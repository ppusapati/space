package mfa

import (
	"net/url"
	"strings"
	"testing"
)

func TestEnrollmentURI_Shape(t *testing.T) {
	secret := []byte("12345678901234567890")
	uri, err := EnrollmentURI("Chetana", "user@example.com", secret)
	if err != nil {
		t.Fatalf("uri: %v", err)
	}
	if !strings.HasPrefix(uri, "otpauth://totp/") {
		t.Errorf("scheme: %q", uri)
	}

	parsed, err := url.Parse(uri)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parsed.Scheme != "otpauth" || parsed.Host != "totp" {
		t.Errorf("scheme/host: %q %q", parsed.Scheme, parsed.Host)
	}

	// Path = /Chetana:user@example.com (with @ percent-encoded).
	wantPathPrefix := "/Chetana:user"
	if !strings.HasPrefix(parsed.Path, wantPathPrefix) {
		t.Errorf("path: got %q want prefix %q", parsed.Path, wantPathPrefix)
	}

	q := parsed.Query()
	if got := q.Get("issuer"); got != "Chetana" {
		t.Errorf("issuer: %q", got)
	}
	if got := q.Get("algorithm"); got != "SHA1" {
		t.Errorf("algorithm: %q", got)
	}
	if got := q.Get("digits"); got != "6" {
		t.Errorf("digits: %q", got)
	}
	if got := q.Get("period"); got != "30" {
		t.Errorf("period: %q", got)
	}
	if got := q.Get("secret"); got == "" {
		t.Error("secret missing")
	} else {
		// Decode and confirm it round-trips back to the input.
		decoded, err := DecodeSecret(got)
		if err != nil {
			t.Fatalf("decode secret: %v", err)
		}
		if string(decoded) != string(secret) {
			t.Error("secret roundtrip mismatch")
		}
	}
}

func TestEnrollmentURI_Validation(t *testing.T) {
	secret := []byte{1, 2, 3, 4}
	cases := []struct {
		name             string
		issuer, account  string
		secret           []byte
	}{
		{"empty issuer", "", "u@x", secret},
		{"empty account", "Chetana", "", secret},
		{"empty secret", "Chetana", "u@x", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := EnrollmentURI(tc.issuer, tc.account, tc.secret); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
