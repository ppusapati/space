/**
 * Tenant Guard Middleware
 *
 * Handles:
 * - Tenant identification from various sources
 * - Tenant validation and status checks
 * - User-tenant access validation
 * - Multi-tenant context setup
 */

import type {
  Middleware,
  MiddlewareEvent,
  MiddlewareResult,
  TenantGuardConfig,
  Tenant,
  User,
} from './types.js';

// ============================================================================
// Default Configuration
// ============================================================================

const DEFAULT_CONFIG: Required<Omit<TenantGuardConfig, 'fetchTenant' | 'onTenantNotFound' | 'onTenantSuspended' | 'validateAccess'>> = {
  tenantSource: 'subdomain',
  tenantHeader: 'X-Tenant-ID',
  tenantCookie: 'tenant',
  tenantQuery: 'tenant',
  pathSegmentIndex: 1,
  excludedRoutes: ['/login', '/register', '/select-tenant', '/api/public', '/health'],
};

// ============================================================================
// Tenant Identification
// ============================================================================

function extractTenantFromSubdomain(url: URL): string | null {
  const hostname = url.hostname;
  const parts = hostname.split('.');

  // Skip if localhost or IP
  if (hostname === 'localhost' || /^\d+\.\d+\.\d+\.\d+$/.test(hostname)) {
    return null;
  }

  // Need at least subdomain.domain.tld
  if (parts.length < 3) {
    return null;
  }

  // Return first subdomain (exclude www)
  const subdomain = parts[0] ?? null;
  if (subdomain === 'www') {
    return parts.length > 3 ? (parts[1] ?? null) : null;
  }

  return subdomain;
}

function extractTenantFromPath(url: URL, segmentIndex: number): string | null {
  const segments = url.pathname.split('/').filter(Boolean);
  return segments[segmentIndex - 1] ?? null;
}

function extractTenantIdentifier(
  event: MiddlewareEvent,
  config: TenantGuardConfig
): string | null {
  const source = config.tenantSource ?? DEFAULT_CONFIG.tenantSource;

  switch (source) {
    case 'subdomain':
      return extractTenantFromSubdomain(event.url);

    case 'path':
      return extractTenantFromPath(
        event.url,
        config.pathSegmentIndex ?? DEFAULT_CONFIG.pathSegmentIndex
      );

    case 'header':
      const headerName = config.tenantHeader ?? DEFAULT_CONFIG.tenantHeader;
      return event.request.headers.get(headerName);

    case 'cookie':
      const cookieName = config.tenantCookie ?? DEFAULT_CONFIG.tenantCookie;
      return event.cookies.get(cookieName) ?? null;

    case 'query':
      const queryParam = config.tenantQuery ?? DEFAULT_CONFIG.tenantQuery;
      return event.url.searchParams.get(queryParam);

    default:
      return null;
  }
}

// ============================================================================
// Mock Tenant Store (replace with real implementation)
// ============================================================================

const tenantCache = new Map<string, Tenant>();

async function defaultFetchTenant(identifier: string): Promise<Tenant | null> {
  // Check by ID or code
  for (const [, tenant] of tenantCache) {
    if (tenant.id === identifier || tenant.code === identifier) {
      return tenant;
    }
  }
  return null;
}

async function defaultValidateAccess(user: User, tenant: Tenant): Promise<boolean> {
  // Check if user belongs to tenant
  return user.tenantId === tenant.id;
}

// ============================================================================
// Route Checking
// ============================================================================

function isExcludedRoute(url: URL, config: TenantGuardConfig): boolean {
  const excludedRoutes = config.excludedRoutes ?? DEFAULT_CONFIG.excludedRoutes;
  const pathname = url.pathname;

  return excludedRoutes.some((route) =>
    pathname === route || pathname.startsWith(route + '/')
  );
}

// ============================================================================
// Tenant Guard Factory
// ============================================================================

/**
 * Create a tenant guard middleware
 */
export function createTenantGuard(config: TenantGuardConfig = {}): Middleware {
  const fetchTenant = config.fetchTenant ?? defaultFetchTenant;
  const validateAccess = config.validateAccess ?? defaultValidateAccess;

  return async (event: MiddlewareEvent): Promise<MiddlewareResult> => {
    const { url, locals } = event;

    // Initialize tenant state
    locals.tenant = null;

    // Skip for excluded routes
    if (isExcludedRoute(url, config)) {
      return { continue: true };
    }

    // Extract tenant identifier
    const identifier = extractTenantIdentifier(event, config);

    if (!identifier) {
      // No tenant identifier found
      if (config.onTenantNotFound) {
        return config.onTenantNotFound(event, null);
      }

      // Default: redirect to tenant selection
      return {
        continue: false,
        redirect: '/select-tenant',
      };
    }

    try {
      // Fetch tenant
      const tenant = await fetchTenant(identifier);

      if (!tenant) {
        // Tenant not found
        if (config.onTenantNotFound) {
          return config.onTenantNotFound(event, identifier);
        }

        return {
          continue: false,
          error: {
            status: 404,
            message: `Tenant not found: ${identifier}`,
          },
        };
      }

      // Check tenant status
      if (tenant.status === 'suspended') {
        if (config.onTenantSuspended) {
          return config.onTenantSuspended(event, tenant);
        }

        return {
          continue: false,
          error: {
            status: 403,
            message: 'This organization has been suspended. Please contact support.',
          },
        };
      }

      if (tenant.status === 'inactive') {
        return {
          continue: false,
          error: {
            status: 403,
            message: 'This organization is inactive.',
          },
        };
      }

      // Set tenant in locals
      locals.tenant = tenant;

      // Validate user access (if user is authenticated)
      if (locals.user) {
        const hasAccess = await validateAccess(locals.user, tenant);

        if (!hasAccess) {
          return {
            continue: false,
            error: {
              status: 403,
              message: 'You do not have access to this organization.',
            },
          };
        }
      }

      return { continue: true };

    } catch (error) {
      console.error('[TenantGuard] Error:', error);
      return {
        continue: false,
        error: {
          status: 500,
          message: 'Failed to validate tenant.',
        },
      };
    }
  };
}

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Register a tenant in cache (for testing/development)
 */
export function registerTenant(tenant: Tenant): void {
  tenantCache.set(tenant.id, tenant);
}

/**
 * Remove a tenant from cache
 */
export function unregisterTenant(tenantId: string): void {
  tenantCache.delete(tenantId);
}

/**
 * Clear all tenants from cache
 */
export function clearTenants(): void {
  tenantCache.clear();
}

/**
 * Check if tenant context is set
 */
export function hasTenant(event: MiddlewareEvent): boolean {
  return event.locals.tenant !== null;
}

/**
 * Require tenant context (throw if not set)
 */
export function requireTenant(event: MiddlewareEvent): Tenant {
  if (!event.locals.tenant) {
    throw new Error('Tenant context required');
  }
  return event.locals.tenant;
}

/**
 * Check if tenant has a specific feature
 */
export function tenantHasFeature(event: MiddlewareEvent, feature: string): boolean {
  return event.locals.tenant?.features.includes(feature) ?? false;
}

/**
 * Require tenant feature (throw if not available)
 */
export function requireTenantFeature(event: MiddlewareEvent, feature: string): void {
  if (!tenantHasFeature(event, feature)) {
    throw new Error(`Feature not available: ${feature}`);
  }
}
