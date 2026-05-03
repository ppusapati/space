package email

import (
	"strings"
	"testing"
)

// REQ-FUNC-PLT-NOTIFY-004: SES endpoint MUST be FIPS-validated at boot.
func TestFIPSAsserts_AcceptsFIPSEndpoint(t *testing.T) {
	cases := []string{
		"https://email-fips.us-east-1.amazonaws.com",
		"https://email-fips.us-gov-west-1.amazonaws.com",
	}
	for _, ep := range cases {
		if err := FIPSAsserts(ep); err != nil {
			t.Errorf("FIPSAsserts(%q): %v", ep, err)
		}
	}
}

func TestFIPSAsserts_RejectsNonFIPS(t *testing.T) {
	cases := []string{
		"",
		"https://email.us-east-1.amazonaws.com",
		"https://example.com",
	}
	for _, ep := range cases {
		if err := FIPSAsserts(ep); err == nil {
			t.Errorf("FIPSAsserts(%q): expected error", ep)
		}
	}
}

func TestValidate(t *testing.T) {
	good := Message{
		From:    "Chetana <noreply@example.com>",
		To:      []string{"user@example.com"},
		Subject: "hi",
		Body:    "body",
	}
	if err := Validate(good); err != nil {
		t.Errorf("good msg: %v", err)
	}

	cases := []struct {
		name string
		mut  func(*Message)
	}{
		{"empty from", func(m *Message) { m.From = "" }},
		{"empty to", func(m *Message) { m.To = nil }},
		{"empty subject", func(m *Message) { m.Subject = "  " }},
		{"empty body", func(m *Message) { m.Body = "  " }},
		{"invalid recipient", func(m *Message) { m.To = []string{"not-an-email"} }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			msg := good
			msg.To = append([]string(nil), good.To...)
			c.mut(&msg)
			if err := Validate(msg); err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestCapturingSender_RecordsAndPropagatesError(t *testing.T) {
	c := &CapturingSender{}
	msg := Message{To: []string{"a"}}
	if err := c.Send(nil, msg); err != nil {
		t.Errorf("nil err: %v", err)
	}
	if len(c.Sent) != 1 {
		t.Errorf("recorded: %d", len(c.Sent))
	}

	c2 := &CapturingSender{Err: testErr("boom")}
	if err := c2.Send(nil, msg); err == nil {
		t.Error("expected propagated error")
	}
	if len(c2.Sent) != 0 {
		t.Errorf("must not record on err: %d", len(c2.Sent))
	}
}

type testErr string

func (e testErr) Error() string { return string(e) }

func TestFIPSAsserts_EndpointWithWhitespace(t *testing.T) {
	if err := FIPSAsserts("  https://email-fips.us-east-1.amazonaws.com  "); err != nil {
		t.Errorf("trimmed endpoint: %v", err)
	}
}

func TestValidate_MultipleRecipients(t *testing.T) {
	good := Message{
		From:    "From <a@x>",
		To:      []string{"x@y", "z@y"},
		Subject: "s",
		Body:    "b",
	}
	if err := Validate(good); err != nil {
		t.Errorf("multiple recipients: %v", err)
	}
	bad := good
	bad.To = []string{"x@y", "not-email"}
	if err := Validate(bad); err == nil {
		t.Error("expected error for one bad recipient")
	}
}

func TestFIPSAsserts_DescriptiveError(t *testing.T) {
	err := FIPSAsserts("https://email.us-east-1.amazonaws.com")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "FIPS") {
		t.Errorf("error should mention FIPS: %v", err)
	}
}
