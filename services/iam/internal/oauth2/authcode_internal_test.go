package oauth2

import (
	"errors"
	"testing"
)

func TestEncodeDecodeAuthCodeBearer_Roundtrip(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	bearer := encodeAuthCodeBearer("rowid", secret)
	id, got, err := decodeAuthCodeBearer(bearer)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if id != "rowid" {
		t.Errorf("id: %q", id)
	}
	if string(got) != string(secret) {
		t.Errorf("secret roundtrip mismatch")
	}
}

func TestDecodeAuthCodeBearer_Malformed(t *testing.T) {
	cases := []string{
		"",
		"no-dot",
		".only-secret",
		"only-id.",
		"id.not!base64!",
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			if _, _, err := decodeAuthCodeBearer(c); !errors.Is(err, ErrAuthCodeNotFound) {
				t.Errorf("got %v want ErrAuthCodeNotFound", err)
			}
		})
	}
}

func TestHashAuthCode_Deterministic(t *testing.T) {
	a := hashAuthCode([]byte("x"))
	b := hashAuthCode([]byte("x"))
	if a != b {
		t.Errorf("hashAuthCode must be deterministic")
	}
	if hashAuthCode([]byte("x")) == hashAuthCode([]byte("y")) {
		t.Errorf("different inputs must hash differently")
	}
}

func TestDecodeBasic(t *testing.T) {
	cases := []struct {
		header     string
		wantID     string
		wantSecret string
		wantOK     bool
	}{
		{"Basic Y2xpZW50OnNlY3JldA==", "client", "secret", true}, // "client:secret"
		{"Basic " + "", "", "", false},
		{"Bearer x", "", "", false},
		{"Basic !!!", "", "", false},
	}
	for _, c := range cases {
		t.Run(c.header, func(t *testing.T) {
			id, secret, ok := decodeBasic(c.header)
			if ok != c.wantOK || id != c.wantID || secret != c.wantSecret {
				t.Errorf("got (%q,%q,%v) want (%q,%q,%v)",
					id, secret, ok, c.wantID, c.wantSecret, c.wantOK)
			}
		})
	}
}
