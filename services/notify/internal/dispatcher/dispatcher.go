// Package dispatcher orchestrates a chetana notification:
//
//   1. Lookup the template (versioned, per-channel) in the store.
//   2. Check the user's preferences (mandatory templates skip
//      this step — REQ-FUNC-PLT-NOTIFY-003).
//   3. Render the template with the supplied variables. Missing
//      required variables surface as MissingVariableError so the
//      caller can return 400 with the variable name (acceptance
//      #1).
//   4. For SMS, consult the per-user 5/h limiter
//      (REQ-FUNC-PLT-NOTIFY-002).
//   5. Hand off to the channel-specific Sender / Publisher.

package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ppusapati/space/services/notify/internal/email"
	"github.com/ppusapati/space/services/notify/internal/inapp"
	"github.com/ppusapati/space/services/notify/internal/limiter"
	"github.com/ppusapati/space/services/notify/internal/preferences"
	"github.com/ppusapati/space/services/notify/internal/sms"
	"github.com/ppusapati/space/services/notify/internal/store"
	"github.com/ppusapati/space/services/notify/internal/template"
)

// SendRequest is the single entry-point for every notification.
type SendRequest struct {
	UserID      string
	UserEmail   string // populated for email channel
	UserPhone   string // populated for sms channel (E.164)
	DisplayName string
	TenantID    string
	TemplateID  string
	Channel     string // store.ChannelEmail / SMS / InApp
	Variables   map[string]any

	// EmailSubject is appended to the rendered body when the
	// channel is email. Templates own the body; subjects stay
	// caller-supplied so dynamic prefixes like "[Chetana
	// Critical]" can be added.
	EmailSubject string

	// EmailFrom overrides the dispatcher's default From mailbox
	// when set. Mandatory templates use the platform-wide
	// no-reply address.
	EmailFrom string
}

// Dispatcher wires the four channels + the preferences + limiter.
type Dispatcher struct {
	templates   *store.TemplateStore
	preferences *preferences.Store
	renderer    *template.Renderer
	emailSender email.Sender
	smsSender   sms.Sender
	inAppPub    inapp.Publisher
	smsLimiter  *limiter.SMSLimiter
	defaults    Defaults
	clk         func() time.Time
}

// Defaults contains tenant-level fallbacks the dispatcher applies
// when the per-call request does not override them.
type Defaults struct {
	EmailFrom string // e.g. "Chetana <noreply@chetana.p9e.in>"
}

// Config wires the Dispatcher.
type Config struct {
	Templates   *store.TemplateStore
	Preferences *preferences.Store
	Renderer    *template.Renderer
	Email       email.Sender
	SMS         sms.Sender
	InApp       inapp.Publisher
	SMSLimiter  *limiter.SMSLimiter
	Defaults    Defaults
	Now         func() time.Time
}

