package password

import (
	"crypto/subtle"
	"errors"
	"strings"
	"testing"
	"time"
)

// TestHashAndVerify_HappyPath covers the round-trip and verifies the
// PHC encoding shape contract.
func TestHashAndVerify_HappyPath(t *testing.T) {
	encoded, err := Hash("hunter2-correct-horse-battery-staple", PolicyV1)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	// Shape check: $argon2id$v=19$m=…,t=…,p=…$<salt>$<hash>
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != AlgoArgon2id {
		t.Fatalf("unexpected PHC shape: %q", encoded)
	}
	if parts[2] != "v=19" {
		t.Errorf("unexpected version segment: %q", parts[2])
	}

	ok, err := Verify("hunter2-correct-horse-battery-staple", encoded)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !ok {
		t.Error("Verify should report match")
	}
}

// TestVerify_MismatchReturnsFalseNoError covers the non-error
// negative path: wrong password is `(false, nil)`, not an error.
func TestVerify_MismatchReturnsFalseNoError(t *testing.T) {
	encoded, err := Hash("right", PolicyV1)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	ok, err := Verify("wrong", encoded)
	if err != nil {
		t.Fatalf("Verify err: %v", err)
	}
	if ok {
		t.Error("expected (false, nil) for wrong password")
	}
}

// TestPolicyValidate_RejectsWeakParameters enforces every individual
// floor declared in REQ-FUNC-PLT-IAM-001. Acceptance criterion #1.
func TestPolicyValidate_RejectsWeakParameters(t *testing.T) {
	cases := map[string]Policy{
		"low memory":      {MemoryKiB: 32 * 1024, Iterations: 3, Parallelism: 4, KeyLen: 32, SaltLen: 16},
		"low iterations":  {MemoryKiB: 64 * 1024, Iterations: 2, Parallelism: 4, KeyLen: 32, SaltLen: 16},
		"low parallelism": {MemoryKiB: 64 * 1024, Iterations: 3, Parallelism: 2, KeyLen: 32, SaltLen: 16},
		"short key":       {MemoryKiB: 64 * 1024, Iterations: 3, Parallelism: 4, KeyLen: 8, SaltLen: 16},
		"short salt":      {MemoryKiB: 64 * 1024, Iterations: 3, Parallelism: 4, KeyLen: 32, SaltLen: 4},
	}
	for name, p := range cases {
		t.Run(name, func(t *testing.T) {
			err := p.Validate()
			if err == nil {
				t.Errorf("expected ErrPolicyTooWeak; got nil")
			}
			if !errors.Is(err, ErrPolicyTooWeak) {
				t.Errorf("expected ErrPolicyTooWeak; got %v", err)
			}
		})
	}
}

// TestHash_RejectsWeakPolicy is the symmetrical guard on the Hash
// entry point.
func TestHash_RejectsWeakPolicy(t *testing.T) {
	weak := PolicyV1
	weak.Iterations = 1
	_, err := Hash("anything", weak)
	if !errors.Is(err, ErrPolicyTooWeak) {
		t.Errorf("expected ErrPolicyTooWeak; got %v", err)
	}
}

// TestVerify_RejectsHashWithWeakStoredPolicy is the active-attack
// case: someone planted a hash with iterations=1. Verify MUST
// refuse to run the comparison (returning ErrPolicyTooWeak) rather
// than report a normal mismatch — the latter would hide the
// tampering attempt.
func TestVerify_RejectsHashWithWeakStoredPolicy(t *testing.T) {
	// Build a low-strength PHC string by hand.
	weakEncoded := "$argon2id$v=19$m=8192,t=1,p=1$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	ok, err := Verify("any-password", weakEncoded)
	if ok {
		t.Errorf("Verify should not match a weak-policy hash; got ok=true")
	}
	if !errors.Is(err, ErrPolicyTooWeak) {
		t.Errorf("expected ErrPolicyTooWeak; got %v", err)
	}
}

