package mfa

import (
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestGenerateSecret_Length(t *testing.T) {
	s, err := GenerateSecret()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(s) != SecretBytes {
		t.Errorf("len: got %d want %d", len(s), SecretBytes)
	}
}

func TestEncodeDecodeSecret_Roundtrip(t *testing.T) {
	in := []byte{0x00, 0x01, 0x02, 0x03, 0xff, 0xfe, 0xfd, 0xfc,
		0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80,
		0x90, 0xa0, 0xb0, 0xc0}
	enc := EncodeSecret(in)
	if strings.Contains(enc, "=") {
		t.Errorf("base32 must be unpadded: %q", enc)
	}
	got, err := DecodeSecret(enc)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if string(got) != string(in) {
		t.Errorf("roundtrip mismatch")
	}

	// Lowercase + spaces tolerated.
	low := strings.ToLower(enc[:8]) + " " + enc[8:]
	got2, err := DecodeSecret(low)
	if err != nil {
		t.Fatalf("decode normalised: %v", err)
	}
	if string(got2) != string(in) {
		t.Error("normalised roundtrip mismatch")
	}
}

// RFC 6238 Appendix B test vectors (SHA-1, 8-digit) cannot be used
// directly because we emit 6-digit codes. Instead we cross-check
// against an independent computation: HMAC-SHA1 of the time step
// truncated per RFC 4226 §5.3. Vector below was computed with
// Go's hmac package against the canonical RFC 6238 secret
// "12345678901234567890" at T0 = 59 seconds → step 1.
func TestGenerate_KnownVector(t *testing.T) {
	secret := []byte("12345678901234567890")
	t0 := time.Unix(59, 0)
	got := Generate(secret, t0)
	want := "287082" // RFC 6238 Appendix B truncated to 6 digits
	if got != want {
		t.Errorf("RFC 6238 vector: got %q want %q", got, want)
	}
}

func TestGenerate_StepBoundary(t *testing.T) {
	secret, _ := GenerateSecret()
	// Anchor on a step-aligned epoch so the assertions stay sharp.
	base := int64(1_700_000_010) // exactly divisible by 30
	t1 := time.Unix(base, 0)
	t2 := time.Unix(base+StepSeconds-1, 0)
	if Generate(secret, t1) != Generate(secret, t2) {
		t.Error("codes within the same step must match")
	}
	// The next second crosses into the next step.
	t3 := time.Unix(base+StepSeconds, 0)
	if Generate(secret, t1) == Generate(secret, t3) {
		t.Error("codes across step boundary must differ")
	}
}

func TestVerify_HappyPath(t *testing.T) {
	secret, _ := GenerateSecret()
	now := time.Unix(1_700_000_030, 0)
	code := Generate(secret, now)
	step, err := Verify(secret, code, now)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if step != uint64(now.Unix()/StepSeconds) {
		t.Errorf("step: got %d want %d", step, now.Unix()/StepSeconds)
	}
}

// REQ-FUNC-PLT-IAM-004: ±1 step tolerance.
func TestVerify_ToleranceWindow(t *testing.T) {
	secret, _ := GenerateSecret()
	now := time.Unix(1_700_000_030, 0)
	prev := now.Add(-StepSeconds * time.Second)
	next := now.Add(StepSeconds * time.Second)

	for _, target := range []time.Time{prev, now, next} {
		code := Generate(secret, target)
		if _, err := Verify(secret, code, now); err != nil {
			t.Errorf("code from step %v rejected at %v: %v", target, now, err)
		}
	}

	// Outside the tolerance window: ±2 steps must be rejected.
	tooEarly := now.Add(-2 * StepSeconds * time.Second)
	tooLate := now.Add(2 * StepSeconds * time.Second)
	for _, target := range []time.Time{tooEarly, tooLate} {
		code := Generate(secret, target)
		if _, err := Verify(secret, code, now); !errors.Is(err, ErrInvalidCode) {
			t.Errorf("code from step %v should be rejected at %v: got %v", target, now, err)
		}
	}
}

func TestVerify_RejectsBadShapes(t *testing.T) {
	secret, _ := GenerateSecret()
	now := time.Now()
	cases := []string{"", "12345", "1234567", "abcdef", "12 345"}
	for _, c := range cases {
		if _, err := Verify(secret, c, now); !errors.Is(err, ErrInvalidCode) {
			t.Errorf("verify(%q): got %v want ErrInvalidCode", c, err)
		}
	}
}

func TestVerify_DifferentSecretFails(t *testing.T) {
	a, _ := GenerateSecret()
	b, _ := GenerateSecret()
	now := time.Now()
	code := Generate(a, now)
	if _, err := Verify(b, code, now); !errors.Is(err, ErrInvalidCode) {
		t.Errorf("got %v want ErrInvalidCode", err)
	}
}

func TestHotp_Determinism(t *testing.T) {
	secret, _ := hex.DecodeString("3132333435363738393031323334353637383930")
	if hotp(secret, 0) != hotp(secret, 0) {
		t.Error("hotp must be deterministic")
	}
}

func TestNormaliseBase32(t *testing.T) {
	cases := map[string]string{
		"abc 234":    "ABC234",
		"ABC-DEF":    "ABCDEF",
		"a b c":      "ABC",
		"NOPADDING ": "NOPADDING",
	}
	for in, want := range cases {
		if got := normaliseBase32(in); got != want {
			t.Errorf("normalise(%q): got %q want %q", in, got, want)
		}
	}
}
