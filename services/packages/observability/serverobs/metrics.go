// metrics.go — Prometheus /metrics endpoint with the platform's standard
// per-process collectors.
//
// TASK-P0-OBS-001 (REQ-FUNC-CMN-003, REQ-NFR-OBS-001).
//
// This file owns the per-process metrics surface served on the metrics
// port (default :9090). It is deliberately distinct from
// packages/metrics — the latter exposes a write-side MetricsProvider
// abstraction that domain code records into; here we are responsible for
// the HTTP plumbing that exposes those values to the Prometheus scraper.
//
// Built-in collectors:
//
//   • chetana_build_info{version,git_sha,go_version}     gauge=1
//   • chetana_dep_check_status{dep="<name>"}             gauge in {0,1}
//   • chetana_dep_check_latency_seconds{dep="<name>"}    gauge
//   • plus the standard process and Go runtime collectors
//
// The dep gauges are populated lazily from the readyChecker snapshot on
// each scrape — no separate goroutine, so the metric values reflect the
// most recent /ready evaluation (within readyCacheTTL).

package serverobs

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// metricsRegistry is the metrics state owned by a single server instance.
// We deliberately do NOT use prometheus.DefaultRegisterer so tests and
// multiple servers can coexist in one process.
type metricsRegistry struct {
	registry *prometheus.Registry

	buildInfo       *prometheus.GaugeVec
	depStatus       *prometheus.GaugeVec
	depLatency      *prometheus.GaugeVec
	rpcDuration     *prometheus.HistogramVec
	rpcTotal        *prometheus.CounterVec
	httpDuration    *prometheus.HistogramVec
	httpTotal       *prometheus.CounterVec
}

// newMetricsRegistry constructs the per-server registry, registers the
// default process + Go runtime collectors, and seeds chetana_build_info
// with the supplied identity.
func newMetricsRegistry(b BuildInfo) *metricsRegistry {
	reg := prometheus.NewRegistry()

	mr := &metricsRegistry{
		registry: reg,
		buildInfo: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chetana_build_info",
			Help: "Build identity. Always 1; labels carry the version/sha/go_version.",
		}, []string{"version", "git_sha", "go_version"}),
		depStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chetana_dep_check_status",
			Help: "Dependency check status: 1=ok, 0=failing.",
		}, []string{"dep"}),
		depLatency: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chetana_dep_check_latency_seconds",
			Help: "Last observed latency of a dependency check.",
		}, []string{"dep"}),
		rpcDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "chetana_rpc_duration_seconds",
			Help:    "ConnectRPC handler duration.",
			Buckets: prometheus.DefBuckets,
		}, []string{"procedure", "code"}),
		rpcTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "chetana_rpc_requests_total",
			Help: "ConnectRPC handler invocation count.",
		}, []string{"procedure", "code"}),
		httpDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "chetana_http_request_duration_seconds",
			Help:    "Non-RPC HTTP handler duration (health, ready, metrics).",
			Buckets: prometheus.DefBuckets,
		}, []string{"path", "method", "status"}),
		httpTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "chetana_http_requests_total",
			Help: "Non-RPC HTTP handler invocation count.",
		}, []string{"path", "method", "status"}),
	}

	reg.MustRegister(
		mr.buildInfo, mr.depStatus, mr.depLatency,
		mr.rpcDuration, mr.rpcTotal,
		mr.httpDuration, mr.httpTotal,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	// Set chetana_build_info to 1; the labels carry the identity.
	mr.buildInfo.WithLabelValues(b.Version, b.GitSHA, resolveGoVersion()).Set(1)
	return mr
}

// updateDepGauges projects the latest readiness aggregate into the dep
// gauges. Called by the metrics handler on each scrape so the snapshot
// is always at most readyCacheTTL old.
func (m *metricsRegistry) updateDepGauges(snap readyAggregate) {
	for _, r := range snap.results {
		v := 0.0
		if r.ok {
			v = 1.0
		}
		m.depStatus.WithLabelValues(r.name).Set(v)
		m.depLatency.WithLabelValues(r.name).Set(r.latency.Seconds())
	}
}

// observeHTTP records a non-RPC HTTP request. Called by the wrapping
// middleware around /health, /ready, and /metrics itself.
func (m *metricsRegistry) observeHTTP(path, method, status string, dur time.Duration) {
	m.httpDuration.WithLabelValues(path, method, status).Observe(dur.Seconds())
	m.httpTotal.WithLabelValues(path, method, status).Inc()
}

// metricsHandler builds the http.Handler exposing the registry. We
// project the latest readyChecker snapshot into the dep gauges before
// each scrape so a separate poller is not required.
func metricsHandler(m *metricsRegistry, rc *readyChecker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rc != nil {
			m.updateDepGauges(rc.snapshot())
		}
		promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
			Registry: m.registry,
		}).ServeHTTP(w, r)
	})
}

// httpMetricsMiddleware records duration and status for every wrapped
// HTTP request. Use sparingly — we wrap /health, /ready, /metrics here
// so operators can see internal-endpoint pressure without turning on
// general HTTP instrumentation.
func httpMetricsMiddleware(m *metricsRegistry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()
			next.ServeHTTP(rec, r)
			m.observeHTTP(r.URL.Path, r.Method, statusLabel(rec.status), time.Since(start))
		})
	}
}

// statusRecorder captures the response status code so the middleware
// can record it in metrics.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// statusLabel reduces the status code cardinality. Class buckets ("2xx",
// "5xx") would also work but operators usually want the exact code on
// internal endpoints.
func statusLabel(code int) string {
	switch code {
	case http.StatusOK:
		return "200"
	case http.StatusServiceUnavailable:
		return "503"
	case http.StatusInternalServerError:
		return "500"
	case http.StatusNotFound:
		return "404"
	}
	// Two-character bucket for everything else keeps cardinality bounded.
	switch {
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 300 && code < 400:
		return "3xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500 && code < 600:
		return "5xx"
	}
	return "unknown"
}
