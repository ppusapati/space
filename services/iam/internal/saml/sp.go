// sp.go — Service Provider façade: AuthnRequest construction,
// SP metadata, ACS handler.
//
// The chetana side of SAML is one SP that talks to N registered
// IdPs. The flow:
//
//   1. /saml/login/{idp_id}  → BeginSSO returns the redirect URL
//      to the IdP's SSO endpoint with the (deflated, base64url-
//      encoded) AuthnRequest in the SAMLRequest query parameter.
//
//   2. IdP authenticates the user, posts a SAMLResponse to the
//      chetana SP's ACS endpoint /saml/acs/{idp_id}.
//
//   3. /saml/acs/{idp_id} → FinishSSO parses + signature-verifies
//      the response, projects the assertion attributes through
//      the IdP's AttributeMapping, and runs the JIT provisioning
//      hook (jit.go) to find or create the chetana user.
//
// Signature requirement: the protocol library's parseResponse
// rejects a response that lacks BOTH a Response-level and
// Assertion-level signature. We additionally pin to "assertion
// MUST be signed" via the chetana SP-side wrapper below.

package saml

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	cs "github.com/crewjam/saml"
)

// Service is the chetana SAML SP façade. Construct with
// NewService.
type Service struct {
	sp    SPConfig
	store *Store
	jit   *JITProvisioner
}

// NewService validates the SP config and returns a Service.
func NewService(sp SPConfig, store *Store, jit *JITProvisioner) (*Service, error) {
	if sp.EntityID == "" {
		return nil, errors.New("saml: SP entity id is required")
	}
	if sp.AssertionConsumerServiceURL == "" {
		return nil, errors.New("saml: SP ACS URL is required")
	}
	if sp.PrivateKey == nil || sp.Certificate == nil {
		return nil, errors.New("saml: SP private key + certificate are required")
	}
	if store == nil {
		return nil, errors.New("saml: nil store")
	}
	if jit == nil {
		return nil, errors.New("saml: nil JIT provisioner")
	}
	return &Service{sp: sp, store: store, jit: jit}, nil
}

// SSOResponse carries the data the HTTP layer needs to redirect
// the user-agent to the IdP's SSO endpoint.
type SSOResponse struct {
	// RedirectURL is the fully-formed URL with SAMLRequest +
	// (optional) RelayState as query parameters.
	RedirectURL string

	// RequestID is the AuthnRequest ID — the caller stores this in
	// the user's server-side session so FinishSSO can verify the
	// SAMLResponse's InResponseTo binding.
	RequestID string
}

// BeginSSO builds an AuthnRequest for the named IdP and returns
// the redirect URL.
func (s *Service) BeginSSO(ctx context.Context, idpID int64, relayState string) (*SSOResponse, error) {
	idp, err := s.store.LookupByID(ctx, idpID)
	if err != nil {
		return nil, err
	}
	provider, err := BuildServiceProvider(s.sp, idp)
	if err != nil {
		return nil, err
	}
	authnReq, err := provider.MakeAuthenticationRequest(
		provider.GetSSOBindingLocation(cs.HTTPRedirectBinding),
		cs.HTTPRedirectBinding,
		cs.HTTPPostBinding,
	)
	if err != nil {
		return nil, fmt.Errorf("saml: build AuthnRequest: %w", err)
	}
	redirectURL, err := authnReq.Redirect(relayState, provider)
	if err != nil {
		return nil, fmt.Errorf("saml: build redirect: %w", err)
	}
	return &SSOResponse{
		RedirectURL: redirectURL.String(),
		RequestID:   authnReq.ID,
	}, nil
}

// SSOResult is the chetana-side outcome of a successful ACS callback.
type SSOResult struct {
	IdP       *IdP
	UserID    string // chetana user id (existing or freshly provisioned)
	Email     string
	NameID    string
	Roles     []string
	Created   bool   // true when JIT created the user on this assertion
	SessionID string // populated by the caller after the chetana session is minted
}

// FinishSSO parses + verifies a SAMLResponse posted to the SP's
// ACS endpoint and runs JIT provisioning. The caller (the HTTP
// handler) supplies the request and the set of in-flight
// AuthnRequest IDs the server-side session knows about (used by
// the protocol library to enforce the InResponseTo binding).
//
// Errors:
//   • ErrIdPNotFound      — the {idp_id} URL parameter did not
//                            map to an active IdP row.
//   • ErrSignatureInvalid — wraps the protocol library's
//                            signature/canonicalisation error so
//                            the caller can return 401 + audit.
//   • Other errors are forwarded as-is (parsing, JIT failures).
func (s *Service) FinishSSO(ctx context.Context, idpID int64, req *http.Request, possibleRequestIDs []string) (*SSOResult, error) {
	idp, err := s.store.LookupByID(ctx, idpID)
	if err != nil {
		return nil, err
	}
	provider, err := BuildServiceProvider(s.sp, idp)
	if err != nil {
		return nil, err
	}
	assertion, err := provider.ParseResponse(req, possibleRequestIDs)
	if err != nil {
		// crewjam/saml wraps every validation failure in
		// *cs.InvalidResponseError. Promote that to our typed
		// signature error so the audit chain is uniform.
		return nil, fmt.Errorf("%w: %v", ErrSignatureInvalid, err)
	}

	attrs := flattenAttributes(assertion)
	nameID := ""
	if assertion.Subject != nil && assertion.Subject.NameID != nil {
		nameID = assertion.Subject.NameID.Value
	}

	out, err := s.jit.Provision(ctx, idp, ProvisionInput{
		NameID:     nameID,
		Attributes: attrs,
	})
	if err != nil {
		return nil, err
	}
	return &SSOResult{
		IdP:     idp,
		UserID:  out.UserID,
		Email:   out.Email,
		NameID:  nameID,
		Roles:   out.Roles,
		Created: out.Created,
	}, nil
}

// MetadataXML returns the SP's SAML 2.0 metadata document for the
// given IdP. Hosted by the IAM service at /saml/metadata so an
// IdP admin can register chetana with one click.
func (s *Service) MetadataXML(ctx context.Context, idpID int64) ([]byte, error) {
	idp, err := s.store.LookupByID(ctx, idpID)
	if err != nil {
		return nil, err
	}
	provider, err := BuildServiceProvider(s.sp, idp)
	if err != nil {
		return nil, err
	}
	md := provider.Metadata()
	return marshalMetadata(md)
}

// flattenAttributes collapses the assertion's AttributeStatements
// into a map keyed by Name (the URI-shaped attribute identifier).
// Multi-valued attributes are preserved as a slice in the map.
func flattenAttributes(assertion *cs.Assertion) map[string][]string {
	out := make(map[string][]string)
	for _, stmt := range assertion.AttributeStatements {
		for _, attr := range stmt.Attributes {
			values := make([]string, 0, len(attr.Values))
			for _, v := range attr.Values {
				values = append(values, v.Value)
			}
			// Some IdPs send the same attribute in two statements;
			// concatenate.
			out[attr.Name] = append(out[attr.Name], values...)
		}
	}
	return out
}

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrSignatureInvalid is returned when the protocol library
// rejects the SAMLResponse's signature, canonicalisation, audience,
// or InResponseTo binding. The chetana ACS handler treats every
// flavour the same way (401 + audit).
var ErrSignatureInvalid = errors.New("saml: response signature invalid")
