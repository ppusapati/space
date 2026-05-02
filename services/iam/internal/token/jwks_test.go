package token

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSigningKey_IsActiveAt(t *testing.T) {
	t0 := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	k := SigningKey{
		Activation: t0,
		Retirement: t0.Add(24 * time.Hour),
	}
	cases := []struct {
		t    time.Time
		want bool
	}{
		{t0.Add(-time.Second), false},
		{t0, true},
		{t0.Add(12 * time.Hour), true},
		{t0.Add(24 * time.Hour), false},
		{t0.Add(48 * time.Hour), false},
	}
	for _, tc := range cases {
		if got := k.IsActiveAt(tc.t); got != tc.want {
			t.Errorf("IsActiveAt(%v): got %v want %v", tc.t, got, tc.want)
		}
	}
}

func TestKeyStore_Add_DuplicateRejected(t *testing.T) {
	store := NewKeyStore(nil)
	priv, _ := GenerateRSAKey(2048)
	k := SigningKey{KeyID: "kid-1", Private: priv, Activation: time.Now()}
	if err := store.Add(k); err != nil {
		t.Fatalf("first add: %v", err)
	}
	if err := store.Add(k); !errors.Is(err, ErrDuplicateKey) {
		t.Errorf("duplicate add: got %v want ErrDuplicateKey", err)
	}
}

func TestKeyStore_Add_Validation(t *testing.T) {
	store := NewKeyStore(nil)
	if err := store.Add(SigningKey{}); err == nil {
		t.Error("empty key id should error")
	}
	if err := store.Add(SigningKey{KeyID: "x"}); err == nil {
		t.Error("nil private key should error")
	}
}

// Rotation overlap: REQ-FUNC-PLT-IAM-002 acceptance #2.
// A new key appears in /jwks.json 24h before becoming the signing key.
func TestKeyStore_RotationOverlap_24hAhead(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)

	// Old key: active now, retires in 25h.
	oldPriv, _ := GenerateRSAKey(2048)
	oldKey := SigningKey{
		KeyID:      "kid-old",
		Private:    oldPriv,
		Activation: now.Add(-24 * time.Hour),
		Retirement: now.Add(25 * time.Hour),
	}

	// New key: published now, activates in 24h.
	newPriv, _ := GenerateRSAKey(2048)
	newKey := SigningKey{
		KeyID:      "kid-new",
		Private:    newPriv,
		Activation: now.Add(24 * time.Hour),
		Retirement: now.Add(7 * 24 * time.Hour),
	}

	store := NewKeyStore(func() time.Time { return now })
	if err := store.Add(oldKey); err != nil {
		t.Fatalf("add old: %v", err)
	}
	if err := store.Add(newKey); err != nil {
		t.Fatalf("add new: %v", err)
	}

	// At now: signing key is the old one.
	signing, err := store.SigningKeyAt(now)
	if err != nil {
		t.Fatalf("signing now: %v", err)
	}
	if signing.KeyID != "kid-old" {
		t.Errorf("signing key now: got %q want kid-old", signing.KeyID)
	}

	// At now: JWKS publishes BOTH keys (so verifiers learn the new one ahead).
	set := store.JWKSet(now)
	if len(set.Keys) != 2 {
		t.Fatalf("JWKS at now: got %d keys, want 2", len(set.Keys))
	}
	gotKIDs := map[string]bool{set.Keys[0].KID: true, set.Keys[1].KID: true}
	if !gotKIDs["kid-old"] || !gotKIDs["kid-new"] {
		t.Errorf("JWKS keys: %+v", set.Keys)
	}

	// One nanosecond before activation: still old.
	signing, err = store.SigningKeyAt(now.Add(24*time.Hour - time.Nanosecond))
	if err != nil {
		t.Fatalf("signing pre-cutover: %v", err)
	}
	if signing.KeyID != "kid-old" {
		t.Errorf("pre-cutover: got %q want kid-old", signing.KeyID)
	}

	// At activation instant: cuts over to new.
	signing, err = store.SigningKeyAt(now.Add(24 * time.Hour))
	if err != nil {
		t.Fatalf("signing at cutover: %v", err)
	}
	if signing.KeyID != "kid-new" {
		t.Errorf("at cutover: got %q want kid-new", signing.KeyID)
	}
}

func TestKeyStore_JWKSet_ExcludesRetired(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	priv, _ := GenerateRSAKey(2048)
	store := NewKeyStore(func() time.Time { return now })
	if err := store.Add(SigningKey{
		KeyID:      "retired",
		Private:    priv,
		Activation: now.Add(-72 * time.Hour),
		Retirement: now.Add(-time.Hour),
	}); err != nil {
		t.Fatalf("add: %v", err)
	}
	set := store.JWKSet(now)
	if len(set.Keys) != 0 {
		t.Errorf("retired key leaked into JWKS: %+v", set.Keys)
	}
}

func TestKeyStore_Active_NoneAvailable(t *testing.T) {
	store := NewKeyStore(nil)
	if _, err := store.Active(); !errors.Is(err, ErrNoActiveKey) {
		t.Errorf("Active on empty: got %v want ErrNoActiveKey", err)
	}
}

func TestKeyStore_PublicKeyForKID(t *testing.T) {
	priv, _ := GenerateRSAKey(2048)
	store := NewKeyStore(nil)
	_ = store.Add(SigningKey{
		KeyID:      "kid-x",
		Private:    priv,
		Activation: time.Now(),
		Retirement: time.Now().Add(time.Hour),
	})
	pub, err := store.PublicKeyForKID("kid-x")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if pub == nil || pub.N.Cmp(priv.PublicKey.N) != 0 {
		t.Error("public key mismatch")
	}
	if _, err := store.PublicKeyForKID("missing"); !errors.Is(err, ErrUnknownKey) {
		t.Errorf("missing kid: got %v want ErrUnknownKey", err)
	}
}

func TestJWKSHandler_ServesJSON(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	priv, _ := GenerateRSAKey(2048)
	store := NewKeyStore(func() time.Time { return now })
	_ = store.Add(SigningKey{
		KeyID:      "kid-h",
		Private:    priv,
		Activation: now.Add(-time.Hour),
		Retirement: now.Add(time.Hour),
	})

	srv := httptest.NewServer(store.JWKSHandler())
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); got != "application/jwk-set+json" {
		t.Errorf("content-type: %q", got)
	}
	if got := resp.Header.Get("Cache-Control"); got != "public, max-age=3600" {
		t.Errorf("cache-control: %q", got)
	}

	var set JWKSet
	if err := json.NewDecoder(resp.Body).Decode(&set); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(set.Keys) != 1 || set.Keys[0].KID != "kid-h" {
		t.Errorf("body: %+v", set)
	}
	if set.Keys[0].Kty != "RSA" || set.Keys[0].Alg != "RS256" || set.Keys[0].Use != "sig" {
		t.Errorf("jwk shape: %+v", set.Keys[0])
	}
	if set.Keys[0].N == "" || set.Keys[0].E == "" {
		t.Error("jwk N/E empty")
	}
}

func TestSHA256KID_Deterministic(t *testing.T) {
	priv, _ := GenerateRSAKey(2048)
	a := SHA256KID(&priv.PublicKey)
	b := SHA256KID(&priv.PublicKey)
	if a != b {
		t.Errorf("SHA256KID not deterministic: %s != %s", a, b)
	}
	if a == "" {
		t.Error("empty KID")
	}
}
