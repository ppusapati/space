package export

import (
	"strings"
	"testing"
	"time"
)

func sampleEnvelope() *Envelope {
	return &Envelope{
		Format:        "json",
		TenantID:      "00000000-0000-0000-0000-000000000001",
		ExportedAt:    time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
		RowCount:      42,
		FirstChainSeq: 1,
		LastChainSeq:  42,
		FirstRowHash:  strings.Repeat("a", 64),
		LastRowHash:   strings.Repeat("b", 64),
		ChainTipSeq:   42,
		ChainTipHash:  strings.Repeat("c", 64),
	}
}

func TestEnvelope_Sign_PopulatesHash(t *testing.T) {
	e := sampleEnvelope()
	if err := e.Sign(); err != nil {
		t.Fatalf("sign: %v", err)
	}
	if len(e.EnvelopeHash) != 64 {
		t.Errorf("hash len: %d", len(e.EnvelopeHash))
	}
}

func TestEnvelope_VerifyRoundtrip(t *testing.T) {
	e := sampleEnvelope()
	if err := e.Sign(); err != nil {
		t.Fatalf("sign: %v", err)
	}
	if err := e.Verify(); err != nil {
		t.Errorf("verify: %v", err)
	}
}

func TestEnvelope_VerifyDetectsTamper(t *testing.T) {
	e := sampleEnvelope()
	_ = e.Sign()
	tampered := *e
	tampered.RowCount++ // change one field
	if err := tampered.Verify(); err == nil {
		t.Error("expected verify to fail on tampered envelope")
	}
}

func TestEnvelope_VerifyRejectsUnsigned(t *testing.T) {
	e := sampleEnvelope()
	if err := e.Verify(); err == nil {
		t.Error("expected error for unsigned envelope")
	}
}

func TestEnvelope_DeterministicSign(t *testing.T) {
	a := sampleEnvelope()
	b := sampleEnvelope()
	if err := a.Sign(); err != nil {
		t.Fatalf("a sign: %v", err)
	}
	if err := b.Sign(); err != nil {
		t.Fatalf("b sign: %v", err)
	}
	if a.EnvelopeHash != b.EnvelopeHash {
		t.Errorf("not deterministic: %q vs %q", a.EnvelopeHash, b.EnvelopeHash)
	}
}

func TestEnvelope_HashChangesPerField(t *testing.T) {
	base := sampleEnvelope()
	_ = base.Sign()

	mutators := []struct {
		name string
		mut  func(*Envelope)
	}{
		{"format", func(e *Envelope) { e.Format = "csv" }},
		{"tenant_id", func(e *Envelope) { e.TenantID = "x" }},
		{"row_count", func(e *Envelope) { e.RowCount++ }},
		{"first_chain_seq", func(e *Envelope) { e.FirstChainSeq++ }},
		{"last_chain_seq", func(e *Envelope) { e.LastChainSeq++ }},
		{"first_row_hash", func(e *Envelope) { e.FirstRowHash = strings.Repeat("z", 64) }},
		{"chain_tip_seq", func(e *Envelope) { e.ChainTipSeq++ }},
		{"exported_at", func(e *Envelope) { e.ExportedAt = e.ExportedAt.Add(time.Second) }},
	}
	for _, m := range mutators {
		t.Run(m.name, func(t *testing.T) {
			clone := sampleEnvelope()
			m.mut(clone)
			_ = clone.Sign()
			if clone.EnvelopeHash == base.EnvelopeHash {
				t.Errorf("%s should change envelope hash", m.name)
			}
		})
	}
}
