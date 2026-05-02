package saml

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"errors"
	"math/big"
	"net/url"
	"testing"
	"time"
)

// genTestCert returns a self-signed RSA-2048 cert + key suitable
// for SP / IdP test fixtures.
func genTestCert(t *testing.T) (*rsa.PrivateKey, *x509.Certificate) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}
	return key, cert
}

func TestEncodeDecodeCertificate_Roundtrip(t *testing.T) {
	_, cert := genTestCert(t)
	pemBytes := EncodeCertificate(cert)
	parsed, err := ParseCertificate(pemBytes)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parsed.SerialNumber.Cmp(cert.SerialNumber) != 0 {
		t.Error("serial mismatch after roundtrip")
	}
}

func TestParseCertificate_RejectsNonPEM(t *testing.T) {
	if _, err := ParseCertificate([]byte("not a pem block")); err == nil {
		t.Error("expected error")
	}
}

func TestBuildServiceProvider_Validation(t *testing.T) {
	key, cert := genTestCert(t)
	idp := &IdP{
		EntityID:    "https://idp.example.com/metadata",
		SSOURL:      "https://idp.example.com/sso",
		Certificate: cert,
	}
	cases := []struct {
		name string
		sp   SPConfig
	}{
		{"empty entity id", SPConfig{AssertionConsumerServiceURL: "https://sp/acs", PrivateKey: key, Certificate: cert}},
		{"empty acs", SPConfig{EntityID: "https://sp/meta", PrivateKey: key, Certificate: cert}},
		{"nil key", SPConfig{EntityID: "https://sp/meta", AssertionConsumerServiceURL: "https://sp/acs", Certificate: cert}},
		{"nil cert", SPConfig{EntityID: "https://sp/meta", AssertionConsumerServiceURL: "https://sp/acs", PrivateKey: key}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := BuildServiceProvider(tc.sp, idp); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestBuildServiceProvider_HappyPath(t *testing.T) {
	key, cert := genTestCert(t)
	sp := SPConfig{
		EntityID:                    "https://iam.test.chetana.p9e.in/saml/metadata",
		AssertionConsumerServiceURL: "https://iam.test.chetana.p9e.in/saml/acs/1",
		PrivateKey:                  key,
		Certificate:                 cert,
	}
	idp := &IdP{
		EntityID:    "https://idp.example.com/metadata",
		SSOURL:      "https://idp.example.com/sso",
		Certificate: cert,
	}
	provider, err := BuildServiceProvider(sp, idp)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if provider.EntityID != sp.EntityID {
		t.Errorf("entity id: %q", provider.EntityID)
	}
	if provider.AcsURL.String() != sp.AssertionConsumerServiceURL {
		t.Errorf("acs url: %q", provider.AcsURL.String())
	}
	if provider.IDPMetadata == nil {
		t.Fatal("idp metadata not built")
	}
	if len(provider.IDPMetadata.IDPSSODescriptors) != 1 {
		t.Errorf("descriptors: %d", len(provider.IDPMetadata.IDPSSODescriptors))
	}
	// The SSO URL parses cleanly.
	if _, err := url.Parse(idp.SSOURL); err != nil {
		t.Errorf("sso url unparseable: %v", err)
	}
}

func TestNewService_Validation(t *testing.T) {
	key, cert := genTestCert(t)
	good := SPConfig{
		EntityID:                    "https://sp/meta",
		AssertionConsumerServiceURL: "https://sp/acs",
		PrivateKey:                  key,
		Certificate:                 cert,
	}
	cases := []struct {
		name string
		sp   SPConfig
	}{
		{"missing entity id", SPConfig{AssertionConsumerServiceURL: "x", PrivateKey: key, Certificate: cert}},
		{"missing acs", SPConfig{EntityID: "x", PrivateKey: key, Certificate: cert}},
		{"missing key", SPConfig{EntityID: "x", AssertionConsumerServiceURL: "x", Certificate: cert}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewService(tc.sp, &Store{}, &JITProvisioner{}); err == nil {
				t.Fatal("expected error")
			}
		})
	}

	if _, err := NewService(good, nil, &JITProvisioner{}); err == nil {
		t.Error("nil store should error")
	}
	if _, err := NewService(good, &Store{}, nil); err == nil {
		t.Error("nil jit should error")
	}
}

func TestNewJITProvisioner_Validation(t *testing.T) {
	if _, err := NewJITProvisioner(nil, "t", nil); err == nil {
		t.Error("nil pool should error")
	}
}

func TestAttributeMapping_JSONRoundtrip(t *testing.T) {
	m := AttributeMapping{
		EmailAttribute:       "email",
		DisplayNameAttribute: "displayName",
		GroupsAttribute:      "groups",
		GroupRoleMap:         map[string]string{"a": "b"},
		DefaultRoles:         []string{"viewer"},
	}
	body, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got AttributeMapping
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.EmailAttribute != m.EmailAttribute || got.GroupsAttribute != m.GroupsAttribute {
		t.Error("scalar mismatch")
	}
	if got.GroupRoleMap["a"] != "b" {
		t.Errorf("map: %v", got.GroupRoleMap)
	}
}

func TestErrors_Sentinel(t *testing.T) {
	for _, e := range []error{ErrIdPNotFound, ErrSignatureInvalid, ErrMissingEmail} {
		if !errors.Is(e, e) {
			t.Errorf("not reflexive: %v", e)
		}
	}
}
