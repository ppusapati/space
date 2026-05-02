// token.go — /oauth2/token handler.
//
// Implements three grants per OAuth 2.1 §1.3:
//
//   • authorization_code  (with PKCE)
//   • refresh_token       (single-use rotation, family invalidation)
//   • client_credentials  (machine-to-machine, no user)
//
// Common pieces:
//
//   • Client authentication first. Methods supported:
//       client_secret_basic — HTTP Basic header
//       client_secret_post  — client_id + client_secret in form
//       none                — public clients (PKCE-only)
//     The chosen method MUST match the client's registered
//     token_endpoint_auth_method. See ResolveClient.
//
//   • Access tokens are minted via internal/token.Issuer; refresh
//     tokens via internal/token.RefreshStore. The same shapes the
//     login flow uses, so an OAuth-issued JWT is verified by the
//     same authz/v1.Verifier that handles login JWTs.

package oauth2

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ppusapati/space/services/iam/internal/token"
)

// TokenIssuer is the access-token + refresh-token surface this
// package consumes. internal/token.Issuer + internal/token.RefreshStore
// satisfy it via the small TokenAdapter the cmd layer constructs.
type TokenIssuer interface {
	IssueAccess(ctx context.Context, in token.IssueInput) (accessToken string, expiresAt time.Time, err error)
	IssueRefresh(ctx context.Context, userID, tenantID, sessionID string) (refreshToken string, expiresAt time.Time, err error)
	RotateRefresh(ctx context.Context, presented string) (refreshToken string, expiresAt time.Time, userID, tenantID, sessionID string, err error)
}

// TokenResponse is the JSON shape per RFC 6749 §5.1 + OIDC core
// §3.1.3.3.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"` // always "Bearer"
	ExpiresIn    int64  `json:"expires_in"` // seconds
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

// TokenHandler glues the client store, auth-code store, and token
// issuer. Construct with NewTokenHandler.
type TokenHandler struct {
	clients *ClientStore
	codes   *AuthCodeStore
	tokens  TokenIssuer
	clk     func() time.Time
}

// NewTokenHandler builds the token endpoint over the supplied
// dependencies.
func NewTokenHandler(clients *ClientStore, codes *AuthCodeStore, tokens TokenIssuer, clock func() time.Time) *TokenHandler {
	if clock == nil {
		clock = time.Now
	}
	return &TokenHandler{
		clients: clients,
		codes:   codes,
		tokens:  tokens,
		clk:     clock,
	}
}

// TokenRequest is the parsed form of an incoming /oauth2/token
// request. The caller (Connect or vanilla HTTP handler) builds
// this from the application/x-www-form-urlencoded body + Basic
// header and calls Exchange.
type TokenRequest struct {
	GrantType string

	// Client authentication. Either ClientSecret or BasicHeader is
	// populated depending on the auth method.
	ClientID     string
	ClientSecret string
	BasicHeader  string // raw "Authorization" header value, optional

	// authorization_code grant.
	Code         string
	RedirectURI  string
	CodeVerifier string

	// refresh_token grant.
	RefreshToken string

	// shared.
	Scope string
}

// Exchange dispatches on grant_type and returns a TokenResponse
// the HTTP handler serialises as JSON.
func (h *TokenHandler) Exchange(ctx context.Context, req TokenRequest) (*TokenResponse, error) {
	client, err := h.resolveClient(ctx, req)
	if err != nil {
		return nil, err
	}
	switch req.GrantType {
	case GrantAuthorizationCode:
		return h.exchangeAuthorizationCode(ctx, client, req)
	case GrantRefreshToken:
		return h.exchangeRefreshToken(ctx, client, req)
	case GrantClientCredentials:
		return h.exchangeClientCredentials(ctx, client, req)
	case "":
		return nil, fmt.Errorf("%w: grant_type is required", ErrInvalidRequest)
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedGrantType, req.GrantType)
	}
}

