// sar.go — Article 15 (Right of access) endpoint.
//
// The user (or their controller acting on their behalf) requests
// a Subject Access Report covering every personal datum the
// platform holds about them. The chetana SP:
//
//   1. Builds the IAM-side Snapshot synchronously (cheap — under
//      a few queries).
//   2. Hands the snapshot to the export service via Exporter.
//      The export service merges it with rows from every other
//      domain service, packages the bundle, and posts it to S3
//      with a short-lived presigned URL.
//
// The user receives a JobID and polls /gdpr/sar/{job_id} until
// the URL is ready. We're well within the 30-day GDPR window —
// the actual job typically completes in minutes.

package gdpr

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SARRequest is the per-call input.
type SARRequest struct {
	UserID         string
	RequestorIP    string
	RequestorAgent string
}

// SARResult is the per-call output.
type SARResult struct {
	JobID         JobID
	Snapshot      *Snapshot
	AcceptedAt    time.Time
}

// SARService handles Article 15 requests.
type SARService struct {
	pool     *pgxpool.Pool
	builder  *SnapshotBuilder
	exporter Exporter
	clk      func() time.Time
}

// NewSARService wires a pool + Exporter into a SAR handler.
// builder=nil → constructs one over the same pool.
func NewSARService(pool *pgxpool.Pool, exporter Exporter, builder *SnapshotBuilder, clock func() time.Time) (*SARService, error) {
	if pool == nil {
		return nil, errors.New("gdpr: nil pool")
	}
	if exporter == nil {
		return nil, errors.New("gdpr: nil exporter")
	}
	if clock == nil {
		clock = time.Now
	}
	if builder == nil {
		builder = NewSnapshotBuilder(pool, clock)
	}
	return &SARService{pool: pool, builder: builder, exporter: exporter, clk: clock}, nil
}

// Request runs the synchronous half of an Article 15 SAR:
//   1. Build the snapshot.
//   2. Enqueue the export job, passing the snapshot through.
//   3. Return the job id + snapshot to the caller.
//
// The caller (the HTTP / Connect handler) responds with the JobID
// + a poll URL; the snapshot is also returned so a "preview"
// flow can show the user their data immediately.
func (s *SARService) Request(ctx context.Context, in SARRequest) (*SARResult, error) {
	if in.UserID == "" {
		return nil, ErrUserNotFound
	}
	snap, err := s.builder.Build(ctx, in.UserID)
	if err != nil {
		return nil, err
	}
	jobID, err := s.exporter.EnqueueSAR(ctx, EnqueueSARInput{
		UserID:         snap.User.ID,
		TenantID:       snap.User.TenantID,
		RequestorIP:    in.RequestorIP,
		RequestorAgent: in.RequestorAgent,
		Snapshot:       snap,
	})
	if err != nil {
		return nil, fmt.Errorf("gdpr: enqueue sar: %w", err)
	}
	return &SARResult{
		JobID:      jobID,
		Snapshot:   snap,
		AcceptedAt: s.clk().UTC(),
	}, nil
}
