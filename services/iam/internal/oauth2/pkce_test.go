package oauth2

import (
	"errors"
	"strings"
	"testing"
)

// RFC 7636 Appendix B test vector.
const (
	rfcVerifier  = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	rfcChallenge = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
)

func TestComputeS256Challenge_RFCVector(t *testing.T) {
	got := ComputeS256Challenge(rfcVerifier)
	if got != rfcChallenge {
		t.Errorf("RFC 7636 Appendix B: got %q want %q", got, rfcChallenge)
	}
}

func TestVerifyVerifier_HappyPath(t *testing.T) {
	if err := VerifyVerifier(rfcVerifier, rfcChallenge); err != nil {
		t.Errorf("verify: %v", err)
	}
}

func TestVerifyVerifier_Mismatch(t *testing.T) {
	bad := strings.Repeat("a", 43)
	if err := VerifyVerifier(bad, rfcChallenge); !errors.Is(err, ErrPKCEMismatch) {
		t.Errorf("got %v want ErrPKCEMismatch", err)
	}
}

func TestVerifyVerifier_InvalidShape(t *testing.T) {
	cases := []struct {
		name, verifier string
	}{
		{"too short", strings.Repeat("a", 42)},
		{"too long", strings.Repeat("a", 129)},
		{"illegal char", strings.Repeat("a", 42) + "!"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := VerifyVerifier(tc.verifier, rfcChallenge); !errors.Is(err, ErrInvalidVerifier) {
				t.Errorf("got %v want ErrInvalidVerifier", err)
			}
		})
	}
}

func TestValidateChallengeShape(t *testing.T) {
	if err := ValidateChallengeShape(rfcChallenge); err != nil {
		t.Errorf("RFC challenge rejected: %v", err)
	}
	cases := []string{
		"",
		strings.Repeat("a", 42),
		strings.Repeat("a", 44),
		strings.Repeat("a", 42) + "=", // padded
		strings.Repeat("a", 42) + "/", // illegal
	}
	for _, c := range cases {
		if err := ValidateChallengeShape(c); !errors.Is(err, ErrInvalidChallenge) {
			t.Errorf("ValidateChallengeShape(%q): got %v want ErrInvalidChallenge", c, err)
		}
	}
}

// REQ-FUNC-PLT-IAM-006 acceptance #1 — plain method explicitly
// rejected.
func TestValidateMethod(t *testing.T) {
	cases := []struct {
		method string
		want   error
	}{
		{"S256", nil},
		{"", ErrMissingChallengeMethod},
		{"plain", ErrPlainMethodForbidden},
		{"S512", ErrUnsupportedChallengeMethod},
		{"sha256", ErrUnsupportedChallengeMethod},
	}
	for _, c := range cases {
		got := ValidateMethod(c.method)
		if c.want == nil && got != nil {
			t.Errorf("ValidateMethod(%q): unexpected error %v", c.method, got)
		}
		if c.want != nil && !errors.Is(got, c.want) {
			t.Errorf("ValidateMethod(%q): got %v want %v", c.method, got, c.want)
		}
	}
}
