/**
 * Cache Utilities
 * Client-side caching for API responses
 * @packageDocumentation
 */
import type { CacheConfig } from '../types/index.js';
/**
 * API response cache
 */
export declare class ApiCache {
    private storage;
    private config;
    constructor(config?: Partial<CacheConfig>);
    /**
     * Creates the appropriate storage backend
     */
    private createStorage;
    /**
     * Gets a cached response
     */
    get<T>(key: string): T | null;
    /**
     * Sets a cached response
     */
    set<T>(key: string, data: T, ttl?: number): void;
    /**
     * Deletes a cached response
     */
    delete(key: string): void;
    /**
     * Deletes all cached responses matching a pattern
     */
    deletePattern(pattern: string | RegExp): void;
    /**
     * Clears all cached responses
     */
    clear(): void;
    /**
     * Gets cache statistics
     */
    stats(): {
        size: number;
        maxSize: number;
        enabled: boolean;
    };
    /**
     * Checks if a key is cached
     */
    has(key: string): boolean;
    /**
     * Gets or sets a cached value
     */
    getOrSet<T>(key: string, fetcher: () => Promise<T>, ttl?: number): Promise<T>;
}
/**
 * Generates a cache key from request parameters
 */
export declare function generateCacheKey(method: string, params?: Record<string, unknown>): string;
/**
 * Gets the API cache instance
 */
export declare function getApiCache(config?: Partial<CacheConfig>): ApiCache;
/**
 * Resets the cache instance
 */
export declare function resetApiCache(): void;
//# sourceMappingURL=cache.d.ts.map