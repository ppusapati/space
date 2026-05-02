/**
 * API Provider Bridge
 *
 * Concrete implementations of the AuthProvider + SessionProvider contracts
 * declared in @chetana/api/providers, backed by the runtime authStore +
 * sessionStore in this package.
 *
 * Why this file exists: @chetana/api defines the contract surface that
 * its interceptors (auth + tenant) consume; @chetana/stores owns the
 * runtime state. Putting concrete impls here breaks the cyclic dependency
 * (api ↔ stores) by inverting which package depends on which —
 * @chetana/stores depends on @chetana/api (for the contract types),
 * not the other way around.
 *
 * Wire-up: call initApiProviders() ONCE from apps/shell/src/hooks.client.ts
 * BEFORE initializeApi(). Order matters because initializeApi() registers
 * interceptors that immediately call getAuthProvider()/getSessionProvider().
 */

import { get } from 'svelte/store';

import {
  configureProviders,
  type AuthProvider,
  type AuthTokens,
  type SessionProvider,
  type SessionContext,
  type ToastProvider,
} from '@chetana/api';

import { authStore } from './global/auth.store.js';
import { sessionStore } from './global/session.store.js';
import { toastStore } from './global/notifications.store.js';

// ============================================================================
// AuthProvider — bridges authStore to the API contract
// ============================================================================
//
// The auth interceptor (createAuthInterceptor in @chetana/api) calls:
//   auth.isAuthenticated() before each request — true → attach Authorization header
//   auth.getTokens() to read accessToken
//   auth.refreshTokens() if the access token is near expiry
//   auth.logout() when refresh fails (redirects to login)
//
// authStore exposes derived stores; we read them via svelte/store get().
// `setTokens`/`clearTokens` exist on the store; `login`/`logout` do not yet
// (commented out in auth.store.ts as of 2026-04-25). Login flow will call
// AuthService/Login directly and use setTokens to land the result.

const authProviderImpl: AuthProvider = {
  isAuthenticated(): boolean {
    const s = get(authStore);
    return Boolean(s.isAuthenticated && s.tokens?.accessToken);
  },

  getTokens(): AuthTokens | null {
    const s = get(authStore);
    if (!s.tokens?.accessToken) return null;
    // authStore's AuthTokens has tokenType + Date expiresAt; the API contract
    // accepts a Date | undefined expiresAt. Drop tokenType (the interceptor
    // hardcodes "Bearer" — it's not a parameter).
    return {
      accessToken: s.tokens.accessToken,
      refreshToken: s.tokens.refreshToken,
      expiresAt: s.tokens.expiresAt,
    };
  },

  async refreshTokens(): Promise<void> {
    // Calls AuthService/RefreshToken with the stored refresh token, lands the
    // new access+refresh pair into authStore + storage. The auth interceptor
    // calls this when the access token is within the refresh window OR after
    // a 401 response; on rejection it bounces the user to /login.
    //
    // Dynamic import for the same reason auth.store.ts does it: keeps api/proto
    // out of the static graph of stores' index.ts so non-auth flows tree-shake.
    const state = get(authStore);
    const stored = state.tokens?.refreshToken;
    if (!stored) {
      throw new Error('refreshTokens: no refresh token in store — cannot refresh');
    }

    const { create } = await import('@bufbuild/protobuf');
    const { getAuthService } = await import('@chetana/api');
    const authPb = await import(
      '@chetana/proto/gen/core/identity/auth/proto/auth_pb.js'
    );

    const res = await getAuthService().refreshToken(
      create(authPb.RefreshTokenRequestSchema, {
        refreshToken: stored,
      }),
    );

    const accessToken = (res as { accessToken?: string }).accessToken ?? '';
    const refreshToken = (res as { refreshToken?: string }).refreshToken ?? '';
    const expiresAtRaw = (res as { expiresAt?: { seconds?: bigint | number } })
      .expiresAt;
    const expiresAt = expiresAtRaw?.seconds
      ? new Date(Number(expiresAtRaw.seconds) * 1000)
      : new Date(Date.now() + 3600_000);

    if (!accessToken) {
      throw new Error('refreshTokens: backend returned empty accessToken');
    }

    // Fan-out the new tokens into authStore + the same storage slot the login
    // flow uses, so a subsequent reload re-hydrates with the refreshed pair.
    // authStore.AuthTokens has stricter shape than the bridge's contract type
    // (non-optional refreshToken/expiresAt + literal tokenType: 'Bearer'), so
    // we build a fresh object here rather than reuse the bridge AuthTokens.
    authStore.setTokens({
      accessToken,
      refreshToken,
      expiresAt,
      tokenType: 'Bearer',
    });
    try {
      const stash = JSON.stringify({
        accessToken,
        refreshToken,
        expiresAt: expiresAt.toISOString(),
        tokenType: 'Bearer',
      });
      // Match the storage slot the login flow used originally (rememberMe
      // decision was made then; we don't have it here). Prefer localStorage
      // when it currently holds the auth_tokens key, else sessionStorage.
      if (localStorage.getItem('auth_tokens') !== null) {
        localStorage.setItem('auth_tokens', stash);
      } else {
        sessionStorage.setItem('auth_tokens', stash);
      }
    } catch {
      // SSR / restricted browser — in-memory state is still updated.
    }
  },

  logout(): void {
    // Clear both auth + session state. The full backend logout (calls
    // AuthService/RevokeSession) lives in the login flow's logout button.
    // This is the local-state-only path the interceptor uses on refresh
    // failure.
    authStore.clearTokens();
    authStore.clearUser();
    authStore.clearSession();
    sessionStore.clearSession();
  },
};

