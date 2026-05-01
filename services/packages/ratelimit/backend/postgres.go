package backend

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"p9e.in/samavaya/packages/p9log"
	"p9e.in/samavaya/packages/ratelimit"
)

// PostgresRateLimiter implements distributed rate limiting with PostgreSQL.
// `logger` stores *p9log.Helper (B.1 sweep).
type PostgresRateLimiter struct {
	pool          *pgxpool.Pool
	logger        *p9log.Helper
	localCache    map[string]*CachedLimit
	cacheMu       sync.RWMutex
	cacheTTL      time.Duration
	syncInterval  time.Duration
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// CachedLimit represents a cached rate limit entry
type CachedLimit struct {
	AllowedCount    int64
	RejectedCount   int64
	CurrentLimit    int64
	WindowStart     time.Time
	LastSync        time.Time
	CachedUntil     time.Time
}

// New creates a new PostgreSQL-backed rate limiter
func New(pool *pgxpool.Pool, logger p9log.Logger) *PostgresRateLimiter {
	pr := &PostgresRateLimiter{
		pool:         pool,
		logger:       p9log.NewHelper(logger),
		localCache:   make(map[string]*CachedLimit),
		cacheTTL:     10 * time.Second,
		syncInterval: 10 * time.Second,
		stopChan:     make(chan struct{}),
	}

	// Start background sync worker
	pr.wg.Add(1)
	go pr.syncWorker()

	return pr
}

// Allow checks if a request is allowed
func (pr *PostgresRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	return pr.AllowN(ctx, key, 1)
}

// AllowN allows N requests
func (pr *PostgresRateLimiter) AllowN(ctx context.Context, key string, n int) (bool, error) {
	// Check cache first
	cached, ok := pr.getCachedLimit(key)
	if ok && time.Now().Before(cached.CachedUntil) {
		// Cache hit and still valid
		if cached.AllowedCount < cached.CurrentLimit {
			// Have capacity in local cache
			cached.AllowedCount -= int64(n)
			pr.setCachedLimit(key, cached)
			return true, nil
		}
		// Cache says we're over limit, but sync with database to be sure
	}

	// Database check (uses advisory locks for coordination)
	const query = `
		UPDATE rate_limits
		SET allowed_requests = allowed_requests + $1,
		    updated_at = NOW()
		WHERE key = $2
		  AND allowed_requests + $1 <= limit_per_second
		RETURNING allowed_requests, rejected_requests, limit_per_second, window_start;
	`

	var allowed, rejected, limit int64
	var windowStart time.Time

	err := pr.pool.QueryRow(ctx, query, n, key).Scan(&allowed, &rejected, &limit, &windowStart)
	if err != nil {
		// Key doesn't exist or update failed - record rejection
		rejected := int64(1)
		if err.Error() == "no rows in result set" {
			rejected = int64(n)
		}

		// Upsert to record rejection
		insertQuery := `
			INSERT INTO rate_limits (key, service_name, allowed_requests, rejected_requests, limit_per_second, window_start)
			VALUES ($1, $2, 0, $3, 100, NOW())
			ON CONFLICT (key) DO UPDATE SET
				rejected_requests = rate_limits.rejected_requests + $3,
				updated_at = NOW()
		`

		service := extractServiceName(key)
		pr.pool.Exec(ctx, insertQuery, key, service, rejected)

		pr.logger.Debug("request rejected",
			"key", key,
			"reason", "rate limit exceeded",
		)

		return false, nil
	}

	// Update local cache
	cached = &CachedLimit{
		AllowedCount:  allowed,
		RejectedCount: rejected,
		CurrentLimit:  limit,
		WindowStart:   windowStart,
		LastSync:      time.Now(),
		CachedUntil:   time.Now().Add(pr.cacheTTL),
	}
	pr.setCachedLimit(key, cached)

	return true, nil
}

// Reserve reserves capacity for a future request
func (pr *PostgresRateLimiter) Reserve(ctx context.Context, key string) (*ratelimit.Reservation, error) {
	if ok, _ := pr.Allow(ctx, key); ok {
		return &ratelimit.Reservation{
			ReadyAt: time.Now(),
			Delay:   0,
			OK:       true,
		}, nil
	}

	// Calculate delay based on request rate
	cached, _ := pr.getCachedLimit(key)
	if cached == nil {
		cached = &CachedLimit{CurrentLimit: 100}
	}

	delay := time.Duration(float64(time.Second) / float64(cached.CurrentLimit))
	return &ratelimit.Reservation{
		ReadyAt: time.Now().Add(delay),
		Delay:   delay,
		OK:       false,
	}, nil
}

// GetStats returns current rate limit stats
func (pr *PostgresRateLimiter) GetStats(ctx context.Context, key string) (*ratelimit.Stats, error) {
	const query = `
		SELECT allowed_requests, rejected_requests, limit_per_second, window_start
		FROM rate_limits
		WHERE key = $1
	`

	var allowed, rejected, limit int64
	var windowStart time.Time

	err := pr.pool.QueryRow(ctx, query, key).Scan(&allowed, &rejected, &limit, &windowStart)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate limit stats: %w", err)
	}

	return &ratelimit.Stats{
		Key:            key,
		AllowedCount:   allowed,
		RejectedCount:  rejected,
		CurrentLimit:   limit,
		WindowStart:    windowStart,
		WindowEnd:      windowStart.Add(time.Second),
	}, nil
}

