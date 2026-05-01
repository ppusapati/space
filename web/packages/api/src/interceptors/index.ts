/**
 * Interceptors Module - Export all interceptors
 * @packageDocumentation
 */

// Auth Interceptor
export { createAuthInterceptor } from './auth.interceptor.js';

// Context Interceptor (organizational context: tenant, company, branch, user)
export { createContextInterceptor, createTenantInterceptor } from './tenant.interceptor.js';

// Error Interceptor
export {
  createErrorInterceptor,
  type ErrorInterceptorOptions,
} from './error.interceptor.js';

// Retry Interceptor
export {
  createRetryInterceptor,
  withRetry,
  createRetryWrapper,
} from './retry.interceptor.js';

// Logging Interceptor
export {
  createLoggingInterceptor,
  createStructuredLogger,
  type LogLevel,
  type LoggingOptions,
  type LogEntry,
} from './logging.interceptor.js';
