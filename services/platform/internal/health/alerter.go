// alerter.go — flap detector + sustained-failure detector.
//
// Detection rules (REQ-FUNC-CMN-004 acceptance #2):
//
//   • Sustained failure: a service has been continuously non-OK
//     for ≥ SustainedThreshold (default 5 min). Severity = "page".
//
//   • Flap: ≥ FlapThreshold (default 3) state transitions inside
//     FlapWindow (default 10 min). Severity = "warn".
//
// Both detection branches call OpenIncident which is idempotent
// on (service, state) WHERE resolved_at IS NULL — so repeated
// ticks against the same condition only update transitions/note,
// they do NOT page repeatedly.
//
// Routing is via injectable Notifier interfaces. The chetana
// notify service (TASK-P1-NOTIFY-001) provides the real impls;
// NopNotifier ships in-package so the alerter is wireable today.

package health

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Notifier is the surface the alerter calls to deliver alerts.
// Implementations route to Slack / email / PagerDuty respectively.
type Notifier interface {
	Notify(ctx context.Context, alert Alert) error
}

// Alert is the per-incident payload routed to a Notifier.
type Alert struct {
	Service     string
	State       string // "flap" | "sustained_failure"
	Severity    string // "warn" | "page"
	OpenedAt    time.Time
	Transitions int
	Note        string
}

// NopNotifier is a no-op Notifier useful for tests + for the
// initial dev posture before NOTIFY-001 ships.
type NopNotifier struct{}

// Notify implements Notifier.
func (NopNotifier) Notify(_ context.Context, _ Alert) error { return nil }

// CapturingNotifier records every alert. Useful for tests.
type CapturingNotifier struct {
	mu     sync.Mutex
	Alerts []Alert
}

// Notify implements Notifier.
func (c *CapturingNotifier) Notify(_ context.Context, a Alert) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Alerts = append(c.Alerts, a)
	return nil
}

// AlerterConfig configures the alerter.
type AlerterConfig struct {
	Store *Store

	// Notifiers — at least one of Slack/Email/Pager should be
	// non-nil in production.
	Slack Notifier
	Email Notifier
	Pager Notifier

	// FlapThreshold is the number of state transitions that
	// trips the flap detector inside FlapWindow.
	FlapThreshold int
	FlapWindow    time.Duration

	// SustainedThreshold is the duration of continuous non-OK
	// status that trips the page.
	SustainedThreshold time.Duration

	Now func() time.Time
}

// Alerter wires the detectors + the routers.
type Alerter struct {
	cfg AlerterConfig
}

// NewAlerter validates the config and returns an Alerter.
// Defaults: flap = 3 transitions / 10 min; sustained = 5 min.
// Missing notifiers default to NopNotifier so the path doesn't
// nil-panic before NOTIFY-001 ships.
func NewAlerter(cfg AlerterConfig) (*Alerter, error) {
	if cfg.Store == nil {
		return nil, errors.New("health: nil store")
	}
	if cfg.FlapThreshold <= 0 {
		cfg.FlapThreshold = 3
	}
	if cfg.FlapWindow <= 0 {
		cfg.FlapWindow = 10 * time.Minute
	}
	if cfg.SustainedThreshold <= 0 {
		cfg.SustainedThreshold = 5 * time.Minute
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.Slack == nil {
		cfg.Slack = NopNotifier{}
	}
	if cfg.Email == nil {
		cfg.Email = NopNotifier{}
	}
	if cfg.Pager == nil {
		cfg.Pager = NopNotifier{}
	}
	return &Alerter{cfg: cfg}, nil
}

// Evaluate runs the detection rules for `service` after a probe
// transitioned from `prevStatus` → `currStatus`. Called by the
// aggregator after every probe.
//
// Returns nil on success; non-nil on store / notifier errors
// (the caller logs but does NOT stop the aggregation loop).
func (a *Alerter) Evaluate(ctx context.Context, service, prevStatus, currStatus string) error {
	// 1. Flap detection — count transitions in the window.
	transitions, err := a.cfg.Store.CountTransitionsSince(ctx, service, a.cfg.FlapWindow)
	if err != nil {
		return err
	}
	if transitions >= a.cfg.FlapThreshold {
		note := fmt.Sprintf("%d transitions in last %s", transitions, a.cfg.FlapWindow)
		inc, err := a.cfg.Store.OpenIncident(ctx, service, StateFlap, SeverityWarn, note, transitions)
		if err != nil {
			return err
		}
		// Route only on the FIRST detection (transitions==threshold).
		// Repeated bumps just update the note via OpenIncident.
		if inc.Transitions == a.cfg.FlapThreshold {
			a.route(ctx, Alert{
				Service:     service,
				State:       StateFlap,
				Severity:    SeverityWarn,
				OpenedAt:    inc.OpenedAt,
				Transitions: transitions,
				Note:        note,
			})
		}
	}

	// 2. Sustained-failure detection — only check when the
	// service is currently non-OK.
	if currStatus != StatusOK {
		dur, ok, err := a.cfg.Store.SustainedSince(ctx, service)
		if err != nil {
			return err
		}
		if ok && dur >= a.cfg.SustainedThreshold {
			note := fmt.Sprintf("%s sustained %s", currStatus, dur.Round(time.Second))
			inc, err := a.cfg.Store.OpenIncident(ctx, service, StateSustainedFailure, SeverityPage, note, 1)
			if err != nil {
				return err
			}
			// Page only once per incident — Transitions stays at
			// 1 for the open incident's lifetime; the second
			// detection tick increments it to 2 and we suppress
			// the duplicate page.
			if inc.Transitions == 1 {
				a.route(ctx, Alert{
					Service:     service,
					State:       StateSustainedFailure,
					Severity:    SeverityPage,
					OpenedAt:    inc.OpenedAt,
					Transitions: 1,
					Note:        note,
				})
			}
			// Bump transitions so the second detection tick is
			// idempotent (no double-page).
			_, _ = a.cfg.Store.OpenIncident(ctx, service, StateSustainedFailure, SeverityPage,
				note, inc.Transitions+1)
		}
	}

	// suppress unused-arg lint
	_ = prevStatus
	_ = strings.TrimSpace
	return nil
}

// route dispatches one Alert to the appropriate notifiers based
// on severity.
func (a *Alerter) route(ctx context.Context, alert Alert) {
	// Best-effort — a notifier failure must NOT stop the others.
	if alert.Severity == SeverityPage {
		_ = a.cfg.Pager.Notify(ctx, alert)
	}
	_ = a.cfg.Slack.Notify(ctx, alert)
	_ = a.cfg.Email.Notify(ctx, alert)
}
