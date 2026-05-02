//go:build integration

// saml_test.go — TASK-P1-IAM-006 SAML integration tests.
//
// Drives a stub SAML IdP (using crewjam/saml's IdentityProvider
// helper) against the chetana SP end-to-end:
//
//   1. The chetana SP builds an AuthnRequest and we hand it to
//      the stub IdP via its ServeSSO HTTP handler.
//   2. The IdP's DefaultAssertionMaker mints a signed Response
//      and we capture the SAMLResponse form value.
//   3. We POST that SAMLResponse back at the chetana SP's ACS
//      endpoint via Service.FinishSSO. The protocol library
//      verifies the XML signature against the IdP cert we
//      registered with the chetana store.
//   4. Assertions: the user is JIT-provisioned with the mapped
//      roles + the unsigned-tampered variant is rejected.

package iam_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	cs "github.com/crewjam/saml"
	"github.com/crewjam/saml/logger"
	"github.com/jackc/pgx/v5/pgxpool"

	chetsaml "github.com/ppusapati/space/services/iam/internal/saml"
)

func newSAMLPool(t *testing.T) *pgxpool.Pool {
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
			`TRUNCATE saml_idps, users RESTART IDENTITY CASCADE`)
		pool.Close()
	})
	if _, err := pool.Exec(context.Background(),
		`TRUNCATE saml_idps, users RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return pool
}

func mustGenCert(t *testing.T) (*rsa.PrivateKey, *x509.Certificate) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
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

// Stub ServiceProviderProvider that returns the chetana SP's
// EntityDescriptor for any requested ID.
type stubSPProvider struct {
	md *cs.EntityDescriptor
}

func (s *stubSPProvider) GetServiceProvider(_ *http.Request, _ string) (*cs.EntityDescriptor, error) {
	return s.md, nil
}

// Stub SessionProvider that always returns the seeded session,
// echoing the configured user attributes onto the assertion.
type stubSessionProvider struct {
	session *cs.Session
}

func (s *stubSessionProvider) GetSession(_ http.ResponseWriter, _ *http.Request, _ *cs.IdpAuthnRequest) *cs.Session {
	return s.session
}

// samlRig is the in-process SP+IdP harness.
type samlRig struct {
	pool      *pgxpool.Pool
	store     *chetsaml.Store
	jit       *chetsaml.JITProvisioner
	svc       *chetsaml.Service
	idp       *cs.IdentityProvider
	idpRow    *chetsaml.IdP
	spURL     string
	spConfig  chetsaml.SPConfig
	tenantID  string
	idpServer *httptest.Server
}

func setupSAMLRig(t *testing.T) *samlRig {
	t.Helper()
	pool := newSAMLPool(t)
	tenantID := "22222222-2222-2222-2222-222222222222"

	// Bootstrap the users table — the JIT provisioner inserts here.
	if _, err := pool.Exec(context.Background(), `
		INSERT INTO users (tenant_id, email_lower, email_display, password_hash, password_algo, status, data_classification, created_at, updated_at)
		VALUES ($1, 'seed@example.com', 'Seed', '', '', 'active', 'cui', now(), now())
	`, tenantID); err != nil {
		// Some test runs already have rows; ignore unique violations
		// because the test cleanup TRUNCATES at start anyway.
		t.Logf("seed user (ignored if dup): %v", err)
	}

	// Stand up a chetana SP credential pair.
	spKey, spCert := mustGenCert(t)
	spURL := "https://sp.test.chetana.p9e.in"
	spCfg := chetsaml.SPConfig{
		EntityID:                    spURL + "/saml/metadata",
		AssertionConsumerServiceURL: spURL + "/saml/acs/1",
		PrivateKey:                  spKey,
		Certificate:                 spCert,
	}

	// Stand up a stub IdP — same crewjam library but in
	// IdentityProvider role. Sign assertions with its own key pair.
	idpKey, idpCert := mustGenCert(t)
	idpServerMux := http.NewServeMux()
	idpServer := httptest.NewServer(idpServerMux)
	t.Cleanup(idpServer.Close)

	idpEntityID := idpServer.URL + "/metadata"
	idpSSOURL := idpServer.URL + "/sso"

	idp := &cs.IdentityProvider{
		Key:         idpKey,
		Certificate: idpCert,
		Logger:      logger.DefaultLogger,
		MetadataURL: mustParseURL(t, idpEntityID),
		SSOURL:      mustParseURL(t, idpSSOURL),
		ServiceProviderProvider: &stubSPProvider{}, // late-bound below
	}

	// Persist the IdP row in the chetana store.
	store := chetsaml.NewStore(pool, time.Now)
	idpID, err := store.CreateForTest(context.Background(), chetsaml.IdP{
		Name:        "Stub IdP",
		EntityID:    idpEntityID,
		SSOURL:      idpSSOURL,
		Certificate: idpCert,
		AttributeMapping: chetsaml.AttributeMapping{
			EmailAttribute:       "urn:oid:0.9.2342.19200300.100.1.3", // mail
			DisplayNameAttribute: "urn:oid:2.16.840.1.113730.3.1.241", // displayName
			GroupsAttribute:      "urn:oid:1.3.6.1.4.1.5923.1.5.1.1",  // isMemberOf
			GroupRoleMap: map[string]string{
				"chetana-operators":     "operator",
				"chetana-mission-leads": "mission_lead",
			},
			DefaultRoles: []string{"viewer"},
		},
	})
	if err != nil {
		t.Fatalf("create idp row: %v", err)
	}
	idpRow, err := store.LookupByID(context.Background(), idpID)
	if err != nil {
		t.Fatalf("lookup idp: %v", err)
	}

	// Build the chetana SP using the freshly-loaded row so the
	// SP's ACS URL bakes in the right idp id.
	spCfg.AssertionConsumerServiceURL = fmt.Sprintf("%s/saml/acs/%d", spURL, idpID)

	jit, err := chetsaml.NewJITProvisioner(pool, tenantID, time.Now)
	if err != nil {
		t.Fatalf("jit: %v", err)
	}
	svc, err := chetsaml.NewService(spCfg, store, jit)
	if err != nil {
		t.Fatalf("service: %v", err)
	}

	// Wire the IdP's SP-provider so it returns the chetana SP's
	// metadata for any lookup. We need the chetana SP's
	// EntityDescriptor for that — build it via crewjam's SP type.
	spProvider, err := chetsaml.BuildServiceProvider(spCfg, idpRow)
	if err != nil {
		t.Fatalf("build sp: %v", err)
	}
	idp.ServiceProviderProvider = &stubSPProvider{md: spProvider.Metadata()}

	// Mount the IdP's /sso handler. The session is set per-test
	// via overrideSession.
	idpServerMux.HandleFunc("/sso", func(w http.ResponseWriter, r *http.Request) {
		idp.ServeSSO(w, r)
	})

	return &samlRig{
		pool:      pool,
		store:     store,
		jit:       jit,
		svc:       svc,
		idp:       idp,
		idpRow:    idpRow,
		spURL:     spURL,
		spConfig:  spCfg,
		tenantID:  tenantID,
		idpServer: idpServer,
	}
}

func mustParseURL(t *testing.T, s string) url.URL {
	t.Helper()
	u, err := url.Parse(s)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return *u
}

// driveSSO walks the full SAML flow: SP builds AuthnRequest →
// IdP receives it via ServeSSO → IdP renders the SAMLResponse
// HTML form → we extract the SAMLResponse value → POST it back
// to the chetana SP's FinishSSO and return the result.
func driveSSO(t *testing.T, rig *samlRig, session *cs.Session) (*chetsaml.SSOResult, error) {
	t.Helper()
	rig.idp.SessionProvider = &stubSessionProvider{session: session}

	// 1. SP builds the redirect URL (which carries the deflated
	//    SAMLRequest).
	sso, err := rig.svc.BeginSSO(context.Background(), rig.idpRow.ID, "test-relay")
	if err != nil {
		return nil, fmt.Errorf("BeginSSO: %w", err)
	}
	parsed, err := url.Parse(sso.RedirectURL)
	if err != nil {
		return nil, fmt.Errorf("parse redirect: %w", err)
	}

	// 2. GET the IdP's /sso with the same query params. The IdP's
	//    ServeSSO writes the SAMLResponse HTML form into the
	//    response body.
	req, err := http.NewRequest("GET", rig.idpServer.URL+"/sso", nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = parsed.RawQuery
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET sso: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("idp sso: status=%d body=%s", resp.StatusCode, body)
	}

	// 3. Pull the SAMLResponse value out of the HTML form.
	samlResponse := extractFormValue(string(body), "SAMLResponse")
	if samlResponse == "" {
		return nil, fmt.Errorf("no SAMLResponse in body: %s", body)
	}

	// 4. Build a synthetic POST request to the chetana ACS, then
	//    drive FinishSSO. The protocol library reads the form
	//    value off the request.
	form := url.Values{}
	form.Set("SAMLResponse", samlResponse)
	form.Set("RelayState", "test-relay")
	acsReq, err := http.NewRequest("POST",
		rig.spConfig.AssertionConsumerServiceURL,
		strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	acsReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return rig.svc.FinishSSO(context.Background(), rig.idpRow.ID, acsReq, []string{sso.RequestID})
}

// extractFormValue is a tiny HTML scraper for the IdP's auto-
// submitting form. The library's template puts each value in a
// `value="..."` attribute on a hidden input; we grab the first
// value attribute that follows the named input.
func extractFormValue(body, name string) string {
	marker := fmt.Sprintf(`name="%s"`, name)
	idx := strings.Index(body, marker)
	if idx < 0 {
		return ""
	}
	rest := body[idx:]
	valueIdx := strings.Index(rest, `value="`)
	if valueIdx < 0 {
		return ""
	}
	rest = rest[valueIdx+len(`value="`):]
	end := strings.Index(rest, `"`)
	if end < 0 {
		return ""
	}
	return rest[:end]
}

