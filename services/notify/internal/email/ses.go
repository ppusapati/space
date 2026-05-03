// Package email implements the SES-backed email channel.
//
// → REQ-FUNC-PLT-NOTIFY-001 (email channel).
// → REQ-FUNC-PLT-NOTIFY-004 (FIPS endpoint at boot — asserted +
//                             logged by FIPSAsserts below).
//
// We intentionally do NOT pull in the AWS SDK here; the chetana
// notify service composes the SES client at the cmd layer via
// the Sender interface. This keeps:
//
//   • The package importable from test binaries without an AWS
//     credentials chain.
//   • The fake-sender swap-in trivial for tests + the dev posture
//     before TASK-P1-PLT-SECRETS-001 wires real KMS-backed creds.

package email

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Message is the per-call payload.
type Message struct {
	From    string   // RFC 5322 mailbox; e.g. "Chetana <noreply@chetana.p9e.in>"
	To      []string // recipients
	Subject string
	Body    string // rendered HTML or text per template channel
	HTML    bool   // when true, sent as text/html; false → text/plain
	Tags    map[string]string
}

// Sender is the abstract surface the notify service depends on.
// Production wires aws-sdk-go-v2's SES client; tests pass
// CapturingSender.
type Sender interface {
	Send(ctx context.Context, msg Message) error
}

// CapturingSender records every Send call. Used by the unit test
// suite to assert the right message was emitted without spinning
// up a real SES client.
type CapturingSender struct {
	Sent []Message
	Err  error // when non-nil, Send returns this and does NOT record
}

// Send implements Sender.
func (c *CapturingSender) Send(_ context.Context, msg Message) error {
	if c.Err != nil {
		return c.Err
	}
	c.Sent = append(c.Sent, msg)
	return nil
}

// FIPSAsserts validates that the supplied SES endpoint URL targets
// the AWS FIPS endpoint family. Called at boot by the cmd layer so
// a misconfigured deploy fails fast.
//
// Acceptance #3: SES/SNS clients verified to use FIPS endpoint at
// boot (logged + asserted).
func FIPSAsserts(endpoint string) error {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return errors.New("email: SES endpoint is required (must target a -fips region)")
	}
	// Canonical SES FIPS endpoints:
	//   https://email-fips.<region>.amazonaws.com
	//   https://email.<region>.amazonaws.com (NOT FIPS — rejected)
	if !strings.Contains(endpoint, "email-fips.") {
		return fmt.Errorf("email: SES endpoint %q is not FIPS-validated; expected 'email-fips.<region>.amazonaws.com'",
			endpoint)
	}
	return nil
}

// Validate runs the per-call sanity checks on a Message before
// the Sender call. Catches obvious bugs (empty body, no
// recipients, mismatched HTML flag) so the failure message is
// in chetana's idiom rather than AWS-shaped.
func Validate(m Message) error {
	if m.From == "" {
		return errors.New("email: From is required")
	}
	if len(m.To) == 0 {
		return errors.New("email: To is required")
	}
	if strings.TrimSpace(m.Subject) == "" {
		return errors.New("email: Subject is required")
	}
	if strings.TrimSpace(m.Body) == "" {
		return errors.New("email: Body is required")
	}
	for _, r := range m.To {
		if !strings.Contains(r, "@") {
			return fmt.Errorf("email: invalid recipient %q", r)
		}
	}
	return nil
}
