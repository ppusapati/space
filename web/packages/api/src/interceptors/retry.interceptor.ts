/**
 * Retry Interceptor
 * Handles automatic retry with exponential backoff
 * @packageDocumentation
 */

import type { Interceptor, ConnectError } from '@connectrpc/connect';
import type { RetryConfig } from '../types/index.js';

// ============================================================================
// DEFAULT CONFIGURATION
// ============================================================================

const DEFAULT_RETRY_CONFIG: RetryConfig = {
  maxRetries: 3,
  initialDelay: 1000,
  maxDelay: 10000,
  backoffMultiplier: 2,
  retryableStatuses: [408, 429, 500, 502, 503, 504],
  retryOnNetworkError: true,
};

/** ConnectRPC codes that are retryable */
const RETRYABLE_CODES = [
  'UNAVAILABLE',
  'DEADLINE_EXCEEDED',
  'RESOURCE_EXHAUSTED',
  'ABORTED',
  'INTERNAL',
];

// ============================================================================
// RETRY INTERCEPTOR
// ============================================================================

/**
 * Creates a retry interceptor with exponential backoff
 */
export function createRetryInterceptor(
  config: Partial<RetryConfig> = {}
): Interceptor {
  const retryConfig: RetryConfig = { ...DEFAULT_RETRY_CONFIG, ...config };

  return (next) => async (req) => {
    let lastError: unknown;
    let attempt = 0;

    while (attempt <= retryConfig.maxRetries) {
      try {
        return await next(req);
      } catch (error) {
        lastError = error;

        // Check if we should retry
        if (!shouldRetry(error, retryConfig, attempt)) {
          throw error;
        }

        // Calculate delay with exponential backoff and jitter
        const delay = calculateDelay(attempt, retryConfig);

        // Log retry attempt
        console.warn(`[API Retry] Attempt ${attempt + 1}/${retryConfig.maxRetries}`, {
          method: req.method.name,
          delay,
          error: getErrorCode(error),
        });

        // Wait before retrying
        await sleep(delay);

        attempt++;
      }
    }

    // All retries exhausted
    throw lastError;
  };
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

/**
 * Determines if a request should be retried
 */
function shouldRetry(
  error: unknown,
  config: RetryConfig,
  attempt: number
): boolean {
  // No more retries left
  if (attempt >= config.maxRetries) {
    return false;
  }

  // Check for network errors
  if (isNetworkError(error) && config.retryOnNetworkError) {
    return true;
  }

  // Check error code
  const code = getErrorCode(error);
  if (code && RETRYABLE_CODES.includes(code)) {
    return true;
  }

  return false;
}

/**
 * Calculates the delay for a retry attempt with jitter
 */
function calculateDelay(attempt: number, config: RetryConfig): number {
  // Exponential backoff
  const exponentialDelay =
    config.initialDelay * Math.pow(config.backoffMultiplier, attempt);

  // Cap at max delay
  const cappedDelay = Math.min(exponentialDelay, config.maxDelay);

  // Add jitter (±25%)
  const jitter = cappedDelay * 0.25 * (Math.random() * 2 - 1);

  return Math.floor(cappedDelay + jitter);
}

/**
 * Checks if error is a network error
 */
function isNetworkError(error: unknown): boolean {
  if (error instanceof TypeError) {
    // Fetch network errors are TypeErrors
    return error.message.includes('fetch') || error.message.includes('network');
  }
  return false;
}

/**
 * Gets the error code from various error types
 */
function getErrorCode(error: unknown): string | undefined {
  if (typeof error === 'object' && error !== null && 'code' in error) {
    return String((error as { code: unknown }).code);
  }
  return undefined;
}

/**
 * Sleeps for a given duration
 */
function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// ============================================================================
// RETRY UTILITIES
// ============================================================================

/**
 * Wraps a function with retry logic
 */
export async function withRetry<T>(
  fn: () => Promise<T>,
  config: Partial<RetryConfig> = {}
): Promise<T> {
  const retryConfig: RetryConfig = { ...DEFAULT_RETRY_CONFIG, ...config };
  let lastError: unknown;
  let attempt = 0;

  while (attempt <= retryConfig.maxRetries) {
    try {
      return await fn();
    } catch (error) {
      lastError = error;

      if (!shouldRetry(error, retryConfig, attempt)) {
        throw error;
      }

      const delay = calculateDelay(attempt, retryConfig);
      await sleep(delay);
      attempt++;
    }
  }

  throw lastError;
}

/**
 * Creates a retry wrapper for a specific function
 */
export function createRetryWrapper<T extends (...args: unknown[]) => Promise<unknown>>(
  fn: T,
  config: Partial<RetryConfig> = {}
): T {
  return (async (...args: Parameters<T>): Promise<ReturnType<T>> => {
    return withRetry(() => fn(...args) as Promise<ReturnType<T>>, config);
  }) as T;
}
