// Package profile loads + validates SpacecraftProfile records.
//
// → REQ-FUNC-SAT-001 (every service's runtime behaviour is configured
//   from the spacecraft profile).
// → design.md §4.5
//
// The Go types in this file are the runtime mirror of
// services/packages/proto/space/satellite/v1/profile.proto. They are
// authored by hand (not generated) so the loader works in any
// chetana service without depending on the generated .pb.go.
//
// When `buf generate` runs in CI a parallel
// services/packages/api/v1/satellite/satellite.pb.go appears; consumers
// that need full proto-reflect features (e.g. Connect handlers
// returning a SpacecraftProfile RPC reply) import that one. Consumers
// that just want to read a YAML profile use this package.
//
// The two type sets MUST stay structurally compatible. profile_test.go
// includes a structural-equivalence check that runs whenever
// satellite.pb.go is generated.
package profile

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"
)

// ----------------------------------------------------------------------
// Type mirror of profile.proto
// ----------------------------------------------------------------------

// Band mirrors satellite.v1.Band.
type Band string

// Canonical Band values; mirror profile.proto and hardware.Band.
const (
	BandUHF Band = "UHF"
	BandS   Band = "S"
	BandX   Band = "X"
)

// Modulation mirrors satellite.v1.Modulation.
type Modulation string

// Canonical Modulation values; mirror profile.proto and hardware.Modulation.
const (
	ModBPSK  Modulation = "BPSK"
	ModQPSK  Modulation = "QPSK"
	ModOQPSK Modulation = "OQPSK"
	Mod8PSK  Modulation = "8PSK"
	ModGMSK  Modulation = "GMSK"
)

// CcsdsProfile mirrors satellite.v1.CcsdsProfile.
type CcsdsProfile string

// Canonical CcsdsProfile values per CCSDS 132.0-B-2 (TM_TF), 232.0-B-3
// (TC_TF), 732.0-B-3 (AOS), and 732.1-B-2 (USLP).
const (
	CcsdsTMTF CcsdsProfile = "TM_TF"
	CcsdsTCTF CcsdsProfile = "TC_TF"
	CcsdsAOS  CcsdsProfile = "AOS"
	CcsdsUSLP CcsdsProfile = "USLP"
)

// SubsystemKind mirrors satellite.v1.Subsystem.Kind.
type SubsystemKind string

// Canonical SubsystemKind values. Each entry maps to one subsystem
// family used by mission ops dashboards and the rule engine.
const (
	SubsystemPower      SubsystemKind = "POWER"
	SubsystemADCS       SubsystemKind = "ADCS"
	SubsystemCDH        SubsystemKind = "CDH"
	SubsystemComms      SubsystemKind = "COMMS"
	SubsystemThermal    SubsystemKind = "THERMAL"
	SubsystemPropulsion SubsystemKind = "PROPULSION"
	SubsystemPayload    SubsystemKind = "PAYLOAD"
	SubsystemStructure  SubsystemKind = "STRUCTURE"
)

// SpacecraftProfile is the YAML-loaded mirror of satellite.v1.SpacecraftProfile.
type SpacecraftProfile struct {
	ProfileID          string         `yaml:"profile_id"          json:"profile_id"`
	SpacecraftID       string         `yaml:"spacecraft_id"       json:"spacecraft_id"`
	BusType            string         `yaml:"bus_type"            json:"bus_type"`
	Bands              []Band         `yaml:"bands"               json:"bands"`
	Modulations        []Modulation   `yaml:"modulations"         json:"modulations"`
	CcsdsProfiles      []CcsdsProfile `yaml:"ccsds_profiles"      json:"ccsds_profiles"`
	LinkBudget         LinkBudget     `yaml:"link_budget"         json:"link_budget"`
	SafetyModes        []SafetyMode   `yaml:"safety_modes"        json:"safety_modes"`
	Subsystems         []Subsystem    `yaml:"subsystems"          json:"subsystems"`
	ItarClassification string         `yaml:"itar_classification" json:"itar_classification"`
	Version            string         `yaml:"version"             json:"version"`
	EffectiveAt        time.Time      `yaml:"effective_at"        json:"effective_at"`
}

// LinkBudget mirrors satellite.v1.LinkBudget.
type LinkBudget struct {
	EirpDbw         float64 `yaml:"eirp_dbw"           json:"eirp_dbw"`
	AntennaGainDb   float64 `yaml:"antenna_gain_db"    json:"antenna_gain_db"`
	NoiseTempK      float64 `yaml:"noise_temp_k"       json:"noise_temp_k"`
	BitRateBps      float64 `yaml:"bit_rate_bps"       json:"bit_rate_bps"`
	RequiredEbN0Db  float64 `yaml:"required_eb_n0_db"  json:"required_eb_n0_db"`
}

// SafetyMode mirrors satellite.v1.SafetyMode.
type SafetyMode struct {
	Name           string `yaml:"name"            json:"name"`
	EntryCriteria  string `yaml:"entry_criteria"  json:"entry_criteria"`
	RecoveryAction string `yaml:"recovery_action" json:"recovery_action"`
}

// Subsystem mirrors satellite.v1.Subsystem.
type Subsystem struct {
	Kind         SubsystemKind `yaml:"kind"          json:"kind"`
	Name         string        `yaml:"name"          json:"name"`
	RedundancyN  int           `yaml:"redundancy_n"  json:"redundancy_n"`
	VendorName   string        `yaml:"vendor_name"   json:"vendor_name"`
	ModelNumber  string        `yaml:"model_number"  json:"model_number"`
}

// ----------------------------------------------------------------------
// Validation
// ----------------------------------------------------------------------

