// direct.go — synchronous Postgres-backed Client.
//
// DirectClient calls into the audit service's chain.Appender
// directly. Used by:
//
//   • Tests that don't want to spin up Kafka.
//   • Services co-located with audit (the v1 single-binary
//     dev posture).
//   • The audit service itself (when it consumes the Kafka topic
//     in the future, the consumer wraps this same client).
//
// Production multi-process deployments wire the KafkaClient
// (TASK-P1-AUDIT-KAFKA, future) so producer-side latency does
// not couple to the audit DB.

package audit

import (
	"context"
	"errors"
)

// DirectClient is the synchronous-INSERT implementation of Client.
type DirectClient struct {
	appender DirectAppender
	tenantID string
}

// DirectAppender is a tiny façade over the per-row write the
// audit service exposes. Production wiring passes a closure that
// calls into chain.Appender.Append; tests pass a fake.
//
// The closure receives the FULL Event (rather than ChainEvent)
// because the audit service's chain package owns the
// canonicalisation; we want a single source of truth for the
// hash, not two copies in two packages.
type DirectAppender func(ctx context.Context, e Event) error

// NewDirectClient builds a Client that emits via the supplied
// appender closure for the given tenant.
func NewDirectClient(tenantID string, appender DirectAppender) (Client, error) {
	if tenantID == "" {
		return nil, errors.New("audit: empty tenant_id")
	}
	if appender == nil {
		return nil, errors.New("audit: nil appender")
	}
	return &DirectClient{appender: appender, tenantID: tenantID}, nil
}

// Emit implements Client.
func (c *DirectClient) Emit(ctx context.Context, e Event) error {
	if e.TenantID == "" {
		e.TenantID = c.tenantID
	}
	if err := validate(&e); err != nil {
		return err
	}
	return c.appender(ctx, e)
}
