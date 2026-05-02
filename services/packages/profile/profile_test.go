package profile_test

import (
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"p9e.in/chetana/packages/profile"
)

// validProfileYAML is the canonical happy-path YAML used by every
// happy-path test. Mirrors the fields a real CubeSat profile would
// carry.
const validProfileYAML = `
profile_id: 01HZA5XXXX0000000000000000
spacecraft_id: "44713"
bus_type: LEO-100kg-3U
bands: [UHF, S]
modulations: [BPSK, QPSK]
ccsds_profiles: [TM_TF, TC_TF]
link_budget:
  eirp_dbw: 1.5
  antenna_gain_db: 14
  noise_temp_k: 150
  bit_rate_bps: 9600
  required_eb_n0_db: 10
safety_modes:
  - name: sun-pointing
    entry_criteria: battery SoC < 30%
    recovery_action: ground command after SoC > 60%
subsystems:
  - kind: POWER
    name: solar+battery
    redundancy_n: 1
    vendor_name: GomSpace
    model_number: NanoPower P31u
  - kind: COMMS
    name: UHF transceiver
    redundancy_n: 1
    vendor_name: EnduroSat
    model_number: UHF-Type-II
itar_classification: public
version: 1.0.0
effective_at: 2026-05-02T00:00:00Z
`

// TestLoadBytes_HappyPath covers acceptance criterion #3
// (round-trip + validation succeed on a well-formed profile).
func TestLoadBytes_HappyPath(t *testing.T) {
	p, err := profile.LoadBytes([]byte(validProfileYAML))
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	if p.ProfileID != "01HZA5XXXX0000000000000000" {
		t.Errorf("profile_id=%q", p.ProfileID)
	}
	if len(p.Bands) != 2 || p.Bands[0] != profile.BandUHF || p.Bands[1] != profile.BandS {
		t.Errorf("bands=%v", p.Bands)
	}
	if len(p.Subsystems) != 2 {
		t.Errorf("subsystems count=%d", len(p.Subsystems))
	}
	if p.LinkBudget.BitRateBps != 9600 {
		t.Errorf("bit_rate_bps=%v", p.LinkBudget.BitRateBps)
	}
	if !p.EffectiveAt.Equal(time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("effective_at=%v", p.EffectiveAt)
	}
}

// TestLoadFromFS covers the fs.FS code path (used by services that
// embed profiles via embed.FS).
func TestLoadFromFS(t *testing.T) {
	fsys := fstest.MapFS{
		"profile.yaml": &fstest.MapFile{Data: []byte(validProfileYAML)},
	}
	p, err := profile.LoadFromFS(fsys, "profile.yaml")
	if err != nil {
		t.Fatalf("LoadFromFS: %v", err)
	}
	if p.ProfileID == "" {
		t.Error("ProfileID empty")
	}
}

