// Package policy hydrates + serves the chetana RBAC + ABAC policy
// set every service consumes via authz/v1.
//
// → REQ-FUNC-PLT-AUTHZ-001..004; design.md §4.1.2.
//
// The loader is the SOURCE OF TRUTH on the IAM service:
//
//   • LoadFromDB pulls the latest policies + role grants from the
//     `policies` and `roles` / `role_permissions` / `user_roles`
//     tables (created by migration 0008).
//   • The loader publishes an *atomic.Pointer[authzv1.PolicySet]
//     so downstream interceptors see consistent snapshots even
//     while a reload is in flight.
//   • Hot-reload is driven by Reload(); cmd/iam wires a periodic
//     ticker (or a NOTIFY/LISTEN trigger once that lands).
//
// Service-side consumers (NOT the IAM service itself) typically
// fetch the same set via the IAM /v1/authz/policies RPC and feed
// the bytes into authzv1.LoadPoliciesYAML — but those services
// can also build policies in-memory at boot from a static YAML
// file when there's no IAM dependency at request time.

package policy

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	authzv1 "p9e.in/chetana/packages/authz/v1"
)

// Loader publishes the active PolicySet. Construct with NewLoader.
type Loader struct {
	pool    *pgxpool.Pool
	current atomic.Pointer[authzv1.PolicySet]
	clk     func() time.Time
}

// NewLoader wraps a pool. clock=nil → time.Now.
func NewLoader(pool *pgxpool.Pool, clock func() time.Time) (*Loader, error) {
	if pool == nil {
		return nil, errors.New("policy: nil pool")
	}
	if clock == nil {
		clock = time.Now
	}
	return &Loader{pool: pool, clk: clock}, nil
}

// Snapshot returns the most recent PolicySet. The returned set
// is immutable — a Reload() that completes after this call will
// not affect already-returned snapshots, so a single Decide call
// always sees a consistent view.
//
// Returns nil before the first successful Reload.
func (l *Loader) Snapshot() *authzv1.PolicySet {
	return l.current.Load()
}

// Reload pulls the latest policies from the DB and publishes a
// new snapshot. Atomic from the consumer's perspective.
func (l *Loader) Reload(ctx context.Context) error {
	policies, err := l.fetchPolicies(ctx)
	if err != nil {
		return fmt.Errorf("policy: reload: %w", err)
	}
	set, err := authzv1.NewPolicySet(policies)
	if err != nil {
		return fmt.Errorf("policy: build set: %w", err)
	}
	l.current.Store(set)
	return nil
}

// fetchPolicies pulls every active row of the `policies` table
// and projects it into the authzv1.Policy shape. Disabled rows
// are skipped.
func (l *Loader) fetchPolicies(ctx context.Context) ([]authzv1.Policy, error) {
	const q = `
SELECT id, COALESCE(description, ''), effect, COALESCE(priority, 0),
       permission, roles, COALESCE(min_clearance, ''),
       COALESCE(require_us_person, false),
       COALESCE(tenant, ''),
       COALESCE(notes, '')
FROM policies
WHERE NOT disabled
ORDER BY priority DESC, id ASC
`
	rows, err := l.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []authzv1.Policy
	for rows.Next() {
		var p authzv1.Policy
		var effect string
		if err := rows.Scan(
			&p.ID, &p.Description, &effect, &p.Priority,
			&p.Permission, &p.Roles, &p.MinClearance,
			&p.RequireUSPerson, &p.Tenant, &p.Notes,
		); err != nil {
			return nil, err
		}
		p.Effect = authzv1.Effect(effect)
		out = append(out, p)
	}
	return out, rows.Err()
}

// LoadStatic builds a PolicySet directly from an in-memory YAML
// document. Useful for the early dev posture before the IAM
// `policies` table is populated.
func LoadStatic(yaml []byte) (*authzv1.PolicySet, error) {
	return authzv1.LoadPoliciesYAML(yaml)
}

// PrimeFromYAML hydrates the loader's snapshot from a YAML
// document so service interceptors that boot before the first
// Reload() still see a usable set.
func (l *Loader) PrimeFromYAML(yaml []byte) error {
	set, err := authzv1.LoadPoliciesYAML(yaml)
	if err != nil {
		return err
	}
	l.current.Store(set)
	return nil
}
