// Package s3 abstracts the chunked-upload + presigned-URL surface
// the export workers depend on.
//
// → REQ-FUNC-CMN-005; design.md §5.2.
//
// We deliberately do NOT pull aws-sdk-go-v2 in here:
//
//   • The cmd layer composes the real client + KMS + FIPS endpoint
//     once TASK-P1-PLT-SECRETS-001 wires the credentials chain.
//
//   • Tests use NopUploader (records bucket+key, generates a
//     deterministic synthetic URL) so the worker's orchestration
//     can be exercised without an S3 bucket.
//
// FIPS posture: callers MUST hand the cmd-layer FIPS-validated
// endpoint URL to FIPSAsserts at boot, mirroring the notify
// service's pattern.

package s3

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// UploadInput is the per-call payload for Uploader.Upload.
type UploadInput struct {
	Bucket      string
	Key         string
	Body        []byte
	ContentType string
	Metadata    map[string]string
}

// UploadResult carries the post-upload pointer.
type UploadResult struct {
	Bucket       string
	Key          string
	ETag         string
	BytesTotal   int64
	StorageClass string
}

// Uploader is the abstract surface workers depend on.
type Uploader interface {
	// Upload writes the supplied bytes into S3 via the multipart
	// API. The chetana production implementation streams chunked
	// uploads; this surface accepts a buffered byte slice for
	// API simplicity — the streaming variant lands when a
	// caller produces > 1GB outputs.
	Upload(ctx context.Context, in UploadInput) (UploadResult, error)

	// Presign returns a presigned GET URL for the object that
	// expires after `validFor`. 24h is the chetana default per
	// REQ-FUNC-CMN-005 acceptance #1.
	Presign(ctx context.Context, bucket, key string, validFor time.Duration) (url string, expiresAt time.Time, err error)

	// Delete removes the object. Called by the cleanup sweep.
	Delete(ctx context.Context, bucket, key string) error
}

// NopUploader is a deterministic in-memory Uploader useful for
// tests + the dev posture before AWS creds are wired.
type NopUploader struct {
	Bucket   string // overrides the per-call Bucket when non-empty
	Stored   map[string][]byte
	Deleted  map[string]bool
}

// Upload records the body in memory.
func (n *NopUploader) Upload(_ context.Context, in UploadInput) (UploadResult, error) {
	bucket := n.Bucket
	if bucket == "" {
		bucket = in.Bucket
	}
	if bucket == "" {
		return UploadResult{}, errors.New("s3: bucket is required")
	}
	if n.Stored == nil {
		n.Stored = map[string][]byte{}
	}
	full := bucket + "/" + in.Key
	n.Stored[full] = append([]byte(nil), in.Body...)
	return UploadResult{
		Bucket:       bucket,
		Key:          in.Key,
		ETag:         fmt.Sprintf("nop-%d-%x", len(in.Body), simpleSum(in.Body)),
		BytesTotal:   int64(len(in.Body)),
		StorageClass: "STANDARD",
	}, nil
}

// Presign returns a synthetic URL.
func (n *NopUploader) Presign(_ context.Context, bucket, key string, validFor time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(validFor)
	return fmt.Sprintf("https://nop-bucket.example/%s/%s?expires=%d", bucket, key, expiresAt.Unix()), expiresAt, nil
}

// Delete removes the object.
func (n *NopUploader) Delete(_ context.Context, bucket, key string) error {
	full := bucket + "/" + key
	if n.Deleted == nil {
		n.Deleted = map[string]bool{}
	}
	n.Deleted[full] = true
	if n.Stored != nil {
		delete(n.Stored, full)
	}
	return nil
}

// simpleSum is a fast non-crypto hash used only for the
// NopUploader's synthetic ETag.
func simpleSum(b []byte) uint64 {
	var sum uint64
	for _, c := range b {
		sum = sum*31 + uint64(c)
	}
	return sum
}

// FIPSAsserts validates that the supplied S3 endpoint URL targets
// the AWS FIPS endpoint family.
//
// Canonical S3 FIPS endpoints:
//   https://s3-fips.<region>.amazonaws.com
//   https://s3-fips.dualstack.<region>.amazonaws.com
//
// Empty endpoint → error (chetana never uses the global SDK
// default in production).
func FIPSAsserts(endpoint string) error {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return errors.New("s3: endpoint is required (must target a -fips region)")
	}
	if !strings.Contains(endpoint, "s3-fips.") {
		return fmt.Errorf("s3: endpoint %q is not FIPS-validated; expected 's3-fips.<region>.amazonaws.com'",
			endpoint)
	}
	return nil
}
