/**
 * ConnectRPC Transport Configuration
 * Creates and configures the transport for API communication
 * @packageDocumentation
 */

import { createConnectTransport } from '@connectrpc/connect-web';
import type { Transport, Interceptor } from '@connectrpc/connect';
import type { ApiConfig } from '../types/index.js';

// ============================================================================
// DEFAULT CONFIGURATION
// ============================================================================

const DEFAULT_CONFIG: Required<Omit<ApiConfig, 'baseUrl' | 'headers'>> = {
  timeout: 30000,
  // Bearer-token auth (Authorization header) — never cookie-based.
  // 'omit' avoids the CORS wildcard-origin trap: when credentials:'include'
  // is sent, the browser silently rejects responses whose
  // Access-Control-Allow-Origin is '*' (it requires the exact origin AND
  // Access-Control-Allow-Credentials:true). Our backend's CORS uses '*'
  // for dev simplicity, so any 'include' fetch fails as "Failed to fetch"
  // before the response is delivered. Tokens travel via Authorization
  // header which is unaffected.
  credentials: 'omit',
  debug: false,
  retry: {
    maxRetries: 3,
    initialDelay: 1000,
    maxDelay: 10000,
    backoffMultiplier: 2,
    retryableStatuses: [408, 429, 500, 502, 503, 504],
    retryOnNetworkError: true,
  },
  cache: {
    enabled: false,
    defaultTtl: 300,
    maxSize: 100,
    storage: 'memory',
  },
};

// ============================================================================
// TRANSPORT CREATION
// ============================================================================

/** Stored configuration for transport */
let currentConfig: ApiConfig | null = null;

/** Stored transport instance */
let transportInstance: Transport | null = null;

/** Custom interceptors */
const customInterceptors: Interceptor[] = [];

/**
 * Creates a ConnectRPC transport with the given configuration
 */
export function createTransport(config: ApiConfig): Transport {
  currentConfig = { ...DEFAULT_CONFIG, ...config };

  const interceptors: Interceptor[] = [
    ...customInterceptors,
  ];

  // Add debug interceptor if enabled
  if (currentConfig.debug) {
    interceptors.push(createDebugInterceptor());
  }

  // Add timeout interceptor
  if (currentConfig.timeout) {
    interceptors.push(createTimeoutInterceptor(currentConfig.timeout));
  }

  transportInstance = createConnectTransport({
    baseUrl: currentConfig.baseUrl,
    fetch: (input, init) => fetch(input, { ...init, credentials: currentConfig?.credentials }),
    interceptors,
  });

  return transportInstance;
}

/**
 * Gets the current transport instance
 */
export function getTransport(): Transport {
  if (!transportInstance) {
    throw new Error('Transport not initialized. Call createTransport first.');
  }
  return transportInstance;
}

/**
 * Gets the current configuration
 */
export function getConfig(): ApiConfig {
  if (!currentConfig) {
    throw new Error('Transport not initialized. Call createTransport first.');
  }
  return currentConfig;
}

/**
 * Updates the transport configuration
 */
export function updateConfig(updates: Partial<ApiConfig>): void {
  if (!currentConfig) {
    throw new Error('Transport not initialized. Call createTransport first.');
  }

  currentConfig = { ...currentConfig, ...updates };

  // Recreate transport with new config
  createTransport(currentConfig);
}

/**
 * Adds a custom interceptor
 */
export function addInterceptor(interceptor: Interceptor): void {
  customInterceptors.push(interceptor);

  // Recreate transport if already initialized
  if (currentConfig) {
    createTransport(currentConfig);
  }
}

/**
 * Removes a custom interceptor
 */
export function removeInterceptor(interceptor: Interceptor): void {
  const index = customInterceptors.indexOf(interceptor);
  if (index > -1) {
    customInterceptors.splice(index, 1);

    // Recreate transport if already initialized
    if (currentConfig) {
      createTransport(currentConfig);
    }
  }
}

/**
 * Clears all custom interceptors
 */
export function clearInterceptors(): void {
  customInterceptors.length = 0;

  // Recreate transport if already initialized
  if (currentConfig) {
    createTransport(currentConfig);
  }
}

// ============================================================================
// BUILT-IN INTERCEPTORS
// ============================================================================

/**
 * Creates a debug interceptor that logs requests and responses
 */
function createDebugInterceptor(): Interceptor {
  return (next) => async (req) => {
    const startTime = performance.now();
    const requestId = crypto.randomUUID();

    console.group(`[API] ${req.method.name}`);
    console.log('Request ID:', requestId);
    console.log('URL:', req.url);
    console.log('Message:', req.message);

    try {
      const response = await next(req);
      const duration = performance.now() - startTime;

      console.log('Response:', response.message);
      console.log('Duration:', `${duration.toFixed(2)}ms`);
      console.groupEnd();

      return response;
    } catch (error) {
      const duration = performance.now() - startTime;

      console.error('Error:', error);
      console.log('Duration:', `${duration.toFixed(2)}ms`);
      console.groupEnd();

      throw error;
    }
  };
}

/**
 * Creates a timeout interceptor
 */
function createTimeoutInterceptor(timeout: number): Interceptor {
  return (next) => async (req) => {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);

    // Merge abort signals if one already exists
    if (req.signal) {
      req.signal.addEventListener('abort', () => controller.abort());
    }

    try {
      const response = await next({
        ...req,
        signal: controller.signal,
      });
      clearTimeout(timeoutId);
      return response;
    } catch (error) {
      clearTimeout(timeoutId);

      if (controller.signal.aborted) {
        throw new Error(`Request timeout after ${timeout}ms`);
      }

      throw error;
    }
  };
}

// ============================================================================
// TRANSPORT UTILITIES
// ============================================================================

/**
 * Checks if transport is initialized
 */
export function isTransportInitialized(): boolean {
  return transportInstance !== null;
}

/**
 * Destroys the current transport
 */
export function destroyTransport(): void {
  transportInstance = null;
  currentConfig = null;
}
