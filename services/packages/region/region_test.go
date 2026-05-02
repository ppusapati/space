package region

import (
	"os"
	"strings"
	"sync"
	"testing"
)

// TestActive_DefaultsWhenEnvUnset verifies that the v1 default
// (us-gov-east-1) is returned when CHETANA_REGION is not set.
// REQ-CONST-003 requires the platform to default to GovCloud.
func TestActive_DefaultsWhenEnvUnset(t *testing.T) {
	resetActive(t)
	t.Setenv(envVar, "")
	if got := Active(); got != USGovEast1 {
		t.Fatalf("Active() = %q, want %q", got, USGovEast1)
	}
}

// TestActive_ReadsEnvVar covers the three regions called out in
// design.md §4.8: us-gov-east-1, eu-central-1, ap-south-1.
func TestActive_ReadsEnvVar(t *testing.T) {
	cases := map[string]Region{
		"us-gov-east-1": USGovEast1,
		"eu-central-1":  EUCentral1,
		"ap-south-1":    APSouth1,
	}
	for raw, want := range cases {
		t.Run(raw, func(t *testing.T) {
			resetActive(t)
			t.Setenv(envVar, raw)
			if got := Active(); got != want {
				t.Errorf("Active() = %q, want %q", got, want)
			}
		})
	}
}

// TestActive_PanicsOnInvalidRegion exercises the fail-fast path. The
// alternative — silently falling back to default — would let services
// boot in the wrong region after a typo, sending traffic to the wrong
// data plane. Crashing at boot is the correct behaviour.
func TestActive_PanicsOnInvalidRegion(t *testing.T) {
	resetActive(t)
	t.Setenv(envVar, "GARBAGE")
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Active() with invalid CHETANA_REGION should panic")
		}
	}()
	Active()
}

// TestResolveOverride verifies the test hook restores its previous
// binding so test isolation is preserved.
func TestResolveOverride(t *testing.T) {
	resetActive(t)
	t.Setenv(envVar, string(USGovEast1))
	if got := Active(); got != USGovEast1 {
		t.Fatalf("baseline Active() = %q, want %q", got, USGovEast1)
	}

	restore := ResolveOverride(EUCentral1)
	if got := Active(); got != EUCentral1 {
		t.Errorf("under override: Active() = %q, want %q", got, EUCentral1)
	}
	restore()
	if got := Active(); got != USGovEast1 {
		t.Errorf("after restore: Active() = %q, want %q", got, USGovEast1)
	}
}

// TestPostgresDSN_RegionInHost verifies the host follows the
// "<region>.db.chetana.internal" convention and that env-var overrides
// take precedence.
func TestPostgresDSN_RegionInHost(t *testing.T) {
	resetActive(t)
	t.Setenv(envVar, string(EUCentral1))
	t.Setenv("CHETANA_DB_HOST", "")
	t.Setenv("CHETANA_DB_USER", "")
	t.Setenv("CHETANA_DB_PASSWORD", "")
	t.Setenv("CHETANA_DB_PORT", "")
	got := PostgresDSN("audit")
	want := "postgres://p9e:p9e@eu-central-1.db.chetana.internal:5432/audit?sslmode=prefer"
	if got != want {
		t.Errorf("PostgresDSN:\n got=%q\nwant=%q", got, want)
	}
}

func TestPostgresDSN_HostOverride(t *testing.T) {
	resetActive(t)
	t.Setenv(envVar, string(USGovEast1))
	t.Setenv("CHETANA_DB_HOST", "my-rds.example.com")
	t.Setenv("CHETANA_DB_USER", "svc_audit")
	t.Setenv("CHETANA_DB_PASSWORD", "secret")
	t.Setenv("CHETANA_DB_PORT", "6432")
	got := PostgresDSN("audit")
	want := "postgres://svc_audit:secret@my-rds.example.com:6432/audit?sslmode=prefer"
	if got != want {
		t.Errorf("PostgresDSN:\n got=%q\nwant=%q", got, want)
	}
}

// TestS3Bucket_RegionInName covers the "chetana-<region>-<prefix>"
// convention from design.md §5.2.
func TestS3Bucket_RegionInName(t *testing.T) {
	resetActive(t)
	t.Setenv(envVar, string(APSouth1))
	cases := map[string]string{
		"eo-scenes":   "chetana-ap-south-1-eo-scenes",
		"audit-cold":  "chetana-ap-south-1-audit-cold",
		"exports":     "chetana-ap-south-1-exports",
	}
	for prefix, want := range cases {
		t.Run(prefix, func(t *testing.T) {
			if got := S3Bucket(prefix); got != want {
				t.Errorf("S3Bucket(%q) = %q, want %q", prefix, got, want)
			}
		})
	}
}

// TestKafkaBootstrap_RegionInHost verifies the canonical address; the
// override branch exercises the local-dev path.
func TestKafkaBootstrap_RegionInHost(t *testing.T) {
	resetActive(t)
	t.Setenv(envVar, string(USGovEast1))
	t.Setenv("CHETANA_KAFKA_BOOTSTRAP", "")
	if got := KafkaBootstrap(); got != "us-gov-east-1.kafka.chetana.internal:9094" {
		t.Errorf("KafkaBootstrap default = %q", got)
	}
}

func TestKafkaBootstrap_Override(t *testing.T) {
	resetActive(t)
	t.Setenv("CHETANA_KAFKA_BOOTSTRAP", "localhost:9092")
	if got := KafkaBootstrap(); got != "localhost:9092" {
		t.Errorf("KafkaBootstrap override = %q, want localhost:9092", got)
	}
}

// TestValidate_TableDriven covers the regex-equivalent shape check.
// Adding a region in the future = appending a row here + a value in
// the const block above.
func TestValidate_TableDriven(t *testing.T) {
	cases := map[string]bool{
		"us-gov-east-1":     true,
		"eu-central-1":      true,
		"ap-south-1":        true,
		"ap-southeast-2":    true,
		"":                  false,
		"us":                false,
		"us-1":              false,
		"US-GOV-EAST-1":     false, // uppercase rejected
		"us-gov-east":       false, // missing trailing digit
		"us--east-1":        false, // empty middle segment
		"us_gov_east_1":     false, // underscore rejected
		"us-gov-east-1a":    false, // suffix non-digit
		"us-gov-east-1.0":   false, // dot rejected
	}
	for raw, valid := range cases {
		t.Run(raw, func(t *testing.T) {
			err := Validate(raw)
			if valid && err != nil {
				t.Errorf("Validate(%q) returned %v, want nil", raw, err)
			}
			if !valid && err == nil {
				t.Errorf("Validate(%q) returned nil, want error", raw)
			}
			if !valid && err != nil && !strings.Contains(err.Error(), "invalid identifier") {
				t.Errorf("Validate(%q) error message lacks 'invalid identifier': %q", raw, err)
			}
		})
	}
}

// resetActive resets the lazy `once` cache so each test sees a fresh
// CHETANA_REGION read. Necessary because Active() caches the result on
// first call; without this the test order would matter.
func resetActive(t *testing.T) {
	t.Helper()
	override.Lock()
	mocked = nil
	override.Unlock()
	once = sync.Once{}
	cached = ""
	// Clear all the related env vars to a known state. t.Setenv handles
	// the per-test restore for any subsequent t.Setenv calls.
	for _, k := range []string{envVar, "CHETANA_DB_HOST", "CHETANA_DB_USER", "CHETANA_DB_PASSWORD", "CHETANA_DB_PORT", "CHETANA_KAFKA_BOOTSTRAP"} {
		_ = os.Unsetenv(k)
	}
}
