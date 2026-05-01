package pgstore

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"

	"p9e.in/samavaya/packages/classregistry"
	"p9e.in/samavaya/packages/database/rlssession"
	"p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/p9context"
	"p9e.in/samavaya/packages/ulid"
)

// Compile-time check: Store implements EntityStore.
var _ classregistry.EntityStore = (*Store)(nil)

// GetByNaturalKey satisfies classregistry.EntityStore.
func (s *Store) GetByNaturalKey(
	ctx context.Context,
	tenantID, domain, class, naturalKey string,
) (*classregistry.ClassEntity, error) {
	var (
		id     string
		label  string
		attrJS []byte
		status string
	)
	var notFound bool
	err := s.withTenantTx(ctx, tenantID, pgx.TxOptions{AccessMode: pgx.ReadOnly}, func(tx pgx.Tx) error {
		scanErr := tx.QueryRow(ctx, `
SELECT id, label, attributes, status
  FROM classregistry.class_entities
 WHERE tenant_id = $1 AND domain = $2 AND class = $3 AND natural_key = $4
   AND deleted_at IS NULL`,
			tenantID, domain, class, naturalKey,
		).Scan(&id, &label, &attrJS, &status)
		if scanErr == pgx.ErrNoRows {
			notFound = true
			return nil
		}
		return scanErr
	})
	if err != nil {
		return nil, fmt.Errorf("read class_entity: %w", err)
	}
	if notFound {
		return nil, errors.NotFound(
			"CLASSREGISTRY_ENTITY_NOT_FOUND",
			fmt.Sprintf("no live entity for tenant=%q domain=%q class=%q natural_key=%q",
				tenantID, domain, class, naturalKey),
		)
	}
	var attrs map[string]classregistry.AttributeValue
	if len(attrJS) > 0 {
		if err := json.Unmarshal(attrJS, &attrs); err != nil {
			return nil, errors.InternalServer(
				"CLASSREGISTRY_ENTITY_DECODE",
				fmt.Sprintf("decode attributes for entity %q: %v", id, err),
			)
		}
	}
	return &classregistry.ClassEntity{
		ID:         id,
		TenantID:   tenantID,
		Domain:     domain,
		Class:      class,
		NaturalKey: naturalKey,
		Label:      label,
		Attributes: attrs,
		Status:     status,
	}, nil
}

// Exists is the existence-check hot path used by LookupResolver.
func (s *Store) Exists(
	ctx context.Context,
	tenantID, domain, class, naturalKey string,
) (bool, error) {
	var one int
	var notFound bool
	err := s.withTenantTx(ctx, tenantID, pgx.TxOptions{AccessMode: pgx.ReadOnly}, func(tx pgx.Tx) error {
		scanErr := tx.QueryRow(ctx, `
SELECT 1
  FROM classregistry.class_entities
 WHERE tenant_id = $1 AND domain = $2 AND class = $3 AND natural_key = $4
   AND deleted_at IS NULL AND status = 'active'
 LIMIT 1`,
			tenantID, domain, class, naturalKey,
		).Scan(&one)
		if scanErr == pgx.ErrNoRows {
			notFound = true
			return nil
		}
		return scanErr
	})
	if err != nil {
		return false, fmt.Errorf("exists class_entity: %w", err)
	}
	if notFound {
		return false, nil
	}
	return true, nil
}

