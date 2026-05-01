/**
 * API Client
 * Main client for making API requests
 * @packageDocumentation
 */

import { createClient } from '@connectrpc/connect';
import type { DescService } from '@bufbuild/protobuf';
import type { Transport, CallOptions, Client } from '@connectrpc/connect';
import { getTransport, getConfig } from './transport.js';
import type {
  ApiError,
  RequestOptions,
} from '../types/index.js';

// ============================================================================
// API CLIENT CLASS
// ============================================================================

/**
 * API Client for making ConnectRPC requests
 */
export class ApiClient {
  private transport: Transport;
  private serviceClients: Map<DescService, unknown> = new Map();

  constructor() {
    this.transport = getTransport();
  }

  /**
   * Gets or creates a client for a service
   */
  getService<T extends DescService>(service: T): Client<T> {
    if (!this.serviceClients.has(service)) {
      const client = createClient(service, this.transport);
      this.serviceClients.set(service, client);
    }
    return this.serviceClients.get(service) as Client<T>;
  }

  /**
   * Makes a unary RPC call
   */
  async call<I extends object, O extends object>(
    service: DescService,
    method: string,
    input: I,
    options?: RequestOptions
  ): Promise<O> {
    const client = this.getService(service) as Record<string, (input: I, options?: CallOptions) => Promise<O>>;
    const methodFn = client[method];

    if (!methodFn) {
      throw createApiError('METHOD_NOT_FOUND', `Method ${method} not found on service`);
    }

    const callOptions = this.buildCallOptions(options);

    try {
      return await methodFn.call(client, input, callOptions);
    } catch (error) {
      throw this.handleError(error);
    }
  }

  /**
   * Makes a server streaming RPC call
   */
  async *stream<I extends object, O extends object>(
    service: DescService,
    method: string,
    input: I,
    options?: RequestOptions
  ): AsyncGenerator<O, void, unknown> {
    const client = this.getService(service) as Record<string, (input: I, options?: CallOptions) => AsyncIterable<O>>;
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
    } catch (error) {
      throw this.handleError(error);
    }
  }

  /**
   * Builds call options from request options
   */
  private buildCallOptions(options?: RequestOptions): CallOptions {
    const config = getConfig();
    const callOptions: CallOptions = {};

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
  private handleError(error: unknown): ApiError {
    if (isApiError(error)) {
      return error;
    }

    // Handle ConnectRPC errors
    if (error instanceof Error) {
      const connectError = error as { code?: string; message: string; metadata?: Headers };

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
  clearCache(): void {
    this.serviceClients.clear();
  }
}

// ============================================================================
// SINGLETON INSTANCE
// ============================================================================

let clientInstance: ApiClient | null = null;

/**
 * Gets the API client instance
 */
export function getApiClient(): ApiClient {
  if (!clientInstance) {
    clientInstance = new ApiClient();
  }
  return clientInstance;
}

/**
 * Resets the API client instance
 */
export function resetApiClient(): void {
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
export function createApiError(
  code: string,
  message: string,
  details?: Record<string, unknown>
): ApiError {
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
export function isApiError(error: unknown): error is ApiError {
  return (
    typeof error === 'object' &&
    error !== null &&
    'code' in error &&
    'message' in error
  );
}

/**
 * Checks if an error code indicates a retryable error
 */
export function isRetryableError(code?: string): boolean {
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
export function createServiceCaller<TService extends DescService>(
  service: TService
) {
  return <TMethod extends keyof Client<TService>>(
    method: TMethod,
    input: Parameters<Client<TService>[TMethod]>[0],
    options?: RequestOptions
  ): ReturnType<Client<TService>[TMethod]> => {
    const client = getApiClient();
    const serviceClient = client.getService(service);
    const methodFn = serviceClient[method] as (...args: unknown[]) => unknown;

    if (!methodFn) {
      throw createApiError('METHOD_NOT_FOUND', `Method ${String(method)} not found`);
    }

    return methodFn.call(serviceClient, input, options ? { headers: new Headers(options.headers) } : undefined) as ReturnType<Client<TService>[TMethod]>;
  };
}
