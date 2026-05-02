// interceptor.go — ConnectRPC interceptor every chetana service
// installs. The interceptor:
//
//   1. Pulls the bearer access token from the request header.
//   2. Verifies it via the cross-service Verifier (from verify.go).
//   3. Looks up the permission for the called procedure via the
//      service-supplied PermissionMap.
//   4. Calls Decide(principal, request, policies).
//   5. Calls the optional SessionValidator hook so a revoked /
//      idle-expired session is rejected immediately
//      (REQ-FUNC-PLT-IAM-009 wired across the fleet).
//   6. Emits a structured AuditEvent for every allow OR deny
//      (REQ-FUNC-PLT-AUTHZ-004).
//
// Procedures with no entry in PermissionMap are treated as
// publicly accessible (e.g. /Health/Check). The service is
// responsible for keeping its PermissionMap exhaustive for
// protected RPCs; the spec's `tools/authz/no-bypass.sh` CI guard
// catches services that try to side-step this check.
//
// REQ-CONST-011: every chetana service installs THIS interceptor;
// no service implements its own authorisation. The grep guard in
// `tools/authz/no-bypass.sh` enforces it.

package authzv1

import (
	"context"
	"errors"
	"strings"
	"time"

	"connectrpc.com/connect"
)

// PermissionMap routes Connect procedure names
// (e.g. "/iam.v1.AuthService/Login") to the canonical
// {module}.{resource}.{action} permission identifier the policy
// set knows about. Empty string → public (no check).
type PermissionMap map[string]string

// SessionValidator is the small surface the interceptor calls
// after JWT verification + Decide succeed. session.Manager from
// services/iam/internal/session satisfies it; tests pass a fake.
//
// Implementations MUST return a typed error matching one of the
// session sentinel errors so the interceptor can map to the
// canonical reason string in the audit event.
type SessionValidator interface {
	Touch(ctx context.Context, sessionID string) (any, error)
}

// PolicySource is the snapshot accessor — the IAM policy loader
// publishes new sets through this interface so a hot-reload does
// not require restarting the interceptor.
type PolicySource interface {
	Snapshot() *PolicySet
}

// AuditSink receives every authz event (allow + deny) the
// interceptor emits. The audit-service producer (TASK-P1-AUDIT-001)
// implements this; tests pass a buffered fake.
type AuditSink interface {
	Emit(ctx context.Context, event AuditEvent)
}

// AuditEvent captures one authorisation decision.
type AuditEvent struct {
	OccurredAt      time.Time
	Procedure       string
	Permission      string
	Effect          Effect
	Reason          string
	MatchedPolicyID string
	UserID          string
	TenantID        string
	SessionID       string
	ClientIP        string
	UserAgent       string
	// Roles + ClearanceLevel + IsUSPerson are echoed for the
	// audit trail so a reviewer sees the principal's posture at
	// decision time.
	Roles          []string
	ClearanceLevel string
	IsUSPerson     bool
	// Error is non-empty when the request failed for an
	// authentication / session / decision-engine error rather
	// than a clean allow/deny.
	Error string
}

// NopAudit is a no-op AuditSink useful for tests.
type NopAudit struct{}

// Emit implements AuditSink.
func (NopAudit) Emit(_ context.Context, _ AuditEvent) {}

// InterceptorConfig configures NewInterceptor.
type InterceptorConfig struct {
	// Verifier validates the bearer access token. Required.
	Verifier *Verifier

	// Policies publishes the active rule set. Required.
	Policies PolicySource

	// Permissions maps Connect procedure names to permission
	// identifiers. Required (may be empty if every endpoint is
	// public, but that's almost certainly a bug).
	Permissions PermissionMap

	// Audit receives every allow/deny event. Defaults to NopAudit.
	Audit AuditSink

	// Sessions, when set, is consulted on every successful
	// allow so a revoked / idle-expired session can be rejected
	// (REQ-FUNC-PLT-IAM-009 wire-up).
	Sessions SessionValidator

	// ClientIPHeader names the request header that carries the
	// originating client IP (set by the platform ingress).
	// Defaults to "X-Forwarded-For".
	ClientIPHeader string

	// Now injects a clock for tests. nil → time.Now.
	Now func() time.Time
}

// Interceptor is the chetana authz interceptor.
type Interceptor struct {
	cfg InterceptorConfig
}

