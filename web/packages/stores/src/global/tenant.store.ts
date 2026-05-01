/**
 * Tenant Store
 * Handles multi-tenancy state, tenant switching, and tenant-specific settings
 */

import { writable, derived, get, type Writable, type Readable } from 'svelte/store';

// ============================================================================
// TYPES
// ============================================================================

export interface Tenant {
  id: string;
  code: string;
  name: string;
  displayName: string;
  logo?: string;
  favicon?: string;
  domain?: string;
  settings: TenantSettings;
  features: TenantFeatures;
  limits: TenantLimits;
  metadata?: Record<string, unknown>;
  status: 'active' | 'suspended' | 'trial' | 'expired';
  plan: 'free' | 'starter' | 'professional' | 'enterprise';
  trialEndsAt?: Date;
  createdAt: Date;
}

export interface TenantSettings {
  // Branding
  primaryColor?: string;
  secondaryColor?: string;
  logoUrl?: string;
  faviconUrl?: string;

  // Localization
  defaultLocale: string;
  supportedLocales: string[];
  defaultTimezone: string;
  defaultCurrency: string;
  dateFormat: string;
  timeFormat: '12h' | '24h';
  numberFormat: {
    decimal: string;
    thousands: string;
  };

  // Security
  passwordPolicy: {
    minLength: number;
    requireUppercase: boolean;
    requireLowercase: boolean;
    requireNumbers: boolean;
    requireSpecialChars: boolean;
    maxAge: number; // Days
  };
  sessionTimeout: number; // Minutes
  mfaRequired: boolean;

  // Email
  emailFromName?: string;
  emailFromAddress?: string;

  // Custom settings
  custom?: Record<string, unknown>;
}

export interface TenantFeatures {
  // Core modules
  identity: boolean;
  workflow: boolean;
  notifications: boolean;
  audit: boolean;
  data: boolean;
  insights: boolean;
  platform: boolean;

  // Business modules
  masters: boolean;
  finance: boolean;
  hr: boolean;
  purchase: boolean;
  inventory: boolean;
  fulfillment: boolean;
  manufacturing: boolean;
  sales: boolean;
  projects: boolean;
  budget: boolean;
  banking: boolean;
  communication: boolean;
  asset: boolean;

  // Premium features
  advancedReporting: boolean;
  customFields: boolean;
  apiAccess: boolean;
  webhooks: boolean;
  sso: boolean;
  customDomain: boolean;
  whiteLabel: boolean;
  prioritySupport: boolean;
}

export interface TenantLimits {
  users: number;
  storage: number; // GB
  apiRequestsPerMonth: number;
  workflowsPerMonth: number;
  documentsPerMonth: number;
  customFieldsPerEntity: number;
  webhooksCount: number;
  integrations: number;
}

export interface TenantUsage {
  users: number;
  storage: number;
  apiRequests: number;
  workflows: number;
  documents: number;
}

export interface TenantState {
  current: Tenant | null;
  available: Tenant[];
  isLoading: boolean;
  isSwitching: boolean;
  error: TenantError | null;
  usage?: TenantUsage;
}

export interface TenantError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

// ============================================================================
// INITIAL STATE
// ============================================================================

const defaultSettings: TenantSettings = {
  defaultLocale: 'en',
  supportedLocales: ['en'],
  defaultTimezone: 'UTC',
  defaultCurrency: 'USD',
  dateFormat: 'YYYY-MM-DD',
  timeFormat: '24h',
  numberFormat: {
    decimal: '.',
    thousands: ',',
  },
  passwordPolicy: {
    minLength: 8,
    requireUppercase: true,
    requireLowercase: true,
    requireNumbers: true,
    requireSpecialChars: false,
    maxAge: 90,
  },
  sessionTimeout: 30,
  mfaRequired: false,
};

const defaultFeatures: TenantFeatures = {
  identity: true,
  workflow: true,
  notifications: true,
  audit: true,
  data: true,
  insights: false,
  platform: true,
  masters: true,
  finance: false,
  hr: false,
  purchase: false,
  inventory: false,
  fulfillment: false,
  manufacturing: false,
  sales: false,
  projects: false,
  budget: false,
  banking: false,
  communication: false,
  asset: false,
  advancedReporting: false,
  customFields: false,
  apiAccess: false,
  webhooks: false,
  sso: false,
  customDomain: false,
  whiteLabel: false,
  prioritySupport: false,
};

const defaultLimits: TenantLimits = {
  users: 5,
  storage: 1,
  apiRequestsPerMonth: 10000,
  workflowsPerMonth: 100,
  documentsPerMonth: 500,
  customFieldsPerEntity: 5,
  webhooksCount: 3,
  integrations: 2,
};

const initialState: TenantState = {
  current: null,
  available: [],
  isLoading: false,
  isSwitching: false,
  error: null,
};

// ============================================================================
// STORE CREATION
// ============================================================================

