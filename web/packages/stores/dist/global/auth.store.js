/**
 * Auth Store
 * Handles authentication state, tokens, and session management
 */
import { writable, derived, get } from 'svelte/store';
// ============================================================================
// INITIAL STATE
// ============================================================================
const initialState = {
    isAuthenticated: false,
    isLoading: false,
    isInitialized: false,
    user: null,
    tokens: null,
    roles: [],
    permissions: [],
    session: null,
    error: null,
};
// ============================================================================
// STORE CREATION
// ============================================================================
function createAuthStore() {
    const store = writable(initialState);
    const { subscribe, set, update } = store;
    // Token refresh interval
    let refreshInterval = null;
    // ============================================================================
    // DERIVED STORES
    // ============================================================================
    const user = derived(store, ($s) => $s.user);
    const isAuthenticated = derived(store, ($s) => $s.isAuthenticated);
    const isLoading = derived(store, ($s) => $s.isLoading);
    const roles = derived(store, ($s) => $s.roles);
    const permissions = derived(store, ($s) => $s.permissions);
    // ============================================================================
    // ACTIONS
    // ============================================================================
    async function login(credentials) {
        update((s) => ({ ...s, isLoading: true, error: null }));
        try {
            // API call would go here
            // const response = await authApi.login(credentials);
            // Placeholder - replace with actual API call
            const mockUser = {
                id: '1',
                email: credentials.email,
                firstName: 'John',
                lastName: 'Doe',
                displayName: 'John Doe',
            };
            const mockTokens = {
                accessToken: 'mock-access-token',
                refreshToken: 'mock-refresh-token',
                expiresAt: new Date(Date.now() + 3600000), // 1 hour
                tokenType: 'Bearer',
            };
            update((s) => ({
                ...s,
                isAuthenticated: true,
                isLoading: false,
                user: mockUser,
                tokens: mockTokens,
                error: null,
            }));
            // Store tokens in localStorage if rememberMe
            if (credentials.rememberMe) {
                localStorage.setItem('auth_tokens', JSON.stringify(mockTokens));
            }
            else {
                sessionStorage.setItem('auth_tokens', JSON.stringify(mockTokens));
            }
            // Start token refresh interval
            startTokenRefresh();
        }
        catch (error) {
            const authError = {
                code: 'LOGIN_FAILED',
                message: error instanceof Error ? error.message : 'Login failed',
            };
            update((s) => ({ ...s, isLoading: false, error: authError }));
            throw error;
        }
    }
    async function logout() {
        update((s) => ({ ...s, isLoading: true }));
        try {
            // API call would go here
            // await authApi.logout();
            // Clear storage
            localStorage.removeItem('auth_tokens');
            sessionStorage.removeItem('auth_tokens');
            // Stop token refresh
            stopTokenRefresh();
            // Reset state
            set(initialState);
        }
        catch (error) {
            // Still logout locally even if API fails
            localStorage.removeItem('auth_tokens');
            sessionStorage.removeItem('auth_tokens');
            stopTokenRefresh();
            set(initialState);
        }
    }
    async function register(data) {
        update((s) => ({ ...s, isLoading: true, error: null }));
        try {
            // API call would go here
            // await authApi.register(data);
            update((s) => ({ ...s, isLoading: false }));
        }
        catch (error) {
            const authError = {
                code: 'REGISTER_FAILED',
                message: error instanceof Error ? error.message : 'Registration failed',
            };
            update((s) => ({ ...s, isLoading: false, error: authError }));
            throw error;
        }
    }
    async function refreshTokens() {
        const state = get(store);
        if (!state.tokens?.refreshToken)
            return;
        try {
            // API call would go here
            // const response = await authApi.refreshToken(state.tokens.refreshToken);
            // Placeholder - replace with actual API call
            const newTokens = {
                accessToken: 'new-access-token',
                refreshToken: 'new-refresh-token',
                expiresAt: new Date(Date.now() + 3600000),
                tokenType: 'Bearer',
            };
            update((s) => ({ ...s, tokens: newTokens }));
            // Update storage
            const stored = localStorage.getItem('auth_tokens');
            if (stored) {
                localStorage.setItem('auth_tokens', JSON.stringify(newTokens));
            }
            else {
                sessionStorage.setItem('auth_tokens', JSON.stringify(newTokens));
            }
        }
        catch (error) {
            // Token refresh failed - logout
            await logout();
        }
    }
    function setTokens(tokens) {
        update((s) => ({ ...s, tokens }));
    }
    function clearTokens() {
        update((s) => ({ ...s, tokens: null }));
        localStorage.removeItem('auth_tokens');
        sessionStorage.removeItem('auth_tokens');
    }
    function setUser(user) {
        update((s) => ({ ...s, user, isAuthenticated: true }));
    }
    function updateUser(updates) {
        update((s) => ({
            ...s,
            user: s.user ? { ...s.user, ...updates } : null,
        }));
    }
    function clearUser() {
        update((s) => ({ ...s, user: null, isAuthenticated: false }));
    }
    function setSession(session) {
        update((s) => ({ ...s, session }));
    }
    function clearSession() {
        update((s) => ({ ...s, session: null }));
    }
    function setRoles(roles) {
        // Extract all permissions from roles
        const allPermissions = roles.flatMap((r) => r.permissions);
        update((s) => ({ ...s, roles, permissions: allPermissions }));
    }
    function setPermissions(permissions) {
        update((s) => ({ ...s, permissions }));
    }
    function hasPermission(resource, action) {
        const state = get(store);
        return state.permissions.some((p) => p.resource === resource && p.action === action);
    }
    function hasRole(roleName) {
        const state = get(store);
        return state.roles.some((r) => r.name === roleName);
    }
    function hasAnyRole(roleNames) {
        const state = get(store);
        return roleNames.some((name) => state.roles.some((r) => r.name === name));
    }
    function hasAllRoles(roleNames) {
        const state = get(store);
        return roleNames.every((name) => state.roles.some((r) => r.name === name));
    }
    async function initialize() {
        update((s) => ({ ...s, isLoading: true }));
        try {
            // Check for stored tokens
            const storedTokens = localStorage.getItem('auth_tokens') ||
                sessionStorage.getItem('auth_tokens');
            if (storedTokens) {
                const tokens = JSON.parse(storedTokens);
                tokens.expiresAt = new Date(tokens.expiresAt);
                // Check if tokens are expired
                if (tokens.expiresAt > new Date()) {
                    update((s) => ({ ...s, tokens }));
                    // Fetch user data
                    // const user = await authApi.getCurrentUser();
                    // update(s => ({ ...s, user, isAuthenticated: true }));
                    startTokenRefresh();
                }
                else {
                    // Tokens expired - try to refresh
                    update((s) => ({ ...s, tokens }));
                    await refreshTokens();
                }
            }
            update((s) => ({ ...s, isLoading: false, isInitialized: true }));
        }
        catch (error) {
            // Clear invalid tokens
            localStorage.removeItem('auth_tokens');
            sessionStorage.removeItem('auth_tokens');
            update((s) => ({ ...s, isLoading: false, isInitialized: true }));
        }
    }
    function reset() {
        stopTokenRefresh();
        set(initialState);
    }
    function setError(error) {
        update((s) => ({ ...s, error }));
    }
    function setLoading(isLoading) {
        update((s) => ({ ...s, isLoading }));
    }
    // ============================================================================
    // HELPERS
    // ============================================================================
    function startTokenRefresh() {
        // Refresh tokens 5 minutes before expiry
        const state = get(store);
        if (!state.tokens)
            return;
        const expiresIn = state.tokens.expiresAt.getTime() - Date.now();
        const refreshIn = Math.max(expiresIn - 5 * 60 * 1000, 60 * 1000); // At least 1 minute
        stopTokenRefresh();
        refreshInterval = setInterval(refreshTokens, refreshIn);
    }
    function stopTokenRefresh() {
        if (refreshInterval) {
            clearInterval(refreshInterval);
            refreshInterval = null;
        }
    }
    // ============================================================================
    // RETURN
    // ============================================================================
    return {
        subscribe,
        // Derived stores
        user,
        isAuthenticated,
        isLoading,
        roles,
        permissions,
        // Actions
        login,
        logout,
        register,
        refreshTokens,
        setTokens,
        clearTokens,
        setUser,
        updateUser,
        clearUser,
        setSession,
        clearSession,
        setRoles,
        setPermissions,
        hasPermission,
        hasRole,
        hasAnyRole,
        hasAllRoles,
        initialize,
        reset,
        setError,
        setLoading,
    };
}
// ============================================================================
// EXPORT
// ============================================================================
export const authStore = createAuthStore();
