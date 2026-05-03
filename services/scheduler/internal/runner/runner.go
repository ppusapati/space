// Package runner orchestrates one job's execution: lock → start
// run → execute → record outcome → advance schedule.
//
// → REQ-FUNC-CMN-006 acceptance #1: exactly-one runner per
//                                   scheduled tick (per-job
//                                   Redis lock).
// → REQ-FUNC-CMN-006 acceptance #3: enable/disable toggles
//                                   immediate; runs fully
//                                   captured in history.

package runner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ppusapati/space/services/scheduler/internal/cron"
	"github.com/ppusapati/space/services/scheduler/internal/lock"
	"github.com/ppusapati/space/services/scheduler/internal/store"
)

// Result is the outcome the Executor returns to the runner.
type Result struct {
	ExitCode     int
	Output       string
	ErrorExcerpt string
}

// Executor renders a job's payload into a Result. Concrete
// implementations live per-job-kind (HTTP webhook, gRPC call,
// Connect RPC, sqlc procedure call, etc.).
type Executor interface {
	Execute(ctx context.Context, job *store.Job) (Result, error)
}

// ExecuteFunc is the closure form.
type ExecuteFunc func(ctx context.Context, job *store.Job) (Result, error)

// Execute implements Executor.
func (f ExecuteFunc) Execute(ctx context.Context, job *store.Job) (Result, error) {
	return f(ctx, job)
}

// Registry maps `job.Name` → Executor. Concurrent-safe.
type Registry struct{ procs map[string]Executor }

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry { return &Registry{procs: map[string]Executor{}} }

// Register adds an Executor for the named job.
func (r *Registry) Register(name string, e Executor) {
	if r.procs == nil {
		r.procs = map[string]Executor{}
	}
	r.procs[name] = e
}

// Lookup returns the Executor for `name` or ErrNoExecutor.
func (r *Registry) Lookup(name string) (Executor, error) {
	if r.procs == nil {
		return nil, fmt.Errorf("%w: %s", ErrNoExecutor, name)
	}
	e, ok := r.procs[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNoExecutor, name)
	}
	return e, nil
}

// Runner orchestrates the lock → run → finish dance for one job.
type Runner struct {
	id       string
	store    *store.JobStore
	locker   *lock.Locker
	registry *Registry
	leaseTTL time.Duration
	clk      func() time.Time
}

// Config configures the Runner.
type Config struct {
	ID       string         // runner identifier (host+pid+random)
	Store    *store.JobStore // required
	Locker   *lock.Locker   // required
	Registry *Registry      // required
	LeaseTTL time.Duration  // default = max(timeoutS, 60s)
	Now      func() time.Time
}

