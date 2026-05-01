// Package config loads sat-mission configuration via packages/config.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	pkgconfig "p9e.in/samavaya/packages/config"
	"p9e.in/samavaya/packages/config/file"

	_ "p9e.in/samavaya/packages/encoding/yaml"
)

type Config struct {
	Service  ServiceConfig  `yaml:"service" json:"service"`
	HTTP     HTTPConfig     `yaml:"http" json:"http"`
	Metrics  MetricsConfig  `yaml:"metrics" json:"metrics"`
	Shutdown ShutdownConfig `yaml:"shutdown" json:"shutdown"`
	Database DatabaseConfig `yaml:"database" json:"database"`
	CORS     CORSConfig     `yaml:"cors" json:"cors"`
}

type ServiceConfig struct {
	Name        string `yaml:"name" json:"name"`
	Environment string `yaml:"environment" json:"environment"`
	LogLevel    string `yaml:"log_level" json:"log_level"`
}
type HTTPConfig struct{ Addr string `yaml:"addr" json:"addr"` }
type MetricsConfig struct{ Addr string `yaml:"addr" json:"addr"` }
type ShutdownConfig struct{ Timeout time.Duration `yaml:"timeout" json:"timeout"` }
type DatabaseConfig struct{ DSN string `yaml:"dsn" json:"dsn"` }
type CORSConfig struct{ AllowedOrigins []string `yaml:"allowed_origins" json:"allowed_origins"` }

func Load() (Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "config/config.yaml"
	}
	c := pkgconfig.New(pkgconfig.WithSource(file.NewSource(path)))
	if err := c.Load(); err != nil {
		return Config{}, fmt.Errorf("sat-mission: load %s: %w", path, err)
	}
	defer c.Close()
	var cfg Config
	if err := c.Scan(&cfg); err != nil {
		return Config{}, fmt.Errorf("sat-mission: scan: %w", err)
	}
	if cfg.Service.Name == "" {
		cfg.Service.Name = "sat-mission"
	}
	if cfg.Service.Environment == "" {
		cfg.Service.Environment = "development"
	}
	if cfg.Service.LogLevel == "" {
		cfg.Service.LogLevel = "info"
	}
	if cfg.HTTP.Addr == "" {
		cfg.HTTP.Addr = ":8080"
	}
	if cfg.Metrics.Addr == "" {
		cfg.Metrics.Addr = ":9090"
	}
	if cfg.Shutdown.Timeout == 0 {
		cfg.Shutdown.Timeout = 30 * time.Second
	}
	if len(cfg.CORS.AllowedOrigins) == 0 {
		cfg.CORS.AllowedOrigins = []string{"*"}
	}
	if v := os.Getenv("DATABASE_URL"); v != "" {
		cfg.Database.DSN = v
	}
	if v := os.Getenv("HTTP_ADDR"); v != "" {
		cfg.HTTP.Addr = v
	}
	if v := os.Getenv("METRICS_ADDR"); v != "" {
		cfg.Metrics.Addr = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Service.LogLevel = v
	}
	if v := os.Getenv("SHUTDOWN_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Shutdown.Timeout = d
		}
	}
	if v := os.Getenv("ALLOWED_ORIGINS"); v != "" {
		out := []string{}
		for _, p := range strings.Split(v, ",") {
			if s := strings.TrimSpace(p); s != "" {
				out = append(out, s)
			}
		}
		if len(out) > 0 {
			cfg.CORS.AllowedOrigins = out
		}
	}
	if cfg.Database.DSN == "" {
		return Config{}, errors.New("sat-mission: database.dsn (or DATABASE_URL) is required")
	}
	return cfg, nil
}

func (c Config) HTTPPort() string {
	if len(c.HTTP.Addr) > 0 && c.HTTP.Addr[0] == ':' {
		return c.HTTP.Addr[1:]
	}
	return c.HTTP.Addr
}
