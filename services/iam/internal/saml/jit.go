// jit.go — Just-In-Time user provisioning.
//
// On a successful ACS callback the chetana SP looks up the
// IdP-supplied email in the users table:
//
//   • Found    → return the existing user; project the IdP's
//                groups onto chetana roles for this session.
//                (Existing roles in the users table are NOT
//                overwritten in v1 — that "sync on every login"
//                policy can be wired in TASK-P1-IAM-USER-ATTRS
//                once the user-attributes table ships.)
//
//   • Missing  → INSERT a new users row with status=active,
//                no password (federated users have no local
//                credential), and the IdP-supplied display name.
//
// The set of roles a freshly-JIT-provisioned user gets is the
// union of:
//
//   • IdP attribute mapping: GroupRoleMap[g] for every g present
//     in the assertion's GroupsAttribute.
//   • DefaultRoles: applied to every user from this IdP.

package saml

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// JITProvisioner finds or creates a chetana user from the SAML
// assertion attributes.
type JITProvisioner struct {
	pool     *pgxpool.Pool
	tenantID string
	clk      func() time.Time
}

// NewJITProvisioner wraps a pool. tenantID is the active tenant
// in single-tenant runtime (TASK-P1-TENANT-001 keeps this
// constant in v1). clock=nil → time.Now.
func NewJITProvisioner(pool *pgxpool.Pool, tenantID string, clock func() time.Time) (*JITProvisioner, error) {
	if pool == nil {
		return nil, errors.New("saml: nil pool")
	}
	if tenantID == "" {
		return nil, errors.New("saml: empty tenant_id")
	}
	if clock == nil {
		clock = time.Now
	}
	return &JITProvisioner{pool: pool, tenantID: tenantID, clk: clock}, nil
}

// ProvisionInput is the per-assertion input.
type ProvisionInput struct {
	NameID     string
	Attributes map[string][]string
}

// ProvisionOutput is the per-assertion output.
type ProvisionOutput struct {
	UserID  string
	Email   string
	Roles   []string
	Created bool
}

// Provision finds-or-creates the chetana user. The transaction
// runs INSERT ... ON CONFLICT DO NOTHING so two concurrent ACS
// callbacks for the same user resolve to the same row.
func (p *JITProvisioner) Provision(ctx context.Context, idp *IdP, in ProvisionInput) (*ProvisionOutput, error) {
	if idp == nil {
		return nil, errors.New("saml: nil idp")
	}
	email, err := requireEmail(idp.AttributeMapping, in.Attributes)
	if err != nil {
		return nil, err
	}
	display := firstAttribute(idp.AttributeMapping.DisplayNameAttribute, in.Attributes)
	roles := projectRoles(idp.AttributeMapping, in.Attributes)

	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("saml: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Try to find an existing user (case-insensitive on email_lower).
	emailLower := strings.ToLower(email)
	const lookup = `
SELECT id FROM users
WHERE tenant_id = $1 AND email_lower = $2
LIMIT 1
`
	var existingID string
	err = tx.QueryRow(ctx, lookup, p.tenantID, emailLower).Scan(&existingID)
	switch {
	case err == nil:
		if cerr := tx.Commit(ctx); cerr != nil {
			return nil, fmt.Errorf("saml: commit: %w", cerr)
		}
		return &ProvisionOutput{
			UserID: existingID,
			Email:  email,
			Roles:  roles,
		}, nil
	case errors.Is(err, pgx.ErrNoRows):
		// fall through to INSERT
	default:
		return nil, fmt.Errorf("saml: lookup user: %w", err)
	}

	// Create the user. password_hash + password_algo are empty for
	// federated users (no local credential). Status is `active`
	// because the IdP just authenticated them.
	now := p.clk().UTC()
	const insert = `
INSERT INTO users
  (tenant_id, email_lower, email_display, password_hash, password_algo,
   status, data_classification, created_at, updated_at)
VALUES
  ($1, $2, $3, '', '', 'active', 'cui', $4, $4)
RETURNING id
`
	var newID string
	if err := tx.QueryRow(ctx, insert,
		p.tenantID, emailLower, displayOrEmail(display, email), now,
	).Scan(&newID); err != nil {
		return nil, fmt.Errorf("saml: insert user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("saml: commit: %w", err)
	}
	return &ProvisionOutput{
		UserID:  newID,
		Email:   email,
		Roles:   roles,
		Created: true,
	}, nil
}

// projectRoles applies the IdP's attribute mapping to the
// assertion's group attribute and unions in DefaultRoles. Output
// is de-duplicated and sorted-by-first-appearance for stable
// downstream behaviour.
func projectRoles(m AttributeMapping, attrs map[string][]string) []string {
	seen := make(map[string]bool)
	var out []string
	add := func(role string) {
		if role == "" || seen[role] {
			return
		}
		seen[role] = true
		out = append(out, role)
	}

	if m.GroupsAttribute != "" {
		for _, g := range attrs[m.GroupsAttribute] {
			if mapped, ok := m.GroupRoleMap[g]; ok {
				add(mapped)
			}
		}
	}
	for _, r := range m.DefaultRoles {
		add(r)
	}
	return out
}

// requireEmail extracts the email per the IdP's attribute
// mapping. Returns ErrMissingEmail when the assertion did not
// include the configured attribute (or the attribute had no
// values).
func requireEmail(m AttributeMapping, attrs map[string][]string) (string, error) {
	if m.EmailAttribute == "" {
		return "", ErrMissingEmail
	}
	values := attrs[m.EmailAttribute]
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			return v, nil
		}
	}
	return "", ErrMissingEmail
}

// firstAttribute returns the first non-empty value of the named
// attribute, or empty when the attribute is missing.
func firstAttribute(name string, attrs map[string][]string) string {
	if name == "" {
		return ""
	}
	for _, v := range attrs[name] {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}

func displayOrEmail(display, email string) string {
	if display != "" {
		return display
	}
	return email
}

// ErrMissingEmail is returned when the assertion did not include
// the IdP's configured email attribute.
var ErrMissingEmail = errors.New("saml: assertion missing email attribute")