// TestRoundTrip_YAML covers acceptance criterion #3 explicitly:
// parse → render → parse → DeepEqual.
func TestRoundTrip_YAML(t *testing.T) {
	first, err := profile.LoadBytes([]byte(validProfileYAML))
	if err != nil {
		t.Fatalf("first parse: %v", err)
	}
	rendered, err := profile.Marshal(first)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	second, err := profile.LoadBytes(rendered)
	if err != nil {
		t.Fatalf("second parse:\n%s\nerr=%v", string(rendered), err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Errorf("round-trip diff:\n--- first ---\n%+v\n--- second ---\n%+v", first, second)
	}
}

// TestValidation_TableDriven covers every validation rule. Each row
// mutates one field of the validProfileYAML and asserts the
// expected error substring appears.
func TestValidation_TableDriven(t *testing.T) {
	cases := []struct {
		name        string
		yaml        string
		wantSubstrs []string
	}{
		{
			name:        "missing profile_id",
			yaml:        replace(validProfileYAML, `profile_id: 01HZA5XXXX0000000000000000`, `profile_id: ""`),
			wantSubstrs: []string{"profile_id is required"},
		},
		{
			name:        "missing spacecraft_id",
			yaml:        replace(validProfileYAML, `spacecraft_id: "44713"`, `spacecraft_id: ""`),
			wantSubstrs: []string{"spacecraft_id is required"},
		},
		{
			name:        "missing bus_type",
			yaml:        replace(validProfileYAML, `bus_type: LEO-100kg-3U`, `bus_type: ""`),
			wantSubstrs: []string{"bus_type is required"},
		},
		{
			name:        "invalid band",
			yaml:        replace(validProfileYAML, `bands: [UHF, S]`, `bands: [UHF, Ka]`),
			wantSubstrs: []string{`bands[1] = "Ka"`},
		},
		{
			name:        "invalid modulation",
			yaml:        replace(validProfileYAML, `modulations: [BPSK, QPSK]`, `modulations: [BPSK, FSK]`),
			wantSubstrs: []string{`modulations[1] = "FSK"`},
		},
		{
			name:        "invalid ccsds profile",
			yaml:        replace(validProfileYAML, `ccsds_profiles: [TM_TF, TC_TF]`, `ccsds_profiles: [TM_TF, ARQ]`),
			wantSubstrs: []string{`ccsds_profiles[1] = "ARQ"`},
		},
		{
			name:        "invalid subsystem kind",
			yaml:        replace(validProfileYAML, `kind: POWER`, `kind: WARP_DRIVE`),
			wantSubstrs: []string{`kind = "WARP_DRIVE"`},
		},
		{
			name:        "subsystem missing name",
			yaml:        replace(validProfileYAML, `name: solar+battery`, `name: ""`),
			wantSubstrs: []string{".name is required"},
		},
		{
			name:        "subsystem zero redundancy",
			yaml:        replace(validProfileYAML, `redundancy_n: 1`, `redundancy_n: 0`),
			wantSubstrs: []string{".redundancy_n must be >= 1"},
		},
		{
			name:        "invalid classification",
			yaml:        replace(validProfileYAML, `itar_classification: public`, `itar_classification: secret`),
			wantSubstrs: []string{`itar_classification = "secret"`},
		},
		{
			name:        "missing version",
			yaml:        replace(validProfileYAML, `version: 1.0.0`, `version: ""`),
			wantSubstrs: []string{"version is required"},
		},
		{
			name:        "zero bit_rate_bps",
			yaml:        replace(validProfileYAML, `bit_rate_bps: 9600`, `bit_rate_bps: 0`),
			wantSubstrs: []string{"bit_rate_bps must be > 0"},
		},
		{
			name:        "zero noise_temp_k",
			yaml:        replace(validProfileYAML, `noise_temp_k: 150`, `noise_temp_k: 0`),
			wantSubstrs: []string{"noise_temp_k must be > 0"},
		},
		{
			name:        "antenna_gain_db too high",
			yaml:        replace(validProfileYAML, `antenna_gain_db: 14`, `antenna_gain_db: 999`),
			wantSubstrs: []string{"antenna_gain_db must be in [0, 100]"},
		},
		{
			name:        "required_eb_n0_db out of range",
			yaml:        replace(validProfileYAML, `required_eb_n0_db: 10`, `required_eb_n0_db: 100`),
			wantSubstrs: []string{"required_eb_n0_db must be in [0, 40]"},
		},
		{
			name:        "implausible eirp_dbw",
			yaml:        replace(validProfileYAML, `eirp_dbw: 1.5`, `eirp_dbw: 200`),
			wantSubstrs: []string{"eirp_dbw > 100 dBW"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := profile.LoadBytes([]byte(c.yaml))
			if err == nil {
				t.Fatalf("expected validation error containing %q; got nil", c.wantSubstrs)
			}
			for _, s := range c.wantSubstrs {
				if !strings.Contains(err.Error(), s) {
					t.Errorf("error missing substring %q\nfull err:\n%v", s, err)
				}
			}
		})
	}
}

// TestValidation_AggregatesAllErrors verifies that joinErrors returns
// every violation in one error, not just the first. Operators want to
// fix the whole profile in one editor pass.
func TestValidation_AggregatesAllErrors(t *testing.T) {
	bad := `
profile_id: ""
spacecraft_id: ""
bus_type: ""
bands: []
modulations: []
ccsds_profiles: []
link_budget:
  bit_rate_bps: -1
  noise_temp_k: 0
itar_classification: top-secret
version: ""
`
	_, err := profile.LoadBytes([]byte(bad))
	if err == nil {
		t.Fatal("expected aggregated error")
	}
	wantSubs := []string{
		"profile_id is required",
		"spacecraft_id is required",
		"bus_type is required",
		"version is required",
		"at least one band is required",
		"at least one modulation is required",
		"at least one ccsds_profile is required",
		"itar_classification",
		"effective_at is required",
	}
	for _, s := range wantSubs {
		if !strings.Contains(err.Error(), s) {
			t.Errorf("aggregated error missing substring %q\nfull err:\n%v", s, err)
		}
	}
}

// TestLoadBytes_RejectsInvalidYAML covers the parse-failure path.
func TestLoadBytes_RejectsInvalidYAML(t *testing.T) {
	_, err := profile.LoadBytes([]byte("not: yaml: at all: {{{"))
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "parse YAML") {
		t.Errorf("error message lacks 'parse YAML': %v", err)
	}
}

// TestMarshal_NilRejected covers the defensive-nil branch.
func TestMarshal_NilRejected(t *testing.T) {
	if _, err := profile.Marshal(nil); err == nil {
		t.Error("Marshal(nil) should error")
	}
}

// replace is a small helper that replaces a single occurrence of `old`
// with `new` in `s`. Used only by the validation table to mutate the
// canonical YAML one field at a time.
func replace(s, old, new string) string {
	return strings.Replace(s, old, new, 1)
}
