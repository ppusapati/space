// Package config holds the eo-catalog service-specific configuration on
// top of the shared `pkg/config.Common` primitives.
package config

import (
	"errors"
	"fmt"

	pkgcfg "github.com/ppusapati/space/pkg/config"
)

// Config is the full eo-catalog configuration loaded from the environment.
type Config struct {
	pkgcfg.Common
	// DSN is the PostgreSQL DSN.
	DSN string
	// CursorSecret is the HMAC key used to sign pagination cursors.
	CursorSecret string
}

// Load reads the environment.
func Load() (Config, error) {
	c, err := pkgcfg.LoadCommon()
	if err != nil {
		return Config{}, err
	}
	dsn, err := pkgcfg.MustString("DATABASE_URL")
	if err != nil {
		return Config{}, fmt.Errorf("eo-catalog: %w", err)
	}
	secret := pkgcfg.String("CURSOR_SECRET", "")
	if len(secret) < 16 {
		return Config{}, errors.New("eo-catalog: CURSOR_SECRET must be at least 16 chars")
	}
	return Config{Common: c, DSN: dsn, CursorSecret: secret}, nil
}