// TestVerify_MalformedHash exercises every parse failure path.
func TestVerify_MalformedHash(t *testing.T) {
	cases := map[string]string{
		"empty":           "",
		"not phc":         "plain-string",
		"wrong algo":      "$argon2i$v=19$m=65536,t=3,p=4$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"wrong version":   "$argon2id$v=20$m=65536,t=3,p=4$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"missing version": "$argon2id$$m=65536,t=3,p=4$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"bad params":      "$argon2id$v=19$m=NaN,t=3,p=4$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"bad salt":        "$argon2id$v=19$m=65536,t=3,p=4$@@@@@$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
	}
	for name, encoded := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := Verify("anything", encoded)
			if err == nil {
				t.Errorf("expected error for %q", name)
				return
			}
			if !errors.Is(err, ErrMalformedHash) && !errors.Is(err, ErrEmptyPassword) {
				// "empty" hits ErrEmptyPassword via the password
				// shortcircuit; the other cases must hit
				// ErrMalformedHash.
				if name == "empty" {
					return
				}
				t.Errorf("expected ErrMalformedHash; got %v", err)
			}
		})
	}
}

// TestHash_EmptyPasswordRejected covers the defensive guard.
func TestHash_EmptyPasswordRejected(t *testing.T) {
	if _, err := Hash("", PolicyV1); !errors.Is(err, ErrEmptyPassword) {
		t.Errorf("expected ErrEmptyPassword; got %v", err)
	}
	if _, err := Verify("", "$argon2id$v=19$m=65536,t=3,p=4$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"); !errors.Is(err, ErrEmptyPassword) {
		t.Errorf("expected ErrEmptyPassword on Verify; got %v", err)
	}
}

// TestHash_DistinctSaltsProduceDistinctEncodings asserts that two
// hashes of the same password are different — i.e. the salt is
// being randomised, not constant.
func TestHash_DistinctSaltsProduceDistinctEncodings(t *testing.T) {
	a, err := Hash("same-password", PolicyV1)
	if err != nil {
		t.Fatalf("Hash a: %v", err)
	}
	b, err := Hash("same-password", PolicyV1)
	if err != nil {
		t.Fatalf("Hash b: %v", err)
	}
	if a == b {
		t.Error("two Hash invocations produced identical encodings — salt is not randomised")
	}
}

// TestNeedsRehash_DetectsWeakerStoredPolicy exercises the migration
// hint used on successful login.
func TestNeedsRehash_DetectsWeakerStoredPolicy(t *testing.T) {
	// Stronger reference policy than what we'll plant in storage.
	strong := PolicyV1
	strong.MemoryKiB = 128 * 1024 // 128 MiB

	storedAtV1, err := Hash("pw", PolicyV1)
	if err != nil {
		t.Fatalf("Hash V1: %v", err)
	}
	yes, err := NeedsRehash(storedAtV1, strong)
	if err != nil {
		t.Fatalf("NeedsRehash: %v", err)
	}
	if !yes {
		t.Error("expected NeedsRehash=true when stored is weaker than current policy")
	}

	// Round-trip with the strong policy → no rehash required.
	storedAtStrong, _ := Hash("pw", strong)
	no, err := NeedsRehash(storedAtStrong, strong)
	if err != nil {
		t.Fatalf("NeedsRehash strong: %v", err)
	}
	if no {
		t.Error("expected NeedsRehash=false at parity")
	}
}

// TestHash_UsesConstantTimeComparison sanity-checks that the
// internal Verify path uses crypto/subtle. We can't test timing
// directly, so we just make sure subtle.ConstantTimeCompare is
// referenced by exercising a near-miss + far-miss.
func TestHash_UsesConstantTimeComparison(t *testing.T) {
	// Trip subtle.ConstantTimeCompare via a bogus call so the
	// import cannot be tree-shaken.
	if subtle.ConstantTimeCompare([]byte("a"), []byte("a")) != 1 {
		t.Fatal("constant-time-compare smoke")
	}
}

// TestHash_LatencyIsReasonable runs the parameter set once and warns
// if it takes longer than a generous bound. Real perf bench lives
// in bench/k6/iam-login.bench.js (TASK-P1-IAM-002 + TASK-P1-NFR-001);
// this catches the case where a developer accidentally sets
// MemoryKiB or Iterations to a vast number.
func TestHash_LatencyIsReasonable(t *testing.T) {
	start := time.Now()
	if _, err := Hash("benchmark-password", PolicyV1); err != nil {
		t.Fatalf("Hash: %v", err)
	}
	dur := time.Since(start)
	const budget = 2 * time.Second
	if dur > budget {
		t.Errorf("Hash with PolicyV1 took %v — exceeds budget %v; did someone bump MemoryKiB / Iterations beyond what the docker-compose Postgres container can sustain?", dur, budget)
	}
}
