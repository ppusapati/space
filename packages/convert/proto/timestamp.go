package proto

import (
	"database/sql"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ==================== Time ↔ Timestamp Conversions ====================

// TimestampFromTime converts time.Time to *timestamppb.Timestamp.
func TimestampFromTime(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

// TimeFromTimestamp converts *timestamppb.Timestamp to time.Time.
// Returns zero time if the Timestamp is nil.
func TimeFromTimestamp(ts *timestamppb.Timestamp) time.Time {
	if ts != nil {
		return ts.AsTime()
	}
	return time.Time{}
}

// ==================== *Time ↔ Timestamp Conversions ====================

// TimestampFromTimePtr converts *time.Time to *timestamppb.Timestamp.
// Returns nil if the pointer is nil.
func TimestampFromTimePtr(t *time.Time) *timestamppb.Timestamp {
	if t != nil {
		return timestamppb.New(*t)
	}
	return nil
}

// TimePtrFromTimestamp converts *timestamppb.Timestamp to *time.Time.
// Returns nil if the Timestamp is nil.
func TimePtrFromTimestamp(ts *timestamppb.Timestamp) *time.Time {
	if ts != nil {
		t := ts.AsTime()
		return &t
	}
	return nil
}

// ==================== NullTime ↔ Timestamp Conversions ====================

// TimestampFromNullTime converts sql.NullTime to *timestamppb.Timestamp.
// Returns nil if the NullTime is not valid.
func TimestampFromNullTime(nt sql.NullTime) *timestamppb.Timestamp {
	if nt.Valid {
		return timestamppb.New(nt.Time)
	}
	return nil
}

// NullTimeFromTimestamp converts *timestamppb.Timestamp to sql.NullTime.
// Returns an invalid NullTime if the Timestamp is nil.
func NullTimeFromTimestamp(ts *timestamppb.Timestamp) sql.NullTime {
	if ts != nil {
		return sql.NullTime{Time: ts.AsTime(), Valid: true}
	}
	return sql.NullTime{Valid: false}
}

// ==================== Validation Helpers ====================

// IsValidTimestamp checks if a timestamp is valid (not nil and has valid time).
func IsValidTimestamp(ts *timestamppb.Timestamp) bool {
	if ts == nil {
		return false
	}
	return ts.IsValid()
}

// TimestampNow returns the current time as a Timestamp.
func TimestampNow() *timestamppb.Timestamp {
	return timestamppb.Now()
}


// TimestampFromTimeValue converts a time.Time value to *timestamppb.Timestamp.
// Alias for TimestampFromTime for compatibility.
func TimestampFromTimeValue(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

// TimePtrToTimestamp converts *time.Time to *timestamppb.Timestamp.
// Alias for TimestampFromTimePtr for compatibility.
func TimePtrToTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t != nil {
		return timestamppb.New(*t)
	}
	return nil
}

// StringPtrFromString converts a string to *string. Returns nil for empty string.
func StringPtrFromString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// StringFromStringPtr converts *string to string. Returns empty string if nil.
func StringFromStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// TimeToTimestamp converts time.Time to *timestamppb.Timestamp. Alias for TimestampFromTime.
func TimeToTimestamp(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

// FromTime converts time.Time to *timestamppb.Timestamp. Alias for TimestampFromTime.
func FromTime(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

// TimeToProto converts time.Time to *timestamppb.Timestamp. Alias for TimestampFromTime.
func TimeToProto(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}
