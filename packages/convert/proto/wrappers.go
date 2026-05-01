// Package proto provides conversions between protobuf types and Go types.
package proto

import (
	"database/sql"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

// ==================== String Conversions ====================

// StringValueFromNullString converts sql.NullString to *wrapperspb.StringValue.
// Returns nil if the NullString is not valid.
func StringValueFromNullString(s sql.NullString) *wrapperspb.StringValue {
	if s.Valid {
		return &wrapperspb.StringValue{Value: s.String}
	}
	return nil
}

// NullStringFromStringValue converts *wrapperspb.StringValue to sql.NullString.
// Returns an invalid NullString if the StringValue is nil.
func NullStringFromStringValue(sv *wrapperspb.StringValue) sql.NullString {
	if sv != nil {
		return sql.NullString{String: sv.Value, Valid: true}
	}
	return sql.NullString{Valid: false}
}

// StringValueFromPtr converts *string to *wrapperspb.StringValue.
// Returns nil if the pointer is nil.
func StringValueFromPtr(s *string) *wrapperspb.StringValue {
	if s != nil {
		return &wrapperspb.StringValue{Value: *s}
	}
	return nil
}

// StringPtrFromStringValue converts *wrapperspb.StringValue to *string.
// Returns nil if the StringValue is nil.
func StringPtrFromStringValue(sv *wrapperspb.StringValue) *string {
	if sv != nil {
		return &sv.Value
	}
	return nil
}

// StringFromStringValue converts *wrapperspb.StringValue to string.
// Returns empty string if the StringValue is nil.
func StringFromStringValue(sv *wrapperspb.StringValue) string {
	if sv != nil {
		return sv.Value
	}
	return ""
}

// StringValueFrom creates a StringValue from a string.
func StringValueFrom(s string) *wrapperspb.StringValue {
	return &wrapperspb.StringValue{Value: s}
}

// ==================== Int64 Conversions ====================

// Int64ValueFromNullInt64 converts sql.NullInt64 to *wrapperspb.Int64Value.
// Returns nil if the NullInt64 is not valid.
func Int64ValueFromNullInt64(ni sql.NullInt64) *wrapperspb.Int64Value {
	if ni.Valid {
		return &wrapperspb.Int64Value{Value: ni.Int64}
	}
	return nil
}

// NullInt64FromInt64Value converts *wrapperspb.Int64Value to sql.NullInt64.
// Returns an invalid NullInt64 if the Int64Value is nil.
func NullInt64FromInt64Value(iv *wrapperspb.Int64Value) sql.NullInt64 {
	if iv != nil {
		return sql.NullInt64{Int64: iv.Value, Valid: true}
	}
	return sql.NullInt64{Valid: false}
}

// Int64ValueFromPtr converts *int64 to *wrapperspb.Int64Value.
// Returns nil if the pointer is nil.
func Int64ValueFromPtr(i *int64) *wrapperspb.Int64Value {
	if i != nil {
		return &wrapperspb.Int64Value{Value: *i}
	}
	return nil
}

// Int64PtrFromInt64Value converts *wrapperspb.Int64Value to *int64.
// Returns nil if the Int64Value is nil.
func Int64PtrFromInt64Value(iv *wrapperspb.Int64Value) *int64 {
	if iv != nil {
		return &iv.Value
	}
	return nil
}

// Int64FromInt64Value converts *wrapperspb.Int64Value to int64.
// Returns 0 if the Int64Value is nil.
func Int64FromInt64Value(iv *wrapperspb.Int64Value) int64 {
	if iv != nil {
		return iv.Value
	}
	return 0
}

// Int64FromInt64ValueOr converts *wrapperspb.Int64Value to int64 with a default.
func Int64FromInt64ValueOr(iv *wrapperspb.Int64Value, defaultVal int64) int64 {
	if iv != nil {
		return iv.Value
	}
	return defaultVal
}

// Int64ValueFrom creates an Int64Value from an int64.
func Int64ValueFrom(i int64) *wrapperspb.Int64Value {
	return &wrapperspb.Int64Value{Value: i}
}

// ==================== Int32 Conversions ====================

// Int32ValueFromNullInt32 converts sql.NullInt32 to *wrapperspb.Int32Value.
// Returns nil if the NullInt32 is not valid.
func Int32ValueFromNullInt32(ni sql.NullInt32) *wrapperspb.Int32Value {
	if ni.Valid {
		return &wrapperspb.Int32Value{Value: ni.Int32}
	}
	return nil
}

// NullInt32FromInt32Value converts *wrapperspb.Int32Value to sql.NullInt32.
// Returns an invalid NullInt32 if the Int32Value is nil.
func NullInt32FromInt32Value(iv *wrapperspb.Int32Value) sql.NullInt32 {
	if iv != nil {
		return sql.NullInt32{Int32: iv.Value, Valid: true}
	}
	return sql.NullInt32{Valid: false}
}

// Int32ValueFromPtr converts *int32 to *wrapperspb.Int32Value.
// Returns nil if the pointer is nil.
func Int32ValueFromPtr(i *int32) *wrapperspb.Int32Value {
	if i != nil {
		return &wrapperspb.Int32Value{Value: *i}
	}
	return nil
}

// Int32PtrFromInt32Value converts *wrapperspb.Int32Value to *int32.
// Returns nil if the Int32Value is nil.
func Int32PtrFromInt32Value(iv *wrapperspb.Int32Value) *int32 {
	if iv != nil {
		return &iv.Value
	}
	return nil
}

// Int32FromInt32Value converts *wrapperspb.Int32Value to int32.
// Returns 0 if the Int32Value is nil.
func Int32FromInt32Value(iv *wrapperspb.Int32Value) int32 {
	if iv != nil {
		return iv.Value
	}
	return 0
}

// Int32FromInt32ValueOr converts *wrapperspb.Int32Value to int32 with a default.
func Int32FromInt32ValueOr(iv *wrapperspb.Int32Value, defaultVal int32) int32 {
	if iv != nil {
		return iv.Value
	}
	return defaultVal
}

// Int32ValueFrom creates an Int32Value from an int32.
func Int32ValueFrom(i int32) *wrapperspb.Int32Value {
	return &wrapperspb.Int32Value{Value: i}
}

// ==================== Bool Conversions ====================

// BoolValueFromNullBool converts sql.NullBool to *wrapperspb.BoolValue.
// Returns nil if the NullBool is not valid.
func BoolValueFromNullBool(nb sql.NullBool) *wrapperspb.BoolValue {
	if nb.Valid {
		return &wrapperspb.BoolValue{Value: nb.Bool}
	}
	return nil
}

// NullBoolFromBoolValue converts *wrapperspb.BoolValue to sql.NullBool.
// Returns an invalid NullBool if the BoolValue is nil.
func NullBoolFromBoolValue(bv *wrapperspb.BoolValue) sql.NullBool {
	if bv != nil {
		return sql.NullBool{Bool: bv.Value, Valid: true}
	}
	return sql.NullBool{Valid: false}
}

// BoolValueFromPtr converts *bool to *wrapperspb.BoolValue.
// Returns nil if the pointer is nil.
func BoolValueFromPtr(b *bool) *wrapperspb.BoolValue {
	if b != nil {
		return &wrapperspb.BoolValue{Value: *b}
	}
	return nil
}

// BoolPtrFromBoolValue converts *wrapperspb.BoolValue to *bool.
// Returns nil if the BoolValue is nil.
func BoolPtrFromBoolValue(bv *wrapperspb.BoolValue) *bool {
	if bv != nil {
		return &bv.Value
	}
	return nil
}

// BoolFromBoolValue converts *wrapperspb.BoolValue to bool.
// Returns false if the BoolValue is nil.
func BoolFromBoolValue(bv *wrapperspb.BoolValue) bool {
	if bv != nil {
		return bv.Value
	}
	return false
}

// BoolFromBoolValueOr converts *wrapperspb.BoolValue to bool with a default.
func BoolFromBoolValueOr(bv *wrapperspb.BoolValue, defaultVal bool) bool {
	if bv != nil {
		return bv.Value
	}
	return defaultVal
}

// BoolValueFrom creates a BoolValue from a bool.
func BoolValueFrom(b bool) *wrapperspb.BoolValue {
	return &wrapperspb.BoolValue{Value: b}
}

// ==================== Float64/Double Conversions ====================

// DoubleValueFromNullFloat64 converts sql.NullFloat64 to *wrapperspb.DoubleValue.
// Returns nil if the NullFloat64 is not valid.
func DoubleValueFromNullFloat64(nf sql.NullFloat64) *wrapperspb.DoubleValue {
	if nf.Valid {
		return &wrapperspb.DoubleValue{Value: nf.Float64}
	}
	return nil
}

// NullFloat64FromDoubleValue converts *wrapperspb.DoubleValue to sql.NullFloat64.
// Returns an invalid NullFloat64 if the DoubleValue is nil.
func NullFloat64FromDoubleValue(dv *wrapperspb.DoubleValue) sql.NullFloat64 {
	if dv != nil {
		return sql.NullFloat64{Float64: dv.Value, Valid: true}
	}
	return sql.NullFloat64{Valid: false}
}

// DoubleValueFromPtr converts *float64 to *wrapperspb.DoubleValue.
// Returns nil if the pointer is nil.
func DoubleValueFromPtr(f *float64) *wrapperspb.DoubleValue {
	if f != nil {
		return &wrapperspb.DoubleValue{Value: *f}
	}
	return nil
}

// Float64PtrFromDoubleValue converts *wrapperspb.DoubleValue to *float64.
// Returns nil if the DoubleValue is nil.
func Float64PtrFromDoubleValue(dv *wrapperspb.DoubleValue) *float64 {
	if dv != nil {
		return &dv.Value
	}
	return nil
}

// Float64FromDoubleValue converts *wrapperspb.DoubleValue to float64.
// Returns 0 if the DoubleValue is nil.
func Float64FromDoubleValue(dv *wrapperspb.DoubleValue) float64 {
	if dv != nil {
		return dv.Value
	}
	return 0
}

// DoubleValueFrom creates a DoubleValue from a float64.
func DoubleValueFrom(f float64) *wrapperspb.DoubleValue {
	return &wrapperspb.DoubleValue{Value: f}
}

// DoubleValueFromFloat32 creates a DoubleValue from a float32 (widening conversion).
func DoubleValueFromFloat32(f float32) *wrapperspb.DoubleValue {
	return &wrapperspb.DoubleValue{Value: float64(f)}
}

// ==================== Float32 Conversions ====================

// FloatValueFromNullFloat64 converts sql.NullFloat64 to *wrapperspb.FloatValue.
// Returns nil if the NullFloat64 is not valid.
// Note: This is a narrowing conversion from float64 to float32.
func FloatValueFromNullFloat64(nf sql.NullFloat64) *wrapperspb.FloatValue {
	if nf.Valid {
		return &wrapperspb.FloatValue{Value: float32(nf.Float64)}
	}
	return nil
}

// NullFloat64FromFloatValue converts *wrapperspb.FloatValue to sql.NullFloat64.
// Returns an invalid NullFloat64 if the FloatValue is nil.
// Note: This is a widening conversion from float32 to float64.
func NullFloat64FromFloatValue(fv *wrapperspb.FloatValue) sql.NullFloat64 {
	if fv != nil {
		return sql.NullFloat64{Float64: float64(fv.Value), Valid: true}
	}
	return sql.NullFloat64{Valid: false}
}

// FloatValueFromPtr converts *float32 to *wrapperspb.FloatValue.
// Returns nil if the pointer is nil.
func FloatValueFromPtr(f *float32) *wrapperspb.FloatValue {
	if f != nil {
		return &wrapperspb.FloatValue{Value: *f}
	}
	return nil
}

// Float32PtrFromFloatValue converts *wrapperspb.FloatValue to *float32.
// Returns nil if the FloatValue is nil.
func Float32PtrFromFloatValue(fv *wrapperspb.FloatValue) *float32 {
	if fv != nil {
		return &fv.Value
	}
	return nil
}

// Float32FromFloatValue converts *wrapperspb.FloatValue to float32.
// Returns 0 if the FloatValue is nil.
func Float32FromFloatValue(fv *wrapperspb.FloatValue) float32 {
	if fv != nil {
		return fv.Value
	}
	return 0
}

// FloatValueFrom creates a FloatValue from a float32.
func FloatValueFrom(f float32) *wrapperspb.FloatValue {
	return &wrapperspb.FloatValue{Value: f}
}

// ==================== UInt32 Conversions ====================

// UInt32ValueFromPtr converts *uint32 to *wrapperspb.UInt32Value.
// Returns nil if the pointer is nil.
func UInt32ValueFromPtr(u *uint32) *wrapperspb.UInt32Value {
	if u != nil {
		return &wrapperspb.UInt32Value{Value: *u}
	}
	return nil
}

// UInt32PtrFromUInt32Value converts *wrapperspb.UInt32Value to *uint32.
// Returns nil if the UInt32Value is nil.
func UInt32PtrFromUInt32Value(uv *wrapperspb.UInt32Value) *uint32 {
	if uv != nil {
		return &uv.Value
	}
	return nil
}

// UInt32FromUInt32Value converts *wrapperspb.UInt32Value to uint32.
// Returns 0 if the UInt32Value is nil.
func UInt32FromUInt32Value(uv *wrapperspb.UInt32Value) uint32 {
	if uv != nil {
		return uv.Value
	}
	return 0
}

// UInt32FromUInt32ValueOr converts *wrapperspb.UInt32Value to uint32 with a default.
func UInt32FromUInt32ValueOr(uv *wrapperspb.UInt32Value, defaultVal uint32) uint32 {
	if uv != nil {
		return uv.Value
	}
	return defaultVal
}

// UInt32ValueFrom creates a UInt32Value from a uint32.
func UInt32ValueFrom(u uint32) *wrapperspb.UInt32Value {
	return &wrapperspb.UInt32Value{Value: u}
}

// ==================== UInt64 Conversions ====================

// UInt64ValueFromPtr converts *uint64 to *wrapperspb.UInt64Value.
// Returns nil if the pointer is nil.
func UInt64ValueFromPtr(u *uint64) *wrapperspb.UInt64Value {
	if u != nil {
		return &wrapperspb.UInt64Value{Value: *u}
	}
	return nil
}

// UInt64PtrFromUInt64Value converts *wrapperspb.UInt64Value to *uint64.
// Returns nil if the UInt64Value is nil.
func UInt64PtrFromUInt64Value(uv *wrapperspb.UInt64Value) *uint64 {
	if uv != nil {
		return &uv.Value
	}
	return nil
}

// UInt64FromUInt64Value converts *wrapperspb.UInt64Value to uint64.
// Returns 0 if the UInt64Value is nil.
func UInt64FromUInt64Value(uv *wrapperspb.UInt64Value) uint64 {
	if uv != nil {
		return uv.Value
	}
	return 0
}

// UInt64FromUInt64ValueOr converts *wrapperspb.UInt64Value to uint64 with a default.
func UInt64FromUInt64ValueOr(uv *wrapperspb.UInt64Value, defaultVal uint64) uint64 {
	if uv != nil {
		return uv.Value
	}
	return defaultVal
}

// UInt64ValueFrom creates a UInt64Value from a uint64.
func UInt64ValueFrom(u uint64) *wrapperspb.UInt64Value {
	return &wrapperspb.UInt64Value{Value: u}
}

// ==================== Bytes Conversions ====================

// BytesValueFromSlice converts []byte to *wrapperspb.BytesValue.
// Returns nil if the slice is nil or empty.
func BytesValueFromSlice(b []byte) *wrapperspb.BytesValue {
	if b != nil && len(b) > 0 {
		return &wrapperspb.BytesValue{Value: b}
	}
	return nil
}

// BytesFromBytesValue converts *wrapperspb.BytesValue to []byte.
// Returns nil if the BytesValue is nil.
func BytesFromBytesValue(bv *wrapperspb.BytesValue) []byte {
	if bv != nil {
		return bv.Value
	}
	return nil
}

// BytesValueFrom creates a BytesValue from a []byte.
// Unlike BytesValueFromSlice, this creates a BytesValue even for empty slices.
func BytesValueFrom(b []byte) *wrapperspb.BytesValue {
	return &wrapperspb.BytesValue{Value: b}
}

// ==================== String Slice Conversions ====================

// StringValuesFromSlice converts []string to []*wrapperspb.StringValue.
// Returns nil if the slice is nil.
func StringValuesFromSlice(strs []string) []*wrapperspb.StringValue {
	if strs == nil {
		return nil
	}
	result := make([]*wrapperspb.StringValue, len(strs))
	for i, s := range strs {
		result[i] = &wrapperspb.StringValue{Value: s}
	}
	return result
}

// StringSliceFromStringValues converts []*wrapperspb.StringValue to []string.
// Returns nil if the slice is nil. Skips nil values in the input.
func StringSliceFromStringValues(wrappers []*wrapperspb.StringValue) []string {
	if wrappers == nil {
		return nil
	}
	result := make([]string, 0, len(wrappers))
	for _, w := range wrappers {
		if w != nil {
			result = append(result, w.Value)
		}
	}
	return result
}
