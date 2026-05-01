package sql

import (
	"database/sql"
	"encoding/json"

	"github.com/sqlc-dev/pqtype"
)

// ==================== NullRawMessage ↔ Map Conversions ====================

// MapFromNullRawMessage converts pqtype.NullRawMessage to map[string]any.
// Returns nil if the NullRawMessage is not valid or on unmarshal error.
func MapFromNullRawMessage(nrm pqtype.NullRawMessage) map[string]any {
	if !nrm.Valid || len(nrm.RawMessage) == 0 {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(nrm.RawMessage, &m); err != nil {
		return nil
	}
	return m
}

// MapFromNullRawMessageWithError converts pqtype.NullRawMessage to map[string]any with error return.
func MapFromNullRawMessageWithError(nrm pqtype.NullRawMessage) (map[string]any, error) {
	if !nrm.Valid || len(nrm.RawMessage) == 0 {
		return nil, nil
	}
	var m map[string]any
	if err := json.Unmarshal(nrm.RawMessage, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// NullRawMessageFromMap converts map[string]any to pqtype.NullRawMessage.
// Returns an invalid NullRawMessage if the map is nil or on marshal error.
func NullRawMessageFromMap(m map[string]any) pqtype.NullRawMessage {
	if m == nil {
		return pqtype.NullRawMessage{Valid: false}
	}
	b, err := json.Marshal(m)
	if err != nil {
		return pqtype.NullRawMessage{Valid: false}
	}
	return pqtype.NullRawMessage{
		RawMessage: b,
		Valid:      true,
	}
}

// NullRawMessageFromMapWithError converts map[string]any to pqtype.NullRawMessage with error return.
func NullRawMessageFromMapWithError(m map[string]any) (pqtype.NullRawMessage, error) {
	if m == nil {
		return pqtype.NullRawMessage{Valid: false}, nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return pqtype.NullRawMessage{Valid: false}, err
	}
	return pqtype.NullRawMessage{
		RawMessage: b,
		Valid:      true,
	}, nil
}

// ==================== NullString ↔ Map Conversions (JSON stored as string) ====================

// MapFromNullString converts sql.NullString (containing JSON) to map[string]any.
// Returns nil if the NullString is not valid or on unmarshal error.
func MapFromNullString(ns sql.NullString) map[string]any {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(ns.String), &m); err != nil {
		return nil
	}
	return m
}

// MapFromNullStringWithError converts sql.NullString to map[string]any with error return.
func MapFromNullStringWithError(ns sql.NullString) (map[string]any, error) {
	if !ns.Valid || ns.String == "" {
		return nil, nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(ns.String), &m); err != nil {
		return nil, err
	}
	return m, nil
}

// NullStringFromMap converts map[string]any to sql.NullString (as JSON).
// Returns an invalid NullString if the map is nil or on marshal error.
func NullStringFromMap(m map[string]any) sql.NullString {
	if m == nil {
		return sql.NullString{Valid: false}
	}
	b, err := json.Marshal(m)
	if err != nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{
		String: string(b),
		Valid:  true,
	}
}

// NullStringFromMapWithError converts map[string]any to sql.NullString with error return.
func NullStringFromMapWithError(m map[string]any) (sql.NullString, error) {
	if m == nil {
		return sql.NullString{Valid: false}, nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return sql.NullString{Valid: false}, err
	}
	return sql.NullString{
		String: string(b),
		Valid:  true,
	}, nil
}

// ==================== Bytes ↔ Map Conversions ====================

// MapFromBytes converts JSON bytes to map[string]any.
// Returns nil if the bytes are nil/empty or on unmarshal error.
func MapFromBytes(b []byte) map[string]any {
	if b == nil || len(b) == 0 {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	return m
}

// MapFromBytesWithError converts JSON bytes to map[string]any with error return.
func MapFromBytesWithError(b []byte) (map[string]any, error) {
	if b == nil || len(b) == 0 {
		return nil, nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// BytesFromMap converts map[string]any to JSON bytes.
// Returns nil if the map is nil or on marshal error.
func BytesFromMap(m map[string]any) []byte {
	if m == nil {
		return nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return b
}

// BytesFromMapWithError converts map[string]any to JSON bytes with error return.
func BytesFromMapWithError(m map[string]any) ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// ==================== String Map Conversions ====================

// StringMapFromMap converts map[string]any to map[string]string.
// Non-string values are converted using their string representation.
// Returns nil if the input map is nil.
func StringMapFromMap(m map[string]any) map[string]string {
	if m == nil {
		return nil
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		if s, ok := v.(string); ok {
			result[k] = s
		}
	}
	return result
}

// MapFromStringMap converts map[string]string to map[string]any.
// Returns nil if the input map is nil.
func MapFromStringMap(m map[string]string) map[string]any {
	if m == nil {
		return nil
	}
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// ==================== String Map ↔ Bytes Conversions ====================

// StringMapFromBytes converts JSON bytes to map[string]string.
// Returns nil if the bytes are nil/empty or on unmarshal error.
func StringMapFromBytes(b []byte) map[string]string {
	if b == nil || len(b) == 0 {
		return nil
	}
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	return m
}

// StringMapFromBytesWithError converts JSON bytes to map[string]string with error return.
func StringMapFromBytesWithError(b []byte) (map[string]string, error) {
	if b == nil || len(b) == 0 {
		return nil, nil
	}
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// BytesFromStringMap converts map[string]string to JSON bytes.
// Returns nil if the map is nil or on marshal error.
func BytesFromStringMap(m map[string]string) []byte {
	if m == nil {
		return nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return b
}

// BytesFromStringMapWithError converts map[string]string to JSON bytes with error return.
func BytesFromStringMapWithError(m map[string]string) ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// ==================== Int64 Map ↔ Bytes Conversions ====================

// Int64MapFromBytes converts JSON bytes to map[string]int64.
// Returns nil if the bytes are nil/empty or on unmarshal error.
func Int64MapFromBytes(b []byte) map[string]int64 {
	if b == nil || len(b) == 0 {
		return nil
	}
	var m map[string]int64
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	return m
}

// Int64MapFromBytesWithError converts JSON bytes to map[string]int64 with error return.
func Int64MapFromBytesWithError(b []byte) (map[string]int64, error) {
	if b == nil || len(b) == 0 {
		return nil, nil
	}
	var m map[string]int64
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// BytesFromInt64Map converts map[string]int64 to JSON bytes.
// Returns nil if the map is nil or on marshal error.
func BytesFromInt64Map(m map[string]int64) []byte {
	if m == nil {
		return nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return b
}

// BytesFromInt64MapWithError converts map[string]int64 to JSON bytes with error return.
func BytesFromInt64MapWithError(m map[string]int64) ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}


// BytesToMap converts a JSON byte slice to a map[string]interface{}.
func BytesToMap(data []byte) (map[string]interface{}, error) {
	if data == nil || len(data) == 0 {
		return nil, nil
	}
	result := make(map[string]interface{})
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// MapToBytes converts a map[string]interface{} to a JSON byte slice.
func MapToBytes(m map[string]interface{}) ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}