// NewInterceptor wires the supplied dependencies. Returns an
// error when a required field is missing.
func NewInterceptor(cfg InterceptorConfig) (*Interceptor, error) {
	if cfg.Verifier == nil {
		return nil, errors.New("authz: verifier is required")
	}
	if cfg.Policies == nil {
		return nil, errors.New("authz: policies source is required")
	}
	if cfg.Permissions == nil {
		return nil, errors.New("authz: permissions map is required (may be empty)")
	}
	if cfg.Audit == nil {
		cfg.Audit = NopAudit{}
	}
	if cfg.ClientIPHeader == "" {
		cfg.ClientIPHeader = "X-Forwarded-For"
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Interceptor{cfg: cfg}, nil
}

// WrapUnary implements connect.Interceptor.
func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		procedure := req.Spec().Procedure
		permission, protected := i.cfg.Permissions[procedure]
		if !protected || permission == "" {
			// Unprotected procedure — pass through without
			// emitting an audit event (the grep guard catches
			// services that mistakenly omit a procedure).
			return next(ctx, req)
		}

		// 1. Verify bearer.
		bearer := bearerFromHeader(req.Header().Get("Authorization"))
		if bearer == "" {
			return nil, i.fail(ctx, procedure, permission, nil,
				clientCtxFromHeader(req.Header(), i.cfg.ClientIPHeader),
				connect.CodeUnauthenticated, "missing bearer token")
		}
		principal, err := i.cfg.Verifier.VerifyAccessToken(ctx, bearer)
		if err != nil {
			return nil, i.fail(ctx, procedure, permission, nil,
				clientCtxFromHeader(req.Header(), i.cfg.ClientIPHeader),
				connect.CodeUnauthenticated, "invalid bearer token: "+err.Error())
		}

		// 2. Decide.
		policies := i.cfg.Policies.Snapshot()
		decision, derr := Decide(principal, Request{
			Permission: permission,
			TenantID:   principal.TenantID,
		}, policies)
		if derr != nil {
			return nil, i.fail(ctx, procedure, permission, principal,
				clientCtxFromHeader(req.Header(), i.cfg.ClientIPHeader),
				connect.CodeInternal, "decision engine: "+derr.Error())
		}
		if decision.Effect != EffectAllow {
			i.cfg.Audit.Emit(ctx, AuditEvent{
				OccurredAt:      i.cfg.Now().UTC(),
				Procedure:       procedure,
				Permission:      permission,
				Effect:          decision.Effect,
				Reason:          decision.Reason,
				MatchedPolicyID: decision.MatchedPolicyID,
				UserID:          principal.UserID,
				TenantID:        principal.TenantID,
				SessionID:       principal.SessionID,
				Roles:           principal.Roles,
				ClearanceLevel:  principal.ClearanceLevel,
				IsUSPerson:      principal.IsUSPerson,
			})
			return nil, connect.NewError(connect.CodePermissionDenied,
				errors.New(decision.Reason))
		}

		// 3. Session liveness — short-circuit if the session was
		// revoked / idle-expired since the JWT was minted.
		if i.cfg.Sessions != nil && principal.SessionID != "" {
			if _, err := i.cfg.Sessions.Touch(ctx, principal.SessionID); err != nil {
				return nil, i.fail(ctx, procedure, permission, principal,
					clientCtxFromHeader(req.Header(), i.cfg.ClientIPHeader),
					connect.CodeUnauthenticated, "session: "+err.Error())
			}
		}

		// 4. Allow → audit + downstream call.
		i.cfg.Audit.Emit(ctx, AuditEvent{
			OccurredAt:      i.cfg.Now().UTC(),
			Procedure:       procedure,
			Permission:      permission,
			Effect:          decision.Effect,
			Reason:          decision.Reason,
			MatchedPolicyID: decision.MatchedPolicyID,
			UserID:          principal.UserID,
			TenantID:        principal.TenantID,
			SessionID:       principal.SessionID,
			Roles:           principal.Roles,
			ClearanceLevel:  principal.ClearanceLevel,
			IsUSPerson:      principal.IsUSPerson,
		})
		return next(ctx, req)
	})
}

