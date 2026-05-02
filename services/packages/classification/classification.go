// Package classification loads and exposes the platform's data
// classification scheme defined in compliance/classification.yaml.
//
// Consumers:
//   • services/packages/api/                 (envelope serializer —
//                                              sets _meta.classification)
//   • services/packages/authz/decision.go    (clearance >= classification)
//   • services/audit/                        (per-row classification tag)
//   • Helm charts                             (pod / topic labelling)
//
// → REQ-FUNC-PLT-AUTHZ-001 (clearance comparison)
// → REQ-FUNC-PLT-AUDIT-003 (per-classification retention)
// → REQ-COMP-ITAR-002       (ITAR namespace + container labels)
// → design.md §4.6
//
// The YAML lives in repo at compliance/classification.yaml and is
// embedded at compile time so service container images do not need to
// ship a copy of the compliance/ directory. Tests inject a fixture FS
// via LoadFromFS.
package classification

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed embedded/classification.yaml
var defaultEmbedFS embed.FS

const (
	// embeddedPath is the path inside defaultEmbedFS where the YAML is
	// copied during build. The build script (or a Makefile target /
	// task recipe) copies compliance/classification.yaml into
	// services/packages/classification/embedded/ before compilation.
	// We keep the embed path stable so the embedded copy and the
	// source-of-truth file in compliance/ never drift silently —
	// tests assert the bytes match.
	embeddedPath = "embedded/classification.yaml"

	// supportedSchemaVersion is the YAML schema version this package
	// can parse. Bumping the YAML's `version:` requires a coordinated
	// release of this package.
	supportedSchemaVersion = 1
)

// Level is one of the canonical classification levels. Stable string
// constants — code MUST NOT compare integers directly; use Compare or
// AtLeast.
type Level string

// Canonical classification levels, ordered from least to most restrictive.
// See compliance/classification.yaml for the full semantics.
const (
	Public     Level = "public"
	Internal   Level = "internal"
	Restricted Level = "restricted"
	CUI        Level = "cui"
	ITAR       Level = "itar"
)

// Scheme is the parsed classification.yaml. Returned by Load /
// LoadFromFS; treat as immutable.
type Scheme struct {
	Version          int               `yaml:"version"`
	Levels           []LevelDefinition `yaml:"levels"`
	Default          Level             `yaml:"default"`
	ResourceDefaults map[string]Level  `yaml:"resource_defaults"`
	EgressChannels   map[string]EgressRule `yaml:"egress_channels"`
	AuditRetention   map[Level]int     `yaml:"audit_retention_online_days"`
	Labels           Labels            `yaml:"labels"`

	// ord maps level name → ordinal for fast comparison. Populated by
	// validate(); not parsed from YAML.
	ord map[Level]int
}

// LevelDefinition is one row in the YAML levels list.
type LevelDefinition struct {
	Name        Level  `yaml:"name"`
	Level       int    `yaml:"level"`
	Description string `yaml:"description"`
	Default     bool   `yaml:"default,omitempty"`
}

// EgressRule constrains the maximum classification a payload may carry
// over a given egress channel.
type EgressRule struct {
	AllowedMax Level `yaml:"allowed_max"`
}

// Labels carries the conventions service code consults when emitting
// container images / Kafka topics / Postgres rows.
type Labels struct {
	ContainerImageLabelKey   string `yaml:"container_image_label_key"`
	KafkaTopicPrefixTemplate string `yaml:"kafka_topic_prefix_template"`
	PostgresColumnName       string `yaml:"postgres_column_name"`
}

// once + cached cache the loaded Scheme so repeated Load calls are
// cheap. Tests use LoadFromFS to bypass the cache.
var (
	once   sync.Once
	cached *Scheme
	loadEr error
)

// Load returns the embedded Scheme, parsed once per process. Service
// boot code calls this once and passes the resulting *Scheme through
// the dependency graph.
func Load() (*Scheme, error) {
	once.Do(func() {
		s, err := LoadFromFS(defaultEmbedFS, embeddedPath)
		cached, loadEr = s, err
	})
	return cached, loadEr
}

// LoadFromFS parses the classification YAML at `path` inside `fsys`.
// Tests use this with testing/fstest.MapFS to inject fixtures.
func LoadFromFS(fsys fs.FS, path string) (*Scheme, error) {
	raw, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, fmt.Errorf("classification: read %s: %w", path, err)
	}
	var s Scheme
	if err := yaml.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("classification: parse %s: %w", path, err)
	}
	if err := s.validate(); err != nil {
		return nil, err
	}
	return &s, nil
}

