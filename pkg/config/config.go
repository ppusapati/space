// Package config loads service configuration from environment variables
// with strongly-typed accessors and exhaustive validation. The Load
// function is the only entry point; every service constructs its own
// config struct that embeds the relevant primitives below.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

// ErrMissing is returned when a required environment variable is unset
// or empty.
var ErrMissing = errors.New("required environment variable missing")

// String returns the value of envvar `key`, or `fallback` if unset.
func String(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// MustString returns the value of envvar `key` or returns ErrMissing.
func MustString(key string) (string, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return "", fmt.Errorf("%w: %s", ErrMissing, key)
	}
	return v, nil
}

// Int returns the integer value of envvar `key`, or `fallback` if unset
// or unparseable.
func Int(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

// Bool returns true when the envvar is set to one of {"1", "true",
// "yes", "on"} (case-insensitive). Otherwise returns `fallback`.
func Bool(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	switch v {
	case "1", "t", "T", "true", "TRUE", "True", "yes", "YES", "on", "ON":
		return true
	case "0", "f", "F", "false", "FALSE", "False", "no", "NO", "off", "OFF":
		return false
	}
	return fallback
}

// Duration parses a Go-style duration string (e.g., "30s", "5m"). On
// any error the fallback is returned.
func Duration(key string, fallback time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

// Common holds the configuration knobs every service needs.
type Common struct {
	// ServiceName is used as the OTEL service.name resource attribute,
	// the Prometheus subsystem, and the gRPC reflection name.
	ServiceName string
	// Environment is one of "dev", "staging", "prod".
	Environment string
	// LogLevel is one of "debug", "info", "warn", "error".
	LogLevel string
	// HTTPAddr is the listen address for the public HTTP/2 ConnectRPC
	// endpoint, e.g. ":8080".
	HTTPAddr string
	// MetricsAddr is the listen address for the Prometheus scrape
	// endpoint, e.g. ":9090".
	MetricsAddr string
	// ShutdownTimeout caps the time given to in-flight requests on
	// SIGTERM.
	ShutdownTimeout time.Duration
}

// LoadCommon populates a Common from the environment.
//
// Required:   SERVICE_NAME
// Optional:   ENVIRONMENT (default "dev"), LOG_LEVEL (default "info"),
//             HTTP_ADDR (default ":8080"), METRICS_ADDR (default ":9090"),
//             SHUTDOWN_TIMEOUT (default "15s").
func LoadCommon() (Common, error) {
	name, err := MustString("SERVICE_NAME")
	if err != nil {
		return Common{}, err
	}
	return Common{
		ServiceName:     name,
		Environment:     String("ENVIRONMENT", "dev"),
		LogLevel:        String("LOG_LEVEL", "info"),
		HTTPAddr:        String("HTTP_ADDR", ":8080"),
		MetricsAddr:     String("METRICS_ADDR", ":9090"),
		ShutdownTimeout: Duration("SHUTDOWN_TIMEOUT", 15*time.Second),
	}, nil
}
