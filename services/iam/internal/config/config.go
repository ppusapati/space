// Package config loads IAM service configuration via env vars +
// region helper. Phase 1; future tasks add structured config files
// (TASK-P1-IAM-005 OIDC clients, TASK-P1-IAM-006 SAML IdPs) which
// will move to YAML.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"p9e.in/chetana/packages/region"
)

// Config holds the runtime configuration for the IAM service.
type Config struct {
	HTTPAddr        string
	MetricsAddr     string
	ShutdownTimeout time.Duration

	// TenantID is the active tenant in single-tenant runtime.
	// Multi-tenant lookup lands in v1.x.
	TenantID string

	// Database is the Postgres DSN. Defaults to region.PostgresDSN("iam").
	DatabaseDSN string

	// RedisAddr is the Redis bootstrap address used by the rate
	// limiter. Defaults to "<region>.redis.chetana.internal:6379"
	// via region helper convention; dev override via CHETANA_REDIS_ADDR.
	RedisAddr string

	// Build identity injected at link time via -ldflags.
	Version string
	GitSHA  string
}

// Load reads config from env vars + region defaults.
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:        getenvOr("HTTP_ADDR", ":8080"),
		MetricsAddr:     getenvOr("METRICS_ADDR", ":9090"),
		ShutdownTimeout: 30 * time.Second,
		TenantID:        getenvOr("CHETANA_TENANT_ID", "00000000-0000-0000-0000-000000000001"),
		DatabaseDSN:     getenvOr("DATABASE_URL", region.PostgresDSN("iam")),
		RedisAddr:       getenvOr("CHETANA_REDIS_ADDR", region.Active().String()+".redis.chetana.internal:6379"),
		Version:         getenvOr("CHETANA_VERSION", "v0.0.0-dev"),
		GitSHA:          getenvOr("CHETANA_GIT_SHA", "unknown"),
	}
	if v := os.Getenv("SHUTDOWN_TIMEOUT"); v != "" {
		dur, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("config: parse SHUTDOWN_TIMEOUT %q: %w", v, err)
		}
		cfg.ShutdownTimeout = dur
	}
	if cfg.TenantID == "" {
		return Config{}, errors.New("config: CHETANA_TENANT_ID is required")
	}
	return cfg, nil
}

func getenvOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