// ============================================================================
// SessionProvider — bridges sessionStore to the API contract
// ============================================================================
//
// The tenant interceptor (createTenantInterceptor) calls:
//   session.getSessionId() to attach X-Session-ID
//   session.getContext() to attach X-Tenant-ID/X-Company-ID/X-Branch-ID
//
// **Note: the JWT-mode backend ignores these headers** (it reads tenant
// scope from verified JWT claims instead). So this is mostly a dev-mode
// fallback; in JWT mode the headers are redundant noise. They're still
// sent because the interceptor is unconditional and the cost is negligible.

const sessionProviderImpl: SessionProvider = {
  getSessionId(): string | null {
    const s = get(authStore);
    return s.session?.id ?? null;
  },

  getContext(): SessionContext | null {
    // sessionStore.context is a derived Readable<OrganizationContext | null>;
    // we synchronously snapshot it via svelte/store get().
    const ctx = get(sessionStore.context);
    if (!ctx) return null;
    return {
      tenantId: ctx.tenantId,
      companyId: ctx.companyId,
      branchId: ctx.branchId,
      userId: ctx.userId,
    };
  },
};

// ============================================================================
// ToastProvider — bridges toastStore to the API contract
// ============================================================================
//
// The error interceptor uses this to show a user-visible toast when an
// RPC fails. Without it the API package has no way to notify the UI.

const toastProviderImpl: ToastProvider = {
  error(message: string, options?: { title?: string; duration?: number }): void {
    toastStore.error(message, {
      title: options?.title,
      duration: options?.duration,
    });
  },
};

// ============================================================================
// PUBLIC ENTRY POINT
// ============================================================================

/**
 * Wire the concrete provider implementations into the @chetana/api
 * provider registry. Call this ONCE from hooks.client.ts BEFORE
 * initializeApi() — the interceptors registered by initializeApi will
 * immediately call getAuthProvider()/getSessionProvider() during their
 * first request, and getX() throws if no provider is configured.
 *
 * Idempotent: calling configureProviders() with the same impls is safe.
 */
export function initApiProviders(): void {
  configureProviders({
    auth: authProviderImpl,
    session: sessionProviderImpl,
    toast: toastProviderImpl,
  });
}
