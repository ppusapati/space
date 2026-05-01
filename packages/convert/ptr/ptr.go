// Package ptr provides generic pointer utilities.
package ptr

import "strconv"

// Ptr returns a pointer to the given value.
func Ptr[T any](v T) *T {
	return &v
}

// Deref returns the value that p points to.
// Returns the zero value of T if p is nil.
func Deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// DerefOr returns the value that p points to, or defaultVal if p is nil.
func DerefOr[T any](p *T, defaultVal T) T {
	if p == nil {
		return defaultVal
	}
	return *p
}

// PtrOrNil returns a pointer to v if v is not the zero value of T, otherwise nil.
func PtrOrNil[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}

// PtrSlice converts a slice of values to a slice of pointers.
func PtrSlice[T any](s []T) []*T {
	if s == nil {
		return nil
	}
	result := make([]*T, len(s))
	for i := range s {
		result[i] = &s[i]
	}
	return result
}

// DerefSlice converts a slice of pointers to a slice of values.
// Nil pointers are converted to zero values.
func DerefSlice[T any](s []*T) []T {
	if s == nil {
		return nil
	}
	result := make([]T, len(s))
	for i, p := range s {
		if p != nil {
			result[i] = *p
		}
	}
	return result
}

// DerefSliceSkipNil converts a slice of pointers to a slice of values,
// skipping nil pointers.
func DerefSliceSkipNil[T any](s []*T) []T {
	if s == nil {
		return nil
	}
	result := make([]T, 0, len(s))
	for _, p := range s {
		if p != nil {
			result = append(result, *p)
		}
	}
	return result
}

// Equal returns true if both pointers are nil or both point to equal values.
func Equal[T comparable](a, b *T) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// Clone creates a new pointer with a copy of the value.
// Returns nil if p is nil.
func Clone[T any](p *T) *T {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

// Coalesce returns the first non-nil pointer, or nil if all are nil.
func Coalesce[T any](ptrs ...*T) *T {
	for _, p := range ptrs {
		if p != nil {
			return p
		}
	}
	return nil
}

// CoalesceValue returns the first non-nil pointer's value, or defaultVal if all are nil.
func CoalesceValue[T any](defaultVal T, ptrs ...*T) T {
	for _, p := range ptrs {
		if p != nil {
			return *p
		}
	}
	return defaultVal
}

// ==================== Type-specific helpers for common types ====================

// String creates a pointer to a string.
func String(s string) *string {
	return &s
}

// StringOrNil returns a pointer to s if s is non-empty, otherwise nil.
// This is commonly used for proto conversions where empty string means unset.
func StringOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// StringValue returns the string value or empty string if nil.
func StringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// StringVal returns the string value or empty string if nil. Alias for StringValue.
func StringVal(s *string) string {
	return StringValue(s)
}

// Int creates a pointer to an int.
func Int(i int) *int {
	return &i
}

// IntValue returns the int value or 0 if nil.
func IntValue(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// Int32 creates a pointer to an int32.
func Int32(i int32) *int32 {
	return &i
}

// Int32Value returns the int32 value or 0 if nil.
func Int32Value(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

// Int64 creates a pointer to an int64.
func Int64(i int64) *int64 {
	return &i
}

// Int64Value returns the int64 value or 0 if nil.
func Int64Value(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

// Float32 creates a pointer to a float32.
func Float32(f float32) *float32 {
	return &f
}

// Float32Value returns the float32 value or 0 if nil.
func Float32Value(f *float32) float32 {
	if f == nil {
		return 0
	}
	return *f
}

// Float64 creates a pointer to a float64.
func Float64(f float64) *float64 {
	return &f
}

// Float64Value returns the float64 value or 0 if nil.
func Float64Value(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

// Bool creates a pointer to a bool.
func Bool(b bool) *bool {
	return &b
}

// BoolValue returns the bool value or false if nil.
func BoolValue(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}


// Of creates a pointer to the given value. Alias for Ptr.
func Of[T any](v T) *T {
	return &v
}

// BoolPtr creates a pointer to a bool value.
func BoolPtr(b bool) *bool {
	return &b
}

// Int32OrNil returns a pointer to an int32 value if non-zero, otherwise nil.
func Int32OrNil(v int32) *int32 {
	if v == 0 {
		return nil
	}
	return &v
}

// Int64OrNil returns a pointer to an int64 value if non-zero, otherwise nil.
func Int64OrNil(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}

// FormatInt converts an int value to its string representation.
func FormatInt(v int) string {
	return strconv.Itoa(v)
}

// FormatFloat converts a float64 value to its string representation.
// An optional precision parameter controls the number of decimal places.
// If no precision is given, uses full precision (-1).
func FormatFloat(f float64, precision ...int) string {
	prec := -1
	if len(precision) > 0 {
		prec = precision[0]
	}
	return strconv.FormatFloat(f, 'f', prec, 64)
}
