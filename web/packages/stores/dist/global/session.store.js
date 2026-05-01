/**
 * Session Store
 * Manages user session with full organizational context
 * tenant_id, company_id, branch_id, user_id
 */
import { writable, derived, get } from 'svelte/store';
// ============================================================================
// INITIAL STATE
// ============================================================================
const initialState = {
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
    const store = writable(initialState);
    const { subscribe, set, update } = store;
    // Activity tracking interval
    let activityInterval = null;
    // ============================================================================
    // DERIVED STORES
    // ============================================================================
    /** Current organization context */
    const context = derived(store, ($s) => $s.session?.context ?? null);
    /** Current tenant */
    const tenant = derived(store, ($s) => $s.session?.tenant ?? null);
    /** Current company */
    const company = derived(store, ($s) => $s.session?.company ?? null);
    /** Current branch */
    const branch = derived(store, ($s) => $s.session?.branch ?? null);
    /** Current fiscal context */
    const fiscalContext = derived(store, ($s) => $s.session?.fiscalContext ?? null);
    /** Available tenants */
    const availableTenants = derived(store, ($s) => $s.session?.availableTenants ?? []);
    /** Available companies */
    const availableCompanies = derived(store, ($s) => $s.session?.availableCompanies ?? []);
    /** Available branches */
    const availableBranches = derived(store, ($s) => $s.session?.availableBranches ?? []);
    /** Is session valid */
    const isValid = derived(store, ($s) => {
        if (!$s.session)
            return false;
        return new Date() < new Date($s.session.expiresAt);
    });
    /** Session context header values for API calls */
    const contextHeaders = derived(store, ($s) => {
        if (!$s.session?.context)
            return {};
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
    async function initialize() {
        update((s) => ({ ...s, isLoading: true }));
        try {
            // Check for stored session
            const storedSession = sessionStorage.getItem('session_data');
            if (storedSession) {
                const session = JSON.parse(storedSession, dateReviver);
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
        }
        catch (error) {
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
    function setSession(session) {
        update((s) => ({ ...s, session, error: null }));
        // Store in sessionStorage
        sessionStorage.setItem('session_data', JSON.stringify(session));
        startActivityTracking();
    }
    /**
     * Switch tenant context
     */
    async function switchTenant(tenantId) {
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
        }
        catch (error) {
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
    async function switchCompany(companyId) {
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
        }
        catch (error) {
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
    async function switchBranch(branchId) {
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
        }
        catch (error) {
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
    async function switchFiscalPeriod(periodId) {
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
        }
        catch (error) {
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
    function updatePreferences(preferences) {
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
    function updateActivity() {
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
    function clearSession() {
        stopActivityTracking();
        sessionStorage.removeItem('session_data');
        set(initialState);
    }
    /**
     * Get current context for API headers
     */
    function getContextHeaders() {
        const state = get(store);
        if (!state.session?.context)
            return {};
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
    function startActivityTracking() {
        stopActivityTracking();
        // Update activity every 5 minutes
        activityInterval = setInterval(() => {
            updateActivity();
        }, 5 * 60 * 1000);
    }
    function stopActivityTracking() {
        if (activityInterval) {
            clearInterval(activityInterval);
            activityInterval = null;
        }
    }
    /**
     * JSON date reviver for parsing stored session
     */
    function dateReviver(key, value) {
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
