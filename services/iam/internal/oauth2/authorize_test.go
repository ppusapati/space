package oauth2

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
)

// We can't construct an AuthCodeStore against a real DB in unit
// tests; the tests below either use scenarios that error BEFORE
// hitting the store, or stub the store via a tiny test helper.
//
// For end-to-end "issue + redeem" coverage, see the integration
// test in services/iam/test/oidc_e2e_test.go.

func clientForAuth() *Client {
	return &Client{
		ClientID:                "client-1",
		RedirectURIs:            []string{"https://app.chetana.p9e.in/cb"},
		GrantTypes:              []string{GrantAuthorizationCode, GrantRefreshToken},
		Scopes:                  []string{"openid", "profile", "telemetry.read"},
		TokenEndpointAuthMethod: AuthBasic,
	}
}

func validRequest(c *Client) AuthorizeRequest {
	return AuthorizeRequest{
		Client:              c,
		UserID:              "11111111-1111-1111-1111-111111111111",
		TenantID:            "22222222-2222-2222-2222-222222222222",
		SessionID:           "sess-1",
		ResponseType:        "code",
		RedirectURI:         "https://app.chetana.p9e.in/cb",
		Scopes:              []string{"openid", "telemetry.read"},
		State:               "opaque-state",
		CodeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		CodeChallengeMethod: MethodS256,
	}
}

// TestIssueCode_ValidationOrder covers every error path that fires
// before we touch the store.
func TestIssueCode_ValidationOrder(t *testing.T) {
	a := NewAuthorizer(nil) // store nil — these should never call into it
	cases := []struct {
		name string
		mut  func(*AuthorizeRequest)
		want error
	}{
		{"unsupported response_type", func(r *AuthorizeRequest) { r.ResponseType = "token" }, ErrUnsupportedResponseType},
		{"client lacks auth_code grant", func(r *AuthorizeRequest) { r.Client.GrantTypes = []string{GrantClientCredentials} }, ErrUnauthorizedClient},
		{"plain PKCE forbidden", func(r *AuthorizeRequest) { r.CodeChallengeMethod = "plain" }, ErrPlainMethodForbidden},
		{"missing PKCE method", func(r *AuthorizeRequest) { r.CodeChallengeMethod = "" }, ErrMissingChallengeMethod},
		{"unsupported PKCE method", func(r *AuthorizeRequest) { r.CodeChallengeMethod = "S512" }, ErrUnsupportedChallengeMethod},
		{"bad challenge shape", func(r *AuthorizeRequest) { r.CodeChallenge = "too-short" }, ErrInvalidChallenge},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Fresh client per subtest — the client_lacks_grant case
			// mutates the GrantTypes slice and we must not let that
			// bleed into the other cases.
			req := validRequest(clientForAuth())
			tc.mut(&req)
			_, err := a.IssueCode(context.Background(), req)
			if !errors.Is(err, tc.want) {
				t.Errorf("got %v want %v", err, tc.want)
			}
		})
	}
}

func TestIssueCode_UserMustBeAuthenticated(t *testing.T) {
	a := NewAuthorizer(nil)
	req := validRequest(clientForAuth())
	req.UserID = ""
	if _, err := a.IssueCode(context.Background(), req); err == nil {
		t.Fatal("expected error")
	}
}

func TestBuildErrorRedirect(t *testing.T) {
	got, err := BuildErrorRedirect(
		"https://app.chetana.p9e.in/cb?existing=1",
		"invalid_request", "code_verifier missing", "state-xyz",
	)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	q := parsed.Query()
	if q.Get("error") != "invalid_request" {
		t.Errorf("error: %q", q.Get("error"))
	}
	if q.Get("error_description") != "code_verifier missing" {
		t.Errorf("description: %q", q.Get("error_description"))
	}
	if q.Get("state") != "state-xyz" {
		t.Errorf("state: %q", q.Get("state"))
	}
	if q.Get("existing") != "1" {
		t.Errorf("existing query lost: %q", q.Get("existing"))
	}
	if !strings.HasPrefix(got, "https://app.chetana.p9e.in/cb?") {
		t.Errorf("redirect base: %q", got)
	}
}

func TestBuildErrorRedirect_BadURL(t *testing.T) {
	if _, err := BuildErrorRedirect("://bad-uri", "x", "y", "z"); err == nil {
		t.Error("expected error for malformed redirect URI")
	}
}
