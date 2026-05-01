/**
 * @samavāya/testing
 * Testing utilities for Samavāya ERP
 *
 * Provides fixtures, mocks, and helpers used across all packages and apps.
 *
 * @example
 * import { createMockUser, createMockTenant, mockApiClient } from '@samavāya/testing';
 * import { mockFormSchema, createMockEntity } from '@samavāya/testing/fixtures';
 * import { mockConnectRPC, mockFetch } from '@samavāya/testing/mocks';
 *
 * @packageDocumentation
 */

// Fixtures
export * from './fixtures/index';

// Mocks
export * from './mocks/index';

// Setup utilities
export * from './setup/index';

// Re-export testing utilities for convenience
export { vi, expect, describe, it, test, beforeEach, afterEach, beforeAll, afterAll } from 'vitest';
