// Package strconv provides string conversion utilities beyond the standard library.
package strconv

import "strconv"

// Float64FromString parses a string to float64.
// Returns 0 if the string is empty or cannot be parsed.
func Float64FromString(s string) float64 {
	if s == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// StringFromFloat64 formats a float64 to string.
// Returns empty string if the value is 0.
func StringFromFloat64(f float64) string {
	if f == 0 {
		return ""
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// Float64FromStringFull formats a float64 to string, including zero values.
func StringFromFloat64Full(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// Int64FromString parses a string to int64.
// Returns 0 if the string is empty or cannot be parsed.
func Int64FromString(s string) int64 {
	if s == "" {
		return 0
	}
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

// StringFromInt64 formats an int64 to string.
// Returns empty string if the value is 0.
func StringFromInt64(i int64) string {
	if i == 0 {
		return ""
	}
	return strconv.FormatInt(i, 10)
}

// StringFromInt64Full formats an int64 to string, including zero values.
func StringFromInt64Full(i int64) string {
	return strconv.FormatInt(i, 10)
}

// Int32FromString parses a string to int32.
// Returns 0 if the string is empty or cannot be parsed.
func Int32FromString(s string) int32 {
	if s == "" {
		return 0
	}
	i, _ := strconv.ParseInt(s, 10, 32)
	return int32(i)
}

// StringFromInt32 formats an int32 to string.
// Returns empty string if the value is 0.
func StringFromInt32(i int32) string {
	if i == 0 {
		return ""
	}
	return strconv.FormatInt(int64(i), 10)
}

// StringFromInt32Full formats an int32 to string, including zero values.
func StringFromInt32Full(i int32) string {
	return strconv.FormatInt(int64(i), 10)
}

// BoolFromString parses a string to bool.
// Returns false if the string is empty or cannot be parsed.
func BoolFromString(s string) bool {
	if s == "" {
		return false
	}
	b, _ := strconv.ParseBool(s)
	return b
}

// StringFromBool formats a bool to string.
func StringFromBool(b bool) string {
	return strconv.FormatBool(b)
}
