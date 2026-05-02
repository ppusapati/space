// Package token implements JWT issuance + verification + refresh-token
// rotation for the chetana IAM service.
//
// → REQ-FUNC-PLT-IAM-002 (token issuance + rotation)
// → REQ-FUNC-PLT-IAM-008 (claim set: tenant_id, is_us_person, …)
// → REQ-NFR-SEC-001       (FIPS-validated signer)
// → design.md §4.1.1
//
// The package owns three concerns:
//   • jwt.go     — Issuer + Claims + signing
//   • jwks.go    — KeyStore + /.well-known/jwks.json + rotation
//   • refresh.go — single-use refresh tokens + family invalidation
package token

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Default lifetimes per REQ-FUNC-PLT-IAM-002.
const (
	DefaultAccessTokenTTL  = 15 * time.Minute
	DefaultRefreshTokenTTL = 7 * 24 * time.Hour
)

// SigningAlgorithm is the JWS algorithm chetana uses for access
// tokens. RS256 satisfies REQ-NFR-SEC-001 when the underlying RSA
// key generation runs through the FIPS-validated provider (the
// crypto/fips guard in services/packages/crypto/ asserts this at
// boot).
const SigningAlgorithm = "RS256"

// Claims is the canonical set of JWT claims produced by the IAM
// service. The shape mirrors design.md §4.1.1 verbatim — every
// field has a downstream consumer (ABAC decision, audit chain,
// realtime-gw topic auth).
type Claims struct {
	jwt.RegisteredClaims

	// Application-defined claims. Names match the design exactly
	// so token introspection tools see the canonical shape.
	TenantID       string   `json:"tenant_id"`
	IsUSPerson     bool     `json:"is_us_person"`
	ClearanceLevel string   `json:"clearance_level"` // "internal" | "restricted" | "cui" | "itar"
	Nationality    string   `json:"nationality"`     // ISO-3166-alpha-2
	Roles          []string `json:"roles"`
	Scopes         []string `json:"scopes"`
	SessionID      string   `json:"session_id"`
	AMR            []string `json:"amr"`             // RFC 8176 — Authentication Methods References
}

// Principal is the verified identity downstream services consume.
// Returned by the verifier in services/packages/authz; constructed
// from a Claims after successful signature + expiry check.
//
// Kept in this package (not the authz package) so tests in both
// packages can reuse the same struct shape.
type Principal struct {
	UserID         string
	TenantID       string
	SessionID      string
	IsUSPerson     bool
	ClearanceLevel string
	Nationality    string
	Roles          []string
	Scopes         []string
	AMR            []string
	IssuedAt       time.Time
	ExpiresAt      time.Time
	JTI            string
}

// AsPrincipal projects Claims into a Principal. Used by the
// verifier in the authz package; lives in this package because
// the projection is one direction (token → principal) and we want
// to keep the JWT-specific bits encapsulated here.
func (c *Claims) AsPrincipal() *Principal {
	p := &Principal{
		UserID:         c.Subject,
		TenantID:       c.TenantID,
		SessionID:      c.SessionID,
		IsUSPerson:     c.IsUSPerson,
		ClearanceLevel: c.ClearanceLevel,
		Nationality:    c.Nationality,
		Roles:          append([]string(nil), c.Roles...),
		Scopes:         append([]string(nil), c.Scopes...),
		AMR:            append([]string(nil), c.AMR...),
		JTI:            c.ID,
	}
	if c.IssuedAt != nil {
		p.IssuedAt = c.IssuedAt.Time
	}
	if c.ExpiresAt != nil {
		p.ExpiresAt = c.ExpiresAt.Time
	}
	return p
}

// IssueInput is the call shape for Issuer.IssueAccessToken. The
// caller (login handler, refresh handler) fills in the verified
// claims; the issuer adds iss/iat/exp/jti and signs.
type IssueInput struct {
	UserID         string
	TenantID       string
	SessionID      string
	IsUSPerson     bool
	ClearanceLevel string
	Nationality    string
	Roles          []string
	Scopes         []string
	Audience       []string
	AMR            []string
}

// Issuer signs JWTs for the IAM service. Construct with
// NewIssuer; the type holds an active KeyStore (one signing key,
// optionally a next-key for the rotation overlap window).
type Issuer struct {
	store    *KeyStore
	issuer   string
	clock    func() time.Time
	tokenTTL time.Duration
}

