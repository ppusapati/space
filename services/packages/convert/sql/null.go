// Package sql provides conversions between SQL nullable types and Go types.
package sql

import (
	"database/sql"
	"time"
)

// ==================== String Conversions ====================

// NullStringFromPtr converts *string to sql.NullString.
// Returns an invalid NullString if the pointer is nil.
func NullStringFromPtr(s *string) sql.NullString {
	if s != nil {
		return sql.NullString{String: *s, Valid: true}
	}
	return sql.NullString{Valid: false}
}

// StringPtrFromNullString converts sql.NullString to *string.
// Returns nil if the NullString is not valid.
func StringPtrFromNullString(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

// NullStringFrom creates a valid NullString from a string.
func NullStringFrom(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

// StringFromNullString converts sql.NullString to string.
// Returns empty string if the NullString is not valid.
func StringFromNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// StringFromNullStringOr converts sql.NullString to string with a default.
func StringFromNullStringOr(ns sql.NullString, defaultVal string) string {
	if ns.Valid {
		return ns.String
	}
	return defaultVal
}

// ==================== Int64 Conversions ====================

// NullInt64FromPtr converts *int64 to sql.NullInt64.
// Returns an invalid NullInt64 if the pointer is nil.
func NullInt64FromPtr(i *int64) sql.NullInt64 {
	if i != nil {
		return sql.NullInt64{Int64: *i, Valid: true}
	}
	return sql.NullInt64{Valid: false}
}

// Int64PtrFromNullInt64 converts sql.NullInt64 to *int64.
// Returns nil if the NullInt64 is not valid.
func Int64PtrFromNullInt64(ni sql.NullInt64) *int64 {
	if ni.Valid {
		return &ni.Int64
	}
	return nil
}

// NullInt64From creates a valid NullInt64 from an int64.
func NullInt64From(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: i, Valid: true}
}

// Int64FromNullInt64 converts sql.NullInt64 to int64.
// Returns 0 if the NullInt64 is not valid.
func Int64FromNullInt64(ni sql.NullInt64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return 0
}

// Int64FromNullInt64Or converts sql.NullInt64 to int64 with a default.
func Int64FromNullInt64Or(ni sql.NullInt64, defaultVal int64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return defaultVal
}

// ==================== Int32 Conversions ====================

// NullInt32FromPtr converts *int32 to sql.NullInt32.
// Returns an invalid NullInt32 if the pointer is nil.
func NullInt32FromPtr(i *int32) sql.NullInt32 {
	if i != nil {
		return sql.NullInt32{Int32: *i, Valid: true}
	}
	return sql.NullInt32{Valid: false}
}

// Int32PtrFromNullInt32 converts sql.NullInt32 to *int32.
// Returns nil if the NullInt32 is not valid.
func Int32PtrFromNullInt32(ni sql.NullInt32) *int32 {
	if ni.Valid {
		return &ni.Int32
	}
	return nil
}

// NullInt32From creates a valid NullInt32 from an int32.
func NullInt32From(i int32) sql.NullInt32 {
	return sql.NullInt32{Int32: i, Valid: true}
}

// Int32FromNullInt32 converts sql.NullInt32 to int32.
// Returns 0 if the NullInt32 is not valid.
func Int32FromNullInt32(ni sql.NullInt32) int32 {
	if ni.Valid {
		return ni.Int32
	}
	return 0
}

// Int32FromNullInt32Or converts sql.NullInt32 to int32 with a default.
func Int32FromNullInt32Or(ni sql.NullInt32, defaultVal int32) int32 {
	if ni.Valid {
		return ni.Int32
	}
	return defaultVal
}

// ==================== Int16 Conversions ====================

// NullInt16FromPtr converts *int16 to sql.NullInt16.
// Returns an invalid NullInt16 if the pointer is nil.
func NullInt16FromPtr(i *int16) sql.NullInt16 {
	if i != nil {
		return sql.NullInt16{Int16: *i, Valid: true}
	}
	return sql.NullInt16{Valid: false}
}

// Int16PtrFromNullInt16 converts sql.NullInt16 to *int16.
// Returns nil if the NullInt16 is not valid.
func Int16PtrFromNullInt16(ni sql.NullInt16) *int16 {
	if ni.Valid {
		return &ni.Int16
	}
	return nil
}

// NullInt16From creates a valid NullInt16 from an int16.
func NullInt16From(i int16) sql.NullInt16 {
	return sql.NullInt16{Int16: i, Valid: true}
}

// Int16FromNullInt16 converts sql.NullInt16 to int16.
// Returns 0 if the NullInt16 is not valid.
func Int16FromNullInt16(ni sql.NullInt16) int16 {
	if ni.Valid {
		return ni.Int16
	}
	return 0
}

// ==================== Bool Conversions ====================

// NullBoolFromPtr converts *bool to sql.NullBool.
// Returns an invalid NullBool if the pointer is nil.
func NullBoolFromPtr(b *bool) sql.NullBool {
	if b != nil {
		return sql.NullBool{Bool: *b, Valid: true}
	}
	return sql.NullBool{Valid: false}
}

// BoolPtrFromNullBool converts sql.NullBool to *bool.
// Returns nil if the NullBool is not valid.
func BoolPtrFromNullBool(nb sql.NullBool) *bool {
	if nb.Valid {
		return &nb.Bool
	}
	return nil
}

// NullBoolFrom creates a valid NullBool from a bool.
func NullBoolFrom(b bool) sql.NullBool {
	return sql.NullBool{Bool: b, Valid: true}
}

// BoolFromNullBool converts sql.NullBool to bool.
// Returns false if the NullBool is not valid.
func BoolFromNullBool(nb sql.NullBool) bool {
	if nb.Valid {
		return nb.Bool
	}
	return false
}

// BoolFromNullBoolOr converts sql.NullBool to bool with a default.
func BoolFromNullBoolOr(nb sql.NullBool, defaultVal bool) bool {
	if nb.Valid {
		return nb.Bool
	}
	return defaultVal
}

// ==================== Float64 Conversions ====================

// NullFloat64FromPtr converts *float64 to sql.NullFloat64.
// Returns an invalid NullFloat64 if the pointer is nil.
func NullFloat64FromPtr(f *float64) sql.NullFloat64 {
	if f != nil {
		return sql.NullFloat64{Float64: *f, Valid: true}
	}
	return sql.NullFloat64{Valid: false}
}

// Float64PtrFromNullFloat64 converts sql.NullFloat64 to *float64.
// Returns nil if the NullFloat64 is not valid.
func Float64PtrFromNullFloat64(nf sql.NullFloat64) *float64 {
	if nf.Valid {
		return &nf.Float64
	}
	return nil
}

// NullFloat64From creates a valid NullFloat64 from a float64.
func NullFloat64From(f float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: f, Valid: true}
}

