// adapters.go — small cross-package adapters cmd/iam composes
// (RETROFIT-001 wiring).
//
// Each adapter bridges a chetana shipped subsystem to one of the
// per-subsystem callbacks the consumer's HandlerConfig accepts.
// They live in cmd/iam (not a shared internal/adapters package)
// because the bridge is the cmd-layer's responsibility — the
// consumer + producer packages stay one-way.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ppusapati/space/services/iam/internal/gdpr"
	"github.com/ppusapati/space/services/iam/internal/login"
	"github.com/ppusapati/space/services/iam/internal/oauth2"
	"github.com/ppusapati/space/services/iam/internal/session"
	"github.com/ppusapati/space/services/iam/internal/token"
	"github.com/ppusapati/space/services/iam/internal/webauthn"
)

// ----------------------------------------------------------------------
// tokenAdapter — bridges token.LoginIssuer to login.TokenIssuer.
// ----------------------------------------------------------------------

type tokenAdapter struct{ inner *token.LoginIssuer }

func (a *tokenAdapter) IssueLoginTokens(ctx context.Context, in login.TokenIssueInput) (login.TokenIssueOutput, error) {
	out, err := a.inner.IssueLoginTokens(ctx, token.LoginIssueInput{
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
		return login.TokenIssueOutput{}, err
	}
	return login.TokenIssueOutput{
		AccessToken:         out.AccessToken,
		AccessTokenExpires:  out.AccessTokenExpires,
		RefreshToken:        out.RefreshToken,
		RefreshTokenExpires: out.RefreshTokenExpires,
	}, nil
}

// ----------------------------------------------------------------------
// sessionAdapter — bridges session.Manager to login.SessionCreator
// AND oauth2.SessionCreator (D1 + D2). Both consumers expect the
// same shape; this single adapter satisfies both.
// ----------------------------------------------------------------------

type sessionAdapter struct{ inner *session.Manager }

func (a *sessionAdapter) Create(ctx context.Context, in login.SessionCreateInput) (login.SessionCreateOutput, error) {
	created, err := a.inner.Create(ctx, session.CreateInput{
		UserID:             in.UserID,
		TenantID:           in.TenantID,
		ClientIP:           in.ClientIP,
		UserAgent:          in.UserAgent,
		AMR:                in.AMR,
		DataClassification: in.DataClassification,
	})
	if err != nil {
		return login.SessionCreateOutput{}, err
	}
	return login.SessionCreateOutput{SessionID: created.SessionID}, nil
}

// oauth2SessionAdapter is the same adapter shape under
// oauth2.SessionCreator's slightly-narrower signature (returns
// only the sessionID string).
type oauth2SessionAdapter struct{ inner *session.Manager }

func (a *oauth2SessionAdapter) Create(ctx context.Context, in oauth2.SessionCreateInput) (string, error) {
	created, err := a.inner.Create(ctx, session.CreateInput{
		UserID:             in.UserID,
		TenantID:           in.TenantID,
		ClientIP:           in.ClientIP,
		UserAgent:          in.UserAgent,
		AMR:                in.AMR,
		DataClassification: in.DataClassification,
	})
	if err != nil {
		return "", err
	}
	return created.SessionID, nil
}

// ----------------------------------------------------------------------
// auditAdapter — bridges the chetana audit-svc HTTP route to the
// per-package audit-emitter interfaces (login.AuditEmitter +
// webauthn.AuditEmitter). RETROFIT-001 A1 + A8.
//
// The cmd-layer wires this against the audit-svc's
// /v1/audit/append POST route. When the audit URL is unset (dev
// posture), the adapter falls back to logging only — matching
// the chetana ALLOW-LISTED stub policy.
// ----------------------------------------------------------------------

type auditAdapter struct {
	url    string
	tenant string
	client *http.Client
	logger *slog.Logger
}

func newAuditAdapter(auditURL, tenantID string, logger *slog.Logger) *auditAdapter {
	return &auditAdapter{
		url:    strings.TrimRight(auditURL, "/"),
		tenant: tenantID,
		client: &http.Client{Timeout: 5 * time.Second},
		logger: logger,
	}
}

// Emit satisfies login.AuditEmitter.
func (a *auditAdapter) Emit(ctx context.Context, e login.Event) error {
	a.send(ctx, map[string]any{
		"tenant_id":         e.TenantID,
		"actor_user_id":     e.UserID,
		"action":            "iam.login." + string(e.Outcome),
		"decision":          loginOutcomeDecision(e.Outcome),
		"actor_client_ip":   e.ClientIP,
		"actor_user_agent":  e.UserAgent,
		"reason":            e.Reason,
		"event_time":        e.OccurredAt.UTC().Format(time.RFC3339Nano),
		"classification":    "cui",
		"metadata":          map[string]string{"email_lower": e.EmailLower},
	})
	return nil
}

// EmitWebAuthn satisfies webauthn.AuditEmitter (RETROFIT-001 A8).
// The webauthn package's AuditEvent shape has different field
// names than login.Event so the adapter projects manually.
func (a *auditAdapter) EmitWebAuthn(ctx context.Context, e webauthn.AuditEvent) error {
	a.send(ctx, map[string]any{
		"tenant_id":      a.tenant,
		"actor_user_id":  e.UserID,
		"action":         "iam.webauthn." + string(e.Outcome),
		"decision":       webauthnOutcomeDecision(e.Outcome),
		"reason":         e.Reason,
		"event_time":     e.OccurredAt.UTC().Format(time.RFC3339Nano),
		"classification": "cui",
		"resource":       e.CredentialID,
	})
	return nil
}

// send is the underlying HTTP POST. Best-effort: an audit failure
// MUST NOT fail the primary flow (REQ-FUNC-PLT-AUDIT-006 says the
// chain is the source of truth, but a chain stall is recovered
// out-of-band).
func (a *auditAdapter) send(ctx context.Context, payload map[string]any) {
	if a.url == "" {
		// Allowlisted dev-posture fallback — logged only.
		a.logger.Debug("audit emit (audit-svc not wired)", slog.Any("event", payload))
		return
	}
	body, _ := jsonMarshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.url+"/v1/audit/append",
		strings.NewReader(string(body)))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		a.logger.Warn("audit emit failed", slog.Any("err", err))
		return
	}
	_ = resp.Body.Close()
}

