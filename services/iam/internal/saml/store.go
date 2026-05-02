// Package saml implements SAML 2.0 SP-initiated SSO + JIT user
// provisioning for the chetana IAM service.
//
// → REQ-FUNC-PLT-IAM-007 (SAML 2.0 SP, signed-assertion validation,
//                          JIT user provisioning with attribute → role
//                          mapping per IdP).
// → design.md §4.1.1.
//
// We delegate the SAML protocol layer (XML c14n, XML signature
// verification, AuthnRequest/Response marshalling, NameID parsing,
// SubjectConfirmation + audience + InResponseTo + NotBefore/
// NotOnOrAfter checks) to github.com/crewjam/saml + the underlying
// goxmldsig library. This package owns:
//
//   • Persistence of registered IdPs (entity_id, sso_url, x509_cert,
//     attribute_mapping JSONB).
//   • Constructing a per-IdP *saml.ServiceProvider from a row plus
//     the chetana SP's own private key + cert.
//   • The chetana JIT provisioning policy: on first ACS callback
//     for a new NameID, create a `users` row and project the IdP-
//     supplied attributes onto chetana roles via the per-IdP
//     attribute_mapping configuration.

package saml

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"time"

	cs "github.com/crewjam/saml"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// base64Std is the encoder used to emit X509Certificate values in
// the in-memory IdP metadata document. RFC 7468 / SAML metadata
// uses standard (padded) base64.
var base64Std = base64.StdEncoding

// IdP is the in-memory shape of one saml_idps row.
type IdP struct {
	ID                int64
	Name              string // human-readable label, e.g. "Acme SSO"
	EntityID          string // IdP entity ID
	SSOURL            string // IdP's SingleSignOnService Location
	SLOURL            string // IdP's SingleLogoutService Location (optional)
	Certificate       *x509.Certificate
	AttributeMapping  AttributeMapping
	Disabled          bool
	CreatedAt         time.Time
}

// AttributeMapping is the per-IdP configuration that drives JIT
// provisioning. Each field names the SAML attribute to read from
// the assertion's AttributeStatement.
//
//   EmailAttribute       — the user's email; primary key for matching
//                           an existing chetana user.
//   DisplayNameAttribute — the user's display name; optional.
//   GroupsAttribute      — multi-valued list of IdP group/role
//                           identifiers; optional.
//   GroupRoleMap         — mapping from IdP group identifier →
//                           chetana role. Unmapped groups are
//                           ignored (they do NOT become roles).
//   DefaultRoles         — roles every JIT-provisioned user gets,
//                           regardless of IdP attributes.
type AttributeMapping struct {
	EmailAttribute       string            `json:"email_attribute"`
	DisplayNameAttribute string            `json:"display_name_attribute"`
	GroupsAttribute      string            `json:"groups_attribute"`
	GroupRoleMap         map[string]string `json:"group_role_map"`
	DefaultRoles         []string          `json:"default_roles"`
}

// SPConfig is the chetana side of the SP. Same SP cert/key is
// shared across every registered IdP (a typical posture — only
// IdPs care about per-tenant SP isolation, and chetana is single-
// tenant in v1).
type SPConfig struct {
	// EntityID is the SP entity ID — typically
	// "https://iam.<region>.chetana.p9e.in/saml/metadata".
	EntityID string

	// AssertionConsumerServiceURL is the SP's ACS endpoint,
	// "https://iam.<region>.chetana.p9e.in/saml/acs".
	AssertionConsumerServiceURL string

	// PrivateKey + Certificate are the SP's signing credentials.
	// Used to sign AuthnRequests when the IdP requires it.
	PrivateKey  *rsa.PrivateKey
	Certificate *x509.Certificate
}

// Store wraps a pgx pool with the IdP persistence helpers.
type Store struct {
	pool *pgxpool.Pool
	clk  func() time.Time
}

// NewStore wraps a pool. clock=nil → time.Now.
func NewStore(pool *pgxpool.Pool, clock func() time.Time) *Store {
	if clock == nil {
		clock = time.Now
	}
	return &Store{pool: pool, clk: clock}
}

// LookupByEntityID fetches an IdP by its `entity_id`. Returns
// ErrIdPNotFound when the row is missing or disabled.
func (s *Store) LookupByEntityID(ctx context.Context, entityID string) (*IdP, error) {
	const q = `
SELECT id, name, entity_id, sso_url, COALESCE(slo_url, ''),
       x509_cert, attribute_mapping, disabled, created_at
FROM saml_idps
WHERE entity_id = $1
`
	return s.scanOne(ctx, q, entityID)
}

// LookupByID fetches an IdP by its surrogate id. Used by the ACS
// handler (the IdP id is part of the SP's unique
// AssertionConsumerServiceURL — `/saml/acs/{idp_id}`).
func (s *Store) LookupByID(ctx context.Context, id int64) (*IdP, error) {
	const q = `
SELECT id, name, entity_id, sso_url, COALESCE(slo_url, ''),
       x509_cert, attribute_mapping, disabled, created_at
FROM saml_idps
WHERE id = $1
`
	return s.scanOne(ctx, q, id)
}

