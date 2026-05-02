// Package search implements the audit-events read query surface.
//
// → REQ-FUNC-PLT-AUDIT-003 (search filters: time range, actor,
//                            action, resource, free-text JSONB).
// → design.md §5.4.
//
// The query DSL is a small typed struct (Query) that the
// implementation translates into a parametrised SQL statement.
// We deliberately do NOT accept raw SQL fragments — all filters
// are bound parameters, no string concatenation.
//
// Pagination is keyset-based on (event_time DESC, id DESC) so
// scrolling stays O(log n) even on the 100M-row table the
// acceptance gate names.

package search

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Query is the per-call input. Every field is optional except
// TenantID. Unset filters apply no restriction.
type Query struct {
	TenantID    string
	Start       time.Time // inclusive; zero = no lower bound
	End         time.Time // inclusive; zero = no upper bound
	ActorUserID string    // exact match
	Action      string    // exact match (a future enhancement adds prefix/wildcard)
	Resource    string    // exact match
	Decision    string    // "allow" | "deny" | "ok" | "fail" | "info"
	Procedure   string    // exact match (the Connect procedure name)
	// FreeText is matched against the metadata JSONB via
	// jsonb_path_exists — supports a minimal path-and-value
	// shape, e.g. "k1=v1" tests for top-level key k1 == v1.
	FreeText string
	Limit    int  // page size; capped at 500
	// Cursor: keyset pointer. Empty = first page.
	BeforeTime time.Time
	BeforeID   int64
}

// Result is one page of audit events.
type Result struct {
	Hits       []Hit
	NextCursor *Cursor // nil when no more pages
}

// Hit is the projected row shape returned to API callers.
// The chain hashes are NOT included by default — callers that
// need attestation use the export envelope (export/csv.go +
// export/json.go) which embeds the chain-tip signature.
type Hit struct {
	ID              int64
	TenantID        string
	EventTime       time.Time
	ActorUserID     string
	ActorSessionID  string
	ActorClientIP   string
	ActorUserAgent  string
	Action          string
	Resource        string
	Decision        string
	Reason          string
	MatchedPolicyID string
	Procedure       string
	Classification  string
	Metadata        []byte // raw JSON
}

// Cursor is the opaque keyset pointer the caller echoes back to
// fetch the next page.
type Cursor struct {
	BeforeTime time.Time `json:"before_time"`
	BeforeID   int64     `json:"before_id"`
}

// Service runs the queries.
type Service struct {
	pool *pgxpool.Pool
}

// NewService wraps a pool.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// MaxLimit caps the per-page result size. The IAM-controlled
// admin UI defaults to 100; bulk exports use the streaming
// export endpoints (TASK-P1-AUDIT-002 export/*) which bypass
// this cap.
const MaxLimit = 500

