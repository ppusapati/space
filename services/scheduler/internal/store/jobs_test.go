package store

import (
	"testing"
	"time"
)

func TestRetryPolicy_Backoff(t *testing.T) {
	cases := []struct {
		policy  RetryPolicy
		attempt int
		want    time.Duration
	}{
		{RetryPolicy{MaxAttempts: 1, BackoffS: 0}, 1, 0},
		{RetryPolicy{MaxAttempts: 1, BackoffS: 5}, 1, 0}, // first attempt has no backoff
		{RetryPolicy{MaxAttempts: 3, BackoffS: 10}, 2, 10 * time.Second},
		{RetryPolicy{MaxAttempts: 3, BackoffS: 10}, 3, 20 * time.Second},
		{RetryPolicy{MaxAttempts: 3, BackoffS: 0}, 2, 0},
	}
	for _, c := range cases {
		if got := c.policy.Backoff(c.attempt); got != c.want {
			t.Errorf("Backoff(%d) on %+v: got %v want %v", c.attempt, c.policy, got, c.want)
		}
	}
}

func TestDefaultRetryPolicy(t *testing.T) {
	p := DefaultRetryPolicy()
	if p.MaxAttempts != 1 {
		t.Errorf("max attempts: %d", p.MaxAttempts)
	}
	if p.BackoffS != 0 {
		t.Errorf("backoff: %d", p.BackoffS)
	}
}

func TestStatusConstants_Distinct(t *testing.T) {
	all := map[string]bool{
		StatusRunning:   true,
		StatusSucceeded: true,
		StatusFailed:    true,
		StatusTimeout:   true,
		StatusSkipped:   true,
	}
	if len(all) != 5 {
		t.Errorf("want 5 distinct statuses, got %d", len(all))
	}
}

func TestTriggerConstants(t *testing.T) {
	if TriggerCron == TriggerManual {
		t.Error("triggers must differ")
	}
}
