package authzv1

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testIssuerURL = "https://iam.test.chetana.p9e.in"

// signTestToken builds an RS256 JWT compatible with the IAM-issued shape.
func signTestToken(t *testing.T, kid string, key *rsa.PrivateKey, mut func(*accessClaims)) string {
	t.Helper()
	now := time.Now().UTC()
	claims := accessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    testIssuerURL,
			Subject:   "user-1",
			Audience:  jwt.ClaimStrings{"chetana-api"},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			ID:        "jti-1",
		},
		TenantID:       "tenant-1",
		IsUSPerson:     true,
		ClearanceLevel: "cui",
		Nationality:    "US",
		Roles:          []string{"operator"},
		Scopes:         []string{"telemetry.read"},
		SessionID:      "sess-1",
		AMR:            []string{"pwd"},
	}
	if mut != nil {
		mut(&claims)
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tok.Header["kid"] = kid
	signed, err := tok.SignedString(key)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return signed
}

func newTestKey(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("gen: %v", err)
	}
	return priv, "kid-test-1"
}

func TestVerifyAccessToken_HappyPath(t *testing.T) {
	priv, kid := newTestKey(t)
	v := NewVerifierWithKeys(VerifierConfig{
		ExpectedIssuer:   testIssuerURL,
		ExpectedAudience: "chetana-api",
	}, map[string]*rsa.PublicKey{kid: &priv.PublicKey})

	signed := signTestToken(t, kid, priv, nil)
	p, err := v.VerifyAccessToken(context.Background(), signed)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if p.UserID != "user-1" || p.TenantID != "tenant-1" || p.SessionID != "sess-1" {
		t.Errorf("identity wrong: %+v", p)
	}
	if !p.IsUSPerson || p.ClearanceLevel != "cui" || p.Nationality != "US" {
		t.Errorf("attributes wrong: %+v", p)
	}
	if len(p.Roles) != 1 || p.Roles[0] != "operator" {
		t.Errorf("roles: %v", p.Roles)
	}
	if p.Issuer != testIssuerURL {
		t.Errorf("issuer: %q", p.Issuer)
	}
	if p.JTI != "jti-1" {
		t.Errorf("jti: %q", p.JTI)
	}
}

func TestVerifyAccessToken_InvalidSignature(t *testing.T) {
	priv, kid := newTestKey(t)
	other, _ := rsa.GenerateKey(rand.Reader, 2048)

	v := NewVerifierWithKeys(VerifierConfig{ExpectedIssuer: testIssuerURL},
		map[string]*rsa.PublicKey{kid: &other.PublicKey})

	signed := signTestToken(t, kid, priv, nil)
	if _, err := v.VerifyAccessToken(context.Background(), signed); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("got %v want ErrInvalidToken", err)
	}
}

func TestVerifyAccessToken_Expired(t *testing.T) {
	priv, kid := newTestKey(t)
	v := NewVerifierWithKeys(VerifierConfig{ExpectedIssuer: testIssuerURL},
		map[string]*rsa.PublicKey{kid: &priv.PublicKey})

	signed := signTestToken(t, kid, priv, func(c *accessClaims) {
		past := time.Now().Add(-time.Hour)
		c.IssuedAt = jwt.NewNumericDate(past.Add(-15 * time.Minute))
		c.NotBefore = jwt.NewNumericDate(past.Add(-15 * time.Minute))
		c.ExpiresAt = jwt.NewNumericDate(past)
	})
	if _, err := v.VerifyAccessToken(context.Background(), signed); !errors.Is(err, ErrTokenExpired) {
		t.Errorf("got %v want ErrTokenExpired", err)
	}
}

func TestVerifyAccessToken_NotYetValid(t *testing.T) {
	priv, kid := newTestKey(t)
	v := NewVerifierWithKeys(VerifierConfig{ExpectedIssuer: testIssuerURL},
		map[string]*rsa.PublicKey{kid: &priv.PublicKey})

	signed := signTestToken(t, kid, priv, func(c *accessClaims) {
		future := time.Now().Add(2 * time.Hour)
		c.NotBefore = jwt.NewNumericDate(future)
		c.IssuedAt = jwt.NewNumericDate(future)
		c.ExpiresAt = jwt.NewNumericDate(future.Add(15 * time.Minute))
	})
	if _, err := v.VerifyAccessToken(context.Background(), signed); !errors.Is(err, ErrTokenNotYetValid) {
		t.Errorf("got %v want ErrTokenNotYetValid", err)
	}
}

