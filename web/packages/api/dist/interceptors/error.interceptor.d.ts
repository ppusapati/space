/**
 * Error Interceptor
 * Standardizes error handling and notifications
 * @packageDocumentation
 */
import type { Interceptor } from '@connectrpc/connect';
import type { ApiError } from '../types/index.js';
/** Options for the error interceptor */
export interface ErrorInterceptorOptions {
    /** Whether to show toast notifications for errors */
    showToasts?: boolean;
    /** Whether to log errors to console */
    logErrors?: boolean;
    /** Custom error handler */
    onError?: (error: ApiError) => void;
    /** Error codes to suppress from notifications */
    suppressCodes?: string[];
}
/**
 * Creates an error interceptor
 */
export declare function createErrorInterceptor(options?: ErrorInterceptorOptions): Interceptor;
//# sourceMappingURL=error.interceptor.d.ts.map