package login

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// TestIPLimiter_DefaultsApplied covers the constructor's value-fill
// path so callers passing a zero IPLimiterConfig get the documented
// 10/min limit.
func TestIPLimiter_DefaultsApplied(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"}) // not dialled
	defer rdb.Close()
	l := NewIPLimiter(rdb, IPLimiterConfig{})
	if l.cfg.MaxAttemptsPerWindow != 10 {
		t.Errorf("MaxAttemptsPerWindow=%d, want 10", l.cfg.MaxAttemptsPerWindow)
	}
	if l.cfg.Window != time.Minute {
		t.Errorf("Window=%v, want 1m", l.cfg.Window)
	}
	if l.cfg.KeyPrefix != "iam:login:ip:" {
		t.Errorf("KeyPrefix=%q, want iam:login:ip:", l.cfg.KeyPrefix)
	}
	if l.cfg.Now == nil {
		t.Error("Now func should be defaulted")
	}
}

// TestIPLimiter_RejectsEmptyIP exercises the empty-IP guard. The
// guard short-circuits before any Redis call so the test runs
// against an unconnected client and never hits the network. Real
// integration coverage of the admit/deny + Retry-After path lives
// in services/iam/test/login_e2e_test.go (build tag: integration).
func TestIPLimiter_RejectsEmptyIP(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"}) // never dialled
	defer rdb.Close()
	l := NewIPLimiter(rdb, IPLimiterConfig{})
	_, err := l.Allow(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty IP")
	}
	if !strings.Contains(err.Error(), "empty client IP") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestIPLimiter_AcceptsClockOverride documents the test-injected
// clock contract.
func TestIPLimiter_AcceptsClockOverride(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})
	defer rdb.Close()
	frozen := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)
	l := NewIPLimiter(rdb, IPLimiterConfig{
		Now: func() time.Time { return frozen },
	})
	if got := l.cfg.Now(); !got.Equal(frozen) {
		t.Errorf("clock override not honoured: got %v, want %v", got, frozen)
	}
}

// TestIPLimiterConfig_PreservesExplicitValues covers the
// constructor's "respect non-zero values" branch.
func TestIPLimiterConfig_PreservesExplicitValues(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})
	defer rdb.Close()
	l := NewIPLimiter(rdb, IPLimiterConfig{
		MaxAttemptsPerWindow: 50,
		Window:               5 * time.Minute,
		KeyPrefix:            "test:",
	})
	if l.cfg.MaxAttemptsPerWindow != 50 {
		t.Errorf("max=%d, want 50", l.cfg.MaxAttemptsPerWindow)
	}
	if l.cfg.Window != 5*time.Minute {
		t.Errorf("window=%v, want 5m", l.cfg.Window)
	}
	if l.cfg.KeyPrefix != "test:" {
		t.Errorf("prefix=%q, want test:", l.cfg.KeyPrefix)
	}
}

// errSentinel is a helper for asserting unwrapping in callers.
var errSentinel = errors.New("sentinel")