// Acceptance #1: signed assertion JIT-provisions a new user with
// the IdP's mapped roles.
func TestSAML_SignedAssertion_JITProvisionsNewUser(t *testing.T) {
	rig := setupSAMLRig(t)

	session := &cs.Session{
		ID:           "test-session",
		NameID:       "alice@example.com",
		NameIDFormat: "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress",
		CreateTime:   time.Now(),
		ExpireTime:   time.Now().Add(time.Hour),
		Index:        "1",
		UserEmail:    "alice@example.com",
		UserCommonName: "Alice Example",
		Groups: []string{
			"chetana-operators",
			"chetana-mission-leads",
			"unmapped-group",
		},
	}

	out, err := driveSSO(t, rig, session)
	if err != nil {
		t.Fatalf("driveSSO: %v", err)
	}
	if out.UserID == "" {
		t.Fatal("missing user id")
	}
	if !out.Created {
		t.Error("user should be JIT-provisioned (created)")
	}
	if out.Email != "alice@example.com" {
		t.Errorf("email: %q", out.Email)
	}
	wantRoles := map[string]bool{"operator": true, "mission_lead": true, "viewer": true}
	for _, r := range out.Roles {
		delete(wantRoles, r)
	}
	if len(wantRoles) > 0 {
		t.Errorf("missing roles: %v (got %v)", wantRoles, out.Roles)
	}

	// Re-run: existing user, NOT created again, same id.
	out2, err := driveSSO(t, rig, session)
	if err != nil {
		t.Fatalf("driveSSO 2: %v", err)
	}
	if out2.Created {
		t.Error("second run should NOT create the user")
	}
	if out2.UserID != out.UserID {
		t.Errorf("user id changed across runs: %q vs %q", out.UserID, out2.UserID)
	}
}

