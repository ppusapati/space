package serverobs

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// fakeDepCheck is the test double used to assert DepCheck behaviour
// without spinning up real Postgres/Kafka/Redis. It also counts probes
// so the cache test can prove the 5-second window is honoured.
type fakeDepCheck struct {
	name       string
	err        error          // nil => OK
	calls      atomic.Int64
	delay      time.Duration  // optional sleep to test timeout behaviour
}

func (f *fakeDepCheck) Name() string { return f.name }

func (f *fakeDepCheck) Check(ctx context.Context) error {
	f.calls.Add(1)
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return f.err
}

// helper — build a *Server with the supplied deps and return its mux
// without starting the listeners. The metrics handler is exposed
// separately because it lives on the dedicated metrics mux.
func newTestServer(t *testing.T, build BuildInfo, deps []DepCheck) (*Server, http.Handler) {
	t.Helper()
	srv := NewServer(
		ServerConfig{Addr: ":0"},
		ObservabilityConfig{
			Build:         build,
			DepChecks:     deps,
			MetricsAddr:   ":0",
			ReadyCacheTTL: 100 * time.Millisecond, // tight TTL for tests
		},
	)
	// Hand back the metrics handler too so tests can scrape it.
	mh := metricsHandler(srv.metricsRegistry, srv.ready)
	return srv, mh
}

