// Package pagination implements an opaque, signed cursor based on a
// (created_at, id) tuple. Cursor encoding is "base64(timestamp_unix_ns
// + ":" + id)" with an HMAC tag so callers cannot forge a position past
// the boundary of a tenant.
package pagination

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MaxPageSize is the hard upper bound on any List RPC.
const MaxPageSize = 1000

// DefaultPageSize is the default if the caller does not request one.
const DefaultPageSize = 50

// ErrInvalidCursor is returned when a cursor cannot be parsed or
// authenticated.
var ErrInvalidCursor = errors.New("invalid cursor")

// Cursor is the decoded position of a paginated list.
type Cursor struct {
	// CreatedAt is the lexicographically-significant ordering key; rows
	// are sorted descending by this then by ID for stable paging.
	CreatedAt time.Time
	// ID disambiguates rows with identical CreatedAt.
	ID uuid.UUID
}

// Encode returns a URL-safe base64 string with an HMAC-SHA256 tag.
//
// Format on the wire (before base64): "<unix_ns>:<uuid>|<hex_tag>"
func Encode(secret []byte, c Cursor) string {
	body := fmt.Sprintf("%d:%s", c.CreatedAt.UnixNano(), c.ID)
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(body))
	tag := mac.Sum(nil)
	combined := body + "|" + base64.RawURLEncoding.EncodeToString(tag)
	return base64.RawURLEncoding.EncodeToString([]byte(combined))
}

// Decode validates the HMAC and returns the position. An empty cursor
// returns the zero Cursor and a nil error so handlers can treat
// "no cursor" as "first page".
func Decode(secret []byte, cursor string) (Cursor, error) {
	if cursor == "" {
		return Cursor{}, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return Cursor{}, ErrInvalidCursor
	}
	parts := strings.SplitN(string(raw), "|", 2)
	if len(parts) != 2 {
		return Cursor{}, ErrInvalidCursor
	}
	body, tagB64 := parts[0], parts[1]
	expected := hmac.New(sha256.New, secret)
	_, _ = expected.Write([]byte(body))
	wantTag := expected.Sum(nil)
	gotTag, err := base64.RawURLEncoding.DecodeString(tagB64)
	if err != nil || !hmac.Equal(wantTag, gotTag) {
		return Cursor{}, ErrInvalidCursor
	}
	bodyParts := strings.SplitN(body, ":", 2)
	if len(bodyParts) != 2 {
		return Cursor{}, ErrInvalidCursor
	}
	ns, err := strconv.ParseInt(bodyParts[0], 10, 64)
	if err != nil {
		return Cursor{}, ErrInvalidCursor
	}
	id, err := uuid.Parse(bodyParts[1])
	if err != nil {
		return Cursor{}, ErrInvalidCursor
	}
	return Cursor{CreatedAt: time.Unix(0, ns).UTC(), ID: id}, nil
}

// ClampPageSize returns a sane page size: at least 1, at most
// [MaxPageSize], with [DefaultPageSize] used when 0 is requested.
func ClampPageSize(requested int) int {
	if requested <= 0 {
		return DefaultPageSize
	}
	if requested > MaxPageSize {
		return MaxPageSize
	}
	return requested
}
