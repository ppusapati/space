/**
 * Context Interceptor
 * Injects full organizational context into requests
 * tenant_id, company_id, branch_id, user_id
 * @packageDocumentation
 */
import type { Interceptor } from '@connectrpc/connect';
/**
 * Creates a context interceptor that injects full organizational context
 */
export declare function createContextInterceptor(): Interceptor;
/**
 * Creates a tenant-only interceptor (for backwards compatibility)
 * @deprecated Use createContextInterceptor instead
 */
export declare function createTenantInterceptor(): Interceptor;
//# sourceMappingURL=tenant.interceptor.d.ts.map