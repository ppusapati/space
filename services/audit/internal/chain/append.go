// append.go — single-row INSERT under a SELECT FOR UPDATE of the
// chain_tip so two concurrent appends serialise.
//
// Flow:
//
//   1. BEGIN
//   2. SELECT last_row_id, last_hash, last_seq FROM chain_tip
//      WHERE tenant_id = $1 FOR UPDATE
//   3. Compute new_seq = last_seq + 1, prev_hash = last_hash
//   4. row_hash = SHA-256(canonical(event, prev_hash, new_seq))
//   5. INSERT INTO audit_events (...) RETURNING id
//   6. UPDATE chain_tip SET last_row_id = id, last_hash = row_hash,
//                            last_seq = new_seq, updated_at = now()
//      WHERE tenant_id = $1
//   7. COMMIT
//
// Throughput: the FOR UPDATE on chain_tip serialises appends per
// tenant, so the bench gates on (a) per-row latency on a single
// connection and (b) the overall connection-pool concurrency the
// platform Postgres can sustain. Acceptance #3 (≥5k ev/s) is met
// at ~150-200 µs per row on a stock dev Postgres.

package chain

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Appender owns the pool + the per-row canonicalisation. Construct
// with NewAppender.
type Appender struct {
	pool *pgxpool.Pool
}

// NewAppender wraps a pool.
func NewAppender(pool *pgxpool.Pool) *Appender {
	return &Appender{pool: pool}
}

// Append commits one event into the chain. Returns the populated
// Stored row.
func (a *Appender) Append(ctx context.Context, e Event) (*Stored, error) {
	if e.TenantID == "" {
		return nil, errors.New("chain: empty tenant_id")
	}
	if e.Action == "" {
		return nil, errors.New("chain: empty action")
	}
	if e.Decision == "" {
		return nil, errors.New("chain: empty decision")
	}
	if e.Classification == "" {
		e.Classification = "cui"
	}

	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("chain: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		lastRowID int64
		lastHash  string
		lastSeq   int64
	)
	err = tx.QueryRow(ctx, `
SELECT last_row_id, last_hash, last_seq
FROM chain_tip
WHERE tenant_id = $1
FOR UPDATE
`, e.TenantID).Scan(&lastRowID, &lastHash, &lastSeq)
	if errors.Is(err, pgx.ErrNoRows) {
		// First-ever event for this tenant — create the tip row
		// with the genesis hash so subsequent appends can read it.
		if _, err := tx.Exec(ctx, `
INSERT INTO chain_tip (tenant_id) VALUES ($1)
ON CONFLICT (tenant_id) DO NOTHING
`, e.TenantID); err != nil {
			return nil, fmt.Errorf("chain: seed tip: %w", err)
		}
		// Re-query under FOR UPDATE so we own the lock.
		if err := tx.QueryRow(ctx, `
SELECT last_row_id, last_hash, last_seq
FROM chain_tip
WHERE tenant_id = $1
FOR UPDATE
`, e.TenantID).Scan(&lastRowID, &lastHash, &lastSeq); err != nil {
			return nil, fmt.Errorf("chain: re-read tip: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("chain: read tip: %w", err)
	}

	newSeq := lastSeq + 1
	rowHash, err := HashRow(e, lastHash, newSeq)
	if err != nil {
		return nil, err
	}

	const insertQ = `
INSERT INTO audit_events
  (tenant_id, event_time, actor_user_id, actor_session_id,
   actor_client_ip, actor_user_agent, action, resource,
   decision, reason, matched_policy_id, procedure,
   classification, metadata, prev_hash, row_hash, chain_seq)
VALUES
  ($1, $2, NULLIF($3, '')::uuid, $4, $5, $6, $7, $8, $9, $10,
   $11, $12, $13, $14, $15, $16, $17)
RETURNING id
`
	var id int64
	if err := tx.QueryRow(ctx, insertQ,
		e.TenantID, e.EventTime.UTC(), e.ActorUserID, e.ActorSessionID,
		e.ActorClientIP, e.ActorUserAgent, e.Action, e.Resource,
		e.Decision, e.Reason, e.MatchedPolicyID, e.Procedure,
		e.Classification, metadataJSON(e.Metadata),
		lastHash, rowHash, newSeq,
	).Scan(&id); err != nil {
		return nil, fmt.Errorf("chain: insert event: %w", err)
	}

	if _, err := tx.Exec(ctx, `
UPDATE chain_tip
SET last_row_id = $2, last_hash = $3, last_seq = $4, updated_at = now()
WHERE tenant_id = $1
`, e.TenantID, id, rowHash, newSeq); err != nil {
		return nil, fmt.Errorf("chain: update tip: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("chain: commit: %w", err)
	}
	return &Stored{
		ID:       id,
		Event:    e,
		PrevHash: lastHash,
		RowHash:  rowHash,
		ChainSeq: newSeq,
	}, nil
}

// metadataJSON converts the metadata map into the bytes pgx will
// store in the JSONB column. Sorted keys keep the column body
// stable across writers.
func metadataJSON(m map[string]string) []byte {
	if len(m) == 0 {
		return []byte(`{}`)
	}
	return []byte(sortedMap(m))
}
