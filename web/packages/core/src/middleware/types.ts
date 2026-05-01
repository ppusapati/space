/**
 * Middleware Types
 */

// ============================================================================
// Core Types
// ============================================================================

export interface User {
  id: string;
  email: string;
  name: string;
  roles: string[];
  permissions: string[];
  tenantId?: string;
  metadata?: Record<string, unknown>;
}

export interface Tenant {
  id: string;
  code: string;
  name: string;
  status: 'active' | 'suspended' | 'inactive';
  features: string[];
  settings?: Record<string, unknown>;
}

export interface Session {
  id: string;
  userId: string;
  tenantId?: string;
  expiresAt: Date;
  createdAt: Date;
  lastActiveAt: Date;
  metadata?: Record<string, unknown>;
}

export interface RequestContext {
  user: User | null;
  tenant: Tenant | null;
  session: Session | null;
  sessionId: string | null;
}

// ============================================================================
// Middleware Types
// ============================================================================

export interface MiddlewareEvent {
  url: URL;
  request: Request;
  cookies: {
    get(name: string): string | undefined;
    set(name: string, value: string, opts?: Record<string, unknown>): void;
    delete(name: string, opts?: Record<string, unknown>): void;
  };
  locals: RequestContext & Record<string, unknown>;
  params: Record<string, string>;
  route: {
    id: string | null;
  };
}

export interface MiddlewareResult {
  /** Whether to continue to next middleware */
  continue: boolean;
  /** Optional redirect URL */
  redirect?: string;
  /** Optional error to throw */
  error?: {
    status: number;
    message: string;
  };
  /** Modified locals */
  locals?: Partial<RequestContext>;
}

export type Middleware = (
  event: MiddlewareEvent
) => Promise<MiddlewareResult> | MiddlewareResult;

export type MiddlewareCondition = (event: MiddlewareEvent) => boolean;

// ============================================================================
// Auth Guard Types
// ============================================================================

export interface AuthGuardConfig {
  /** Routes that don't require authentication */
  publicRoutes?: string[];
  /** Route patterns that don't require authentication */
  publicPatterns?: RegExp[];
  /** Where to redirect unauthenticated users */
  loginPath?: string;
  /** Where to redirect authenticated users from login page */
  afterLoginPath?: string;
  /** Custom session validator */
  validateSession?: (sessionId: string) => Promise<Session | null>;
  /** Custom user fetcher */
  fetchUser?: (userId: string) => Promise<User | null>;
  /** Custom unauthorized handler */
  onUnauthorized?: (event: MiddlewareEvent) => MiddlewareResult;
}

// ============================================================================
// Tenant Guard Types
// ============================================================================

export interface TenantGuardConfig {
  /** How to extract tenant identifier */
  tenantSource?: 'subdomain' | 'path' | 'header' | 'cookie' | 'query';
  /** Header name if using header source */
  tenantHeader?: string;
  /** Cookie name if using cookie source */
  tenantCookie?: string;
  /** Query parameter if using query source */
  tenantQuery?: string;
  /** Path segment index if using path source */
  pathSegmentIndex?: number;
  /** Custom tenant fetcher */
  fetchTenant?: (identifier: string) => Promise<Tenant | null>;
  /** Routes that don't require tenant context */
  excludedRoutes?: string[];
  /** What to do when tenant is not found */
  onTenantNotFound?: (event: MiddlewareEvent, identifier: string | null) => MiddlewareResult;
  /** What to do when tenant is suspended */
  onTenantSuspended?: (event: MiddlewareEvent, tenant: Tenant) => MiddlewareResult;
  /** Validate user has access to tenant */
  validateAccess?: (user: User, tenant: Tenant) => Promise<boolean>;
}

// ============================================================================
// Route Guard Types
// ============================================================================

export interface RoutePermission {
  /** Route pattern (glob or exact) */
  pattern: string;
  /** Required roles (any of these) */
  roles?: string[];
  /** Required permissions (all of these) */
  permissions?: string[];
  /** Required feature flags */
  features?: string[];
  /** Custom condition */
  condition?: (event: MiddlewareEvent) => boolean | Promise<boolean>;
}

export interface RouteGuardConfig {
  /** Route permission definitions */
  routes: RoutePermission[];
  /** Default behavior for unmatched routes */
  defaultAllow?: boolean;
  /** Where to redirect on forbidden */
  forbiddenPath?: string;
  /** Custom forbidden handler */
  onForbidden?: (event: MiddlewareEvent, route: RoutePermission) => MiddlewareResult;
}

// ============================================================================
// CSRF Protection Types
// ============================================================================

export interface CSRFConfig {
  /** Cookie name for CSRF token */
  cookieName?: string;
  /** Header name for CSRF token */
  headerName?: string;
  /** Form field name for CSRF token */
  fieldName?: string;
  /** Methods that require CSRF validation */
  methods?: string[];
  /** Routes excluded from CSRF check */
  excludedRoutes?: string[];
}

// ============================================================================
// Rate Limiting Types
// ============================================================================

export interface RateLimitConfig {
  /** Maximum requests per window */
  maxRequests: number;
  /** Window size in seconds */
  windowSeconds: number;
  /** Key generator (default: IP address) */
  keyGenerator?: (event: MiddlewareEvent) => string;
  /** Routes to apply rate limiting */
  routes?: string[];
  /** Routes to exclude from rate limiting */
  excludedRoutes?: string[];
  /** Custom response when rate limited */
  onRateLimited?: (event: MiddlewareEvent) => MiddlewareResult;
}
