// Package middleware provides ConnectRPC interceptors for recovery,
// request logging, and correlation-ID propagation. Each interceptor is
// composable; services chain them in main.go via connect.WithInterceptors.
package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/ppusapati/space/pkg/observability"
)

const (
	// HeaderCorrelationID names the header used to propagate a
	// caller-supplied correlation id. When absent, the interceptor
	// generates a fresh UUIDv4.
	HeaderCorrelationID = "X-Correlation-Id"
	// HeaderTenantID identifies the tenant. Required by every
	// non-public RPC.
	HeaderTenantID = "X-Tenant-Id"
)

type ctxKey string

const (
	correlationIDKey ctxKey = "correlation-id"
	tenantIDKey      ctxKey = "tenant-id"
)

// CorrelationID returns the correlation id stored in ctx, or "" if none.
func CorrelationID(ctx context.Context) string {
	if v, ok := ctx.Value(correlationIDKey).(string); ok {
		return v
	}
	return ""
}

// TenantID returns the tenant id stored in ctx, or "" if none.
func TenantID(ctx context.Context) string {
	if v, ok := ctx.Value(tenantIDKey).(string); ok {
		return v
	}
	return ""
}

// Recovery converts panics inside RPC handlers into a Connect Internal
// error, attaching the panic message and a stack trace to the logger
// for diagnostics.
func Recovery(logger *slog.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (resp connect.AnyResponse, err error) {
			defer func() {
				if r := recover(); r != nil {
					stack := debug.Stack()
					logger.ErrorContext(ctx, "panic in handler",
						slog.Any("panic", r),
						slog.String("stack", string(stack)),
					)
					err = connect.NewError(
						connect.CodeInternal,
						fmt.Errorf("internal server error"),
					)
					_ = errors.New("recovered panic")
				}
			}()
			return next(ctx, req)
		}
	}
}

// CorrelationAndTenant is the standard request-context interceptor: it
// stamps a correlation id (echoing the incoming header or generating a
// fresh UUIDv4) and optionally promotes the tenant header to the
// context. Both ids are added to the request-scoped logger and to the
// outbound response headers.
func CorrelationAndTenant() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			cid := req.Header().Get(HeaderCorrelationID)
			if cid == "" {
				cid = uuid.NewString()
			}
			tid := req.Header().Get(HeaderTenantID)
			ctx = context.WithValue(ctx, correlationIDKey, cid)
			if tid != "" {
				ctx = context.WithValue(ctx, tenantIDKey, tid)
			}
			logger := observability.LoggerFromContext(ctx).With(
				slog.String("correlation_id", cid),
			)
			if tid != "" {
				logger = logger.With(slog.String("tenant_id", tid))
			}
			ctx = observability.WithLogger(ctx, logger)
			resp, err := next(ctx, req)
			if resp != nil {
				resp.Header().Set(HeaderCorrelationID, cid)
				if tid != "" {
					resp.Header().Set(HeaderTenantID, tid)
				}
			}
			return resp, err
		}
	}
}

// AccessLog logs one record per RPC at info level with method, status
// code, duration, and error class.
func AccessLog() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			start := time.Now()
			resp, err := next(ctx, req)
			logger := observability.LoggerFromContext(ctx)
			attrs := []any{
				slog.String("rpc", req.Spec().Procedure),
				slog.Duration("dur", time.Since(start)),
			}
			if err != nil {
				code := connect.CodeOf(err)
				attrs = append(attrs, slog.String("code", code.String()), slog.String("err", err.Error()))
				logger.LogAttrs(ctx, slog.LevelWarn, "rpc.error", asAttrs(attrs)...)
			} else {
				logger.LogAttrs(ctx, slog.LevelInfo, "rpc.ok", asAttrs(attrs)...)
			}
			return resp, err
		}
	}
}

func asAttrs(in []any) []slog.Attr {
	out := make([]slog.Attr, 0, len(in))
	for _, v := range in {
		if a, ok := v.(slog.Attr); ok {
			out = append(out, a)
		}
	}
	return out
}
