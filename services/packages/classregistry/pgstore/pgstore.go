// Package pgstore is the PostgreSQL-backed OverrideStore used by the
// classregistry tenant overlay (F.6.2). It reads and writes the
// `classregistry.tenant_class_overrides` table created by migration
// 000070 and emits operational audit rows into
// `classregistry.tenant_override_audit`.
//
// The implementation is deliberately narrow: four operations
// (ListForTenantDomain, UpsertOverride, DeleteOverride, ListForTenant)
// plus the override-audit insert tied to write paths. Consumers wire
// an instance via fx by providing the shared *pgxpool.Pool.
package pgstore

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/chetana/packages/classregistry"
	"p9e.in/chetana/packages/database/rlssession"
	"p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/p9context"
	"p9e.in/chetana/packages/ulid"
)

// Store is the pgxpool-backed override store. Satisfies
// classregistry.OverrideStore + classregistry.EntityStore + the admin
// write surface consumed by the classregistry admin RPCs.
//
// The hooks field is optional. When set, every class_entity Upsert
// fires pre-write + post-write hooks through the registry. When nil,
// Upsert skips hook firing — the pre-F.6.L3 compatibility path.
type Store struct {
	pool  *pgxpool.Pool
	hooks classregistry.HookRegistry
}

// New builds a Store bound to the shared application pool. The pool
// must target a database where migration 000070 + 000072 have been
// run. Hooks are not wired — call WithHooks to add a HookRegistry.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// WithHooks returns a Store that consults hooks on every class_entity
// write. This is a fluent setter rather than a constructor param so
// fx wiring can build the Store first + bind hooks in a separate
// fx.Invoke when the hook registry depends on other services that
// might not be available at Store construction time.
func (s *Store) WithHooks(h classregistry.HookRegistry) *Store {
	s.hooks = h
	return s
}

// OverrideRow is the administrative view of a tenant_class_overrides
// row. Returned by List* operations on the admin port.
type OverrideRow struct {
	ID        string
	TenantID  string
	Domain    string
	Class     string
	Override  classregistry.ClassOverride
	CreatedBy string
	UpdatedBy string
}

