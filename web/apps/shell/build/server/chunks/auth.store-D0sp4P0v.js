import { d as derived, g as get, w as writable } from './index-CBcFMcIv.js';

const initialState = {
  isAuthenticated: false,
  isLoading: false,
  isInitialized: false,
  user: null,
  tokens: null,
  roles: [],
  permissions: [],
  session: null,
  error: null
};
function createAuthStore() {
  const store = writable(initialState);
  const { subscribe, set, update } = store;
  const user = derived(store, ($s) => $s.user);
  const isAuthenticated = derived(store, ($s) => $s.isAuthenticated);
  const isLoading = derived(store, ($s) => $s.isLoading);
  const roles = derived(store, ($s) => $s.roles);
  const permissions = derived(store, ($s) => $s.permissions);
  async function register(data) {
    update((s) => ({ ...s, isLoading: true, error: null }));
    try {
      update((s) => ({ ...s, isLoading: false }));
    } catch (error) {
      const authError = {
        code: "REGISTER_FAILED",
        message: error instanceof Error ? error.message : "Registration failed"
      };
      update((s) => ({ ...s, isLoading: false, error: authError }));
      throw error;
    }
  }
  function setTokens(tokens) {
    update((s) => ({ ...s, tokens }));
  }
  function clearTokens() {
    update((s) => ({ ...s, tokens: null }));
    localStorage.removeItem("auth_tokens");
    sessionStorage.removeItem("auth_tokens");
  }
  function setUser(user2) {
    update((s) => ({ ...s, user: user2, isAuthenticated: true }));
  }
  function updateUser(updates) {
    update((s) => ({
      ...s,
      user: s.user ? { ...s.user, ...updates } : null
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
  function setRoles(roles2) {
    const allPermissions = roles2.flatMap((r) => r.permissions);
    update((s) => ({ ...s, roles: roles2, permissions: allPermissions }));
  }
  function setPermissions(permissions2) {
    update((s) => ({ ...s, permissions: permissions2 }));
  }
  function hasPermission(resource, action) {
    const state = get(store);
    return state.permissions.some(
      (p) => p.resource === resource && p.action === action
    );
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
  function reset() {
    set(initialState);
  }
  function setError(error) {
    update((s) => ({ ...s, error }));
  }
  function setLoading(isLoading2) {
    update((s) => ({ ...s, isLoading: isLoading2 }));
  }
  return {
    subscribe,
    // Derived stores
    user,
    isAuthenticated,
    isLoading,
    roles,
    permissions,
    // Actions
    // login,
    // logout,
    register,
    // refreshTokens,
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
    // initialize,
    reset,
    setError,
    setLoading
  };
}
const authStore = createAuthStore();

export { authStore as a };
//# sourceMappingURL=auth.store-D0sp4P0v.js.map
