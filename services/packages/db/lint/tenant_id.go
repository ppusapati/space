// Package lint contains the chetana database-schema lints.
//
// tenant_id.go implements the REQ-CONST-007 / REQ-FUNC-PLT-TENANT-001
// guard: every domain table created by `services/**/migrations/*.sql`
// MUST carry a `tenant_id` column so the schema is forward-
// compatible with multi-tenant runtime.
//
// The check is pure SQL-text scanning; no Postgres connection
// required, so it runs fast in CI on every PR. Tables explicitly
// flagged exempt (system tables, cross-tenant audit chain) live
// in the `Exempt` allowlist below.
//
// Usage:
//
//	import "p9e.in/chetana/packages/db/lint"
//	violations, err := lint.CheckMigrations("services")
//	if err != nil { ... }
//	if len(violations) > 0 { os.Exit(1) }
//
// A thin CLI wrapper lives at `tools/db/tenantid-lint/main.go`
// (TASK-P1-TENANT-001 ships the library; the CLI binary is
// stamped out in TASK-P1-CI-001's lint stage).

package lint

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Exempt names tables that are allowed to omit `tenant_id`.
// Cross-tenant operational tables go here; domain tables MUST
// NOT be added without code-owner approval.
//
// Two categories of exemption:
//
//   1. Genuinely cross-tenant tables (the `tenants` registry
//      itself, OAuth clients keyed by client_id, SAML IdPs keyed
//      by entity_id) — these will never have a tenant_id column.
//
//   2. Grandfathered IAM tables that pre-date this lint guard.
//      These have a follow-up task (TASK-P1-IAM-TENANT-RETROFIT)
//      to retro-add the column + a backfill data migration; until
//      then they live behind the Exempt list rather than blocking
//      every CI run.
var Exempt = map[string]bool{
	// (1) Genuinely cross-tenant.
	"tenants":          true, // the tenants registry itself
	"oauth2_clients":   true, // platform-wide; client_id is the scope
	"saml_idps":        true, // platform-wide; entity_id is the scope
	"role_permissions": true, // M2M; role_id carries the scope via roles.tenant_id
	"user_roles":       true, // M2M; user_id carries the scope
	"policies":         true, // policies have an explicit `tenant` text col instead

	// (2) Grandfathered IAM tables — TASK-P1-IAM-TENANT-RETROFIT.
	"mfa_totp_secrets":     true, // TODO: retrofit tenant_id (linked to user_id's tenant)
	"mfa_backup_codes":     true, // TODO: retrofit tenant_id
	"webauthn_credentials": true, // TODO: retrofit tenant_id
	"password_resets":      true, // TODO: retrofit tenant_id
}

// Violation is one offending CREATE TABLE statement.
type Violation struct {
	File      string
	Table     string
	Line      int
	Statement string
}

// String formats a Violation for CLI output.
func (v Violation) String() string {
	return fmt.Sprintf("%s:%d: table %q is missing tenant_id (and is not in Exempt)",
		v.File, v.Line, v.Table)
}

// CheckMigrations walks `root` looking for *.sql files under any
// `migrations/` directory and lints every CREATE TABLE block.
// Returns the list of violations + the first I/O error (if any).
func CheckMigrations(root string) ([]Violation, error) {
	if root == "" {
		return nil, errors.New("lint: root is required")
	}
	var violations []Violation
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// Only inspect SQL files inside a `migrations/` folder.
		if filepath.Ext(path) != ".sql" {
			return nil
		}
		if !strings.Contains(filepath.ToSlash(path), "/migrations/") {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fileViolations := CheckSQL(string(body), path)
		violations = append(violations, fileViolations...)
		return nil
	})
	return violations, err
}

// reCreateTable matches `CREATE TABLE [IF NOT EXISTS] <name> (`
// case-insensitively. We capture the table name + the open-paren
// offset so the body scan starts at the right place.
var reCreateTable = regexp.MustCompile(
	`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?([a-z_][a-z0-9_]*)\s*\(`,
)

// reTenantID matches a `tenant_id` column declaration. We require
// the name as a whole word at the start of a line (or after a
// comma) so a comment or a string literal containing the substring
// doesn't false-positive.
var reTenantID = regexp.MustCompile(`(?i)(?:^|\n|,)\s*tenant_id\s+`)

// CheckSQL scans a single SQL document and returns violations.
// `filename` is used in the Violation's File field.
func CheckSQL(body, filename string) []Violation {
	var violations []Violation
	matches := reCreateTable.FindAllStringSubmatchIndex(body, -1)
	for _, m := range matches {
		// m[0]/m[1]: full match start/end. m[2]/m[3]: capture
		// (table name) start/end. The body of the CREATE TABLE
		// statement starts at m[1]-1 (the `(`).
		tableName := body[m[2]:m[3]]
		if Exempt[tableName] {
			continue
		}
		// Find the matching close-paren for this table body. We
		// walk forward from the open-paren counting paren depth.
		depth := 0
		end := -1
		for i := m[1] - 1; i < len(body); i++ {
			switch body[i] {
			case '(':
				depth++
			case ')':
				depth--
				if depth == 0 {
					end = i
					break
				}
			}
			if end >= 0 {
				break
			}
		}
		if end < 0 {
			// Unterminated CREATE TABLE — let the SQL parser
			// surface this; we just skip the lint.
			continue
		}
		stmt := body[m[0] : end+1]
		if !reTenantID.MatchString(stmt) {
			line := 1 + strings.Count(body[:m[0]], "\n")
			violations = append(violations, Violation{
				File:      filename,
				Table:     tableName,
				Line:      line,
				Statement: stmt,
			})
		}
	}
	return violations
}
