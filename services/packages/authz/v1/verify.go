// Package authzv1 is the v1 IAM-aligned JWT verifier consumed by every
// chetana service's auth interceptor.
//
// → REQ-FUNC-PLT-IAM-002 acceptance #4: services/packages/authz/v1
//   exposes VerifyAccessToken(ctx, token) returning the populated
//   principal struct.
// → REQ-CONST-011 (no duplication: every service interceptor calls
//                  this single function).
// → design.md §4.1.1 token model + §4.1.2 ABAC decision.
//
// The verifier is decoupled from the IAM service — it pulls the
// public-key set from a JWKS endpoint and validates RS256 signatures
// against it. Downstream services instantiate one Verifier at boot,
// pointed at IAM's /.well-known/jwks.json, and call VerifyAccessToken
// on every authenticated request.
//
// Sibling-package note: the legacy authz package (parent dir) holds
// CustomClaims + interceptor scaffolding from the previous platform.
// authz/v1 is the package new chetana services should import; it
// avoids the legacy package's protobuf init dependency so test
// binaries don't trip the proto-init panic.

package authzv1

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Principal is the verified identity downstream services consume.
// Populated by VerifyAccessToken from the JWT claims.
//
// Distinct from the legacy InjectedUserInfo (which still serves the
// gRPC interceptor in this package) because the v1 IAM token shape
// is different — see design.md §4.1.1 vs the existing
// CustomClaims.
type Principal struct {
	UserID         string
	TenantID       string
	SessionID      string
	IsUSPerson     bool
	ClearanceLevel string   // "internal" | "restricted" | "cui" | "itar"
	Nationality    string   // ISO-3166-alpha-2
	Roles          []string // RBAC roles
	Scopes         []string // OAuth2 scopes
	AMR            []string // RFC 8176 authentication methods references
	IssuedAt       time.Time
	ExpiresAt      time.Time
	JTI            string
	Audience       []string
	Issuer         string
}

// VerifierConfig configures the verifier.
type VerifierConfig struct {
	// JWKSURL is the URL of the IAM service's JWKS endpoint
	// (typically https://iam.<region>.chetana.p9e.in/.well-known/jwks.json).
	JWKSURL string

	// JWKSRefreshInterval is how often the verifier re-fetches the
	// JWKS in the background. Defaults to 1h. The IAM service
	// publishes new keys 24h before they become active so a 1h
	// refresh keeps the cache safely fresh.
	JWKSRefreshInterval time.Duration

	// HTTPClient overrides the default http client used to fetch
	// the JWKS. Pass an mTLS-equipped client when fetching from
	// inside the cluster.
	HTTPClient *http.Client

	// ExpectedIssuer is matched against the JWT iss claim. Empty
	// string disables the check (NOT recommended).
	ExpectedIssuer string

	// ExpectedAudience, when non-empty, is matched against the JWT
	// aud claim. The token is accepted when aud contains this
	// value.
	ExpectedAudience string

	// ClockSkew is the leeway applied to exp / nbf checks.
	// Defaults to 30s.
	ClockSkew time.Duration

	// Clock injects a time source for tests. nil → time.Now.
	Clock func() time.Time
}

// Verifier validates IAM-issued access tokens. Construct with
// NewVerifier; the type holds a cached JWKS that is refreshed
// periodically in the background.
type Verifier struct {
	cfg VerifierConfig

	mu       sync.RWMutex
	keys     map[string]*rsa.PublicKey // kid → key
	lastFetch time.Time
}

// NewVerifier builds a verifier and performs an initial JWKS fetch.
// Returns an error when the initial fetch fails so the caller can
// bail at boot rather than serve unauthenticated requests.
func NewVerifier(ctx context.Context, cfg VerifierConfig) (*Verifier, error) {
	if cfg.JWKSURL == "" {
		return nil, errors.New("authz: empty JWKSURL")
	}
	if cfg.JWKSRefreshInterval <= 0 {
		cfg.JWKSRefreshInterval = time.Hour
	}
	if cfg.ClockSkew <= 0 {
		cfg.ClockSkew = 30 * time.Second
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	}
	if cfg.Clock == nil {
		cfg.Clock = time.Now
	}
	v := &Verifier{cfg: cfg, keys: map[string]*rsa.PublicKey{}}
	if err := v.refresh(ctx); err != nil {
		return nil, fmt.Errorf("authz: initial JWKS fetch: %w", err)
	}
	return v, nil
}

