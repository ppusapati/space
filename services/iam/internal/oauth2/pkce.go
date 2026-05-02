// Package oauth2 implements the chetana IAM service's OAuth 2.1 +
// OpenID Connect authorisation server.
//
// → REQ-FUNC-PLT-IAM-006 (OIDC issuer + OAuth 2.1: auth-code with
//                          PKCE, refresh-token, client-credentials).
// → design.md §4.1.1 (token model + ABAC decision).
//
// Strict choices for v1 (NOT runtime-configurable per OAuth 2.1 BCP):
//
//   • PKCE method MUST be S256. The legacy `plain` method is
//     explicitly rejected — IETF OAuth 2.1 §4.1.1.6 forbids it.
//   • Implicit + ROPC grants are NOT implemented. Authorisation-code
//     (with PKCE) and client-credentials are the only flows.
//   • Redirect URI matching is EXACT byte-for-byte (no scheme-host-
//     only matching, no glob, no port-range — see OAuth 2.1
//     §1.4.2).
//   • Auth codes are single-use, 10-minute TTL, and bound to
//     (client_id, user_id, code_challenge, redirect_uri, scopes) at
//     issue time.
//   • Client secrets are stored argon2id-hashed (REQ-FUNC-PLT-IAM-001
//     parity); public clients (e.g. SPA) carry an empty secret hash
//     and rely on PKCE alone for proof-of-possession.

package oauth2

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
)

// PKCE parameters per RFC 7636 + OAuth 2.1.
const (
	// MinVerifierLength is the lower bound for a code_verifier per
	// RFC 7636 §4.1 (43 base64url characters of entropy).
	MinVerifierLength = 43

	// MaxVerifierLength is the upper bound (128 chars per RFC 7636).
	MaxVerifierLength = 128

	// MethodS256 is the only PKCE challenge method we accept.
	MethodS256 = "S256"

	// MethodPlain is the deprecated plain method. We reject it
	// explicitly to surface a clear error to misconfigured clients
	// rather than silently accepting it.
	MethodPlain = "plain"
)

// ComputeS256Challenge returns base64url-unpadded(SHA256(verifier))
// per RFC 7636 §4.2. Exposed for tests + the chetana CLI's PKCE
// helper.
func ComputeS256Challenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// ValidateChallengeShape rejects challenges that don't fit the
// RFC 7636 base64url-unpadded SHA-256 output shape (43 characters
// from the URL-safe alphabet, no padding). Catches typos and the
// common copy-paste-with-padding mistake at the /authorize edge
// before we issue a code.
func ValidateChallengeShape(challenge string) error {
	if len(challenge) != 43 {
		return ErrInvalidChallenge
	}
	for i := 0; i < len(challenge); i++ {
		c := challenge[i]
		switch {
		case c >= 'A' && c <= 'Z',
			c >= 'a' && c <= 'z',
			c >= '0' && c <= '9',
			c == '-', c == '_':
			// ok
		default:
			return ErrInvalidChallenge
		}
	}
	return nil
}

// VerifyVerifier checks the presented verifier against the stored
// challenge. The caller has already pinned the challenge method to
// MethodS256 at /authorize time; this function does the actual
// SHA-256 + constant-time compare.
//
// Returns nil on match; ErrInvalidVerifier on length / charset
// violations; ErrPKCEMismatch on a clean shape that simply doesn't
// hash to the stored challenge.
func VerifyVerifier(verifier, storedChallenge string) error {
	if len(verifier) < MinVerifierLength || len(verifier) > MaxVerifierLength {
		return ErrInvalidVerifier
	}
	for i := 0; i < len(verifier); i++ {
		c := verifier[i]
		switch {
		case c >= 'A' && c <= 'Z',
			c >= 'a' && c <= 'z',
			c >= '0' && c <= '9',
			c == '-', c == '_', c == '.', c == '~':
			// ok per RFC 7636 §4.1 unreserved alphabet
		default:
			return ErrInvalidVerifier
		}
	}
	got := ComputeS256Challenge(verifier)
	if subtle.ConstantTimeCompare([]byte(got), []byte(storedChallenge)) != 1 {
		return ErrPKCEMismatch
	}
	return nil
}

// ValidateMethod rejects everything other than S256. The caller
// passes the `code_challenge_method` query parameter exactly as
// the client sent it — empty + "plain" + anything else are all
// errors. RFC 7636 says missing-method MAY default to plain; OAuth
// 2.1 §4.1.1.6 forbids plain so we treat missing as an error too.
func ValidateMethod(method string) error {
	if method == "" {
		return ErrMissingChallengeMethod
	}
	if method == MethodPlain {
		return ErrPlainMethodForbidden
	}
	if method != MethodS256 {
		return ErrUnsupportedChallengeMethod
	}
	return nil
}

// ----------------------------------------------------------------------
// PKCE errors
// ----------------------------------------------------------------------

// ErrInvalidChallenge is returned when the code_challenge value is
// malformed (wrong length, illegal alphabet).
var ErrInvalidChallenge = errors.New("oauth2: invalid code_challenge shape")

// ErrInvalidVerifier is returned when the code_verifier is
// malformed (length out of bounds, illegal alphabet).
var ErrInvalidVerifier = errors.New("oauth2: invalid code_verifier shape")

// ErrPKCEMismatch is returned when SHA256(verifier) does not match
// the stored challenge.
var ErrPKCEMismatch = errors.New("oauth2: PKCE verifier did not match stored challenge")

// ErrMissingChallengeMethod is returned when /authorize was called
// without a code_challenge_method parameter.
var ErrMissingChallengeMethod = errors.New("oauth2: code_challenge_method is required")

// ErrPlainMethodForbidden is returned when /authorize requested the
// `plain` PKCE method (OAuth 2.1 §4.1.1.6 forbids it).
var ErrPlainMethodForbidden = errors.New("oauth2: PKCE method 'plain' is forbidden; use S256")

// ErrUnsupportedChallengeMethod is returned when the requested
// method is neither plain nor S256.
var ErrUnsupportedChallengeMethod = errors.New("oauth2: unsupported code_challenge_method")
