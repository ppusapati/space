// deps.go — health/readiness dependency checks.
//
// TASK-P0-OBS-001 (REQ-FUNC-CMN-002).
//
// A DepCheck answers a single question: "is this upstream dependency
// reachable right now?" Implementations are passed to NewServer in the
// Config.DepChecks slice; the /ready handler aggregates them with a
// 5-second cache so a flapping upstream does not drown the server in
// connection storms.
//
// Built-in implementations cover Postgres (pg_isready-equivalent),
// Kafka (broker metadata), and Redis (PING). Services may add their
// own by implementing the DepCheck interface.

package serverobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// DepCheck represents a single upstream dependency probe. The Name is
// used to label the /ready response and the dep_check_status metric
// (so it MUST be a stable, low-cardinality string — "postgres", "kafka",
// "redis", "iam"). Check is invoked at most once per cache window with
// a per-call timeout supplied by the caller via ctx.
type DepCheck interface {
	Name() string
	Check(ctx context.Context) error
}

// PostgresCheck reports the dependency healthy when the supplied pool
// answers Ping within the call's deadline.
type PostgresCheck struct {
	Pool *pgxpool.Pool
}

// Name implements DepCheck.
func (c PostgresCheck) Name() string { return "postgres" }

// Check implements DepCheck.
func (c PostgresCheck) Check(ctx context.Context) error {
	if c.Pool == nil {
		return errors.New("postgres: nil pool")
	}
	return c.Pool.Ping(ctx)
}

// KafkaCheck reports the dependency healthy when the supplied client
// returns a non-empty broker list (a stand-in for "cluster reachable").
// We use the existing sarama client without dialing a fresh connection
// per probe — that's cheap and avoids file-descriptor churn under load.
type KafkaCheck struct {
	Client sarama.Client
}

// Name implements DepCheck.
func (c KafkaCheck) Name() string { return "kafka" }

// Check implements DepCheck.
func (c KafkaCheck) Check(ctx context.Context) error {
	if c.Client == nil {
		return errors.New("kafka: nil client")
	}
	// sarama.Client lacks a context-aware probe; we satisfy the
	// caller's deadline by running the probe on a goroutine and
	// racing it against ctx.Done().
	type result struct {
		brokers []*sarama.Broker
		err     error
	}
	resCh := make(chan result, 1)
	go func() {
		// RefreshMetadata keeps the cluster view fresh and exercises
		// the connection. A nil topic argument refreshes all topics.
		err := c.Client.RefreshMetadata()
		if err != nil {
			resCh <- result{err: fmt.Errorf("refresh metadata: %w", err)}
			return
		}
		brokers := c.Client.Brokers()
		if len(brokers) == 0 {
			resCh <- result{err: errors.New("no brokers known")}
			return
		}
		resCh <- result{brokers: brokers}
	}()
	select {
	case <-ctx.Done():
		return fmt.Errorf("kafka check: %w", ctx.Err())
	case r := <-resCh:
		return r.err
	}
}

// RedisCheck reports the dependency healthy when PING returns within
// the call's deadline.
type RedisCheck struct {
	Client redis.UniversalClient
}

// Name implements DepCheck.
func (c RedisCheck) Name() string { return "redis" }

// Check implements DepCheck.
func (c RedisCheck) Check(ctx context.Context) error {
	if c.Client == nil {
		return errors.New("redis: nil client")
	}
	return c.Client.Ping(ctx).Err()
}

// FuncDepCheck adapts a plain func into a DepCheck. Useful for ad-hoc
// service-specific probes (e.g. "auth-svc reachable") without defining
// a struct.
type FuncDepCheck struct {
	NameValue string
	Probe     func(ctx context.Context) error
}

// Name implements DepCheck.
func (f FuncDepCheck) Name() string { return f.NameValue }

// Check implements DepCheck.
func (f FuncDepCheck) Check(ctx context.Context) error {
	if f.Probe == nil {
		return errors.New("FuncDepCheck: nil probe")
	}
	return f.Probe(ctx)
}

// depCheckResult is the outcome of one DepCheck.Check invocation,
// recorded in the readiness cache.
type depCheckResult struct {
	name     string
	ok       bool
	err      string // empty when ok
	latency  time.Duration
	checkedAt time.Time
}
