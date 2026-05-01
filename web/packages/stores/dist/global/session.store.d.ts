/**
 * Session Store
 * Manages user session with full organizational context
 * tenant_id, company_id, branch_id, user_id
 */
import { type Readable } from 'svelte/store';
/** Organization context for multi-tenant hierarchy */
export interface OrganizationContext {
    /** Tenant ID - top level organization (SaaS customer) */
    tenantId: string;
    /** Company ID - legal entity within tenant */
    companyId: string;
    /** Branch ID - physical or logical location within company */
    branchId: string;
    /** User ID - authenticated user */
    userId: string;
}
/** Tenant - SaaS customer organization */
export interface SessionTenant {
    id: string;
    code: string;
    name: string;
    domain?: string;
    logo?: string;
    isActive: boolean;
}
/** Company - legal entity within tenant */
export interface SessionCompany {
    id: string;
    tenantId: string;
    code: string;
    name: string;
    legalName?: string;
    taxId?: string;
    logo?: string;
    currency: string;
    fiscalYearStart: number;
    isActive: boolean;
}
/** Branch - location within company */
export interface SessionBranch {
    id: string;
    companyId: string;
    code: string;
    name: string;
    type: BranchType;
    address?: BranchAddress;
    phone?: string;
    email?: string;
    timezone: string;
    isDefault: boolean;
    isActive: boolean;
}
export type BranchType = 'headquarters' | 'branch' | 'warehouse' | 'store' | 'factory' | 'office';
export interface BranchAddress {
    line1: string;
    line2?: string;
    city: string;
    state: string;
    postalCode: string;
    country: string;
}
/** Fiscal period context */
export interface FiscalContext {
    fiscalYearId: string;
    fiscalYearName: string;
    fiscalPeriodId: string;
    fiscalPeriodName: string;
    periodStart: Date;
    periodEnd: Date;
    isClosed: boolean;
}
/** User preferences within session */
export interface SessionPreferences {
    language: string;
    timezone: string;
    dateFormat: string;
    timeFormat: string;
    numberFormat: string;
    currency: string;
    theme: 'light' | 'dark' | 'system';
    density: 'compact' | 'default' | 'comfortable';
}
/** Full session data */
export interface SessionData {
    /** Session ID */
    id: string;
    /** Organization context */
    context: OrganizationContext;
    /** Current tenant */
    tenant: SessionTenant;
    /** Available tenants for user */
    availableTenants: SessionTenant[];
    /** Current company */
    company: SessionCompany;
    /** Available companies for user in current tenant */
    availableCompanies: SessionCompany[];
    /** Current branch */
    branch: SessionBranch;
    /** Available branches for user in current company */
    availableBranches: SessionBranch[];
    /** Fiscal context */
    fiscalContext: FiscalContext | null;
    /** User preferences */
    preferences: SessionPreferences;
    /** Session timestamps */
    createdAt: Date;
    lastActivityAt: Date;
    expiresAt: Date;
    /** Device/client info */
    deviceInfo: DeviceInfo;
}
export interface DeviceInfo {
    userAgent: string;
    platform: string;
    browser: string;
    ipAddress?: string;
    location?: string;
}
/** Session state */
export interface SessionState {
    isInitialized: boolean;
    isLoading: boolean;
    isSwitching: boolean;
    session: SessionData | null;
    error: SessionError | null;
}
export interface SessionError {
    code: string;
    message: string;
    details?: Record<string, unknown>;
}
export declare const sessionStore: {
    subscribe: (this: void, run: import("svelte/store").Subscriber<SessionState>, invalidate?: () => void) => import("svelte/store").Unsubscriber;
    context: Readable<OrganizationContext | null>;
    tenant: Readable<SessionTenant | null>;
    company: Readable<SessionCompany | null>;
    branch: Readable<SessionBranch | null>;
    fiscalContext: Readable<FiscalContext | null>;
    availableTenants: Readable<SessionTenant[]>;
    availableCompanies: Readable<SessionCompany[]>;
    availableBranches: Readable<SessionBranch[]>;
    isValid: Readable<boolean>;
    contextHeaders: Readable<Record<string, string>>;
    initialize: () => Promise<void>;
    setSession: (session: SessionData) => void;
    switchTenant: (tenantId: string) => Promise<void>;
    switchCompany: (companyId: string) => Promise<void>;
    switchBranch: (branchId: string) => Promise<void>;
    switchFiscalPeriod: (periodId: string) => Promise<void>;
    updatePreferences: (preferences: Partial<SessionPreferences>) => void;
    updateActivity: () => void;
    clearSession: () => void;
    getContextHeaders: () => Record<string, string>;
};
//# sourceMappingURL=session.store.d.ts.map