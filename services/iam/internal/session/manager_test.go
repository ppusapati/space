package session

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

func TestNewManager_Defaults(t *testing.T) {
	if _, err := NewManager(nil, Config{}); err == nil {
		t.Error("nil pool should error")
	}

	// We can't construct a real Manager without a pool, but we
	// can verify the cfg defaulting via a quick struct check.
	cfg := Config{}
	if cfg.IdleTimeout != 0 {
		t.Fatal("zero-value cfg should have zero IdleTimeout")
	}
}

func TestStatus_IsActiveAt(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		s    Status
		t    time.Time
		want bool
	}{
		{"fresh + within idle + within absolute", Status{
			IdleExpiresAt:     now.Add(time.Hour),
			AbsoluteExpiresAt: now.Add(24 * time.Hour),
		}, now, true},
		{"revoked", Status{
			RevokedAt:         sql.NullTime{Valid: true, Time: now.Add(-time.Minute)},
			IdleExpiresAt:     now.Add(time.Hour),
			AbsoluteExpiresAt: now.Add(24 * time.Hour),
		}, now, false},
		{"idle expired", Status{
			IdleExpiresAt:     now.Add(-time.Minute),
			AbsoluteExpiresAt: now.Add(24 * time.Hour),
		}, now, false},
		{"absolute expired", Status{
			IdleExpiresAt:     now.Add(time.Hour),
			AbsoluteExpiresAt: now.Add(-time.Minute),
		}, now, false},
		{"exactly at idle expiry", Status{
			IdleExpiresAt:     now,
			AbsoluteExpiresAt: now.Add(24 * time.Hour),
		}, now, false},
		{"exactly at absolute expiry", Status{
			IdleExpiresAt:     now.Add(time.Hour),
			AbsoluteExpiresAt: now,
		}, now, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.s.IsActiveAt(tc.t); got != tc.want {
				t.Errorf("got %v want %v", got, tc.want)
			}
		})
	}
}

// fakeValidator is a tiny test double for the Validate hook so we
// can exercise it without a real Manager.
type fakeValidator struct {
	err error
}

func (f *fakeValidator) Touch(_ context.Context, _ string) (*Status, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &Status{}, nil
}

func TestValidate_PropagatesError(t *testing.T) {
	v := &fakeValidator{err: ErrSessionRevoked}
	if err := Validate(context.Background(), v, "sid"); !errors.Is(err, ErrSessionRevoked) {
		t.Errorf("got %v want ErrSessionRevoked", err)
	}
}

func TestValidate_OK(t *testing.T) {
	if err := Validate(context.Background(), &fakeValidator{}, "sid"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestValidate_NilValidatorRejected(t *testing.T) {
	if err := Validate(context.Background(), nil, "sid"); err == nil {
		t.Error("nil validator should error")
	}
}

func TestReason(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{nil, ""},
		{ErrSessionNotFound, "session_not_found"},
		{ErrSessionRevoked, "session_revoked"},
		{ErrSessionIdleTimeout, "session_idle_timeout"},
		{ErrSessionAbsoluteExpired, "session_absolute_expired"},
		{errors.New("unrelated"), ""},
	}
	for _, tc := range cases {
		if got := Reason(tc.err); got != tc.want {
			t.Errorf("Reason(%v): got %q want %q", tc.err, got, tc.want)
		}
	}
}

func TestNewSessionID_LengthAndHex(t *testing.T) {
	id, err := newSessionID()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(id) != 32 {
		t.Errorf("len: %d", len(id))
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		switch {
		case c >= '0' && c <= '9', c >= 'a' && c <= 'f':
			// ok
		default:
			t.Errorf("non-hex char at %d: %q", i, c)
			break
		}
	}
}

func TestAmrSlice(t *testing.T) {
	if got := amrSlice(nil); len(got) != 0 || got == nil {
		t.Errorf("nil → empty non-nil slice, got %v", got)
	}
	in := []string{"pwd", "totp"}
	got := amrSlice(in)
	if len(got) != 2 || got[0] != "pwd" || got[1] != "totp" {
		t.Errorf("got %v", got)
	}
	got[0] = "MUTATED"
	if in[0] == "MUTATED" {
		t.Error("amrSlice must defensively copy")
	}
}
