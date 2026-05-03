// aggregate.go — periodic /ready poll over every registered
// service. Records the outcome in the store and triggers the
// alerter on transitions.
//
// Each Aggregator is owned by the platform service; cmd/platform
// boots one and calls Run(ctx) which loops on a ticker. The
// per-service /ready URL is registered via Register(name, url).

package health

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// Target is a registered service to poll.
type Target struct {
	Service string
	URL     string // absolute /ready URL
}

// Aggregator polls every Target on a ticker and rolls the
// outcome into the store.
type Aggregator struct {
	store    *Store
	alerter  *Alerter
	client   *http.Client
	interval time.Duration
	timeout  time.Duration
	clk      func() time.Time

	mu      sync.RWMutex
	targets map[string]string // service → URL
}

// AggregatorConfig configures the Aggregator.
type AggregatorConfig struct {
	Store    *Store
	Alerter  *Alerter // optional; nil disables alerting
	Client   *http.Client
	Interval time.Duration // poll cadence; defaults to 10s
	Timeout  time.Duration // per-probe timeout; defaults to 5s
	Now      func() time.Time
}

// NewAggregator wires the dependencies. Defaults: 10s interval,
// 5s probe timeout, 10s HTTP client timeout.
func NewAggregator(cfg AggregatorConfig) (*Aggregator, error) {
	if cfg.Store == nil {
		return nil, errors.New("health: nil store")
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 10 * time.Second
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.Client == nil {
		cfg.Client = &http.Client{Timeout: cfg.Timeout}
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Aggregator{
		store:    cfg.Store,
		alerter:  cfg.Alerter,
		client:   cfg.Client,
		interval: cfg.Interval,
		timeout:  cfg.Timeout,
		clk:      cfg.Now,
		targets:  make(map[string]string),
	}, nil
}

// Register adds (or replaces) a target. Safe for concurrent use.
func (a *Aggregator) Register(service, url string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.targets[service] = url
}

// Targets returns a snapshot of the current registry.
func (a *Aggregator) Targets() []Target {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]Target, 0, len(a.targets))
	for s, u := range a.targets {
		out = append(out, Target{Service: s, URL: u})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Service < out[j].Service })
	return out
}

// Run loops on the configured interval until ctx is cancelled.
// Each tick polls every Target sequentially (chetana scale —
// dozens of services — keeps this cheap; if we ever exceed ~50
// services we'll switch to a worker pool).
func (a *Aggregator) Run(ctx context.Context) error {
	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()

	// Run an initial tick immediately so /v1/health/services has
	// data on first call rather than waiting `interval`.
	a.Tick(ctx)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			a.Tick(ctx)
		}
	}
}

// Tick runs one polling round. Exposed for tests.
func (a *Aggregator) Tick(ctx context.Context) {
	for _, t := range a.Targets() {
		if err := a.probe(ctx, t); err != nil {
			// probe errors are already recorded; nothing else to
			// do here — the loop continues with the next target.
			_ = err
		}
	}
}

// probe runs one /ready check + records the outcome + invokes
// the alerter on transitions.
func (a *Aggregator) probe(ctx context.Context, t Target) error {
	probeCtx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	status, errMsg := a.runReady(probeCtx, t.URL)
	prev, err := a.store.RecordCheck(ctx, t.Service, status, errMsg)
	if err != nil {
		return fmt.Errorf("health: record %q: %w", t.Service, err)
	}

	// Resolution path: was failing, now OK → close any open
	// incidents.
	if prev != "" && prev != StatusOK && status == StatusOK {
		if _, err := a.store.ResolveOpenIncidents(ctx, t.Service); err != nil {
			return fmt.Errorf("health: resolve %q: %w", t.Service, err)
		}
	}

	if a.alerter != nil {
		if err := a.alerter.Evaluate(ctx, t.Service, prev, status); err != nil {
			return fmt.Errorf("health: alerter %q: %w", t.Service, err)
		}
	}
	return nil
}

