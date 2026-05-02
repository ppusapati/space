// jwks.go — RSA key store + JWKS HTTP endpoint + key rotation.
//
// → REQ-FUNC-PLT-IAM-002 (JWKS rotation: a new key appears in
//                          /jwks.json 24h before becoming the active
//                          signing key).
// → design.md §4.1.1
//
// Lifecycle of a key:
//
//   added    → in JWKS, NOT signing yet (rotation overlap window)
//   active   → in JWKS, used for signing
//   retired  → in JWKS for the legacy-token TTL window, then dropped
//
// The "added" → "active" promotion is wall-clock-driven by
// SigningKeyAt(now): keys whose Activation time has passed AND whose
// Retirement time is in the future are eligible to sign; among them
// the one with the most recent Activation wins. This lets operators
// stage a rotation by inserting a new key with Activation = now + 24h
// and walking away — the cutover happens automatically.

package token

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sort"
	"sync"
	"time"
)

// SigningKey is one entry in the KeyStore: an RSA private key plus
// the metadata that drives rotation.
type SigningKey struct {
	// KeyID is the JWS `kid` header. Stable for the life of the key.
	KeyID string

	// Private holds the RSA private key. The verifier exposes only
	// the public half (KeyStore.PublicKeyForKID).
	Private *rsa.PrivateKey

	// Activation is the wall-clock instant from which this key
	// becomes the active signing key. Until Activation the key is
	// published in JWKS (so verifiers learn of it ahead of time)
	// but is not chosen by SigningKeyAt.
	Activation time.Time

	// Retirement is the wall-clock instant after which this key is
	// dropped from JWKS entirely. Tokens signed with this key
	// remain verifiable until Retirement; clients must rotate
	// before that.
	Retirement time.Time
}

// IsActiveAt reports whether `t` is inside [Activation, Retirement).
func (k SigningKey) IsActiveAt(t time.Time) bool {
	if t.Before(k.Activation) {
		return false
	}
	if !k.Retirement.IsZero() && !t.Before(k.Retirement) {
		return false
	}
	return true
}

// IsPublishableAt reports whether the key should appear in the JWKS
// at time t. Keys remain publishable from creation through
// Retirement (so verifiers learn of new keys 24h ahead and stale
// ones disappear once their tokens cannot be valid any more).
func (k SigningKey) IsPublishableAt(t time.Time) bool {
	if !k.Retirement.IsZero() && !t.Before(k.Retirement) {
		return false
	}
	return true
}

// KeyStore holds the IAM service's RSA key set. Construct with
// NewKeyStore; populate with Add. The store is goroutine-safe.
type KeyStore struct {
	mu    sync.RWMutex
	keys  []SigningKey
	clock func() time.Time
}

// NewKeyStore returns an empty KeyStore. Pass clock=nil for
// time.Now; tests inject a frozen clock.
func NewKeyStore(clock func() time.Time) *KeyStore {
	if clock == nil {
		clock = time.Now
	}
	return &KeyStore{clock: clock}
}

// Add inserts a key. Returns ErrDuplicateKey when KeyID is already
// present.
func (s *KeyStore) Add(k SigningKey) error {
	if k.KeyID == "" {
		return errors.New("token: empty KeyID")
	}
	if k.Private == nil {
		return errors.New("token: nil private key")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.keys {
		if existing.KeyID == k.KeyID {
			return fmt.Errorf("token: %w: kid=%s", ErrDuplicateKey, k.KeyID)
		}
	}
	s.keys = append(s.keys, k)
	return nil
}

// Active returns the key the issuer should sign with right now.
// Among all active keys it picks the one with the most-recent
// Activation, so an operator who stages a new key with
// Activation = now + 24h gets clean cutover at that instant.
func (s *KeyStore) Active() (SigningKey, error) {
	return s.SigningKeyAt(s.clock())
}

// SigningKeyAt returns the key the issuer should sign with at
// instant t. Exposed for tests; callers in production use Active().
func (s *KeyStore) SigningKeyAt(t time.Time) (SigningKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var (
		best       SigningKey
		bestSet    bool
	)
	for _, k := range s.keys {
		if !k.IsActiveAt(t) {
			continue
		}
		if !bestSet || k.Activation.After(best.Activation) {
			best = k
			bestSet = true
		}
	}
	if !bestSet {
		return SigningKey{}, ErrNoActiveKey
	}
	return best, nil
}

// snapshot returns a copy of the current key set. Used by JWKSet
// + PublicKeyForKID to read without holding the lock for long.
func (s *KeyStore) snapshot() []SigningKey {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]SigningKey, len(s.keys))
	copy(out, s.keys)
	return out
}

