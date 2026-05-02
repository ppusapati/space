// authorize.go — /oauth2/authorize handler.
//
// The authorisation endpoint is the entry point of the auth-code
// flow. The handler is split in two:
//
//   • AuthorizeRequest is the parsed + validated form of the
//     incoming query parameters. The caller (the Connect layer
//     once iam.proto regenerates) parses the request, runs the
//     login-required prompt machinery, and then hands a fully
//     populated AuthorizeRequest plus the authenticated user to
//     IssueCode.
//
//   • IssueCode binds (client, user, redirect, scopes, PKCE
//     challenge, nonce) into a one-shot auth code via the
//     AuthCodeStore and returns the redirect URL the caller sends
//     to the user-agent.
//
// We keep the HTTP-shaped layer thin so the same logic can serve
// the future Connect RPC variant the SPA will use.

package oauth2

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// AuthorizeRequest is the validated form of the /oauth2/authorize
// query parameters. The caller MUST set every field — IssueCode
// trusts what it is given.
type AuthorizeRequest struct {
	Client              *Client
	UserID              string
	TenantID            string
	SessionID           string
	ResponseType        string // must be "code"
	RedirectURI         string // already validated by Client.AllowsRedirectURI
	Scopes              []string
	State               string // echoed back to the client; opaque
	CodeChallenge       string
	CodeChallengeMethod string
	Nonce               string // OIDC; echoed into id_token at /token
}

// Authorizer ties an AuthCodeStore + ClientStore for the
// authorisation endpoint. Construct with NewAuthorizer.
type Authorizer struct {
	codes *AuthCodeStore
}

// NewAuthorizer builds an Authorizer over the supplied stores.
func NewAuthorizer(codes *AuthCodeStore) *Authorizer {
	return &Authorizer{codes: codes}
}

// IssueCode validates `req` and mints an auth code. Returns the
// fully composed redirect URL — the caller HTTP-redirects the
// user-agent there.
//
// Validation order (matches OAuth 2.1 §4.1.2 + §5):
//
//   1. response_type == "code"          → ErrUnsupportedResponseType
//   2. client allows the auth_code grant → ErrUnauthorizedClient
//   3. PKCE method validates             → returns the underlying
//                                          PKCE error
//   4. PKCE challenge has the right shape→ ErrInvalidChallenge
//   5. scopes are within the client's allow-list → keeps only
//                                          the intersection
//   6. mint + persist the code            → ErrServer on store error
func (a *Authorizer) IssueCode(ctx context.Context, req AuthorizeRequest) (string, error) {
	if req.Client == nil {
		return "", errors.New("oauth2: nil client")
	}
	if req.ResponseType != "code" {
		return "", fmt.Errorf("%w: %q", ErrUnsupportedResponseType, req.ResponseType)
	}
	if !req.Client.AllowsGrant(GrantAuthorizationCode) {
		return "", ErrUnauthorizedClient
	}
	if err := ValidateMethod(req.CodeChallengeMethod); err != nil {
		return "", err
	}
	if err := ValidateChallengeShape(req.CodeChallenge); err != nil {
		return "", err
	}
	if req.UserID == "" {
		return "", errors.New("oauth2: user must be authenticated before IssueCode")
	}

	scopes := req.Client.IntersectScopes(req.Scopes)

	out, err := a.codes.Issue(ctx, AuthCodeIssue{
		ClientID:            req.Client.ClientID,
		UserID:              req.UserID,
		TenantID:            req.TenantID,
		SessionID:           req.SessionID,
		RedirectURI:         req.RedirectURI,
		Scopes:              scopes,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		Nonce:               req.Nonce,
	})
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrServer, err)
	}

	parsed, err := url.Parse(req.RedirectURI)
	if err != nil {
		return "", fmt.Errorf("oauth2: parse redirect: %w", err)
	}
	q := parsed.Query()
	q.Set("code", out.Code)
	if req.State != "" {
		q.Set("state", req.State)
	}
	parsed.RawQuery = q.Encode()
	return parsed.String(), nil
}

// BuildErrorRedirect composes an OAuth-error redirect URL per
// RFC 6749 §4.1.2.1. The error code is the spec-defined string
// (`invalid_request`, `unauthorized_client`, `invalid_scope`, …);
// description is opaque text shown to the developer.
func BuildErrorRedirect(redirectURI, errCode, errDescription, state string) (string, error) {
	parsed, err := url.Parse(redirectURI)
	if err != nil {
		return "", fmt.Errorf("oauth2: parse redirect: %w", err)
	}
	q := parsed.Query()
	q.Set("error", errCode)
	if errDescription != "" {
		q.Set("error_description", errDescription)
	}
	if state != "" {
		q.Set("state", state)
	}
	parsed.RawQuery = q.Encode()
	return parsed.String(), nil
}

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrUnsupportedResponseType is returned when response_type is
// anything other than "code".
var ErrUnsupportedResponseType = errors.New("oauth2: unsupported response_type")

// ErrUnauthorizedClient is returned when the client's allow-list
// does not include the auth_code grant.
var ErrUnauthorizedClient = errors.New("oauth2: client is not authorised for this grant")

// ErrServer is the generic wrapper for downstream-store failures.
// The /oauth2 RPC translates this to "server_error".
var ErrServer = errors.New("oauth2: server error")
