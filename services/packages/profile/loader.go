// loader.go — YAML ↔ proto round-trip for SpacecraftProfile.
//
// Functions:
//   • LoadFile          — read + parse + validate a YAML file from
//                          the OS filesystem.
//   • LoadFromFS        — same, against an arbitrary fs.FS (tests).
//   • LoadBytes         — parse + validate raw YAML bytes.
//   • Marshal           — render a SpacecraftProfile back to YAML
//                          bytes. Round-trip-stable.
//
// All loader entry points run Validate before returning the profile
// — callers never see an inconsistent SpacecraftProfile.

package profile

import (
	"fmt"
	"io/fs"

	"gopkg.in/yaml.v3"
)

// LoadFile reads a YAML profile from disk, parses, and validates it.
// Convenience wrapper around LoadBytes for service entrypoints.
func LoadFile(path string) (*SpacecraftProfile, error) {
	raw, err := defaultReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("profile: read %s: %w", path, err)
	}
	return LoadBytes(raw)
}

// LoadFromFS reads a YAML profile from an arbitrary fs.FS (typically
// embed.FS or testing/fstest.MapFS), parses, and validates it.
func LoadFromFS(fsys fs.FS, path string) (*SpacecraftProfile, error) {
	raw, err := readFromFS(fsys, path)
	if err != nil {
		return nil, fmt.Errorf("profile: read %s: %w", path, err)
	}
	return LoadBytes(raw)
}

// LoadBytes parses raw YAML bytes into a SpacecraftProfile and
// validates the result. Returns the populated profile + nil on
// success; nil + error on parse OR validation failure.
func LoadBytes(raw []byte) (*SpacecraftProfile, error) {
	var p SpacecraftProfile
	if err := yaml.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("profile: parse YAML: %w", err)
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return &p, nil
}

// Marshal renders a SpacecraftProfile to YAML bytes. The output is
// deterministic — repeated calls on the same struct produce
// byte-identical output. Round-trip stability:
//
//   parsed, _ := LoadBytes(input)
//   output, _ := Marshal(parsed)
//   reparsed, _ := LoadBytes(output)
//   reflect.DeepEqual(parsed, reparsed) // true
//
// The function does NOT call Validate; callers Marshal known-good
// profiles. To Marshal an unvalidated draft (e.g. an editor saving
// in-progress work), the caller can wrap explicitly.
func Marshal(p *SpacecraftProfile) ([]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("profile: cannot marshal nil profile")
	}
	out, err := yaml.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("profile: marshal YAML: %w", err)
	}
	return out, nil
}
