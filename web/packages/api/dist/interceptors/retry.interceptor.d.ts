/**
 * Retry Interceptor
 * Handles automatic retry with exponential backoff
 * @packageDocumentation
 */
import type { Interceptor } from '@connectrpc/connect';
import type { RetryConfig } from '../types/index.js';
/**
 * Creates a retry interceptor with exponential backoff
 */
export declare function createRetryInterceptor(config?: Partial<RetryConfig>): Interceptor;
/**
 * Wraps a function with retry logic
 */
export declare function withRetry<T>(fn: () => Promise<T>, config?: Partial<RetryConfig>): Promise<T>;
/**
 * Creates a retry wrapper for a specific function
 */
export declare function createRetryWrapper<T extends (...args: unknown[]) => Promise<unknown>>(fn: T, config?: Partial<RetryConfig>): T;
//# sourceMappingURL=retry.interceptor.d.ts.map