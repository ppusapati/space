package token

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func newTestStore(t *testing.T, activation time.Time) (*KeyStore, SigningKey) {
	t.Helper()
	priv, err := GenerateRSAKey(2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	key := SigningKey{
		KeyID:      "kid-test-1",
		Private:    priv,
		Activation: activation,
		Retirement: activation.Add(48 * time.Hour),
	}
	store := NewKeyStore(func() time.Time { return activation })
	if err := store.Add(key); err != nil {
		t.Fatalf("add key: %v", err)
	}
	return store, key
}

func newTestIssuer(t *testing.T) (*Issuer, *KeyStore, time.Time) {
	t.Helper()
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	store, _ := newTestStore(t, now)
	iss, err := NewIssuer(store, IssuerConfig{
		Issuer:         "https://iam.test.chetana.p9e.in",
		AccessTokenTTL: 15 * time.Minute,
		Clock:          func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("new issuer: %v", err)
	}
	return iss, store, now
}

func TestIssueAccessToken_HappyPath(t *testing.T) {
	iss, _, now := newTestIssuer(t)

	signed, claims, err := iss.IssueAccessToken(IssueInput{
		UserID:         "user-1",
		TenantID:       "tenant-1",
		SessionID:      "sess-1",
		IsUSPerson:     true,
		ClearanceLevel: "cui",
		Nationality:    "US",
		Roles:          []string{"operator"},
		Scopes:         []string{"telemetry.read"},
		Audience:       []string{"chetana-api"},
		AMR:            []string{"pwd", "mfa"},
	})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if signed == "" {
		t.Fatal("empty signed token")
	}
	if strings.Count(signed, ".") != 2 {
		t.Fatalf("expected 3-segment JWT, got %q", signed)
	}

	if claims.Subject != "user-1" {
		t.Errorf("subject: got %q want %q", claims.Subject, "user-1")
	}
	if claims.TenantID != "tenant-1" {
		t.Errorf("tenant_id: got %q want %q", claims.TenantID, "tenant-1")
	}
	if claims.SessionID != "sess-1" {
		t.Errorf("session_id: got %q want %q", claims.SessionID, "sess-1")
	}
	if !claims.IsUSPerson {
		t.Error("is_us_person: want true")
	}
	if claims.ClearanceLevel != "cui" {
		t.Errorf("clearance: got %q", claims.ClearanceLevel)
	}
	if claims.Nationality != "US" {
		t.Errorf("nationality: got %q", claims.Nationality)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != "operator" {
		t.Errorf("roles: %v", claims.Roles)
	}
	if len(claims.AMR) != 2 || claims.AMR[0] != "pwd" {
		t.Errorf("amr: %v", claims.AMR)
	}
	if claims.Issuer != "https://iam.test.chetana.p9e.in" {
		t.Errorf("issuer: got %q", claims.Issuer)
	}
	if claims.ID == "" {
		t.Error("jti must be populated")
	}
	if claims.IssuedAt == nil || !claims.IssuedAt.Time.Equal(now) {
		t.Errorf("iat: got %v want %v", claims.IssuedAt, now)
	}
	if claims.ExpiresAt == nil || !claims.ExpiresAt.Time.Equal(now.Add(15*time.Minute)) {
		t.Errorf("exp: got %v", claims.ExpiresAt)
	}
}

func TestIssueAccessToken_DefaultAudience(t *testing.T) {
	iss, _, _ := newTestIssuer(t)
	_, claims, err := iss.IssueAccessToken(IssueInput{
		UserID:    "u",
		TenantID:  "t",
		SessionID: "s",
	})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if len(claims.Audience) != 1 || claims.Audience[0] != "chetana-api" {
		t.Errorf("default audience: %v", claims.Audience)
	}
}

func TestIssueAccessToken_ValidationErrors(t *testing.T) {
	iss, _, _ := newTestIssuer(t)

	tests := []struct {
		name string
		in   IssueInput
	}{
		{"empty user", IssueInput{TenantID: "t", SessionID: "s"}},
		{"empty tenant", IssueInput{UserID: "u", SessionID: "s"}},
		{"empty session", IssueInput{UserID: "u", TenantID: "t"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := iss.IssueAccessToken(tc.in); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestNewIssuer_Validation(t *testing.T) {
	store, _ := newTestStore(t, time.Now())

	if _, err := NewIssuer(nil, IssuerConfig{Issuer: "x"}); err == nil {
		t.Error("nil store should error")
	}
	if _, err := NewIssuer(store, IssuerConfig{Issuer: ""}); err == nil {
		t.Error("empty issuer should error")
	}

	emptyStore := NewKeyStore(nil)
	if _, err := NewIssuer(emptyStore, IssuerConfig{Issuer: "x"}); err == nil {
		t.Error("empty key store should error")
	}
}

func TestClaims_AsPrincipal(t *testing.T) {
	iss, _, now := newTestIssuer(t)
	_, claims, err := iss.IssueAccessToken(IssueInput{
		UserID:         "u",
		TenantID:       "t",
		SessionID:      "s",
		ClearanceLevel: "restricted",
		Roles:          []string{"r1", "r2"},
		Scopes:         []string{"sc1"},
		AMR:            []string{"pwd"},
	})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	p := claims.AsPrincipal()
	if p.UserID != "u" || p.TenantID != "t" || p.SessionID != "s" {
		t.Errorf("identity: %+v", p)
	}
	if p.ClearanceLevel != "restricted" {
		t.Errorf("clearance: %q", p.ClearanceLevel)
	}
	if len(p.Roles) != 2 || len(p.Scopes) != 1 || len(p.AMR) != 1 {
		t.Errorf("collections: %+v", p)
	}
	if !p.IssuedAt.Equal(now) {
		t.Errorf("iat: %v", p.IssuedAt)
	}
	if !p.ExpiresAt.Equal(now.Add(15 * time.Minute)) {
		t.Errorf("exp: %v", p.ExpiresAt)
	}

	// Defensive copy: mutating principal slices must not affect claims.
	p.Roles[0] = "MUTATED"
	if claims.Roles[0] == "MUTATED" {
		t.Error("AsPrincipal must defensively copy Roles")
	}
}

func TestParseUnverified(t *testing.T) {
	iss, _, _ := newTestIssuer(t)
	signed, _, err := iss.IssueAccessToken(IssueInput{
		UserID:    "u",
		TenantID:  "t",
		SessionID: "s",
	})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	got, err := ParseUnverified(signed)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got.Subject != "u" || got.TenantID != "t" {
		t.Errorf("parsed claims wrong: %+v", got)
	}
}

func TestParseUnverified_Garbage(t *testing.T) {
	if _, err := ParseUnverified("not.a.jwt!!"); err == nil {
		t.Error("garbage should error")
	}
}

func TestGenerateRSAKey_RejectsWeakBits(t *testing.T) {
	if _, err := GenerateRSAKey(1024); err == nil {
		t.Error("1024-bit RSA must be rejected")
	}
}

func TestNewIssuer_DefaultsTTL(t *testing.T) {
	store, _ := newTestStore(t, time.Now())
	iss, err := NewIssuer(store, IssuerConfig{Issuer: "x"})
	if err != nil {
		t.Fatalf("new issuer: %v", err)
	}
	if iss.tokenTTL != DefaultAccessTokenTTL {
		t.Errorf("ttl: got %v want %v", iss.tokenTTL, DefaultAccessTokenTTL)
	}
}

func TestIssueAccessToken_ReturnsErrorWhenNoActiveKey(t *testing.T) {
	// Drive the store clock past the only key's Retirement; Issuer.Active()
	// then has nothing to pick.
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	priv, err := GenerateRSAKey(2048)
	if err != nil {
		t.Fatalf("gen: %v", err)
	}
	currentTime := now
	store := NewKeyStore(func() time.Time { return currentTime })
	if err := store.Add(SigningKey{
		KeyID:      "k",
		Private:    priv,
		Activation: now,
		Retirement: now.Add(time.Hour),
	}); err != nil {
		t.Fatalf("add: %v", err)
	}
	iss, err := NewIssuer(store, IssuerConfig{
		Issuer: "x",
		Clock:  func() time.Time { return currentTime },
	})
	if err != nil {
		t.Fatalf("new issuer: %v", err)
	}
	// Advance both clocks past retirement.
	currentTime = now.Add(2 * time.Hour)
	if _, _, err := iss.IssueAccessToken(IssueInput{
		UserID:    "u",
		TenantID:  "t",
		SessionID: "s",
	}); err == nil || !errors.Is(err, ErrNoActiveKey) {
		t.Errorf("expected ErrNoActiveKey, got %v", err)
	}
}
