package mfa

import (
	"testing"
	"time"
)

// The replay-cache logic lives on Store but exercises no DB. We
// construct a Store with a nil pool and call only the in-memory
// methods, which is safe.
func newReplayOnlyStore(now time.Time) *Store {
	return &Store{
		clk:    func() time.Time { return now },
		replay: make(map[replayKey]struct{}),
	}
}

func TestConsumeReplayWindow_FirstSeenWins(t *testing.T) {
	now := time.Unix(1_700_000_010, 0)
	s := newReplayOnlyStore(now)

	if !s.ConsumeReplayWindow("user-1", 100, "123456") {
		t.Error("first presentation must succeed")
	}
	if s.ConsumeReplayWindow("user-1", 100, "123456") {
		t.Error("replay must be rejected")
	}
}

func TestConsumeReplayWindow_DifferentUserNotBlocked(t *testing.T) {
	now := time.Unix(1_700_000_010, 0)
	s := newReplayOnlyStore(now)

	if !s.ConsumeReplayWindow("user-1", 100, "123456") {
		t.Error("user-1 first presentation")
	}
	if !s.ConsumeReplayWindow("user-2", 100, "123456") {
		t.Error("user-2 must not be blocked by user-1's entry")
	}
}

func TestConsumeReplayWindow_DifferentStepNotBlocked(t *testing.T) {
	now := time.Unix(1_700_000_010, 0)
	s := newReplayOnlyStore(now)

	if !s.ConsumeReplayWindow("user-1", 100, "123456") {
		t.Error("step 100")
	}
	if !s.ConsumeReplayWindow("user-1", 101, "123456") {
		t.Error("step 101 must not be blocked")
	}
}

func TestConsumeReplayWindow_GarbageCollectsStaleEntries(t *testing.T) {
	now := time.Unix(1_700_000_010, 0)
	s := newReplayOnlyStore(now)

	// Seed an entry for an old step.
	oldStep := uint64(now.Unix()/StepSeconds) - 50
	s.replay[replayKey{UserID: "u", Step: oldStep, Code: "111111"}] = struct{}{}

	// Force the GC threshold to be in the past so the next call sweeps.
	s.replayCleanupAfter = now.Add(-time.Second)

	if !s.ConsumeReplayWindow("u", uint64(now.Unix()/StepSeconds), "222222") {
		t.Error("fresh insertion failed")
	}

	if _, present := s.replay[replayKey{UserID: "u", Step: oldStep, Code: "111111"}]; present {
		t.Error("stale entry should have been swept")
	}
}
