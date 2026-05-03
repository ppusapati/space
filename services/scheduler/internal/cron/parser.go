// Package cron wraps github.com/robfig/cron/v3 with chetana-side
// validation + next-tick computation.
//
// Why robfig/cron rather than rolling our own:
//
//   • Battle-tested against the canonical Vixie + RFC-style
//     spec; supports both 5-field POSIX and 6-field-with-seconds
//     extensions.
//
//   • The Specification interface is exactly the
//     `Next(time.Time) time.Time` shape the runner needs.
//
// The chetana scheduler uses the standard 5-field syntax
// (`minute hour day-of-month month day-of-week`); the optional
// seconds field is intentionally NOT enabled — sub-minute
// scheduling is overkill for our workload + would let a
// misconfigured job fire 60× faster than intended.

package cron

import (
	"errors"
	"fmt"
	"strings"
	"time"

	cron3 "github.com/robfig/cron/v3"
)

// Schedule wraps a parsed cron expression.
type Schedule struct {
	expr string
	loc  *time.Location
	spec cron3.Schedule
}

// Parse validates `expr` (5-field cron) under `tz` ("" → UTC).
// Returns ErrInvalidSchedule on any parse failure with a
// descriptive message.
func Parse(expr, tz string) (*Schedule, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, ErrInvalidSchedule
	}
	loc := time.UTC
	if tz != "" && tz != "UTC" {
		l, err := time.LoadLocation(tz)
		if err != nil {
			return nil, fmt.Errorf("%w: timezone %q: %v", ErrInvalidSchedule, tz, err)
		}
		loc = l
	}
	parser := cron3.NewParser(cron3.Minute | cron3.Hour | cron3.Dom | cron3.Month | cron3.Dow)
	spec, err := parser.Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSchedule, err)
	}
	return &Schedule{expr: expr, loc: loc, spec: spec}, nil
}

// Next returns the next scheduled instant strictly AFTER `from`.
func (s *Schedule) Next(from time.Time) time.Time {
	return s.spec.Next(from.In(s.loc))
}

// Expression returns the canonical text form of the schedule.
func (s *Schedule) Expression() string { return s.expr }

// Timezone returns the IANA tz name backing the schedule.
func (s *Schedule) Timezone() string { return s.loc.String() }

// ErrInvalidSchedule is returned by Parse when the expression
// (or timezone) cannot be parsed.
var ErrInvalidSchedule = errors.New("cron: invalid schedule expression")