// ListForTenantDomain satisfies classregistry.OverrideStore — the
// per-request merge call made by the tenantView.
func (s *Store) ListForTenantDomain(
	ctx context.Context,
	tenantID, domain string,
) (map[string]classregistry.ClassOverride, error) {
	if s.pool == nil {
		return nil, errors.InternalServer(
			"CLASSREGISTRY_PGSTORE_NO_POOL",
			"pgstore was constructed without a pgxpool.Pool",
		)
	}
	out := map[string]classregistry.ClassOverride{}
	err := s.withTenantTx(ctx, tenantID, pgx.TxOptions{AccessMode: pgx.ReadOnly}, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `
SELECT class, override_json
  FROM classregistry.tenant_class_overrides
 WHERE tenant_id = $1
   AND domain = $2
   AND deleted_at IS NULL`,
			tenantID, domain,
		)
		if err != nil {
			return fmt.Errorf("query tenant overrides: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var class string
			var raw []byte
			if err := rows.Scan(&class, &raw); err != nil {
				return fmt.Errorf("scan tenant override row: %w", err)
			}
			var ov classregistry.ClassOverride
			if err := json.Unmarshal(raw, &ov); err != nil {
				return errors.InternalServer(
					"CLASSREGISTRY_OVERRIDE_DECODE",
					fmt.Sprintf("decode tenant override for %q/%q: %v", domain, class, err),
				)
			}
			out[class] = ov
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ListForTenant returns every live override row for a tenant across
// every domain. Admin-facing.
func (s *Store) ListForTenant(ctx context.Context, tenantID string) ([]OverrideRow, error) {
	var out []OverrideRow
	err := s.withTenantTx(ctx, tenantID, pgx.TxOptions{AccessMode: pgx.ReadOnly}, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `
SELECT id, tenant_id, domain, class, override_json, created_by, updated_by
  FROM classregistry.tenant_class_overrides
 WHERE tenant_id = $1
   AND deleted_at IS NULL
 ORDER BY domain, class`,
			tenantID,
		)
		if err != nil {
			return fmt.Errorf("query tenant overrides: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var r OverrideRow
			var raw []byte
			if err := rows.Scan(&r.ID, &r.TenantID, &r.Domain, &r.Class, &raw, &r.CreatedBy, &r.UpdatedBy); err != nil {
				return fmt.Errorf("scan tenant override row: %w", err)
			}
			if err := json.Unmarshal(raw, &r.Override); err != nil {
				return errors.InternalServer(
					"CLASSREGISTRY_OVERRIDE_DECODE",
					fmt.Sprintf("decode tenant override %q: %v", r.ID, err),
				)
			}
			out = append(out, r)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UpsertInput is the admin-port write payload for upsertOverride.
type UpsertInput struct {
	TenantID string
	Domain   string
	Class    string
	Override classregistry.ClassOverride
	ActorID  string
	Reason   string // free-text audit rationale; optional
}

// UpsertOverride inserts or updates a tenant override row. The write
// and its audit trail ship as a single transaction so the audit log
// never disagrees with the live row. Returns the row's ULID.
func (s *Store) UpsertOverride(ctx context.Context, in UpsertInput) (string, error) {
	if err := validateUpsert(in); err != nil {
		return "", err
	}
	nextJSON, err := json.Marshal(in.Override)
	if err != nil {
		return "", errors.InternalServer(
			"CLASSREGISTRY_OVERRIDE_ENCODE",
			fmt.Sprintf("encode override for %q/%q: %v", in.Domain, in.Class, err),
		)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// SET LOCAL app.tenant_id so FORCE RLS policies pass under
	// non-superuser DB role (chetana_app).
	if err := rlssession.SetLocal(ctx, tx, p9context.RLSScope{TenantID: in.TenantID}); err != nil {
		return "", fmt.Errorf("upsert override: %w", err)
	}

	// Look up existing row (if any) to capture previous_json and ID
	// so upsert+audit share the same identifier.
	var (
		existingID   string
		previousJSON []byte
	)
	err = tx.QueryRow(ctx, `
SELECT id, override_json
  FROM classregistry.tenant_class_overrides
 WHERE tenant_id = $1 AND domain = $2 AND class = $3 AND deleted_at IS NULL`,
		in.TenantID, in.Domain, in.Class,
	).Scan(&existingID, &previousJSON)

	var (
		id     string
		action string
	)
	switch err {
	case nil:
		id = existingID
		action = "updated"
		_, err = tx.Exec(ctx, `
UPDATE classregistry.tenant_class_overrides
   SET override_json = $1,
       updated_by    = $2,
       updated_at    = NOW()
 WHERE id = $3`,
			nextJSON, in.ActorID, id,
		)
		if err != nil {
			return "", fmt.Errorf("update tenant override: %w", err)
		}
	case pgx.ErrNoRows:
		id = ulid.New().String()
		action = "created"
		_, err = tx.Exec(ctx, `
INSERT INTO classregistry.tenant_class_overrides
       (id, tenant_id, domain, class, override_json, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $6)`,
			id, in.TenantID, in.Domain, in.Class, nextJSON, in.ActorID,
		)
		if err != nil {
			return "", fmt.Errorf("insert tenant override: %w", err)
		}
	default:
		return "", fmt.Errorf("read existing tenant override: %w", err)
	}

	if err := insertAudit(ctx, tx, auditEntry{
		OverrideID:   id,
		TenantID:     in.TenantID,
		Domain:       in.Domain,
		Class:        in.Class,
		Action:       action,
		PreviousJSON: previousJSON,
		NextJSON:     nextJSON,
		ActorID:      in.ActorID,
		Reason:       in.Reason,
	}); err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit tx: %w", err)
	}
	return id, nil
}

// DeleteInput is the admin-port payload for deleteOverride.
type DeleteInput struct {
	TenantID string
	Domain   string
	Class    string
	ActorID  string
	Reason   string
}

// DeleteOverride soft-deletes a tenant override. Returns ErrNotFound
// (wrapped with pgx.ErrNoRows semantics via errors.NotFound) when the
// row doesn't exist.
func (s *Store) DeleteOverride(ctx context.Context, in DeleteInput) error {
	if in.TenantID == "" || in.Domain == "" || in.Class == "" {
		return errors.BadRequest(
			"CLASSREGISTRY_OVERRIDE_DELETE_MISSING_KEY",
			"tenant_id, domain, and class are all required",
		)
	}
	if in.ActorID == "" {
		return errors.BadRequest(
			"CLASSREGISTRY_OVERRIDE_DELETE_MISSING_ACTOR",
			"actor_id is required for audit",
		)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := rlssession.SetLocal(ctx, tx, p9context.RLSScope{TenantID: in.TenantID}); err != nil {
		return fmt.Errorf("delete override: %w", err)
	}

	var (
		id           string
		previousJSON []byte
	)
	err = tx.QueryRow(ctx, `
SELECT id, override_json
  FROM classregistry.tenant_class_overrides
 WHERE tenant_id = $1 AND domain = $2 AND class = $3 AND deleted_at IS NULL`,
		in.TenantID, in.Domain, in.Class,
	).Scan(&id, &previousJSON)
	if err == pgx.ErrNoRows {
		return errors.NotFound(
			"CLASSREGISTRY_OVERRIDE_NOT_FOUND",
			fmt.Sprintf("no live override for tenant %q domain %q class %q", in.TenantID, in.Domain, in.Class),
		)
	}
	if err != nil {
		return fmt.Errorf("read existing tenant override: %w", err)
	}

	_, err = tx.Exec(ctx, `
UPDATE classregistry.tenant_class_overrides
   SET deleted_at = NOW(),
       deleted_by = $1
 WHERE id = $2`,
		in.ActorID, id,
	)
	if err != nil {
		return fmt.Errorf("soft-delete tenant override: %w", err)
	}

	if err := insertAudit(ctx, tx, auditEntry{
		OverrideID:   id,
		TenantID:     in.TenantID,
		Domain:       in.Domain,
		Class:        in.Class,
		Action:       "deleted",
		PreviousJSON: previousJSON,
		NextJSON:     nil,
		ActorID:      in.ActorID,
		Reason:       in.Reason,
	}); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Internal
// ---------------------------------------------------------------------------

type auditEntry struct {
	OverrideID   string
	TenantID     string
	Domain       string
	Class        string
	Action       string
	PreviousJSON []byte
	NextJSON     []byte
	ActorID      string
	Reason       string
}

func insertAudit(ctx context.Context, tx pgx.Tx, e auditEntry) error {
	_, err := tx.Exec(ctx, `
INSERT INTO classregistry.tenant_override_audit
       (id, override_id, tenant_id, domain, class, action, previous_json, next_json, actor_id, actor_reason)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		ulid.New().String(),
		e.OverrideID, e.TenantID, e.Domain, e.Class, e.Action,
		nullableJSON(e.PreviousJSON), nullableJSON(e.NextJSON),
		e.ActorID, nullableString(e.Reason),
	)
	if err != nil {
		return fmt.Errorf("insert override audit row: %w", err)
	}
	return nil
}

func nullableJSON(b []byte) interface{} {
	if len(b) == 0 {
		return nil
	}
	return b
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func validateUpsert(in UpsertInput) error {
	if in.TenantID == "" {
		return errors.BadRequest("CLASSREGISTRY_OVERRIDE_UPSERT_MISSING_TENANT", "tenant_id is required")
	}
	if in.Domain == "" {
		return errors.BadRequest("CLASSREGISTRY_OVERRIDE_UPSERT_MISSING_DOMAIN", "domain is required")
	}
	if in.Class == "" {
		return errors.BadRequest("CLASSREGISTRY_OVERRIDE_UPSERT_MISSING_CLASS", "class is required")
	}
	if in.ActorID == "" {
		return errors.BadRequest("CLASSREGISTRY_OVERRIDE_UPSERT_MISSING_ACTOR", "actor_id is required for audit")
	}
	return nil
}
