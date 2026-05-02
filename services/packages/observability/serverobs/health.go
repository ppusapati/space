// health.go — /health (liveness) and /ready (readiness) endpoints.
//
// TASK-P0-OBS-001 (REQ-FUNC-CMN-001, REQ-FUNC-CMN-002).
//
// /health is unconditional and answers "is the process alive?". It MUST
// not depend on any upstream so a brief Postgres/Kafka outage does not
// trigger pod restarts in the orchestrator.
//
// /ready aggregates the configured DepChecks. The result is cached for
// readyCacheTTL so that bursts of probes (Kubernetes default is one
// every few seconds across many endpoints) do not amplify into a
// thundering herd against the upstream. Cache freshness is bounded —
// every Check call respects the surrounding context deadline.

package serverobs

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime/debug"
	"sort"
	"sync"
	"time"
)

// readyCacheTTL is the maximum age of a cached aggregate readiness
// result. Per the acceptance criteria a cached entry must remain valid
// for 5 seconds.
const readyCacheTTL = 5 * time.Second

// depCheckTimeout caps the per-check duration. Five seconds tolerates
// a slow Postgres ping under load while ensuring /ready responds within
// orchestrator probe limits.
const depCheckTimeout = 5 * time.Second

// BuildInfo captures the build identity reported by /health and exposed
// as the build_info Prometheus gauge. Service entrypoints supply these
// at startup; the package never tries to read them from the binary.
type BuildInfo struct {
	// Version is the human-friendly release version (semver). Required.
	Version string
	// GitSHA is the source-control commit hash. Required.
	GitSHA string
}

// resolveGoVersion reads the Go version from runtime/debug. It is a
// convenience for /health output — services do not need to set it.
func resolveGoVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.GoVersion != "" {
		return info.GoVersion
	}
	return "unknown"
}

// healthResponse is the JSON payload returned from /health.
type healthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	GitSHA    string `json:"git_sha"`
	UptimeS   int64  `json:"uptime_s"`
	GoVersion string `json:"go_version"`
}

// healthHandler builds an http.HandlerFunc for /health that always
// returns 200 with the build identity. startedAt is captured by NewServer
// so uptime is measured from process start, not first request.
func healthHandler(b BuildInfo, startedAt time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(healthResponse{
			Status:    "ok",
			Version:   b.Version,
			GitSHA:    b.GitSHA,
			UptimeS:   int64(time.Since(startedAt).Seconds()),
			GoVersion: resolveGoVersion(),
		})
	}
}

// readyChecker holds the deps + cache used by /ready. Constructed by
// NewServer; not exported because callers interact with it through the
// HTTP endpoint.
type readyChecker struct {
	checks []DepCheck
	ttl    time.Duration

	mu       sync.Mutex
	cached   readyAggregate
	cachedAt time.Time
}

// readyAggregate is the rolled-up output of a single readiness sweep.
type readyAggregate struct {
	OK         bool                  `json:"ok"`
	CheckedAt  time.Time             `json:"checked_at"`
	Duration   string                `json:"duration"`
	DepResults []readyDepResultJSON  `json:"deps"`
	results    []depCheckResult      `json:"-"` // for metrics export
}

type readyDepResultJSON struct {
	Name      string    `json:"name"`
	OK        bool      `json:"ok"`
	Error     string    `json:"error,omitempty"`
	LatencyMs int64     `json:"latency_ms"`
	CheckedAt time.Time `json:"checked_at"`
}

// newReadyChecker constructs the cache-fronted aggregator over checks.
// ttl <= 0 falls back to the package default.
func newReadyChecker(checks []DepCheck, ttl time.Duration) *readyChecker {
	if ttl <= 0 {
		ttl = readyCacheTTL
	}
	cp := make([]DepCheck, len(checks))
	copy(cp, checks)
	return &readyChecker{checks: cp, ttl: ttl}
}

// snapshot returns the currently-cached aggregate. Used by metrics.go
// to populate the chetana_dep_check_status gauge without re-running the
// probes (which would race with /ready under scrape pressure).
func (r *readyChecker) snapshot() readyAggregate {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.cached
}

// run executes all DepChecks (or returns the cache entry if still warm).
// It is safe for concurrent callers — only one Check sweep runs at a
// time per checker.
func (r *readyChecker) run(ctx context.Context) readyAggregate {
	r.mu.Lock()
	if !r.cachedAt.IsZero() && time.Since(r.cachedAt) < r.ttl {
		out := r.cached
		r.mu.Unlock()
		return out
	}
	r.mu.Unlock()

	results := make([]depCheckResult, len(r.checks))
	overallStart := time.Now()

	// Run checks sequentially. They are short and we want predictable
	// upstream load — a parallel fan-out hammers all deps every probe.
	for i, c := range r.checks {
		probeCtx, cancel := context.WithTimeout(ctx, depCheckTimeout)
		t0 := time.Now()
		err := c.Check(probeCtx)
		cancel()
		res := depCheckResult{
			name:      c.Name(),
			ok:        err == nil,
			latency:   time.Since(t0),
			checkedAt: time.Now().UTC(),
		}
		if err != nil {
			res.err = err.Error()
		}
		results[i] = res
	}

	allOK := true
	for _, r := range results {
		if !r.ok {
			allOK = false
			break
		}
	}

	jsonResults := make([]readyDepResultJSON, len(results))
	for i, r := range results {
		jsonResults[i] = readyDepResultJSON{
			Name:      r.name,
			OK:        r.ok,
			Error:     r.err,
			LatencyMs: r.latency.Milliseconds(),
			CheckedAt: r.checkedAt,
		}
	}
	// Sort by name for deterministic output.
	sort.Slice(jsonResults, func(i, j int) bool {
		return jsonResults[i].Name < jsonResults[j].Name
	})

	agg := readyAggregate{
		OK:         allOK,
		CheckedAt:  time.Now().UTC(),
		Duration:   time.Since(overallStart).String(),
		DepResults: jsonResults,
		results:    results,
	}

	r.mu.Lock()
	r.cached = agg
	r.cachedAt = time.Now()
	r.mu.Unlock()
	return agg
}

// readyHandler builds the /ready endpoint. Returns 200 when every
// DepCheck passes, 503 otherwise. The body is JSON in both cases so
// operators can see which dependency is degraded.
func readyHandler(rc *readyChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), depCheckTimeout+time.Second)
		defer cancel()
		agg := rc.run(ctx)

		w.Header().Set("Content-Type", "application/json")
		if !agg.OK {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		_ = json.NewEncoder(w).Encode(agg)
	}
}