// Validate enforces every invariant the profile must satisfy at boot
// time. Returns a joined error containing every violation so the
// operator can fix them all in one pass.
func (p *SpacecraftProfile) Validate() error {
	var errs []error

	if p.ProfileID == "" {
		errs = append(errs, errors.New("profile_id is required"))
	}
	if p.SpacecraftID == "" {
		errs = append(errs, errors.New("spacecraft_id is required"))
	}
	if p.BusType == "" {
		errs = append(errs, errors.New("bus_type is required"))
	}
	if p.Version == "" {
		errs = append(errs, errors.New("version is required"))
	}
	if len(p.Bands) == 0 {
		errs = append(errs, errors.New("at least one band is required"))
	}
	for i, b := range p.Bands {
		if !validBand(b) {
			errs = append(errs, fmt.Errorf("bands[%d] = %q is not a valid Band (one of: UHF, S, X)", i, b))
		}
	}
	if len(p.Modulations) == 0 {
		errs = append(errs, errors.New("at least one modulation is required"))
	}
	for i, m := range p.Modulations {
		if !validModulation(m) {
			errs = append(errs, fmt.Errorf("modulations[%d] = %q is not a valid Modulation (one of: BPSK, QPSK, OQPSK, 8PSK, GMSK)", i, m))
		}
	}
	if len(p.CcsdsProfiles) == 0 {
		errs = append(errs, errors.New("at least one ccsds_profile is required"))
	}
	for i, c := range p.CcsdsProfiles {
		if !validCcsds(c) {
			errs = append(errs, fmt.Errorf("ccsds_profiles[%d] = %q is not a valid CcsdsProfile (one of: TM_TF, TC_TF, AOS, USLP)", i, c))
		}
	}
	for i, s := range p.Subsystems {
		if !validSubsystemKind(s.Kind) {
			errs = append(errs, fmt.Errorf("subsystems[%d].kind = %q is not a valid SubsystemKind", i, s.Kind))
		}
		if s.Name == "" {
			errs = append(errs, fmt.Errorf("subsystems[%d].name is required", i))
		}
		if s.RedundancyN < 1 {
			errs = append(errs, fmt.Errorf("subsystems[%d].redundancy_n must be >= 1", i))
		}
	}
	if !validClassification(p.ItarClassification) {
		errs = append(errs, fmt.Errorf("itar_classification = %q must be one of: public, itar", p.ItarClassification))
	}
	if err := p.LinkBudget.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("link_budget: %w", err))
	}
	if p.EffectiveAt.IsZero() {
		errs = append(errs, errors.New("effective_at is required"))
	}

	return joinErrors(errs)
}

// Validate enforces link-budget sanity. Real link budgets are signed
// off by RF engineering; this guard catches obviously-bad values
// that would corrupt downstream calculations.
func (b *LinkBudget) Validate() error {
	var errs []error
	if b.BitRateBps <= 0 {
		errs = append(errs, errors.New("bit_rate_bps must be > 0"))
	}
	if b.NoiseTempK <= 0 {
		errs = append(errs, errors.New("noise_temp_k must be > 0"))
	}
	if b.AntennaGainDb < 0 || b.AntennaGainDb > 100 {
		errs = append(errs, errors.New("antenna_gain_db must be in [0, 100]"))
	}
	// Required Eb/N0 for any practical modulation is 0..40 dB.
	if b.RequiredEbN0Db < 0 || b.RequiredEbN0Db > 40 {
		errs = append(errs, errors.New("required_eb_n0_db must be in [0, 40]"))
	}
	// EIRP can be negative (very low-power CubeSats); cap upper end
	// to flag obvious typos.
	if b.EirpDbw > 100 {
		errs = append(errs, errors.New("eirp_dbw > 100 dBW is implausible — likely a typo"))
	}
	return joinErrors(errs)
}

// ----------------------------------------------------------------------
// Validation helpers
// ----------------------------------------------------------------------

func validBand(b Band) bool {
	switch b {
	case BandUHF, BandS, BandX:
		return true
	}
	return false
}

func validModulation(m Modulation) bool {
	switch m {
	case ModBPSK, ModQPSK, ModOQPSK, Mod8PSK, ModGMSK:
		return true
	}
	return false
}

func validCcsds(c CcsdsProfile) bool {
	switch c {
	case CcsdsTMTF, CcsdsTCTF, CcsdsAOS, CcsdsUSLP:
		return true
	}
	return false
}

func validSubsystemKind(k SubsystemKind) bool {
	switch k {
	case SubsystemPower, SubsystemADCS, SubsystemCDH, SubsystemComms,
		SubsystemThermal, SubsystemPropulsion, SubsystemPayload, SubsystemStructure:
		return true
	}
	return false
}

func validClassification(s string) bool {
	switch strings.ToLower(s) {
	case "public", "itar":
		return true
	}
	return false
}

// joinErrors collapses a slice of errors into a single multi-line
// error, or returns nil when the slice is empty.
func joinErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	parts := make([]string, len(errs))
	for i, e := range errs {
		parts[i] = "  - " + e.Error()
	}
	return fmt.Errorf("profile validation failed:\n%s", strings.Join(parts, "\n"))
}

// ----------------------------------------------------------------------
// Loader convenience helpers (file system access). The actual YAML
// (un)marshalling lives in loader.go so this file stays YAML-agnostic.
// ----------------------------------------------------------------------

// readFile is a tiny indirection so tests can swap in a mock fs.
type readFileFn func(name string) ([]byte, error)

var defaultReadFile readFileFn = os.ReadFile

// readFromFS reads the file at `path` from the supplied fs; falls
// back to defaultReadFile when fsys is nil. Used by LoadFile +
// LoadFromFS.
func readFromFS(fsys fs.FS, path string) ([]byte, error) {
	if fsys == nil {
		return defaultReadFile(path)
	}
	return fs.ReadFile(fsys, path)
}
