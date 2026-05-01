/**
 * Cache Utilities
 * Client-side caching for API responses
 * @packageDocumentation
 */
// ============================================================================
// MEMORY CACHE STORAGE
// ============================================================================
/**
 * In-memory cache storage
 */
class MemoryCacheStorage {
    cache = new Map();
    maxSize;
    constructor(maxSize) {
        this.maxSize = maxSize;
    }
    get(key) {
        const entry = this.cache.get(key);
        return entry;
    }
    set(key, entry) {
        // Evict oldest entries if at capacity
        if (this.cache.size >= this.maxSize && !this.cache.has(key)) {
            const oldestKey = this.cache.keys().next().value;
            if (oldestKey) {
                this.cache.delete(oldestKey);
            }
        }
        this.cache.set(key, entry);
    }
    delete(key) {
        this.cache.delete(key);
    }
    clear() {
        this.cache.clear();
    }
    keys() {
        return Array.from(this.cache.keys());
    }
    size() {
        return this.cache.size;
    }
}
// ============================================================================
// WEB STORAGE CACHE
// ============================================================================
/**
 * Web storage (localStorage/sessionStorage) cache
 */
class WebStorageCache {
    storage;
    prefix;
    maxSize;
    constructor(storage, maxSize, prefix = 'api_cache_') {
        this.storage = storage;
        this.prefix = prefix;
        this.maxSize = maxSize;
    }
    get(key) {
        try {
            const item = this.storage.getItem(this.prefix + key);
            if (!item)
                return null;
            return JSON.parse(item);
        }
        catch {
            return null;
        }
    }
    set(key, entry) {
        try {
            // Check size and evict if necessary
            const keys = this.keys();
            if (keys.length >= this.maxSize && !keys.includes(key)) {
                // Remove oldest entry
                const entries = keys
                    .map((k) => ({ key: k, entry: this.get(k) }))
                    .filter((e) => e.entry !== null)
                    .sort((a, b) => (a.entry?.timestamp ?? 0) - (b.entry?.timestamp ?? 0));
                if (entries[0]) {
                    this.delete(entries[0].key);
                }
            }
            this.storage.setItem(this.prefix + key, JSON.stringify(entry));
        }
        catch (error) {
            // Storage full, clear old entries
            console.warn('Cache storage full, clearing old entries');
            this.clearExpired();
            try {
                this.storage.setItem(this.prefix + key, JSON.stringify(entry));
            }
            catch {
                // Still full, clear all
                this.clear();
            }
        }
    }
    delete(key) {
        this.storage.removeItem(this.prefix + key);
    }
    clear() {
        const keys = this.keys();
        for (const key of keys) {
            this.storage.removeItem(this.prefix + key);
        }
    }
    keys() {
        const keys = [];
        for (let i = 0; i < this.storage.length; i++) {
            const key = this.storage.key(i);
            if (key?.startsWith(this.prefix)) {
                keys.push(key.slice(this.prefix.length));
            }
        }
        return keys;
    }
    size() {
        return this.keys().length;
    }
    /**
     * Clears expired entries
     */
    clearExpired() {
        const now = Date.now();
        const keys = this.keys();
        for (const key of keys) {
            const entry = this.get(key);
            if (entry && now > entry.timestamp + entry.ttl * 1000) {
                this.delete(key);
            }
        }
    }
}
// ============================================================================
// API CACHE
// ============================================================================
/**
 * API response cache
 */
export class ApiCache {
    storage;
    config;
    constructor(config = {}) {
        this.config = {
            enabled: config.enabled ?? true,
            defaultTtl: config.defaultTtl ?? 300,
            maxSize: config.maxSize ?? 100,
            storage: config.storage ?? 'memory',
        };
        this.storage = this.createStorage();
    }
    /**
     * Creates the appropriate storage backend
     */
    createStorage() {
        switch (this.config.storage) {
            case 'localStorage':
                if (typeof localStorage !== 'undefined') {
                    return new WebStorageCache(localStorage, this.config.maxSize);
                }
                break;
            case 'sessionStorage':
                if (typeof sessionStorage !== 'undefined') {
                    return new WebStorageCache(sessionStorage, this.config.maxSize);
                }
                break;
        }
        return new MemoryCacheStorage(this.config.maxSize);
    }
    /**
     * Gets a cached response
     */
    get(key) {
        if (!this.config.enabled)
            return null;
        const entry = this.storage.get(key);
        if (!entry)
            return null;
        // Check if expired
        const now = Date.now();
        if (now > entry.timestamp + entry.ttl * 1000) {
            this.storage.delete(key);
            return null;
        }
        return entry.data;
    }
    /**
     * Sets a cached response
     */
    set(key, data, ttl) {
        if (!this.config.enabled)
            return;
        const entry = {
            data,
            timestamp: Date.now(),
            ttl: ttl ?? this.config.defaultTtl,
            key,
        };
        this.storage.set(key, entry);
    }
    /**
     * Deletes a cached response
     */
    delete(key) {
        this.storage.delete(key);
    }
    /**
     * Deletes all cached responses matching a pattern
     */
    deletePattern(pattern) {
        const regex = typeof pattern === 'string' ? new RegExp(pattern) : pattern;
        const keys = this.storage.keys();
        for (const key of keys) {
            if (regex.test(key)) {
                this.storage.delete(key);
            }
        }
    }
    /**
     * Clears all cached responses
     */
    clear() {
        this.storage.clear();
    }
    /**
     * Gets cache statistics
     */
    stats() {
        return {
            size: this.storage.size(),
            maxSize: this.config.maxSize,
            enabled: this.config.enabled,
        };
    }
    /**
     * Checks if a key is cached
     */
    has(key) {
        return this.get(key) !== null;
    }
    /**
     * Gets or sets a cached value
     */
    async getOrSet(key, fetcher, ttl) {
        const cached = this.get(key);
        if (cached !== null) {
            return cached;
        }
        const data = await fetcher();
        this.set(key, data, ttl);
        return data;
    }
}
// ============================================================================
// CACHE KEY GENERATION
// ============================================================================
/**
 * Generates a cache key from request parameters
 */
export function generateCacheKey(method, params) {
    const parts = [method];
    if (params && Object.keys(params).length > 0) {
        // Sort keys for consistent ordering
        const sortedParams = Object.keys(params)
            .sort()
            .reduce((acc, key) => {
            acc[key] = params[key];
            return acc;
        }, {});
        parts.push(hashObject(sortedParams));
    }
    return parts.join(':');
}
/**
 * Creates a simple hash of an object
 */
function hashObject(obj) {
    const str = JSON.stringify(obj);
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
        const char = str.charCodeAt(i);
        hash = (hash << 5) - hash + char;
        hash = hash & hash; // Convert to 32bit integer
    }
    return Math.abs(hash).toString(36);
}
// ============================================================================
// SINGLETON INSTANCE
// ============================================================================
let cacheInstance = null;
/**
 * Gets the API cache instance
 */
export function getApiCache(config) {
    if (!cacheInstance || config) {
        cacheInstance = new ApiCache(config);
    }
    return cacheInstance;
}
/**
 * Resets the cache instance
 */
export function resetApiCache() {
    if (cacheInstance) {
        cacheInstance.clear();
        cacheInstance = null;
    }
}
//# sourceMappingURL=cache.js.map