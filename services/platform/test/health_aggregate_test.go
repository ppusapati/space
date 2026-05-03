//go:build integration

// health_aggregate_test.go — TASK-P1-PLT-HEALTH-001 acceptance.
//
// #1: aggregator polls every registered service and Roll() yields
//      one entry per service with last-seen + status + error rate.
// #2: a 5-minute sustained failure → exactly one PagerDuty
//      incident; flap (≥3 transitions in 10 min) → single warning.

package platform_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ppusapati/space/services/platform/internal/health"
)

func newHealthPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("PLATFORM_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PLATFORM_TEST_DATABASE_URL not set — skipping integration test")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`TRUNCATE service_health, health_incidents, service_transitions RESTART IDENTITY`)
		pool.Close()
	})
	if _, err := pool.Exec(context.Background(),
		`TRUNCATE service_health, health_incidents, service_transitions RESTART IDENTITY`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return pool
}

// fakeServer returns 200 by default but flips to 500 when down=1.
func fakeServer(down *atomic.Int32) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if down.Load() == 1 {
			http.Error(w, "down", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
}

// Acceptance #1: aggregator rolls up per-service last-seen + status.
func TestHealth_AggregatorRollUp(t *testing.T) {
	pool := newHealthPool(t)
	st, err := health.NewStore(pool, time.Now)
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	agg, err := health.NewAggregator(health.AggregatorConfig{Store: st})
	if err != nil {
		t.Fatalf("agg: %v", err)
	}

	healthy := atomic.Int32{}
	failing := atomic.Int32{}
	failing.Store(1)
	healthySrv := fakeServer(&healthy)
	failingSrv := fakeServer(&failing)
	t.Cleanup(func() { healthySrv.Close(); failingSrv.Close() })

	agg.Register("svc-healthy", healthySrv.URL+"/ready")
	agg.Register("svc-failing", failingSrv.URL+"/ready")

	agg.Tick(context.Background())

	report, err := agg.Report(context.Background())
	if err != nil {
		t.Fatalf("report: %v", err)
	}
	if len(report.Services) != 2 {
		t.Fatalf("services: %d", len(report.Services))
	}
	byName := make(map[string]health.ServiceSummary)
	for _, s := range report.Services {
		byName[s.Service] = s
	}
	if byName["svc-healthy"].Status != "ok" {
		t.Errorf("healthy status: %q", byName["svc-healthy"].Status)
	}
	if byName["svc-failing"].Status != "down" {
		t.Errorf("failing status: %q", byName["svc-failing"].Status)
	}
	if byName["svc-failing"].ErrorCount == 0 {
		t.Errorf("error count should be > 0")
	}
}

// Acceptance #2 (flap path): ≥3 transitions in 10min → single warn.
func TestHealth_FlapDetectionEmitsSingleWarning(t *testing.T) {
	pool := newHealthPool(t)
	st, _ := health.NewStore(pool, time.Now)

	pager := &health.CapturingNotifier{}
	slack := &health.CapturingNotifier{}
	alerter, err := health.NewAlerter(health.AlerterConfig{
		Store:              st,
		Slack:              slack,
		Pager:              pager,
		FlapThreshold:      3,
		FlapWindow:         10 * time.Minute,
		SustainedThreshold: 5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("alerter: %v", err)
	}
	agg, _ := health.NewAggregator(health.AggregatorConfig{Store: st, Alerter: alerter})

	flapping := atomic.Int32{}
	srv := fakeServer(&flapping)
	t.Cleanup(srv.Close)
	agg.Register("svc-flap", srv.URL+"/ready")

	// Flap: ok → down → ok → down → ok → down (5 transitions).
	for i := 0; i < 6; i++ {
		flapping.Store(int32(i % 2))
		agg.Tick(context.Background())
	}

	if len(pager.Alerts) != 0 {
		t.Errorf("flap should not page: %d alerts", len(pager.Alerts))
	}
	flapSlackAlerts := 0
	for _, a := range slack.Alerts {
		if a.State == health.StateFlap {
			flapSlackAlerts++
		}
	}
	if flapSlackAlerts != 1 {
		t.Errorf("flap should emit exactly 1 slack alert; got %d", flapSlackAlerts)
	}
}

// Acceptance #2 (sustained path): 5-minute sustained failure →
// exactly one PagerDuty incident.
//
// We use a clock-injected store so we don't need to wait 5 real
// minutes.
func TestHealth_SustainedFailureEmitsSinglePage(t *testing.T) {
	pool := newHealthPool(t)
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	currentTime := now
	clock := func() time.Time { return currentTime }
	st, _ := health.NewStore(pool, clock)

	pager := &health.CapturingNotifier{}
	slack := &health.CapturingNotifier{}
	alerter, _ := health.NewAlerter(health.AlerterConfig{
		Store:              st,
		Slack:              slack,
		Pager:              pager,
		FlapThreshold:      100, // disable flap path for this test
		FlapWindow:         10 * time.Minute,
		SustainedThreshold: 5 * time.Minute,
		Now:                clock,
	})
	agg, _ := health.NewAggregator(health.AggregatorConfig{Store: st, Alerter: alerter})

	failing := atomic.Int32{}
	failing.Store(1)
	srv := fakeServer(&failing)
	t.Cleanup(srv.Close)
	agg.Register("svc-down", srv.URL+"/ready")

	// First tick — ok → down transition logged at currentTime.
	agg.Tick(context.Background())

	// Advance clock past the 5-minute sustained threshold + tick
	// again. This time the alerter sees sustained_dur >= 5m.
	currentTime = now.Add(6 * time.Minute)
	agg.Tick(context.Background())

	// Second post-threshold tick — should NOT page again
	// (idempotent OpenIncident bumps Transitions).
	currentTime = now.Add(7 * time.Minute)
	agg.Tick(context.Background())

	pageCount := 0
	for _, a := range pager.Alerts {
		if a.State == health.StateSustainedFailure {
			pageCount++
		}
	}
	if pageCount != 1 {
		t.Errorf("sustained failure should page exactly once; got %d", pageCount)
	}

	// And the open-incidents block should carry the row.
	report, _ := agg.Report(context.Background())
	openMatch := 0
	for _, inc := range report.Open {
		if inc.Service == "svc-down" && inc.State == health.StateSustainedFailure {
			openMatch++
		}
	}
	if openMatch != 1 {
		t.Errorf("expected 1 open sustained incident; got %d", openMatch)
	}
}

// Sustained-failure recovery: when the service returns to OK,
// the open incident is resolved.
func TestHealth_RecoveryResolvesIncidents(t *testing.T) {
	pool := newHealthPool(t)
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	currentTime := now
	clock := func() time.Time { return currentTime }
	st, _ := health.NewStore(pool, clock)
	alerter, _ := health.NewAlerter(health.AlerterConfig{
		Store:              st,
		FlapThreshold:      100,
		SustainedThreshold: 1 * time.Second,
		Now:                clock,
	})
	agg, _ := health.NewAggregator(health.AggregatorConfig{Store: st, Alerter: alerter})

	failing := atomic.Int32{}
	failing.Store(1)
	srv := fakeServer(&failing)
	t.Cleanup(srv.Close)
	agg.Register("svc-recover", srv.URL+"/ready")

	agg.Tick(context.Background())
	currentTime = now.Add(2 * time.Second)
	agg.Tick(context.Background())

	// Now recover.
	failing.Store(0)
	currentTime = now.Add(3 * time.Second)
	agg.Tick(context.Background())

	report, _ := agg.Report(context.Background())
	for _, inc := range report.Open {
		if inc.Service == "svc-recover" {
			t.Errorf("incident not resolved: %+v", inc)
		}
	}
}

func TestHealth_ProbeDistinguishesStatusCodes(t *testing.T) {
	pool := newHealthPool(t)
	st, _ := health.NewStore(pool, time.Now)
	agg, _ := health.NewAggregator(health.AggregatorConfig{Store: st})

	for _, tc := range []struct {
		name       string
		status     int
		wantStatus string
	}{
		{"200 → ok", http.StatusOK, "ok"},
		{"503 → down", http.StatusServiceUnavailable, "down"},
		{"429 → degraded", http.StatusTooManyRequests, "degraded"},
		{"301 → degraded", http.StatusMovedPermanently, "degraded"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()
			agg.Register("svc", srv.URL+"/ready")
			agg.Tick(context.Background())
			report, _ := agg.Report(context.Background())
			var got string
			for _, s := range report.Services {
				if s.Service == "svc" {
					got = s.Status
				}
			}
			if got != tc.wantStatus {
				t.Errorf("status: got %q want %q", got, tc.wantStatus)
			}
		})
	}
}
