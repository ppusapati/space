// Package password implements the platform's password hashing
// contract: Argon2id with parameters meeting REQ-FUNC-PLT-IAM-001
// (memory ≥ 64 MiB, iterations ≥ 3, parallelism ≥ 4) and PHC-string
// encoding so a future algorithm migration can co-exist with the
// current set in the same column.
//
// Hash format (PHC string per draft-ietf-irtf-cfrg-argon2-13 Annex A):
//
//	$argon2id$v=19$m=65536,t=3,p=4$<base64-salt>$<base64-hash>
//
// The Hash function returns this exact shape; Verify parses and
// re-derives. Both functions reject parameters below the policy
// minimums so an attacker who plants a weak hash through SQL
// injection cannot survive a Verify call.
package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Policy is the per-deployment Argon2id parameter set. Values are
// fixed in the variant constants below so service code cannot weaken
// them by accident.
//
// Memory is in KiB (Argon2 calls this `m`). Iterations is the time
// parameter `t`. Parallelism is `p`. KeyLen is the derived hash
// length in bytes; SaltLen is the random salt length.
type Policy struct {
	MemoryKiB   uint32
	Iterations  uint32
	Parallelism uint8
	KeyLen      uint32
	SaltLen     uint32
}

// PolicyV1 is the Phase-1 baseline.
//
// Numbers are the lower bounds in REQ-FUNC-PLT-IAM-001 (memory ≥
// 64 MiB, iterations ≥ 3, parallelism ≥ 4). These can be re-tuned in
// a follow-up after the IAM bench (TASK-P1-IAM-002 includes
// iam-login.bench.js) reports actual hash latency on the production
// hardware profile; downward changes are forbidden by Validate().
var PolicyV1 = Policy{
	MemoryKiB:   64 * 1024, // 64 MiB
	Iterations:  3,
	Parallelism: 4,
	KeyLen:      32,
	SaltLen:     16,
}

// Algo names — only argon2id supported in v1. Stored in the
// password_algo column so a v2 migration can add a sibling algo
// without breaking existing rows.
const (
	AlgoArgon2id = "argon2id"
)

// Argon2id version supported by golang.org/x/crypto/argon2.
const argon2Version = argon2.Version

// Hash derives an Argon2id hash of `password` using `policy` and
// returns the PHC-encoded string suitable for storage in
// users.password_hash.
//
// Errors:
//   - ErrPolicyTooWeak when policy violates the REQ-FUNC-PLT-IAM-001
//     minimums.
//   - ErrEmptyPassword when password is empty (callers should reject
//     these before reaching the hash function — guard rail).
func Hash(password string, policy Policy) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}
	if err := policy.Validate(); err != nil {
		return "", err
	}
	salt := make([]byte, policy.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("password: read salt: %w", err)
	}
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		policy.Iterations,
		policy.MemoryKiB,
		policy.Parallelism,
		policy.KeyLen,
	)
	return encodePHC(policy, salt, hash), nil
}

// Verify reports whether `password` matches the supplied PHC-encoded
// hash. Comparison is constant-time. The function also re-validates
// the parameters embedded in the hash — a hash with weaker params
// than PolicyV1 is treated as invalid even if the hash bytes match,
// so an attacker who replaces a row with a 1-iteration hash gets
// rejected.
//
// Returns:
//   - (true, nil)  — match.
//   - (false, nil) — mismatch.
//   - (false, err) — malformed hash, weak parameters, etc.
func Verify(password, encoded string) (bool, error) {
	if password == "" {
		return false, ErrEmptyPassword
	}
	policy, salt, hash, err := decodePHC(encoded)
	if err != nil {
		return false, err
	}
	if err := policy.Validate(); err != nil {
		// Policy below the floor — refuse the comparison entirely.
		// The caller logs this as a tampering attempt.
		return false, fmt.Errorf("verify: %w", err)
	}
	candidate := argon2.IDKey(
		[]byte(password),
		salt,
		policy.Iterations,
		policy.MemoryKiB,
		policy.Parallelism,
		uint32(len(hash)),
	)
	if subtle.ConstantTimeCompare(candidate, hash) == 1 {
		return true, nil
	}
	return false, nil
}

// NeedsRehash reports whether an existing hash should be migrated
// to a stronger parameter set. Login flows call this after Verify
// returns true; if it reports true, the cleartext password the user
// just submitted is fed back through Hash with the current policy
// and the new encoding written back.
func NeedsRehash(encoded string, policy Policy) (bool, error) {
	stored, _, _, err := decodePHC(encoded)
	if err != nil {
		return false, err
	}
	switch {
	case stored.MemoryKiB < policy.MemoryKiB:
		return true, nil
	case stored.Iterations < policy.Iterations:
		return true, nil
	case stored.Parallelism < policy.Parallelism:
		return true, nil
	case stored.KeyLen < policy.KeyLen:
		return true, nil
	}
	return false, nil
}