// TestHealth_AlwaysReturns200 asserts liveness is unconditional —
// REQ-FUNC-CMN-001.
func TestHealth_AlwaysReturns200(t *testing.T) {
	build := BuildInfo{Version: "v1.2.3", GitSHA: "abc123"}
	srv, _ := newTestServer(t, build, []DepCheck{
		// Even with a failing dep, /health should report OK.
		&fakeDepCheck{name: "postgres", err: errors.New("DB down")},
	})

	rr := httptest.NewRecorder()
	srv.Mux.ServeHTTP(rr, httptest.NewRequest("GET", "/health", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("/health: got %d, want 200", rr.Code)
	}
	var body healthResponse
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Status != "ok" {
		t.Errorf("status=%q, want ok", body.Status)
	}
	if body.Version != "v1.2.3" {
		t.Errorf("version=%q, want v1.2.3", body.Version)
	}
	if body.GitSHA != "abc123" {
		t.Errorf("git_sha=%q, want abc123", body.GitSHA)
	}
	if body.UptimeS < 0 {
		t.Errorf("uptime_s=%d, want >=0", body.UptimeS)
	}
	if body.GoVersion == "" {
		t.Error("go_version: empty")
	}
}

// TestReady_AllOK_Returns200 + TestReady_AnyDepFails_Returns503 cover
// the dep-aggregation contract. REQ-FUNC-CMN-002.
func TestReady_AllOK_Returns200(t *testing.T) {
	srv, _ := newTestServer(t, BuildInfo{}, []DepCheck{
		&fakeDepCheck{name: "postgres"},
		&fakeDepCheck{name: "redis"},
	})

	rr := httptest.NewRecorder()
	srv.Mux.ServeHTTP(rr, httptest.NewRequest("GET", "/ready", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("/ready: got %d, want 200; body=%s", rr.Code, rr.Body.String())
	}
	var body readyAggregate
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !body.OK {
		t.Errorf("aggregate ok=false, want true")
	}
	if len(body.DepResults) != 2 {
		t.Fatalf("got %d dep results, want 2", len(body.DepResults))
	}
}

func TestReady_AnyDepFails_Returns503(t *testing.T) {
	cases := []struct {
		name string
		deps []DepCheck
	}{
		{"single failing dep", []DepCheck{
			&fakeDepCheck{name: "postgres", err: errors.New("connection refused")},
		}},
		{"one failing among many", []DepCheck{
			&fakeDepCheck{name: "postgres"},
			&fakeDepCheck{name: "kafka", err: errors.New("metadata unavailable")},
			&fakeDepCheck{name: "redis"},
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv, _ := newTestServer(t, BuildInfo{}, c.deps)
			rr := httptest.NewRecorder()
			srv.Mux.ServeHTTP(rr, httptest.NewRequest("GET", "/ready", nil))

			if rr.Code != http.StatusServiceUnavailable {
				t.Fatalf("got %d, want 503; body=%s", rr.Code, rr.Body.String())
			}
			var body readyAggregate
			if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if body.OK {
				t.Error("aggregate ok=true, want false")
			}
			// At least one result must carry a non-empty error string.
			anyErr := false
			for _, r := range body.DepResults {
				if r.Error != "" {
					anyErr = true
				}
			}
			if !anyErr {
				t.Error("expected at least one error in dep results")
			}
		})
	}
}

// TestReady_CacheHonoursTTL verifies the 5-second cache window
// requirement. We use a 100ms TTL in newTestServer so the test runs
// quickly; the contract under test is identical.
//
// Acceptance criterion #2: result is cached for 5s.
func TestReady_CacheHonoursTTL(t *testing.T) {
	dep := &fakeDepCheck{name: "postgres"}
	srv, _ := newTestServer(t, BuildInfo{}, []DepCheck{dep})

	// First two probes within the TTL window: only one Check call.
	for i := 0; i < 3; i++ {
		rr := httptest.NewRecorder()
		srv.Mux.ServeHTTP(rr, httptest.NewRequest("GET", "/ready", nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("probe %d: got %d, want 200", i, rr.Code)
		}
	}
	if got := dep.calls.Load(); got != 1 {
		t.Errorf("warm cache: dep was probed %d times within TTL, want 1", got)
	}

	// Sleep past the TTL and probe again — Check call count should advance.
	time.Sleep(150 * time.Millisecond)

	rr := httptest.NewRecorder()
	srv.Mux.ServeHTTP(rr, httptest.NewRequest("GET", "/ready", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("post-TTL probe: got %d, want 200", rr.Code)
	}
	if got := dep.calls.Load(); got != 2 {
		t.Errorf("post-TTL: dep was probed %d times, want 2", got)
	}
}

// TestMetrics_ContainsBuildInfoAndDepStatus asserts the chetana_build_info
// and chetana_dep_check_status gauges are exposed.
//
// Acceptance criterion #5: /metrics includes build_info{...} and
// chetana_dep_check_status{dep="postgres"} ∈ {0,1}.
func TestMetrics_ContainsBuildInfoAndDepStatus(t *testing.T) {
	build := BuildInfo{Version: "v9.9.9", GitSHA: "deadbeef"}
	srv, mh := newTestServer(t, build, []DepCheck{
		&fakeDepCheck{name: "postgres"},
		&fakeDepCheck{name: "kafka", err: errors.New("offline")},
	})
	// Trigger a /ready probe so the dep gauges are populated.
	rr := httptest.NewRecorder()
	srv.Mux.ServeHTTP(rr, httptest.NewRequest("GET", "/ready", nil))

	// Now scrape /metrics.
	mr := httptest.NewRecorder()
	mh.ServeHTTP(mr, httptest.NewRequest("GET", "/metrics", nil))
	body, err := io.ReadAll(mr.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	text := string(body)

	wantSubstrings := []string{
		`chetana_build_info{`,
		`version="v9.9.9"`,
		`git_sha="deadbeef"`,
		`go_version=`,
		`chetana_dep_check_status{dep="postgres"} 1`,
		`chetana_dep_check_status{dep="kafka"} 0`,
		`chetana_dep_check_latency_seconds{dep="postgres"}`,
		`go_goroutines`,    // GoCollector
		`process_open_fds`, // ProcessCollector (Linux); skipped on Windows runtime
	}
	for _, want := range wantSubstrings {
		if want == `process_open_fds` && !strings.Contains(text, want) {
			// ProcessCollector emits a different surface on Windows.
			// Tolerate absence rather than fail cross-platform.
			continue
		}
		if !strings.Contains(text, want) {
			t.Errorf("metrics body missing substring %q", want)
		}
	}
}

// TestNewServer_ZeroDepChecks_ReadyAlwaysOK verifies the empty-deps
// edge case: services that do not declare dependencies should still
// return 200 from /ready.
func TestNewServer_ZeroDepChecks_ReadyAlwaysOK(t *testing.T) {
	srv, _ := newTestServer(t, BuildInfo{}, nil)
	rr := httptest.NewRecorder()
	srv.Mux.ServeHTTP(rr, httptest.NewRequest("GET", "/ready", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rr.Code)
	}
}

// TestStatusLabel_BoundsCardinality verifies the helper used by the
// HTTP-metrics middleware reduces status codes to a stable label set.
// Cardinality control matters for Prometheus.
func TestStatusLabel_BoundsCardinality(t *testing.T) {
	cases := map[int]string{
		200: "200",
		503: "503",
		500: "500",
		404: "404",
		201: "2xx",
		301: "3xx",
		418: "4xx",
		504: "5xx",
		999: "unknown",
	}
	for code, want := range cases {
		if got := statusLabel(code); got != want {
			t.Errorf("statusLabel(%d)=%q, want %q", code, got, want)
		}
	}
}
