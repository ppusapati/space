package sql

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Type aliases for common pgtype types.
type Numeric = pgtype.Numeric
type Date = pgtype.Date
type Timestamptz = pgtype.Timestamptz

// TimestamptzToTimePtr converts pgtype.Timestamptz to *time.Time.
// Returns nil if the Timestamptz is not valid.
// This is an alias for TimePtrFromTimestamptz.
func TimestamptzToTimePtr(ts pgtype.Timestamptz) *time.Time {
	return TimePtrFromTimestamptz(ts)
}

// TextToStringPtr converts pgtype.Text to *string.
func TextToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

// StringPtrToText converts *string to pgtype.Text.
func StringPtrToText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// TimePtrToDate converts *time.Time to pgtype.Date. Alias for DateFromTimePtr.
func TimePtrToDate(t *time.Time) pgtype.Date {
	return DateFromTimePtr(t)
}

// TimePtrToTimestamptz converts *time.Time to pgtype.Timestamptz. Alias for TimestamptzFromTimePtr.
func TimePtrToTimestamptz(t *time.Time) pgtype.Timestamptz {
	return TimestamptzFromTimePtr(t)
}

// TimeFromTimestamptzPtr converts *pgtype.Timestamptz to time.Time.
func TimeFromTimestamptzPtr(ts *pgtype.Timestamptz) time.Time {
	if ts == nil || !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}

// TimePtrFromTimestamptzPtr converts *pgtype.Timestamptz to *time.Time.
func TimePtrFromTimestamptzPtr(ts *pgtype.Timestamptz) *time.Time {
	if ts == nil || !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}
