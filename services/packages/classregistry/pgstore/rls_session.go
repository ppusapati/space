// Package pgstore — RLS session helper.
//
// All Store operations must run with `app.tenant_id` set so the FORCE RLS
// policy on classregistry.class_entities passes when the application
// connects as a non-superuser role (chetana_app, NOSUPERUSER NOBYPASSRLS).
//
// Without this, every read returns 0 rows under chetana_app — the WHERE
// clause `tenant_id = $1` in app code passes, but the policy
// `tenant_id = current_setting('app.tenant_id', true)::CHAR(26)` fails
// because the setting is unset.
//
// withTenantTx acquires a connection, opens a transaction, sets
// `app.tenant_id` LOCAL to that transaction, calls fn, and commits.
// Read paths get a slight overhead (one extra round-trip to SET, plus
// BEGIN/COMMIT) — acceptable in exchange for working RLS. Write paths
// already use transactions, so the overhead is just the SET LOCAL.
//
// SET LOCAL is preferred over SET because:
//  - LOCAL scopes to the current transaction, not the connection. The
//    pool can hand the connection to the next request without leaking
//    the previous request's tenant_id.
//  - SET (session-level) requires explicit RESET on release; if the
//    application crashes between SET and RESET, the connection returns
//    to the pool poisoned with the wrong tenant_id.

package pgstore

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"p9e.in/chetana/packages/database/rlssession"
	"p9e.in/chetana/packages/p9context"
)

// withTenantTx runs fn inside a transaction with `app.tenant_id` SET LOCAL.
// tenantID must be non-empty — empty would skip the SET (per rlssession's
// empty-skip semantics) and the FORCE RLS policy would reject every read.
func (s *Store) withTenantTx(
	ctx context.Context,
	tenantID string,
	opts pgx.TxOptions,
	fn func(tx pgx.Tx) error,
) error {
	if tenantID == "" {
		return fmt.Errorf("withTenantTx: tenantID is required")
	}
	tx, err := s.pool.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := rlssession.SetLocal(ctx, tx, p9context.RLSScope{TenantID: tenantID}); err != nil {
		return fmt.Errorf("withTenantTx: %w", err)
	}

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
