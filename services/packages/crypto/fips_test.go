package crypto

import (
	"errors"
	"strings"
	"testing"
)

// TestFIPSStatus_AlwaysReportsAProvider verifies that whichever build tag
// is active, FIPSStatus returns a non-empty Provider. Regression guard for
// build-tag drift.
func TestFIPSStatus_AlwaysReportsAProvider(t *testing.T) {
	st := FIPSStatus()
	if st.Provider == "" {
		t.Fatal("FIPSStatus returned empty provider; build-tag wiring is broken")
	}
	switch st.Provider {
	case "boringcrypto", "stdlib":
		// expected
	default:
		t.Fatalf("FIPSStatus returned unexpected provider %q", st.Provider)
	}
}

// TestAssertFIPS_NoEnforcement_ReturnsNil verifies that without
// CHETANA_REQUIRE_FIPS set, AssertFIPS never returns an error regardless
// of provider. This keeps local dev / non-FIPS CI builds runnable.
func TestAssertFIPS_NoEnforcement_ReturnsNil(t *testing.T) {
	t.Setenv(requireFIPSEnv, "")
	if err := AssertFIPS(nil); err != nil {
		t.Fatalf("AssertFIPS without enforcement should return nil; got %v", err)
	}
}

// TestAssertFIPS_EnforcementWithoutBoring_ReturnsError verifies the
// failure mode demanded by acceptance criterion #4 — the process must
// not proceed past AssertFIPS when FIPS is required but not provided.
//
// This test only runs when the active provider is "stdlib" (the default
// non-boringcrypto build). Under //go:build boringcrypto it would be a
// false test (the provider IS enabled) so we skip.
func TestAssertFIPS_EnforcementWithoutBoring_ReturnsError(t *testing.T) {
	if FIPSStatus().Enabled {
		t.Skip("running under boringcrypto build; nothing to enforce")
	}
	t.Setenv(requireFIPSEnv, "1")
	err := AssertFIPS(nil)
	if err == nil {
		t.Fatal("AssertFIPS with CHETANA_REQUIRE_FIPS=1 and stdlib provider must error")
	}
	if !strings.Contains(err.Error(), "boringcrypto") {
		t.Errorf("error message should mention boringcrypto remediation; got %q", err)
	}
}

// TestRequireFIPS_TruthyValues verifies the env var parser recognises the
// documented truthy spellings.
func TestRequireFIPS_TruthyValues(t *testing.T) {
	cases := map[string]bool{
		"":      false,
		"0":     false,
		"false": false,
		"no":    false,
		"off":   false,
		"junk":  false,
		"1":     true,
		"true":  true,
		"yes":   true,
		"on":    true,
		"TRUE":  true, // case insensitive
	}
	for in, want := range cases {
		t.Run(in, func(t *testing.T) {
			t.Setenv(requireFIPSEnv, in)
			if got := requireFIPS(); got != want {
				t.Errorf("requireFIPS(%q) = %v, want %v", in, got, want)
			}
		})
	}
}

// TestErrFIPSRequiredIsExported verifies the sentinel is reachable via
// errors.Is for callers that want programmatic detection. (We don't yet
// wrap with this sentinel — verify exposure so future enrichment is
// straightforward.)
func TestErrFIPSRequiredIsExported(t *testing.T) {
	if errors.Is(nil, ErrFIPSRequired) {
		t.Fatal("nil should not match ErrFIPSRequired")
	}
	if !errors.Is(ErrFIPSRequired, ErrFIPSRequired) {
		t.Fatal("sentinel should match itself via errors.Is")
	}
}