// IssuerConfig configures the JWT issuer.
type IssuerConfig struct {
	// Issuer is the canonical iss claim — typically
	// "https://iam.chetana.<region>.p9e.in".
	Issuer string

	// AccessTokenTTL overrides DefaultAccessTokenTTL. Must be > 0.
	AccessTokenTTL time.Duration

	// Clock injects a time source for tests. nil → time.Now.
	Clock func() time.Time
}

// NewIssuer builds a token issuer over the supplied KeyStore. The
// store MUST contain at least one signing key; new tokens are
// signed with whichever key the store reports as Active.
func NewIssuer(store *KeyStore, cfg IssuerConfig) (*Issuer, error) {
	if store == nil {
		return nil, errors.New("token: nil key store")
	}
	if cfg.Issuer == "" {
		return nil, errors.New("token: empty issuer")
	}
	if cfg.AccessTokenTTL <= 0 {
		cfg.AccessTokenTTL = DefaultAccessTokenTTL
	}
	if cfg.Clock == nil {
		cfg.Clock = time.Now
	}
	if _, err := store.Active(); err != nil {
		return nil, fmt.Errorf("token: key store has no active key: %w", err)
	}
	return &Issuer{
		store:    store,
		issuer:   cfg.Issuer,
		clock:    cfg.Clock,
		tokenTTL: cfg.AccessTokenTTL,
	}, nil
}

// IssueAccessToken signs a fresh access token and returns the
// compact-serialised string + the canonical Claims it carries.
// Returning the Claims spares the caller a re-parse round-trip.
//
// The jti is a 128-bit random hex token; iat = now; exp = now + TTL.
func (i *Issuer) IssueAccessToken(in IssueInput) (string, *Claims, error) {
	if in.UserID == "" {
		return "", nil, errors.New("token: empty UserID")
	}
	if in.TenantID == "" {
		return "", nil, errors.New("token: empty TenantID")
	}
	if in.SessionID == "" {
		return "", nil, errors.New("token: empty SessionID")
	}

	now := i.clock().UTC()
	jti, err := newRandomHex(16)
	if err != nil {
		return "", nil, fmt.Errorf("token: jti: %w", err)
	}
	audience := jwt.ClaimStrings(in.Audience)
	if len(audience) == 0 {
		audience = jwt.ClaimStrings{"chetana-api"}
	}

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    i.issuer,
			Subject:   in.UserID,
			Audience:  audience,
			ExpiresAt: jwt.NewNumericDate(now.Add(i.tokenTTL)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        jti,
		},
		TenantID:       in.TenantID,
		IsUSPerson:     in.IsUSPerson,
		ClearanceLevel: in.ClearanceLevel,
		Nationality:    in.Nationality,
		Roles:          append([]string(nil), in.Roles...),
		Scopes:         append([]string(nil), in.Scopes...),
		SessionID:      in.SessionID,
		AMR:            append([]string(nil), in.AMR...),
	}

	signingKey, err := i.store.Active()
	if err != nil {
		return "", nil, fmt.Errorf("token: active key: %w", err)
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tok.Header["kid"] = signingKey.KeyID
	signed, err := tok.SignedString(signingKey.Private)
	if err != nil {
		return "", nil, fmt.Errorf("token: sign: %w", err)
	}
	return signed, claims, nil
}

// ParseUnverified decodes a token without verifying its signature.
// Used in narrow paths where the caller has another assurance the
// payload is trusted (e.g. logging the jti of an already-verified
// token). NEVER use this for authorization decisions.
func ParseUnverified(raw string) (*Claims, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	var c Claims
	if _, _, err := parser.ParseUnverified(raw, &c); err != nil {
		return nil, fmt.Errorf("token: parse unverified: %w", err)
	}
	return &c, nil
}

// PublicKeyForKID returns the RSA public key matching the given key
// ID, or an error when the key is unknown. Used by the verifier in
// services/packages/authz; lives here so the key-store integration
// stays encapsulated.
func (s *KeyStore) PublicKeyForKID(kid string) (*rsa.PublicKey, error) {
	for _, k := range s.snapshot() {
		if k.KeyID == kid {
			return &k.Private.PublicKey, nil
		}
	}
	return nil, fmt.Errorf("token: %w: kid=%s", ErrUnknownKey, kid)
}

// newRandomHex returns 2*n hex characters (n random bytes).
func newRandomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
