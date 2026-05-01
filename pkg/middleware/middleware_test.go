package middleware

import (
	"bytes"
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/ppusapati/space/pkg/observability"
)

func makeReq(headers map[string]string) connect.AnyRequest {
	r := connect.NewRequest(&emptypb.Empty{})
	for k, v := range headers {
		r.Header().Set(k, v)
	}
	return r
}

func TestCorrelationAndTenantStampsContext(t *testing.T) {
	icpt := CorrelationAndTenant()
	var seen struct{ cid, tid string }
	next := func(ctx context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		seen.cid = CorrelationID(ctx)
		seen.tid = TenantID(ctx)
		return connect.NewResponse(&emptypb.Empty{}), nil
	}
	req := makeReq(map[string]string{
		HeaderCorrelationID: "abc",
		HeaderTenantID:      "tenant-1",
	})
	resp, err := icpt(next)(context.Background(), req)
	if err != nil {
		t.Fatalf("interceptor err: %v", err)
	}
	if seen.cid != "abc" || seen.tid != "tenant-1" {
		t.Fatalf("ctx values: %+v", seen)
	}
	if resp.Header().Get(HeaderCorrelationID) != "abc" {
		t.Fatalf("response missing correlation id; got headers %v", resp.Header())
	}
}

func TestCorrelationAndTenantGeneratesIDWhenMissing(t *testing.T) {
	icpt := CorrelationAndTenant()
	var seen string
	next := func(ctx context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		seen = CorrelationID(ctx)
		return connect.NewResponse(&emptypb.Empty{}), nil
	}
	if _, err := icpt(next)(context.Background(), makeReq(nil)); err != nil {
		t.Fatalf("interceptor err: %v", err)
	}
	if seen == "" {
		t.Fatal("expected generated correlation id")
	}
}

func TestRecoveryConvertsPanicToInternal(t *testing.T) {
	var buf bytes.Buffer
	logger := observability.NewLogger(observability.LogConfig{Level: "error", Service: "test", Environment: "test", Writer: &buf})
	icpt := Recovery(logger)
	next := func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		panic("boom")
	}
	_, err := icpt(next)(context.Background(), makeReq(nil))
	if err == nil {
		t.Fatal("expected error from recovered panic")
	}
	var ce *connect.Error
	if !errors.As(err, &ce) || ce.Code() != connect.CodeInternal {
		t.Fatalf("expected Internal connect error, got %v", err)
	}
	if !strings.Contains(buf.String(), "panic in handler") {
		t.Fatal("recovery did not log the panic")
	}
}

func TestAccessLogEmitsAttrs(t *testing.T) {
	var buf bytes.Buffer
	logger := observability.NewLogger(observability.LogConfig{Level: "info", Service: "test", Environment: "dev", Writer: &buf})
	icpt := AccessLog()
	next := func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		return connect.NewResponse(&emptypb.Empty{}), nil
	}
	ctx := observability.WithLogger(context.Background(), logger)
	if _, err := icpt(next)(ctx, makeReq(nil)); err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.Contains(buf.String(), "rpc.ok") {
		t.Fatalf("access log missing rpc.ok: %s", buf.String())
	}
}

// Ensures the http.NewServer test recorder pattern still imports.
var _ = httptest.NewRecorder
