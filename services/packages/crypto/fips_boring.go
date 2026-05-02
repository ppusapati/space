//go:build boringcrypto

package crypto

import "crypto/boring"

// fipsStatus reports the BoringCrypto provider when GOEXPERIMENT=boringcrypto
// is set at build time. boring.Enabled() is the source of truth — when
// running on a runtime that did not link the BoringCrypto module it
// returns false even though this build tag is present.
func fipsStatus() Status {
	return Status{
		Provider: "boringcrypto",
		Enabled:  boring.Enabled(),
	}
}
