package gdpr

import (
	"strings"
	"testing"
)

// REQ-COMP-GDPR-001: erasure must produce a deterministic hash so
// audit-chain joins (keyed on user_id) keep working without
// re-identifying the subject.
func TestAnonymisedEmailFor_Deterministic(t *testing.T) {
	a := AnonymisedEmailFor("u-1", "t-1")
	b := AnonymisedEmailFor("u-1", "t-1")
	if a != b {
		t.Errorf("not deterministic: %q vs %q", a, b)
	}
}

func TestAnonymisedEmailFor_Shape(t *testing.T) {
	got := AnonymisedEmailFor("11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222")
	if !strings.HasPrefix(got, "anon-") {
		t.Errorf("missing anon- prefix: %q", got)
	}
	rest := strings.TrimPrefix(got, "anon-")
	if len(rest) != 32 {
		t.Errorf("hash hex length: got %d want 32", len(rest))
	}
	for i := 0; i < len(rest); i++ {
		c := rest[i]
		switch {
		case c >= '0' && c <= '9', c >= 'a' && c <= 'f':
			// ok
		default:
			t.Errorf("non-hex char %q at %d in %q", c, i, rest)
			break
		}
	}
}

func TestAnonymisedEmailFor_DifferentInputsDiffer(t *testing.T) {
	cases := []struct{ a, b, c, d string }{
		{"u-1", "t-1", "u-2", "t-1"}, // different user
		{"u-1", "t-1", "u-1", "t-2"}, // different tenant
	}
	for _, tc := range cases {
		x := AnonymisedEmailFor(tc.a, tc.b)
		y := AnonymisedEmailFor(tc.c, tc.d)
		if x == y {
			t.Errorf("%v / %v collided to %q", []string{tc.a, tc.b}, []string{tc.c, tc.d}, x)
		}
	}
}
