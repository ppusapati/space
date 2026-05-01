package proto

import (
	"encoding/json"

	"google.golang.org/protobuf/types/known/structpb"
)

// ==================== Map ↔ Struct Conversions ====================

// StructFromMap converts map[string]any to *structpb.Struct.
// Returns nil if the map is nil or on conversion error.
func StructFromMap(m map[string]any) *structpb.Struct {
	if m == nil {
		return nil
	}
	s, err := structpb.NewStruct(m)
	if err != nil {
		return nil
	}
	return s
}

// StructFromMapWithError converts map[string]any to *structpb.Struct with error return.
// Returns nil, nil if the map is nil.
func StructFromMapWithError(m map[string]any) (*structpb.Struct, error) {
	if m == nil {
		return nil, nil
	}
	return structpb.NewStruct(m)
}

// MapFromStruct converts *structpb.Struct to map[string]any.
// Returns nil if the Struct is nil.
func MapFromStruct(s *structpb.Struct) map[string]any {
	if s == nil {
		return nil
	}
	return s.AsMap()
}

// ==================== JSON ↔ Struct Conversions ====================

// StructFromJSON converts JSON bytes to *structpb.Struct.
// Returns nil on unmarshal error or if the JSON doesn't represent an object.
func StructFromJSON(b []byte) *structpb.Struct {
	if b == nil || len(b) == 0 {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	s, err := structpb.NewStruct(m)
	if err != nil {
		return nil
	}
	return s
}

// StructFromJSONWithError converts JSON bytes to *structpb.Struct with error return.
func StructFromJSONWithError(b []byte) (*structpb.Struct, error) {
	if b == nil || len(b) == 0 {
		return nil, nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return structpb.NewStruct(m)
}

// JSONFromStruct converts *structpb.Struct to JSON bytes.
// Returns nil if the Struct is nil or on marshal error.
func JSONFromStruct(s *structpb.Struct) []byte {
	if s == nil {
		return nil
	}
	b, err := json.Marshal(s.AsMap())
	if err != nil {
		return nil
	}
	return b
}

// JSONFromStructWithError converts *structpb.Struct to JSON bytes with error return.
func JSONFromStructWithError(s *structpb.Struct) ([]byte, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s.AsMap())
}

// ==================== Utility Functions ====================

// IsEmptyStruct checks if a Struct is nil or has no fields.
func IsEmptyStruct(s *structpb.Struct) bool {
	return s == nil || len(s.Fields) == 0
}

// StructFields returns the number of fields in a Struct.
// Returns 0 if the Struct is nil.
func StructFields(s *structpb.Struct) int {
	if s == nil {
		return 0
	}
	return len(s.Fields)
}

// EmptyStruct returns an empty Struct.
func EmptyStruct() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{})
	return s
}