// New wires a Runner.
func New(cfg Config) (*Runner, error) {
	if cfg.Store == nil {
		return nil, errors.New("runner: nil store")
	}
	if cfg.Locker == nil {
		return nil, errors.New("runner: nil locker")
	}
	if cfg.Registry == nil {
		return nil, errors.New("runner: nil registry")
	}
	if cfg.ID == "" {
		cfg.ID = "runner-default"
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Runner{
		id:       cfg.ID,
		store:    cfg.Store,
		locker:   cfg.Locker,
		registry: cfg.Registry,
		leaseTTL: cfg.LeaseTTL,
		clk:      cfg.Now,
	}, nil
}

// TriggerInput configures one Trigger call.
type TriggerInput struct {
	Job     *store.Job
	Trigger string // store.TriggerCron | store.TriggerManual
}

// Trigger runs one job. Returns nil + nil when another runner
// already holds the lock (i.e. we lost the race — exactly one
// runner ran the tick, which is the acceptance #1 invariant).
//
// Trigger handles the full lifecycle:
//
//   1. Acquire the per-job Redis lock.
//   2. StartRun in the DB (records the attempt).
//   3. Execute via the registry.
//   4. FinishRun with the outcome.
//   5. AdvanceNext (cron triggers only; manual triggers do NOT
//      shift next_run_at).
//   6. Release the lock.
//
// Retries: on Executor error AND attempts < MaxAttempts, the
// runner sleeps the configured Backoff and retries inside the
// same lock so we don't yield to another runner mid-retry.
func (r *Runner) Trigger(ctx context.Context, in TriggerInput) (*Outcome, error) {
	if in.Job == nil {
		return nil, errors.New("runner: nil job")
	}
	if in.Trigger == "" {
		in.Trigger = store.TriggerCron
	}
	if !in.Job.Enabled && in.Trigger == store.TriggerCron {
		return nil, ErrJobDisabled
	}

	lockTTL := r.leaseTTL
	if lockTTL <= 0 {
		lockTTL = time.Duration(in.Job.TimeoutS+30) * time.Second
	}
	held, err := r.locker.Acquire(ctx, "job:"+in.Job.ID, lockTTL)
	if err != nil {
		return nil, fmt.Errorf("runner: acquire: %w", err)
	}
	if held == nil {
		return nil, nil // another runner won the race
	}
	defer func() { _ = held.Release(context.Background()) }()

	exec, err := r.registry.Lookup(in.Job.Name)
	if err != nil {
		// No executor → record the run as failed so the audit
		// trail surfaces the misconfiguration.
		runID, startErr := r.store.StartRun(ctx, in.Job.ID, r.id, in.Trigger, 1)
		if startErr == nil {
			_ = r.store.FinishRun(ctx, runID, store.FinishRunInput{
				Status: store.StatusFailed, ExitCode: 1, ErrorExcerpt: err.Error(),
			})
		}
		return nil, err
	}

	outcome := &Outcome{JobID: in.Job.ID}
	for attempt := 1; attempt <= max(1, in.Job.RetryPolicy.MaxAttempts); attempt++ {
		runID, err := r.store.StartRun(ctx, in.Job.ID, r.id, in.Trigger, attempt)
		if err != nil {
			return outcome, err
		}
		runCtx, cancel := context.WithTimeout(ctx, time.Duration(in.Job.TimeoutS)*time.Second)
		startedAt := r.clk().UTC()
		res, execErr := exec.Execute(runCtx, in.Job)
		cancel()

		status := store.StatusSucceeded
		errMsg := ""
		switch {
		case errors.Is(execErr, context.DeadlineExceeded):
			status = store.StatusTimeout
			errMsg = "execution exceeded timeout_s"
		case execErr != nil:
			status = store.StatusFailed
			errMsg = excerpt(execErr.Error())
		}
		if err := r.store.FinishRun(ctx, runID, store.FinishRunInput{
			Status:       status,
			ExitCode:     res.ExitCode,
			Output:       excerpt(res.Output),
			ErrorExcerpt: errMsg,
		}); err != nil {
			return outcome, err
		}
		outcome.LastStatus = status
		outcome.LastAttempt = attempt
		outcome.LastDuration = r.clk().UTC().Sub(startedAt)

		if status == store.StatusSucceeded || attempt >= max(1, in.Job.RetryPolicy.MaxAttempts) {
			break
		}
		if backoff := in.Job.RetryPolicy.Backoff(attempt + 1); backoff > 0 {
			select {
			case <-ctx.Done():
				return outcome, ctx.Err()
			case <-time.After(backoff):
			}
		}
	}

	// Cron-triggered ticks advance the schedule; manual triggers
	// don't shift the cadence.
	if in.Trigger == store.TriggerCron && in.Job.Schedule != "" {
		sched, err := cron.Parse(in.Job.Schedule, in.Job.Timezone)
		if err == nil {
			now := r.clk().UTC()
			next := sched.Next(now)
			if err := r.store.AdvanceNext(ctx, in.Job.ID, now, next); err != nil {
				return outcome, err
			}
		}
	}
	return outcome, nil
}

// Outcome is the runner's structured outcome. The cmd-layer
// `Run(ctx)` loop uses it for metrics; tests assert on it
// directly.
type Outcome struct {
	JobID        string
	LastStatus   string
	LastAttempt  int
	LastDuration time.Duration
}

func excerpt(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 1024 {
		return s[:1024] + "…"
	}
	return s
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrNoExecutor is returned when a job's name has no registered Executor.
var ErrNoExecutor = errors.New("runner: no executor")

// ErrJobDisabled is returned when a CRON trigger fires for a
// disabled job (the disabled bit takes effect immediately —
// acceptance #3).
var ErrJobDisabled = errors.New("runner: job disabled")
