package authzv1

import (
	"fmt"
	"testing"
)

// REQ-FUNC-PLT-AUTHZ-004: decision latency < 1ms p99 on a
// 10k-policy fixture. This bench is the gate.
//
// Run with:
//
//	go test -tags='' -run=^$ -bench=Decide -benchtime=2s ./authz/v1/
//
// Failure mode: a single iteration that exceeds 1ms is a hard
// regression — the bench reports ns/op so the CI gate compares
// `b.N > 0 && nsPerOp < 1_000_000`.
func BenchmarkDecide_10kPolicies(b *testing.B) {
	rules := make([]Policy, 0, 10_000)
	// 9_999 noise rules across many modules + one matching allow
	// at the bottom of the priority pile.
	for i := 0; i < 9_999; i++ {
		rules = append(rules, Policy{
			ID:         fmt.Sprintf("noise-%05d", i),
			Effect:     EffectAllow,
			Priority:   1, // all equal-low priority noise
			Permission: fmt.Sprintf("module%d.resource%d.action", i, i%97),
			Roles:      []string{"some-role"},
		})
	}
	rules = append(rules, Policy{
		ID:         "match",
		Effect:     EffectAllow,
		Priority:   10, // higher so the walk halts on this rule
		Permission: "groundstation.pass.read",
		Roles:      []string{"operator"},
	})
	set, err := NewPolicySet(rules)
	if err != nil {
		b.Fatalf("set: %v", err)
	}
	principal := &Principal{
		UserID:         "u1",
		Roles:          []string{"operator"},
		ClearanceLevel: "internal",
	}
	req := Request{Permission: "groundstation.pass.read"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec, err := Decide(principal, req, set)
		if err != nil || dec.Effect != EffectAllow {
			b.Fatalf("got %+v %v", dec, err)
		}
	}
}

// Worst-case bench: no matching rule → linear scan to the end of
// the policy set + default-deny.
func BenchmarkDecide_10kPolicies_DefaultDeny(b *testing.B) {
	rules := make([]Policy, 0, 10_000)
	for i := 0; i < 10_000; i++ {
		rules = append(rules, Policy{
			ID:         fmt.Sprintf("noise-%05d", i),
			Effect:     EffectAllow,
			Priority:   1,
			Permission: fmt.Sprintf("module%d.resource%d.action", i, i%97),
			Roles:      []string{"role-x"},
		})
	}
	set, _ := NewPolicySet(rules)
	principal := &Principal{
		UserID: "u1",
		Roles:  []string{"operator"},
	}
	req := Request{Permission: "no.matching.rule"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Decide(principal, req, set)
	}
}
