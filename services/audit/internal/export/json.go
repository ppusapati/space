// json.go — JSON export. Streams rows as NDJSON (one event per
// line) wrapped in a signed envelope.
//
// Output shape:
//
//	{"envelope": {... signed ...}}
//	{"event": {...row 1...}}
//	{"event": {...row 2...}}
//	...
//
// The envelope is the FIRST line so a streaming consumer can
// hash-check it before processing any rows.

package export

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ppusapati/space/services/audit/internal/search"
)

// JSONExporter writes audit events as NDJSON with a signed
// envelope header.
type JSONExporter struct {
	search *search.Service
	pool   *pgxpool.Pool
	clk    func() time.Time
}

// NewJSONExporter wraps the search service + the pool. clock=nil → time.Now.
func NewJSONExporter(s *search.Service, pool *pgxpool.Pool, clock func() time.Time) (*JSONExporter, error) {
	if s == nil {
		return nil, errors.New("export: nil search service")
	}
	if pool == nil {
		return nil, errors.New("export: nil pool")
	}
	if clock == nil {
		clock = time.Now
	}
	return &JSONExporter{search: s, pool: pool, clk: clock}, nil
}

// Export streams the events matching `q` to `w` with a signed
// envelope header. Returns the envelope it stamped (which the
// HTTP handler can also expose via a response header).
func (j *JSONExporter) Export(ctx context.Context, q search.Query, w io.Writer) (*Envelope, error) {
	if q.TenantID == "" {
		return nil, errors.New("export: tenant_id is required")
	}

	// 1. Capture the chain tip + first-row attestation BEFORE
	// streaming rows so the envelope's row_count matches what
	// we end up writing.
	tipSeq, tipHash, err := j.search.ChainTipFor(ctx, q.TenantID)
	if err != nil {
		return nil, fmt.Errorf("export: chain tip: %w", err)
	}
	first, last, count, err := j.firstLastForRange(ctx, q)
	if err != nil {
		return nil, err
	}

	env := &Envelope{
		Format:        "json",
		TenantID:      q.TenantID,
		ExportedAt:    j.clk().UTC(),
		RowCount:      count,
		FirstChainSeq: first.ChainSeq,
		FirstRowHash:  first.RowHash,
		LastChainSeq:  last.ChainSeq,
		LastRowHash:   last.RowHash,
		ChainTipSeq:   tipSeq,
		ChainTipHash:  tipHash,
	}
	if err := env.Sign(); err != nil {
		return nil, err
	}

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(map[string]any{"envelope": env}); err != nil {
		return nil, fmt.Errorf("export: write envelope: %w", err)
	}

	if err := j.search.Stream(ctx, q, func(h search.Hit) error {
		return enc.Encode(map[string]any{"event": jsonHit(h)})
	}); err != nil {
		return nil, err
	}
	return env, nil
}

// jsonHit projects a search.Hit into a stable, JSON-friendly map.
func jsonHit(h search.Hit) map[string]any {
	return map[string]any{
		"id":                h.ID,
		"tenant_id":         h.TenantID,
		"event_time":        h.EventTime.UTC().Format(time.RFC3339Nano),
		"actor_user_id":     h.ActorUserID,
		"actor_session_id":  h.ActorSessionID,
		"actor_client_ip":   h.ActorClientIP,
		"actor_user_agent":  h.ActorUserAgent,
		"action":            h.Action,
		"resource":          h.Resource,
		"decision":          h.Decision,
		"reason":            h.Reason,
		"matched_policy_id": h.MatchedPolicyID,
		"procedure":         h.Procedure,
		"classification":    h.Classification,
		"metadata":          json.RawMessage(h.Metadata),
	}
}

// rowRef is the tiny struct firstLastForRange returns so the
// envelope can attest both ends of the range without a second
// scan.
type rowRef struct {
	ChainSeq int64
	RowHash  string
}

// firstLastForRange runs two cheap range-bounded queries to grab
// the first + last (chain_seq, row_hash) the export will cover,
// plus the count. Done as separate queries so the streaming
// scan can stay forward-only.
func (j *JSONExporter) firstLastForRange(ctx context.Context, q search.Query) (rowRef, rowRef, int, error) {
	var (
		first rowRef
		last  rowRef
		count int
	)
	const firstQ = `
SELECT chain_seq, row_hash FROM audit_events
WHERE tenant_id = $1
  AND ($2::timestamptz IS NULL OR event_time >= $2)
  AND ($3::timestamptz IS NULL OR event_time <= $3)
ORDER BY event_time ASC, id ASC LIMIT 1
`
	const lastQ = `
SELECT chain_seq, row_hash FROM audit_events
WHERE tenant_id = $1
  AND ($2::timestamptz IS NULL OR event_time >= $2)
  AND ($3::timestamptz IS NULL OR event_time <= $3)
ORDER BY event_time DESC, id DESC LIMIT 1
`
	const countQ = `
SELECT count(*) FROM audit_events
WHERE tenant_id = $1
  AND ($2::timestamptz IS NULL OR event_time >= $2)
  AND ($3::timestamptz IS NULL OR event_time <= $3)
`
	startArg := nullTime(q.Start)
	endArg := nullTime(q.End)
	if err := j.pool.QueryRow(ctx, firstQ, q.TenantID, startArg, endArg).Scan(&first.ChainSeq, &first.RowHash); err != nil {
		// Empty range — leave the row refs zero-valued and let the
		// envelope's row_count == 0 carry the meaning.
		return first, last, 0, nil
	}
	_ = j.pool.QueryRow(ctx, lastQ, q.TenantID, startArg, endArg).Scan(&last.ChainSeq, &last.RowHash)
	_ = j.pool.QueryRow(ctx, countQ, q.TenantID, startArg, endArg).Scan(&count)
	return first, last, count, nil
}

// nullTime wraps a possibly-zero time so the SQL `$2::timestamptz IS NULL`
// guard works correctly. pgx treats a zero time as zero-value, NOT
// NULL, so we explicitly hand it nil for the "no bound" case.
func nullTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.UTC()
}

// EncodeRowHashSummary returns a compact "first→last" string for
// log lines. Used by tests + the future SOC2 evidence harness.
func EncodeRowHashSummary(env *Envelope) string {
	return fmt.Sprintf("rows=%d first_seq=%d first=%s last_seq=%d last=%s",
		env.RowCount, env.FirstChainSeq, short(env.FirstRowHash),
		env.LastChainSeq, short(env.LastRowHash))
}

func short(h string) string {
	if len(h) <= 8 {
		return h
	}
	// First 8 hex chars give 32 bits — plenty for visual checks.
	return hex.EncodeToString([]byte(h[:4]))
}
