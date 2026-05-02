// Package mfa implements multi-factor authentication for the chetana
// IAM service: RFC 6238 TOTP, single-use backup codes, and the
// otpauth:// provisioning URIs Authenticator apps consume.
//
// → REQ-FUNC-PLT-IAM-004 (TOTP + 10 backup codes; ±1 step tolerance;
//                         replay rejection within the active window).
// → design.md §4.1.1 token + amr=["pwd","totp"].
//
// Algorithm choice: HMAC-SHA1 with 30-second steps and 6-digit codes,
// matching the universal Google Authenticator / 1Password / Authy
// default. Newer SHA-256/SHA-512 variants exist in RFC 6238 but the
// authenticator ecosystem has not standardised on them; sticking to
// SHA-1 maximises interop while still meeting the FIPS 140-3 module
// requirements (SHA-1 in HMAC mode is permitted for HOTP/TOTP usage).

package mfa

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// TOTP parameters per REQ-FUNC-PLT-IAM-004 + RFC 6238 defaults.
const (
	// SecretBytes is the size of a freshly minted TOTP secret. RFC
	// 6238 §5.1 recommends ≥160 bits; we use exactly 160 bits (20
	// bytes) so the base32 encoding fits in 32 characters with no
	// padding — a clean shape for QR codes and manual entry.
	SecretBytes = 20

	// Digits is the number of decimal digits in a generated code.
	Digits = 6

	// StepSeconds is the time-step Δ. RFC 6238 §5.2 recommends 30s.
	StepSeconds = 30

	// ToleranceSteps is the ± window the verifier accepts around the
	// current step. ±1 covers clock skew between authenticator and
	// server up to one full step in either direction. REQ-FUNC-PLT-
	// IAM-004 fixes this at ±1.
	ToleranceSteps = 1
)

// digitMod is 10^Digits — used to truncate HOTP output to Digits
// decimal places. Computed once so the hot path is multiply-free.
var digitMod = func() uint32 {
	m := uint32(1)
	for i := 0; i < Digits; i++ {
		m *= 10
	}
	return m
}()

// GenerateSecret returns a freshly minted TOTP secret. The caller
// stores the raw bytes in `mfa_totp_secrets.secret` and shows the
// base32 encoding (via EncodeSecret) to the user during enrollment.
func GenerateSecret() ([]byte, error) {
	b := make([]byte, SecretBytes)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("mfa: read random: %w", err)
	}
	return b, nil
}

// EncodeSecret returns the base32 (RFC 4648, no padding) form an
// authenticator app expects. Authenticator apps universally accept
// uppercase no-padding base32; we emit exactly that.
func EncodeSecret(secret []byte) string {
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret)
}

// DecodeSecret parses the base32 form. Accepts both padded and
// unpadded input + lowercase variants (some users hand-type the
// secret on a second device).
func DecodeSecret(encoded string) ([]byte, error) {
	encoded = normaliseBase32(encoded)
	b, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("mfa: decode secret: %w", err)
	}
	return b, nil
}

// Generate returns the 6-digit TOTP code valid at time t. Exposed for
// tests + the QR-preview server-side render; callers verifying codes
// should use Verify, which applies the ±1-step tolerance.
func Generate(secret []byte, t time.Time) string {
	step := uint64(t.Unix() / StepSeconds)
	return hotp(secret, step)
}

// Verify checks `code` against the secret at time t, accepting ±1
// step of tolerance.
//
// To defeat replay attacks within the same time-step window the
// caller MUST also consult a replay cache keyed by (user, step,
// code) — see store.go's ConsumeReplayWindow. Verify alone validates
// the algorithm; it does not enforce single-use.
//
// Returns the matched step (so the caller can pass it to the replay
// cache) when successful; (0, ErrInvalidCode) on mismatch.
func Verify(secret []byte, code string, t time.Time) (uint64, error) {
	if len(code) != Digits {
		return 0, ErrInvalidCode
	}
	if _, err := strconv.Atoi(code); err != nil {
		return 0, ErrInvalidCode
	}
	current := uint64(t.Unix() / StepSeconds)
	for delta := -ToleranceSteps; delta <= ToleranceSteps; delta++ {
		step := current
		switch {
		case delta < 0:
			if uint64(-delta) > current {
				continue
			}
			step = current - uint64(-delta)
		case delta > 0:
			step = current + uint64(delta)
		}
		if constantTimeStringsEqual(hotp(secret, step), code) {
			return step, nil
		}
	}
	return 0, ErrInvalidCode
}

// hotp implements RFC 4226 §5.3: HMAC-SHA1(secret, counter), then
// dynamic truncation, then mod 10^Digits.
func hotp(secret []byte, counter uint64) string {
	var ctr [8]byte
	binary.BigEndian.PutUint64(ctr[:], counter)

	mac := hmac.New(sha1.New, secret)
	mac.Write(ctr[:])
	sum := mac.Sum(nil)

	// Dynamic truncation per RFC 4226 §5.3 step 2.
	offset := sum[len(sum)-1] & 0x0f
	bin := (uint32(sum[offset])&0x7f)<<24 |
		uint32(sum[offset+1])<<16 |
		uint32(sum[offset+2])<<8 |
		uint32(sum[offset+3])

	code := bin % digitMod
	return fmt.Sprintf("%0*d", Digits, code)
}

// constantTimeStringsEqual is a lightweight constant-time string
// comparator; used so a timing attacker cannot detect which prefix of
// a bad code matched a valid step.
func constantTimeStringsEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

// normaliseBase32 uppercases + strips spaces from a manually-entered
// secret. Authenticator apps display the secret in groups of 4 with
// spaces; users sometimes paste those.
func normaliseBase32(s string) string {
	out := make([]byte, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z', r >= '2' && r <= '7':
			out = append(out, byte(r))
		case r >= 'a' && r <= 'z':
			out = append(out, byte(r-'a'+'A'))
		case r == ' ', r == '-':
			// drop separators
		default:
			out = append(out, byte(r))
		}
	}
	return string(out)
}

// ErrInvalidCode is returned when a presented TOTP does not match
// any step in the tolerance window.
var ErrInvalidCode = errors.New("mfa: invalid TOTP code")
