// Package export emits CSV + JSON downloads of audit events
// wrapped in a signed envelope so the recipient can independently
// re-verify the chain (REQ-FUNC-PLT-AUDIT-004 acceptance #2).
//
// The envelope shape:
//
//	{
//	  "format":          "json" | "csv",
//	  "tenant_id":       "...",
//	  "exported_at":     "RFC 3339 nanos",
//	  "row_count":       N,
//	  "first_chain_seq": <seq of first row>,
//	  "last_chain_seq":  <seq of last row>,
//	  "first_row_hash":  "<hex SHA-256>",
//	  "last_row_hash":   "<hex SHA-256>",
//	  "chain_tip_seq":   <chain_tip.last_seq AT export time>,
//	  "chain_tip_hash":  "<chain_tip.last_hash AT export time>",
//	  "envelope_hash":   "<hex SHA-256 of the canonical envelope sans envelope_hash>"
//	}
//
// envelope_hash is the SHA-256 of the JSON-canonical bytes of
// every other field. A consumer recomputes that hash + (if they
// wish to confirm chain continuity) walks the included row hashes
// against their own re-fetch from the audit service.

package export

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"time"
)

// Envelope is the per-export attestation header.
type Envelope struct {
	Format         string    `json:"format"`
	TenantID       string    `json:"tenant_id"`
	ExportedAt     time.Time `json:"exported_at"`
	RowCount       int       `json:"row_count"`
	FirstChainSeq  int64     `json:"first_chain_seq"`
	LastChainSeq   int64     `json:"last_chain_seq"`
	FirstRowHash   string    `json:"first_row_hash"`
	LastRowHash    string    `json:"last_row_hash"`
	ChainTipSeq    int64     `json:"chain_tip_seq"`
	ChainTipHash   string    `json:"chain_tip_hash"`
	EnvelopeHash   string    `json:"envelope_hash"`
}

// Sign computes EnvelopeHash from the canonical bytes of every
// other field and stamps it onto the envelope. Idempotent.
func (e *Envelope) Sign() error {
	body, err := canonicaliseEnvelope(e)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(body)
	e.EnvelopeHash = hex.EncodeToString(sum[:])
	return nil
}

// Verify recomputes EnvelopeHash and asserts it matches the
// supplied envelope's existing field. Used by the export-test +
// the future consumer-side attestation tooling.
func (e Envelope) Verify() error {
	if e.EnvelopeHash == "" {
		return errors.New("export: envelope is unsigned")
	}
	want := e.EnvelopeHash
	clone := e
	clone.EnvelopeHash = ""
	body, err := canonicaliseEnvelope(&clone)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(body)
	got := hex.EncodeToString(sum[:])
	if want != got {
		return errors.New("export: envelope hash mismatch")
	}
	return nil
}

// canonicaliseEnvelope returns the lex-key-ordered JSON bytes of
// the envelope with envelope_hash forcibly stripped.
func canonicaliseEnvelope(e *Envelope) ([]byte, error) {
	pairs := map[string]any{
		"chain_tip_hash":   e.ChainTipHash,
		"chain_tip_seq":    e.ChainTipSeq,
		"exported_at":      e.ExportedAt.UTC().Format(time.RFC3339Nano),
		"first_chain_seq":  e.FirstChainSeq,
		"first_row_hash":   e.FirstRowHash,
		"format":           e.Format,
		"last_chain_seq":   e.LastChainSeq,
		"last_row_hash":    e.LastRowHash,
		"row_count":        e.RowCount,
		"tenant_id":        e.TenantID,
	}
	keys := make([]string, 0, len(pairs))
	for k := range pairs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]byte, 0, 256)
	out = append(out, '{')
	for i, k := range keys {
		if i > 0 {
			out = append(out, ',')
		}
		kb, _ := json.Marshal(k)
		vb, err := json.Marshal(pairs[k])
		if err != nil {
			return nil, err
		}
		out = append(out, kb...)
		out = append(out, ':')
		out = append(out, vb...)
	}
	out = append(out, '}')
	return out, nil
}
