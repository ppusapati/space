// Package audit is the typed client every chetana service uses to
// emit audit events.
//
// → REQ-FUNC-PLT-AUDIT-001 / 002 / 006; design.md §4.2.
//
// Two transport choices:
//
//   • Direct (DirectClient) — synchronous Postgres INSERT through
//     the audit service's chain.Appender. Simplest; suitable for
//     in-process or co-located services. Used by tests + the IAM
//     service in v1 before Kafka is wired.
//
//   • Kafka (KafkaClient — future) — fire-and-forget produce to
//     `audit.events.v1`; the audit service's consumer drains the
//     topic and runs the chain.Appender. The wire-format Event
//     struct here is what the producer marshals.
//
// Both implementations satisfy the same Client interface so the
// rest of the platform's interceptors don't care which transport
// is wired at boot.

package audit

import (
	"context"
	"errors"
	"time"
)

// Event is the wire-format audit event the platform's typed
// interceptor + the IAM-side audit emitters produce. The shape
// stays identical to the chain.Event the audit service hashes —
// the audit service constructs a chain.Event from this struct
// 1:1, so a future field addition lands in BOTH places (the
// hash will then change for new events but not for old ones).
type Event struct {
	TenantID         string
	OccurredAt       time.Time
	ActorUserID      string
	ActorSessionID   string
	ActorClientIP    string
	ActorUserAgent   string
	Action           string // canonical {module}.{resource}.{action}
	Resource         string // optional resource id the action targets
	Decision         string // "allow" | "deny" | "ok" | "fail" | "info"
	Reason           string // matched policy id OR error reason text
	MatchedPolicyID  string
	Procedure        string // Connect procedure name when wrapping an RPC
	Classification   string // public | internal | restricted | cui | itar
	Metadata         map[string]string
}

// Client is the surface the per-service interceptor + ad-hoc
// emitters depend on. Implementations MUST be safe for concurrent
// use.
type Client interface {
	Emit(ctx context.Context, event Event) error
}

// NopClient is a no-op Client useful for unit tests in services
// that don't care about the audit chain.
type NopClient struct{}

// Emit implements Client.
func (NopClient) Emit(_ context.Context, _ Event) error { return nil }

// validate normalises the supplied Event in place + checks the
// per-field invariants. Hot-path: every Emit call goes through
// this so it stays cheap (no allocations beyond the timestamp).
func validate(e *Event) error {
	if e.Action == "" {
		return errors.New("audit: empty action")
	}
	if e.Decision == "" {
		return errors.New("audit: empty decision")
	}
	if e.OccurredAt.IsZero() {
		e.OccurredAt = time.Now().UTC()
	} else {
		e.OccurredAt = e.OccurredAt.UTC()
	}
	if e.Classification == "" {
		e.Classification = "cui"
	}
	switch e.Decision {
	case "allow", "deny", "ok", "fail", "info":
		// ok
	default:
		return errors.New("audit: decision must be allow|deny|ok|fail|info")
	}
	switch e.Classification {
	case "public", "internal", "restricted", "cui", "itar":
		// ok
	default:
		return errors.New("audit: classification must be public|internal|restricted|cui|itar")
	}
	return nil
}
