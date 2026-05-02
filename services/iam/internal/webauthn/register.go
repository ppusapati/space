// register.go — registration ceremony wrapper.
//
// Two-phase flow per W3C §7.1:
//
//   BeginRegistration(user)        →  options + session blob
//                                     (caller serialises the options
//                                     to the browser, persists the
//                                     session blob in the user's
//                                     server-side session row).
//
//   FinishRegistration(user, session, request)
//                                  →  validated *webauthn.Credential
//                                     written to webauthn_credentials.
//
// The protocol library does the heavy lifting (clientDataJSON +
// CBOR attestation parse, signature verification, attestation-format
// dispatch, RP-ID hash check, challenge match, origin check,
// algorithm-allow-list enforcement). We persist the result and emit
// an audit event.

package webauthn

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// Service is the chetana wrapper around *webauthn.WebAuthn (the
// protocol library's main type) plus our store + audit emitter.
// Construct with NewService.
type Service struct {
	web   *webauthn.WebAuthn
	store *Store
	audit AuditEmitter
}

// Config configures the Service.
type Config struct {
	// RPDisplayName is the human-readable name the authenticator
	// shows to the user (e.g. "Chetana").
	RPDisplayName string

	// RPID is the Relying Party identifier — the effective domain
	// of the IAM service (e.g. "chetana.p9e.in"). Authenticators
	// hash this into the credential and refuse to release it on
	// any other domain. MUST match the browser's location origin
	// hostname (or be a registrable suffix of it).
	RPID string

	// RPOrigins is the allow-list of origins from which assertions
	// may be presented (e.g. ["https://chetana.p9e.in"]).
	RPOrigins []string

	// AttestationPreference is "none" | "indirect" | "direct" |
	// "enterprise". Most chetana flows use "direct" so the audit
	// trail captures the attestation; passkey-style flows pass
	// "none" because the platform attesters are uniformly trusted.
	AttestationPreference protocol.ConveyancePreference

	// AuthenticatorSelection lets the caller pin to platform
	// authenticators (TouchID/Windows Hello) vs cross-platform
	// (security keys) and require user verification (UV) vs
	// presence (UP) only.
	AuthenticatorSelection protocol.AuthenticatorSelection
}

// NewService validates the config and returns a Service. The
// underlying *webauthn.WebAuthn config is validated lazily (on
// first call) by the library; we surface the error early so a
// misconfigured boot fails before serving traffic.
func NewService(cfg Config, store *Store, audit AuditEmitter) (*Service, error) {
	if cfg.RPID == "" {
		return nil, errors.New("webauthn: empty RPID")
	}
	if cfg.RPDisplayName == "" {
		cfg.RPDisplayName = "Chetana"
	}
	if len(cfg.RPOrigins) == 0 {
		return nil, errors.New("webauthn: at least one RPOrigin is required")
	}
	if cfg.AttestationPreference == "" {
		cfg.AttestationPreference = protocol.PreferDirectAttestation
	}
	if store == nil {
		return nil, errors.New("webauthn: nil store")
	}
	if audit == nil {
		audit = NopAudit{}
	}
	web, err := webauthn.New(&webauthn.Config{
		RPDisplayName:          cfg.RPDisplayName,
		RPID:                   cfg.RPID,
		RPOrigins:              cfg.RPOrigins,
		AttestationPreference:  cfg.AttestationPreference,
		AuthenticatorSelection: cfg.AuthenticatorSelection,
	})
	if err != nil {
		return nil, fmt.Errorf("webauthn: config: %w", err)
	}
	return &Service{web: web, store: store, audit: audit}, nil
}

// BeginRegistration starts the registration ceremony. Returns the
// PublicKeyCredentialCreationOptions the caller serialises to the
// browser and the SessionData the caller MUST persist (typically
// in the server-side session row) until FinishRegistration is
// called.
//
// The exclusion list is built from the user's currently-active
// credentials so an authenticator that already enrolled cannot
// silently re-register.
func (s *Service) BeginRegistration(_ context.Context, user *User) (*protocol.CredentialCreation, *webauthn.SessionData, error) {
	if user == nil {
		return nil, nil, errors.New("webauthn: nil user")
	}
	exclusions := webauthn.Credentials(user.WebAuthnCredentials()).CredentialDescriptors()
	return s.web.BeginRegistration(user, webauthn.WithExclusions(exclusions))
}

// FinishRegistration completes the registration ceremony. The
// protocol library validates everything; we persist the resulting
// credential + emit an audit event.
//
// Returns ErrCredentialExists when the authenticator returned an
// id that's already on file (defence in depth — BeginRegistration
// already excludes the active set).
func (s *Service) FinishRegistration(ctx context.Context, user *User, session webauthn.SessionData, req *http.Request) (*webauthn.Credential, error) {
	cred, err := s.web.FinishRegistration(user, session, req)
	if err != nil {
		return nil, fmt.Errorf("webauthn: finish registration: %w", err)
	}
	if err := s.store.SaveCredential(ctx, string(user.WebAuthnID()), cred); err != nil {
		return nil, err
	}
	_ = s.audit.Emit(ctx, AuditEvent{
		UserID:       string(user.WebAuthnID()),
		CredentialID: encodeCredentialID(cred.ID),
		Outcome:      OutcomeRegistered,
		OccurredAt:   s.store.clk().UTC(),
	})
	return cred, nil
}