// JWKSet builds the public-facing JWKS payload at instant t. Keys
// past their Retirement are excluded.
func (s *KeyStore) JWKSet(t time.Time) JWKSet {
	keys := s.snapshot()
	out := JWKSet{Keys: make([]JWK, 0, len(keys))}
	for _, k := range keys {
		if !k.IsPublishableAt(t) {
			continue
		}
		out.Keys = append(out.Keys, marshalRSAPublic(&k.Private.PublicKey, k.KeyID))
	}
	// Stable order — kid ascending — so caches don't churn between
	// scrapes.
	sort.Slice(out.Keys, func(i, j int) bool { return out.Keys[i].KID < out.Keys[j].KID })
	return out
}

// JWKSHandler returns an http.Handler serving the canonical
// /.well-known/jwks.json payload. Cache-Control headers tell
// downstream verifiers they may cache for 1 hour; the rotation
// overlap is wider than 1h so a stale cache cannot trip a verifier.
func (s *KeyStore) JWKSHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := json.MarshalIndent(s.JWKSet(s.clock()), "", "  ")
		if err != nil {
			http.Error(w, "json marshal", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/jwk-set+json")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		_, _ = w.Write(body)
	})
}

// GenerateRSAKey is a convenience helper for the service entrypoint
// + tests. Production should generate keys out-of-band and store
// the private bytes in AWS Secrets Manager (REQ-NFR-SEC-003); this
// helper stays here so dev environments can boot without ceremony.
func GenerateRSAKey(bits int) (*rsa.PrivateKey, error) {
	if bits < 2048 {
		return nil, errors.New("token: RSA key must be >= 2048 bits")
	}
	return rsa.GenerateKey(rand.Reader, bits)
}

// ----------------------------------------------------------------------
// JWKS payload types
// ----------------------------------------------------------------------

// JWKSet is the JSON payload returned by /.well-known/jwks.json per
// RFC 7517 §5.
type JWKSet struct {
	Keys []JWK `json:"keys"`
}

// JWK is one entry in a JWKS — RFC 7517 §4 + RFC 7518 §6.3 for
// RSA public keys. The chetana issuer only emits RS256 today; if
// we ever add ECDSA we'll add a second JWK shape (or use a
// generic representation).
type JWK struct {
	Kty string `json:"kty"`           // "RSA"
	Use string `json:"use"`           // "sig"
	Alg string `json:"alg"`           // "RS256"
	KID string `json:"kid"`
	N   string `json:"n"`             // base64url-unpadded big-endian modulus
	E   string `json:"e"`             // base64url-unpadded big-endian exponent
}

// marshalRSAPublic builds a JWK from an *rsa.PublicKey + key id.
// base64url-unpadded encoding per RFC 7518 §6.3.1.
func marshalRSAPublic(pub *rsa.PublicKey, kid string) JWK {
	return JWK{
		Kty: "RSA",
		Use: "sig",
		Alg: SigningAlgorithm,
		KID: kid,
		N:   base64URLUnpaddedBigInt(pub.N),
		E:   base64URLUnpaddedExponent(pub.E),
	}
}

func base64URLUnpaddedBigInt(n *big.Int) string {
	return base64.RawURLEncoding.EncodeToString(n.Bytes())
}

func base64URLUnpaddedExponent(e int) string {
	// RFC 7518 §6.3.1.2 says the exponent is the minimum number of
	// big-endian bytes; standard usage is 3 bytes for 0x010001.
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(e))
	for len(buf) > 1 && buf[0] == 0 {
		buf = buf[1:]
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}

// SHA256KID derives a stable Key ID from an RSA public key by
// hashing its SubjectPublicKeyInfo-equivalent canonical bytes.
// Convenience for tests + dev startup.
func SHA256KID(pub *rsa.PublicKey) string {
	h := sha256.New()
	h.Write(pub.N.Bytes())
	binary.Write(h, binary.BigEndian, uint32(pub.E))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil)[:16])
}

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrNoActiveKey is returned when the KeyStore contains no key
// whose Activation has elapsed.
var ErrNoActiveKey = errors.New("token: no active signing key")

// ErrDuplicateKey is returned by KeyStore.Add when a key with the
// same KeyID is already present.
var ErrDuplicateKey = errors.New("token: duplicate key")

// ErrUnknownKey is returned by PublicKeyForKID when the requested
// key id is not in the store.
var ErrUnknownKey = errors.New("token: unknown key id")
