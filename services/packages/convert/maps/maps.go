// Package maps provides map conversion utilities.
package maps

import (
	"encoding/json"
	"strconv"
)

// StringToAny converts map[string]string to map[string]any.
func StringToAny(m map[string]string) map[string]any {
	if m == nil {
		return nil
	}
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// AnyToString converts map[string]any to map[string]string.
// Non-string values are converted using fmt.Sprintf("%v", v).
func AnyToString(m map[string]any) map[string]string {
	if m == nil {
		return nil
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		if s, ok := v.(string); ok {
			result[k] = s
		} else if v != nil {
			result[k] = stringFromAny(v)
		}
	}
	return result
}

// stringFromAny converts any value to string.
func stringFromAny(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case int:
		return intToString(int64(val))
	case int8:
		return intToString(int64(val))
	case int16:
		return intToString(int64(val))
	case int32:
		return intToString(int64(val))
	case int64:
		return intToString(val)
	case uint:
		return uintToString(uint64(val))
	case uint8:
		return uintToString(uint64(val))
	case uint16:
		return uintToString(uint64(val))
	case uint32:
		return uintToString(uint64(val))
	case uint64:
		return uintToString(val)
	case float32:
		return floatToString(float64(val))
	case float64:
		return floatToString(val)
	default:
		return ""
	}
}

// intToString converts int64 to string without importing strconv.
func intToString(n int64) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if negative {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// uintToString converts uint64 to string without importing strconv.
func uintToString(n uint64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// floatToString converts float64 to string using strconv.
func floatToString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// Clone creates a shallow copy of the map.
func Clone[K comparable, V any](m map[K]V) map[K]V {
	if m == nil {
		return nil
	}
	result := make(map[K]V, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// Merge merges multiple maps into one. Later maps override earlier ones.
func Merge[K comparable, V any](maps ...map[K]V) map[K]V {
	var total int
	for _, m := range maps {
		total += len(m)
	}
	result := make(map[K]V, total)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// MapFromJSONString converts a JSON string to map[string]any.
// Returns nil if the string is empty or cannot be unmarshaled.
func MapFromJSONString(s string) map[string]any {
	if s == "" {
		return nil
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return nil
	}
	return result
}

// JSONStringFromMap converts map[string]any to a JSON string.
// Returns empty string if the map is nil or cannot be marshaled.
func JSONStringFromMap(m map[string]any) string {
	if m == nil {
		return ""
	}
	data, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(data)
}
