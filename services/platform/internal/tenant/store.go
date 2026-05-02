// Package tenant implements the chetana platform-tenants service.
//
// → REQ-FUNC-PLT-TENANT-001 (single tenant in v1; multi-ready
//                              data model).
// → REQ-FUNC-PLT-TENANT-002 (per-tenant security policy:
//                              MFA required, session timeout,
//                              password policy).
// → REQ-FUNC-PLT-TENANT-003 (no row-level security in v1; tenant
//                              isolation enforced at the
//                              application layer + the
//                              `tenant_id NOT NULL` lint guard).
// → design.md §3.1.
//
// Single-tenant runtime posture
//
// chetana ships as single-tenant in v1: one row in the `tenants`
// table, seeded by the migration; every domain row carries that
// `tenant_id` so the schema is forward-compatible with the
// multi-tenant runtime that lands in v1.x.
//
// Why no RLS?
//
// REQ-FUNC-PLT-TENANT-003 explicitly prohibits PostgreSQL Row-
// Level Security in v1. Reasoning (from the design doc):
//
//   • RLS bypasses the application-layer audit chain — a SQL
//     query that RLS implicitly filters does not appear in the
//     authz.AuditEvent stream.
//   • The single-tenant deployment makes RLS the wrong tool
//     (zero rows would ever be filtered).
//   • The lint guard in `services/packages/db/lint/tenant_id.go`
//     gives us most of the safety RLS provides at lower
//     operational cost.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Tenant is the in-memory shape of one `tenants` row.
type Tenant struct {
	ID                 string
	Name               string
	DisplayName        string
	Status             string
	DataClassification string
	SecurityPolicy     SecurityPolicy
	Quotas             Quotas
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// SecurityPolicy holds the per-tenant security knobs the IAM
// service consults at login + RPC time. The shape is JSONB on
// disk so a new knob doesn't require a migration; new fields here
// must default to zero-values that match the legacy behaviour so
// pre-existing rows keep working.
type SecurityPolicy struct {
	MFARequired           bool          `json:"mfa_required"`
	SessionIdleTimeout    time.Duration `json:"session_idle_timeout"`
	SessionAbsoluteLimit  time.Duration `json:"session_absolute_limit"`
	MaxConcurrentSessions int           `json:"max_concurrent_sessions"`
	PasswordMinLength     int           `json:"password_min_length"`
	PasswordRequireMixed  bool          `json:"password_require_mixed"`
}

// DefaultSecurityPolicy mirrors the IAM-001..009 defaults so a
// freshly-seeded tenant matches what the rest of the platform
// expects out of the box.
func DefaultSecurityPolicy() SecurityPolicy {
	return SecurityPolicy{
		MFARequired:           false,
		SessionIdleTimeout:    time.Hour,
		SessionAbsoluteLimit:  24 * time.Hour,
		MaxConcurrentSessions: 5,
		PasswordMinLength:     12,
		PasswordRequireMixed:  true,
	}
}

// Quotas holds per-tenant resource caps. Surfaced in the platform
// API so customers can see usage; enforced by the per-domain
// services that consume them (e.g. eo-pipeline checks
// MaxObjectsPerDay before accepting an upload).
type Quotas struct {
	MaxUsers           int `json:"max_users"`
	MaxRolesPerUser    int `json:"max_roles_per_user"`
	MaxAPIRequestsHour int `json:"max_api_requests_hour"`
}

// DefaultQuotas mirrors the v1 single-tenant ceiling.
func DefaultQuotas() Quotas {
	return Quotas{
		MaxUsers:           1_000,
		MaxRolesPerUser:    32,
		MaxAPIRequestsHour: 1_000_000,
	}
}

// Status constants for the tenants.status column. Aligned with
// the CHECK constraint in the migration.
const (
	StatusActive    = "active"
	StatusSuspended = "suspended"
	StatusArchived  = "archived"
)

// Store wraps a pgxpool.Pool with the tenants CRUD helpers.
type Store struct {
	pool *pgxpool.Pool
	clk  func() time.Time
}

// NewStore wraps a pool. clock=nil → time.Now.
func NewStore(pool *pgxpool.Pool, clock func() time.Time) *Store {
	if clock == nil {
		clock = time.Now
	}
	return &Store{pool: pool, clk: clock}
}

// Get returns the single tenant by id. Returns ErrTenantNotFound
// when the row is missing.
func (s *Store) Get(ctx context.Context, tenantID string) (*Tenant, error) {
	if tenantID == "" {
		return nil, ErrTenantNotFound
	}
	const q = `
SELECT id, name, display_name, status,
       data_classification, security_policy, quotas,
       created_at, updated_at
FROM tenants
WHERE id = $1
`
	var (
		t              Tenant
		policyRaw      []byte
		quotasRaw      []byte
	)
	err := s.pool.QueryRow(ctx, q, tenantID).Scan(
		&t.ID, &t.Name, &t.DisplayName, &t.Status,
		&t.DataClassification, &policyRaw, &quotasRaw,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTenantNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("tenant: get: %w", err)
	}
	if err := json.Unmarshal(policyRaw, &t.SecurityPolicy); err != nil {
		return nil, fmt.Errorf("tenant: parse security_policy: %w", err)
	}
	if err := json.Unmarshal(quotasRaw, &t.Quotas); err != nil {
		return nil, fmt.Errorf("tenant: parse quotas: %w", err)
	}
	return &t, nil
}

// UpdateSecurityPolicy patches the per-tenant security knobs.
// Used by the platform admin RPC.
func (s *Store) UpdateSecurityPolicy(ctx context.Context, tenantID string, policy SecurityPolicy) error {
	body, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("tenant: marshal policy: %w", err)
	}
	tag, err := s.pool.Exec(ctx, `
UPDATE tenants
SET security_policy = $2, updated_at = $3
WHERE id = $1
`, tenantID, body, s.clk().UTC())
	if err != nil {
		return fmt.Errorf("tenant: update policy: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTenantNotFound
	}
	return nil
}

// UpdateQuotas patches the per-tenant quota knobs.
func (s *Store) UpdateQuotas(ctx context.Context, tenantID string, quotas Quotas) error {
	body, err := json.Marshal(quotas)
	if err != nil {
		return fmt.Errorf("tenant: marshal quotas: %w", err)
	}
	tag, err := s.pool.Exec(ctx, `
UPDATE tenants
SET quotas = $2, updated_at = $3
WHERE id = $1
`, tenantID, body, s.clk().UTC())
	if err != nil {
		return fmt.Errorf("tenant: update quotas: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTenantNotFound
	}
	return nil
}

// CreateForTest inserts a row directly. Production deploys
// receive their tenant via the seed migration; tests + ops
// scripts use this helper.
func (s *Store) CreateForTest(ctx context.Context, t Tenant) error {
	if t.SecurityPolicy == (SecurityPolicy{}) {
		t.SecurityPolicy = DefaultSecurityPolicy()
	}
	if t.Quotas == (Quotas{}) {
		t.Quotas = DefaultQuotas()
	}
	if t.Status == "" {
		t.Status = StatusActive
	}
	if t.DataClassification == "" {
		t.DataClassification = "cui"
	}
	policyRaw, err := json.Marshal(t.SecurityPolicy)
	if err != nil {
		return fmt.Errorf("tenant: marshal policy: %w", err)
	}
	quotasRaw, err := json.Marshal(t.Quotas)
	if err != nil {
		return fmt.Errorf("tenant: marshal quotas: %w", err)
	}
	if _, err := s.pool.Exec(ctx, `
INSERT INTO tenants
  (id, name, display_name, status, data_classification,
   security_policy, quotas)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (id) DO NOTHING
`, t.ID, t.Name, t.DisplayName, t.Status, t.DataClassification,
		policyRaw, quotasRaw,
	); err != nil {
		return fmt.Errorf("tenant: insert: %w", err)
	}
	return nil
}

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrTenantNotFound is returned by Get / UpdateSecurityPolicy /
// UpdateQuotas when no row matches.
var ErrTenantNotFound = errors.New("tenant: not found")
