package algorithms

import (
	"context"
	"sync"
	"time"

	"p9e.in/samavaya/packages/ratelimit"
)

// TokenBucketLimiter implements classic token bucket rate limiting
type TokenBucketLimiter struct {
	capacity   float64 // Maximum tokens in bucket
	refillRate float64 // Tokens per second
	buckets    map[string]*TokenBucket
	mu         sync.RWMutex
}

// TokenBucket represents a single bucket for a key
type TokenBucket struct {
	tokens      float64
	capacity    float64
	refillRate  float64
	lastRefill  time.Time
	mu          sync.Mutex
}

// NewTokenBucketLimiter creates a new token bucket rate limiter
func NewTokenBucketLimiter(capacity float64, refillRate float64) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		capacity:   capacity,
		refillRate: refillRate,
		buckets:    make(map[string]*TokenBucket),
	}
}

// Allow checks if a request is allowed
func (tbl *TokenBucketLimiter) Allow(ctx context.Context, key string) (bool, error) {
	return tbl.AllowN(ctx, key, 1)
}

// AllowN allows N tokens
func (tbl *TokenBucketLimiter) AllowN(ctx context.Context, key string, n int) (bool, error) {
	bucket := tbl.getOrCreateBucket(key)

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	bucket.tokens = min(bucket.capacity, bucket.tokens+elapsed*bucket.refillRate)
	bucket.lastRefill = now

	// Check if we have enough tokens
	needed := float64(n)
	if bucket.tokens >= needed {
		bucket.tokens -= needed
		return true, nil
	}

	return false, nil
}

// Reserve reserves tokens for a future request
func (tbl *TokenBucketLimiter) Reserve(ctx context.Context, key string) (*ratelimit.Reservation, error) {
	bucket := tbl.getOrCreateBucket(key)

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	bucket.tokens = min(bucket.capacity, bucket.tokens+elapsed*bucket.refillRate)
	bucket.lastRefill = now

	// If we have tokens, allow immediately
	if bucket.tokens >= 1.0 {
		bucket.tokens -= 1.0
		return &ratelimit.Reservation{
			ReadyAt: time.Now(),
			Delay:   0,
			OK:       true,
		}, nil
	}

	// Calculate delay until next token is available
	tokensNeeded := 1.0 - bucket.tokens
	delaySeconds := tokensNeeded / bucket.refillRate
	delay := time.Duration(delaySeconds*1000) * time.Millisecond

	return &ratelimit.Reservation{
		ReadyAt: time.Now().Add(delay),
		Delay:   delay,
		OK:       false,
	}, nil
}

// GetStats returns current rate limit stats
func (tbl *TokenBucketLimiter) GetStats(ctx context.Context, key string) (*ratelimit.Stats, error) {
	bucket := tbl.getOrCreateBucket(key)

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Refill to get accurate current state
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	currentTokens := min(bucket.capacity, bucket.tokens+elapsed*bucket.refillRate)

	return &ratelimit.Stats{
		Key:           key,
		AllowedCount:  int64(currentTokens),
		RejectedCount: 0, // Token bucket doesn't track rejections
		CurrentLimit:  int64(bucket.refillRate),
		WindowStart:   bucket.lastRefill,
		WindowEnd:     time.Now().Add(time.Duration(bucket.refillRate) * time.Second),
		Metrics: map[string]interface{}{
			"tokens":        currentTokens,
			"capacity":      bucket.capacity,
			"refill_rate":   bucket.refillRate,
		},
	}, nil
}

// Reset resets the bucket for a key
func (tbl *TokenBucketLimiter) Reset(ctx context.Context, key string) error {
	tbl.mu.Lock()
	defer tbl.mu.Unlock()

	delete(tbl.buckets, key)
	return nil
}

// SetCapacity updates the capacity for a key
func (tbl *TokenBucketLimiter) SetCapacity(key string, capacity float64) {
	bucket := tbl.getOrCreateBucket(key)
	bucket.mu.Lock()
	defer bucket.mu.Unlock()
	bucket.capacity = capacity
}

// SetRefillRate updates the refill rate for a key
func (tbl *TokenBucketLimiter) SetRefillRate(key string, refillRate float64) {
	bucket := tbl.getOrCreateBucket(key)
	bucket.mu.Lock()
	defer bucket.mu.Unlock()
	bucket.refillRate = refillRate
}

// Helper methods

func (tbl *TokenBucketLimiter) getOrCreateBucket(key string) *TokenBucket {
	tbl.mu.Lock()
	defer tbl.mu.Unlock()

	if bucket, ok := tbl.buckets[key]; ok {
		return bucket
	}

	bucket := &TokenBucket{
		tokens:     tbl.capacity,
		capacity:   tbl.capacity,
		refillRate: tbl.refillRate,
		lastRefill: time.Now(),
	}
	tbl.buckets[key] = bucket
	return bucket
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
