package oidc

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func defaultConfig() Config {
	return Config{
		Issuer:                "https://iam.test.chetana.p9e.in",
		AuthorizationEndpoint: "https://iam.test.chetana.p9e.in/oauth2/authorize",
		TokenEndpoint:         "https://iam.test.chetana.p9e.in/oauth2/token",
		UserInfoEndpoint:      "https://iam.test.chetana.p9e.in/oauth2/userinfo",
		JWKSURI:               "https://iam.test.chetana.p9e.in/.well-known/jwks.json",
		SupportedScopes:       []string{"profile", "email", "telemetry.read"},
	}
}

func TestBuildDocument_HappyPath(t *testing.T) {
	doc, err := BuildDocument(defaultConfig())
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if doc.Issuer != "https://iam.test.chetana.p9e.in" {
		t.Errorf("issuer: %q", doc.Issuer)
	}
	// "openid" must be auto-injected at the front of scopes.
	if doc.ScopesSupported[0] != "openid" {
		t.Errorf("scopes_supported missing openid: %v", doc.ScopesSupported)
	}
	// REQ-FUNC-PLT-IAM-006 acceptance #1: only S256 advertised.
	if len(doc.CodeChallengeMethodsSupported) != 1 || doc.CodeChallengeMethodsSupported[0] != "S256" {
		t.Errorf("code_challenge_methods_supported: %v", doc.CodeChallengeMethodsSupported)
	}
	// REQ-FUNC-PLT-IAM-006 acceptance #3: client_credentials advertised.
	want := map[string]bool{
		"authorization_code": true, "refresh_token": true, "client_credentials": true,
	}
	for _, g := range doc.GrantTypesSupported {
		delete(want, g)
	}
	if len(want) > 0 {
		t.Errorf("missing grant types: %v", want)
	}
	// chetana-specific claims must be advertised.
	for _, c := range []string{"tenant_id", "session_id", "is_us_person", "clearance_level"} {
		if !contains(doc.ClaimsSupported, c) {
			t.Errorf("claims_supported missing %q", c)
		}
	}
}

func TestBuildDocument_DoesNotDuplicateOpenIDScope(t *testing.T) {
	cfg := defaultConfig()
	cfg.SupportedScopes = []string{"openid", "profile"}
	doc, err := BuildDocument(cfg)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	count := 0
	for _, s := range doc.ScopesSupported {
		if s == "openid" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("openid appeared %d times in %v", count, doc.ScopesSupported)
	}
}

func TestBuildDocument_RejectsRelativeURLs(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*Config)
	}{
		{"relative issuer", func(c *Config) { c.Issuer = "/iam" }},
		{"relative authorization", func(c *Config) { c.AuthorizationEndpoint = "/auth" }},
		{"relative token", func(c *Config) { c.TokenEndpoint = "/token" }},
		{"relative jwks", func(c *Config) { c.JWKSURI = "/jwks" }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := defaultConfig()
			c.mut(&cfg)
			if _, err := BuildDocument(cfg); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestHandler_ServesValidJSON(t *testing.T) {
	doc, err := BuildDocument(defaultConfig())
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	srv := httptest.NewServer(Handler(doc))
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if got := resp.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("content-type: %q", got)
	}
	if !strings.HasPrefix(resp.Header.Get("Cache-Control"), "public") {
		t.Errorf("cache-control: %q", resp.Header.Get("Cache-Control"))
	}
	var got Document
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Issuer != doc.Issuer {
		t.Errorf("decoded issuer: %q want %q", got.Issuer, doc.Issuer)
	}
}

func contains(s []string, want string) bool {
	for _, v := range s {
		if v == want {
			return true
		}
	}
	return false
}
