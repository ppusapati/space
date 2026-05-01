/**
 * Authentication Interceptor
 * Handles token injection and refresh
 * @packageDocumentation
 */

import type { Interceptor } from '@connectrpc/connect';
import { getAuthProvider } from '../providers.js';
import { createApiError } from '../client/client.js';

// ============================================================================
// AUTH INTERCEPTOR
// ============================================================================

/**
 * Paths that don't require authentication.
 *
 * Mirrors the backend's jwtAuthMiddleware skip-list (app/cmd/main.go) and
 * tenant.interceptor's CONTEXT_EXEMPT_PATHS. Matched via
 * `path.includes(...)` so each entry must be a unique substring of the
 * full pathname. ConnectRPC paths look like
 * `/core.identity.auth.api.v1.AuthService/Login` — the canonical exempt
 * substring is the fully-qualified RPC procedure path. The legacy
 * `/auth/login`-style entries never matched because the dotted package
 * qualifier sits between the leading `/` and the word `auth`.
 *
 * The first 5 entries are the pre-login flows (Login + 2FA + RefreshToken
 * + password recovery). RegisterTenantUser is included for fresh signup.
 * .well-known + health/version stay as legacy REST entries.
 */
const PUBLIC_PATHS = [
  '/core.identity.auth.api.v1.AuthService/Login',
  '/core.identity.auth.api.v1.AuthService/LoginWithOTP',
  '/core.identity.auth.api.v1.AuthService/RefreshToken',
  '/core.identity.auth.api.v1.AuthService/ForgotPassword',
  '/core.identity.auth.api.v1.AuthService/ResetPassword',
  '/.well-known/jwks.json',
  '/health',
  '/ready',
  '/version',
];

/** Token refresh threshold in ms (refresh if expires in less than 5 minutes) */
const REFRESH_THRESHOLD = 5 * 60 * 1000;

/** Whether a refresh is in progress */
let isRefreshing = false;

/** Queue of requests waiting for token refresh */
let refreshQueue: Array<{
  resolve: (token: string) => void;
  reject: (error: Error) => void;
}> = [];

/**
 * Creates an authentication interceptor
 */
export function createAuthInterceptor(): Interceptor {
  return (next) => async (req) => {
    // Skip auth for public paths
    if (isPublicPath(req.url)) {
      return next(req);
    }

    const auth = getAuthProvider();

    // Check if authenticated
    if (!auth.isAuthenticated() || !auth.getTokens()) {
      throw createApiError('UNAUTHENTICATED', 'User is not authenticated');
    }

    // Check if token needs refresh
    if (shouldRefreshToken(auth.getTokens()?.expiresAt)) {
      try {
        await refreshTokenIfNeeded();
      } catch {
        // Token refresh failed, redirect to login
        auth.logout();
        throw createApiError('UNAUTHENTICATED', 'Session expired');
      }
    }

    // Get fresh token after potential refresh
    const token = getAuthProvider().getTokens()?.accessToken;

    if (!token) {
      throw createApiError('UNAUTHENTICATED', 'No access token available');
    }

    // Add authorization header
    req.header.set('Authorization', `Bearer ${token}`);

    try {
      return await next(req);
    } catch (error) {
      // Handle 401 errors
      if (isUnauthorizedError(error)) {
        // Try to refresh token and retry
        try {
          await refreshTokenIfNeeded(true);
          const freshToken = getAuthProvider().getTokens()?.accessToken;
          req.header.set('Authorization', `Bearer ${freshToken}`);
          return await next(req);
        } catch {
          getAuthProvider().logout();
          throw createApiError('UNAUTHENTICATED', 'Session expired');
        }
      }
      throw error;
    }
  };
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

/**
 * Checks if a path is public (doesn't require auth)
 */
function isPublicPath(url: string): boolean {
  const path = new URL(url).pathname;
  return PUBLIC_PATHS.some((publicPath) => path.includes(publicPath));
}

/**
 * Checks if token should be refreshed
 */
function shouldRefreshToken(expiresAt: Date | undefined): boolean {
  if (!expiresAt) return true;

  const now = Date.now();
  const expiryTime = new Date(expiresAt).getTime();

  return expiryTime - now < REFRESH_THRESHOLD;
}

/**
 * Refreshes the token if needed
 */
async function refreshTokenIfNeeded(force = false): Promise<void> {
  const auth = getAuthProvider();

  if (!force && !shouldRefreshToken(auth.getTokens()?.expiresAt)) {
    return;
  }

  if (isRefreshing) {
    // Wait for ongoing refresh
    return new Promise((resolve, reject) => {
      refreshQueue.push({ resolve: () => resolve(), reject });
    });
  }

  isRefreshing = true;

  try {
    await auth.refreshTokens();

    // Resolve all queued requests
    refreshQueue.forEach(({ resolve }) => resolve(''));
    refreshQueue = [];
  } catch (error) {
    // Reject all queued requests
    const err = error instanceof Error ? error : new Error('Token refresh failed');
    refreshQueue.forEach(({ reject }) => reject(err));
    refreshQueue = [];
    throw error;
  } finally {
    isRefreshing = false;
  }
}

/**
 * Checks if error is an unauthorized error
 */
function isUnauthorizedError(error: unknown): boolean {
  if (typeof error === 'object' && error !== null) {
    const err = error as { code?: string };
    return err.code === 'UNAUTHENTICATED' || err.code === 'permission_denied';
  }
  return false;
}
