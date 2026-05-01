// Package config holds the sat-telemetry service configuration.
package config

import (
	"errors"
	"os"
	"strings"
	"time"
)

// Config is the sat-telemetry configuration loaded from environment.
type Config struct {
	ServiceName     string
	Environment     string
	LogLevel        string
	HTTPAddr        string
	MetricsAddr     string
	ShutdownTimeout time.Duration
	DSN             string
	AllowedOrigins  []string
}

// Load reads the environment.
func Load() (Config, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return Config{}, errors.New("sat-telemetry: DATABASE_URL required")
	}
	return Config{
		ServiceName:     getenv("SERVICE_NAME", "sat-telemetry"),
		Environment:     getenv("ENVIRONMENT", "development"),
		LogLevel:        getenv("LOG_LEVEL", "info"),
		HTTPAddr:        getenv("HTTP_ADDR", ":8080"),
		MetricsAddr:     getenv("METRICS_ADDR", ":9090"),
		ShutdownTimeout: getenvDuration("SHUTDOWN_TIMEOUT", 30*time.Second),
		DSN:             dsn,
		AllowedOrigins:  getenvList("ALLOWED_ORIGINS", []string{"*"}),
	}, nil
}

// HTTPPort returns the port suffix from HTTPAddr (e.g. ":8080" -> "8080").
func (c Config) HTTPPort() string {
	if len(c.HTTPAddr) > 0 && c.HTTPAddr[0] == ':' {
		return c.HTTPAddr[1:]
	}
	return c.HTTPAddr
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func getenvList(key string, def []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	out := make([]string, 0)
	for _, p := range strings.Split(v, ",") {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}
