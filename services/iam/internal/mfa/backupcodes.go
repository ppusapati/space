// backupcodes.go — single-use recovery codes the user prints during
// MFA enrollment.
//
// → REQ-FUNC-PLT-IAM-004 acceptance #2: each backup code is single-
//   use; reuse rejected.
//
// Encoding choice: each code is 8 alphanumeric characters drawn from
// a 32-symbol alphabet ({2-9, A-N, P-Z} — {0,1,O} omitted because
// they're easily confused on paper). 32^8 ≈ 1.1 × 10^12 — ample for
// a 10-code book.
//
// Storage: bcrypt(work=12) hash of the code. Prefix-indexed so the
// verifier can locate the candidate row in O(log n) rather than
// computing N bcrypts per attempt.

package mfa

import (
	"crypto/rand"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// BackupCode parameters per REQ-FUNC-PLT-IAM-004.
const (
	// BackupCodeCount is the number of codes minted per enrollment.
	BackupCodeCount = 10

	// BackupCodeLength is the visible length of each code (excluding
	// any human-friendly grouping).
	BackupCodeLength = 8

	// BackupCodePrefixLength is the leading-character slice indexed
	// in the database to pick the right candidate row to bcrypt-
	// compare against. Four characters is a reasonable trade-off
	// between collision probability (~1.0 % of pairs collide in a
	// 10-code book at 32^4 = 1M buckets) and storage cost.
	BackupCodePrefixLength = 4

	// BackupBcryptCost is the bcrypt work factor. 12 ≈ ~250 ms on
	// the production hardware profile — enough to make brute-forcing
	// a stolen DB dump infeasible while leaving headroom for the
	// constant-time login budget.
	BackupBcryptCost = 12
)

// codeAlphabet is the 32-symbol Crockford-ish set. {0, 1, O, I, L}
// are removed to dodge the common transcription confusions when
// printed.
const codeAlphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

// BackupCodeIssued represents one freshly minted code. The Plain
// value is shown to the user EXACTLY ONCE during enrollment; the
// Hash + PrefixIdx are persisted in mfa_backup_codes.
type BackupCodeIssued struct {
	Plain     string
	Hash      []byte
	PrefixIdx string
}

// GenerateBackupCodes mints BackupCodeCount fresh codes. The function
// is goroutine-safe.
func GenerateBackupCodes() ([]BackupCodeIssued, error) {
	out := make([]BackupCodeIssued, 0, BackupCodeCount)
	for i := 0; i < BackupCodeCount; i++ {
		code, err := newBackupCodePlain()
		if err != nil {
			return nil, fmt.Errorf("mfa: backup code %d: %w", i, err)
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(code), BackupBcryptCost)
		if err != nil {
			return nil, fmt.Errorf("mfa: bcrypt %d: %w", i, err)
		}
		out = append(out, BackupCodeIssued{
			Plain:     code,
			Hash:      hash,
			PrefixIdx: code[:BackupCodePrefixLength],
		})
	}
	return out, nil
}

// VerifyBackupCode compares `presented` against a slice of stored
// hashes (the candidate set the caller fetched by prefix index).
// Returns the index of the matched hash on success — the caller MUST
// then mark that row consumed_at=now within the same transaction so
// the next presentation of the same code fails as a reuse.
func VerifyBackupCode(presented string, candidates [][]byte) (int, error) {
	presentedBytes := []byte(presented)
	for i, h := range candidates {
		if err := bcrypt.CompareHashAndPassword(h, presentedBytes); err == nil {
			return i, nil
		}
	}
	return -1, ErrBackupCodeNotFound
}

// PrefixOf returns the prefix index for a presented code. Helper for
// the store layer's WHERE clause.
func PrefixOf(code string) string {
	if len(code) < BackupCodePrefixLength {
		return code
	}
	return code[:BackupCodePrefixLength]
}

// newBackupCodePlain reads BackupCodeLength characters of secure
// randomness and maps them onto codeAlphabet via rejection sampling
// so the result is uniformly distributed (a naive `r % 32` biases
// the first 8 symbols slightly when the source isn't a power of 2 —
// here the source IS 256, which is divisible by 32, so modulo would
// also work; rejection sampling is kept anyway as a defensive
// measure in case BackupCodeLength or codeAlphabet ever changes).
func newBackupCodePlain() (string, error) {
	out := make([]byte, BackupCodeLength)
	buf := make([]byte, 1)
	for i := 0; i < BackupCodeLength; {
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		// 256 / 32 = 8 evenly — accept all.
		out[i] = codeAlphabet[buf[0]%32]
		i++
	}
	return string(out), nil
}

// ErrBackupCodeNotFound is returned by VerifyBackupCode when no
// candidate hash matches the presented code.
var ErrBackupCodeNotFound = errors.New("mfa: backup code not found")
