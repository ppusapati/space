// Package config loads gs-rf configuration via packages/config.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	pkgconfig "p9e.in/chetana/packages/config"
	"p9e.in/chetana/packages/config/file"

	_ "p9e.in/chetana/packages/encoding/yaml"
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

type HTTPConfig struct {
	Addr string `yaml:"addr" json:"addr"`
}

type MetricsConfig struct {
	Addr string `yaml:"addr" json:"addr"`
}

type ShutdownConfig struct {
	Timeout time.Duration `yaml:"timeout" json:"timeout"`
}

type DatabaseConfig struct {
	DSN string `yaml:"dsn" json:"dsn"`
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins" json:"allowed_origins"`
}

func Load() (Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "config/config.yaml"
	}
	src := file.NewSource(path)
	c := pkgconfig.New(pkgconfig.WithSource(src))
	if err := c.Load(); err != nil {
		return Config{}, fmt.Errorf("gs-rf: load %s: %w", path, err)
	}
	defer c.Close()

	var cfg Config
	if err := c.Scan(&cfg); err != nil {
		return Config{}, fmt.Errorf("gs-rf: scan config: %w", err)
	}
	applyDefaults(&cfg)
	overlayEnv(&cfg)
	if err := validate(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func applyDefaults(c *Config) {
	if c.Service.Name == "" {
		c.Service.Name = "gs-rf"
	}
	if c.Service.Environment == "" {
		c.Service.Environment = "development"
	}
	if c.Service.LogLevel == "" {
		c.Service.LogLevel = "info"
	}
	if c.HTTP.Addr == "" {
		c.HTTP.Addr = ":8080"
	}
	if c.Metrics.Addr == "" {
		c.Metrics.Addr = ":9090"
	}
	if c.Shutdown.Timeout == 0 {
		c.Shutdown.Timeout = 30 * time.Second
	}
	if len(c.CORS.AllowedOrigins) == 0 {
		c.CORS.AllowedOrigins = []string{"*"}
	}
}

func overlayEnv(c *Config) {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		c.Database.DSN = v
	}
	if v := os.Getenv("SERVICE_NAME"); v != "" {
		c.Service.Name = v
	}
	if v := os.Getenv("ENVIRONMENT"); v != "" {
		c.Service.Environment = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		c.Service.LogLevel = v
	}
	if v := os.Getenv("HTTP_ADDR"); v != "" {
		c.HTTP.Addr = v
	}
	if v := os.Getenv("METRICS_ADDR"); v != "" {
		c.Metrics.Addr = v
	}
	if v := os.Getenv("SHUTDOWN_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			c.Shutdown.Timeout = d
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
			c.CORS.AllowedOrigins = out
		}
	}
}

func validate(c *Config) error {
	if c.Database.DSN == "" {
		return errors.New("gs-rf: database.dsn (or DATABASE_URL) is required")
	}
	return nil
}

func (c Config) HTTPPort() string {
	if len(c.HTTP.Addr) > 0 && c.HTTP.Addr[0] == ':' {
		return c.HTTP.Addr[1:]
	}
	return c.HTTP.Addr
}
