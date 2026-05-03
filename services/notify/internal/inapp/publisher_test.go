package inapp

import (
	"testing"
)

func TestValidate(t *testing.T) {
	good := Message{UserID: "u", Title: "t", Body: "b"}
	if err := Validate(good); err != nil {
		t.Errorf("good: %v", err)
	}
	cases := []struct {
		name string
		mut  func(*Message)
	}{
		{"empty user", func(m *Message) { m.UserID = "" }},
		{"empty title", func(m *Message) { m.Title = "  " }},
		{"empty body", func(m *Message) { m.Body = "" }},
		{"unknown severity", func(m *Message) { m.Severity = "fatal" }},
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

func TestValidate_AcceptedSeverities(t *testing.T) {
	for _, s := range []string{"", "info", "warn", "critical"} {
		msg := Message{UserID: "u", Title: "t", Body: "b", Severity: s}
		if err := Validate(msg); err != nil {
			t.Errorf("severity %q: %v", s, err)
		}
	}
}

func TestCapturingPublisher(t *testing.T) {
	c := &CapturingPublisher{}
	if err := c.Publish(nil, Message{UserID: "u"}); err != nil {
		t.Errorf("err: %v", err)
	}
	if len(c.Sent) != 1 {
		t.Errorf("recorded: %d", len(c.Sent))
	}
}

func TestTopic_Constant(t *testing.T) {
	if Topic != "notify.inapp.v1" {
		t.Errorf("topic name regression: %q", Topic)
	}
}
