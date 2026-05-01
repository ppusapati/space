/**
 * Auth Store
 * Handles authentication state, tokens, and session management
 */

import { writable, derived, get, type Writable, type Readable } from 'svelte/store';

// Static type-only imports — erased at runtime, no circular dependency risk
// import type { LoginResponse, RefreshTokenResponse, ValidateTokenResponse } from '@samavāya/proto/gen/core/identity/auth/proto/auth_pb.js';

// ============================================================================
// TYPES
// ============================================================================

export interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  displayName: string;
  avatar?: string;
  phone?: string;
  locale?: string;
  timezone?: string;
  metadata?: Record<string, unknown>;
}

export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
  expiresAt: Date;
  tokenType: 'Bearer';
}

export interface Permission {
  resource: string;
  action: string;
  conditions?: Record<string, unknown>;
}

export interface Role {
  id: string;
  name: string;
  displayName: string;
  permissions: Permission[];
}

export interface Session {
  id: string;
  userId: string;
  deviceInfo?: {
    userAgent: string;
    platform: string;
    browser: string;
  };
  ipAddress?: string;
  location?: string;
  createdAt: Date;
  lastActiveAt: Date;
  expiresAt: Date;
  isCurrent: boolean;
}

export interface AuthState {
  isAuthenticated: boolean;
  isLoading: boolean;
  isInitialized: boolean;
  user: User | null;
  tokens: AuthTokens | null;
  roles: Role[];
  permissions: Permission[];
  session: Session | null;
  error: AuthError | null;
}

export interface AuthError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

export interface LoginCredentials {
  email: string;
  password: string;
  rememberMe?: boolean;
  mfaCode?: string;
}

export interface RegisterData {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
  phone?: string;
}

export interface AuthStoreActions {
  // Authentication
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => Promise<void>;
  register: (data: RegisterData) => Promise<void>;

  // Token management
  refreshTokens: () => Promise<void>;
  setTokens: (tokens: AuthTokens) => void;
  clearTokens: () => void;

  // User management
  setUser: (user: User) => void;
  updateUser: (updates: Partial<User>) => void;
  clearUser: () => void;

  // Session management
  setSession: (session: Session) => void;
  clearSession: () => void;

  // Permissions
  setRoles: (roles: Role[]) => void;
  setPermissions: (permissions: Permission[]) => void;
  hasPermission: (resource: string, action: string) => boolean;
  hasRole: (roleName: string) => boolean;
  hasAnyRole: (roleNames: string[]) => boolean;
  hasAllRoles: (roleNames: string[]) => boolean;

  // State management
  initialize: () => Promise<void>;
  reset: () => void;
  setError: (error: AuthError | null) => void;
  setLoading: (isLoading: boolean) => void;
}

// ============================================================================
// INITIAL STATE
// ============================================================================

