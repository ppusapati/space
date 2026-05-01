/**
 * @samavāya/api
 * ConnectRPC API client for samavāya ERP
 * @packageDocumentation
 */
export * from './types/index.js';
export * from './client/index.js';
export * from './interceptors/index.js';
export * from './utils/index.js';
export * from './services/index.js';
import type { ApiConfig } from './types/index.js';
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
export declare function initializeApi(options: InitOptions): void;
//# sourceMappingURL=index.d.ts.map