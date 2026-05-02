// Package oidc implements the OpenID Connect Discovery 1.0 endpoint
// for the chetana IAM service.
//
// → REQ-FUNC-PLT-IAM-006 acceptance #2: discovery doc validates
//   against OIDC Discovery 1.0.
// → design.md §4.1.1.
//
// The handler emits a static, build-time-frozen JSON document at
// /.well-known/openid-configuration listing the issuer's URLs +
// supported algorithms + supported scopes / response types /
// grant types. coreos/go-oidc validates this shape at boot when a
// downstream service constructs a Provider from the issuer URL.

package oidc

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// Config supplies the deployment-specific knobs for the discovery
// document. All URL fields MUST be absolute https (loopback http
// is permitted for dev only).
type Config struct {
	Issuer                string // canonical iss claim, e.g. "https://iam.us-east.chetana.p9e.in"
	AuthorizationEndpoint string // e.g. <issuer>/oauth2/authorize
	TokenEndpoint         string // e.g. <issuer>/oauth2/token
	UserInfoEndpoint      string // e.g. <issuer>/oauth2/userinfo
	JWKSURI               string // e.g. <issuer>/.well-known/jwks.json
	RegistrationEndpoint  string // optional; empty when DCR is disabled

	// SupportedScopes is the union of scopes any registered client
	// may request. Per OIDC §3 the special "openid" scope MUST be
	// included.
	SupportedScopes []string
}

// Document is the JSON shape returned by /.well-known/openid-
// configuration per OIDC Discovery 1.0 §3 + §4.
//
// Fields are ordered so the JSON output matches the spec's
// example. The set published here is the chetana-specific subset
// that's actually true of the deployment — unsupported features
// are NOT advertised (the spec explicitly says omit-rather-than-
// lie).
type Document struct {
	Issuer                                     string   `json:"issuer"`
	AuthorizationEndpoint                      string   `json:"authorization_endpoint"`
	TokenEndpoint                              string   `json:"token_endpoint"`
	UserInfoEndpoint                           string   `json:"userinfo_endpoint,omitempty"`
	JWKSURI                                    string   `json:"jwks_uri"`
	RegistrationEndpoint                       string   `json:"registration_endpoint,omitempty"`
	ScopesSupported                            []string `json:"scopes_supported"`
	ResponseTypesSupported                     []string `json:"response_types_supported"`
	GrantTypesSupported                        []string `json:"grant_types_supported"`
	SubjectTypesSupported                      []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported           []string `json:"id_token_signing_alg_values_supported"`
	TokenEndpointAuthMethodsSupported          []string `json:"token_endpoint_auth_methods_supported"`
	CodeChallengeMethodsSupported              []string `json:"code_challenge_methods_supported"`
	ClaimsSupported                            []string `json:"claims_supported"`
	RequireRequestURIRegistration              bool     `json:"require_request_uri_registration"`
	RequestParameterSupported                  bool     `json:"request_parameter_supported"`
	RequestURIParameterSupported               bool     `json:"request_uri_parameter_supported"`
	ClaimsParameterSupported                   bool     `json:"claims_parameter_supported"`
}

// BuildDocument validates the supplied config and returns a fully
// populated Document. Returns an error when the config is missing
// a required field or carries an obviously bad URL.
func BuildDocument(cfg Config) (*Document, error) {
	if err := validate(cfg); err != nil {
		return nil, err
	}
	scopes := cfg.SupportedScopes
	if !containsString(scopes, "openid") {
		scopes = append([]string{"openid"}, scopes...)
	}
	return &Document{
		Issuer:                cfg.Issuer,
		AuthorizationEndpoint: cfg.AuthorizationEndpoint,
		TokenEndpoint:         cfg.TokenEndpoint,
		UserInfoEndpoint:      cfg.UserInfoEndpoint,
		JWKSURI:               cfg.JWKSURI,
		RegistrationEndpoint:  cfg.RegistrationEndpoint,

		ScopesSupported: scopes,
		// chetana implements only auth-code; OAuth 2.1 has dropped
		// implicit + ROPC, and we don't expose hybrid flows.
		ResponseTypesSupported: []string{"code"},
		GrantTypesSupported: []string{
			"authorization_code",
			"refresh_token",
			"client_credentials",
		},
		// Public subject_type — the iss + sub pair is a stable
		// global identity. OIDC §8 also defines "pairwise"; we
		// don't implement it.
		SubjectTypesSupported: []string{"public"},
		// We sign id_tokens with RS256 (matches the access JWT).
		IDTokenSigningAlgValuesSupported: []string{"RS256"},
		// Accepted token-endpoint client-auth methods.
		TokenEndpointAuthMethodsSupported: []string{
			"client_secret_basic",
			"client_secret_post",
			"none",
		},
		// PKCE: S256 only — plain is forbidden by OAuth 2.1.
		CodeChallengeMethodsSupported: []string{"S256"},
		// Claims the platform may emit. Tracks the access JWT
		// claim set in services/iam/internal/token/jwt.go.
		ClaimsSupported: []string{
			"sub", "iss", "aud", "exp", "iat", "nbf", "jti",
			"tenant_id", "session_id",
			"is_us_person", "clearance_level", "nationality",
			"roles", "scopes", "amr",
		},
		// Request-object features off — chetana only supports the
		// vanilla query-parameter authorisation request.
		RequestParameterSupported:    false,
		RequestURIParameterSupported: false,
		ClaimsParameterSupported:     false,
	}, nil
}

// Handler returns an http.Handler that serves the discovery
// document at /.well-known/openid-configuration with a 1-hour
// cache window. Mounted by cmd/iam at boot.
func Handler(doc *Document) http.Handler {
	body, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		// Build-time failure should not happen — Document is a
		// pure-data struct — but if it does, surface a 500.
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "marshal", http.StatusInternalServerError)
		})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		_, _ = w.Write(body)
	})
}

// ----------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------

func validate(cfg Config) error {
	if !isAbsoluteURL(cfg.Issuer) {
		return errors.New("oidc: Issuer must be an absolute URL")
	}
	if !isAbsoluteURL(cfg.AuthorizationEndpoint) {
		return errors.New("oidc: AuthorizationEndpoint must be an absolute URL")
	}
	if !isAbsoluteURL(cfg.TokenEndpoint) {
		return errors.New("oidc: TokenEndpoint must be an absolute URL")
	}
	if !isAbsoluteURL(cfg.JWKSURI) {
		return errors.New("oidc: JWKSURI must be an absolute URL")
	}
	if cfg.UserInfoEndpoint != "" && !isAbsoluteURL(cfg.UserInfoEndpoint) {
		return errors.New("oidc: UserInfoEndpoint must be an absolute URL when set")
	}
	return nil
}

func isAbsoluteURL(s string) bool {
	return strings.HasPrefix(s, "https://") ||
		strings.HasPrefix(s, "http://localhost") ||
		strings.HasPrefix(s, "http://127.0.0.1")
}

func containsString(s []string, want string) bool {
	for _, v := range s {
		if v == want {
			return true
		}
	}
	return false
}
