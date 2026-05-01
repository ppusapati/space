package ids

import (
	"testing"
)

func TestUUIDv7Monotonic(t *testing.T) {
	a := MustUUIDv7()
	b := MustUUIDv7()
	// UUIDv7 IDs minted in sequence within the same process must compare
	// `a <= b` lexically because the leading 48 bits encode milliseconds
	// since the Unix epoch.
	if a.String() > b.String() {
		t.Fatalf("UUIDv7 not monotonic: %s > %s", a, b)
	}
}

func TestUUIDv7Unique(t *testing.T) {
	seen := map[string]struct{}{}
	for i := 0; i < 1000; i++ {
		u := MustUUIDv7().String()
		if _, dup := seen[u]; dup {
			t.Fatalf("duplicate UUIDv7 at iteration %d: %s", i, u)
		}
		seen[u] = struct{}{}
	}
}

func TestULIDGeneratorMonotonic(t *testing.T) {
	g := NewULIDGenerator()
	a := g.MustNext()
	b := g.MustNext()
	if a.Compare(b) >= 0 {
		t.Fatalf("ULID not monotonic: %s >= %s", a, b)
	}
}
