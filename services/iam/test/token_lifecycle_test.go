//go:build integration

// token_lifecycle_test.go — TASK-P1-IAM-002 acceptance.
//
// Exercises the full token lifecycle end-to-end against a real
// Postgres so the family-invalidation invariant (the most security-
// critical behaviour of the refresh-token system) is verified against
// the real concurrency semantics, not a mock.
//
// Coverage:
//
//	1. Login mints (access JWT, refresh) pair.
//	2. Access JWT verifies via the IAM JWKS endpoint.
//	3. Refresh rotation produces a fresh credential pair; old refresh
//	   becomes consumed.
//	4. Replaying the original refresh after rotation is detected as
//	   reuse and revokes the entire family — the legitimate "current"
//	   refresh is also invalidated.

package iam_test

import (
	"context"
	"crypto/rsa"
	"errors"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	authzv1 "p9e.in/chetana/packages/authz/v1"

	"github.com/ppusapati/space/services/iam/internal/token"
)

func newTokenPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("IAM_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("IAM_TEST_DATABASE_URL not set — skipping integration test")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`TRUNCATE refresh_tokens RESTART IDENTITY CASCADE`)
		pool.Close()
	})
	if _, err := pool.Exec(context.Background(),
		`TRUNCATE refresh_tokens RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return pool
}

func TestTokenLifecycle_LoginVerifyRotateReuse(t *testing.T) {
	pool := newTokenPool(t)

	// 1. Build IAM signing surface.
	priv, err := token.GenerateRSAKey(2048)
	if err != nil {
		t.Fatalf("rsa: %v", err)
	}
	kid := token.SHA256KID(&priv.PublicKey)
	ks := token.NewKeyStore(time.Now)
	if err := ks.Add(token.SigningKey{
		KeyID:      kid,
		Private:    priv,
		Activation: time.Now().Add(-time.Minute),
		Retirement: time.Now().Add(time.Hour),
	}); err != nil {
		t.Fatalf("add key: %v", err)
	}
	const issuerURL = "https://iam.test.chetana.p9e.in"
	iss, err := token.NewIssuer(ks, token.IssuerConfig{
		Issuer:         issuerURL,
		AccessTokenTTL: 15 * time.Minute,
	})
	if err != nil {
		t.Fatalf("issuer: %v", err)
	}
	rs := token.NewRefreshStore(pool, time.Now)
	li := token.NewLoginIssuer(iss, rs, time.Now)

	// 2. Stand up the JWKS endpoint and a verifier pointing at it.
	srv := httptest.NewServer(ks.JWKSHandler())
	defer srv.Close()

	v, err := authzv1.NewVerifier(context.Background(), authzv1.VerifierConfig{
		JWKSURL:        srv.URL,
		ExpectedIssuer: issuerURL,
	})
	if err != nil {
		t.Fatalf("verifier: %v", err)
	}

	ctx := context.Background()

	// 3. Login mints (access JWT, refresh).
	out, err := li.IssueLoginTokens(ctx, token.LoginIssueInput{
		UserID:    "11111111-1111-1111-1111-111111111111",
		TenantID:  "22222222-2222-2222-2222-222222222222",
		SessionID: "sess-lifecycle",
		AMR:       []string{"pwd"},
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if out.AccessToken == "" || out.RefreshToken == "" {
		t.Fatal("login output missing tokens")
	}

	// 4. Access JWT verifies through the JWKS-fetching verifier.
	p, err := v.VerifyAccessToken(ctx, out.AccessToken)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if p.SessionID != "sess-lifecycle" {
		t.Errorf("session_id: %q", p.SessionID)
	}

	// 5. Rotate refresh → fresh pair, old refresh consumed.
	newRefresh, err := rs.Rotate(ctx, out.RefreshToken)
	if err != nil {
		t.Fatalf("rotate: %v", err)
	}
	if newRefresh.Token == out.RefreshToken {
		t.Error("rotated refresh must differ from original")
	}

	// 6. Reuse: present the original refresh again. Must trip
	// detection AND revoke the entire family — the new refresh
	// must then be invalid as well.
	if _, err := rs.Rotate(ctx, out.RefreshToken); !errors.Is(err, token.ErrReusedRefresh) {
		t.Fatalf("reuse: got %v want ErrReusedRefresh", err)
	}
	if _, err := rs.Rotate(ctx, newRefresh.Token); !errors.Is(err, token.ErrInvalidRefresh) {
		t.Fatalf("post-revoke rotate: got %v want ErrInvalidRefresh", err)
	}

	// We do not assert on the pubkey itself but use it to ensure the
	// test compiles against rsa.PublicKey when imported.
	var _ *rsa.PublicKey = &priv.PublicKey
}