// NewVerifierWithKeys is the test-friendly constructor that bypasses
// HTTP fetching. Pass a pre-populated kid → public-key map.
func NewVerifierWithKeys(cfg VerifierConfig, keys map[string]*rsa.PublicKey) *Verifier {
	if cfg.ClockSkew <= 0 {
		cfg.ClockSkew = 30 * time.Second
	}
	if cfg.Clock == nil {
		cfg.Clock = time.Now
	}
	cp := make(map[string]*rsa.PublicKey, len(keys))
	for k, v := range keys {
		cp[k] = v
	}
	return &Verifier{cfg: cfg, keys: cp}
}

// VerifyAccessToken parses + verifies a bearer token and returns
// the populated Principal. Callers that need the raw claims can
// re-parse with token.ParseUnverified — but for authorization
// decisions, use Principal.
//
// Error semantics:
//   - ErrInvalidToken      — bad signature, unknown kid, malformed.
//   - ErrTokenExpired      — exp < now (with clock skew applied).
//   - ErrTokenNotYetValid  — nbf > now.
//   - ErrIssuerMismatch    — iss != ExpectedIssuer.
//   - ErrAudienceMismatch  — aud does not contain ExpectedAudience.
func (v *Verifier) VerifyAccessToken(ctx context.Context, raw string) (*Principal, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
		jwt.WithLeeway(v.cfg.ClockSkew),
		jwt.WithTimeFunc(v.cfg.Clock),
		jwt.WithExpirationRequired(),
	)

	var claims accessClaims
	tok, err := parser.ParseWithClaims(raw, &claims, v.keyFor(ctx))
	if err != nil {
		// The jwt-go errors aren't always great for our consumer;
		// translate the common ones.
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, ErrTokenExpired
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, ErrTokenNotYetValid
		case errors.Is(err, jwt.ErrSignatureInvalid):
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if !tok.Valid {
		return nil, ErrInvalidToken
	}

	if v.cfg.ExpectedIssuer != "" && claims.Issuer != v.cfg.ExpectedIssuer {
		return nil, fmt.Errorf("%w: got %q want %q", ErrIssuerMismatch, claims.Issuer, v.cfg.ExpectedIssuer)
	}
	if v.cfg.ExpectedAudience != "" {
		ok := false
		for _, a := range claims.Audience {
			if a == v.cfg.ExpectedAudience {
				ok = true
				break
			}
		}
		if !ok {
			return nil, fmt.Errorf("%w: got %v want %q", ErrAudienceMismatch, claims.Audience, v.cfg.ExpectedAudience)
		}
	}

	p := &Principal{
		UserID:         claims.Subject,
		TenantID:       claims.TenantID,
		SessionID:      claims.SessionID,
		IsUSPerson:     claims.IsUSPerson,
		ClearanceLevel: claims.ClearanceLevel,
		Nationality:    claims.Nationality,
		Roles:          append([]string(nil), claims.Roles...),
		Scopes:         append([]string(nil), claims.Scopes...),
		AMR:            append([]string(nil), claims.AMR...),
		Audience:       append([]string(nil), claims.Audience...),
		Issuer:         claims.Issuer,
		JTI:            claims.ID,
	}
	if claims.IssuedAt != nil {
		p.IssuedAt = claims.IssuedAt.Time
	}
	if claims.ExpiresAt != nil {
		p.ExpiresAt = claims.ExpiresAt.Time
	}
	return p, nil
}

// Refresh fetches the JWKS and updates the in-memory cache. Call
// periodically (the constructor does not start a background
// refresher — services that want one wrap this in a ticker, see
// example below).
//
//	go func() {
//	    t := time.NewTicker(v.RefreshInterval())
//	    for { <-t.C; _ = v.Refresh(ctx) }
//	}()
func (v *Verifier) Refresh(ctx context.Context) error { return v.refresh(ctx) }

// RefreshInterval is the configured refresh interval.
func (v *Verifier) RefreshInterval() time.Duration { return v.cfg.JWKSRefreshInterval }

