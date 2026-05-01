// Package config holds the eo-catalog service configuration.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config is the eo-catalog configuration loaded from environment.
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
		return Config{}, errors.New("eo-catalog: DATABASE_URL required")
	}
	c := Config{
		ServiceName:     getenv("SERVICE_NAME", "eo-catalog"),
		Environment:     getenv("ENVIRONMENT", "development"),
		LogLevel:        getenv("LOG_LEVEL", "info"),
		HTTPAddr:        getenv("HTTP_ADDR", ":8080"),
		MetricsAddr:     getenv("METRICS_ADDR", ":9090"),
		ShutdownTimeout: getenvDuration("SHUTDOWN_TIMEOUT", 30*time.Second),
		DSN:             dsn,
		AllowedOrigins:  getenvList("ALLOWED_ORIGINS", []string{"*"}),
	}
	return c, nil
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
	cur := ""
	for _, c := range v {
		if c == ',' {
			if s := trim(cur); s != "" {
				out = append(out, s)
			}
			cur = ""
			continue
		}
		cur += string(c)
	}
	if s := trim(cur); s != "" {
		out = append(out, s)
	}
	return out
}

func trim(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

// httpPort returns the port suffix from HTTPAddr (e.g. ":8080" → "8080").
func (c Config) HTTPPort() string {
	if len(c.HTTPAddr) > 0 && c.HTTPAddr[0] == ':' {
		return c.HTTPAddr[1:]
	}
	return c.HTTPAddr
}

// String returns a redacted summary safe for logs.
func (c Config) String() string {
	return fmt.Sprintf("eo-catalog{env=%s log=%s http=%s metrics=%s timeout=%s dsn=set:%s}",
		c.Environment, c.LogLevel, c.HTTPAddr, c.MetricsAddr, c.ShutdownTimeout,
		strconv.FormatBool(c.DSN != ""))
}