func loginOutcomeDecision(o login.Outcome) string {
	switch o {
	case login.OutcomeSuccess:
		return "ok"
	case login.OutcomeBadCredentials, login.OutcomeUserNotFound, login.OutcomeUserDisabled:
		return "deny"
	case login.OutcomeLocked, login.OutcomeRateLimited:
		return "deny"
	case login.OutcomeError:
		return "fail"
	}
	return "info"
}

func webauthnOutcomeDecision(o webauthn.AuditOutcome) string {
	switch o {
	case webauthn.OutcomeAssertionOK, webauthn.OutcomeRegistered:
		return "ok"
	case webauthn.OutcomeAssertionFail, webauthn.OutcomeCloneDetected:
		return "deny"
	case webauthn.OutcomeCredentialDisabled:
		return "info"
	}
	return "info"
}

// ----------------------------------------------------------------------
// notifyAdapter — bridges the chetana notify-svc HTTP route to
// reset.Notifier (RETROFIT-001 A2).
// ----------------------------------------------------------------------

type notifyAdapter struct {
	url    string
	client *http.Client
	logger *slog.Logger
}

func newNotifyAdapter(notifyURL string, logger *slog.Logger) *notifyAdapter {
	return &notifyAdapter{
		url:    strings.TrimRight(notifyURL, "/"),
		client: &http.Client{Timeout: 5 * time.Second},
		logger: logger,
	}
}

// SendPasswordReset satisfies reset.Notifier.
func (n *notifyAdapter) SendPasswordReset(ctx context.Context, email, token string, expiresAt time.Time) error {
	if n.url == "" {
		// Allowlisted dev-posture fallback. The reset confirmation
		// link is logged so a developer can copy it for the
		// confirm step.
		n.logger.Warn("notify-svc not wired — password reset link logged only",
			slog.String("email", email),
			slog.String("token_preview", tokenPreview(token)),
			slog.Time("expires_at", expiresAt),
		)
		return nil
	}
	body, _ := jsonMarshal(map[string]any{
		"template_id":    "security.password.reset",
		"channel":        "email",
		"user_email":     email,
		"email_subject":  "Reset your Chetana password",
		"variables": map[string]any{
			"reset_link":  fmt.Sprintf("/reset-password?token=%s", url.QueryEscape(token)),
			"expires_at":  expiresAt.UTC().Format(time.RFC3339),
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		n.url+"/v1/notify/send", strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}

func tokenPreview(t string) string {
	if len(t) <= 12 {
		return t
	}
	return t[:8] + "…"
}

// ----------------------------------------------------------------------
// exporterAdapter — bridges the chetana export-svc HTTP route to
// gdpr.Exporter (RETROFIT-001 A3).
// ----------------------------------------------------------------------

type exporterAdapter struct {
	url    string
	client *http.Client
	logger *slog.Logger
}

func newExporterAdapter(exportURL string, logger *slog.Logger) *exporterAdapter {
	return &exporterAdapter{
		url:    strings.TrimRight(exportURL, "/"),
		client: &http.Client{Timeout: 5 * time.Second},
		logger: logger,
	}
}

// EnqueueSAR satisfies gdpr.Exporter.
func (e *exporterAdapter) EnqueueSAR(ctx context.Context, in gdpr.EnqueueSARInput) (gdpr.JobID, error) {
	if e.url == "" {
		// Allowlisted dev-posture fallback.
		e.logger.Warn("export-svc not wired — SAR enqueue logged only",
			slog.String("user_id", in.UserID),
			slog.String("tenant_id", in.TenantID),
		)
		return gdpr.JobID("dev-stub-" + in.UserID), nil
	}
	body, _ := jsonMarshal(map[string]any{
		"tenant_id":    in.TenantID,
		"requested_by": in.UserID,
		"kind":         "gdpr_sar",
		"payload": map[string]any{
			"user_id":   in.UserID,
			"tenant_id": in.TenantID,
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		e.url+"/v1/export/submit", strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := e.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("export-svc submit: HTTP %d", resp.StatusCode)
	}
	var out struct {
		JobID string `json:"job_id"`
	}
	if err := jsonDecode(resp.Body, &out); err != nil {
		return "", fmt.Errorf("export-svc submit: decode: %w", err)
	}
	return gdpr.JobID(out.JobID), nil
}
