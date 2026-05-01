/**
 * ConnectRPC Transport Configuration
 * Creates and configures the transport for API communication
 * @packageDocumentation
 */
import type { Transport, Interceptor } from '@connectrpc/connect';
import type { ApiConfig } from '../types/index.js';
/**
 * Creates a ConnectRPC transport with the given configuration
 */
export declare function createTransport(config: ApiConfig): Transport;
/**
 * Gets the current transport instance
 */
export declare function getTransport(): Transport;
/**
 * Gets the current configuration
 */
export declare function getConfig(): ApiConfig;
/**
 * Updates the transport configuration
 */
export declare function updateConfig(updates: Partial<ApiConfig>): void;
/**
 * Adds a custom interceptor
 */
export declare function addInterceptor(interceptor: Interceptor): void;
/**
 * Removes a custom interceptor
 */
export declare function removeInterceptor(interceptor: Interceptor): void;
/**
 * Clears all custom interceptors
 */
export declare function clearInterceptors(): void;
/**
 * Checks if transport is initialized
 */
export declare function isTransportInitialized(): boolean;
/**
 * Destroys the current transport
 */
export declare function destroyTransport(): void;
//# sourceMappingURL=transport.d.ts.map