const initialState: AuthState = {
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
  const store = writable<AuthState>(initialState);
  const { subscribe, set, update } = store;

  // Token refresh interval
  let refreshInterval: ReturnType<typeof setInterval> | null = null;

  // ============================================================================
  // DERIVED STORES
  // ============================================================================

  const user: Readable<User | null> = derived(store, ($s) => $s.user);
  const isAuthenticated: Readable<boolean> = derived(store, ($s) => $s.isAuthenticated);
  const isLoading: Readable<boolean> = derived(store, ($s) => $s.isLoading);
  const roles: Readable<Role[]> = derived(store, ($s) => $s.roles);
  const permissions: Readable<Permission[]> = derived(store, ($s) => $s.permissions);

  // ============================================================================
  // ACTIONS
  // ============================================================================

  // async function login(credentials: LoginCredentials): Promise<void> {
  //   update((s) => ({ ...s, isLoading: true, error: null }));

  //   try {
  //     // Dynamic imports to avoid circular deps (stores ← api ← stores)
  //     const { create } = await import('@bufbuild/protobuf');
  //     // const { getAuthService } = await import('@samavāya/api');
  //     const { LoginRequestSchema } = await import(
  //       '@samavāya/proto/gen/core/identity/auth/proto/auth_pb.js'
  //     );

  //     const res: LoginResponse = await getAuthService().login(
  //       create(LoginRequestSchema, {
  //         identifier: { case: 'email', value: credentials.email },
  //         password: credentials.password,
  //         rememberMe: credentials.rememberMe ?? false,
  //       }),
  //     ) as LoginResponse;

  //     if (res.requiresTwoFactor) {
  //       update((s) => ({ ...s, isLoading: false }));
  //       return;
  //     }

  //     const user = res.user;
  //     const expiresAt = res.expiresAt
  //       ? new Date(Number(res.expiresAt.seconds) * 1000)
  //       : new Date(Date.now() + 3600000);

  //     const tokens: AuthTokens = {
  //       accessToken: res.accessToken,
  //       refreshToken: res.refreshToken,
  //       expiresAt,
  //       tokenType: 'Bearer',
  //     };

  //     update((s) => ({
  //       ...s,
  //       isAuthenticated: true,
  //       isLoading: false,
  //       user: {
  //         id: user?.userId ?? '',
  //         email: user?.email ?? '',
  //         firstName: user?.fullname?.split(' ')[0] ?? '',
  //         lastName: user?.fullname?.split(' ').slice(1).join(' ') ?? '',
  //         displayName: user?.fullname ?? '',
  //       },
  //       tokens,
  //       error: null,
  //     }));

  //     // Store tokens
  //     const tokenJson = JSON.stringify(tokens);
  //     if (credentials.rememberMe) {
  //       localStorage.setItem('auth_tokens', tokenJson);
  //     } else {
  //       sessionStorage.setItem('auth_tokens', tokenJson);
  //     }

  //     startTokenRefresh();
  //   } catch (error) {
  //     const authError: AuthError = {
  //       code: 'LOGIN_FAILED',
  //       message: error instanceof Error ? error.message : 'Login failed',
  //     };
  //     update((s) => ({ ...s, isLoading: false, error: authError }));
  //     throw error;
  //   }
  // }

  // async function logout(): Promise<void> {
  //   update((s) => ({ ...s, isLoading: true }));

  //   try {
  //     const state = get(store);
  //     if (state.session?.id) {
  //       const { create } = await import('@bufbuild/protobuf');
  //       const { getAuthService } = await import('@samavāya/api');
  //       const { LogoutRequestSchema } = await import(
  //         '@samavāya/proto/gen/core/identity/auth/proto/auth_pb.js'
  //       );
  //       await getAuthService().logout(
  //         create(LogoutRequestSchema, { sessionId: state.session.id }),
  //       );
  //     }
  //   } catch {
  //     // Still logout locally even if API fails
  //   } finally {
  //     localStorage.removeItem('auth_tokens');
  //     sessionStorage.removeItem('auth_tokens');
  //     stopTokenRefresh();
  //     set(initialState);
  //   }
  // }

  async function register(data: RegisterData): Promise<void> {
    update((s) => ({ ...s, isLoading: true, error: null }));

    try {
      // API call would go here
      // await authApi.register(data);

      update((s) => ({ ...s, isLoading: false }));
    } catch (error) {
      const authError: AuthError = {
        code: 'REGISTER_FAILED',
        message: error instanceof Error ? error.message : 'Registration failed',
      };
      update((s) => ({ ...s, isLoading: false, error: authError }));
      throw error;
    }
  }

  // ============================================================================
  // LOGIN / LOGOUT — wired to AuthService via @samavāya/api
  // ============================================================================
  //
  // These call the live ConnectRPC AuthService at /core.identity.auth.api.v1.AuthService/Login.
  // The longer commented-out variants above were the original draft; this is
  // the simpler version that landed once the apiProviderBridge unblocked the
  // cyclic dep between stores ↔ api.
  //
  // Token persistence: rememberMe → localStorage (survives tab close);
  // !rememberMe → sessionStorage (lost on close). Both keys are 'auth_tokens'.
  // hooks.client.ts hydrates from storage on app boot.

  async function login(credentials: LoginCredentials): Promise<void> {
    update((s) => ({ ...s, isLoading: true, error: null }));
    try {
      // Dynamic imports avoid pulling api/proto into the static import graph
      // of stores' index.ts (keeps tree-shaking healthier in non-login flows).
      const { create } = await import('@bufbuild/protobuf');
      const { getAuthService } = await import('@samavāya/api');
      const authPb = await import('@samavāya/proto/gen/core/identity/auth/proto/auth_pb.js');

      const res = await getAuthService().login(
        create(authPb.LoginRequestSchema, {
          identifier: { case: 'email', value: credentials.email },
          password: credentials.password,
          rememberMe: credentials.rememberMe ?? false,
        }),
      );

      // The backend may return requires_two_factor=true; surfacing that to the
      // UI is FE.5 work — for now treat it as an unsupported error.
      if ((res as { requiresTwoFactor?: boolean }).requiresTwoFactor) {
        update((s) => ({ ...s, isLoading: false }));
        throw new Error('Two-factor auth not yet supported by the UI');
      }

      const accessToken = (res as { accessToken?: string }).accessToken ?? '';
      const refreshToken = (res as { refreshToken?: string }).refreshToken ?? '';
      const expiresAtRaw = (res as { expiresAt?: { seconds?: bigint | number } }).expiresAt;
      const expiresAt = expiresAtRaw?.seconds
        ? new Date(Number(expiresAtRaw.seconds) * 1000)
        : new Date(Date.now() + 3600_000); // 1h fallback

      const tokens: AuthTokens = {
        accessToken,
        refreshToken,
        expiresAt,
        tokenType: 'Bearer',
      };

      // Stash tokens for hydration on next boot.
      const tokenJson = JSON.stringify({
        ...tokens,
        expiresAt: expiresAt.toISOString(),
      });
      try {
        if (credentials.rememberMe) {
          localStorage.setItem('auth_tokens', tokenJson);
        } else {
          sessionStorage.setItem('auth_tokens', tokenJson);
        }
      } catch {
        // SSR or restricted browser — safe to skip; in-memory state still set.
      }

      const userRaw = (res as { user?: { userId?: string; email?: string; fullname?: string } }).user;

      // Capture sessionId from the Login response so the SessionProvider's
      // getSessionId() (consumed by the tenant interceptor for X-Session-ID
      // and by the FE logout flow to call AuthService/RevokeSession) returns
      // the real id instead of null. Backend's LoginResponse.session_id is a
      // ULID; it's the same value the JWT's session_id claim holds and what
      // the auth_sessions table keys by.
      const sessionId = (res as { sessionId?: string }).sessionId ?? '';

      update((s) => ({
        ...s,
        isAuthenticated: true,
        isLoading: false,
        user: userRaw
          ? {
              id: userRaw.userId ?? '',
              email: userRaw.email ?? credentials.email,
              firstName: (userRaw.fullname ?? '').split(' ')[0] ?? '',
              lastName: (userRaw.fullname ?? '').split(' ').slice(1).join(' '),
              displayName: userRaw.fullname ?? credentials.email,
            }
          : null,
        tokens,
        // Minimal Session shape — id is the only field consumers actually
        // read today (SessionProvider.getSessionId, logout flow). Other
        // fields are populated by a follow-up GetSession call when the
        // app needs richer session metadata. createdAt + lastActiveAt
        // default to now since we don't have the server's value yet.
        session: sessionId
          ? {
              id: sessionId,
              userId: userRaw?.userId ?? '',
              createdAt: new Date(),
              lastActiveAt: new Date(),
              expiresAt,
              isCurrent: true,
            }
          : s.session,
        error: null,
      }));
    } catch (error) {
      const authError: AuthError = {
        code: 'LOGIN_FAILED',
        message: error instanceof Error ? error.message : 'Login failed',
      };
      update((s) => ({ ...s, isLoading: false, error: authError }));
      throw error;
    }
  }

  async function logout(): Promise<void> {
    update((s) => ({ ...s, isLoading: true }));
    const state = get(store);
    try {
      if (state.session?.id) {
        const { create } = await import('@bufbuild/protobuf');
        const { getAuthService } = await import('@samavāya/api');
        const authPb = await import('@samavāya/proto/gen/core/identity/auth/proto/auth_pb.js');
        // RevokeSession is the typed name in the backend; if it doesn't
        // exist on the generated client, the catch swallows the error and
        // we still clear local state below.
        const svc = getAuthService() as unknown as Record<string, (req: unknown) => Promise<unknown>>;
        if (typeof svc.revokeSession === 'function') {
          await svc.revokeSession(create(authPb.RevokeSessionRequestSchema, { sessionId: state.session.id }));
        } else if (typeof svc.logout === 'function') {
          await svc.logout(create(authPb.LogoutRequestSchema, { sessionId: state.session.id }));
        }
      }
    } catch {
      // Network or 401 — still clear local state below.
    } finally {
      try {
        localStorage.removeItem('auth_tokens');
        sessionStorage.removeItem('auth_tokens');
      } catch {
        // SSR or restricted — ignore.
      }
      set(initialState);
    }
  }

  // async function refreshTokens(): Promise<void> {
  //   const state = get(store);
  //   if (!state.tokens?.refreshToken) return;

  //   try {
  //     const { create } = await import('@bufbuild/protobuf');
  //     const { getAuthService } = await import('@samavāya/api');
  //     const { RefreshTokenRequestSchema } = await import(
  //       '@samavāya/proto/gen/core/identity/auth/proto/auth_pb.js'
  //     );

  //     const res: RefreshTokenResponse = await getAuthService().refreshToken(
  //       create(RefreshTokenRequestSchema, {
  //         refreshToken: state.tokens.refreshToken,
  //       }),
  //     ) as RefreshTokenResponse;

  //     const expiresAt = res.expiresAt
  //       ? new Date(Number(res.expiresAt.seconds) * 1000)
  //       : new Date(Date.now() + 3600000);

  //     const newTokens: AuthTokens = {
  //       accessToken: res.accessToken,
  //       refreshToken: res.refreshToken,
  //       expiresAt,
  //       tokenType: 'Bearer',
  //     };

  //     update((s) => ({ ...s, tokens: newTokens }));

  //     // Update storage
  //     const tokenJson = JSON.stringify(newTokens);
  //     if (localStorage.getItem('auth_tokens')) {
  //       localStorage.setItem('auth_tokens', tokenJson);
  //     } else {
  //       sessionStorage.setItem('auth_tokens', tokenJson);
  //     }
  //   } catch {
  //     // Token refresh failed - logout
  //     await logout();
  //   }
  // }

  function setTokens(tokens: AuthTokens): void {
    update((s) => ({ ...s, tokens }));
  }

  function clearTokens(): void {
    update((s) => ({ ...s, tokens: null }));
    localStorage.removeItem('auth_tokens');
    sessionStorage.removeItem('auth_tokens');
  }

  function setUser(user: User): void {
    update((s) => ({ ...s, user, isAuthenticated: true }));
  }

  function updateUser(updates: Partial<User>): void {
    update((s) => ({
      ...s,
      user: s.user ? { ...s.user, ...updates } : null,
    }));
  }

  function clearUser(): void {
    update((s) => ({ ...s, user: null, isAuthenticated: false }));
  }

  function setSession(session: Session): void {
    update((s) => ({ ...s, session }));
  }

  function clearSession(): void {
    update((s) => ({ ...s, session: null }));
  }

  function setRoles(roles: Role[]): void {
    // Extract all permissions from roles
    const allPermissions = roles.flatMap((r) => r.permissions);
    update((s) => ({ ...s, roles, permissions: allPermissions }));
  }

  function setPermissions(permissions: Permission[]): void {
    update((s) => ({ ...s, permissions }));
  }

  function hasPermission(resource: string, action: string): boolean {
    const state = get(store);
    return state.permissions.some(
      (p) => p.resource === resource && p.action === action
    );
  }

  function hasRole(roleName: string): boolean {
    const state = get(store);
    return state.roles.some((r) => r.name === roleName);
  }

  function hasAnyRole(roleNames: string[]): boolean {
    const state = get(store);
    return roleNames.some((name) => state.roles.some((r) => r.name === name));
  }

  function hasAllRoles(roleNames: string[]): boolean {
    const state = get(store);
    return roleNames.every((name) => state.roles.some((r) => r.name === name));
  }

  // async function initialize(): Promise<void> {
  //   update((s) => ({ ...s, isLoading: true }));

  //   try {
  //     // Check for stored tokens
  //     const storedTokens =
  //       localStorage.getItem('auth_tokens') ||
  //       sessionStorage.getItem('auth_tokens');

  //     if (storedTokens) {
  //       const tokens: AuthTokens = JSON.parse(storedTokens);
  //       tokens.expiresAt = new Date(tokens.expiresAt);

  //       if (tokens.expiresAt > new Date()) {
  //         update((s) => ({ ...s, tokens }));

  //         // Validate token and get user info from backend
  //         try {
  //           const { create } = await import('@bufbuild/protobuf');
  //           const { getAuthService } = await import('@samavāya/api');
  //           const { ValidateTokenRequestSchema } = await import(
  //             '@samavāya/proto/gen/core/identity/auth/proto/auth_pb.js'
  //           );

  //           const res: ValidateTokenResponse = await getAuthService().validateToken(
  //             create(ValidateTokenRequestSchema, { token: tokens.accessToken }),
  //           ) as ValidateTokenResponse;

  //           if (res.valid && res.user) {
  //             update((s) => ({
  //               ...s,
  //               isAuthenticated: true,
  //               user: {
  //                 id: res.user!.userId,
  //                 email: res.user!.email,
  //                 firstName: res.user!.fullname?.split(' ')[0] ?? '',
  //                 lastName: res.user!.fullname?.split(' ').slice(1).join(' ') ?? '',
  //                 displayName: res.user!.fullname ?? '',
  //               },
  //             }));
  //           }
  //         } catch {
  //           // Validation failed — token may be invalid, still try refresh
  //         }

  //         startTokenRefresh();
  //       } else {
  //         // Tokens expired - try to refresh
  //         update((s) => ({ ...s, tokens }));
  //         await refreshTokens();
  //       }
  //     }

  //     update((s) => ({ ...s, isLoading: false, isInitialized: true }));
  //   } catch {
  //     // Clear invalid tokens
  //     localStorage.removeItem('auth_tokens');
  //     sessionStorage.removeItem('auth_tokens');
  //     update((s) => ({ ...s, isLoading: false, isInitialized: true }));
  //   }
  // }

  function reset(): void {
    stopTokenRefresh();
    set(initialState);
  }

  function setError(error: AuthError | null): void {
    update((s) => ({ ...s, error }));
  }

  function setLoading(isLoading: boolean): void {
    update((s) => ({ ...s, isLoading }));
  }

  // ============================================================================
  // HELPERS
  // ============================================================================

  // function startTokenRefresh(): void {
  //   // Refresh tokens 5 minutes before expiry
  //   const state = get(store);
  //   if (!state.tokens) return;

  //   const expiresIn = state.tokens.expiresAt.getTime() - Date.now();
  //   const refreshIn = Math.max(expiresIn - 5 * 60 * 1000, 60 * 1000); // At least 1 minute

  //   stopTokenRefresh();
  //   refreshInterval = setInterval(refreshTokens, refreshIn);
  // }

  function stopTokenRefresh(): void {
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
    setLoading,
  };
}

// ============================================================================
// EXPORT
// ============================================================================

export const authStore = createAuthStore();
