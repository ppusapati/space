package pagination

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	secret := []byte("test-key")
	c := Cursor{
		CreatedAt: time.Unix(0, 1_700_000_000_123_456_789).UTC(),
		ID:        uuid.MustParse("01928fbb-9b9e-7000-8000-000000000001"),
	}
	enc := Encode(secret, c)
	got, err := Decode(secret, enc)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !got.CreatedAt.Equal(c.CreatedAt) || got.ID != c.ID {
		t.Fatalf("round-trip mismatch: %+v vs %+v", got, c)
	}
}

func TestDecodeEmptyReturnsZeroCursor(t *testing.T) {
	c, err := Decode([]byte("k"), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !c.CreatedAt.IsZero() || c.ID != uuid.Nil {
		t.Fatalf("expected zero cursor, got %+v", c)
	}
}

func TestDecodeRejectsTamperedTag(t *testing.T) {
	secret := []byte("test-key")
	c := Cursor{CreatedAt: time.Unix(1, 0), ID: uuid.MustParse("01928fbb-9b9e-7000-8000-000000000002")}
	enc := Encode(secret, c)
	// Flip a byte in the tag region — must fail HMAC verification.
	tampered := []byte(enc)
	tampered[len(tampered)-1] ^= 0x1
	if _, err := Decode(secret, string(tampered)); !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("expected ErrInvalidCursor, got %v", err)
	}
}

func TestDecodeRejectsForeignSecret(t *testing.T) {
	enc := Encode([]byte("alpha"), Cursor{CreatedAt: time.Unix(1, 0), ID: uuid.New()})
	if _, err := Decode([]byte("beta"), enc); !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("expected ErrInvalidCursor, got %v", err)
	}
}

func TestClampPageSize(t *testing.T) {
	cases := []struct {
		in, want int
	}{
		{0, DefaultPageSize},
		{-1, DefaultPageSize},
		{1, 1},
		{500, 500},
		{MaxPageSize, MaxPageSize},
		{MaxPageSize + 1, MaxPageSize},
	}
	for _, c := range cases {
		if got := ClampPageSize(c.in); got != c.want {
			t.Fatalf("ClampPageSize(%d)=%d, want %d", c.in, got, c.want)
		}
	}
}