// validate checks the parsed scheme is internally consistent and
// matches this package's expectations. Called once during Load.
func (s *Scheme) validate() error {
	if s.Version != supportedSchemaVersion {
		return fmt.Errorf("classification: schema version %d not supported (this package supports v%d)",
			s.Version, supportedSchemaVersion)
	}
	if len(s.Levels) == 0 {
		return errors.New("classification: no levels defined")
	}

	s.ord = make(map[Level]int, len(s.Levels))
	seen := map[int]bool{}
	defaultCount := 0
	for _, l := range s.Levels {
		if l.Name == "" {
			return errors.New("classification: level with empty name")
		}
		if _, dup := s.ord[l.Name]; dup {
			return fmt.Errorf("classification: duplicate level name %q", l.Name)
		}
		if seen[l.Level] {
			return fmt.Errorf("classification: duplicate level ordinal %d for %q", l.Level, l.Name)
		}
		seen[l.Level] = true
		s.ord[l.Name] = l.Level
		if l.Default {
			defaultCount++
		}
	}
	if defaultCount > 1 {
		return errors.New("classification: more than one level marked default")
	}

	if s.Default == "" {
		return errors.New("classification: top-level `default` must be set")
	}
	if _, ok := s.ord[s.Default]; !ok {
		return fmt.Errorf("classification: top-level default %q is not a defined level", s.Default)
	}

	for resource, lvl := range s.ResourceDefaults {
		if _, ok := s.ord[lvl]; !ok {
			return fmt.Errorf("classification: resource_defaults[%q] = %q is not a defined level", resource, lvl)
		}
	}
	for ch, rule := range s.EgressChannels {
		if _, ok := s.ord[rule.AllowedMax]; !ok {
			return fmt.Errorf("classification: egress_channels[%q].allowed_max = %q is not a defined level",
				ch, rule.AllowedMax)
		}
	}
	for lvl := range s.AuditRetention {
		if _, ok := s.ord[lvl]; !ok {
			return fmt.Errorf("classification: audit_retention_online_days[%q] is not a defined level", lvl)
		}
	}

	if s.Labels.ContainerImageLabelKey == "" ||
		s.Labels.KafkaTopicPrefixTemplate == "" ||
		s.Labels.PostgresColumnName == "" {
		return errors.New("classification: labels.* keys must all be set")
	}

	return nil
}

// Ordinal returns the integer rank of a level. Returns -1 (and false)
// if the level is unknown.
func (s *Scheme) Ordinal(l Level) (int, bool) {
	o, ok := s.ord[l]
	return o, ok
}

// Compare returns -1, 0, +1 according to whether `a` is below, equal
// to, or above `b` in the classification ordering. Unknown levels
// compare as -1 — the caller should treat unknown levels as the most
// permissive (least restrictive) interpretation, which is wrong in
// most contexts; prefer KnowsLevel + explicit handling.
func (s *Scheme) Compare(a, b Level) int {
	ao := s.ord[a]
	bo := s.ord[b]
	switch {
	case ao < bo:
		return -1
	case ao > bo:
		return 1
	default:
		return 0
	}
}

// KnowsLevel reports whether the level name is defined in the scheme.
func (s *Scheme) KnowsLevel(l Level) bool {
	_, ok := s.ord[l]
	return ok
}

// AtLeast reports whether `clearance` satisfies the requirement to
// access data classified at `required`. Equivalent to
// Compare(clearance, required) >= 0 with explicit unknown handling.
func (s *Scheme) AtLeast(clearance, required Level) bool {
	if !s.KnowsLevel(clearance) || !s.KnowsLevel(required) {
		return false
	}
	return s.Compare(clearance, required) >= 0
}

// MaxAllowedFor returns the AllowedMax for a named egress channel.
// Returns ("", false) when the channel is not configured; the caller
// should refuse to serialise the payload in that case.
func (s *Scheme) MaxAllowedFor(channel string) (Level, bool) {
	rule, ok := s.EgressChannels[channel]
	if !ok {
		return "", false
	}
	return rule.AllowedMax, true
}

// AllowEgress reports whether a payload classified at `payload` may be
// sent over the named channel.
func (s *Scheme) AllowEgress(channel string, payload Level) bool {
	max, ok := s.MaxAllowedFor(channel)
	if !ok {
		return false
	}
	return s.AtLeast(max, payload)
}

// DefaultFor returns the resource-specific default classification, or
// the global default when the resource kind is not specially
// configured.
func (s *Scheme) DefaultFor(resourceKind string) Level {
	if l, ok := s.ResourceDefaults[resourceKind]; ok {
		return l
	}
	return s.Default
}

// AuditRetentionDays returns the online audit retention requirement in
// days for the supplied classification.
func (s *Scheme) AuditRetentionDays(l Level) (int, bool) {
	d, ok := s.AuditRetention[l]
	return d, ok
}

// LevelNames returns every level name in ascending ordinal order. Used
// by tests + the YAML round-trip helpers.
func (s *Scheme) LevelNames() []Level {
	out := make([]Level, 0, len(s.Levels))
	for _, l := range s.Levels {
		out = append(out, l.Name)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return s.ord[out[i]] < s.ord[out[j]]
	})
	return out
}

// ParseLevel converts a free-form string (case-insensitive, trims
// whitespace) into a typed Level value. Returns ("", false) if the
// string is not a known level.
func ParseLevel(s string) (Level, bool) {
	canon := Level(strings.ToLower(strings.TrimSpace(s)))
	switch canon {
	case Public, Internal, Restricted, CUI, ITAR:
		return canon, true
	}
	return "", false
}
