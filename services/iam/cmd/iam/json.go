// json.go — tiny JSON helpers used by the cmd/iam adapters.
// Kept in a sibling file so adapters.go stays focused on the
// adapter shapes.

package main

import (
	"encoding/json"
	"io"
)

func jsonMarshal(v any) ([]byte, error) { return json.Marshal(v) }

func jsonDecode(r io.Reader, v any) error { return json.NewDecoder(r).Decode(v) }
