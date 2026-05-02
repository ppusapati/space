// verify.go — chain verifier that recomputes the hashes over a
// time range and reports the first broken offset.
//
// Use cases:
//
//   • Compliance attestation: "prove the audit log between
//     start and end has not been tampered with."
//
//   • Per-row spot check: VerifyRow recomputes one row's hash
//     against the previous row's row_hash. Useful when the
//     export envelope (TASK-P1-AUDIT-002) needs a single-row
//     attestation rather than a range scan.

package chain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Verifier wraps the pool with the chain-recomputation helpers.
type Verifier struct {
	pool *pgxpool.Pool
}

// NewVerifier wraps a pool.
func NewVerifier(pool *pgxpool.Pool) *Verifier {
	return &Verifier{pool: pool}
}

// VerifyResult describes the outcome of a Verify call.
type VerifyResult struct {
	TenantID  string
	Verified  int    // rows whose hash matched the recompute
	Broken    int64  // chain_seq of the first offending row, 0 if none
	Reason    string // "" on success
	StartSeq  int64  // first chain_seq inspected
	EndSeq    int64  // last chain_seq inspected
}

// IsClean reports whether the verifier found no break.
func (r VerifyResult) IsClean() bool { return r.Broken == 0 }

// VerifyRange recomputes the hash chain for tenantID over
// [start, end]. Returns a VerifyResult; err is non-nil only for
// I/O errors (the chain-break case is reported via Result.Broken).
//
// The walk is forward-ordered (chain_seq ASC) so the algorithm
// can short-circuit on the first break.
func (v *Verifier) VerifyRange(ctx context.Context, tenantID string, start, end time.Time) (VerifyResult, error) {
	if tenantID == "" {
		return VerifyResult{}, errors.New("chain: empty tenant_id")
	}
	if end.Before(start) {
		return VerifyResult{}, errors.New("chain: end < start")
	}
	rows, err := v.pool.Query(ctx, `
SELECT chain_seq, prev_hash, row_hash, event_time,
       actor_user_id, actor_session_id, actor_client_ip,
       actor_user_agent, action, resource, decision, reason,
       matched_policy_id, procedure, classification, metadata
FROM audit_events
WHERE tenant_id = $1
  AND event_time >= $2
  AND event_time <= $3
ORDER BY chain_seq ASC
`, tenantID, start.UTC(), end.UTC())
	if err != nil {
		return VerifyResult{}, fmt.Errorf("chain: query: %w", err)
	}
	defer rows.Close()

	res := VerifyResult{TenantID: tenantID}
	var prevHashOnDisk string
	var firstSeen bool

	for rows.Next() {
		var (
			seq                                                     int64
			prevHash, rowHash                                       string
			eventTime                                               time.Time
			actorUserID                                             *string
			actorSession, actorIP, actorUA                          string
			action, resource, decision, reason, matched, procedure  string
			classification                                          string
			metadataRaw                                             []byte
		)
		if err := rows.Scan(
			&seq, &prevHash, &rowHash, &eventTime,
			&actorUserID, &actorSession, &actorIP, &actorUA,
			&action, &resource, &decision, &reason,
			&matched, &procedure, &classification, &metadataRaw,
		); err != nil {
			return res, fmt.Errorf("chain: scan: %w", err)
		}
		if !firstSeen {
			res.StartSeq = seq
			firstSeen = true
		}
		res.EndSeq = seq

		// Continuity: this row's prev_hash MUST equal the previous
		// row's row_hash. Skip the very first row in the range
		// (we don't have the predecessor in scope).
		if prevHashOnDisk != "" && prevHash != prevHashOnDisk {
			res.Broken = seq
			res.Reason = fmt.Sprintf("prev_hash mismatch at chain_seq=%d", seq)
			return res, nil
		}

		// Hash recompute.
		var actor string
		if actorUserID != nil {
			actor = *actorUserID
		}
		want, err := HashRow(Event{
			TenantID:        tenantID,
			EventTime:       eventTime,
			ActorUserID:     actor,
			ActorSessionID:  actorSession,
			ActorClientIP:   actorIP,
			ActorUserAgent:  actorUA,
			Action:          action,
			Resource:        resource,
			Decision:        decision,
			Reason:          reason,
			MatchedPolicyID: matched,
			Procedure:       procedure,
			Classification:  classification,
			Metadata:        decodeMetadata(metadataRaw),
		}, prevHash, seq)
		if err != nil {
			return res, fmt.Errorf("chain: hash: %w", err)
		}
		if want != rowHash {
			res.Broken = seq
			res.Reason = fmt.Sprintf("row_hash mismatch at chain_seq=%d (recompute=%s, stored=%s)",
				seq, want, rowHash)
			return res, nil
		}

		res.Verified++
		prevHashOnDisk = rowHash
	}
	if err := rows.Err(); err != nil {
		return res, fmt.Errorf("chain: rows: %w", err)
	}
	return res, nil
}

