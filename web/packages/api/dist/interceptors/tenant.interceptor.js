/**
 * Context Interceptor
 * Injects full organizational context into requests
 * tenant_id, company_id, branch_id, user_id
 * @packageDocumentation
 */
import { sessionStore } from '@samavāya/stores';
import { get } from 'svelte/store';
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
};
/** Paths that don't require organizational context */
const CONTEXT_EXEMPT_PATHS = [
    '/auth/',
    '/session/init',
    '/tenants/available',
    '/health',
    '/version',
];
// ============================================================================
// CONTEXT INTERCEPTOR
// ============================================================================
/**
 * Creates a context interceptor that injects full organizational context
 */
export function createContextInterceptor() {
    return (next) => async (req) => {
        // Skip context injection for exempt paths
        if (isContextExempt(req.url)) {
            return next(req);
        }
        const state = get(sessionStore);
        // Check if session exists
        if (!state.session) {
            throw createApiError('NO_SESSION', 'No active session');
        }
        const { context } = state.session;
        // Validate required context fields
        if (!context.tenantId) {
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
        req.header.set(CONTEXT_HEADERS.SESSION, state.session.id);
        return next(req);
    };
}
/**
 * Creates a tenant-only interceptor (for backwards compatibility)
 * @deprecated Use createContextInterceptor instead
 */
export function createTenantInterceptor() {
    return createContextInterceptor();
}
// ============================================================================
// HELPER FUNCTIONS
// ============================================================================
/**
 * Checks if a path is exempt from context requirement
 */
function isContextExempt(url) {
    const path = new URL(url).pathname;
    return CONTEXT_EXEMPT_PATHS.some((exemptPath) => path.includes(exemptPath));
}
//# sourceMappingURL=tenant.interceptor.js.map