// Package webauthn wraps the W3C-conformant go-webauthn protocol
// library with the chetana-specific glue: credential persistence,
// sign-count clone-detection policy, and audit-event emission.
//
// → REQ-FUNC-PLT-IAM-005 (WebAuthn Level 2; sign-count clone
//                          detection disables the credential and
//                          emits an audit event).
// → design.md §4.1.1 (MFA factor list includes WebAuthn).
//
// We delegate the W3C protocol parsing/verification (clientDataJSON,
// CBOR attestation, COSE key extraction, signature verification
// across RSA/EC2/OKP, attestation-format checks for none/packed/
// fido-u2f/tpm/android-key/android-safetynet/apple) to
// github.com/go-webauthn/webauthn rather than re-implementing
// security-critical crypto from scratch. This package owns:
//
//   - Postgres persistence of the [webauthn.Credential] records
//     (the User adapter loads them; Register/Assert mutate them).
//   - The chetana clone-detection policy: when the protocol
//     library's Authenticator.CloneWarning fires (per W3C §7.2
//     step 17 — sign-count did not strictly increase), this
//     package marks the credential disabled, emits a
//     "webauthn.clone_detected" audit event, and returns
//     ErrCloneDetected so the caller fails the assertion.

package webauthn

import (
	"context"
	"time"
)

// AuditEvent is one entry in the WebAuthn audit stream. Mirrors
// the shape used by the login + token packages for a uniform
// downstream audit consumer (REQ-NFR-AUDIT-001).
type AuditEvent struct {
	UserID       string
	CredentialID string // base64url, the credential's id attribute
	Outcome      AuditOutcome
	OccurredAt   time.Time
	Reason       string
}

// AuditOutcome enumerates the WebAuthn audit outcomes the platform
// records. Most-important is OutcomeCloneDetected — the explicit
// W3C-defined signal that an authenticator has been cloned.
type AuditOutcome string

// Canonical audit outcomes.
const (
	OutcomeRegistered     AuditOutcome = "registered"
	OutcomeAssertionOK    AuditOutcome = "assertion_ok"
	OutcomeAssertionFail  AuditOutcome = "assertion_fail"
	OutcomeCloneDetected  AuditOutcome = "clone_detected"
	OutcomeCredentialDisabled AuditOutcome = "credential_disabled"
)

// AuditEmitter publishes a WebAuthn event to the audit pipeline.
// The IAM service writes via Kafka in production
// (TASK-P1-AUDIT-001 supplies the wire); a no-op implementation is
// used in unit tests.
type AuditEmitter interface {
	Emit(ctx context.Context, event AuditEvent) error
}

// NopAudit is a no-op AuditEmitter useful for tests.
type NopAudit struct{}

// Emit implements AuditEmitter.
func (NopAudit) Emit(_ context.Context, _ AuditEvent) error { return nil }
