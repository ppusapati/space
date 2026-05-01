package proto

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// StringToPtr converts a string to *string. Returns nil for empty string.
func StringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// PtrToString converts *string to string. Returns empty string if nil.
func PtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Int32ToPtr converts int32 to *int32. Returns nil for zero.
func Int32ToPtr(v int32) *int32 {
	if v == 0 {
		return nil
	}
	return &v
}

// BoolToPtr converts bool to *bool.
func BoolToPtr(b bool) *bool {
	return &b
}

// ToBoolPtr converts bool to *bool. Alias for BoolToPtr.
func ToBoolPtr(b bool) *bool {
	return &b
}

// BoolPtrFromBool converts bool to *bool. Alias for BoolToPtr.
func BoolPtrFromBool(b bool) *bool {
	return &b
}

// BoolFromBoolPtr converts *bool to bool. Returns false if nil.
func BoolFromBoolPtr(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// ToStringPtr converts a string to *string. Alias for StringToPtr.
func ToStringPtr(s string) *string {
	return StringToPtr(s)
}

// ToTimePtr converts *timestamppb.Timestamp to *time.Time.
func ToTimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

// TimestampToTimePtr converts *timestamppb.Timestamp to *time.Time.
func TimestampToTimePtr(ts *timestamppb.Timestamp) *time.Time {
	return ToTimePtr(ts)
}

// TimestampToTime converts *timestamppb.Timestamp to time.Time.
func TimestampToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}

// TimestampToPtr converts time.Time to *timestamppb.Timestamp.
func TimestampToPtr(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

// StringToDecimal converts string to decimal.Decimal. Returns zero on parse failure.
func StringToDecimal(s string) decimal.Decimal {
	d, _ := decimal.NewFromString(s)
	return d
}

// ToDecimal converts string to decimal.Decimal. Alias for StringToDecimal.
func ToDecimal(s string) decimal.Decimal {
	return StringToDecimal(s)
}

// DecimalToString converts decimal.Decimal to string.
func DecimalToString(d decimal.Decimal) string {
	return d.String()
}

// ToInt32Ptr converts int32 to *int32. Returns nil for zero.
func ToInt32Ptr(v int32) *int32 {
	if v == 0 {
		return nil
	}
	return &v
}

// PtrToInt32 converts *int32 to int32. Returns 0 if nil.
func PtrToInt32(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}

// ToTime converts *timestamppb.Timestamp to time.Time. Returns zero time if nil.
func ToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}

// PtrToBool converts *bool to bool. Returns false if nil.
func PtrToBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// ToInt32PtrFromString converts string to *int32. Returns nil for empty or invalid.
func ToInt32PtrFromString(s string) *int32 {
	if s == "" {
		return nil
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return nil
	}
	v := int32(d.IntPart())
	return &v
}

// PtrToTimestamp converts *time.Time to *timestamppb.Timestamp. Returns nil if nil.
func PtrToTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

// FromTimePtr converts *time.Time to *timestamppb.Timestamp. Alias for PtrToTimestamp.
func FromTimePtr(t *time.Time) *timestamppb.Timestamp {
	return PtrToTimestamp(t)
}

// FromInt32Ptr converts *int32 to int32. Returns 0 if nil. Alias for PtrToInt32.
func FromInt32Ptr(v *int32) int32 {
	return PtrToInt32(v)
}

// FromStringPtr converts *string to string. Returns empty string if nil.
func FromStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// FromBoolPtr converts *bool to bool. Returns false if nil.
func FromBoolPtr(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// DecimalFromString converts string to decimal.Decimal. Alias for StringToDecimal.
func DecimalFromString(s string) decimal.Decimal {
	return StringToDecimal(s)
}

// BytesToMap converts JSON bytes to map[string]interface{}.
func BytesToMap(data []byte) map[string]interface{} {
	if data == nil || len(data) == 0 {
		return nil
	}
	result := make(map[string]interface{})
	if err := json.Unmarshal(data, &result); err != nil {
		return nil
	}
	return result
}

// MapToBytes converts a map to JSON bytes.
func MapToBytes(m map[string]interface{}) ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// StringToStringPtr converts a string to *string. Alias for StringToPtr.
func StringToStringPtr(s string) *string {
	return StringToPtr(s)
}

// StringPtrToString converts *string to string. Returns empty string if nil.
// Alias for FromStringPtr.
func StringPtrToString(s *string) string {
	return FromStringPtr(s)
}

// TimePtrToDateString converts *time.Time to a date string (YYYY-MM-DD format).
// Returns empty string if nil.
func TimePtrToDateString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

// FormatInt converts an int to string.
func FormatInt(v int) string {
	return strconv.Itoa(v)
}

// FormatInt32 converts an int32 to string.
func FormatInt32(v int32) string {
	return strconv.FormatInt(int64(v), 10)
}

// FormatFloat converts a float64 to string.
// An optional precision parameter controls the number of decimal places.
// If no precision is given, uses full precision (-1).
func FormatFloat(f float64, precision ...int) string {
	prec := -1
	if len(precision) > 0 {
		prec = precision[0]
	}
	return strconv.FormatFloat(f, 'f', prec, 64)
}

// Int32ToInt32Ptr converts int32 to *int32. Always returns a pointer, even for zero.
func Int32ToInt32Ptr(v int32) *int32 {
	return &v
}

// Int32PtrToInt32 converts *int32 to int32. Returns 0 if nil.
func Int32PtrToInt32(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}

// BytesFromStringMap converts map[string]string to JSON bytes.
func BytesFromStringMap(m map[string]string) []byte {
	if m == nil {
		return nil
	}
	data, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return data
}

// StringMapFromBytes converts JSON bytes to map[string]string.
func StringMapFromBytes(data []byte) map[string]string {
	if data == nil || len(data) == 0 {
		return nil
	}
	result := make(map[string]string)
	if err := json.Unmarshal(data, &result); err != nil {
		return nil
	}
	return result
}

// TimeStringToTimePtr converts a time string (RFC3339 or date format) to *time.Time.
// Returns nil for empty string or parse failure.
func TimeStringToTimePtr(s string) *time.Time {
	if s == "" {
		return nil
	}
	// Try RFC3339 first
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// Try date-only format
		t, err = time.Parse("2006-01-02", s)
		if err != nil {
			return nil
		}
	}
	return &t
}

// TimePtrToTimeString converts *time.Time to an RFC3339 string.
// Returns empty string if nil.
func TimePtrToTimeString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

// DateStringToTimePtr converts a date string (YYYY-MM-DD) to *time.Time.
// Returns nil for empty string or parse failure.
func DateStringToTimePtr(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}

// BoolToBoolPtr converts bool to *bool. Always returns a pointer.
func BoolToBoolPtr(b bool) *bool {
	return &b
}

// BoolPtrToBool converts *bool to bool. Returns false if nil.
func BoolPtrToBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}