// List returns a filtered, sorted page + total count.
func (s *Store) List(
	ctx context.Context,
	filter classregistry.EntityListFilter,
) ([]*classregistry.ClassEntity, int32, error) {
	if filter.TenantID == "" {
		return nil, 0, errors.BadRequest(
			"CLASSREGISTRY_ENTITY_LIST_NO_TENANT",
			"EntityListFilter.TenantID is required",
		)
	}
	where := `tenant_id = $1 AND deleted_at IS NULL`
	args := []interface{}{filter.TenantID}
	if filter.Domain != "" {
		args = append(args, filter.Domain)
		where += fmt.Sprintf(" AND domain = $%d", len(args))
	}
	if filter.Class != "" {
		args = append(args, filter.Class)
		where += fmt.Sprintf(" AND class = $%d", len(args))
	}
	if filter.NaturalKeyPrefix != "" {
		args = append(args, filter.NaturalKeyPrefix+"%")
		where += fmt.Sprintf(" AND natural_key LIKE $%d", len(args))
	}
	if !filter.IncludeArchived {
		where += " AND status = 'active'"
	}

	limit := int32(50)
	if filter.Limit > 0 && filter.Limit <= 500 {
		limit = filter.Limit
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	var total int32
	var out []*classregistry.ClassEntity
	err := s.withTenantTx(ctx, filter.TenantID, pgx.TxOptions{AccessMode: pgx.ReadOnly}, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx,
			`SELECT COUNT(*)::int FROM classregistry.class_entities WHERE `+where, args...,
		).Scan(&total); err != nil {
			return fmt.Errorf("count class_entities: %w", err)
		}

		pageArgs := append(append([]interface{}{}, args...), limit, offset)
		q := `SELECT id, domain, class, natural_key, label, attributes, status
                FROM classregistry.class_entities
               WHERE ` + where + `
               ORDER BY domain, class, natural_key
               LIMIT $` + fmt.Sprintf("%d", len(pageArgs)-1) + ` OFFSET $` + fmt.Sprintf("%d", len(pageArgs))

		rows, err := tx.Query(ctx, q, pageArgs...)
		if err != nil {
			return fmt.Errorf("query class_entities: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				id, domain, class, naturalKey, label, status string
				attrJS                                       []byte
			)
			if err := rows.Scan(&id, &domain, &class, &naturalKey, &label, &attrJS, &status); err != nil {
				return fmt.Errorf("scan class_entity: %w", err)
			}
			var attrs map[string]classregistry.AttributeValue
			if len(attrJS) > 0 {
				if err := json.Unmarshal(attrJS, &attrs); err != nil {
					return errors.InternalServer(
						"CLASSREGISTRY_ENTITY_DECODE",
						fmt.Sprintf("decode attributes for entity %q: %v", id, err),
					)
				}
			}
			out = append(out, &classregistry.ClassEntity{
				ID:         id,
				TenantID:   filter.TenantID,
				Domain:     domain,
				Class:      class,
				NaturalKey: naturalKey,
				Label:      label,
				Attributes: attrs,
				Status:     status,
			})
		}
		return rows.Err()
	})
	if err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// Upsert inserts or updates a class_entity row. The caller is
// responsible for having run registry.ValidateAttributes against the
// entity's class beforehand — we do NOT re-validate here, to let the
// provisioner + admin paths stay in control of which registry view
// (global vs per-tenant overlay) drives validation.
func (s *Store) Upsert(
	ctx context.Context,
	entity *classregistry.ClassEntity,
	actorID string,
) (string, error) {
	if err := validateEntityScope(entity, actorID); err != nil {
		return "", err
	}

	attrsJSON, err := json.Marshal(entity.Attributes)
	if err != nil {
		return "", errors.InternalServer(
			"CLASSREGISTRY_ENTITY_ENCODE",
			fmt.Sprintf("encode attributes for %s/%s/%s: %v",
				entity.Domain, entity.Class, entity.NaturalKey, err),
		)
	}
	status := entity.Status
	if status == "" {
		status = "active"
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Set RLS session variable so FORCE RLS policies pass under
	// non-superuser DB role (samavaya_app). See pgstore/rls_session.go.
	if err := rlssession.SetLocal(ctx, tx, p9context.RLSScope{TenantID: entity.TenantID}); err != nil {
		return "", fmt.Errorf("upsert: %w", err)
	}

	// Read the prior row (for pre-write hook's `old` argument). Not
	// finding one is fine — inserts call hooks with old=nil.
	var (
		existingID   string
		oldLabel     string
		oldAttrJSON  []byte
		oldStatus    string
	)
	err = tx.QueryRow(ctx, `
SELECT id, label, attributes, status
  FROM classregistry.class_entities
 WHERE tenant_id = $1 AND domain = $2 AND class = $3 AND natural_key = $4
   AND deleted_at IS NULL`,
		entity.TenantID, entity.Domain, entity.Class, entity.NaturalKey,
	).Scan(&existingID, &oldLabel, &oldAttrJSON, &oldStatus)

	// Fire pre-write hook. old is nil for inserts; populated for updates.
	if s.hooks != nil {
		var old *classregistry.ClassEntity
		if err == nil {
			var oldAttrs map[string]classregistry.AttributeValue
			if len(oldAttrJSON) > 0 {
				_ = json.Unmarshal(oldAttrJSON, &oldAttrs)
			}
			old = &classregistry.ClassEntity{
				ID:         existingID,
				TenantID:   entity.TenantID,
				Domain:     entity.Domain,
				Class:      entity.Class,
				NaturalKey: entity.NaturalKey,
				Label:      oldLabel,
				Attributes: oldAttrs,
				Status:     oldStatus,
			}
		}
		if hookErr := s.hooks.FirePreWrite(ctx, entity.TenantID, old, entity); hookErr != nil {
			return "", hookErr
		}
	}

	var id string
	switch err {
	case nil:
		id = existingID
		_, err = tx.Exec(ctx, `
UPDATE classregistry.class_entities
   SET label      = $1,
       attributes = $2,
       status     = $3,
       updated_by = $4,
       updated_at = NOW()
 WHERE id = $5`,
			entity.Label, attrsJSON, status, actorID, id,
		)
		if err != nil {
			return "", fmt.Errorf("update class_entity: %w", err)
		}
	case pgx.ErrNoRows:
		id = ulid.New().String()
		_, err = tx.Exec(ctx, `
INSERT INTO classregistry.class_entities
       (id, tenant_id, domain, class, natural_key, label, attributes, status, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)`,
			id, entity.TenantID, entity.Domain, entity.Class, entity.NaturalKey,
			entity.Label, attrsJSON, status, actorID,
		)
		if err != nil {
			return "", fmt.Errorf("insert class_entity: %w", err)
		}
	default:
		return "", fmt.Errorf("read existing class_entity: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit tx: %w", err)
	}

	// Fire post-write hook. Errors are logged by the caller but do not
	// unwind the commit — post-write hooks are observational.
	if s.hooks != nil {
		committedEntity := *entity
		committedEntity.ID = id
		if committedEntity.Status == "" {
			committedEntity.Status = status
		}
		_ = s.hooks.FirePostWrite(ctx, entity.TenantID, &committedEntity)
	}

	return id, nil
}

// Archive flips status to 'archived'.
func (s *Store) Archive(ctx context.Context, tenantID, domain, class, naturalKey, actorID string) error {
	if actorID == "" {
		return errors.BadRequest(
			"CLASSREGISTRY_ENTITY_ARCHIVE_MISSING_ACTOR",
			"actor_id is required",
		)
	}
	var rowsAffected int64
	err := s.withTenantTx(ctx, tenantID, pgx.TxOptions{}, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, `
UPDATE classregistry.class_entities
   SET status     = 'archived',
       updated_by = $1,
       updated_at = NOW()
 WHERE tenant_id = $2 AND domain = $3 AND class = $4 AND natural_key = $5
   AND deleted_at IS NULL`,
			actorID, tenantID, domain, class, naturalKey,
		)
		if err != nil {
			return err
		}
		rowsAffected = ct.RowsAffected()
		return nil
	})
	if err != nil {
		return fmt.Errorf("archive class_entity: %w", err)
	}
	if rowsAffected == 0 {
		return errors.NotFound(
			"CLASSREGISTRY_ENTITY_NOT_FOUND",
			fmt.Sprintf("no live entity to archive: %s/%s/%s for tenant %q",
				domain, class, naturalKey, tenantID),
		)
	}
	return nil
}

