// Package cleanup runs the daily sweep that removes expired
// exports + their S3 objects.
//
// → REQ-FUNC-CMN-005 acceptance #3: cleanup removes S3 objects +
//   DB rows for jobs older than retention.

package cleanup

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ppusapati/space/services/export/internal/queue"
	"github.com/ppusapati/space/services/export/internal/s3"
)

// Sweeper removes expired exports.
type Sweeper struct {
	store    *queue.Store
	uploader s3.Uploader
	limit    int
	clk      func() time.Time
}

// Config configures the Sweeper.
type Config struct {
	Store    *queue.Store
	Uploader s3.Uploader
	Limit    int           // max jobs per run; default 100
	Now      func() time.Time
}

// New wires a Sweeper.
func New(cfg Config) (*Sweeper, error) {
	if cfg.Store == nil {
		return nil, errors.New("cleanup: nil store")
	}
	if cfg.Uploader == nil {
		return nil, errors.New("cleanup: nil uploader")
	}
	if cfg.Limit <= 0 {
		cfg.Limit = 100
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Sweeper{
		store: cfg.Store, uploader: cfg.Uploader,
		limit: cfg.Limit, clk: cfg.Now,
	}, nil
}

// SweepResult is the outcome of one Sweep call.
type SweepResult struct {
	Inspected int
	Deleted   int
	Errors    int
}

// Sweep removes up to Limit expired jobs in one pass.
func (s *Sweeper) Sweep(ctx context.Context) (SweepResult, error) {
	expired, err := s.store.ListExpired(ctx, s.limit)
	if err != nil {
		return SweepResult{}, fmt.Errorf("cleanup: list: %w", err)
	}
	res := SweepResult{Inspected: len(expired)}
	for _, j := range expired {
		// Delete the S3 object (best-effort; missing object is OK).
		if j.S3Bucket != "" && j.S3Key != "" {
			if err := s.uploader.Delete(ctx, j.S3Bucket, j.S3Key); err != nil {
				res.Errors++
				continue
			}
		}
		if err := s.store.MarkExpired(ctx, j.ID); err != nil {
			res.Errors++
			continue
		}
		res.Deleted++
	}
	return res, nil
}

// Run loops on a daily ticker until ctx is cancelled. Exposed
// directly so cmd/export can wire it.
func (s *Sweeper) Run(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	// initial sweep so a fresh boot processes any backlog before
	// waiting `interval`.
	_, _ = s.Sweep(ctx)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			_, _ = s.Sweep(ctx)
		}
	}
}