// Validate reports whether a Policy meets REQ-FUNC-PLT-IAM-001.
// Returns ErrPolicyTooWeak with a descriptive wrapped error.
func (p Policy) Validate() error {
	if p.MemoryKiB < PolicyV1.MemoryKiB {
		return fmt.Errorf("%w: memory %d KiB < required %d KiB",
			ErrPolicyTooWeak, p.MemoryKiB, PolicyV1.MemoryKiB)
	}
	if p.Iterations < PolicyV1.Iterations {
		return fmt.Errorf("%w: iterations %d < required %d",
			ErrPolicyTooWeak, p.Iterations, PolicyV1.Iterations)
	}
	if p.Parallelism < PolicyV1.Parallelism {
		return fmt.Errorf("%w: parallelism %d < required %d",
			ErrPolicyTooWeak, p.Parallelism, PolicyV1.Parallelism)
	}
	if p.KeyLen < 16 {
		return fmt.Errorf("%w: key_len %d < 16 bytes", ErrPolicyTooWeak, p.KeyLen)
	}
	if p.SaltLen < 8 {
		return fmt.Errorf("%w: salt_len %d < 8 bytes", ErrPolicyTooWeak, p.SaltLen)
	}
	return nil
}

// ----------------------------------------------------------------------
// PHC-string encoding
// ----------------------------------------------------------------------

// encodePHC builds the canonical Argon2id PHC string. base64 encoding
// per the spec is unpadded URL-safe-less variant (rawStdEncoding
// without padding); we use raw std encoding (no padding) which is
// what golang.org/x/crypto/argon2 documentation calls out.
func encodePHC(p Policy, salt, hash []byte) string {
	b64 := base64.RawStdEncoding
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2Version,
		p.MemoryKiB, p.Iterations, p.Parallelism,
		b64.EncodeToString(salt),
		b64.EncodeToString(hash),
	)
}

// decodePHC parses an Argon2id PHC string. The grammar is
// deliberately strict — anything that doesn't exactly match the
// expected shape is rejected as malformed.
func decodePHC(encoded string) (Policy, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	// Expected: ["", "argon2id", "v=19", "m=…,t=…,p=…", "<salt>", "<hash>"].
	if len(parts) != 6 || parts[0] != "" || parts[1] != AlgoArgon2id {
		return Policy{}, nil, nil, fmt.Errorf("password: %w: not an argon2id PHC string", ErrMalformedHash)
	}
	if !strings.HasPrefix(parts[2], "v=") {
		return Policy{}, nil, nil, fmt.Errorf("password: %w: missing version", ErrMalformedHash)
	}
	version, err := strconv.Atoi(strings.TrimPrefix(parts[2], "v="))
	if err != nil || version != argon2Version {
		return Policy{}, nil, nil, fmt.Errorf("password: %w: unsupported version %s", ErrMalformedHash, parts[2])
	}
	policy := Policy{}
	for _, kv := range strings.Split(parts[3], ",") {
		seg := strings.SplitN(kv, "=", 2)
		if len(seg) != 2 {
			return Policy{}, nil, nil, fmt.Errorf("password: %w: bad parameter %q", ErrMalformedHash, kv)
		}
		val, err := strconv.ParseUint(seg[1], 10, 32)
		if err != nil {
			return Policy{}, nil, nil, fmt.Errorf("password: %w: bad parameter %q", ErrMalformedHash, kv)
		}
		switch seg[0] {
		case "m":
			policy.MemoryKiB = uint32(val)
		case "t":
			policy.Iterations = uint32(val)
		case "p":
			policy.Parallelism = uint8(val)
		default:
			return Policy{}, nil, nil, fmt.Errorf("password: %w: unknown parameter %q", ErrMalformedHash, seg[0])
		}
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return Policy{}, nil, nil, fmt.Errorf("password: %w: bad salt encoding: %v", ErrMalformedHash, err)
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return Policy{}, nil, nil, fmt.Errorf("password: %w: bad hash encoding: %v", ErrMalformedHash, err)
	}
	policy.SaltLen = uint32(len(salt))
	policy.KeyLen = uint32(len(hash))
	return policy, salt, hash, nil
}

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrPolicyTooWeak is returned by Hash, Verify, and Policy.Validate
// when the supplied (or stored) parameters are below the minimums in
// REQ-FUNC-PLT-IAM-001.
var ErrPolicyTooWeak = errors.New("password: policy below REQ-FUNC-PLT-IAM-001 minimums")

// ErrEmptyPassword is returned when the caller supplies an empty
// password to Hash or Verify. Empty passwords MUST be rejected at
// the input layer; this is a defensive guard.
var ErrEmptyPassword = errors.New("password: empty password")

// ErrMalformedHash is returned by Verify and NeedsRehash when the
// stored hash string does not parse as a valid Argon2id PHC string.
var ErrMalformedHash = errors.New("password: malformed hash string")