// resolveClient extracts the client_id / client_secret pair from
// either the Basic header or the form body, then authenticates.
//
// The chosen channel MUST match the client's registered
// token_endpoint_auth_method (RFC 6749 §2.3.1 + OIDC core §9).
func (h *TokenHandler) resolveClient(ctx context.Context, req TokenRequest) (*Client, error) {
	if req.BasicHeader != "" {
		id, secret, ok := decodeBasic(req.BasicHeader)
		if !ok {
			return nil, ErrClientAuthFailed
		}
		c, err := h.clients.Authenticate(ctx, id, secret)
		if err != nil {
			return nil, err
		}
		if c.TokenEndpointAuthMethod != AuthBasic && c.TokenEndpointAuthMethod != "" {
			return nil, fmt.Errorf("%w: client must authenticate via %s",
				ErrClientAuthFailed, c.TokenEndpointAuthMethod)
		}
		return c, nil
	}
	// Body-based: secret_post for confidential clients, none for public.
	if req.ClientID == "" {
		return nil, ErrClientAuthFailed
	}
	c, err := h.clients.Authenticate(ctx, req.ClientID, req.ClientSecret)
	if err != nil {
		return nil, err
	}
	switch c.TokenEndpointAuthMethod {
	case AuthBasic:
		return nil, fmt.Errorf("%w: client requires HTTP Basic auth", ErrClientAuthFailed)
	case AuthNone:
		if req.ClientSecret != "" {
			return nil, ErrClientAuthFailed
		}
	}
	return c, nil
}

// exchangeAuthorizationCode redeems an auth code + verifies PKCE +
// mints (access, refresh).
func (h *TokenHandler) exchangeAuthorizationCode(ctx context.Context, client *Client, req TokenRequest) (*TokenResponse, error) {
	if !client.AllowsGrant(GrantAuthorizationCode) {
		return nil, ErrUnauthorizedClient
	}
	if req.Code == "" {
		return nil, fmt.Errorf("%w: code is required", ErrInvalidRequest)
	}
	if req.CodeVerifier == "" {
		return nil, fmt.Errorf("%w: code_verifier is required (PKCE)", ErrInvalidRequest)
	}

	rec, err := h.codes.Redeem(ctx, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, ErrAuthCodeNotFound):
			return nil, fmt.Errorf("%w: code not found or malformed", ErrInvalidGrant)
		case errors.Is(err, ErrAuthCodeExpired):
			return nil, fmt.Errorf("%w: code expired", ErrInvalidGrant)
		case errors.Is(err, ErrAuthCodeReused):
			return nil, fmt.Errorf("%w: code already redeemed", ErrInvalidGrant)
		}
		return nil, fmt.Errorf("%w: %v", ErrServer, err)
	}
	if rec.ClientID != client.ClientID {
		return nil, fmt.Errorf("%w: code was issued to a different client", ErrInvalidGrant)
	}
	if req.RedirectURI != rec.RedirectURI {
		return nil, fmt.Errorf("%w: redirect_uri does not match the value used at /authorize", ErrInvalidGrant)
	}
	if rec.CodeChallengeMethod != MethodS256 {
		// Defence in depth — IssueCode pinned this to S256, but
		// double-check at redemption.
		return nil, fmt.Errorf("%w: code was not bound to PKCE S256", ErrInvalidGrant)
	}
	if err := VerifyVerifier(req.CodeVerifier, rec.CodeChallenge); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidGrant, err)
	}

	access, accessExpires, err := h.tokens.IssueAccess(ctx, token.IssueInput{
		UserID:    rec.UserID,
		TenantID:  rec.TenantID,
		SessionID: rec.SessionID,
		Scopes:    rec.Scopes,
		Audience:  []string{"chetana-api"},
		AMR:       []string{"pwd"},
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrServer, err)
	}

	resp := &TokenResponse{
		AccessToken: access,
		TokenType:   "Bearer",
		ExpiresIn:   int64(time.Until(accessExpires).Seconds()),
		Scope:       JoinScope(rec.Scopes),
	}
	if client.AllowsGrant(GrantRefreshToken) {
		refresh, _, err := h.tokens.IssueRefresh(ctx, rec.UserID, rec.TenantID, rec.SessionID)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrServer, err)
		}
		resp.RefreshToken = refresh
	}

	// OIDC: when the openid scope is present we also mint an
	// id_token. We piggyback on the same access-token issuer (RS256
	// + iss + aud); the audience is the client_id per OIDC core
	// §3.1.3.7 step 3. The nonce stored at /authorize is echoed.
	if hasScope(rec.Scopes, "openid") {
		idToken, _, err := h.tokens.IssueAccess(ctx, token.IssueInput{
			UserID:    rec.UserID,
			TenantID:  rec.TenantID,
			SessionID: rec.SessionID,
			Audience:  []string{client.ClientID},
			AMR:       []string{"pwd"},
		})
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrServer, err)
		}
		resp.IDToken = idToken
	}
	return resp, nil
}

