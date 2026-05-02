package mfa

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestGenerateBackupCodes_Shape(t *testing.T) {
	codes, err := GenerateBackupCodes()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(codes) != BackupCodeCount {
		t.Errorf("count: got %d want %d", len(codes), BackupCodeCount)
	}
	seen := make(map[string]bool)
	for i, c := range codes {
		if len(c.Plain) != BackupCodeLength {
			t.Errorf("code %d length: %d", i, len(c.Plain))
		}
		for _, r := range c.Plain {
			if !strings.ContainsRune(codeAlphabet, r) {
				t.Errorf("code %d %q has illegal char %q", i, c.Plain, r)
			}
		}
		if c.PrefixIdx != c.Plain[:BackupCodePrefixLength] {
			t.Errorf("prefix %d: got %q want %q", i, c.PrefixIdx, c.Plain[:BackupCodePrefixLength])
		}
		if seen[c.Plain] {
			t.Errorf("code %d %q duplicates an earlier code", i, c.Plain)
		}
		seen[c.Plain] = true
		// Hash must verify the plain.
		if err := bcrypt.CompareHashAndPassword(c.Hash, []byte(c.Plain)); err != nil {
			t.Errorf("code %d hash does not verify plain: %v", i, err)
		}
	}
}

func TestVerifyBackupCode_FindsMatch(t *testing.T) {
	codes, err := GenerateBackupCodes()
	if err != nil {
		t.Fatalf("gen: %v", err)
	}
	candidates := make([][]byte, len(codes))
	for i, c := range codes {
		candidates[i] = c.Hash
	}

	// Pick the 5th code.
	target := codes[4]
	idx, err := VerifyBackupCode(target.Plain, candidates)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if idx != 4 {
		t.Errorf("idx: got %d want 4", idx)
	}
}

func TestVerifyBackupCode_RejectsUnknown(t *testing.T) {
	codes, _ := GenerateBackupCodes()
	candidates := make([][]byte, len(codes))
	for i, c := range codes {
		candidates[i] = c.Hash
	}
	if _, err := VerifyBackupCode("ZZZZZZZZ", candidates); !errors.Is(err, ErrBackupCodeNotFound) {
		t.Errorf("got %v want ErrBackupCodeNotFound", err)
	}
}

func TestPrefixOf(t *testing.T) {
	if got := PrefixOf("ABCD1234"); got != "ABCD" {
		t.Errorf("PrefixOf: %q", got)
	}
	if got := PrefixOf("AB"); got != "AB" {
		t.Errorf("PrefixOf short: %q", got)
	}
}
