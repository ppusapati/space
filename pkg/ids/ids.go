// Package ids provides UUIDv7 and ULID generators with deterministic
// monotonic ordering inside a single process.
//
// UUIDv7 is preferred for primary keys because it embeds a millisecond
// timestamp, sorts naturally, and is index-friendly in PostgreSQL.
// ULIDs are exposed for cases where a 26-character base32 token is
// required (correlation IDs, file names).
package ids

import (
	"crypto/rand"
	mathrand "math/rand/v2"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

// NewUUIDv7 returns a fresh UUIDv7. UUIDv7 IDs are time-ordered to the
// millisecond and globally unique.
func NewUUIDv7() (uuid.UUID, error) {
	return uuid.NewV7()
}

// MustUUIDv7 is like NewUUIDv7 but panics on error. The underlying RNG
// only fails when the system is out of entropy, which on Linux/macOS
// is treated as fatal.
func MustUUIDv7() uuid.UUID {
	u, err := uuid.NewV7()
	if err != nil {
		panic("ids: UUIDv7 generation failed: " + err.Error())
	}
	return u
}

// ULIDGenerator is a goroutine-safe, monotonic ULID generator.
type ULIDGenerator struct {
	mu      sync.Mutex
	entropy *ulid.MonotonicEntropy
}

// NewULIDGenerator returns a generator seeded from the OS RNG.
func NewULIDGenerator() *ULIDGenerator {
	// Use math/rand/v2 with a CSPRNG-derived seed so we get a stream
	// without depending on math/rand's global state.
	var seed [32]byte
	if _, err := rand.Read(seed[:]); err != nil {
		panic("ids: cannot seed ULID entropy: " + err.Error())
	}
	src := mathrand.NewChaCha8(seed)
	entropy := ulid.Monotonic(newRandReader(src), 0)
	return &ULIDGenerator{entropy: entropy}
}

// Next returns the next monotonic ULID.
func (g *ULIDGenerator) Next() (ulid.ULID, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return ulid.New(ulid.Timestamp(time.Now()), g.entropy)
}

// MustNext returns the next ULID or panics on entropy exhaustion.
func (g *ULIDGenerator) MustNext() ulid.ULID {
	v, err := g.Next()
	if err != nil {
		panic("ids: ULID generation failed: " + err.Error())
	}
	return v
}

// chachaReader adapts a ChaCha8 source to io.Reader for ulid.Monotonic.
type chachaReader struct{ src *mathrand.ChaCha8 }

func newRandReader(src *mathrand.ChaCha8) *chachaReader { return &chachaReader{src: src} }

func (r *chachaReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = byte(r.src.Uint64() & 0xff)
	}
	return len(p), nil
}
