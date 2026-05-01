/**
 * Tenant Store
 * Handles multi-tenancy state, tenant switching, and tenant-specific settings
 */
import { writable, derived, get } from 'svelte/store';
// ============================================================================
// INITIAL STATE
// ============================================================================
const defaultSettings = {
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
const defaultFeatures = {
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
const defaultLimits = {
    users: 5,
    storage: 1,
    apiRequestsPerMonth: 10000,
    workflowsPerMonth: 100,
    documentsPerMonth: 500,
    customFieldsPerEntity: 5,
    webhooksCount: 3,
    integrations: 2,
};
const initialState = {
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
    const store = writable(initialState);
    const { subscribe, set, update } = store;
    // ============================================================================
    // DERIVED STORES
    // ============================================================================
    const current = derived(store, ($s) => $s.current);
    const available = derived(store, ($s) => $s.available);
    const isLoading = derived(store, ($s) => $s.isLoading);
    const settings = derived(store, ($s) => $s.current?.settings ?? null);
    const features = derived(store, ($s) => $s.current?.features ?? null);
    const limits = derived(store, ($s) => $s.current?.limits ?? null);
    // ============================================================================
    // ACTIONS
    // ============================================================================
    async function loadTenants() {
        update((s) => ({ ...s, isLoading: true, error: null }));
        try {
            // API call would go here
            // const tenants = await tenantApi.getAvailableTenants();
            // Placeholder
            const mockTenants = [];
            update((s) => ({
                ...s,
                available: mockTenants,
                isLoading: false,
            }));
        }
        catch (error) {
            const tenantError = {
                code: 'LOAD_FAILED',
                message: error instanceof Error ? error.message : 'Failed to load tenants',
            };
            update((s) => ({ ...s, isLoading: false, error: tenantError }));
            throw error;
        }
    }
    async function switchTenant(tenantId) {
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
        }
        catch (error) {
            const tenantError = {
                code: 'SWITCH_FAILED',
                message: error instanceof Error ? error.message : 'Failed to switch tenant',
            };
            update((s) => ({ ...s, isSwitching: false, error: tenantError }));
            throw error;
        }
    }
    function setCurrentTenant(tenant) {
        update((s) => ({ ...s, current: tenant }));
        localStorage.setItem('current_tenant_id', tenant.id);
    }
    function clearCurrentTenant() {
        update((s) => ({ ...s, current: null }));
        localStorage.removeItem('current_tenant_id');
    }
    function setAvailableTenants(tenants) {
        update((s) => ({ ...s, available: tenants }));
    }
    async function loadUsage() {
        const state = get(store);
        if (!state.current)
            return;
        try {
            // API call would go here
            // const usage = await tenantApi.getUsage(state.current.id);
            const mockUsage = {
                users: 3,
                storage: 0.5,
                apiRequests: 2500,
                workflows: 25,
                documents: 150,
            };
            update((s) => ({ ...s, usage: mockUsage }));
        }
        catch (error) {
            console.error('Failed to load tenant usage', error);
        }
    }
    async function updateSettings(updates) {
        const state = get(store);
        if (!state.current)
            return;
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
        }
        catch (error) {
            const tenantError = {
                code: 'UPDATE_FAILED',
                message: error instanceof Error ? error.message : 'Failed to update settings',
            };
            update((s) => ({ ...s, error: tenantError }));
            throw error;
        }
    }
    function hasFeature(feature) {
        const state = get(store);
        return state.current?.features[feature] ?? false;
    }
    function isWithinLimit(limit, currentValue) {
        const state = get(store);
        if (!state.current)
            return false;
        return currentValue < state.current.limits[limit];
    }
    function getUsagePercentage(limit) {
        const state = get(store);
        if (!state.current || !state.usage)
            return 0;
        const limitValue = state.current.limits[limit];
        const usageKey = limit;
        const usageValue = state.usage[usageKey] ?? 0;
        return Math.round((usageValue / limitValue) * 100);
    }
    async function initialize() {
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
        }
        else if (get(store).available.length === 1) {
            // Auto-select if only one tenant
            const tenant = get(store).available[0];
            if (tenant) {
                setCurrentTenant(tenant);
                await loadUsage();
            }
        }
    }
    function reset() {
        localStorage.removeItem('current_tenant_id');
        set(initialState);
    }
    function setError(error) {
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