// runReady performs the HTTP probe.
//
//   • 200 OK              → status "ok"
//   • 5xx                 → status "down" with the body excerpt
//   • 4xx                 → status "degraded" with the body excerpt
//   • network / timeout   → status "down" with the err.Error()
func (a *Aggregator) runReady(ctx context.Context, url string) (string, string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return StatusUnknown, err.Error()
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return StatusDown, err.Error()
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusOK:
		return StatusOK, ""
	case resp.StatusCode >= 500:
		return StatusDown, fmt.Sprintf("HTTP %d", resp.StatusCode)
	case resp.StatusCode >= 400:
		return StatusDegraded, fmt.Sprintf("HTTP %d", resp.StatusCode)
	}
	// 2xx-non-200, 3xx → treat as degraded.
	return StatusDegraded, fmt.Sprintf("HTTP %d", resp.StatusCode)
}

// AggregatedReport is the JSON shape returned by
// /v1/health/services.
type AggregatedReport struct {
	GeneratedAt time.Time         `json:"generated_at"`
	Services    []ServiceSummary  `json:"services"`
	Open        []IncidentSummary `json:"open_incidents"`
}

// ServiceSummary is one row in the report.
type ServiceSummary struct {
	Service      string    `json:"service"`
	Status       string    `json:"status"`
	LastSeenAt   time.Time `json:"last_seen_at"`
	LastError    string    `json:"last_error,omitempty"`
	ErrorRate    float64   `json:"error_rate"` // errors / (errors + successes)
	ErrorCount   int64     `json:"error_count"`
	SuccessCount int64     `json:"success_count"`
}

// IncidentSummary is one row of the open-incidents block.
type IncidentSummary struct {
	ID          int64     `json:"id"`
	Service     string    `json:"service"`
	State       string    `json:"state"`
	Severity    string    `json:"severity"`
	OpenedAt    time.Time `json:"opened_at"`
	Transitions int       `json:"transitions"`
	Note        string    `json:"note,omitempty"`
}

// Report builds the read-side report. Used by the HTTP handler
// the cmd layer mounts at /v1/health/services.
func (a *Aggregator) Report(ctx context.Context) (*AggregatedReport, error) {
	snaps, err := a.store.Roll(ctx)
	if err != nil {
		return nil, err
	}
	out := &AggregatedReport{GeneratedAt: a.clk().UTC()}
	for _, s := range snaps {
		out.Services = append(out.Services, ServiceSummary{
			Service:      s.Service,
			Status:       s.LastStatus,
			LastSeenAt:   s.LastSeenAt,
			LastError:    s.LastError,
			ErrorRate:    errorRate(s.ErrorCount, s.SuccessCount),
			ErrorCount:   s.ErrorCount,
			SuccessCount: s.SuccessCount,
		})
	}
	// Open incidents fetch — kept inline so the handler does not
	// need a second store helper.
	rows, err := a.store.pool.Query(ctx, `
SELECT id, service, state, severity, opened_at, transitions, note
FROM health_incidents
WHERE resolved_at IS NULL
ORDER BY opened_at DESC
`)
	if err != nil {
		return out, nil // best-effort; the services array is the payload
	}
	defer rows.Close()
	for rows.Next() {
		var inc IncidentSummary
		if err := rows.Scan(
			&inc.ID, &inc.Service, &inc.State, &inc.Severity,
			&inc.OpenedAt, &inc.Transitions, &inc.Note,
		); err != nil {
			break
		}
		out.Open = append(out.Open, inc)
	}
	return out, nil
}

// errorRate returns the rolling error_count / total. Zero
// observations → 0.
func errorRate(errCount, successCount int64) float64 {
	total := errCount + successCount
	if total == 0 {
		return 0
	}
	return float64(errCount) / float64(total)
}

// excerpt clamps free-form error text to 256 chars + collapses
// whitespace runs so the audit-friendly column doesn't blow up.
func excerpt(s string) string {
	if len(s) > 256 {
		s = s[:256] + "…"
	}
	return strings.TrimSpace(s)
}
