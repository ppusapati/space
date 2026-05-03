package runner

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/redis/go-redis/v9"

	"github.com/ppusapati/space/services/scheduler/internal/lock"
	"github.com/ppusapati/space/services/scheduler/internal/store"
)

// fakeRedis returns a *redis.Client that is never connected; the
// Locker constructor accepts it but no Acquire happens here.
func fakeRedis() *redis.Client {
	return redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})
}

// newFakeLocker constructs a real *lock.Locker over a never-
// dialled redis client. Safe because the unit tests below only
// exercise the constructor.
func newFakeLocker(t *testing.T) *lock.Locker {
	t.Helper()
	l, err := lock.NewLocker(fakeRedis())
	if err != nil {
		t.Fatalf("locker: %v", err)
	}
	return l
}

func TestRegistry_RegisterLookup(t *testing.T) {
	r := NewRegistry()
	r.Register("daily-cleanup", ExecuteFunc(func(ctx context.Context, j *store.Job) (Result, error) {
		return Result{ExitCode: 0, Output: "ok"}, nil
	}))
	e, err := r.Lookup("daily-cleanup")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	res, err := e.Execute(context.Background(), &store.Job{})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if res.Output != "ok" {
		t.Errorf("output: %q", res.Output)
	}
}

func TestRegistry_LookupUnknown(t *testing.T) {
	r := NewRegistry()
	if _, err := r.Lookup("missing"); !errors.Is(err, ErrNoExecutor) {
		t.Errorf("got %v want ErrNoExecutor", err)
	}
}

func TestNew_RejectsMissingDeps(t *testing.T) {
	locker := newFakeLocker(t)
	cases := []func(c *Config){
		func(c *Config) { c.Store = nil },
		func(c *Config) { c.Locker = nil },
		func(c *Config) { c.Registry = nil },
	}
	for _, mut := range cases {
		cfg := Config{Store: &store.JobStore{}, Locker: locker, Registry: NewRegistry()}
		mut(&cfg)
		if _, err := New(cfg); err == nil {
			t.Errorf("expected error for mutation")
		}
	}
}

func TestNew_AppliesDefaults(t *testing.T) {
	r, err := New(Config{
		Store: &store.JobStore{}, Locker: newFakeLocker(t), Registry: NewRegistry(),
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if r.id == "" {
		t.Error("id default")
	}
}

func TestExcerpt(t *testing.T) {
	short := "short"
	if got := excerpt(short); got != short {
		t.Errorf("short pass-through: %q", got)
	}
	long := strings.Repeat("x", 1500)
	got := excerpt(long)
	if !strings.HasSuffix(got, "…") {
		t.Errorf("long should be truncated with ellipsis")
	}
	if len(got) > 1100 {
		t.Errorf("excerpt length: %d", len(got))
	}
	if got := excerpt("  trim me  "); got != "trim me" {
		t.Errorf("trim: %q", got)
	}
}

func TestMax(t *testing.T) {
	if max(3, 7) != 7 {
		t.Error("max(3,7)")
	}
	if max(7, 3) != 7 {
		t.Error("max(7,3)")
	}
}

