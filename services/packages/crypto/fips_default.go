//go:build !boringcrypto

package crypto

// fipsStatus reports the standard library provider for builds without
// GOEXPERIMENT=boringcrypto. Enabled is always false — these builds do
// not satisfy the REQ-NFR-SEC-001 FIPS contract and should not run in
// production.
func fipsStatus() Status {
	return Status{
		Provider: "stdlib",
		Enabled:  false,
	}
}
