/**
 * Test Fixtures
 * Pre-built data objects for use in tests
 */

import type { BaseEntity } from '@samavāya/types';

// ============================================================================
// ENTITY FIXTURES
// ============================================================================

export function createMockEntity<T extends object>(
  overrides?: Partial<T & BaseEntity>
): T & BaseEntity {
  return {
    id: `mock-${Math.random().toString(36).slice(2, 9)}`,
    createdAt: new Date('2026-01-01T00:00:00Z'),
    updatedAt: new Date('2026-01-01T00:00:00Z'),
    ...overrides,
  } as T & BaseEntity;
}

// ============================================================================
// USER / AUTH FIXTURES
// ============================================================================

export interface MockUser {
  id: string;
  name: string;
  email: string;
  tenantId: string;
  vertical: 'agriculture' | 'manufacturing' | 'water' | 'construction';
  roles: string[];
}

export function createMockUser(overrides?: Partial<MockUser>): MockUser {
  return {
    id: 'user-001',
    name: 'Test User',
    email: 'test@samavaya.example',
    tenantId: 'tenant-001',
    vertical: 'agriculture',
    roles: ['user'],
    ...overrides,
  };
}

// ============================================================================
// TENANT FIXTURES
// ============================================================================

export interface MockTenant {
  id: string;
  name: string;
  vertical: string;
  plan: string;
}

export function createMockTenant(overrides?: Partial<MockTenant>): MockTenant {
  return {
    id: 'tenant-001',
    name: 'Test Organisation',
    vertical: 'agriculture',
    plan: 'standard',
    ...overrides,
  };
}

// ============================================================================
// FORM FIXTURES
// ============================================================================

export function createMockFormSchema(overrides?: object): object {
  return {
    id: 'mock-form',
    title: 'Mock Form',
    fields: [
      {
        name: 'name',
        type: 'text',
        label: 'Name',
        required: true,
      },
      {
        name: 'email',
        type: 'email',
        label: 'Email',
        required: true,
      },
    ],
    layout: { type: 'vertical' },
    ...overrides,
  };
}

// ============================================================================
// PAGINATION FIXTURES
// ============================================================================

export function createMockPagination(overrides?: object) {
  return {
    page: 1,
    pageSize: 10,
    total: 100,
    ...overrides,
  };
}

// ============================================================================
// API RESPONSE FIXTURES
// ============================================================================

export function createMockListResponse<T>(items: T[], total?: number) {
  return {
    items,
    total: total ?? items.length,
    page: 1,
    pageSize: 10,
  };
}

export function createMockErrorResponse(message = 'Mock error', code = 'MOCK_ERROR') {
  return {
    error: { code, message },
  };
}
