/**
 * Test Mocks
 * Vi mocks for API clients, stores, and external services
 */
import { vi } from 'vitest';

// ============================================================================
// API CLIENT MOCK
// ============================================================================

/** Mock a ConnectRPC service method */
export function mockRPCMethod<T>(returnValue: T) {
  return vi.fn().mockResolvedValue(returnValue);
}

/** Create a mock API client with all methods returning undefined by default */
export function createMockApiClient<T extends object>(methods: Partial<T> = {}): T {
  return new Proxy(methods as T, {
    get(target, prop) {
      if (prop in target) return (target as Record<string | symbol, unknown>)[prop];
      return vi.fn().mockResolvedValue(undefined);
    },
  });
}

// ============================================================================
// FETCH MOCK
// ============================================================================

/** Mock global fetch with a JSON response */
export function mockFetch(response: unknown, status = 200) {
  return vi.spyOn(global, 'fetch').mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(response),
    text: () => Promise.resolve(JSON.stringify(response)),
    headers: new Headers(),
  } as Response);
}

// ============================================================================
// STORE MOCK
// ============================================================================

/** Create a writable-like mock store for testing */
export function createMockStore<T>(initialValue: T) {
  let value = initialValue;
  const subscribers = new Set<(v: T) => void>();

  return {
    subscribe(fn: (v: T) => void) {
      subscribers.add(fn);
      fn(value);
      return () => subscribers.delete(fn);
    },
    set(newValue: T) {
      value = newValue;
      subscribers.forEach(fn => fn(value));
    },
    get: () => value,
    update(fn: (v: T) => T) {
      this.set(fn(value));
    },
  };
}

// ============================================================================
// ROUTER MOCK
// ============================================================================

/** Mock SvelteKit navigation */
export function createMockNavigation() {
  return {
    goto: vi.fn().mockResolvedValue(undefined),
    back: vi.fn(),
    forward: vi.fn(),
    preloadData: vi.fn(),
    preloadCode: vi.fn(),
    invalidate: vi.fn(),
    invalidateAll: vi.fn(),
  };
}

/** Mock SvelteKit page store */
export function createMockPage(overrides?: object) {
  return {
    url: new URL('http://localhost/'),
    params: {},
    route: { id: '/' },
    status: 200,
    error: null,
    data: {},
    form: null,
    state: {},
    ...overrides,
  };
}

// ============================================================================
// TIMER MOCKS
// ============================================================================

export const useFakeTimers = () => {
  vi.useFakeTimers();
  return () => vi.useRealTimers();
};