// New wires a Dispatcher.
func New(cfg Config) (*Dispatcher, error) {
	if cfg.Templates == nil {
		return nil, errors.New("dispatcher: nil template store")
	}
	if cfg.Preferences == nil {
		return nil, errors.New("dispatcher: nil preferences store")
	}
	if cfg.Renderer == nil {
		cfg.Renderer = template.NewRenderer()
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Dispatcher{
		templates:   cfg.Templates,
		preferences: cfg.Preferences,
		renderer:    cfg.Renderer,
		emailSender: cfg.Email,
		smsSender:   cfg.SMS,
		inAppPub:    cfg.InApp,
		smsLimiter:  cfg.SMSLimiter,
		defaults:    cfg.Defaults,
		clk:         cfg.Now,
	}, nil
}

// Result is the outcome of a Send call.
type Result struct {
	Outcome    Outcome
	TemplateID string
	Channel    string
	Reason     string
}

// Outcome enumerates the dispatcher's outcomes.
type Outcome string

// Canonical outcomes.
const (
	OutcomeSent       Outcome = "sent"
	OutcomeOptedOut   Outcome = "opted_out"
	OutcomeRateLimit  Outcome = "rate_limited"
	OutcomeMissingVar Outcome = "missing_variable"
	OutcomeError      Outcome = "internal_error"
)

// Send runs the orchestration. Returns the outcome envelope plus
// the typed error for the caller to map onto an HTTP / Connect
// status code.
func (d *Dispatcher) Send(ctx context.Context, req SendRequest) (*Result, error) {
	if req.TemplateID == "" || req.Channel == "" {
		return nil, errors.New("dispatcher: TemplateID + Channel required")
	}
	res := &Result{TemplateID: req.TemplateID, Channel: req.Channel}

	tmpl, err := d.templates.LookupActive(ctx, req.TemplateID, req.Channel)
	if err != nil {
		res.Outcome = OutcomeError
		res.Reason = err.Error()
		return res, err
	}

	allowed, err := d.preferences.IsAllowed(ctx, req.UserID, req.TemplateID, tmpl.Mandatory)
	if err != nil {
		res.Outcome = OutcomeError
		res.Reason = err.Error()
		return res, err
	}
	if !allowed {
		res.Outcome = OutcomeOptedOut
		return res, nil
	}

	// Inject the platform-controlled defaults the templates
	// always reference (display_name, occurred_at) so callers
	// don't have to.
	if req.Variables == nil {
		req.Variables = map[string]any{}
	}
	if _, ok := req.Variables["display_name"]; !ok {
		req.Variables["display_name"] = req.DisplayName
	}
	if _, ok := req.Variables["occurred_at"]; !ok {
		req.Variables["occurred_at"] = d.clk().UTC().Format(time.RFC3339)
	}

	body, err := d.renderer.Render(tmpl, req.Variables)
	if err != nil {
		var missing *template.MissingVariableError
		if errors.As(err, &missing) {
			res.Outcome = OutcomeMissingVar
			res.Reason = strings.Join(missing.Missing, ",")
			return res, err
		}
		res.Outcome = OutcomeError
		res.Reason = err.Error()
		return res, err
	}

	switch req.Channel {
	case store.ChannelEmail:
		return d.sendEmail(ctx, req, body, res)
	case store.ChannelSMS:
		return d.sendSMS(ctx, req, body, res)
	case store.ChannelInApp:
		return d.sendInApp(ctx, req, tmpl, body, res)
	default:
		res.Outcome = OutcomeError
		res.Reason = "unknown channel"
		return res, fmt.Errorf("dispatcher: unknown channel %q", req.Channel)
	}
}

func (d *Dispatcher) sendEmail(ctx context.Context, req SendRequest, body string, res *Result) (*Result, error) {
	if d.emailSender == nil {
		res.Outcome = OutcomeError
		res.Reason = "email sender not configured"
		return res, errors.New("dispatcher: email channel requested but sender is nil")
	}
	from := req.EmailFrom
	if from == "" {
		from = d.defaults.EmailFrom
	}
	msg := email.Message{
		From:    from,
		To:      []string{req.UserEmail},
		Subject: req.EmailSubject,
		Body:    body,
	}
	if err := email.Validate(msg); err != nil {
		res.Outcome = OutcomeError
		res.Reason = err.Error()
		return res, err
	}
	if err := d.emailSender.Send(ctx, msg); err != nil {
		res.Outcome = OutcomeError
		res.Reason = err.Error()
		return res, err
	}
	res.Outcome = OutcomeSent
	return res, nil
}

func (d *Dispatcher) sendSMS(ctx context.Context, req SendRequest, body string, res *Result) (*Result, error) {
	if d.smsSender == nil {
		res.Outcome = OutcomeError
		res.Reason = "sms sender not configured"
		return res, errors.New("dispatcher: sms channel requested but sender is nil")
	}
	if d.smsLimiter != nil {
		decision, err := d.smsLimiter.Allow(ctx, req.UserID)
		if err != nil {
			res.Outcome = OutcomeError
			res.Reason = err.Error()
			return res, err
		}
		if !decision.Allowed {
			res.Outcome = OutcomeRateLimit
			res.Reason = fmt.Sprintf("hit %d/%d in window", decision.HitsInWindow, decision.Limit)
			return res, nil
		}
	}
	msg := sms.Message{To: req.UserPhone, Body: body}
	if err := sms.Validate(msg); err != nil {
		res.Outcome = OutcomeError
		res.Reason = err.Error()
		return res, err
	}
	if err := d.smsSender.Send(ctx, msg); err != nil {
		res.Outcome = OutcomeError
		res.Reason = err.Error()
		return res, err
	}
	res.Outcome = OutcomeSent
	return res, nil
}

func (d *Dispatcher) sendInApp(ctx context.Context, req SendRequest, _ *template.Template, body string, res *Result) (*Result, error) {
	if d.inAppPub == nil {
		res.Outcome = OutcomeError
		res.Reason = "inapp publisher not configured"
		return res, errors.New("dispatcher: inapp channel requested but publisher is nil")
	}
	severity, _ := req.Variables["severity"].(string)
	if severity == "" {
		severity = "info"
	}
	title, _ := req.Variables["title"].(string)
	if title == "" {
		title = req.EmailSubject // fall back to the email subject if the caller didn't pass one
	}
	msg := inapp.Message{
		UserID:     req.UserID,
		TenantID:   req.TenantID,
		Title:      title,
		Body:       body,
		Severity:   severity,
		OccurredAt: d.clk().UTC(),
	}
	if err := inapp.Validate(msg); err != nil {
		res.Outcome = OutcomeError
		res.Reason = err.Error()
		return res, err
	}
	if err := d.inAppPub.Publish(ctx, msg); err != nil {
		res.Outcome = OutcomeError
		res.Reason = err.Error()
		return res, err
	}
	res.Outcome = OutcomeSent
	return res, nil
}
