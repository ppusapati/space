/**
 * Tenant Store
 * Handles multi-tenancy state, tenant switching, and tenant-specific settings
 */
import { type Readable } from 'svelte/store';
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
    primaryColor?: string;
    secondaryColor?: string;
    logoUrl?: string;
    faviconUrl?: string;
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
    passwordPolicy: {
        minLength: number;
        requireUppercase: boolean;
        requireLowercase: boolean;
        requireNumbers: boolean;
        requireSpecialChars: boolean;
        maxAge: number;
    };
    sessionTimeout: number;
    mfaRequired: boolean;
    emailFromName?: string;
    emailFromAddress?: string;
    custom?: Record<string, unknown>;
}
export interface TenantFeatures {
    identity: boolean;
    workflow: boolean;
    notifications: boolean;
    audit: boolean;
    data: boolean;
    insights: boolean;
    platform: boolean;
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
    storage: number;
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
declare const defaultSettings: TenantSettings;
declare const defaultFeatures: TenantFeatures;
declare const defaultLimits: TenantLimits;
export declare const tenantStore: {
    subscribe: (this: void, run: import("svelte/store").Subscriber<TenantState>, invalidate?: () => void) => import("svelte/store").Unsubscriber;
    current: Readable<Tenant | null>;
    available: Readable<Tenant[]>;
    isLoading: Readable<boolean>;
    settings: Readable<TenantSettings | null>;
    features: Readable<TenantFeatures | null>;
    limits: Readable<TenantLimits | null>;
    loadTenants: () => Promise<void>;
    switchTenant: (tenantId: string) => Promise<void>;
    setCurrentTenant: (tenant: Tenant) => void;
    clearCurrentTenant: () => void;
    setAvailableTenants: (tenants: Tenant[]) => void;
    loadUsage: () => Promise<void>;
    updateSettings: (updates: Partial<TenantSettings>) => Promise<void>;
    hasFeature: (feature: keyof TenantFeatures) => boolean;
    isWithinLimit: (limit: keyof TenantLimits, currentValue: number) => boolean;
    getUsagePercentage: (limit: keyof TenantLimits) => number;
    initialize: () => Promise<void>;
    reset: () => void;
    setError: (error: TenantError | null) => void;
};
export { defaultSettings, defaultFeatures, defaultLimits };
//# sourceMappingURL=tenant.store.d.ts.map