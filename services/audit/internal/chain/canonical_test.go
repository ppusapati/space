package chain

import (
	"strings"
	"testing"
	"time"
)

func sampleEvent() Event {
	return Event{
		TenantID:        "00000000-0000-0000-0000-000000000001",
		EventTime:       time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
		ActorUserID:     "11111111-1111-1111-1111-111111111111",
		ActorSessionID:  "sess-1",
		ActorClientIP:   "10.0.0.1",
		ActorUserAgent:  "chetana-test/1.0",
		Action:          "iam.user.read",
		Resource:        "user-42",
		Decision:        "allow",
		Reason:          "allowed_by_rule",
		MatchedPolicyID: "ops-pass-read",
		Procedure:       "/iam.v1.AuthService/Login",
		Classification:  "cui",
		Metadata:        map[string]string{"k1": "v1", "k0": "v0"},
	}
}

func TestCanonicalise_Deterministic(t *testing.T) {
	e := sampleEvent()
	a, err := Canonicalise(e, GenesisHash, 1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	b, err := Canonicalise(e, GenesisHash, 1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if string(a) != string(b) {
		t.Errorf("not deterministic")
	}
}

func TestCanonicalise_KeyOrderStable(t *testing.T) {
	body, _ := Canonicalise(sampleEvent(), GenesisHash, 1)
	s := string(body)
	// keys must appear in lexicographic order. Check a known-safe
	// ordering by indexOf comparisons.
	pairs := []struct{ a, b string }{
		{`"action"`, `"actor_client_ip"`},
		{`"actor_user_agent"`, `"actor_user_id"`},
		{`"chain_seq"`, `"classification"`},
		{`"matched_policy_id"`, `"metadata"`},
		{`"prev_hash"`, `"procedure"`},
		{`"resource"`, `"tenant_id"`},
	}
	for _, p := range pairs {
		ai := strings.Index(s, p.a)
		bi := strings.Index(s, p.b)
		if ai < 0 || bi < 0 {
			t.Errorf("missing key: %s or %s in %s", p.a, p.b, s)
		}
		if ai >= bi {
			t.Errorf("key order: %s should precede %s; got %d vs %d", p.a, p.b, ai, bi)
		}
	}
}

func TestCanonicalise_MetadataKeyOrderStable(t *testing.T) {
	e := sampleEvent()
	e.Metadata = map[string]string{"z": "z", "a": "a", "m": "m"}
	a, _ := Canonicalise(e, GenesisHash, 1)
	for i := 0; i < 5; i++ {
		b, _ := Canonicalise(e, GenesisHash, 1)
		if string(a) != string(b) {
			t.Fatalf("iteration %d: metadata order unstable", i)
		}
	}
	// And the canonical bytes must show keys a, m, z in lex order.
	s := string(a)
	ai := strings.Index(s, `"a":"a"`)
	mi := strings.Index(s, `"m":"m"`)
	zi := strings.Index(s, `"z":"z"`)
	if !(ai < mi && mi < zi) {
		t.Errorf("metadata key order: a@%d m@%d z@%d", ai, mi, zi)
	}
}

func TestHashRow_DifferentEventsDiffer(t *testing.T) {
	a, _ := HashRow(sampleEvent(), GenesisHash, 1)
	e2 := sampleEvent()
	e2.Action = "iam.user.write" // change a single field
	b, _ := HashRow(e2, GenesisHash, 1)
	if a == b {
		t.Error("hash should differ for different events")
	}
	if len(a) != 64 {
		t.Errorf("hex len: %d want 64", len(a))
	}
}

func TestHashRow_PrevHashAffectsResult(t *testing.T) {
	a, _ := HashRow(sampleEvent(), GenesisHash, 1)
	b, _ := HashRow(sampleEvent(), strings.Repeat("a", 64), 1)
	if a == b {
		t.Error("prev_hash should be part of the hash")
	}
}

func TestHashRow_ChainSeqAffectsResult(t *testing.T) {
	a, _ := HashRow(sampleEvent(), GenesisHash, 1)
	b, _ := HashRow(sampleEvent(), GenesisHash, 2)
	if a == b {
		t.Error("chain_seq should be part of the hash")
	}
}

func TestHashRow_TimestampNanoPreserved(t *testing.T) {
	e1 := sampleEvent()
	e2 := sampleEvent()
	e2.EventTime = e1.EventTime.Add(1 * time.Nanosecond)
	a, _ := HashRow(e1, GenesisHash, 1)
	b, _ := HashRow(e2, GenesisHash, 1)
	if a == b {
		t.Error("nanosecond drift should change the hash")
	}
}

func TestCanonicalise_NilMetadataTreatedAsEmpty(t *testing.T) {
	e1 := sampleEvent()
	e1.Metadata = nil
	e2 := sampleEvent()
	e2.Metadata = map[string]string{}
	a, _ := HashRow(e1, GenesisHash, 1)
	b, _ := HashRow(e2, GenesisHash, 1)
	if a != b {
		t.Errorf("nil + empty metadata should hash the same")
	}
}

func TestGenesisHash_AllZero(t *testing.T) {
	if GenesisHash != strings.Repeat("0", 64) {
		t.Errorf("genesis: %q", GenesisHash)
	}
}
