package cache

import (
	"context"
	"sync"
	"time"

	"p9e.in/chetana/packages/p9log"
	"p9e.in/chetana/packages/registry"
)

// CacheEntry holds cached data with expiration
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// IsExpired checks if the cache entry is expired
func (ce *CacheEntry) IsExpired() bool {
	return time.Now().After(ce.ExpiresAt)
}

// CachedRegistry wraps a RegistryBackend with an in-memory cache layer.
// `logger` stores *p9log.Helper (B.1 sweep).
type CachedRegistry struct {
	backend    registry.RegistryBackend
	logger     *p9log.Helper
	ttl        time.Duration
	instancesCache map[string]*CacheEntry // service_name -> instances
	instanceCache  map[string]*CacheEntry // instance_id -> instance
	cacheMu        sync.RWMutex
}

// New creates a new cached registry
func New(backend registry.RegistryBackend, logger p9log.Logger, ttl time.Duration) *CachedRegistry {
	return &CachedRegistry{
		backend:         backend,
		logger:          p9log.NewHelper(logger),
		ttl:             ttl,
		instancesCache:  make(map[string]*CacheEntry),
		instanceCache:   make(map[string]*CacheEntry),
	}
}

// Register registers a service instance and invalidates cache
func (cr *CachedRegistry) Register(ctx context.Context, instance *registry.ServiceInstance) error {
	err := cr.backend.Register(ctx, instance)
	if err != nil {
		return err
	}

	// Invalidate cache for this service
	cr.invalidateServiceCache(instance.ServiceName)
	cr.invalidateInstanceCache(instance.ID)

	return nil
}

// Deregister deregisters a service instance and invalidates cache
func (cr *CachedRegistry) Deregister(ctx context.Context, instanceID string) error {
	// Get instance to know which service to invalidate
	instance, _ := cr.backend.GetInstance(ctx, instanceID)

	err := cr.backend.Deregister(ctx, instanceID)
	if err != nil {
		return err
	}

	cr.invalidateInstanceCache(instanceID)
	if instance != nil {
		cr.invalidateServiceCache(instance.ServiceName)
	}

	return nil
}

// GetInstance retrieves a specific instance, using cache if available
func (cr *CachedRegistry) GetInstance(ctx context.Context, instanceID string) (*registry.ServiceInstance, error) {
	cr.cacheMu.RLock()
	entry := cr.instanceCache[instanceID]
	cr.cacheMu.RUnlock()

	if entry != nil && !entry.IsExpired() {
		instance := entry.Data.(*registry.ServiceInstance)
		cr.logger.Debug("instance cache hit", "instance_id", instanceID)
		return instance, nil
	}

	// Cache miss, fetch from backend
	instance, err := cr.backend.GetInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	// Update cache
	cr.cacheMu.Lock()
	cr.instanceCache[instanceID] = &CacheEntry{
		Data:      instance,
		ExpiresAt: time.Now().Add(cr.ttl),
	}
	cr.cacheMu.Unlock()

	return instance, nil
}

// GetInstances retrieves all healthy instances, using cache if available
func (cr *CachedRegistry) GetInstances(ctx context.Context, serviceName string) ([]*registry.ServiceInstance, error) {
	cr.cacheMu.RLock()
	entry := cr.instancesCache[serviceName]
	cr.cacheMu.RUnlock()

	if entry != nil && !entry.IsExpired() {
		instances := entry.Data.([]*registry.ServiceInstance)
		cr.logger.Debug("service cache hit", "service_name", serviceName, "count", len(instances))
		return instances, nil
	}

	// Cache miss, fetch from backend
	instances, err := cr.backend.GetInstances(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	// Update cache
	cr.cacheMu.Lock()
	cr.instancesCache[serviceName] = &CacheEntry{
		Data:      instances,
		ExpiresAt: time.Now().Add(cr.ttl),
	}
	cr.cacheMu.Unlock()

	return instances, nil
}

// GetInstancesByHealth retrieves instances filtered by health status
func (cr *CachedRegistry) GetInstancesByHealth(ctx context.Context, serviceName string, health registry.HealthStatus) ([]*registry.ServiceInstance, error) {
	// For now, no caching on filtered queries (could be optimized)
	return cr.backend.GetInstancesByHealth(ctx, serviceName, health)
}

// GetAllInstances retrieves all registered instances
func (cr *CachedRegistry) GetAllInstances(ctx context.Context) ([]*registry.ServiceInstance, error) {
	// Don't cache this as it's rarely needed in hot path
	return cr.backend.GetAllInstances(ctx)
}

// UpdateHealth updates an instance's health and invalidates cache
func (cr *CachedRegistry) UpdateHealth(ctx context.Context, instanceID string, health registry.HealthStatus) error {
	err := cr.backend.UpdateHealth(ctx, instanceID, health)
	if err != nil {
		return err
	}

	// Invalidate caches
	cr.invalidateInstanceCache(instanceID)

	// Also invalidate service cache (since health affects GetInstances results)
	instance, _ := cr.backend.GetInstance(ctx, instanceID)
	if instance != nil {
		cr.invalidateServiceCache(instance.ServiceName)
	}

	return nil
}

// UpdateMetadata updates an instance's metadata and invalidates cache
func (cr *CachedRegistry) UpdateMetadata(ctx context.Context, instanceID string, metadata map[string]string) error {
	err := cr.backend.UpdateMetadata(ctx, instanceID, metadata)
	if err != nil {
		return err
	}

	cr.invalidateInstanceCache(instanceID)

	return nil
}

// Heartbeat records a heartbeat (no cache invalidation needed)
func (cr *CachedRegistry) Heartbeat(ctx context.Context, instanceID string) error {
	return cr.backend.Heartbeat(ctx, instanceID)
}

// CleanupStale removes stale instances and invalidates all caches
func (cr *CachedRegistry) CleanupStale(ctx context.Context, ttl time.Duration) (int, error) {
	count, err := cr.backend.CleanupStale(ctx, ttl)
	if err != nil {
		return 0, err
	}

	// Invalidate all caches since we don't know which services were affected
	cr.invalidateAllCaches()

	return count, nil
}

// Close closes the backend
func (cr *CachedRegistry) Close() error {
	return cr.backend.Close()
}

// Helper methods

// invalidateServiceCache invalidates cache for a service
func (cr *CachedRegistry) invalidateServiceCache(serviceName string) {
	cr.cacheMu.Lock()
	delete(cr.instancesCache, serviceName)
	cr.cacheMu.Unlock()
}

// invalidateInstanceCache invalidates cache for an instance
func (cr *CachedRegistry) invalidateInstanceCache(instanceID string) {
	cr.cacheMu.Lock()
	delete(cr.instanceCache, instanceID)
	cr.cacheMu.Unlock()
}

// invalidateAllCaches invalidates all caches
func (cr *CachedRegistry) invalidateAllCaches() {
	cr.cacheMu.Lock()
	cr.instancesCache = make(map[string]*CacheEntry)
	cr.instanceCache = make(map[string]*CacheEntry)
	cr.cacheMu.Unlock()
}

// GetCacheStats returns cache statistics for debugging
func (cr *CachedRegistry) GetCacheStats() map[string]interface{} {
	cr.cacheMu.RLock()
	defer cr.cacheMu.RUnlock()

	return map[string]interface{}{
		"service_cache_entries":   len(cr.instancesCache),
		"instance_cache_entries":  len(cr.instanceCache),
		"total_cache_entries":     len(cr.instancesCache) + len(cr.instanceCache),
	}
}
