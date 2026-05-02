// login.go — adapter that fits the login handler's TokenIssuer
// interface to this package's Issuer + RefreshStore pair.
//
// Lives here (not in internal/login) so internal/login keeps zero
// dependencies on the token package — that one-way arrow keeps the
// unit tests honest and matches the layered design in plan/todo.md.

package token

import (
	"context"
	"fmt"
	"time"
)

// LoginIssuer combines an access-token Issuer and a RefreshStore. It
// satisfies internal/login.TokenIssuer; cmd/iam wires both halves at
// service boot and hands the assembled value to the login handler.
type LoginIssuer struct {
	access  *Issuer
	refresh *RefreshStore
	clock   func() time.Time
}

// NewLoginIssuer returns an adapter that mints (access, refresh)
// pairs on every successful login.
func NewLoginIssuer(access *Issuer, refresh *RefreshStore, clock func() time.Time) *LoginIssuer {
	if clock == nil {
		clock = time.Now
	}
	return &LoginIssuer{access: access, refresh: refresh, clock: clock}
}

// LoginIssueInput mirrors internal/login.TokenIssueInput. We restate
// it here so the token package does not import internal/login.
type LoginIssueInput struct {
	UserID         string
	TenantID       string
	SessionID      string
	IsUSPerson     bool
	ClearanceLevel string
	Nationality    string
	Roles          []string
	Scopes         []string
	AMR            []string
}

// LoginIssueOutput mirrors internal/login.TokenIssueOutput.
type LoginIssueOutput struct {
	AccessToken         string
	AccessTokenExpires  time.Time
	RefreshToken        string
	RefreshTokenExpires time.Time
}

// IssueLoginTokens mints the access JWT then opens a fresh
// refresh-token family for the session. The two writes are
// independent — if refresh-token persistence fails the access token
// must NOT be returned (otherwise the client gets a JWT with no
// rotation path). We return the error directly so the login handler
// reports 500 and the access JWT is discarded with the call.
func (l *LoginIssuer) IssueLoginTokens(ctx context.Context, in LoginIssueInput) (LoginIssueOutput, error) {
	access, claims, err := l.access.IssueAccessToken(IssueInput{
		UserID:         in.UserID,
		TenantID:       in.TenantID,
		SessionID:      in.SessionID,
		IsUSPerson:     in.IsUSPerson,
		ClearanceLevel: in.ClearanceLevel,
		Nationality:    in.Nationality,
		Roles:          in.Roles,
		Scopes:         in.Scopes,
		AMR:            in.AMR,
	})
	if err != nil {
		return LoginIssueOutput{}, fmt.Errorf("token: issue access: %w", err)
	}

	refresh, err := l.refresh.Issue(ctx, RefreshIssue{
		UserID:    in.UserID,
		TenantID:  in.TenantID,
		SessionID: in.SessionID,
		IssuedAt:  l.clock().UTC(),
	})
	if err != nil {
		return LoginIssueOutput{}, fmt.Errorf("token: issue refresh: %w", err)
	}

	return LoginIssueOutput{
		AccessToken:         access,
		AccessTokenExpires:  claims.ExpiresAt.Time,
		RefreshToken:        refresh.Token,
		RefreshTokenExpires: refresh.ExpiresAt,
	}, nil
}
