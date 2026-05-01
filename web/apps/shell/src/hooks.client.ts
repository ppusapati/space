/**
 * Client-side initialization for the shell app.
 *
 * Order matters:
 *   1. initApiProviders()  — wire concrete AuthProvider/SessionProvider/ToastProvider
 *                            into @samavāya/api's provider registry. The interceptors
 *                            registered by initializeApi() call getAuthProvider()
 *                            and getSessionProvider() on the very first request, so
 *                            providers MUST be configured before initializeApi.
 *   2. hydrateAuth()       — pull persisted tokens out of localStorage/sessionStorage
 *                            and seed authStore. Safe no-op when nothing is stored
 *                            (first launch / logged-out state).
 *   3. initializeApi(...)  — create the ConnectRPC transport + register the auth +
 *                            tenant + error + retry + logging interceptors.
 *
 * Backend URL: VITE_API_URL env var, default http://localhost:9090 (the
 * monolith's default port). Set VITE_API_URL=http://localhost:8088 (or
 * whatever) when pointing at a non-default backend.
 */

import { initializeApi } from '@samavāya/api';
import { authStore, initApiProviders } from '@samavāya/stores';

// 1. Wire stores → api providers. Idempotent.
initApiProviders();

// 2. Hydrate tokens from storage so a refreshed page stays logged in.
function hydrateAuth(): void {
  if (typeof window === 'undefined') return; // SSR safety
  try {
    const raw = localStorage.getItem('auth_tokens') ?? sessionStorage.getItem('auth_tokens');
    if (!raw) return;
    const parsed = JSON.parse(raw) as {
      accessToken?: string;
      refreshToken?: string;
      expiresAt?: string;
      tokenType?: string;
    };
    if (!parsed.accessToken) return;
    const expiresAt = parsed.expiresAt ? new Date(parsed.expiresAt) : new Date(Date.now() + 3600_000);
    if (expiresAt.getTime() < Date.now()) {
      // Stale token — drop it; user will be bounced to /login on first request.
      localStorage.removeItem('auth_tokens');
      sessionStorage.removeItem('auth_tokens');
      return;
    }
    authStore.setTokens({
      accessToken: parsed.accessToken,
      refreshToken: parsed.refreshToken ?? '',
      expiresAt,
      tokenType: 'Bearer',
    });
  } catch (err) {
    console.warn('[hooks.client] hydrateAuth failed (non-fatal):', err);
  }
}
hydrateAuth();

// 3. Initialize the API client. Failures here are fatal for the app —
// without a transport every RPC call crashes, so log loudly and let the
// error boundary catch it.
try {
  initializeApi({
    baseUrl: import.meta.env.VITE_API_URL || 'http://localhost:9090',
    timeout: 30000,
    // 'omit' for Bearer-token auth — see transport.ts default for the
    // CORS wildcard-origin trap rationale. Switch to 'include' only if
    // we move to cookie-based session auth and the backend echoes the
    // exact Origin + Access-Control-Allow-Credentials: true.
    credentials: 'omit',
    debug: import.meta.env.DEV,
    withAuth: true,
    withTenant: true,
    withErrorHandling: true,
    withRetry: true,
    withLogging: import.meta.env.DEV,
  });
} catch (error) {
  console.error('[hooks.client] Failed to initialize API:', error);
}
