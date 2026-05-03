// Package store implements the notify service's template +
// preferences persistence.

package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ppusapati/space/services/notify/internal/template"
)

// Channel constants. Aligned with the CHECK constraint in the
// migration.
const (
	ChannelEmail = "email"
	ChannelSMS   = "sms"
	ChannelInApp = "inapp"
)

// TemplateStore wraps a pgx pool with the template helpers.
type TemplateStore struct {
	pool *pgxpool.Pool
	clk  func() time.Time
}

// NewTemplateStore wraps a pool. clock=nil → time.Now.
func NewTemplateStore(pool *pgxpool.Pool, clock func() time.Time) (*TemplateStore, error) {
	if pool == nil {
		return nil, errors.New("store: nil pool")
	}
	if clock == nil {
		clock = time.Now
	}
	return &TemplateStore{pool: pool, clk: clock}, nil
}

// LookupActive fetches the highest-version active template by id
// for a given channel. Returns ErrTemplateNotFound when no row.
func (s *TemplateStore) LookupActive(ctx context.Context, id, channel string) (*template.Template, error) {
	if id == "" || channel == "" {
		return nil, ErrTemplateNotFound
	}
	const q = `
SELECT id, version, channel, body, variables_schema, mandatory
FROM notification_templates
WHERE id = $1 AND channel = $2 AND active = true
ORDER BY version DESC
LIMIT 1
`
	var (
		t          template.Template
		schemaRaw  []byte
	)
	err := s.pool.QueryRow(ctx, q, id, channel).Scan(
		&t.ID, &t.Version, &t.Channel, &t.Body, &schemaRaw, &t.Mandatory,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTemplateNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store: lookup template: %w", err)
	}
	if err := json.Unmarshal(schemaRaw, &t.Variables); err != nil {
		return nil, fmt.Errorf("store: parse schema: %w", err)
	}
	return &t, nil
}

// CreateForTest inserts a template row directly. Production
// deploys register templates via an admin RPC; tests + ops
// scripts use this.
func (s *TemplateStore) CreateForTest(ctx context.Context, t template.Template, active bool) error {
	schema, err := json.Marshal(t.Variables)
	if err != nil {
		return fmt.Errorf("store: marshal schema: %w", err)
	}
	if t.Version <= 0 {
		t.Version = 1
	}
	const q = `
INSERT INTO notification_templates
  (id, version, channel, body, variables_schema, mandatory, active)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (id, version, channel) DO NOTHING
`
	if _, err := s.pool.Exec(ctx, q,
		t.ID, t.Version, t.Channel, t.Body, schema, t.Mandatory, active,
	); err != nil {
		return fmt.Errorf("store: insert template: %w", err)
	}
	return nil
}

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrTemplateNotFound is returned by LookupActive when no row matches.
var ErrTemplateNotFound = errors.New("store: template not found")
