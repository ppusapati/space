// Package gdpr implements the chetana IAM service's data-subject
// rights endpoints required by GDPR + the platform DPIA.
//
// → REQ-FUNC-PLT-IAM-011 (SAR + erasure + portability + rectification).
// → REQ-COMP-GDPR-001.
// → design.md §9.2 (Records of Processing Activities + DPIA).
//
// Article coverage:
//
//   • Article 15 (Right of access)        — sar.go     — ExportRequest()
//   • Article 16 (Right to rectification) — rectify.go — Rectify()
//   • Article 17 (Right to erasure)       — erase.go   — Erase()
//   • Article 20 (Right to data portability) — portability.go — Snapshot()
//
// Wiring split:
//
// SAR completion is asynchronous. The user-facing "request export"
// endpoint enqueues a job onto the export service (TASK-P1-EXPORT-001)
// and returns a job_id immediately; the export service does the
// heavy lifting (S3 multipart upload, presigned URL, lifecycle
// rule). Until that service ships, the chetana IAM uses the
// `Exporter` interface and a NopExporter for tests / early
// integration. The shape of the call (job_id, expected delivery
// channel) does not change when the real producer plugs in.

package gdpr

import (
	"context"
	"errors"
	"time"
)

// JobID is the opaque identifier the export service assigns to an
// in-flight SAR job. The user polls /gdpr/sar/{job_id} (a future
// endpoint) to discover when the export is ready.
type JobID string

// Exporter enqueues an asynchronous SAR job onto the export
// service (TASK-P1-EXPORT-001). The chetana IAM service does NOT
// implement S3 multipart / presigned URL / lifecycle here — that
// surface lives in the export service.
type Exporter interface {
	// EnqueueSAR enqueues a SAR job for the supplied user. The
	// snapshot is the IAM-side data the export service will
	// merge with rows fetched from every other domain service.
	EnqueueSAR(ctx context.Context, in EnqueueSARInput) (JobID, error)
}

// EnqueueSARInput is the per-call payload.
type EnqueueSARInput struct {
	UserID         string
	TenantID       string
	RequestorIP    string
	RequestorAgent string
	// Snapshot is the IAM-side data captured at request time.
	// Carried through the queue so the export service does NOT
	// need to round-trip back to IAM during job execution.
	Snapshot *Snapshot
}

// NopExporter is a no-op Exporter useful for tests and for early
// integration when the export service is not deployed yet. It
// returns a deterministic synthetic JobID so the caller can test
// the request-side wiring without a real queue.
type NopExporter struct{}

// EnqueueSAR implements Exporter.
func (NopExporter) EnqueueSAR(_ context.Context, in EnqueueSARInput) (JobID, error) {
	if in.UserID == "" {
		return "", errors.New("gdpr: empty user_id")
	}
	return JobID("nop-" + in.UserID + "-" + time.Now().UTC().Format("20060102T150405Z")), nil
}
