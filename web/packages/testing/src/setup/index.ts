/**
 * Test Setup Utilities
 * Configure vitest environment, global mocks, and cleanup helpers
 */
import { vi, beforeEach, afterEach } from 'vitest';

// ============================================================================
// VITEST SETUP HELPERS
// ============================================================================

/**
 * Standard test setup for Samavāya ERP tests.
 * Call in a vitest setup file or at the top of test suites.
 *
 * @example
 * // vitest.setup.ts
 * import { setupSamavayaTests } from '@samavāya/testing/setup';
 * setupSamavayaTests();
 */
export function setupSamavayaTests() {
  beforeEach(() => {
    vi.clearAllMocks();
    // Reset localStorage
    if (typeof localStorage !== 'undefined') {
      localStorage.clear();
    }
    // Reset sessionStorage
    if (typeof sessionStorage !== 'undefined') {
      sessionStorage.clear();
    }
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });
}

// ============================================================================
// VITEST GLOBAL CONFIG
// ============================================================================

/**
 * Vitest config preset for Samavāya packages.
 * Use in vitest.config.ts files.
 *
 * @example
 * // vitest.config.ts
 * import { defineConfig } from 'vitest/config';
 * import { samavayaVitestConfig } from '@samavāya/testing/setup';
 *
 * export default defineConfig({ test: samavayaVitestConfig });
 */
export const samavayaVitestConfig = {
  environment: 'jsdom',
  globals: true,
  setupFiles: [],
  coverage: {
    provider: 'v8' as const,
    reporter: ['text', 'json', 'html'],
    exclude: ['node_modules/', 'dist/', '**/*.d.ts', '**/*.config.*', '**/index.ts'],
  },
};

// ============================================================================
// RENDER HELPERS (Svelte)
// ============================================================================

/**
 * Wrapper around @testing-library/svelte render with common defaults.
 * Requires @testing-library/svelte to be installed.
 */
export async function renderComponent<T extends object>(
  Component: new (...args: unknown[]) => T,
  props?: Record<string, unknown>
) {
  const { render } = await import('@testing-library/svelte');
  return render(Component as Parameters<typeof render>[0], { props: props ?? {} });
}

// ============================================================================
// ASSERTION HELPERS
// ============================================================================

/** Wait for a condition to be true (polling) */
export async function waitFor(
  condition: () => boolean,
  timeout = 1000,
  interval = 50
): Promise<void> {
  const start = Date.now();
  while (!condition()) {
    if (Date.now() - start > timeout) {
      throw new Error(`waitFor timeout after ${timeout}ms`);
    }
    await new Promise(resolve => setTimeout(resolve, interval));
  }
}

/** Assert that an async function throws */
export async function expectToThrow(fn: () => Promise<unknown>, message?: string): Promise<Error> {
  try {
    await fn();
    throw new Error('Expected function to throw but it did not');
  } catch (err) {
    if (message) {
      const error = err as Error;
      if (!error.message.includes(message)) {
        throw new Error(`Expected error message to include "${message}" but got "${error.message}"`);
      }
    }
    return err as Error;
  }
}
