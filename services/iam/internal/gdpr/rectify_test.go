package gdpr

import (
	"errors"
	"testing"
)

func TestLooksLikeEmail(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"user@example.com", true},
		{"a@b.c", true},
		{"", false},
		{"a", false},
		{"@example.com", false},
		{"user@", false},
		{"user@nodomain", false},
		{"user with spaces@x.y", false},
		{"u@x.y\n", false},
	}
	for _, tc := range cases {
		if got := looksLikeEmail(tc.in); got != tc.want {
			t.Errorf("looksLikeEmail(%q): got %v want %v", tc.in, got, tc.want)
		}
	}
}

func TestLooksLikeEmail_LengthBounds(t *testing.T) {
	// 320 chars max per the function's guard.
	long := "x@" + repeatDot(318)
	if !looksLikeEmail(long) {
		t.Errorf("at length cap should be valid")
	}
	if looksLikeEmail("x@" + repeatDot(319)) {
		t.Errorf("over length cap should be invalid")
	}
}

func repeatDot(n int) string {
	out := make([]byte, n)
	for i := range out {
		if i%2 == 0 {
			out[i] = 'a'
		} else {
			out[i] = '.'
		}
	}
	return string(out)
}

func TestErrors_Sentinel(t *testing.T) {
	for _, e := range []error{
		ErrUserNotFound,
		ErrInvalidEmail,
		ErrEmailInUse,
		ErrAlreadyErased,
	} {
		if !errors.Is(e, e) {
			t.Errorf("not reflexive: %v", e)
		}
		if e.Error() == "" {
			t.Errorf("empty error: %v", e)
		}
	}
}
