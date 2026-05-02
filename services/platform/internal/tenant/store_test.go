package tenant

import (
	"testing"
	"time"
)

func TestDefaultSecurityPolicy(t *testing.T) {
	got := DefaultSecurityPolicy()
	if got.SessionIdleTimeout != time.Hour {
		t.Errorf("idle: %v", got.SessionIdleTimeout)
	}
	if got.SessionAbsoluteLimit != 24*time.Hour {
		t.Errorf("absolute: %v", got.SessionAbsoluteLimit)
	}
	if got.MaxConcurrentSessions != 5 {
		t.Errorf("concurrent: %d", got.MaxConcurrentSessions)
	}
	if got.PasswordMinLength != 12 {
		t.Errorf("password min: %d", got.PasswordMinLength)
	}
	if !got.PasswordRequireMixed {
		t.Error("password mixed should default true")
	}
	if got.MFARequired {
		t.Error("MFA should default false (single-tenant v1 posture)")
	}
}

func TestDefaultQuotas(t *testing.T) {
	got := DefaultQuotas()
	if got.MaxUsers != 1_000 {
		t.Errorf("users: %d", got.MaxUsers)
	}
	if got.MaxRolesPerUser != 32 {
		t.Errorf("roles: %d", got.MaxRolesPerUser)
	}
	if got.MaxAPIRequestsHour != 1_000_000 {
		t.Errorf("rps: %d", got.MaxAPIRequestsHour)
	}
}

func TestStatusConstants_DistinctAndKnown(t *testing.T) {
	statuses := map[string]bool{
		StatusActive:    true,
		StatusSuspended: true,
		StatusArchived:  true,
	}
	if len(statuses) != 3 {
		t.Errorf("expected 3 unique statuses, got %d", len(statuses))
	}
}

// NewStore wires up a defaulted clock when nil is supplied.
func TestNewStore_DefaultClock(t *testing.T) {
	s := NewStore(nil, nil)
	if s == nil {
		t.Fatal("NewStore should return a non-nil store even with nil pool")
	}
	if s.clk == nil {
		t.Fatal("clock should default to time.Now")
	}
	now := s.clk()
	if now.IsZero() {
		t.Error("default clock should produce a non-zero time")
	}
}
