/**
 * Session Store
 * Manages user session with full organizational context
 * tenant_id, company_id, branch_id, user_id
 */

import { writable, derived, get, type Readable } from 'svelte/store';

// ============================================================================
// TYPES
// ============================================================================

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
  fiscalYearStart: number; // Month 1-12
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

// ============================================================================
// INITIAL STATE
// ============================================================================

const initialState: SessionState = {
  isInitialized: false,
  isLoading: false,
  isSwitching: false,
  session: null,
  error: null,
};

// ============================================================================
// STORE CREATION
// ============================================================================

function createSessionStore() {
  const store = writable<SessionState>(initialState);
  const { subscribe, set, update } = store;

  // Activity tracking interval
  let activityInterval: ReturnType<typeof setInterval> | null = null;

  // ============================================================================
  // DERIVED STORES
  // ============================================================================

  /** Current organization context */
  const context: Readable<OrganizationContext | null> = derived(
    store,
    ($s) => $s.session?.context ?? null
  );

  /** Current tenant */
  const tenant: Readable<SessionTenant | null> = derived(
    store,
    ($s) => $s.session?.tenant ?? null
  );

  /** Current company */
  const company: Readable<SessionCompany | null> = derived(
    store,
    ($s) => $s.session?.company ?? null
  );

  /** Current branch */
  const branch: Readable<SessionBranch | null> = derived(
    store,
    ($s) => $s.session?.branch ?? null
  );

  /** Current fiscal context */
  const fiscalContext: Readable<FiscalContext | null> = derived(
    store,
    ($s) => $s.session?.fiscalContext ?? null
  );

  /** Available tenants */
  const availableTenants: Readable<SessionTenant[]> = derived(
    store,
    ($s) => $s.session?.availableTenants ?? []
  );

  /** Available companies */
  const availableCompanies: Readable<SessionCompany[]> = derived(
    store,
    ($s) => $s.session?.availableCompanies ?? []
  );

  /** Available branches */
  const availableBranches: Readable<SessionBranch[]> = derived(
    store,
    ($s) => $s.session?.availableBranches ?? []
  );

  /** Is session valid */
  const isValid: Readable<boolean> = derived(store, ($s) => {
    if (!$s.session) return false;
    return new Date() < new Date($s.session.expiresAt);
  });

  /** Session context header values for API calls */
  const contextHeaders: Readable<Record<string, string>> = derived(store, ($s): Record<string, string> => {
    if (!$s.session?.context) return {};
    return {
      'X-Tenant-ID': $s.session.context.tenantId,
      'X-Company-ID': $s.session.context.companyId,
      'X-Branch-ID': $s.session.context.branchId,
      'X-User-ID': $s.session.context.userId,
    };
  });

  // ============================================================================
  // ACTIONS
  // ============================================================================

  /**
   * Initialize session from server or storage
   */
  async function initialize(): Promise<void> {
    update((s) => ({ ...s, isLoading: true }));

    try {
      // Check for stored session
      const storedSession = sessionStorage.getItem('session_data');

      if (storedSession) {
        const session: SessionData = JSON.parse(storedSession, dateReviver);

        // Validate session hasn't expired
        if (new Date() < new Date(session.expiresAt)) {
          update((s) => ({
            ...s,
            session,
            isLoading: false,
            isInitialized: true,
          }));

          startActivityTracking();
          return;
        }
      }

      // No valid stored session - will need to fetch from server after auth
      update((s) => ({ ...s, isLoading: false, isInitialized: true }));
    } catch (error) {
      update((s) => ({
        ...s,
        isLoading: false,
        isInitialized: true,
        error: {
          code: 'SESSION_INIT_FAILED',
          message: 'Failed to initialize session',
        },
      }));
    }
  }

  /**
   * Set session data (typically after login)
   */
  function setSession(session: SessionData): void {
    update((s) => ({ ...s, session, error: null }));

    // Store in sessionStorage
    sessionStorage.setItem('session_data', JSON.stringify(session));

    startActivityTracking();
  }

  /**
   * Switch tenant context
   */
  async function switchTenant(tenantId: string): Promise<void> {
    const state = get(store);
    if (!state.session) {
      throw new Error('No active session');
    }

    const targetTenant = state.session.availableTenants.find((t) => t.id === tenantId);
    if (!targetTenant) {
      throw new Error('Tenant not available');
    }

    update((s) => ({ ...s, isSwitching: true }));

    try {
      // API call to switch tenant and get new context
      // const response = await sessionApi.switchTenant(tenantId);

      // Placeholder - would come from API
      // This would return new companies/branches for the selected tenant
      update((s) => ({
        ...s,
        isSwitching: false,
        session: s.session
          ? {
              ...s.session,
              tenant: targetTenant,
              context: {
                ...s.session.context,
                tenantId,
              },
            }
          : null,
      }));

      // Update storage
      const updatedState = get(store);
      if (updatedState.session) {
        sessionStorage.setItem('session_data', JSON.stringify(updatedState.session));
      }
    } catch (error) {
      update((s) => ({
        ...s,
        isSwitching: false,
        error: {
          code: 'SWITCH_TENANT_FAILED',
          message: error instanceof Error ? error.message : 'Failed to switch tenant',
        },
      }));
      throw error;
    }
  }

  /**
   * Switch company context
   */
  async function switchCompany(companyId: string): Promise<void> {
    const state = get(store);
    if (!state.session) {
      throw new Error('No active session');
    }

    const targetCompany = state.session.availableCompanies.find((c) => c.id === companyId);
    if (!targetCompany) {
      throw new Error('Company not available');
    }

    update((s) => ({ ...s, isSwitching: true }));

    try {
      // API call to switch company and get new branches
      // const response = await sessionApi.switchCompany(companyId);

      update((s) => ({
        ...s,
        isSwitching: false,
        session: s.session
          ? {
              ...s.session,
              company: targetCompany,
              context: {
                ...s.session.context,
                companyId,
              },
            }
          : null,
      }));

      // Update storage
      const updatedState = get(store);
      if (updatedState.session) {
        sessionStorage.setItem('session_data', JSON.stringify(updatedState.session));
      }
    } catch (error) {
      update((s) => ({
        ...s,
        isSwitching: false,
        error: {
          code: 'SWITCH_COMPANY_FAILED',
          message: error instanceof Error ? error.message : 'Failed to switch company',
        },
      }));
      throw error;
    }
  }

  /**
   * Switch branch context
   */
  async function switchBranch(branchId: string): Promise<void> {
    const state = get(store);
    if (!state.session) {
      throw new Error('No active session');
    }

    const targetBranch = state.session.availableBranches.find((b) => b.id === branchId);
    if (!targetBranch) {
      throw new Error('Branch not available');
    }

    update((s) => ({ ...s, isSwitching: true }));

    try {
      update((s) => ({
        ...s,
        isSwitching: false,
        session: s.session
          ? {
              ...s.session,
              branch: targetBranch,
              context: {
                ...s.session.context,
                branchId,
              },
            }
          : null,
      }));

      // Update storage
      const updatedState = get(store);
      if (updatedState.session) {
        sessionStorage.setItem('session_data', JSON.stringify(updatedState.session));
      }
    } catch (error) {
      update((s) => ({
        ...s,
        isSwitching: false,
        error: {
          code: 'SWITCH_BRANCH_FAILED',
          message: error instanceof Error ? error.message : 'Failed to switch branch',
        },
      }));
      throw error;
    }
  }

  /**
   * Switch fiscal period
   */
  async function switchFiscalPeriod(periodId: string): Promise<void> {
    const state = get(store);
    if (!state.session) {
      throw new Error('No active session');
    }

    update((s) => ({ ...s, isSwitching: true }));

    try {
      // API call to get fiscal period details
      // const fiscalContext = await sessionApi.getFiscalPeriod(periodId);

      // Placeholder
      update((s) => ({
        ...s,
        isSwitching: false,
      }));
    } catch (error) {
      update((s) => ({
        ...s,
        isSwitching: false,
        error: {
          code: 'SWITCH_PERIOD_FAILED',
          message: error instanceof Error ? error.message : 'Failed to switch fiscal period',
        },
      }));
      throw error;
    }
  }

  /**
   * Update user preferences
   */
  function updatePreferences(preferences: Partial<SessionPreferences>): void {
    update((s) => ({
      ...s,
      session: s.session
        ? {
            ...s.session,
            preferences: { ...s.session.preferences, ...preferences },
          }
        : null,
    }));

    // Update storage
    const state = get(store);
    if (state.session) {
      sessionStorage.setItem('session_data', JSON.stringify(state.session));
    }
  }

  /**
   * Update last activity timestamp
   */
  function updateActivity(): void {
    update((s) => ({
      ...s,
      session: s.session
        ? {
            ...s.session,
            lastActivityAt: new Date(),
          }
        : null,
    }));
  }

  /**
   * Clear session (on logout)
   */
  function clearSession(): void {
    stopActivityTracking();
    sessionStorage.removeItem('session_data');
    set(initialState);
  }

  /**
   * Get current context for API headers
   */
  function getContextHeaders(): Record<string, string> {
    const state = get(store);
    if (!state.session?.context) return {};

    return {
      'X-Tenant-ID': state.session.context.tenantId,
      'X-Company-ID': state.session.context.companyId,
      'X-Branch-ID': state.session.context.branchId,
      'X-User-ID': state.session.context.userId,
    };
  }

  // ============================================================================
  // HELPERS
  // ============================================================================

  function startActivityTracking(): void {
    stopActivityTracking();

    // Update activity every 5 minutes
    activityInterval = setInterval(() => {
      updateActivity();
    }, 5 * 60 * 1000);
  }

  function stopActivityTracking(): void {
    if (activityInterval) {
      clearInterval(activityInterval);
      activityInterval = null;
    }
  }

  /**
   * JSON date reviver for parsing stored session
   */
  function dateReviver(key: string, value: unknown): unknown {
    if (typeof value === 'string') {
      const dateKeys = ['createdAt', 'lastActivityAt', 'expiresAt', 'periodStart', 'periodEnd'];
      if (dateKeys.includes(key)) {
        return new Date(value);
      }
    }
    return value;
  }

  // ============================================================================
  // RETURN
  // ============================================================================

  return {
    subscribe,
    // Derived stores
    context,
    tenant,
    company,
    branch,
    fiscalContext,
    availableTenants,
    availableCompanies,
    availableBranches,
    isValid,
    contextHeaders,
    // Actions
    initialize,
    setSession,
    switchTenant,
    switchCompany,
    switchBranch,
    switchFiscalPeriod,
    updatePreferences,
    updateActivity,
    clearSession,
    getContextHeaders,
  };
}

// ============================================================================
// EXPORT
// ============================================================================

export const sessionStore = createSessionStore();
