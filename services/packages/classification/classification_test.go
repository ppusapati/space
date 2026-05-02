package classification

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// TestLoad_EmbeddedSchemeIsValid verifies the YAML embedded in this
// package via //go:embed parses, validates, and exposes every level
// the spec calls out (TASK-P0-COMP-001 acceptance criterion #3).
func TestLoad_EmbeddedSchemeIsValid(t *testing.T) {
	s, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	wantLevels := []Level{Public, Internal, Restricted, CUI, ITAR}
	for _, l := range wantLevels {
		if !s.KnowsLevel(l) {
			t.Errorf("expected level %q in scheme", l)
		}
	}
	if s.Default != Internal {
		t.Errorf("expected default=internal; got %q", s.Default)
	}
	if got := s.LevelNames(); len(got) != len(wantLevels) {
		t.Errorf("LevelNames length = %d, want %d", len(got), len(wantLevels))
	}
}

// TestLoad_EmbeddedMatchesSourceOfTruth verifies the embedded YAML
// matches compliance/classification.yaml byte-for-byte. Regression
// guard against the embedded copy drifting.
func TestLoad_EmbeddedMatchesSourceOfTruth(t *testing.T) {
	embedded, err := defaultEmbedFS.ReadFile(embeddedPath)
	if err != nil {
		t.Fatalf("read embedded: %v", err)
	}
	source, err := os.ReadFile(filepath.Join("..", "..", "..", "compliance", "classification.yaml"))
	if err != nil {
		t.Skipf("source file not reachable from test working dir: %v", err)
	}
	if string(embedded) != string(source) {
		t.Errorf("embedded classification.yaml has drifted from compliance/classification.yaml — re-run the build script that copies it into services/packages/classification/embedded/")
	}
}

// TestCompare_EnforcesOrdering covers the comparator that authz.go
// will call to gate access (clearance >= required).
func TestCompare_EnforcesOrdering(t *testing.T) {
	s, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	cases := []struct {
		a, b Level
		want int
	}{
		{Public, Public, 0},
		{Public, Internal, -1},
		{Internal, Public, 1},
		{ITAR, CUI, 1},
		{CUI, ITAR, -1},
		{ITAR, ITAR, 0},
	}
	for _, c := range cases {
		t.Run(string(c.a)+"_vs_"+string(c.b), func(t *testing.T) {
			if got := s.Compare(c.a, c.b); got != c.want {
				t.Errorf("Compare(%q,%q)=%d, want %d", c.a, c.b, got, c.want)
			}
		})
	}
}

// TestAtLeast covers the boolean clearance-meets-requirement helper
// that downstream interceptors use.
func TestAtLeast(t *testing.T) {
	s, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	cases := []struct {
		clearance, required Level
		want                bool
	}{
		{Public, Public, true},
		{Internal, Public, true},
		{ITAR, ITAR, true},
		{CUI, ITAR, false},
		{Public, ITAR, false},
		{ITAR, Public, true},
		{"unknown", Public, false},
	}
	for _, c := range cases {
		t.Run(string(c.clearance)+"_meets_"+string(c.required), func(t *testing.T) {
			if got := s.AtLeast(c.clearance, c.required); got != c.want {
				t.Errorf("AtLeast(%q,%q)=%v, want %v", c.clearance, c.required, got, c.want)
			}
		})
	}
}

// TestAllowEgress covers the channel-cap rule. Public surface MUST
// reject CUI / ITAR; internal RPC MUST allow ITAR.
func TestAllowEgress(t *testing.T) {
	s, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	cases := []struct {
		channel string
		payload Level
		want    bool
	}{
		{"public_api_v1", Public, true},
		{"public_api_v1", Internal, false},
		{"public_api_v1", ITAR, false},
		{"internal_rpc", ITAR, true},
		{"notify_email", CUI, true},
		{"notify_email", ITAR, false},
		{"notify_sms", Internal, true},
		{"notify_sms", CUI, false},
		{"unknown_channel", Public, false}, // unknown defaults to deny
	}
	for _, c := range cases {
		t.Run(c.channel+"_"+string(c.payload), func(t *testing.T) {
			if got := s.AllowEgress(c.channel, c.payload); got != c.want {
				t.Errorf("AllowEgress(%q,%q)=%v, want %v", c.channel, c.payload, got, c.want)
			}
		})
	}
}

// TestDefaultFor exercises the resource-defaults override path.
func TestDefaultFor(t *testing.T) {
	s, err := Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	cases := map[string]Level{
		"audit_event":         CUI,
		"command":             ITAR,
		"telemetry_sample":    ITAR,
		"spacecraft_profile":  ITAR,
		"eo_scene":            Internal,
		"stac_item_public":    Public,
		"user":                CUI,
		"unknown_kind":        Internal, // falls back to global default
	}
	for kind, want := range cases {
		t.Run(kind, func(t *testing.T) {
			if got := s.DefaultFor(kind); got != want {
				t.Errorf("DefaultFor(%q)=%q, want %q", kind, got, want)
			}
		})
	}
}

// TestParseLevel exercises the input-string normaliser.
func TestParseLevel(t *testing.T) {
	cases := map[string]struct {
		want Level
		ok   bool
	}{
		"public":     {Public, true},
		"PUBLIC":     {Public, true},
		"  itar  ":   {ITAR, true},
		"Internal":   {Internal, true},
		"unknown":    {"", false},
		"":           {"", false},
	}
	for in, want := range cases {
		t.Run(in, func(t *testing.T) {
			got, ok := ParseLevel(in)
			if got != want.want || ok != want.ok {
				t.Errorf("ParseLevel(%q)=(%q,%v); want (%q,%v)", in, got, ok, want.want, want.ok)
			}
		})
	}
}

// TestLoadFromFS_RejectsBadSchema covers the validate() error paths
// using small fixtures via testing/fstest.MapFS.
func TestLoadFromFS_RejectsBadSchema(t *testing.T) {
	cases := map[string]string{
		"unsupported version": `version: 999
levels:
  - {name: public, level: 0}
default: public
labels: {container_image_label_key: x, kafka_topic_prefix_template: y, postgres_column_name: z}
`,
		"empty levels": `version: 1
levels: []
default: public
labels: {container_image_label_key: x, kafka_topic_prefix_template: y, postgres_column_name: z}
`,
		"duplicate level name": `version: 1
levels:
  - {name: public, level: 0}
  - {name: public, level: 1}
default: public
labels: {container_image_label_key: x, kafka_topic_prefix_template: y, postgres_column_name: z}
`,
		"default not defined": `version: 1
levels:
  - {name: public, level: 0}
default: itar
labels: {container_image_label_key: x, kafka_topic_prefix_template: y, postgres_column_name: z}
`,
		"resource_default unknown level": `version: 1
levels:
  - {name: public, level: 0}
default: public
resource_defaults: {audit_event: itar}
labels: {container_image_label_key: x, kafka_topic_prefix_template: y, postgres_column_name: z}
`,
		"missing labels block": `version: 1
levels:
  - {name: public, level: 0}
default: public
labels: {container_image_label_key: "", kafka_topic_prefix_template: "", postgres_column_name: ""}
`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			fsys := fstest.MapFS{
				"c.yaml": &fstest.MapFile{Data: []byte(body)},
			}
			_, err := LoadFromFS(fsys, "c.yaml")
			if err == nil {
				t.Fatalf("LoadFromFS should have errored for %q", name)
			}
			if !strings.Contains(err.Error(), "classification") {
				t.Errorf("error message lacks 'classification' prefix: %v", err)
			}
		})
	}
}