func TestVerifyAccessToken_IssuerMismatch(t *testing.T) {
	priv, kid := newTestKey(t)
	v := NewVerifierWithKeys(VerifierConfig{ExpectedIssuer: "https://other"},
		map[string]*rsa.PublicKey{kid: &priv.PublicKey})

	signed := signTestToken(t, kid, priv, nil)
	if _, err := v.VerifyAccessToken(context.Background(), signed); !errors.Is(err, ErrIssuerMismatch) {
		t.Errorf("got %v want ErrIssuerMismatch", err)
	}
}

func TestVerifyAccessToken_AudienceMismatch(t *testing.T) {
	priv, kid := newTestKey(t)
	v := NewVerifierWithKeys(VerifierConfig{
		ExpectedIssuer:   testIssuerURL,
		ExpectedAudience: "wrong-audience",
	}, map[string]*rsa.PublicKey{kid: &priv.PublicKey})

	signed := signTestToken(t, kid, priv, nil)
	if _, err := v.VerifyAccessToken(context.Background(), signed); !errors.Is(err, ErrAudienceMismatch) {
		t.Errorf("got %v want ErrAudienceMismatch", err)
	}
}

func TestVerifyAccessToken_UnknownKidWithoutJWKSURL(t *testing.T) {
	priv, _ := newTestKey(t)
	other, _ := rsa.GenerateKey(rand.Reader, 2048)

	v := NewVerifierWithKeys(VerifierConfig{ExpectedIssuer: testIssuerURL},
		map[string]*rsa.PublicKey{"only-known-kid": &other.PublicKey})

	signed := signTestToken(t, "unknown-kid", priv, nil)
	if _, err := v.VerifyAccessToken(context.Background(), signed); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("got %v want ErrInvalidToken", err)
	}
}

func TestVerifyAccessToken_RotationOverlap_PicksUpNewKid(t *testing.T) {
	// Server rotates: first only key1 in JWKS, then both, then only key2.
	priv1, _ := rsa.GenerateKey(rand.Reader, 2048)
	priv2, _ := rsa.GenerateKey(rand.Reader, 2048)

	stage := atomic.Int32{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/jwk-set+json")
		var keys []map[string]string
		if stage.Load() == 0 {
			keys = []map[string]string{jwkPayload("kid-1", &priv1.PublicKey)}
		} else {
			keys = []map[string]string{
				jwkPayload("kid-1", &priv1.PublicKey),
				jwkPayload("kid-2", &priv2.PublicKey),
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": keys})
	}))
	defer srv.Close()

	v, err := NewVerifier(context.Background(), VerifierConfig{
		JWKSURL:        srv.URL,
		ExpectedIssuer: testIssuerURL,
	})
	if err != nil {
		t.Fatalf("new verifier: %v", err)
	}

	// Server rolls forward — kid-2 now appears in JWKS.
	stage.Store(1)

	signed := signTestToken(t, "kid-2", priv2, nil)
	if _, err := v.VerifyAccessToken(context.Background(), signed); err != nil {
		t.Fatalf("verify after rotation: %v", err)
	}
}

func TestNewVerifier_FailsWithoutURL(t *testing.T) {
	if _, err := NewVerifier(context.Background(), VerifierConfig{}); err == nil {
		t.Error("missing URL should error")
	}
}

func TestNewVerifier_FailsOnUnreachableJWKS(t *testing.T) {
	if _, err := NewVerifier(context.Background(), VerifierConfig{
		JWKSURL: "http://127.0.0.1:1/never-listening",
	}); err == nil {
		t.Error("unreachable JWKS should error at boot")
	}
}

func TestParseJWKS_RoundTrip(t *testing.T) {
	priv, _ := newTestKey(t)
	body, err := json.Marshal(map[string]any{
		"keys": []map[string]string{jwkPayload("kid-x", &priv.PublicKey)},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	keys, err := parseJWKS(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	got, ok := keys["kid-x"]
	if !ok {
		t.Fatal("kid-x missing")
	}
	if got.N.Cmp(priv.PublicKey.N) != 0 {
		t.Error("modulus mismatch")
	}
	if got.E != priv.PublicKey.E {
		t.Errorf("exponent: got %d want %d", got.E, priv.PublicKey.E)
	}
}

// jwkPayload builds a JWK JSON object compatible with parseJWKS.
func jwkPayload(kid string, pub *rsa.PublicKey) map[string]string {
	eBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(eBuf, uint32(pub.E))
	for len(eBuf) > 1 && eBuf[0] == 0 {
		eBuf = eBuf[1:]
	}
	return map[string]string{
		"kty": "RSA",
		"use": "sig",
		"alg": "RS256",
		"kid": kid,
		"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(eBuf),
	}
}

// keep big.Int+fmt referenced even if not directly used in assertions.
var _ = big.NewInt
var _ = fmt.Sprintf
