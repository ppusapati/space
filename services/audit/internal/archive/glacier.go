// Package archive ships audit-events chunks to S3 Glacier for the
// cold (year-6 → year-12) retention tier.
//
// → REQ-FUNC-PLT-AUDIT-005 acceptance #3: records older than 5y
//   archived to Glacier; pointer stored in audit_archives table.
//
// The actual S3 multipart upload + Glacier transition lives in
// the export service (TASK-P1-EXPORT-001) — chetana-pattern: the
// audit service hands its chunks to a generic Archiver interface
// and lets the export service own the S3 wire format.
//
// Until EXPORT-001 ships we use NopArchiver: the chunk is built +
// the envelope is signed, but no upload happens. The audit_archives
// row is still written so a future re-run with the real Archiver
// can pick up where the dev posture left off.

package archive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ppusapati/space/services/audit/internal/export"
	"github.com/ppusapati/space/services/audit/internal/search"
)

// Archiver is the surface the chetana audit service's Glacier
// shipper depends on. Real impl ships to S3 + transitions to
// Glacier; NopArchiver records the call but does no I/O.
type Archiver interface {
	Upload(ctx context.Context, in UploadInput) (UploadResult, error)
}

// UploadInput is the per-call payload.
type UploadInput struct {
	TenantID   string
	RangeStart time.Time
	RangeEnd   time.Time
	Body       []byte // the JSON export bytes
	Envelope   *export.Envelope
}

// UploadResult carries the S3 pointer + bytes-uploaded the
// caller persists in audit_archives.
type UploadResult struct {
	Bucket          string
	Key             string
	ETag            string
	BytesCompressed int64
	StorageClass    string
}

// NopArchiver is a no-op Archiver for tests + early integration.
// It returns a deterministic synthetic key so the DB row write
// path can be exercised without S3.
type NopArchiver struct {
	Bucket string // optional; defaults to "chetana-audit-cold"
}

// Upload implements Archiver.
func (n NopArchiver) Upload(_ context.Context, in UploadInput) (UploadResult, error) {
	if in.TenantID == "" {
		return UploadResult{}, errors.New("archive: empty tenant_id")
	}
	bucket := n.Bucket
	if bucket == "" {
		bucket = "chetana-audit-cold"
	}
	key := fmt.Sprintf("audit/%s/%s.json",
		in.TenantID, in.RangeStart.UTC().Format("20060102T150405Z"))
	return UploadResult{
		Bucket:          bucket,
		Key:             key,
		StorageClass:    "GLACIER",
		BytesCompressed: int64(len(in.Body)),
	}, nil
}

// Service ties the search + export + archiver helpers together.
type Service struct {
	pool     *pgxpool.Pool
	exporter *export.JSONExporter
	archiver Archiver
	search   *search.Service
	clk      func() time.Time
}

// NewService wraps the dependencies.
func NewService(pool *pgxpool.Pool, exporter *export.JSONExporter, archiver Archiver, sv *search.Service, clock func() time.Time) (*Service, error) {
	if pool == nil {
		return nil, errors.New("archive: nil pool")
	}
	if exporter == nil {
		return nil, errors.New("archive: nil exporter")
	}
	if archiver == nil {
		return nil, errors.New("archive: nil archiver")
	}
	if sv == nil {
		return nil, errors.New("archive: nil search service")
	}
	if clock == nil {
		clock = time.Now
	}
	return &Service{pool: pool, exporter: exporter, archiver: archiver, search: sv, clk: clock}, nil
}

// ArchiveRange archives every audit_events row for `tenantID` in
// [start, end] to Glacier and writes a pointer row into
// audit_archives. Idempotent: re-running the same range produces
// the same s3_key and the audit_archives UNIQUE constraint
// rejects the duplicate insert.
func (s *Service) ArchiveRange(ctx context.Context, tenantID string, start, end time.Time) (*UploadResult, error) {
	if tenantID == "" {
		return nil, errors.New("archive: empty tenant_id")
	}
	if !end.After(start) {
		return nil, errors.New("archive: end must be after start")
	}

	// Build the export body in memory. Acceptance #3 only fires
	// for chunks older than 5 years — typical chunk size is one
	// month, so the in-memory buffer is bounded.
	buf := newSizedBuffer()
	q := search.Query{TenantID: tenantID, Start: start, End: end}
	env, err := s.exporter.Export(ctx, q, buf)
	if err != nil {
		return nil, fmt.Errorf("archive: export: %w", err)
	}
	if env.RowCount == 0 {
		return nil, ErrNoRowsToArchive
	}

	res, err := s.archiver.Upload(ctx, UploadInput{
		TenantID:   tenantID,
		RangeStart: start,
		RangeEnd:   end,
		Body:       buf.Bytes(),
		Envelope:   env,
	})
	if err != nil {
		return nil, fmt.Errorf("archive: upload: %w", err)
	}

	// Persist the pointer + the envelope so a future re-verify
	// can rebuild the chain attestation without re-pulling the
	// archive bytes.
	envBytes, err := json.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("archive: marshal envelope: %w", err)
	}
	if _, err := s.pool.Exec(ctx, `
INSERT INTO audit_archives
  (tenant_id, range_start, range_end, first_chain_seq,
   last_chain_seq, row_count, s3_bucket, s3_key,
   s3_storage_class, s3_etag, bytes_compressed, envelope)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
ON CONFLICT (s3_bucket, s3_key) DO NOTHING
`, tenantID, start.UTC(), end.UTC(),
		env.FirstChainSeq, env.LastChainSeq, int64(env.RowCount),
		res.Bucket, res.Key, res.StorageClass, res.ETag, res.BytesCompressed,
		envBytes,
	); err != nil {
		return nil, fmt.Errorf("archive: insert pointer: %w", err)
	}
	return &res, nil
}

// ErrNoRowsToArchive is returned by ArchiveRange when the
// supplied range contains no events. Idempotent — caller can
// safely treat as "nothing to do".
var ErrNoRowsToArchive = errors.New("archive: no rows in range")

// sizedBuffer is a tiny in-memory buffer that tracks bytes for
// the upload size report. We use bytes.Buffer underneath via a
// tiny wrapper kept local so the internal/archive surface stays
// dep-thin.
type sizedBuffer struct {
	buf []byte
}

func newSizedBuffer() *sizedBuffer { return &sizedBuffer{} }

func (b *sizedBuffer) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *sizedBuffer) Bytes() []byte { return b.buf }
