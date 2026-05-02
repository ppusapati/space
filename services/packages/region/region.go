// Package region centralises the region-aware naming conventions every
// chetana service needs. v1 deploys to a single GovCloud cluster
// (REQ-CONST-003) but every data-plane resource (Postgres, S3, Kafka)
// is addressed through this package so multi-region rollout in v1.x is
// pure configuration — no code changes (REQ-NFR-SCALE-003,
// design.md §4.8).
//
// Conventions:
//
//   • Postgres host:    <region>.db.chetana.internal
//   • S3 bucket:        chetana-<region>-<prefix>
//   • Kafka bootstrap:  <region>.kafka.chetana.internal:9094
//   • Audit region tag: the value of CHETANA_REGION (verbatim)
//
// All helpers take the region from CHETANA_REGION at process start.
// Tests should use ResolveOverride to inject a region without mutating
// the environment.
package region

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// envVar is the canonical environment variable that every chetana
// process reads to determine its region.
const envVar = "CHETANA_REGION"

// defaultRegion is the v1 cluster region per REQ-CONST-003.
const defaultRegion = "us-gov-east-1"

// Region is a strongly-typed region identifier. We use a named string
// instead of an enum so future regions can be added without recompiling
// every dependent service.
type Region string

// Known regions called out in plan/design.md §4.8. These are constants
// for readability; arbitrary Region values are valid (any string that
// passes Validate).
const (
	USGovEast1   Region = "us-gov-east-1"
	EUCentral1   Region = "eu-central-1"
	APSouth1     Region = "ap-south-1"
)

// String implements fmt.Stringer.
func (r Region) String() string { return string(r) }

// validRegionPattern enforces the AWS-style "<continent>-<area>-<n>" shape
// without depending on `regexp` at the top level (cheap call site, hot at
// boot only).
//
// Accepts any string of the form "<lc>+(-<lc>+)+-<digit>+" where <lc>
// is a lowercase ASCII letter. Examples: us-gov-east-1, eu-central-1,
// ap-south-1, ap-southeast-2.
func validRegion(r string) bool {
	if r == "" {
		return false
	}
	parts := strings.Split(r, "-")
	if len(parts) < 3 {
		return false
	}
	// Last part must be one or more digits.
	last := parts[len(parts)-1]
	if last == "" {
		return false
	}
	for _, b := range last {
		if b < '0' || b > '9' {
			return false
		}
	}
	// All other parts must be lowercase ASCII letters, non-empty.
	for _, p := range parts[:len(parts)-1] {
		if p == "" {
			return false
		}
		for _, b := range p {
			if b < 'a' || b > 'z' {
				return false
			}
		}
	}
	return true
}

// Validate returns an error when the region identifier does not match
// the expected shape. Useful for service entrypoints that want to fail
// fast on misconfiguration.
func Validate(r string) error {
	if !validRegion(r) {
		return fmt.Errorf("region: invalid identifier %q (expected lowercase <continent>-<area>-<n>)", r)
	}
	return nil
}

// once / cached / overrideMu guard the lazy resolution of the active
// region so unit tests can override it without touching the
// environment.
var (
	once     sync.Once
	cached   Region
	override sync.Mutex
	mocked   *Region
)

// ResolveOverride sets the region returned by Active for the duration
// of the test; the returned function restores the previous binding.
// Production code MUST NOT call this.
func ResolveOverride(r Region) func() {
	override.Lock()
	defer override.Unlock()
	prev := mocked
	rr := r
	mocked = &rr
	return func() {
		override.Lock()
		defer override.Unlock()
		mocked = prev
	}
}

// Active returns the region the process was started in. It reads
// CHETANA_REGION on first call and caches the result; subsequent calls
// are lock-free fast paths. Tests bypass the cache via ResolveOverride.
//
// If CHETANA_REGION is unset Active returns USGovEast1 (the v1 default
// per REQ-CONST-003) and emits no error — local-dev workflows do not
// need to set the env var. If CHETANA_REGION is set but malformed,
// Active panics; failing fast at boot is preferred over silently using
// a wrong bucket.
func Active() Region {
	override.Lock()
	if mocked != nil {
		r := *mocked
		override.Unlock()
		return r
	}
	override.Unlock()

	once.Do(func() {
		raw := strings.TrimSpace(os.Getenv(envVar))
		if raw == "" {
			cached = defaultRegion
			return
		}
		if !validRegion(raw) {
			panic(fmt.Sprintf("region: %s=%q is not a valid region identifier", envVar, raw))
		}
		cached = Region(raw)
	})
	return cached
}

// PostgresDSN returns the canonical Postgres DSN for the active region
// using the supplied database name. The host follows the
// "<region>.db.chetana.internal" convention; user/password are read
// from CHETANA_DB_USER / CHETANA_DB_PASSWORD (defaults p9e/p9e for
// local dev). Per REQ-NFR-SEC-002 the connection always asks for TLS;
// docker-compose Postgres ignores the request, RDS enforces it.
func PostgresDSN(database string) string {
	r := Active()
	user := envOrDefault("CHETANA_DB_USER", "p9e")
	pass := envOrDefault("CHETANA_DB_PASSWORD", "p9e")
	host := envOrDefault("CHETANA_DB_HOST", string(r)+".db.chetana.internal")
	port := envOrDefault("CHETANA_DB_PORT", "5432")
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=prefer",
		user, pass, host, port, database,
	)
}

// S3Bucket returns the bucket name for the supplied logical prefix in
// the active region. Examples:
//
//   region.S3Bucket("eo-scenes")  -> "chetana-us-gov-east-1-eo-scenes"
//   region.S3Bucket("audit-cold") -> "chetana-us-gov-east-1-audit-cold"
//
// AWS bucket names are global; the region tag in the bucket name keeps
// per-region buckets unambiguous and aligns with the design.md §5.2
// convention.
func S3Bucket(prefix string) string {
	r := Active()
	return fmt.Sprintf("chetana-%s-%s", r, prefix)
}

// KafkaBootstrap returns the bootstrap address for the active region's
// Kafka cluster. v1 deploys MSK in GovCloud; the convention scales to
// future regions transparently.
//
// Caller may override via CHETANA_KAFKA_BOOTSTRAP for local dev (e.g.
// "localhost:9092" against the docker-compose Kafka).
func KafkaBootstrap() string {
	if v := strings.TrimSpace(os.Getenv("CHETANA_KAFKA_BOOTSTRAP")); v != "" {
		return v
	}
	r := Active()
	return string(r) + ".kafka.chetana.internal:9094"
}

// envOrDefault reads an env var, returning def when unset/empty.
func envOrDefault(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