// WrapStreamingClient implements connect.Interceptor with a no-op
// — chetana services consume connect handlers; client-side auth
// is owned by the caller.
func (i *Interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

// WrapStreamingHandler implements connect.Interceptor for
// server-side streaming RPCs. Same flow as unary; the actual
// stream body is delegated to `next` after the auth + decision
// gates pass.
func (i *Interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		procedure := conn.Spec().Procedure
		permission, protected := i.cfg.Permissions[procedure]
		if !protected || permission == "" {
			return next(ctx, conn)
		}
		bearer := bearerFromHeader(conn.RequestHeader().Get("Authorization"))
		if bearer == "" {
			return i.failStream(ctx, procedure, permission, nil,
				clientCtxFromHeader(conn.RequestHeader(), i.cfg.ClientIPHeader),
				connect.CodeUnauthenticated, "missing bearer token")
		}
		principal, err := i.cfg.Verifier.VerifyAccessToken(ctx, bearer)
		if err != nil {
			return i.failStream(ctx, procedure, permission, nil,
				clientCtxFromHeader(conn.RequestHeader(), i.cfg.ClientIPHeader),
				connect.CodeUnauthenticated, "invalid bearer token: "+err.Error())
		}
		decision, derr := Decide(principal, Request{
			Permission: permission, TenantID: principal.TenantID,
		}, i.cfg.Policies.Snapshot())
		if derr != nil {
			return i.failStream(ctx, procedure, permission, principal,
				clientCtxFromHeader(conn.RequestHeader(), i.cfg.ClientIPHeader),
				connect.CodeInternal, "decision engine: "+derr.Error())
		}
		if decision.Effect != EffectAllow {
			i.cfg.Audit.Emit(ctx, AuditEvent{
				OccurredAt:      i.cfg.Now().UTC(),
				Procedure:       procedure,
				Permission:      permission,
				Effect:          decision.Effect,
				Reason:          decision.Reason,
				MatchedPolicyID: decision.MatchedPolicyID,
				UserID:          principal.UserID,
				TenantID:        principal.TenantID,
				SessionID:       principal.SessionID,
				Roles:           principal.Roles,
				ClearanceLevel:  principal.ClearanceLevel,
				IsUSPerson:      principal.IsUSPerson,
			})
			return connect.NewError(connect.CodePermissionDenied, errors.New(decision.Reason))
		}
		if i.cfg.Sessions != nil && principal.SessionID != "" {
			if _, err := i.cfg.Sessions.Touch(ctx, principal.SessionID); err != nil {
				return i.failStream(ctx, procedure, permission, principal,
					clientCtxFromHeader(conn.RequestHeader(), i.cfg.ClientIPHeader),
					connect.CodeUnauthenticated, "session: "+err.Error())
			}
		}
		i.cfg.Audit.Emit(ctx, AuditEvent{
			OccurredAt:      i.cfg.Now().UTC(),
			Procedure:       procedure,
			Permission:      permission,
			Effect:          decision.Effect,
			Reason:          decision.Reason,
			MatchedPolicyID: decision.MatchedPolicyID,
			UserID:          principal.UserID,
			TenantID:        principal.TenantID,
			SessionID:       principal.SessionID,
			Roles:           principal.Roles,
			ClearanceLevel:  principal.ClearanceLevel,
			IsUSPerson:      principal.IsUSPerson,
		})
		return next(ctx, conn)
	})
}

// fail emits a deny audit event and returns the wrapped connect
// error. Used by the failure branches in WrapUnary.
func (i *Interceptor) fail(
	ctx context.Context, procedure, permission string,
	principal *Principal, clientCtx clientCtx,
	code connect.Code, reason string,
) error {
	ev := AuditEvent{
		OccurredAt: i.cfg.Now().UTC(),
		Procedure:  procedure,
		Permission: permission,
		Effect:     EffectDeny,
		Reason:     reason,
		ClientIP:   clientCtx.IP,
		UserAgent:  clientCtx.UA,
		Error:      reason,
	}
	if principal != nil {
		ev.UserID = principal.UserID
		ev.TenantID = principal.TenantID
		ev.SessionID = principal.SessionID
		ev.Roles = principal.Roles
		ev.ClearanceLevel = principal.ClearanceLevel
		ev.IsUSPerson = principal.IsUSPerson
	}
	i.cfg.Audit.Emit(ctx, ev)
	return connect.NewError(code, errors.New(reason))
}

// failStream is fail's streaming-handler twin.
func (i *Interceptor) failStream(
	ctx context.Context, procedure, permission string,
	principal *Principal, clientCtx clientCtx,
	code connect.Code, reason string,
) error {
	return i.fail(ctx, procedure, permission, principal, clientCtx, code, reason)
}

// clientCtx captures the originating IP + user agent for the audit
// event. The interceptor is OK if neither is populated — the audit
// row simply has empty strings.
type clientCtx struct {
	IP string
	UA string
}

func clientCtxFromHeader(h interface {
	Get(string) string
}, ipHeader string) clientCtx {
	return clientCtx{
		IP: h.Get(ipHeader),
		UA: h.Get("User-Agent"),
	}
}

// bearerFromHeader strips the "Bearer " prefix.
func bearerFromHeader(s string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(s, prefix) {
		return ""
	}
	return strings.TrimSpace(s[len(prefix):])
}
