package webauthn

import (
	"context"
	"errors"
	"testing"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// Verify the chetana User adapter satisfies the protocol library's
// User interface — a compile-time + behaviour check.
func TestUser_ImplementsLibraryInterface(t *testing.T) {
	var _ webauthn.User = (*User)(nil)

	u := &User{
		id:          []byte("11111111-1111-1111-1111-111111111111"),
		name:        "user@example.com",
		displayName: "User Example",
		credentials: []webauthn.Credential{
			{ID: []byte("cred-a")},
			{ID: []byte("cred-b")},
		},
	}
	if string(u.WebAuthnID()) != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("WebAuthnID: %q", u.WebAuthnID())
	}
	if u.WebAuthnName() != "user@example.com" {
		t.Errorf("WebAuthnName: %q", u.WebAuthnName())
	}
	if u.WebAuthnDisplayName() != "User Example" {
		t.Errorf("WebAuthnDisplayName: %q", u.WebAuthnDisplayName())
	}

	got := u.WebAuthnCredentials()
	if len(got) != 2 {
		t.Fatalf("creds: got %d want 2", len(got))
	}
	// Defensive copy: mutating returned slice must not affect the user.
	got[0].ID = []byte("MUTATED")
	if string(u.credentials[0].ID) == "MUTATED" {
		t.Error("WebAuthnCredentials must return a defensive copy")
	}
}

func TestNewService_ValidatesConfig(t *testing.T) {
	store := &Store{}
	cases := []struct {
		name string
		cfg  Config
	}{
		{"empty RPID", Config{RPOrigins: []string{"https://x"}}},
		{"empty origins", Config{RPID: "x", RPOrigins: nil}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewService(tc.cfg, store, NopAudit{}); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestNewService_NilStoreRejected(t *testing.T) {
	if _, err := NewService(Config{
		RPID:      "chetana.p9e.in",
		RPOrigins: []string{"https://chetana.p9e.in"},
	}, nil, NopAudit{}); err == nil {
		t.Error("nil store should error")
	}
}

func TestNewService_DefaultsAndBuilds(t *testing.T) {
	svc, err := NewService(Config{
		RPID:      "chetana.p9e.in",
		RPOrigins: []string{"https://chetana.p9e.in"},
	}, &Store{}, nil) // nil audit → NopAudit
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if svc.web == nil {
		t.Error("web not initialised")
	}
	if svc.audit == nil {
		t.Error("audit defaulted to nil")
	}
}

// Acceptance #2: decreasing sign-count → CloneWarning fires.
//
// This exercises the protocol library's UpdateCounter (the
// authoritative implementation of W3C §7.2 step 17) against the
// canonical clone-detection scenarios. Our Service.FinishAssertion
// reads CloneWarning off the same Authenticator and triggers the
// disable + audit path; the integration test wires the full
// end-to-end flow.
func TestAuthenticator_CloneDetection_PolicyMatrix(t *testing.T) {
	cases := []struct {
		name         string
		stored       uint32
		reported     uint32
		wantWarning  bool
	}{
		{"strict increase", 5, 6, false},
		{"large jump", 5, 100, false},
		{"equal nonzero", 5, 5, true},
		{"decrease", 10, 5, true},
		{"decrease to zero", 10, 0, true},
		{"both zero — no signal", 0, 0, false},
		{"first nonzero on a zero-counter device", 0, 1, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := webauthn.Authenticator{SignCount: tc.stored}
			a.UpdateCounter(tc.reported)
			if a.CloneWarning != tc.wantWarning {
				t.Errorf("stored=%d reported=%d: got CloneWarning=%v want %v",
					tc.stored, tc.reported, a.CloneWarning, tc.wantWarning)
			}
		})
	}
}

// recordingAudit captures every emitted event so tests can assert
// the chetana audit chain reflects the W3C-defined clone-detection
// outcome.
type recordingAudit struct {
	events []AuditEvent
}

func (r *recordingAudit) Emit(_ context.Context, e AuditEvent) error {
	r.events = append(r.events, e)
	return nil
}

func TestEncodeCredentialID_Base64URLUnpadded(t *testing.T) {
	got := encodeCredentialID([]byte{0x00, 0x10, 0x20, 0x30})
	want := "ABAgMA"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestParseJoinTransports_Roundtrip(t *testing.T) {
	in := []protocol.AuthenticatorTransport{
		protocol.USB,
		protocol.NFC,
		protocol.Internal,
	}
	got := parseTransports(joinTransports(in))
	if len(got) != len(in) {
		t.Fatalf("len: got %d want %d", len(got), len(in))
	}
	for i := range in {
		if got[i] != in[i] {
			t.Errorf("[%d]: got %q want %q", i, got[i], in[i])
		}
	}

	// Empty roundtrip.
	if got := parseTransports(joinTransports(nil)); len(got) != 0 {
		t.Errorf("empty: got %v", got)
	}

	// Tolerant of stray whitespace.
	if got := parseTransports("usb, nfc , internal"); len(got) != 3 {
		t.Errorf("loose parse: %v", got)
	}
}

func TestErrors_AreSentinel(t *testing.T) {
	for _, e := range []error{
		ErrUserNotFound,
		ErrCredentialExists,
		ErrCredentialNotFound,
		ErrCloneDetected,
	} {
		if e == nil || e.Error() == "" {
			t.Errorf("sentinel error empty: %v", e)
		}
		// errors.Is must be reflexive.
		if !errors.Is(e, e) {
			t.Errorf("errors.Is reflexivity failed for %v", e)
		}
	}
}