// keyFor returns a jwt-go KeyFunc that maps the token's kid header
// to the cached public key. Refreshes the JWKS if the kid is
// unknown — gives an honest service that just rotated its key one
// shot at picking up the new key without waiting for the next
// scheduled refresh.
func (v *Verifier) keyFor(ctx context.Context) jwt.Keyfunc {
	return func(tok *jwt.Token) (any, error) {
		kid, ok := tok.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, ErrInvalidToken
		}
		v.mu.RLock()
		key, found := v.keys[kid]
		v.mu.RUnlock()
		if found {
			return key, nil
		}
		// Cache miss — try a refresh, then look again.
		if err := v.refresh(ctx); err != nil {
			return nil, fmt.Errorf("%w: kid %q not found and refresh failed: %v", ErrInvalidToken, kid, err)
		}
		v.mu.RLock()
		key, found = v.keys[kid]
		v.mu.RUnlock()
		if !found {
			return nil, fmt.Errorf("%w: kid %q not found after refresh", ErrInvalidToken, kid)
		}
		return key, nil
	}
}

func (v *Verifier) refresh(ctx context.Context) error {
	if v.cfg.JWKSURL == "" {
		// Test-injected key set; nothing to refresh.
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.cfg.JWKSURL, nil)
	if err != nil {
		return fmt.Errorf("authz: build req: %w", err)
	}
	req.Header.Set("Accept", "application/jwk-set+json")

	resp, err := v.cfg.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("authz: http: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authz: jwks http %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("authz: read jwks: %w", err)
	}
	keys, err := parseJWKS(body)
	if err != nil {
		return fmt.Errorf("authz: parse jwks: %w", err)
	}

	v.mu.Lock()
	v.keys = keys
	v.lastFetch = v.cfg.Clock()
	v.mu.Unlock()
	return nil
}

// parseJWKS decodes a JWKS payload into kid → public-key map.
// Exported via NewVerifierWithKeys's payload-equivalent path; kept
// unexported here because the public surface returns *rsa.PublicKey
// directly.
func parseJWKS(body []byte) (map[string]*rsa.PublicKey, error) {
	var set struct {
		Keys []struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			Alg string `json:"alg"`
			Use string `json:"use"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(body, &set); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	out := make(map[string]*rsa.PublicKey, len(set.Keys))
	for _, k := range set.Keys {
		if k.Kty != "RSA" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			return nil, fmt.Errorf("decode n: %w", err)
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			return nil, fmt.Errorf("decode e: %w", err)
		}
		var e int
		for _, b := range eBytes {
			e = e<<8 | int(b)
		}
		pub := &rsa.PublicKey{
			N: new(big.Int).SetBytes(nBytes),
			E: e,
		}
		out[k.Kid] = pub
	}
	return out, nil
}

// accessClaims mirrors the IAM-issued JWT shape. Defined here (not
// imported from services/iam) so the authz package has zero
// dependencies on any service module — services/packages MUST stay
// importable from every service.
type accessClaims struct {
	jwt.RegisteredClaims
	TenantID       string   `json:"tenant_id"`
	IsUSPerson     bool     `json:"is_us_person"`
	ClearanceLevel string   `json:"clearance_level"`
	Nationality    string   `json:"nationality"`
	Roles          []string `json:"roles"`
	Scopes         []string `json:"scopes"`
	SessionID      string   `json:"session_id"`
	AMR            []string `json:"amr"`
}

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrInvalidToken is returned when the JWT signature, encoding, or
// kid lookup fails.
var ErrInvalidToken = errors.New("authz: invalid access token")

// ErrTokenExpired is returned when exp < now.
var ErrTokenExpired = errors.New("authz: access token expired")

// ErrTokenNotYetValid is returned when nbf > now.
var ErrTokenNotYetValid = errors.New("authz: access token not yet valid")

// ErrIssuerMismatch is returned when iss != ExpectedIssuer.
var ErrIssuerMismatch = errors.New("authz: token issuer mismatch")

// ErrAudienceMismatch is returned when aud does not contain
// ExpectedAudience.
var ErrAudienceMismatch = errors.New("authz: token audience mismatch")