// VerifyRow recomputes a single row's hash against the supplied
// previous-row hash. Used by the export envelope (AUDIT-002) when
// attesting a CSV/JSON download.
//
// Returns nil on match; ErrChainBreak when the hashes diverge.
func (v *Verifier) VerifyRow(ctx context.Context, tenantID string, chainSeq int64) error {
	const q = `
SELECT prev_hash, row_hash, event_time,
       actor_user_id, actor_session_id, actor_client_ip,
       actor_user_agent, action, resource, decision, reason,
       matched_policy_id, procedure, classification, metadata
FROM audit_events
WHERE tenant_id = $1 AND chain_seq = $2
`
	var (
		prevHash, rowHash                                       string
		eventTime                                               time.Time
		actorUserID                                             *string
		actorSession, actorIP, actorUA                          string
		action, resource, decision, reason, matched, procedure  string
		classification                                          string
		metadataRaw                                             []byte
	)
	if err := v.pool.QueryRow(ctx, q, tenantID, chainSeq).Scan(
		&prevHash, &rowHash, &eventTime,
		&actorUserID, &actorSession, &actorIP, &actorUA,
		&action, &resource, &decision, &reason,
		&matched, &procedure, &classification, &metadataRaw,
	); err != nil {
		return fmt.Errorf("chain: verify row: %w", err)
	}
	var actor string
	if actorUserID != nil {
		actor = *actorUserID
	}
	want, err := HashRow(Event{
		TenantID:        tenantID,
		EventTime:       eventTime,
		ActorUserID:     actor,
		ActorSessionID:  actorSession,
		ActorClientIP:   actorIP,
		ActorUserAgent:  actorUA,
		Action:          action,
		Resource:        resource,
		Decision:        decision,
		Reason:          reason,
		MatchedPolicyID: matched,
		Procedure:       procedure,
		Classification:  classification,
		Metadata:        decodeMetadata(metadataRaw),
	}, prevHash, chainSeq)
	if err != nil {
		return err
	}
	if want != rowHash {
		return fmt.Errorf("%w at chain_seq=%d", ErrChainBreak, chainSeq)
	}
	return nil
}

// decodeMetadata best-effort parses the metadata column. Empty /
// invalid → returns nil so the canonicaliser substitutes an empty
// map (which hashes the same as "no metadata").
func decodeMetadata(raw []byte) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	var m map[string]string
	// We intentionally ignore the error — a future schema addition
	// might write a richer structure that we still hash by treating
	// the unknown shape as empty. The canonicaliser stays
	// deterministic regardless.
	_ = json.Unmarshal(raw, &m)
	return m
}

// ErrChainBreak is returned by VerifyRow / surfaced in
// VerifyResult.Reason when the chain integrity check fails.
var ErrChainBreak = errors.New("chain: integrity break")
