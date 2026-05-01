/**
 * API Client
 * Main client for making API requests
 * @packageDocumentation
 */
import { createClient } from '@connectrpc/connect';
import { getTransport, getConfig } from './transport.js';
// ============================================================================
// API CLIENT CLASS
// ============================================================================
/**
 * API Client for making ConnectRPC requests
 */
export class ApiClient {
    transport;
    serviceClients = new Map();
    constructor() {
        this.transport = getTransport();
    }
    /**
     * Gets or creates a client for a service
     */
    getService(service) {
        if (!this.serviceClients.has(service)) {
            const client = createClient(service, this.transport);
            this.serviceClients.set(service, client);
        }
        return this.serviceClients.get(service);
    }
    /**
     * Makes a unary RPC call
     */
    async call(service, method, input, options) {
        const client = this.getService(service);
        const methodFn = client[method];
        if (!methodFn) {
            throw createApiError('METHOD_NOT_FOUND', `Method ${method} not found on service`);
        }
        const callOptions = this.buildCallOptions(options);
        try {
            return await methodFn.call(client, input, callOptions);
        }
        catch (error) {
            throw this.handleError(error);
        }
    }
    /**
     * Makes a server streaming RPC call
     */
    async *stream(service, method, input, options) {
        const client = this.getService(service);
        const methodFn = client[method];
        if (!methodFn) {
            throw createApiError('METHOD_NOT_FOUND', `Method ${method} not found on service`);
        }
        const callOptions = this.buildCallOptions(options);
        try {
            const stream = methodFn.call(client, input, callOptions);
            for await (const message of stream) {
                yield message;
            }
        }
        catch (error) {
            throw this.handleError(error);
        }
    }
    /**
     * Builds call options from request options
     */
    buildCallOptions(options) {
        const config = getConfig();
        const callOptions = {};
        // Add headers
        const headers = new Headers();
        if (config.headers) {
            for (const [key, value] of Object.entries(config.headers)) {
                headers.set(key, value);
            }
        }
        if (options?.headers) {
            for (const [key, value] of Object.entries(options.headers)) {
                headers.set(key, value);
            }
        }
        callOptions.headers = headers;
        // Add abort signal
        if (options?.signal) {
            callOptions.signal = options.signal;
        }
        // Add timeout
        if (options?.timeout) {
            callOptions.timeoutMs = options.timeout;
        }
        return callOptions;
    }
    /**
     * Handles and transforms errors
     */
    handleError(error) {
        if (isApiError(error)) {
            return error;
        }
        // Handle ConnectRPC errors
        if (error instanceof Error) {
            const connectError = error;
            return {
                code: connectError.code ?? 'UNKNOWN',
                message: connectError.message,
                retryable: isRetryableError(connectError.code),
            };
        }
        return {
            code: 'UNKNOWN',
            message: 'An unknown error occurred',
            retryable: false,
        };
    }
    /**
     * Clears cached service clients
     */
    clearCache() {
        this.serviceClients.clear();
    }
}
// ============================================================================
// SINGLETON INSTANCE
// ============================================================================
let clientInstance = null;
/**
 * Gets the API client instance
 */
export function getApiClient() {
    if (!clientInstance) {
        clientInstance = new ApiClient();
    }
    return clientInstance;
}
/**
 * Resets the API client instance
 */
export function resetApiClient() {
    if (clientInstance) {
        clientInstance.clearCache();
        clientInstance = null;
    }
}
// ============================================================================
// HELPER FUNCTIONS
// ============================================================================
/**
 * Creates an API error
 */
export function createApiError(code, message, details) {
    return {
        code,
        message,
        details,
        retryable: isRetryableError(code),
    };
}
/**
 * Checks if an error is an ApiError
 */
export function isApiError(error) {
    return (typeof error === 'object' &&
        error !== null &&
        'code' in error &&
        'message' in error);
}
/**
 * Checks if an error code indicates a retryable error
 */
export function isRetryableError(code) {
    const retryableCodes = [
        'UNAVAILABLE',
        'DEADLINE_EXCEEDED',
        'RESOURCE_EXHAUSTED',
        'ABORTED',
        'INTERNAL',
    ];
    return code ? retryableCodes.includes(code) : false;
}
/**
 * Creates a typed service caller
 */
export function createServiceCaller(service) {
    return (method, input, options) => {
        const client = getApiClient();
        const serviceClient = client.getService(service);
        const methodFn = serviceClient[method];
        if (!methodFn) {
            throw createApiError('METHOD_NOT_FOUND', `Method ${String(method)} not found`);
        }
        return methodFn.call(serviceClient, input, options ? { headers: new Headers(options.headers) } : undefined);
    };
}
//# sourceMappingURL=client.js.map