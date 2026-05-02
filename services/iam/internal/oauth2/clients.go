// clients.go — oauth2_clients persistence + client authentication.
//
// The shape mirrors RFC 6749 §2 + RFC 7591:
//
//   client_id          : opaque, unique. UUID-shaped in production.
//   client_secret_hash : argon2id(client_secret). NULL for public
//                        clients (e.g. SPAs); those rely on PKCE
//                        alone for proof-of-possession.
//   redirect_uris      : exact-match allow-list. OAuth 2.1 §1.4.2
//                        forbids glob/scheme-host-only matching.
//   grant_types        : subset of {"authorization_code",
//                        "refresh_token", "client_credentials"}.
//   scopes             : the maximum scope set this client may
//                        request; the actual scopes minted on a
//                        successful exchange is the intersection of
//                        what was requested and this allow-list.
//   token_endpoint_auth_method : "client_secret_basic" |
//                        "client_secret_post" | "none". "none" is
//                        only valid when client_secret_hash is NULL.

package oauth2

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ppusapati/space/services/iam/internal/password"
)

// Canonical grant-type identifiers.
const (
	GrantAuthorizationCode = "authorization_code"
	GrantRefreshToken      = "refresh_token"
	GrantClientCredentials = "client_credentials"
)

// Canonical token-endpoint auth methods we support.
const (
	AuthBasic  = "client_secret_basic"
	AuthPost   = "client_secret_post"
	AuthNone   = "none"
)

// Client is the in-memory shape of one oauth2_clients row.
type Client struct {
	ClientID                string
	ClientSecretHash        string // PHC-encoded argon2id; empty for public clients
	RedirectURIs            []string
	GrantTypes              []string
	Scopes                  []string
	TokenEndpointAuthMethod string
	Disabled                bool
	CreatedAt               time.Time
}

// IsPublic reports whether the client is a public (no-secret)
// client. Public clients MUST authenticate with PKCE alone.
func (c *Client) IsPublic() bool {
	return c.TokenEndpointAuthMethod == AuthNone || c.ClientSecretHash == ""
}

// AllowsGrant reports whether this client's allow-list permits the
// given grant type.
func (c *Client) AllowsGrant(grant string) bool {
	for _, g := range c.GrantTypes {
		if g == grant {
			return true
		}
	}
	return false
}

// AllowsRedirectURI returns nil when the redirect URI exactly
// matches one of the registered values (case-sensitive,
// byte-for-byte). Per OAuth 2.1 §1.4.2 wildcards are forbidden.
//
// Empty input is treated as "no redirect URI provided" — the
// authorisation server falls back to the SOLE registered URI when
// exactly one is configured (RFC 6749 §3.1.2.3).
func (c *Client) AllowsRedirectURI(uri string) (string, error) {
	if uri == "" {
		if len(c.RedirectURIs) == 1 {
			return c.RedirectURIs[0], nil
		}
		return "", ErrRedirectURIRequired
	}
	if !isAbsoluteHTTPS(uri) && !isLoopback(uri) {
		return "", ErrInvalidRedirectURI
	}
	for _, r := range c.RedirectURIs {
		if r == uri {
			return uri, nil
		}
	}
	return "", ErrRedirectURIMismatch
}

// IntersectScopes returns the intersection of `requested` with the
// client's allow-list, preserving the requested order. Empty
// `requested` defaults to the client's full allow-list (per common
// authorisation-server convention).
func (c *Client) IntersectScopes(requested []string) []string {
	if len(requested) == 0 {
		out := make([]string, len(c.Scopes))
		copy(out, c.Scopes)
		return out
	}
	allow := make(map[string]bool, len(c.Scopes))
	for _, s := range c.Scopes {
		allow[s] = true
	}
	out := make([]string, 0, len(requested))
	for _, s := range requested {
		if allow[s] {
			out = append(out, s)
		}
	}
	return out
}

// ClientStore wraps a pgxpool.Pool with the OAuth2 client lookup +
// authentication helpers.
type ClientStore struct {
	pool *pgxpool.Pool
}

// NewClientStore wraps a pool.
func NewClientStore(pool *pgxpool.Pool) *ClientStore {
	return &ClientStore{pool: pool}
}

