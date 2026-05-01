package checks

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/samavaya/packages/healthcheck"
)

// DatabaseCheck implements database health checking
type DatabaseCheck struct {
	DSN     string
	Query   string
	Timeout time.Duration
	pool    *pgxpool.Pool
	name    string
	ownsPool bool
}

// NewDatabaseCheck creates a new database health check
func NewDatabaseCheck(cfg healthcheck.DatabaseCheckConfig) *DatabaseCheck {
	check := &DatabaseCheck{
		DSN:     cfg.DSN,
		Query:   cfg.Query,
		Timeout: cfg.Timeout,
		name:    fmt.Sprintf("database:%s", cfg.DSN),
		ownsPool: true,
	}

	if check.Query == "" {
		check.Query = "SELECT 1"
	}

	if check.Timeout == 0 {
		check.Timeout = 10 * time.Second
	}

	return check
}

// NewDatabaseCheckWithPool creates a database health check with existing pool
func NewDatabaseCheckWithPool(pool *pgxpool.Pool, query string, timeout time.Duration) *DatabaseCheck {
	check := &DatabaseCheck{
		pool:     pool,
		Query:    query,
		Timeout:  timeout,
		name:     "database:pooled",
		ownsPool: false,
	}

	if check.Query == "" {
		check.Query = "SELECT 1"
	}

	if check.Timeout == 0 {
		check.Timeout = 10 * time.Second
	}

	return check
}

// Check performs the database health check
func (dc *DatabaseCheck) Check(ctx context.Context) (*healthcheck.CheckResult, error) {
	start := time.Now()
	result := &healthcheck.CheckResult{
		CheckedAt: start,
		Details:   make(map[string]interface{}),
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, dc.Timeout)
	defer cancel()

	// Get or create pool
	pool := dc.pool
	if pool == nil {
		// Create temporary pool for this check
		pgCfg, err := pgxpool.ParseConfig(dc.DSN)
		if err != nil {
			result.Status = healthcheck.StatusUnhealthy
			result.Error = err.Error()
			result.Message = fmt.Sprintf("Failed to parse DSN: %v", err)
			result.Duration = time.Since(start)
			result.Details["duration_ms"] = result.Duration.Milliseconds()
			return result, err
		}

		pool, err = pgxpool.NewWithConfig(ctx, pgCfg)
		if err != nil {
			result.Status = healthcheck.StatusUnhealthy
			result.Error = err.Error()
			result.Message = fmt.Sprintf("Failed to create pool: %v", err)
			result.Duration = time.Since(start)
			result.Details["duration_ms"] = result.Duration.Milliseconds()
			return result, err
		}
		defer pool.Close()

		// Perform ping
		if err := pool.Ping(ctx); err != nil {
			result.Status = healthcheck.StatusUnhealthy
			result.Error = err.Error()
			result.Message = fmt.Sprintf("Database ping failed: %v", err)
			result.Duration = time.Since(start)
			result.Details["duration_ms"] = result.Duration.Milliseconds()
			return result, err
		}
	} else {
		// Use existing pool
		if err := pool.Ping(ctx); err != nil {
			result.Status = healthcheck.StatusUnhealthy
			result.Error = err.Error()
			result.Message = fmt.Sprintf("Database ping failed: %v", err)
			result.Duration = time.Since(start)
			result.Details["duration_ms"] = result.Duration.Milliseconds()
			return result, err
		}
	}

	// Execute test query
	var queryResult interface{}
	err := pool.QueryRow(ctx, dc.Query).Scan(&queryResult)
	result.Duration = time.Since(start)
	result.Details["duration_ms"] = result.Duration.Milliseconds()
	result.Details["query"] = dc.Query
	result.Details["pool_owns"] = !dc.ownsPool

	if err != nil {
		result.Status = healthcheck.StatusUnhealthy
		result.Error = err.Error()
		result.Message = fmt.Sprintf("Query failed: %v", err)
		return result, err
	}

	// Get pool stats
	stats := pool.Stat()
	result.Details["connections"] = stats.TotalConns()
	result.Details["idle_conns"] = stats.IdleConns()
	result.Details["acquired_conns"] = stats.AcquiredConns()

	result.Status = healthcheck.StatusHealthy
	result.Message = "Database is healthy"

	return result, nil
}

// Type returns the check type
func (dc *DatabaseCheck) Type() healthcheck.CheckType {
	return healthcheck.CheckTypeDatabase
}

// Name returns the check name
func (dc *DatabaseCheck) Name() string {
	return dc.name
}

// SetName sets the check name
func (dc *DatabaseCheck) SetName(name string) {
	dc.name = name
}
