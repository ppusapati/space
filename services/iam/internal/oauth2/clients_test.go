package oauth2

import (
	"errors"
	"reflect"
	"testing"
)

func TestClient_AllowsRedirectURI(t *testing.T) {
	c := &Client{
		ClientID: "c",
		RedirectURIs: []string{
			"https://app.chetana.p9e.in/callback",
			"http://127.0.0.1:5173/callback",
		},
	}
	cases := []struct {
		name string
		uri  string
		want error
	}{
		{"exact https", "https://app.chetana.p9e.in/callback", nil},
		{"loopback", "http://127.0.0.1:5173/callback", nil},
		{"http non-loopback rejected", "http://evil.com/cb", ErrInvalidRedirectURI},
		{"mismatched path", "https://app.chetana.p9e.in/other", ErrRedirectURIMismatch},
		{"case sensitive", "https://App.chetana.p9e.in/callback", ErrRedirectURIMismatch},
		{"trailing slash differs", "https://app.chetana.p9e.in/callback/", ErrRedirectURIMismatch},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := c.AllowsRedirectURI(tc.uri)
			if tc.want == nil && err != nil {
				t.Errorf("unexpected error %v", err)
			}
			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Errorf("got %v want %v", err, tc.want)
			}
		})
	}
}

func TestClient_AllowsRedirectURI_OmittedWithSingleRegistered(t *testing.T) {
	c := &Client{RedirectURIs: []string{"https://app.chetana.p9e.in/cb"}}
	got, err := c.AllowsRedirectURI("")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != "https://app.chetana.p9e.in/cb" {
		t.Errorf("default uri: %q", got)
	}
}

func TestClient_AllowsRedirectURI_OmittedWithMultiple(t *testing.T) {
	c := &Client{RedirectURIs: []string{"https://a/cb", "https://b/cb"}}
	if _, err := c.AllowsRedirectURI(""); !errors.Is(err, ErrRedirectURIRequired) {
		t.Errorf("got %v want ErrRedirectURIRequired", err)
	}
}

func TestClient_IntersectScopes(t *testing.T) {
	c := &Client{Scopes: []string{"openid", "profile", "email", "telemetry.read"}}
	got := c.IntersectScopes([]string{"openid", "missing", "telemetry.read", "telemetry.write"})
	want := []string{"openid", "telemetry.read"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("intersect: got %v want %v", got, want)
	}

	// Empty request → full allow-list, defensively copied.
	got = c.IntersectScopes(nil)
	if !reflect.DeepEqual(got, c.Scopes) {
		t.Errorf("default: got %v want %v", got, c.Scopes)
	}
	got[0] = "MUTATED"
	if c.Scopes[0] == "MUTATED" {
		t.Error("IntersectScopes must defensively copy when defaulting")
	}
}

func TestClient_IsPublic(t *testing.T) {
	cases := []struct {
		name string
		c    Client
		want bool
	}{
		{"none method", Client{TokenEndpointAuthMethod: AuthNone}, true},
		{"empty hash", Client{TokenEndpointAuthMethod: AuthBasic, ClientSecretHash: ""}, true},
		{"basic with hash", Client{TokenEndpointAuthMethod: AuthBasic, ClientSecretHash: "$argon2id$..."}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.c.IsPublic(); got != tc.want {
				t.Errorf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestSplitJoinScope_Roundtrip(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"  ", nil},
		{"openid profile email", []string{"openid", "profile", "email"}},
		{"  openid   profile  ", []string{"openid", "profile"}},
	}
	for _, c := range cases {
		got := SplitScope(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("SplitScope(%q): got %v want %v", c.in, got, c.want)
		}
	}

	if got := JoinScope([]string{"openid", "profile"}); got != "openid profile" {
		t.Errorf("JoinScope: %q", got)
	}
}

func TestConstantTimeCompareStrings(t *testing.T) {
	if !ConstantTimeCompareStrings("abc", "abc") {
		t.Error("equal must return true")
	}
	if ConstantTimeCompareStrings("abc", "abd") {
		t.Error("different must return false")
	}
	if ConstantTimeCompareStrings("abc", "abcd") {
		t.Error("different lengths must return false")
	}
}

func TestHashClientSecret_Roundtrip(t *testing.T) {
	hash, err := HashClientSecret("hunter2")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "" {
		t.Error("empty hash")
	}
	if _, err := HashClientSecret(""); err == nil {
		t.Error("empty secret should error")
	}
}
