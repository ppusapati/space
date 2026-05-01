/**
 * Authentication Interceptor
 * Handles token injection and refresh
 * @packageDocumentation
 */
import { authStore } from '@samavāya/stores';
import { get } from 'svelte/store';
import { createApiError } from '../client/client.js';
// ============================================================================
// AUTH INTERCEPTOR
// ============================================================================
/** Paths that don't require authentication */
const PUBLIC_PATHS = [
    '/auth/login',
    '/auth/register',
    '/auth/forgot-password',
    '/auth/reset-password',
    '/auth/verify-email',
    '/health',
    '/version',
];
/** Token refresh threshold in ms (refresh if expires in less than 5 minutes) */
const REFRESH_THRESHOLD = 5 * 60 * 1000;
/** Whether a refresh is in progress */
let isRefreshing = false;
/** Queue of requests waiting for token refresh */
let refreshQueue = [];
/**
 * Creates an authentication interceptor
 */
export function createAuthInterceptor() {
    return (next) => async (req) => {
        // Skip auth for public paths
        if (isPublicPath(req.url)) {
            return next(req);
        }
        const store = get(authStore);
        // Check if authenticated
        if (!store.isAuthenticated || !store.tokens) {
            throw createApiError('UNAUTHENTICATED', 'User is not authenticated');
        }
        // Check if token needs refresh
        if (shouldRefreshToken(store.tokens.expiresAt)) {
            try {
                await refreshTokenIfNeeded();
            }
            catch {
                // Token refresh failed, redirect to login
                authStore.logout();
                throw createApiError('UNAUTHENTICATED', 'Session expired');
            }
        }
        // Get fresh token after potential refresh
        const freshStore = get(authStore);
        const token = freshStore.tokens?.accessToken;
        if (!token) {
            throw createApiError('UNAUTHENTICATED', 'No access token available');
        }
        // Add authorization header
        req.header.set('Authorization', `Bearer ${token}`);
        try {
            return await next(req);
        }
        catch (error) {
            // Handle 401 errors
            if (isUnauthorizedError(error)) {
                // Try to refresh token and retry
                try {
                    await refreshTokenIfNeeded(true);
                    const retriedStore = get(authStore);
                    req.header.set('Authorization', `Bearer ${retriedStore.tokens?.accessToken}`);
                    return await next(req);
                }
                catch {
                    authStore.logout();
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
function isPublicPath(url) {
    const path = new URL(url).pathname;
    return PUBLIC_PATHS.some((publicPath) => path.includes(publicPath));
}
/**
 * Checks if token should be refreshed
 */
function shouldRefreshToken(expiresAt) {
    if (!expiresAt)
        return true;
    const now = Date.now();
    const expiryTime = new Date(expiresAt).getTime();
    return expiryTime - now < REFRESH_THRESHOLD;
}
/**
 * Refreshes the token if needed
 */
async function refreshTokenIfNeeded(force = false) {
    const store = get(authStore);
    if (!force && !shouldRefreshToken(store.tokens?.expiresAt)) {
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
        await authStore.refreshTokens();
        // Resolve all queued requests
        refreshQueue.forEach(({ resolve }) => resolve(''));
        refreshQueue = [];
    }
    catch (error) {
        // Reject all queued requests
        const err = error instanceof Error ? error : new Error('Token refresh failed');
        refreshQueue.forEach(({ reject }) => reject(err));
        refreshQueue = [];
        throw error;
    }
    finally {
        isRefreshing = false;
    }
}
/**
 * Checks if error is an unauthorized error
 */
function isUnauthorizedError(error) {
    if (typeof error === 'object' && error !== null) {
        const err = error;
        return err.code === 'UNAUTHENTICATED' || err.code === 'permission_denied';
    }
    return false;
}
//# sourceMappingURL=auth.interceptor.js.map