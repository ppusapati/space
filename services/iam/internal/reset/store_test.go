package reset

import (
	"errors"
	"strings"
	"testing"
)

func TestEncodeDecodeBearer_Roundtrip(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	bearer := encodeBearer("rowid", secret)
	id, got, err := decodeBearer(bearer)
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

func TestDecodeBearer_Malformed(t *testing.T) {
	cases := []string{
		"",
		"no-dot",
		".only-secret",
		"only-id.",
		"id.not!base64!",
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			if _, _, err := decodeBearer(c); !errors.Is(err, ErrTokenNotFound) {
				t.Errorf("got %v want ErrTokenNotFound", err)
			}
		})
	}
}

func TestHashToken_Deterministic(t *testing.T) {
	a := hashToken([]byte("x"))
	b := hashToken([]byte("x"))
	if a != b {
		t.Errorf("hash not deterministic")
	}
	if hashToken([]byte("x")) == hashToken([]byte("y")) {
		t.Errorf("different inputs must hash differently")
	}
}

func TestNewTokenBytes_LengthAndEntropy(t *testing.T) {
	a, err := newTokenBytes()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(a) != TokenBytes {
		t.Errorf("len: %d want %d", len(a), TokenBytes)
	}
	b, _ := newTokenBytes()
	if string(a) == string(b) {
		t.Error("two tokens must differ")
	}
}

func TestNewRowID_LengthAndHex(t *testing.T) {
	id, err := newRowID()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(id) != 32 {
		t.Errorf("len: %d", len(id))
	}
	if strings.Trim(id, "0123456789abcdef") != "" {
		t.Errorf("non-hex: %q", id)
	}
}
