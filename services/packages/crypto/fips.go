// Package crypto contains cryptographic policy helpers shared across
// services. The most important contract is the FIPS self-check — every
// Go binary calls AssertFIPS at boot and refuses to serve if the active
// crypto provider does not satisfy the configured posture.
//
// → REQ-NFR-SEC-001 (FIPS 140-2/3 cryptographic modules)
// → design.md §4.1.3 (FIPS crypto), §4.7
//
// The actual provider check is implemented in two files:
//
//   • fips_boring.go   (//go:build boringcrypto)  — calls boring.Enabled()
//                      to confirm the BoringCrypto module is in use.
//   • fips_default.go  (//go:build !boringcrypto) — reports "stdlib", which
//                      fails AssertFIPS when CHETANA_REQUIRE_FIPS=1.
//
// Production builds compile with GOEXPERIMENT=boringcrypto and therefore
// pick up fips_boring.go. Local dev / CI builds without the experiment
// pick up fips_default.go and may run only when CHETANA_REQUIRE_FIPS is
// unset (or "0" / "false").
package crypto

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// requireFIPSEnv is the environment variable that toggles FIPS enforcement.
// When set to a truthy value ("1", "true", "yes"), AssertFIPS exits the
// process if FIPSStatus().Enabled is false.
const requireFIPSEnv = "CHETANA_REQUIRE_FIPS"

// Status describes the current crypto provider posture. Returned from
// FIPSStatus() and consumed by AssertFIPS for the boot-time check and
// by /metrics for the build_info gauge.
type Status struct {
	// Provider names the active crypto provider ("boringcrypto" or
	// "stdlib"). Defined in fips_boring.go and fips_default.go via
	// build tags.
	Provider string

	// Enabled is true when the provider satisfies the FIPS contract.
	// Only "boringcrypto" sets this true today.
	Enabled bool
}

// FIPSStatus returns the current crypto provider posture. The
// implementation is supplied by either fips_boring.go (build tag
// boringcrypto) or fips_default.go (build tag !boringcrypto).
func FIPSStatus() Status { return fipsStatus() }

// AssertFIPS performs the boot-time FIPS self-check.
//
// Behavior:
//   - Always calls FIPSStatus() and emits a structured log line:
//     "fips: provider=<provider> status=<ok|disabled>".
//   - When CHETANA_REQUIRE_FIPS is truthy and Enabled is false, returns
//     a non-nil error so the caller can exit non-zero before serving any
//     RPCs (acceptance criterion #4).
//   - When CHETANA_REQUIRE_FIPS is unset or falsy, returns nil regardless
//     of provider (local dev / CI without the experiment).
//
// Pass logger=nil to use slog.Default().
func AssertFIPS(logger *slog.Logger) error {
	if logger == nil {
		logger = slog.Default()
	}
	st := FIPSStatus()
	statusLabel := "ok"
	if !st.Enabled {
		statusLabel = "disabled"
	}
	logger.Info("fips self-check",
		slog.String("provider", st.Provider),
		slog.String("status", statusLabel),
	)
	if requireFIPS() && !st.Enabled {
		return fmt.Errorf(
			"fips: %s required (CHETANA_REQUIRE_FIPS=1) but provider %q does not satisfy the contract; rebuild with GOEXPERIMENT=boringcrypto",
			"boringcrypto", st.Provider,
		)
	}
	return nil
}

// MustAssertFIPS is the convenience wrapper that calls AssertFIPS and
// log.Fatals on error. Service entrypoints can call this directly:
//
//	func main() {
//	    crypto.MustAssertFIPS(slog.Default())
//	    // … rest of bootstrap …
//	}
func MustAssertFIPS(logger *slog.Logger) {
	if err := AssertFIPS(logger); err != nil {
		if logger == nil {
			logger = slog.Default()
		}
		logger.Error("fips self-check failed", slog.Any("err", err))
		os.Exit(1)
	}
}

// requireFIPS interprets CHETANA_REQUIRE_FIPS as a boolean.
func requireFIPS() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(requireFIPSEnv))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// ErrFIPSRequired is the sentinel error returned by AssertFIPS when the
// FIPS posture is required but not satisfied. Exposed for callers that
// want to recognise the failure mode programmatically.
var ErrFIPSRequired = errors.New("fips: required provider not enabled")