// Search runs one paginated query.
func (s *Service) Search(ctx context.Context, q Query) (*Result, error) {
	if q.TenantID == "" {
		return nil, errors.New("search: tenant_id is required")
	}
	limit := q.Limit
	if limit <= 0 || limit > MaxLimit {
		limit = MaxLimit
	}

	// Build the parametrised WHERE clause + arg list. We assemble
	// the predicates in a slice so the order is deterministic +
	// the placeholder indices stay aligned.
	conds := []string{"tenant_id = $1"}
	args := []any{q.TenantID}
	add := func(cond string, val any) {
		args = append(args, val)
		conds = append(conds, fmt.Sprintf(cond, len(args)))
	}
	if !q.Start.IsZero() {
		add("event_time >= $%d", q.Start.UTC())
	}
	if !q.End.IsZero() {
		add("event_time <= $%d", q.End.UTC())
	}
	if q.ActorUserID != "" {
		add("actor_user_id = NULLIF($%d, '')::uuid", q.ActorUserID)
	}
	if q.Action != "" {
		add("action = $%d", q.Action)
	}
	if q.Resource != "" {
		add("resource = $%d", q.Resource)
	}
	if q.Decision != "" {
		add("decision = $%d", q.Decision)
	}
	if q.Procedure != "" {
		add("procedure = $%d", q.Procedure)
	}
	if q.FreeText != "" {
		// k=v shape only in v1; richer JSONPath lands later.
		key, val, ok := strings.Cut(q.FreeText, "=")
		if !ok {
			return nil, errors.New("search: free_text must be 'key=value'")
		}
		add("metadata @> jsonb_build_object($%d, $", key)
		// We need TWO placeholders for the @> jsonb pair; the
		// `add` helper above only registers one. Patch the last
		// condition + push the second arg manually.
		conds[len(conds)-1] = fmt.Sprintf("metadata @> jsonb_build_object($%d, $%d)", len(args), len(args)+1)
		args = append(args, val)
	}
	// Keyset pagination.
	if !q.BeforeTime.IsZero() && q.BeforeID > 0 {
		args = append(args, q.BeforeTime.UTC(), q.BeforeID)
		conds = append(conds, fmt.Sprintf(
			"(event_time, id) < ($%d, $%d)", len(args)-1, len(args)))
	}

	sql := fmt.Sprintf(`
SELECT id, tenant_id, event_time, actor_user_id, actor_session_id,
       actor_client_ip, actor_user_agent, action, resource,
       decision, reason, matched_policy_id, procedure,
       classification, metadata
FROM audit_events
WHERE %s
ORDER BY event_time DESC, id DESC
LIMIT %d
`, strings.Join(conds, " AND "), limit+1)

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("search: query: %w", err)
	}
	defer rows.Close()

	var hits []Hit
	for rows.Next() {
		var (
			h           Hit
			actorUserID *string
		)
		if err := rows.Scan(
			&h.ID, &h.TenantID, &h.EventTime, &actorUserID, &h.ActorSessionID,
			&h.ActorClientIP, &h.ActorUserAgent, &h.Action, &h.Resource,
			&h.Decision, &h.Reason, &h.MatchedPolicyID, &h.Procedure,
			&h.Classification, &h.Metadata,
		); err != nil {
			return nil, fmt.Errorf("search: scan: %w", err)
		}
		if actorUserID != nil {
			h.ActorUserID = *actorUserID
		}
		hits = append(hits, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search: rows: %w", err)
	}

	res := &Result{}
	if len(hits) > limit {
		// We over-fetched by one to detect "is there a next page".
		last := hits[limit-1]
		res.NextCursor = &Cursor{BeforeTime: last.EventTime, BeforeID: last.ID}
		hits = hits[:limit]
	}
	res.Hits = hits
	return res, nil
}

// Stream walks every matching row WITHOUT a page cap and invokes
// the supplied callback per row. Used by the export pipeline so
// a 1M-row CSV download doesn't materialise the full slice in
// memory.
func (s *Service) Stream(ctx context.Context, q Query, fn func(Hit) error) error {
	if q.TenantID == "" {
		return errors.New("search: tenant_id is required")
	}
	conds := []string{"tenant_id = $1"}
	args := []any{q.TenantID}
	add := func(cond string, val any) {
		args = append(args, val)
		conds = append(conds, fmt.Sprintf(cond, len(args)))
	}
	if !q.Start.IsZero() {
		add("event_time >= $%d", q.Start.UTC())
	}
	if !q.End.IsZero() {
		add("event_time <= $%d", q.End.UTC())
	}
	sql := fmt.Sprintf(`
SELECT id, tenant_id, event_time, actor_user_id, actor_session_id,
       actor_client_ip, actor_user_agent, action, resource,
       decision, reason, matched_policy_id, procedure,
       classification, metadata
FROM audit_events
WHERE %s
ORDER BY event_time ASC, id ASC
`, strings.Join(conds, " AND "))

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("search: stream query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			h           Hit
			actorUserID *string
		)
		if err := rows.Scan(
			&h.ID, &h.TenantID, &h.EventTime, &actorUserID, &h.ActorSessionID,
			&h.ActorClientIP, &h.ActorUserAgent, &h.Action, &h.Resource,
			&h.Decision, &h.Reason, &h.MatchedPolicyID, &h.Procedure,
			&h.Classification, &h.Metadata,
		); err != nil {
			return fmt.Errorf("search: stream scan: %w", err)
		}
		if actorUserID != nil {
			h.ActorUserID = *actorUserID
		}
		if err := fn(h); err != nil {
			return err
		}
	}
	return rows.Err()
}

// ChainTipFor returns the (last_seq, last_hash) of the chain at
// the moment of the call. Embedded in export envelopes so the
// download is independently re-verifiable.
func (s *Service) ChainTipFor(ctx context.Context, tenantID string) (int64, string, error) {
	var (
		seq  int64
		hash string
	)
	err := s.pool.QueryRow(ctx,
		`SELECT last_seq, last_hash FROM chain_tip WHERE tenant_id = $1`,
		tenantID,
	).Scan(&seq, &hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, "", ErrTenantUnknown
	}
	if err != nil {
		return 0, "", err
	}
	return seq, hash, nil
}

// ErrTenantUnknown is returned when the chain_tip lookup finds
// no row for the supplied tenant.
var ErrTenantUnknown = errors.New("search: tenant unknown")