// Float64FromNullFloat64 converts sql.NullFloat64 to float64.
// Returns 0 if the NullFloat64 is not valid.
func Float64FromNullFloat64(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0
}

// Float64FromNullFloat64Or converts sql.NullFloat64 to float64 with a default.
func Float64FromNullFloat64Or(nf sql.NullFloat64, defaultVal float64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return defaultVal
}

// ==================== Time Conversions ====================

// NullTimeFromPtr converts *time.Time to sql.NullTime.
// Returns an invalid NullTime if the pointer is nil.
func NullTimeFromPtr(t *time.Time) sql.NullTime {
	if t != nil {
		return sql.NullTime{Time: *t, Valid: true}
	}
	return sql.NullTime{Valid: false}
}

// TimePtrFromNullTime converts sql.NullTime to *time.Time.
// Returns nil if the NullTime is not valid.
func TimePtrFromNullTime(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

// NullTimeFrom creates a valid NullTime from a time.Time.
func NullTimeFrom(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: true}
}

// TimeFromNullTime converts sql.NullTime to time.Time.
// Returns zero time if the NullTime is not valid.
func TimeFromNullTime(nt sql.NullTime) time.Time {
	if nt.Valid {
		return nt.Time
	}
	return time.Time{}
}

// TimeFromNullTimeOr converts sql.NullTime to time.Time with a default.
func TimeFromNullTimeOr(nt sql.NullTime, defaultVal time.Time) time.Time {
	if nt.Valid {
		return nt.Time
	}
	return defaultVal
}

// ==================== Byte Conversions ====================

// NullByteFromPtr converts *byte to sql.NullByte.
// Returns an invalid NullByte if the pointer is nil.
func NullByteFromPtr(b *byte) sql.NullByte {
	if b != nil {
		return sql.NullByte{Byte: *b, Valid: true}
	}
	return sql.NullByte{Valid: false}
}

// BytePtrFromNullByte converts sql.NullByte to *byte.
// Returns nil if the NullByte is not valid.
func BytePtrFromNullByte(nb sql.NullByte) *byte {
	if nb.Valid {
		return &nb.Byte
	}
	return nil
}

// NullByteFrom creates a valid NullByte from a byte.
func NullByteFrom(b byte) sql.NullByte {
	return sql.NullByte{Byte: b, Valid: true}
}

// ByteFromNullByte converts sql.NullByte to byte.
// Returns 0 if the NullByte is not valid.
func ByteFromNullByte(nb sql.NullByte) byte {
	if nb.Valid {
		return nb.Byte
	}
	return 0
}