// LookupByID fetches a client by its client_id. Returns
// ErrClientNotFound when no row matches OR the row is disabled.
func (s *ClientStore) LookupByID(ctx context.Context, clientID string) (*Client, error) {
	if clientID == "" {
		return nil, ErrClientNotFound
	}
	const q = `
SELECT client_id, COALESCE(client_secret_hash, ''),
       redirect_uris, grant_types, scopes,
       token_endpoint_auth_method, disabled, created_at
FROM oauth2_clients
WHERE client_id = $1
`
	var c Client
	err := s.pool.QueryRow(ctx, q, clientID).Scan(
		&c.ClientID, &c.ClientSecretHash,
		&c.RedirectURIs, &c.GrantTypes, &c.Scopes,
		&c.TokenEndpointAuthMethod, &c.Disabled, &c.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrClientNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("oauth2: lookup client: %w", err)
	}
	if c.Disabled {
		return nil, ErrClientNotFound
	}
	return &c, nil
}

// Authenticate verifies a presented (client_id, client_secret)
// pair. Public clients (no stored secret) MUST authenticate with
// an empty secret — any non-empty secret on a public client row
// is treated as a misconfigured request and rejected.
//
// Returns the loaded *Client on success.
func (s *ClientStore) Authenticate(ctx context.Context, clientID, clientSecret string) (*Client, error) {
	c, err := s.LookupByID(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if c.IsPublic() {
		if clientSecret != "" {
			return nil, ErrClientAuthFailed
		}
		return c, nil
	}
	ok, err := password.Verify(clientSecret, c.ClientSecretHash)
	if err != nil {
		return nil, fmt.Errorf("oauth2: verify secret: %w", err)
	}
	if !ok {
		return nil, ErrClientAuthFailed
	}
	return c, nil
}

// CreateForTest inserts a row directly. Production deploys
// register clients via a Connect RPC backed by a dynamic-client-
// registration handler (TASK-P1-IAM-DCR-001, future); until then
// tests + ops scripts use this helper.
func (s *ClientStore) CreateForTest(ctx context.Context, c Client) error {
	const q = `
INSERT INTO oauth2_clients
  (client_id, client_secret_hash, redirect_uris, grant_types, scopes,
   token_endpoint_auth_method, disabled)
VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6, $7)
`
	_, err := s.pool.Exec(ctx, q,
		c.ClientID, c.ClientSecretHash,
		c.RedirectURIs, c.GrantTypes, c.Scopes,
		c.TokenEndpointAuthMethod, c.Disabled,
	)
	if err != nil {
		return fmt.Errorf("oauth2: insert client: %w", err)
	}
	return nil
}

// HashClientSecret returns the PHC-encoded argon2id hash of the
// given secret. Used by the dynamic-client-registration path AND
// by tests' CreateForTest helper.
func HashClientSecret(secret string) (string, error) {
	if secret == "" {
		return "", errors.New("oauth2: empty secret")
	}
	return password.Hash(secret, password.PolicyV1)
}

// ----------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------

func isAbsoluteHTTPS(u string) bool {
	parsed, err := url.Parse(u)
	if err != nil {
		return false
	}
	return parsed.Scheme == "https" && parsed.Host != ""
}

func isLoopback(u string) bool {
	parsed, err := url.Parse(u)
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" {
		return false
	}
	host := parsed.Hostname()
	return host == "127.0.0.1" || host == "localhost" || host == "::1"
}

// ConstantTimeCompareStrings is a convenience wrapper used by the
// token endpoint when comparing client-supplied state echoes.
func ConstantTimeCompareStrings(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// SplitScope splits a scope-string per RFC 6749 §3.3
// (space-delimited). Empty input yields an empty slice.
func SplitScope(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Fields(s)
	return parts
}

// JoinScope is the inverse of SplitScope.
func JoinScope(s []string) string {
	return strings.Join(s, " ")
}

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrClientNotFound is returned when no oauth2_clients row matches
// the supplied client_id (or the row is disabled).
var ErrClientNotFound = errors.New("oauth2: client not found")

// ErrClientAuthFailed is returned when the (client_id,
// client_secret) pair does not authenticate.
var ErrClientAuthFailed = errors.New("oauth2: client authentication failed")

// ErrInvalidRedirectURI is returned when the requested redirect
// URI is not absolute https (or a permitted loopback http).
var ErrInvalidRedirectURI = errors.New("oauth2: redirect_uri must be absolute https or loopback http")

// ErrRedirectURIMismatch is returned when the requested URI does
// not exactly match any registered URI.
var ErrRedirectURIMismatch = errors.New("oauth2: redirect_uri did not match any registered value")

// ErrRedirectURIRequired is returned when a redirect_uri was
// omitted and the client has more than one registered.
var ErrRedirectURIRequired = errors.New("oauth2: redirect_uri is required when multiple are registered")
