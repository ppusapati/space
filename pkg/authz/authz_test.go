package authz

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func mintHS256(t *testing.T, secret []byte, claims Claims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	s, err := tok.SignedString(secret)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return s
}

func TestHS256RoundTrip(t *testing.T) {
	secret := []byte("topsecret")
	v, err := NewVerifier(VerifierOptions{
		Method:           "HS256",
		SigningKey:       secret,
		ExpectedIssuer:   "iam",
		ExpectedAudience: "space",
	})
	if err != nil {
		t.Fatalf("verifier: %v", err)
	}
	tok := mintHS256(t, secret, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "iam",
			Audience:  jwt.ClaimStrings{"space"},
			Subject:   "u-1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:   "u-1",
		TenantID: "t-1",
		Roles:    []string{"viewer"},
	})
	c, err := v.VerifyToken(tok)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if c.UserID != "u-1" || c.TenantID != "t-1" || !c.HasRole("viewer") {
		t.Fatalf("claims wrong: %+v", c)
	}
}

func TestVerifyHeaderRequiresBearer(t *testing.T) {
	v, _ := NewVerifier(VerifierOptions{Method: "HS256", SigningKey: []byte("x")})
	header := http.Header{}
	header.Set("Authorization", "Basic foo")
	if _, err := v.VerifyHeader(context.Background(), header); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("want ErrUnauthenticated, got %v", err)
	}
}

func TestRejectsWrongIssuer(t *testing.T) {
	secret := []byte("topsecret")
	v, _ := NewVerifier(VerifierOptions{Method: "HS256", SigningKey: secret, ExpectedIssuer: "iam"})
	tok := mintHS256(t, secret, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "spoof",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	if _, err := v.VerifyToken(tok); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("want ErrUnauthenticated, got %v", err)
	}
}

func TestRequireRole(t *testing.T) {
	check := Require("admin")
	if err := check(&Claims{Roles: []string{"viewer"}}); !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("want ErrPermissionDenied, got %v", err)
	}
	if err := check(&Claims{Roles: []string{"admin"}}); err != nil {
		t.Fatalf("admin should pass: %v", err)
	}
	if err := check(nil); !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("nil claims should be denied, got %v", err)
	}
}

func TestRejectsExpired(t *testing.T) {
	secret := []byte("topsecret")
	v, _ := NewVerifier(VerifierOptions{Method: "HS256", SigningKey: secret})
	tok := mintHS256(t, secret, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	})
	if _, err := v.VerifyToken(tok); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("expired token should be rejected, got %v", err)
	}
}
