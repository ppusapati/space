package sms

import (
	"strings"
	"testing"
)

func TestFIPSAsserts(t *testing.T) {
	if err := FIPSAsserts("https://sns-fips.us-east-1.amazonaws.com"); err != nil {
		t.Errorf("good fips: %v", err)
	}
	bad := []string{
		"",
		"https://sns.us-east-1.amazonaws.com",
		"https://example.com",
	}
	for _, b := range bad {
		if err := FIPSAsserts(b); err == nil {
			t.Errorf("FIPSAsserts(%q): expected error", b)
		}
	}
}

func TestValidate(t *testing.T) {
	good := Message{To: "+15551234567", Body: "hi"}
	if err := Validate(good); err != nil {
		t.Errorf("good: %v", err)
	}
	cases := []struct {
		name string
		mut  func(*Message)
	}{
		{"missing plus", func(m *Message) { m.To = "15551234567" }},
		{"too short", func(m *Message) { m.To = "+1234" }},
		{"too long", func(m *Message) { m.To = "+12345678901234567" }},
		{"empty body", func(m *Message) { m.Body = "  " }},
		{"oversize body", func(m *Message) { m.Body = strings.Repeat("x", 1601) }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			msg := good
			c.mut(&msg)
			if err := Validate(msg); err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestCapturingSender(t *testing.T) {
	c := &CapturingSender{}
	if err := c.Send(nil, Message{To: "+1", Body: "x"}); err != nil {
		t.Errorf("err: %v", err)
	}
	if len(c.Sent) != 1 {
		t.Errorf("recorded: %d", len(c.Sent))
	}
}
