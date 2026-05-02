// Package chain implements the chetana audit hash chain.
//
// → REQ-FUNC-PLT-AUDIT-001 (append-only, hash-chained at the
//                            row level so single-row tampering
//                            is detectable).
// → REQ-FUNC-PLT-AUDIT-002 (verifier recomputes the chain over
//                            a time range and reports the first
//                            broken offset).
// → REQ-NFR-OBS-004; design.md §4.2.

package chain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// Event is the canonical shape of one audit row that the chain
// hashes. The field set is fixed; new optional facets land in
// Metadata so the chain hash stays compatible with rows minted
// before the new field existed.
type Event struct {
	TenantID         string            `json:"tenant_id"`
	EventTime        time.Time         `json:"event_time"`
	ActorUserID      string            `json:"actor_user_id"`
	ActorSessionID   string            `json:"actor_session_id"`
	ActorClientIP    string            `json:"actor_client_ip"`
	ActorUserAgent   string            `json:"actor_user_agent"`
	Action           string            `json:"action"`
	Resource         string            `json:"resource"`
	Decision         string            `json:"decision"`
	Reason           string            `json:"reason"`
	MatchedPolicyID  string            `json:"matched_policy_id"`
	Procedure        string            `json:"procedure"`
	Classification   string            `json:"classification"`
	Metadata         map[string]string `json:"metadata"`
}

// Stored mirrors the audit_events row shape; Append populates
// PrevHash / RowHash / ChainSeq from the chain_tip and the
// canonical bytes.
type Stored struct {
	ID       int64
	Event    Event
	PrevHash string
	RowHash  string
	ChainSeq int64
}

// Canonicalise returns the deterministic JSON encoding of e plus
// the supplied prev_hash / chain_seq. The serialiser:
//
//   • Emits keys in lexicographic order at every nesting level.
//   • Escapes only the JSON-mandatory characters (no HTML
//     escaping, no unicode escaping for printable code points).
//   • Encodes the timestamp as RFC 3339 nanoseconds in UTC so
//     wall-clock drift does not change the hash.
//   • Includes prev_hash + chain_seq inside the hashed payload
//     so a reorder-and-replay attack on a stolen row is caught
//     by the next row's prev_hash check.
//
// The returned bytes are what we SHA-256 to produce row_hash.
func Canonicalise(e Event, prevHash string, chainSeq int64) ([]byte, error) {
	if e.Metadata == nil {
		e.Metadata = map[string]string{}
	}
	out := map[string]any{
		"action":            e.Action,
		"actor_client_ip":   e.ActorClientIP,
		"actor_session_id":  e.ActorSessionID,
		"actor_user_agent":  e.ActorUserAgent,
		"actor_user_id":     e.ActorUserID,
		"chain_seq":         chainSeq,
		"classification":    e.Classification,
		"decision":          e.Decision,
		"event_time":        e.EventTime.UTC().Format(time.RFC3339Nano),
		"matched_policy_id": e.MatchedPolicyID,
		"metadata":          sortedMap(e.Metadata),
		"prev_hash":         prevHash,
		"procedure":         e.Procedure,
		"reason":            e.Reason,
		"resource":          e.Resource,
		"tenant_id":         e.TenantID,
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(out); err != nil {
		return nil, fmt.Errorf("chain: canonicalise: %w", err)
	}
	// json.Encoder appends a trailing newline; strip it so the
	// hash is byte-stable across encoders.
	body := bytes.TrimRight(buf.Bytes(), "\n")
	return body, nil
}

// HashRow returns the hex SHA-256 of the canonicalised event.
func HashRow(e Event, prevHash string, chainSeq int64) (string, error) {
	body, err := Canonicalise(e, prevHash, chainSeq)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:]), nil
}

// GenesisHash is the prev_hash for the very first row of a chain
// (chain_seq=1). Stable, all-zero hex.
const GenesisHash = "0000000000000000000000000000000000000000000000000000000000000000"

// sortedMap returns a copy of m with stable iteration order. The
// JSON encoder emits keys in iteration order for map[string]any,
// so sorting at this layer is what makes the canonical bytes
// deterministic.
//
// Returns map[string]string (the only metadata shape we currently
// support); the JSON encoder still iterates lexicographically
// because the source map's keys are sorted into a slice that the
// encoder then re-flattens via reflect, but we stage the sort by
// returning a json.RawMessage of the body that's already sorted.
func sortedMap(m map[string]string) json.RawMessage {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		kb, _ := json.Marshal(k)
		vb, _ := json.Marshal(m[k])
		buf.Write(kb)
		buf.WriteByte(':')
		buf.Write(vb)
	}
	buf.WriteByte('}')
	return buf.Bytes()
}
