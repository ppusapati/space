// Package sms implements the SNS-backed SMS channel.
//
// → REQ-FUNC-PLT-NOTIFY-001 (SMS channel).
// → REQ-FUNC-PLT-NOTIFY-002 (5/h/user limit — enforced via the
//                             limiter package; the channel itself
//                             does not own the cap).
// → REQ-FUNC-PLT-NOTIFY-004 (FIPS endpoint at boot).

package sms

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Message is the per-call payload.
type Message struct {
	To   string // E.164, e.g. "+15551234567"
	Body string
}

// Sender is the abstract SMS surface. Production wires
// aws-sdk-go-v2's SNS client; tests pass CapturingSender.
type Sender interface {
	Send(ctx context.Context, msg Message) error
}

// CapturingSender records every Send call.
type CapturingSender struct {
	Sent []Message
	Err  error
}

// Send implements Sender.
func (c *CapturingSender) Send(_ context.Context, msg Message) error {
	if c.Err != nil {
		return c.Err
	}
	c.Sent = append(c.Sent, msg)
	return nil
}

// FIPSAsserts validates that the supplied SNS endpoint URL
// targets the AWS FIPS endpoint family.
//
// Canonical SNS FIPS endpoint:
//   https://sns-fips.<region>.amazonaws.com
func FIPSAsserts(endpoint string) error {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return errors.New("sms: SNS endpoint is required (must target a -fips region)")
	}
	if !strings.Contains(endpoint, "sns-fips.") {
		return fmt.Errorf("sms: SNS endpoint %q is not FIPS-validated; expected 'sns-fips.<region>.amazonaws.com'",
			endpoint)
	}
	return nil
}

// Validate runs the per-call sanity checks.
func Validate(m Message) error {
	if !strings.HasPrefix(m.To, "+") {
		return fmt.Errorf("sms: To must be E.164 (got %q)", m.To)
	}
	if len(m.To) < 8 || len(m.To) > 16 {
		return fmt.Errorf("sms: To length out of E.164 bounds (got %q)", m.To)
	}
	if strings.TrimSpace(m.Body) == "" {
		return errors.New("sms: Body is required")
	}
	if len(m.Body) > 1600 {
		// SNS hard cap is 1600 UCS-2 chars per message; longer
		// payloads are silently split which corrupts our audit.
		return errors.New("sms: Body exceeds 1600 char SNS limit")
	}
	return nil
}
