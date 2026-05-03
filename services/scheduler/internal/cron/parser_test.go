package cron

import (
	"errors"
	"testing"
	"time"
)

func TestParse_HappyPath(t *testing.T) {
	s, err := Parse("*/5 * * * *", "")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if s.Expression() != "*/5 * * * *" {
		t.Errorf("expr: %q", s.Expression())
	}
	from := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	next := s.Next(from)
	if !next.After(from) {
		t.Errorf("next must be after from: %v vs %v", next, from)
	}
	if next.Sub(from) > 6*time.Minute {
		t.Errorf("next gap too large: %v", next.Sub(from))
	}
}

func TestParse_RejectsEmpty(t *testing.T) {
	if _, err := Parse("", ""); !errors.Is(err, ErrInvalidSchedule) {
		t.Errorf("got %v want ErrInvalidSchedule", err)
	}
}

func TestParse_RejectsBadExpression(t *testing.T) {
	cases := []string{
		"not a cron",
		"* * * *",      // 4 fields
		"60 * * * *",   // out of range
		"* */0 * * *",  // illegal step
	}
	for _, c := range cases {
		if _, err := Parse(c, ""); err == nil {
			t.Errorf("Parse(%q): expected error", c)
		}
	}
}

func TestParse_RejectsBadTimezone(t *testing.T) {
	if _, err := Parse("* * * * *", "Not/AReal_Zone"); err == nil {
		t.Error("expected error for bogus tz")
	}
}

func TestParse_AcceptsValidTimezone(t *testing.T) {
	s, err := Parse("0 9 * * *", "America/Los_Angeles")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if s.Timezone() != "America/Los_Angeles" {
		t.Errorf("tz: %q", s.Timezone())
	}
}

func TestNext_Hourly(t *testing.T) {
	s, _ := Parse("0 * * * *", "UTC")
	from := time.Date(2026, 5, 1, 12, 30, 0, 0, time.UTC)
	next := s.Next(from)
	want := time.Date(2026, 5, 1, 13, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Errorf("next: got %v want %v", next, want)
	}
}
