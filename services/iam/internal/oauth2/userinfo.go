// userinfo.go — /oauth2/userinfo handler.
//
// Per OIDC core §5.3 the UserInfo endpoint returns claims about
// the principal carried in the bearer access token. We project a
// minimal set; richer attribute resolution (organisation,
// preferred_username, etc.) lands when the user-attributes table
// (TASK-P1-IAM-USER-ATTRS) ships.

package oauth2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	authzv1 "p9e.in/chetana/packages/authz/v1"
)

// UserInfoResponse is the JSON envelope returned from /userinfo.
// Field set chosen to satisfy the OIDC standard claims
// `sub`/`email` plus the chetana-specific identity bits a client
// commonly needs without a follow-up call.
type UserInfoResponse struct {
	Subject        string   `json:"sub"`
	TenantID       string   `json:"tenant_id,omitempty"`
	IsUSPerson     bool     `json:"is_us_person,omitempty"`
	ClearanceLevel string   `json:"clearance_level,omitempty"`
	Nationality    string   `json:"nationality,omitempty"`
	Roles          []string `json:"roles,omitempty"`
	Scopes         []string `json:"scopes,omitempty"`
	SessionID      string   `json:"session_id,omitempty"`
}

// UserInfoHandler verifies the bearer access token and projects
// its claims into a UserInfoResponse. Construct with
// NewUserInfoHandler.
type UserInfoHandler struct {
	verifier *authzv1.Verifier
}

// NewUserInfoHandler builds the handler over the supplied access-
// token verifier. Reusing the cross-service verifier guarantees
// /userinfo enforces the same iss/aud/exp policy every service
// interceptor enforces (REQ-CONST-011).
func NewUserInfoHandler(verifier *authzv1.Verifier) *UserInfoHandler {
	return &UserInfoHandler{verifier: verifier}
}

// Handle executes the verify + project flow for an incoming
// Authorization-bearer request. Returns the response body and any
// non-nil error so the HTTP wrapper can emit the right status.
func (h *UserInfoHandler) Handle(ctx context.Context, authHeader string) (*UserInfoResponse, error) {
	token := bearerToken(authHeader)
	if token == "" {
		return nil, ErrUserInfoMissingToken
	}
	p, err := h.verifier.VerifyAccessToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUserInfoInvalidToken, err)
	}
	return &UserInfoResponse{
		Subject:        p.UserID,
		TenantID:       p.TenantID,
		IsUSPerson:     p.IsUSPerson,
		ClearanceLevel: p.ClearanceLevel,
		Nationality:    p.Nationality,
		Roles:          p.Roles,
		Scopes:         p.Scopes,
		SessionID:      p.SessionID,
	}, nil
}

// ServeHTTP makes the handler usable directly off a mux. The
// Connect-RPC variant lands once iam.proto regenerates with the
// userinfo RPC (still gated by OQ-004).
func (h *UserInfoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := h.Handle(r.Context(), r.Header.Get("Authorization"))
	if err != nil {
		switch {
		case errors.Is(err, ErrUserInfoMissingToken):
			w.Header().Set("WWW-Authenticate", `Bearer realm="chetana"`)
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
		case errors.Is(err, ErrUserInfoInvalidToken):
			w.Header().Set("WWW-Authenticate",
				`Bearer realm="chetana", error="invalid_token"`)
			http.Error(w, "invalid token", http.StatusUnauthorized)
		default:
			http.Error(w, "userinfo error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(body)
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(header[len(prefix):])
}

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrUserInfoMissingToken is returned when the request did not
// carry a Bearer Authorization header.
var ErrUserInfoMissingToken = errors.New("oauth2: userinfo missing bearer token")

// ErrUserInfoInvalidToken is returned when the bearer token
// failed verification.
var ErrUserInfoInvalidToken = errors.New("oauth2: userinfo invalid token")
