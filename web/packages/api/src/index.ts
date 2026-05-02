/**
 * @chetana/api
 * ConnectRPC API client for Chetana ERP
 * @packageDocumentation
 */

// ============================================================================
// PROVIDERS (dependency inversion for stores)
// ============================================================================
export * from './providers.js';

// ============================================================================
// TYPES
// ============================================================================
export * from './types/index.js';

// ============================================================================
// CLIENT
// ============================================================================
export * from './client/index.js';

// ============================================================================
// INTERCEPTORS
// ============================================================================
export * from './interceptors/index.js';

// ============================================================================
// UTILITIES
// ============================================================================
export * from './utils/index.js';

// ============================================================================
// SERVICES (typed ConnectRPC client factories)
// ============================================================================
export * from './services/index.js';

// ============================================================================
// INITIALIZATION HELPER
// ============================================================================

import type { ApiConfig } from './types/index.js';
import { createTransport, addInterceptor } from './client/index.js';
import { createAuthInterceptor } from './interceptors/auth.interceptor.js';
import { createTenantInterceptor } from './interceptors/tenant.interceptor.js';
import { createErrorInterceptor } from './interceptors/error.interceptor.js';
import { createRetryInterceptor } from './interceptors/retry.interceptor.js';
import { createLoggingInterceptor } from './interceptors/logging.interceptor.js';

/** Initialization options */
export interface InitOptions extends ApiConfig {
  /** Whether to add auth interceptor */
  withAuth?: boolean;

  /** Whether to add tenant interceptor */
  withTenant?: boolean;

  /** Whether to add error interceptor */
  withErrorHandling?: boolean;

  /** Whether to add retry interceptor */
  withRetry?: boolean;

  /** Whether to add logging interceptor */
  withLogging?: boolean;
}

/**
 * Initializes the API client with common interceptors
 */
export function initializeApi(options: InitOptions): void {
  const {
    withAuth = true,
    withTenant = true,
    withErrorHandling = true,
    withRetry = true,
    withLogging = false,
    ...config
  } = options;

  // Create transport
  createTransport(config);

  // Add interceptors in order (first added = runs first)
  if (withLogging) {
    addInterceptor(createLoggingInterceptor({ level: config.debug ? 'debug' : 'info' }));
  }

  if (withRetry) {
    addInterceptor(createRetryInterceptor(config.retry));
  }

  if (withErrorHandling) {
    addInterceptor(createErrorInterceptor());
  }

  if (withTenant) {
    addInterceptor(createTenantInterceptor());
  }

  if (withAuth) {
    addInterceptor(createAuthInterceptor());
  }
}
