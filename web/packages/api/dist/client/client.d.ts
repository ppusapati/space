/**
 * API Client
 * Main client for making API requests
 * @packageDocumentation
 */
import type { DescService } from '@bufbuild/protobuf';
import type { Client } from '@connectrpc/connect';
import type { ApiError, RequestOptions } from '../types/index.js';
/**
 * API Client for making ConnectRPC requests
 */
export declare class ApiClient {
    private transport;
    private serviceClients;
    constructor();
    /**
     * Gets or creates a client for a service
     */
    getService<T extends DescService>(service: T): Client<T>;
    /**
     * Makes a unary RPC call
     */
    call<I extends object, O extends object>(service: DescService, method: string, input: I, options?: RequestOptions): Promise<O>;
    /**
     * Makes a server streaming RPC call
     */
    stream<I extends object, O extends object>(service: DescService, method: string, input: I, options?: RequestOptions): AsyncGenerator<O, void, unknown>;
    /**
     * Builds call options from request options
     */
    private buildCallOptions;
    /**
     * Handles and transforms errors
     */
    private handleError;
    /**
     * Clears cached service clients
     */
    clearCache(): void;
}
/**
 * Gets the API client instance
 */
export declare function getApiClient(): ApiClient;
/**
 * Resets the API client instance
 */
export declare function resetApiClient(): void;
/**
 * Creates an API error
 */
export declare function createApiError(code: string, message: string, details?: Record<string, unknown>): ApiError;
/**
 * Checks if an error is an ApiError
 */
export declare function isApiError(error: unknown): error is ApiError;
/**
 * Checks if an error code indicates a retryable error
 */
export declare function isRetryableError(code?: string): boolean;
/**
 * Creates a typed service caller
 */
export declare function createServiceCaller<TService extends DescService>(service: TService): <TMethod extends keyof Client<TService>>(method: TMethod, input: Parameters<Client<TService>[TMethod]>[0], options?: RequestOptions) => ReturnType<Client<TService>[TMethod]>;
//# sourceMappingURL=client.d.ts.map