func hasScope(scopes []string, want string) bool {
	for _, s := range scopes {
		if s == want {
			return true
		}
	}
	return false
}

// exchangeRefreshToken rotates a refresh token and re-issues an
// access token. Reuse detection is owned by the underlying
// RefreshStore; we forward whatever error it returns.
func (h *TokenHandler) exchangeRefreshToken(ctx context.Context, client *Client, req TokenRequest) (*TokenResponse, error) {
	if !client.AllowsGrant(GrantRefreshToken) {
		return nil, ErrUnauthorizedClient
	}
	if req.RefreshToken == "" {
		return nil, fmt.Errorf("%w: refresh_token is required", ErrInvalidRequest)
	}

	newRefresh, _, userID, tenantID, sessionID, err := h.tokens.RotateRefresh(ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidGrant, err)
	}

	scopes := SplitScope(req.Scope)
	if len(scopes) > 0 {
		scopes = client.IntersectScopes(scopes)
	}
	access, accessExpires, err := h.tokens.IssueAccess(ctx, token.IssueInput{
		UserID:    userID,
		TenantID:  tenantID,
		SessionID: sessionID,
		Scopes:    scopes,
		Audience:  []string{"chetana-api"},
		AMR:       []string{"pwd"},
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrServer, err)
	}
	return &TokenResponse{
		AccessToken:  access,
		TokenType:    "Bearer",
		ExpiresIn:    int64(time.Until(accessExpires).Seconds()),
		RefreshToken: newRefresh,
		Scope:        JoinScope(scopes),
	}, nil
}

// exchangeClientCredentials issues an access token for a m2m
// client. No user, no refresh token (RFC 6749 §4.4.3).
func (h *TokenHandler) exchangeClientCredentials(ctx context.Context, client *Client, req TokenRequest) (*TokenResponse, error) {
	if !client.AllowsGrant(GrantClientCredentials) {
		return nil, ErrUnauthorizedClient
	}
	if client.IsPublic() {
		return nil, fmt.Errorf("%w: public clients cannot use client_credentials", ErrUnauthorizedClient)
	}
	scopes := client.IntersectScopes(SplitScope(req.Scope))

	// For m2m, the subject is the client_id itself (per the
	// "service account" convention used by every major IdP).
	access, accessExpires, err := h.tokens.IssueAccess(ctx, token.IssueInput{
		UserID:    client.ClientID,
		TenantID:  "", // m2m token: no tenant context
		SessionID: "client-credentials-" + client.ClientID,
		Scopes:    scopes,
		Audience:  []string{"chetana-api"},
		AMR:       []string{"client_credentials"},
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrServer, err)
	}
	return &TokenResponse{
		AccessToken: access,
		TokenType:   "Bearer",
		ExpiresIn:   int64(time.Until(accessExpires).Seconds()),
		Scope:       JoinScope(scopes),
	}, nil
}

// decodeBasic parses an "Authorization: Basic <b64>" header value
// into (client_id, client_secret).
func decodeBasic(header string) (string, string, bool) {
	const prefix = "Basic "
	if !strings.HasPrefix(header, prefix) {
		return "", "", false
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(header[len(prefix):]))
	if err != nil {
		return "", "", false
	}
	idx := strings.IndexByte(string(raw), ':')
	if idx < 0 {
		return "", "", false
	}
	return string(raw[:idx]), string(raw[idx+1:]), true
}

// WriteJSONError writes the canonical RFC 6749 §5.2 error envelope.
// Exposed so the HTTP layer (or the Connect bridge) can produce a
// consistent shape.
func WriteJSONError(w http.ResponseWriter, status int, errCode, description string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(status)
	body := fmt.Sprintf(`{"error":%q,"error_description":%q}`, errCode, description)
	_, _ = w.Write([]byte(body))
}

// ----------------------------------------------------------------------
// Errors (token endpoint specific — distinct from the /authorize set
// because the canonical RFC 6749 §5.2 error_codes are different)
// ----------------------------------------------------------------------

// ErrInvalidRequest is RFC 6749 §5.2 invalid_request.
var ErrInvalidRequest = errors.New("oauth2: invalid_request")

// ErrInvalidGrant is RFC 6749 §5.2 invalid_grant.
var ErrInvalidGrant = errors.New("oauth2: invalid_grant")

// ErrUnsupportedGrantType is RFC 6749 §5.2 unsupported_grant_type.
var ErrUnsupportedGrantType = errors.New("oauth2: unsupported_grant_type")