// Reset resets the rate limit for a key
func (pr *PostgresRateLimiter) Reset(ctx context.Context, key string) error {
	const query = `
		UPDATE rate_limits
		SET allowed_requests = 0, rejected_requests = 0, window_start = NOW()
		WHERE key = $1
	`

	_, err := pr.pool.Exec(ctx, query, key)
	if err != nil {
		return fmt.Errorf("failed to reset rate limit: %w", err)
	}

	// Clear cache
	pr.cacheMu.Lock()
	delete(pr.localCache, key)
	pr.cacheMu.Unlock()

	return nil
}

// SetLimit updates the rate limit for a key
func (pr *PostgresRateLimiter) SetLimit(ctx context.Context, key string, limitPerSecond int64) error {
	const query = `
		UPDATE rate_limits
		SET limit_per_second = $1, updated_at = NOW()
		WHERE key = $2
	`

	_, err := pr.pool.Exec(ctx, query, limitPerSecond, key)
	if err != nil {
		return fmt.Errorf("failed to set rate limit: %w", err)
	}

	// Clear cache to force refresh
	pr.cacheMu.Lock()
	delete(pr.localCache, key)
	pr.cacheMu.Unlock()

	return nil
}

// Close closes the rate limiter
func (pr *PostgresRateLimiter) Close() error {
	close(pr.stopChan)
	pr.wg.Wait()
	return nil
}

// Helper methods

func (pr *PostgresRateLimiter) getCachedLimit(key string) (*CachedLimit, bool) {
	pr.cacheMu.RLock()
	defer pr.cacheMu.RUnlock()

	limit, ok := pr.localCache[key]
	return limit, ok
}

func (pr *PostgresRateLimiter) setCachedLimit(key string, limit *CachedLimit) {
	pr.cacheMu.Lock()
	defer pr.cacheMu.Unlock()

	pr.localCache[key] = limit
}

// syncWorker periodically syncs cache with database
func (pr *PostgresRateLimiter) syncWorker() {
	defer pr.wg.Done()

	ticker := time.NewTicker(pr.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pr.stopChan:
			return
		case <-ticker.C:
			pr.syncCache()
		}
	}
}

// syncCache syncs all cached entries back to database
func (pr *PostgresRateLimiter) syncCache() {
	pr.cacheMu.RLock()
	keys := make([]string, 0, len(pr.localCache))
	for key := range pr.localCache {
		keys = append(keys, key)
	}
	pr.cacheMu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, key := range keys {
		pr.cacheMu.RLock()
		cached := pr.localCache[key]
		pr.cacheMu.RUnlock()

		if cached == nil {
			continue
		}

		// Update window if it's expired
		if time.Now().Sub(cached.WindowStart) > time.Second {
			const resetQuery = `
				UPDATE rate_limits
				SET allowed_requests = 0, rejected_requests = 0, window_start = NOW()
				WHERE key = $1
			`
			pr.pool.Exec(ctx, resetQuery, key)

			pr.cacheMu.Lock()
			delete(pr.localCache, key)
			pr.cacheMu.Unlock()
		}
	}
}

// extractServiceName extracts service name from key
// Key format: "service-name:client-id"
func extractServiceName(key string) string {
	for i, ch := range key {
		if ch == ':' {
			return key[:i]
		}
	}
	return key
}