function createTenantStore() {
  const store = writable<TenantState>(initialState);
  const { subscribe, set, update } = store;

  // ============================================================================
  // DERIVED STORES
  // ============================================================================

  const current: Readable<Tenant | null> = derived(store, ($s) => $s.current);
  const available: Readable<Tenant[]> = derived(store, ($s) => $s.available);
  const isLoading: Readable<boolean> = derived(store, ($s) => $s.isLoading);
  const settings: Readable<TenantSettings | null> = derived(store, ($s) => $s.current?.settings ?? null);
  const features: Readable<TenantFeatures | null> = derived(store, ($s) => $s.current?.features ?? null);
  const limits: Readable<TenantLimits | null> = derived(store, ($s) => $s.current?.limits ?? null);

  // ============================================================================
  // ACTIONS
  // ============================================================================

  async function loadTenants(): Promise<void> {
    update((s) => ({ ...s, isLoading: true, error: null }));

    try {
      // API call would go here
      // const tenants = await tenantApi.getAvailableTenants();

      // Placeholder
      const mockTenants: Tenant[] = [];

      update((s) => ({
        ...s,
        available: mockTenants,
        isLoading: false,
      }));
    } catch (error) {
      const tenantError: TenantError = {
        code: 'LOAD_FAILED',
        message: error instanceof Error ? error.message : 'Failed to load tenants',
      };
      update((s) => ({ ...s, isLoading: false, error: tenantError }));
      throw error;
    }
  }

  async function switchTenant(tenantId: string): Promise<void> {
    const state = get(store);
    const tenant = state.available.find((t) => t.id === tenantId);

    if (!tenant) {
      throw new Error(`Tenant ${tenantId} not found`);
    }

    update((s) => ({ ...s, isSwitching: true, error: null }));

    try {
      // API call to switch tenant context
      // await tenantApi.switchTenant(tenantId);

      update((s) => ({
        ...s,
        current: tenant,
        isSwitching: false,
      }));

      // Store current tenant in localStorage
      localStorage.setItem('current_tenant_id', tenantId);

      // Notify other stores of tenant change
      // This would typically be done through events or a central bus
    } catch (error) {
      const tenantError: TenantError = {
        code: 'SWITCH_FAILED',
        message: error instanceof Error ? error.message : 'Failed to switch tenant',
      };
      update((s) => ({ ...s, isSwitching: false, error: tenantError }));
      throw error;
    }
  }

  function setCurrentTenant(tenant: Tenant): void {
    update((s) => ({ ...s, current: tenant }));
    localStorage.setItem('current_tenant_id', tenant.id);
  }

  function clearCurrentTenant(): void {
    update((s) => ({ ...s, current: null }));
    localStorage.removeItem('current_tenant_id');
  }

  function setAvailableTenants(tenants: Tenant[]): void {
    update((s) => ({ ...s, available: tenants }));
  }

  async function loadUsage(): Promise<void> {
    const state = get(store);
    if (!state.current) return;

    try {
      // API call would go here
      // const usage = await tenantApi.getUsage(state.current.id);

      const mockUsage: TenantUsage = {
        users: 3,
        storage: 0.5,
        apiRequests: 2500,
        workflows: 25,
        documents: 150,
      };

      update((s) => ({ ...s, usage: mockUsage }));
    } catch (error) {
      console.error('Failed to load tenant usage', error);
    }
  }

  async function updateSettings(updates: Partial<TenantSettings>): Promise<void> {
    const state = get(store);
    if (!state.current) return;

    try {
      // API call would go here
      // await tenantApi.updateSettings(state.current.id, updates);

      update((s) => ({
        ...s,
        current: s.current
          ? {
              ...s.current,
              settings: { ...s.current.settings, ...updates },
            }
          : null,
      }));
    } catch (error) {
      const tenantError: TenantError = {
        code: 'UPDATE_FAILED',
        message: error instanceof Error ? error.message : 'Failed to update settings',
      };
      update((s) => ({ ...s, error: tenantError }));
      throw error;
    }
  }

  function hasFeature(feature: keyof TenantFeatures): boolean {
    const state = get(store);
    return state.current?.features[feature] ?? false;
  }

  function isWithinLimit(limit: keyof TenantLimits, currentValue: number): boolean {
    const state = get(store);
    if (!state.current) return false;
    return currentValue < state.current.limits[limit];
  }

  function getUsagePercentage(limit: keyof TenantLimits): number {
    const state = get(store);
    if (!state.current || !state.usage) return 0;

    const limitValue = state.current.limits[limit];
    const usageKey = limit as keyof TenantUsage;
    const usageValue = state.usage[usageKey] ?? 0;

    return Math.round((usageValue / limitValue) * 100);
  }

  async function initialize(): Promise<void> {
    await loadTenants();

    // Check for stored tenant
    const storedTenantId = localStorage.getItem('current_tenant_id');
    if (storedTenantId) {
      const state = get(store);
      const tenant = state.available.find((t) => t.id === storedTenantId);
      if (tenant) {
        setCurrentTenant(tenant);
        await loadUsage();
      }
    } else if (get(store).available.length === 1) {
      // Auto-select if only one tenant
      const tenant = get(store).available[0];
      if (tenant) {
        setCurrentTenant(tenant);
        await loadUsage();
      }
    }
  }

  function reset(): void {
    localStorage.removeItem('current_tenant_id');
    set(initialState);
  }

  function setError(error: TenantError | null): void {
    update((s) => ({ ...s, error }));
  }

  // ============================================================================
  // RETURN
  // ============================================================================

  return {
    subscribe,
    // Derived stores
    current,
    available,
    isLoading,
    settings,
    features,
    limits,
    // Actions
    loadTenants,
    switchTenant,
    setCurrentTenant,
    clearCurrentTenant,
    setAvailableTenants,
    loadUsage,
    updateSettings,
    hasFeature,
    isWithinLimit,
    getUsagePercentage,
    initialize,
    reset,
    setError,
  };
}

// ============================================================================
// EXPORT
// ============================================================================

export const tenantStore = createTenantStore();

// Export defaults for reference
export { defaultSettings, defaultFeatures, defaultLimits };
