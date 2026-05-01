// Package authz verifies HS256 / RS256 JWTs and exposes a small
// role-based authorisation helper.
//
// The package deliberately does not depend on the iam service — both
// iam (which mints tokens) and every protected service (which verifies
// them) link this package. Verifiers are constructed once at startup
// from a key set and shared across handlers.
package authz

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrUnauthenticated is returned when the bearer token is missing,
// malformed, expired, or fails signature verification.
var ErrUnauthenticated = errors.New("unauthenticated")

// ErrPermissionDenied is returned when the token is valid but does not
// carry the required role.
var ErrPermissionDenied = errors.New("permission denied")

// Claims is the canonical JWT body issued by iam. The `Roles` slice is
// the ground truth for RBAC; the [HasRole] helper checks it.
type Claims struct {
	jwt.RegisteredClaims
	UserID   string   `json:"user_id"`
	TenantID string   `json:"tenant_id"`
	Roles    []string `json:"roles"`
}

// HasRole returns true if the claims include `want`.
func (c *Claims) HasRole(want string) bool {
	return slices.Contains(c.Roles, want)
}

// Verifier validates JWTs and extracts [Claims].
type Verifier struct {
	// SigningKey is HMAC secret bytes (HS256) or PEM-encoded RSA public
	// key bytes (RS256).
	signingKey []byte
	// Method is "HS256" or "RS256".
	method jwt.SigningMethod
	// ExpectedIssuer is enforced if non-empty.
	expectedIssuer string
	// ExpectedAudience is enforced if non-empty.
	expectedAudience string
	// ClockSkew loosens iat / exp checks.
	clockSkew time.Duration

	// rsaPub is the parsed RSA public key when method == RS256.
	rsaKey any
}

// VerifierOptions configures NewVerifier.
type VerifierOptions struct {
	// Method is one of "HS256" or "RS256".
	Method string
	// SigningKey is the HS256 secret (raw bytes) or the RS256 public
	// key (PEM-encoded).
	SigningKey []byte
	// ExpectedIssuer, when non-empty, is enforced.
	ExpectedIssuer string
	// ExpectedAudience, when non-empty, is enforced.
	ExpectedAudience string
	// ClockSkew is added to the now-window when checking iat/exp.
	ClockSkew time.Duration
}

// NewVerifier constructs a Verifier.
func NewVerifier(opts VerifierOptions) (*Verifier, error) {
	v := &Verifier{
		signingKey:       opts.SigningKey,
		expectedIssuer:   opts.ExpectedIssuer,
		expectedAudience: opts.ExpectedAudience,
		clockSkew:        opts.ClockSkew,
	}
	switch opts.Method {
	case "HS256":
		v.method = jwt.SigningMethodHS256
	case "RS256":
		v.method = jwt.SigningMethodRS256
		key, err := jwt.ParseRSAPublicKeyFromPEM(opts.SigningKey)
		if err != nil {
			return nil, fmt.Errorf("authz: parse RSA pub key: %w", err)
		}
		v.rsaKey = key
	default:
		return nil, fmt.Errorf("authz: unsupported method %q", opts.Method)
	}
	return v, nil
}

// VerifyHeader extracts the bearer token from `header` and returns the
// validated claims.
func (v *Verifier) VerifyHeader(_ context.Context, header http.Header) (*Claims, error) {
	auth := header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return nil, fmt.Errorf("%w: missing Bearer", ErrUnauthenticated)
	}
	return v.VerifyToken(strings.TrimPrefix(auth, "Bearer "))
}

// VerifyToken parses and verifies a token string.
func (v *Verifier) VerifyToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	parser := jwt.NewParser(jwt.WithLeeway(v.clockSkew))
	_, err := parser.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != v.method.Alg() {
			return nil, fmt.Errorf("unexpected signing method %q", t.Method.Alg())
		}
		switch v.method.Alg() {
		case "HS256":
			return v.signingKey, nil
		case "RS256":
			return v.rsaKey, nil
		default:
			return nil, errors.New("unsupported method")
		}
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnauthenticated, err)
	}
	if v.expectedIssuer != "" && claims.Issuer != v.expectedIssuer {
		return nil, fmt.Errorf("%w: issuer %q != %q", ErrUnauthenticated, claims.Issuer, v.expectedIssuer)
	}
	if v.expectedAudience != "" && !slices.Contains(claims.Audience, v.expectedAudience) {
		return nil, fmt.Errorf("%w: audience %v missing %q", ErrUnauthenticated, claims.Audience, v.expectedAudience)
	}
	return claims, nil
}

// Require returns a function that asserts the supplied claims hold the
// given role; otherwise it returns ErrPermissionDenied.
func Require(role string) func(*Claims) error {
	return func(c *Claims) error {
		if c == nil || !c.HasRole(role) {
			return fmt.Errorf("%w: requires role %q", ErrPermissionDenied, role)
		}
		return nil
	}
}
