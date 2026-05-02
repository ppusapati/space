// enroll.go — otpauth:// URI generator for QR provisioning.
//
// The URI format is the de-facto standard published by Google:
// https://github.com/google/google-authenticator/wiki/Key-Uri-Format
// Every authenticator app (Google Authenticator, Authy, 1Password,
// Bitwarden, FreeOTP) consumes this exact shape via a QR scan.
//
// → REQ-FUNC-PLT-IAM-004 acceptance #1: enroll → scan QR → submit
//   code completes within one HTTP round-trip.

package mfa

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// EnrollmentURI builds the otpauth://totp/... URI for a given user.
//
// `issuer` is the human-readable service name shown in the
// authenticator (e.g. "Chetana"); `account` is typically the user's
// email. The URI carries the SAME issuer in both the label prefix
// and the issuer query parameter — that redundancy is required by
// some authenticator apps that read only one of the two.
func EnrollmentURI(issuer, account string, secret []byte) (string, error) {
	if strings.TrimSpace(issuer) == "" {
		return "", errors.New("mfa: empty issuer")
	}
	if strings.TrimSpace(account) == "" {
		return "", errors.New("mfa: empty account")
	}
	if len(secret) == 0 {
		return "", errors.New("mfa: empty secret")
	}

	// Label = "issuer:account", URL-encoded as a single path segment.
	label := url.PathEscape(fmt.Sprintf("%s:%s", issuer, account))

	q := url.Values{}
	q.Set("secret", EncodeSecret(secret))
	q.Set("issuer", issuer)
	q.Set("algorithm", "SHA1")
	q.Set("digits", fmt.Sprintf("%d", Digits))
	q.Set("period", fmt.Sprintf("%d", StepSeconds))

	return fmt.Sprintf("otpauth://totp/%s?%s", label, q.Encode()), nil
}
