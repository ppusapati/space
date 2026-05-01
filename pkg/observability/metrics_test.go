package observability

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthzReturnsOK(t *testing.T) {
	reg := NewMetricsRegistry()
	srv := MetricsServer(":0", reg, nil)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/healthz", nil)
	srv.Handler.ServeHTTP(rr, req)
	if rr.Code != 200 || rr.Body.String() != "ok" {
		t.Fatalf("healthz: code=%d body=%q", rr.Code, rr.Body.String())
	}
}

func TestReadyzGatedByCallback(t *testing.T) {
	reg := NewMetricsRegistry()
	ready := false
	srv := MetricsServer(":0", reg, func() bool { return ready })
	rr := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, httptest.NewRequest("GET", "/readyz", nil))
	if rr.Code != 503 {
		t.Fatalf("readyz: expected 503, got %d", rr.Code)
	}
	ready = true
	rr = httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, httptest.NewRequest("GET", "/readyz", nil))
	if rr.Code != 200 {
		t.Fatalf("readyz: expected 200 after ready, got %d", rr.Code)
	}
}

func TestMetricsEndpointExposesGoCollector(t *testing.T) {
	reg := NewMetricsRegistry()
	srv := MetricsServer(":0", reg, nil)
	rr := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, httptest.NewRequest("GET", "/metrics", nil))
	if rr.Code != 200 {
		t.Fatalf("metrics: code=%d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "go_goroutines") {
		t.Fatalf("expected go_goroutines metric, got: %s", rr.Body.String())
	}
}
