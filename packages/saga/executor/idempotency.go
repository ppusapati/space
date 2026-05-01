// Package executor provides idempotency checking for saga steps
package executor

import (
	"fmt"
	"sync"
	"time"

	"p9e.in/samavaya/packages/saga/models"
)

// IdempotencyImpl provides idempotency checking and result caching
type IdempotencyImpl struct {
	mu               sync.RWMutex
	cache            map[string]*cachedStepResult // key -> cached result
	ttl              time.Duration                 // Time to live for cached results
	maxCacheSize     int                           // Maximum cache size
	cacheSizeCounter int                           // Current cache size
}

// cachedStepResult holds a cached step result with metadata
type cachedStepResult struct {
	result    *models.StepResult
	cachedAt  time.Time
	expiresAt time.Time
}

// NewIdempotencyImpl creates a new idempotency checker instance
func NewIdempotencyImpl(ttl time.Duration, maxCacheSize int) *IdempotencyImpl {
	return &IdempotencyImpl{
		cache:        make(map[string]*cachedStepResult),
		ttl:          ttl,
		maxCacheSize: maxCacheSize,
	}
}

// GetCachedResult retrieves a cached result if it exists and hasn't expired
func (i *IdempotencyImpl) GetCachedResult(sagaID string, stepNum int) *models.StepResult {
	i.mu.RLock()
	defer i.mu.RUnlock()

	// 1. Generate idempotency key
	key := generateIdempotencyKey(sagaID, stepNum)

	// 2. Look up in cache
	cached, exists := i.cache[key]
	if !exists {
		return nil // Not in cache
	}

	// 3. Check if expired
	if time.Now().After(cached.expiresAt) {
		// Expired, but we'll let cleanup handle it
		return nil
	}

	// 4. Return cached result
	return cached.result
}

// CacheResult stores a step result with TTL
func (i *IdempotencyImpl) CacheResult(sagaID string, stepNum int, result *models.StepResult) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	// 1. Generate idempotency key
	key := generateIdempotencyKey(sagaID, stepNum)

	// 2. Check cache size and evict if necessary
	if i.cacheSizeCounter >= i.maxCacheSize {
		i.evictOldestExpired()
	}

	// 3. Add to cache with expiration
	now := time.Now()
	cachedResult := &cachedStepResult{
		result:    result,
		cachedAt:  now,
		expiresAt: now.Add(i.ttl),
	}

	i.cache[key] = cachedResult
	i.cacheSizeCounter++

	return nil
}

// IsDuplicate checks if a step has already been executed (duplicate detection)
func (i *IdempotencyImpl) IsDuplicate(sagaID string, stepNum int) bool {
	i.mu.RLock()
	defer i.mu.RUnlock()

	key := generateIdempotencyKey(sagaID, stepNum)
	cached, exists := i.cache[key]

	if !exists {
		return false // Not a duplicate
	}

	// Check if expired
	if time.Now().After(cached.expiresAt) {
		return false // Expired, not considered duplicate
	}

	return true // Found non-expired result, so it's a duplicate
}

// ClearCache removes all cached results (for testing)
func (i *IdempotencyImpl) ClearCache() {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.cache = make(map[string]*cachedStepResult)
	i.cacheSizeCounter = 0
}

// GetCacheStats returns cache statistics
func (i *IdempotencyImpl) GetCacheStats() map[string]interface{} {
	i.mu.RLock()
	defer i.mu.RUnlock()

	expiredCount := 0
	for _, cached := range i.cache {
		if time.Now().After(cached.expiresAt) {
			expiredCount++
		}
	}

	return map[string]interface{}{
		"totalCached":   i.cacheSizeCounter,
		"expiredCount":  expiredCount,
		"maxCacheSize":  i.maxCacheSize,
		"ttlSeconds":    i.ttl.Seconds(),
	}
}

// evictOldestExpired removes the oldest expired entry from cache
func (i *IdempotencyImpl) evictOldestExpired() {
	now := time.Now()

	// First pass: remove any expired entries
	for key, cached := range i.cache {
		if now.After(cached.expiresAt) {
			delete(i.cache, key)
			i.cacheSizeCounter--
			return // Just remove one
		}
	}

	// If no expired entries, remove the oldest one
	var oldestKey string
	var oldestTime time.Time = time.Now().Add(i.ttl) // Initialize to future

	for key, cached := range i.cache {
		if cached.cachedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = cached.cachedAt
		}
	}

	if oldestKey != "" {
		delete(i.cache, oldestKey)
		i.cacheSizeCounter--
	}
}

// generateIdempotencyKey generates a unique key for idempotency caching
func generateIdempotencyKey(sagaID string, stepNum int) string {
	return fmt.Sprintf("saga:%s:step:%d", sagaID, stepNum)
}