// Delete soft-deletes the row.
func (s *Store) Delete(ctx context.Context, tenantID, domain, class, naturalKey, actorID string) error {
	if actorID == "" {
		return errors.BadRequest(
			"CLASSREGISTRY_ENTITY_DELETE_MISSING_ACTOR",
			"actor_id is required",
		)
	}
	var rowsAffected int64
	err := s.withTenantTx(ctx, tenantID, pgx.TxOptions{}, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, `
UPDATE classregistry.class_entities
   SET deleted_at = NOW(),
       deleted_by = $1
 WHERE tenant_id = $2 AND domain = $3 AND class = $4 AND natural_key = $5
   AND deleted_at IS NULL`,
			actorID, tenantID, domain, class, naturalKey,
		)
		if err != nil {
			return err
		}
		rowsAffected = ct.RowsAffected()
		return nil
	})
	if err != nil {
		return fmt.Errorf("delete class_entity: %w", err)
	}
	if rowsAffected == 0 {
		return errors.NotFound(
			"CLASSREGISTRY_ENTITY_NOT_FOUND",
			fmt.Sprintf("no live entity to delete: %s/%s/%s for tenant %q",
				domain, class, naturalKey, tenantID),
		)
	}
	return nil
}

func validateEntityScope(e *classregistry.ClassEntity, actorID string) error {
	if e == nil {
		return errors.BadRequest("CLASSREGISTRY_ENTITY_NIL", "entity is nil")
	}
	if e.TenantID == "" {
		return errors.BadRequest("CLASSREGISTRY_ENTITY_NO_TENANT", "tenant_id is required")
	}
	if e.Domain == "" || e.Class == "" {
		return errors.BadRequest("CLASSREGISTRY_ENTITY_NO_CLASS", "domain + class are required")
	}
	if e.NaturalKey == "" {
		return errors.BadRequest("CLASSREGISTRY_ENTITY_NO_NATURAL_KEY", "natural_key is required")
	}
	if e.Label == "" {
		return errors.BadRequest("CLASSREGISTRY_ENTITY_NO_LABEL", "label is required")
	}
	if actorID == "" {
		return errors.BadRequest("CLASSREGISTRY_ENTITY_NO_ACTOR", "actor_id is required")
	}
	return nil
}
