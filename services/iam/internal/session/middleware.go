// middleware.go — service-side hook that every protected RPC's
// auth interceptor calls.
//
// The interceptor on each chetana service does:
//
//   1. Pull the access token off the incoming request.
//   2. Verify it via authz/v1.Verifier (cryptographic check;
//      no DB I/O).
//   3. Call session.Validate(ctx, principal.SessionID) on the
//      *gateway* (or the IAM service itself) to consult the live
//      session state — revocation + idle-timeout + absolute-
//      expiry checks the JWT can't carry on its own.
//
// Validate is intentionally a tiny, framework-agnostic surface so
// the same hook serves the eventual Connect interceptor, the
// realtime-gw WebSocket upgrade, and any direct HTTP middleware.
// The result is either nil (proceed) or one of the typed errors
// from manager.go that the framework wrapper translates into the
// appropriate transport-level status code + audit event.

package session

import (
	"context"
	"errors"
)

// Validator is the small interface the framework wrappers depend
// on. Manager satisfies it; tests substitute a fake.
type Validator interface {
	Touch(ctx context.Context, sessionID string) (*Status, error)
}

// Validate runs the session liveness check for the principal's
// session_id. Returns nil when the session is still active and
// has been touched (last_seen_at bumped).
//
// Returns one of the typed errors from manager.go on failure —
// callers should map them to user-facing reasons via Reason().
func Validate(ctx context.Context, v Validator, sessionID string) error {
	if v == nil {
		return errors.New("session: nil validator")
	}
	_, err := v.Touch(ctx, sessionID)
	return err
}

// Reason translates a session error into the canonical machine-
// readable reason string the audit pipeline + the
// WWW-Authenticate header use. Returns empty string for nil and
// for errors not produced by this package.
func Reason(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, ErrSessionNotFound):
		return "session_not_found"
	case errors.Is(err, ErrSessionRevoked):
		return "session_revoked"
	case errors.Is(err, ErrSessionIdleTimeout):
		return "session_idle_timeout"
	case errors.Is(err, ErrSessionAbsoluteExpired):
		return "session_absolute_expired"
	}
	return ""
}
