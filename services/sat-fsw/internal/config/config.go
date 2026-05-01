// Package config holds the sat-fsw service configuration.
package config

import (
	"errors"
	"fmt"

	pkgcfg "github.com/ppusapati/space/pkg/config"
)

// Config is the sat-fsw configuration.
type Config struct {
	pkgcfg.Common
	DSN          string
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
		return Config{}, fmt.Errorf("sat-fsw: %w", err)
	}
	secret := pkgcfg.String("CURSOR_SECRET", "")
	if len(secret) < 16 {
		return Config{}, errors.New("sat-fsw: CURSOR_SECRET must be at least 16 chars")
	}
	return Config{Common: c, DSN: dsn, CursorSecret: secret}, nil
}
