// assert.go — assertion (login) ceremony with sign-count clone
// detection.
//
// Two-phase flow per W3C §7.2:
//
//   BeginAssertion(user)           →  options + session blob
//   FinishAssertion(user, session, request)
//                                  →  validated *webauthn.Credential
//
// Clone detection (REQ-FUNC-PLT-IAM-005 acceptance #2):
//
// Per W3C §7.2 step 17 the verifier MUST compare the
// authenticator's reported sign-count against the stored value:
//
//   • strictly greater → update stored value, accept.
//   • equal or smaller (and either side non-zero) → the
//     authenticator may have been cloned — there are now two
//     copies in the wild, both producing assertions, with at
//     least one of their counters lagging.
//
// The protocol library encodes that policy in the returned
// Authenticator.CloneWarning bool. When it fires we:
//
//   1. Disable the credential row (disabled_at = now,
//      disabled_reason = "clone_detected").
//   2. Emit a webauthn.clone_detected AuditEvent.
//   3. Return ErrCloneDetected so the caller fails the login.
//
// The credential remains in the table for forensics; the user
// must register a new key.

package webauthn

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// BeginAssertion starts the assertion ceremony for a known user.
// Returns the PublicKeyCredentialRequestOptions for the browser
// and the SessionData the caller MUST persist until
// FinishAssertion.
func (s *Service) BeginAssertion(_ context.Context, user *User) (*protocol.CredentialAssertion, *webauthn.SessionData, error) {
	if user == nil {
		return nil, nil, errors.New("webauthn: nil user")
	}
	return s.web.BeginLogin(user)
}

// FinishAssertion completes the assertion. Returns the validated
// credential on success.
//
// On clone detection (Authenticator.CloneWarning == true) the
// matched credential is disabled, an audit event is emitted, and
// the call returns ErrCloneDetected. The caller MUST treat that
// error as authentication failure AND surface a re-enrolment flow
// so the user can replace the compromised credential.
func (s *Service) FinishAssertion(ctx context.Context, user *User, session webauthn.SessionData, req *http.Request) (*webauthn.Credential, error) {
	cred, err := s.web.FinishLogin(user, session, req)
	if err != nil {
		_ = s.audit.Emit(ctx, AuditEvent{
			UserID:     string(user.WebAuthnID()),
			Outcome:    OutcomeAssertionFail,
			OccurredAt: s.store.clk().UTC(),
			Reason:     err.Error(),
		})
		return nil, fmt.Errorf("webauthn: finish login: %w", err)
	}

	if cred.Authenticator.CloneWarning {
		_ = s.store.DisableCredential(ctx, cred.ID, "clone_detected")
		_ = s.audit.Emit(ctx, AuditEvent{
			UserID:       string(user.WebAuthnID()),
			CredentialID: encodeCredentialID(cred.ID),
			Outcome:      OutcomeCloneDetected,
			OccurredAt:   s.store.clk().UTC(),
			Reason:       "authenticator sign-count did not strictly increase",
		})
		_ = s.audit.Emit(ctx, AuditEvent{
			UserID:       string(user.WebAuthnID()),
			CredentialID: encodeCredentialID(cred.ID),
			Outcome:      OutcomeCredentialDisabled,
			OccurredAt:   s.store.clk().UTC(),
			Reason:       "clone_detected",
		})
		return nil, ErrCloneDetected
	}

	if err := s.store.UpdateSignCount(ctx, cred.ID, cred.Authenticator.SignCount); err != nil {
		return nil, err
	}
	_ = s.audit.Emit(ctx, AuditEvent{
		UserID:       string(user.WebAuthnID()),
		CredentialID: encodeCredentialID(cred.ID),
		Outcome:      OutcomeAssertionOK,
		OccurredAt:   s.store.clk().UTC(),
	})
	return cred, nil
}

// encodeCredentialID returns the base64url-unpadded form of a
// credential ID. Used in audit events so log consumers see a
// consistent printable identifier.
func encodeCredentialID(id []byte) string {
	return base64.RawURLEncoding.EncodeToString(id)
}

// ErrCloneDetected is returned by FinishAssertion when the
// authenticator's sign-count did not strictly increase from the
// stored value — the explicit W3C-defined signal that the
// credential may have been cloned. The caller MUST fail the login.
var ErrCloneDetected = errors.New("webauthn: authenticator clone detected")
