package store

import (
	"testing"
	"time"
)

// TestLockoutDurationFor enforces REQ-FUNC-PLT-IAM-003 ladder.
func TestLockoutDurationFor(t *testing.T) {
	cases := map[int]time.Duration{
		0: 0,
		1: 15 * time.Minute,
		2: 1 * time.Hour,
		3: 24 * time.Hour,
		4: 0, // out of band — defensive
	}
	for level, want := range cases {
		if got := lockoutDurationFor(level); got != want {
			t.Errorf("lockoutDurationFor(%d)=%v, want %v", level, got, want)
		}
	}
}

// TestUser_IsActive covers the status check helper.
func TestUser_IsActive(t *testing.T) {
	cases := map[string]bool{
		StatusActive:              true,
		StatusPendingVerification: false,
		StatusDisabled:            false,
		StatusDeleted:             false,
		"unknown":                 false,
	}
	for status, want := range cases {
		u := &User{Status: status}
		if got := u.IsActive(); got != want {
			t.Errorf("IsActive(%q)=%v, want %v", status, got, want)
		}
	}
}

// TestUser_IsLockedAt + LockoutRemaining cover the per-row lockout
// helpers.
func TestUser_LockoutHelpers(t *testing.T) {
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	t.Run("not locked when LockedUntil nil", func(t *testing.T) {
		u := &User{}
		if u.IsLockedAt(now) {
			t.Error("IsLockedAt should be false when LockedUntil is nil")
		}
		if got := u.LockoutRemaining(now); got != 0 {
			t.Errorf("LockoutRemaining=%v, want 0", got)
		}
	})

	t.Run("locked when LockedUntil in future", func(t *testing.T) {
		until := now.Add(10 * time.Minute)
		u := &User{LockedUntil: &until}
		if !u.IsLockedAt(now) {
			t.Error("IsLockedAt should be true when LockedUntil > now")
		}
		if got := u.LockoutRemaining(now); got != 10*time.Minute {
			t.Errorf("LockoutRemaining=%v, want 10m", got)
		}
	})

	t.Run("not locked when LockedUntil elapsed", func(t *testing.T) {
		past := now.Add(-time.Minute)
		u := &User{LockedUntil: &past}
		if u.IsLockedAt(now) {
			t.Error("IsLockedAt should be false when LockedUntil < now")
		}
		if got := u.LockoutRemaining(now); got != 0 {
			t.Errorf("LockoutRemaining after expiry = %v, want 0", got)
		}
	})
}