// Acceptance #2: tampering with a signed response invalidates the
// signature and the SP rejects it.
func TestSAML_TamperedAssertion_Rejected(t *testing.T) {
	rig := setupSAMLRig(t)
	session := &cs.Session{
		ID:           "tamper-session",
		NameID:       "evil@example.com",
		NameIDFormat: "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress",
		CreateTime:   time.Now(),
		ExpireTime:   time.Now().Add(time.Hour),
		Index:        "1",
		UserEmail:    "evil@example.com",
		Groups:       []string{"chetana-operators"},
	}
	rig.idp.SessionProvider = &stubSessionProvider{session: session}

	sso, err := rig.svc.BeginSSO(context.Background(), rig.idpRow.ID, "")
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	parsed, err := url.Parse(sso.RedirectURL)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	req, err := http.NewRequest("GET", rig.idpServer.URL+"/sso", nil)
	if err != nil {
		t.Fatalf("new req: %v", err)
	}
	req.URL.RawQuery = parsed.RawQuery
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	samlResponse := extractFormValue(string(body), "SAMLResponse")

	// Tamper: decode, flip a byte in the email-attribute payload,
	// re-encode. The signature must no longer verify.
	raw, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	tampered := bytes.Replace(raw, []byte("evil@example.com"), []byte("admin@example.com"), 1)
	if bytes.Equal(tampered, raw) {
		t.Fatal("test setup: tamper had no effect — email not present in response body")
	}
	tamperedB64 := base64.StdEncoding.EncodeToString(tampered)

	form := url.Values{}
	form.Set("SAMLResponse", tamperedB64)
	acsReq, _ := http.NewRequest("POST",
		rig.spConfig.AssertionConsumerServiceURL,
		strings.NewReader(form.Encode()))
	acsReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, err = rig.svc.FinishSSO(context.Background(), rig.idpRow.ID, acsReq, []string{sso.RequestID})
	if !errors.Is(err, chetsaml.ErrSignatureInvalid) {
		t.Errorf("got %v want ErrSignatureInvalid", err)
	}

	// And the user MUST NOT have been provisioned.
	var count int
	if err := rig.pool.QueryRow(context.Background(),
		`SELECT count(*) FROM users WHERE email_lower IN ('admin@example.com','evil@example.com')`,
	).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Errorf("tampered assertion should not provision user; count=%d", count)
	}
}

// Discovery: SP metadata XML carries the entity id + ACS URL the
// IdP needs to register chetana.
func TestSAML_MetadataXML(t *testing.T) {
	rig := setupSAMLRig(t)
	xmlBytes, err := rig.svc.MetadataXML(context.Background(), rig.idpRow.ID)
	if err != nil {
		t.Fatalf("metadata: %v", err)
	}
	body := string(xmlBytes)
	if !strings.Contains(body, rig.spConfig.EntityID) {
		t.Errorf("metadata missing entity id: %s", body)
	}
	if !strings.Contains(body, rig.spConfig.AssertionConsumerServiceURL) {
		t.Errorf("metadata missing ACS URL: %s", body)
	}
	if !strings.HasPrefix(body, "<?xml") {
		t.Errorf("metadata missing XML decl: %s", body[:64])
	}
}