func (s *Store) scanOne(ctx context.Context, q string, arg any) (*IdP, error) {
	var (
		idp        IdP
		certPEM    []byte
		mappingRaw []byte
	)
	err := s.pool.QueryRow(ctx, q, arg).Scan(
		&idp.ID, &idp.Name, &idp.EntityID, &idp.SSOURL, &idp.SLOURL,
		&certPEM, &mappingRaw, &idp.Disabled, &idp.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrIdPNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("saml: lookup: %w", err)
	}
	if idp.Disabled {
		return nil, ErrIdPNotFound
	}
	cert, err := ParseCertificate(certPEM)
	if err != nil {
		return nil, fmt.Errorf("saml: idp cert: %w", err)
	}
	idp.Certificate = cert
	if err := json.Unmarshal(mappingRaw, &idp.AttributeMapping); err != nil {
		return nil, fmt.Errorf("saml: attribute_mapping: %w", err)
	}
	return &idp, nil
}

// CreateForTest inserts a row directly. Production deploys
// register IdPs via an admin RPC; tests + ops scripts use this.
func (s *Store) CreateForTest(ctx context.Context, idp IdP) (int64, error) {
	mapping, err := json.Marshal(idp.AttributeMapping)
	if err != nil {
		return 0, fmt.Errorf("saml: marshal mapping: %w", err)
	}
	certPEM := EncodeCertificate(idp.Certificate)
	const q = `
INSERT INTO saml_idps
  (name, entity_id, sso_url, slo_url, x509_cert, attribute_mapping, disabled)
VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, $7)
RETURNING id
`
	var id int64
	err = s.pool.QueryRow(ctx, q,
		idp.Name, idp.EntityID, idp.SSOURL, idp.SLOURL,
		certPEM, mapping, idp.Disabled,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("saml: insert idp: %w", err)
	}
	return id, nil
}

// BuildServiceProvider returns a *saml.ServiceProvider configured
// for the given IdP. The returned value embeds the IdP's signing
// cert so that ParseResponse can verify the assertion signature.
func BuildServiceProvider(sp SPConfig, idp *IdP) (*cs.ServiceProvider, error) {
	if sp.EntityID == "" || sp.AssertionConsumerServiceURL == "" {
		return nil, errors.New("saml: SP entity id and ACS URL are required")
	}
	if sp.PrivateKey == nil || sp.Certificate == nil {
		return nil, errors.New("saml: SP private key + certificate are required")
	}
	if idp == nil || idp.Certificate == nil {
		return nil, errors.New("saml: IdP must include a certificate")
	}
	acsURL, err := url.Parse(sp.AssertionConsumerServiceURL)
	if err != nil {
		return nil, fmt.Errorf("saml: parse ACS URL: %w", err)
	}
	ssoURL, err := url.Parse(idp.SSOURL)
	if err != nil {
		return nil, fmt.Errorf("saml: parse SSO URL: %w", err)
	}

	// IdP metadata is built in-memory from the persisted cert +
	// SSO URL — we don't fetch IdP metadata over the network at
	// request time.
	idpMetadata := &cs.EntityDescriptor{
		EntityID: idp.EntityID,
		IDPSSODescriptors: []cs.IDPSSODescriptor{{
			SSODescriptor: cs.SSODescriptor{
				RoleDescriptor: cs.RoleDescriptor{
					KeyDescriptors: []cs.KeyDescriptor{{
						Use: "signing",
						KeyInfo: cs.KeyInfo{
							X509Data: cs.X509Data{
								X509Certificates: []cs.X509Certificate{
									{Data: base64NoPEM(idp.Certificate.Raw)},
								},
							},
						},
					}},
				},
			},
			SingleSignOnServices: []cs.Endpoint{{
				Binding:  cs.HTTPRedirectBinding,
				Location: ssoURL.String(),
			}},
		}},
	}

	return &cs.ServiceProvider{
		EntityID:          sp.EntityID,
		Key:               sp.PrivateKey,
		Certificate:       sp.Certificate,
		AcsURL:            *acsURL,
		IDPMetadata:       idpMetadata,
		AllowIDPInitiated: false,
	}, nil
}

// ParseCertificate decodes a PEM-encoded x509 certificate.
func ParseCertificate(pemBytes []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("saml: not a PEM-encoded certificate")
	}
	return x509.ParseCertificate(block.Bytes)
}

// EncodeCertificate is the inverse — builds a PEM block from an
// x509.Certificate's raw DER bytes.
func EncodeCertificate(cert *x509.Certificate) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
}

// base64NoPEM returns the base64 form expected by the
// X509Certificate XML element (no PEM headers, no newlines).
func base64NoPEM(der []byte) string {
	return base64Std.EncodeToString(der)
}

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrIdPNotFound is returned when no row matches the supplied
// entity_id / id (or the row is disabled).
var ErrIdPNotFound = errors.New("saml: idp not found")
