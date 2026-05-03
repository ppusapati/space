// Package worker drains the export queue, runs the kind-specific
// processor, uploads the result to S3, and writes the presigned
// pointer back to the job row.
//
// Pluggable design: Processors register against (kind →
// ProcessFunc). The chetana service registers `gdpr_sar`,
// `audit_csv`, `audit_json` once those producers are wired;
// new export shapes (NetCDF, GeoTIFF) plug in without touching
// the queue or worker scaffolding.

package worker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ppusapati/space/services/export/internal/queue"
	"github.com/ppusapati/space/services/export/internal/s3"
)

// ProcessOutput is what a Processor returns to the worker.
type ProcessOutput struct {
	Body        []byte
	ContentType string
	Filename    string // used to compose the S3 key
}

// Processor renders a queued job into a body the worker uploads
// to S3.
type Processor interface {
	Process(ctx context.Context, job *queue.Job) (ProcessOutput, error)
}

// ProcessFunc is the closure form of Processor.
type ProcessFunc func(ctx context.Context, job *queue.Job) (ProcessOutput, error)

// Process implements Processor.
func (f ProcessFunc) Process(ctx context.Context, job *queue.Job) (ProcessOutput, error) {
	return f(ctx, job)
}

// Registry maps job.Kind → Processor. Concurrent-safe.
type Registry struct {
	procs map[string]Processor
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry { return &Registry{procs: map[string]Processor{}} }

// Register adds (or replaces) a processor for the given kind.
func (r *Registry) Register(kind string, p Processor) {
	if r.procs == nil {
		r.procs = map[string]Processor{}
	}
	r.procs[kind] = p
}

// Lookup returns the processor for the kind or ErrNoProcessor.
func (r *Registry) Lookup(kind string) (Processor, error) {
	p, ok := r.procs[kind]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNoProcessor, kind)
	}
	return p, nil
}

// Worker drains the queue.
type Worker struct {
	id             string
	store          *queue.Store
	uploader       s3.Uploader
	registry       *Registry
	bucket         string
	leaseTTL       time.Duration
	pollInterval   time.Duration
	presignedFor   time.Duration
}

// Config configures the Worker.
type Config struct {
	ID            string        // worker identifier (host+pid)
	Store         *queue.Store  // required
	Uploader      s3.Uploader   // required
	Registry      *Registry     // required
	Bucket        string        // S3 bucket; required
	LeaseTTL      time.Duration // default 60s
	PollInterval  time.Duration // default 1s
	PresignedFor  time.Duration // default 24h (REQ-FUNC-CMN-005 acceptance #1)
}

// New wires a Worker.
func New(cfg Config) (*Worker, error) {
	if cfg.Store == nil {
		return nil, errors.New("worker: nil store")
	}
	if cfg.Uploader == nil {
		return nil, errors.New("worker: nil uploader")
	}
	if cfg.Registry == nil {
		return nil, errors.New("worker: nil registry")
	}
	if cfg.Bucket == "" {
		return nil, errors.New("worker: bucket is required")
	}
	if cfg.ID == "" {
		cfg.ID = "worker-default"
	}
	if cfg.LeaseTTL <= 0 {
		cfg.LeaseTTL = 60 * time.Second
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = time.Second
	}
	if cfg.PresignedFor <= 0 {
		cfg.PresignedFor = 24 * time.Hour
	}
	return &Worker{
		id:           cfg.ID,
		store:        cfg.Store,
		uploader:     cfg.Uploader,
		registry:     cfg.Registry,
		bucket:       cfg.Bucket,
		leaseTTL:     cfg.LeaseTTL,
		pollInterval: cfg.PollInterval,
		presignedFor: cfg.PresignedFor,
	}, nil
}

// Run loops until ctx is cancelled. Each iteration: try to
// checkout a job; if one is available, process it; if not, sleep
// `PollInterval`.
func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.RunOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
				// best-effort; the next tick retries.
				_ = err
			}
		}
	}
}

// RunOnce handles at most one job. Exposed for tests.
func (w *Worker) RunOnce(ctx context.Context) error {
	job, err := w.store.Checkout(ctx, w.id, w.leaseTTL)
	if err != nil {
		return err
	}
	if job == nil {
		return nil
	}
	return w.process(ctx, job)
}

// process runs the kind-specific Processor + uploads the result +
// writes the presigned pointer back. On error the job is marked
// failed (which the queue's Fail() helper translates into either
// requeue-for-retry or terminal-failed depending on attempts).
func (w *Worker) process(ctx context.Context, job *queue.Job) error {
	proc, err := w.registry.Lookup(job.Kind)
	if err != nil {
		_ = w.store.Fail(ctx, job.ID, w.id, err.Error())
		return err
	}
	out, err := proc.Process(ctx, job)
	if err != nil {
		_ = w.store.Fail(ctx, job.ID, w.id, err.Error())
		return err
	}
	key := composeKey(job, out.Filename)
	upload, err := w.uploader.Upload(ctx, s3.UploadInput{
		Bucket:      w.bucket,
		Key:         key,
		Body:        out.Body,
		ContentType: out.ContentType,
	})
	if err != nil {
		_ = w.store.Fail(ctx, job.ID, w.id, err.Error())
		return err
	}
	url, expiresAt, err := w.uploader.Presign(ctx, upload.Bucket, upload.Key, w.presignedFor)
	if err != nil {
		_ = w.store.Fail(ctx, job.ID, w.id, err.Error())
		return err
	}
	if err := w.store.Complete(ctx, job.ID, w.id, queue.CompleteOutput{
		S3Bucket:       upload.Bucket,
		S3Key:          upload.Key,
		PresignedURL:   url,
		PresignedUntil: expiresAt,
		BytesTotal:     upload.BytesTotal,
	}); err != nil {
		return err
	}
	return nil
}

// composeKey builds an S3 key shape that's stable + scannable +
// per-tenant. Pattern:
//
//	exports/<tenant_id>/<kind>/<yyyy>/<mm>/<job_id>-<filename>
func composeKey(job *queue.Job, filename string) string {
	if filename == "" {
		filename = "output.bin"
	}
	t := job.EnqueuedAt.UTC()
	return fmt.Sprintf("exports/%s/%s/%04d/%02d/%s-%s",
		job.TenantID, sanitiseKind(job.Kind),
		t.Year(), int(t.Month()), job.ID, filename)
}

func sanitiseKind(k string) string {
	k = strings.TrimSpace(k)
	if k == "" {
		return "unknown"
	}
	// Replace anything that isn't alphanumeric/underscore with
	// underscore so the S3 key stays clean.
	out := make([]byte, 0, len(k))
	for i := 0; i < len(k); i++ {
		c := k[i]
		switch {
		case c >= 'a' && c <= 'z',
			c >= 'A' && c <= 'Z',
			c >= '0' && c <= '9',
			c == '_', c == '-':
			out = append(out, c)
		default:
			out = append(out, '_')
		}
	}
	return string(out)
}

// ErrNoProcessor is returned when a job's kind has no registered
// processor.
var ErrNoProcessor = errors.New("worker: no processor for kind")
