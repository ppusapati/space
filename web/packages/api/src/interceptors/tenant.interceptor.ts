/**
 * Context Interceptor
 * Injects full organizational context into requests
 * tenant_id, company_id, branch_id, user_id
 * @packageDocumentation
 */

import type { Interceptor } from '@connectrpc/connect';
import { getSessionProvider } from '../providers.js';
import { createApiError } from '../client/client.js';

// ============================================================================
// CONTEXT HEADERS
// ============================================================================

/** Header names for organizational context */
const CONTEXT_HEADERS = {
  TENANT: 'X-Tenant-ID',
  COMPANY: 'X-Company-ID',
  BRANCH: 'X-Branch-ID',
  USER: 'X-User-ID',
  SESSION: 'X-Session-ID',
} as const;

/**
 * Paths that don't require organizational context.
 *
 * Mirrors the backend's jwtAuthMiddleware skip-list (app/cmd/main.go).
 * The matcher is `path.includes(...)`, so each entry must be a unique
 * substring of the full pathname. ConnectRPC paths look like
 * `/core.identity.auth.api.v1.AuthService/Login` — the canonical exempt
 * substring is the fully-qualified RPC procedure path, NOT a generic
 * `/auth/` prefix (that prefix never matches because the dotted package
 * qualifier sits between the leading `/` and `auth`).
 *
 * Pre-login flows (Login + 2FA + RefreshToken + password recovery) are
 * exempt by procedure path; .well-known + health/version stay as legacy
 * REST entries.
 */
const CONTEXT_EXEMPT_PATHS = [
  '/core.identity.auth.api.v1.AuthService/Login',
  '/core.identity.auth.api.v1.AuthService/LoginWithOTP',
  '/core.identity.auth.api.v1.AuthService/RefreshToken',
  '/core.identity.auth.api.v1.AuthService/ForgotPassword',
  '/core.identity.auth.api.v1.AuthService/ResetPassword',
  '/.well-known/jwks.json',
  '/health',
  '/ready',
  '/version',
  // Legacy REST exemptions kept for backwards-compat with non-Connect
  // services. Safe — these are unique substrings that never appear
  // inside a ConnectRPC procedure path.
  '/session/init',
  '/tenants/available',
];

// ============================================================================
// CONTEXT INTERCEPTOR
// ============================================================================

/**
 * Creates a context interceptor that injects full organizational context
 */
export function createContextInterceptor(): Interceptor {
  return (next) => async (req) => {
    // Skip context injection for exempt paths
    if (isContextExempt(req.url)) {
      return next(req);
    }

    const session = getSessionProvider();
    const sessionId = session.getSessionId();

    // Check if session exists
    if (!sessionId) {
      throw createApiError('NO_SESSION', 'No active session');
    }

    const context = session.getContext();

    // Validate required context fields
    if (!context?.tenantId) {
      throw createApiError('INVALID_CONTEXT', 'No tenant selected');
    }
    if (!context.companyId) {
      throw createApiError('INVALID_CONTEXT', 'No company selected');
    }
    if (!context.branchId) {
      throw createApiError('INVALID_CONTEXT', 'No branch selected');
    }

    // Add all context headers
    req.header.set(CONTEXT_HEADERS.TENANT, context.tenantId);
    req.header.set(CONTEXT_HEADERS.COMPANY, context.companyId);
    req.header.set(CONTEXT_HEADERS.BRANCH, context.branchId);
    req.header.set(CONTEXT_HEADERS.USER, context.userId);
    req.header.set(CONTEXT_HEADERS.SESSION, sessionId);

    return next(req);
  };
}

/**
 * Creates a tenant-only interceptor (for backwards compatibility)
 * @deprecated Use createContextInterceptor instead
 */
export function createTenantInterceptor(): Interceptor {
  return createContextInterceptor();
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

/**
 * Checks if a path is exempt from context requirement
 */
function isContextExempt(url: string): boolean {
  const path = new URL(url).pathname;
  return CONTEXT_EXEMPT_PATHS.some((exemptPath) => path.includes(exemptPath));
}
