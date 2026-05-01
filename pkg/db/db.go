// Package db builds a connected pgxpool with sane production defaults
// and provides a small TxFunc helper for transactional repositories.
package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config configures Open. Only DSN is required; the other fields fall
// back to opinionated defaults documented inline.
type Config struct {
	// DSN is the PostgreSQL connection string (URL or keyword form).
	DSN string
	// MaxConns is the maximum size of the pool. Default 25.
	MaxConns int32
	// MinConns is the minimum number of idle connections held open.
	// Default 1.
	MinConns int32
	// MaxConnIdleTime is the maximum idle time before a connection is
	// closed. Default 30 m.
	MaxConnIdleTime time.Duration
	// MaxConnLifetime caps total per-connection lifetime. Default 1 h.
	MaxConnLifetime time.Duration
	// HealthCheckPeriod is the pgx ping interval. Default 1 m.
	HealthCheckPeriod time.Duration
	// ConnectTimeout caps the initial connect attempt. Default 5 s.
	ConnectTimeout time.Duration
}

// Open creates and pings a connection pool. It returns a *Pool so
// callers don't need to import pgxpool just to type the field.
func Open(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	if cfg.DSN == "" {
		return nil, errors.New("db: DSN is required")
	}
	pcfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("db: parse DSN: %w", err)
	}
	if cfg.MaxConns > 0 {
		pcfg.MaxConns = cfg.MaxConns
	} else {
		pcfg.MaxConns = 25
	}
	if cfg.MinConns > 0 {
		pcfg.MinConns = cfg.MinConns
	} else {
		pcfg.MinConns = 1
	}
	if cfg.MaxConnIdleTime > 0 {
		pcfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	} else {
		pcfg.MaxConnIdleTime = 30 * time.Minute
	}
	if cfg.MaxConnLifetime > 0 {
		pcfg.MaxConnLifetime = cfg.MaxConnLifetime
	} else {
		pcfg.MaxConnLifetime = time.Hour
	}
	if cfg.HealthCheckPeriod > 0 {
		pcfg.HealthCheckPeriod = cfg.HealthCheckPeriod
	} else {
		pcfg.HealthCheckPeriod = time.Minute
	}
	connectTimeout := cfg.ConnectTimeout
	if connectTimeout == 0 {
		connectTimeout = 5 * time.Second
	}
	pool, err := pgxpool.NewWithConfig(ctx, pcfg)
	if err != nil {
		return nil, fmt.Errorf("db: connect: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db: ping: %w", err)
	}
	return pool, nil
}

// TxFunc is the body of a transaction. The pgx.Tx is automatically
// committed when fn returns nil and rolled back otherwise.
type TxFunc func(ctx context.Context, tx pgx.Tx) error

// InTx executes fn inside a SERIALIZABLE-by-default transaction with
// `iso` isolation level. Pass [pgx.ReadCommitted] when serializable is
// too strict.
func InTx(ctx context.Context, pool *pgxpool.Pool, iso pgx.TxIsoLevel, fn TxFunc) error {
	if pool == nil {
		return errors.New("db: nil pool")
	}
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: iso})
	if err != nil {
		return fmt.Errorf("db: begin: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()
	if err = fn(ctx, tx); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("db: commit: %w", err)
	}
	return nil
}
