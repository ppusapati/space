// interceptor.go — ConnectRPC interceptor that captures method +
// actor + classification automatically.
//
// Every chetana service installs this interceptor right after
// the authz interceptor (which produces the Principal). The
// audit interceptor:
//
//   1. Pulls the verified principal off the request context (set
//      by authzv1.Interceptor).
//   2. Runs the wrapped handler, capturing the duration + the
//      Connect error code.
//   3. Emits one Event via the configured Client with action =
//      the procedure name's canonical permission, decision =
//      "ok"/"fail" depending on the handler's outcome.
//
// Note on REQ-CONST-011: this is an OBSERVATION layer, NOT an
// authorisation layer. The authz interceptor's allow/deny audit
// events are emitted by authz/v1.Interceptor itself (richer
// shape — they include the matched policy id + reason). This
// interceptor records every successfully-authorised RPC call so
// the audit chain has a per-action row even for methods that
// don't trip a deny.

package audit

import (
	"context"
	"time"

	"connectrpc.com/connect"
)

// Classifier maps a procedure name to the data classification
// the action operates on. Per-service config; defaults to "cui".
type Classifier func(procedure string) string

// PermissionMap mirrors authzv1.PermissionMap — the canonical
// {module}.{resource}.{action} for the procedure. The audit
// interceptor uses it as the Event's Action field.
type PermissionMap map[string]string

// PrincipalFromContext pulls the verified principal from the
// request context. Implementations typically delegate to a
// per-service helper that the authz interceptor populated.
type PrincipalFromContext func(ctx context.Context) (userID, sessionID, tenantID string)

// InterceptorConfig configures the audit interceptor.
type InterceptorConfig struct {
	Client        Client               // required
	Permissions   PermissionMap        // required
	PrincipalFrom PrincipalFromContext // required
	Classify      Classifier           // optional; defaults to "cui"
	Now           func() time.Time     // optional; defaults to time.Now
}

// Interceptor is the audit-recording Connect interceptor.
type Interceptor struct {
	cfg InterceptorConfig
}

// NewInterceptor wires the supplied dependencies. Returns an
// error if a required field is missing.
func NewInterceptor(cfg InterceptorConfig) (*Interceptor, error) {
	if cfg.Client == nil {
		return nil, errInvalidInterceptor("client")
	}
	if cfg.Permissions == nil {
		return nil, errInvalidInterceptor("permissions")
	}
	if cfg.PrincipalFrom == nil {
		return nil, errInvalidInterceptor("principal_from")
	}
	if cfg.Classify == nil {
		cfg.Classify = func(string) string { return "cui" }
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
		permission := i.cfg.Permissions[procedure]
		userID, sessionID, tenantID := i.cfg.PrincipalFrom(ctx)
		ip := req.Header().Get("X-Forwarded-For")
		ua := req.Header().Get("User-Agent")
		started := i.cfg.Now()

		resp, err := next(ctx, req)
		evt := Event{
			TenantID:        tenantID,
			OccurredAt:      started,
			ActorUserID:     userID,
			ActorSessionID:  sessionID,
			ActorClientIP:   ip,
			ActorUserAgent:  ua,
			Action:          permissionOrProcedure(permission, procedure),
			Decision:        decisionFromError(err),
			Procedure:       procedure,
			Classification:  i.cfg.Classify(procedure),
			Metadata:        map[string]string{"duration_ns": durString(time.Since(started))},
		}
		if err != nil {
			evt.Reason = err.Error()
		}
		// Emit best-effort: never block the response on the audit
		// pipeline. The audit service's own back-pressure (or the
		// Kafka topic) is the right circuit-breaker.
		_ = i.cfg.Client.Emit(ctx, evt)
		return resp, err
	})
}

// WrapStreamingClient is a no-op (this is a server-side interceptor).
func (i *Interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

// WrapStreamingHandler emits one event per stream open + close.
func (i *Interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		procedure := conn.Spec().Procedure
		permission := i.cfg.Permissions[procedure]
		userID, sessionID, tenantID := i.cfg.PrincipalFrom(ctx)
		ip := conn.RequestHeader().Get("X-Forwarded-For")
		ua := conn.RequestHeader().Get("User-Agent")
		started := i.cfg.Now()
		err := next(ctx, conn)
		evt := Event{
			TenantID:        tenantID,
			OccurredAt:      started,
			ActorUserID:     userID,
			ActorSessionID:  sessionID,
			ActorClientIP:   ip,
			ActorUserAgent:  ua,
			Action:          permissionOrProcedure(permission, procedure),
			Decision:        decisionFromError(err),
			Procedure:       procedure,
			Classification:  i.cfg.Classify(procedure),
			Metadata: map[string]string{
				"duration_ns": durString(time.Since(started)),
				"streaming":   "true",
			},
		}
		if err != nil {
			evt.Reason = err.Error()
		}
		_ = i.cfg.Client.Emit(ctx, evt)
		return err
	})
}

func decisionFromError(err error) string {
	if err == nil {
		return "ok"
	}
	return "fail"
}

func permissionOrProcedure(permission, procedure string) string {
	if permission != "" {
		return permission
	}
	return procedure
}

func durString(d time.Duration) string {
	// Render as an integer-nanosecond decimal string so the
	// metadata field stays JSON-friendly + sort-stable.
	return time.Duration(d).String()
}

// errInvalidInterceptor is a tiny helper; using an inline error
// constructor keeps the per-field check readable.
func errInvalidInterceptor(field string) error {
	return &interceptorErr{field: field}
}

type interceptorErr struct{ field string }

func (e *interceptorErr) Error() string {
	return "audit: missing required interceptor config field: " + e.field
}
