/**
 * Route Guard Middleware
 *
 * Handles:
 * - Role-based access control
 * - Permission-based access control
 * - Feature flag gating
 * - Custom route conditions
 */

import type {
  Middleware,
  MiddlewareEvent,
  MiddlewareResult,
  RouteGuardConfig,
  RoutePermission,
  User,
} from './types.js';

// ============================================================================
// Default Configuration
// ============================================================================

const DEFAULT_CONFIG: Required<Omit<RouteGuardConfig, 'onForbidden'>> = {
  routes: [],
  defaultAllow: true,
  forbiddenPath: '/403',
};

// ============================================================================
// Pattern Matching
// ============================================================================

function patternToRegex(pattern: string): RegExp {
  // Convert glob pattern to regex
  const escaped = pattern
    .replace(/[.+?^${}()|[\]\\]/g, '\\$&') // Escape special chars
    .replace(/\*\*/g, '{{GLOBSTAR}}') // Placeholder for **
    .replace(/\*/g, '[^/]*') // * matches any chars except /
    .replace(/{{GLOBSTAR}}/g, '.*'); // ** matches anything

  return new RegExp(`^${escaped}$`);
}

function matchRoute(pathname: string, pattern: string): boolean {
  // Exact match
  if (pattern === pathname) {
    return true;
  }

  // Pattern match
  const regex = patternToRegex(pattern);
  return regex.test(pathname);
}

function findMatchingRoute(
  pathname: string,
  routes: RoutePermission[]
): RoutePermission | null {
  // Find the most specific matching route
  let bestMatch: RoutePermission | null = null;
  let bestSpecificity = -1;

  for (const route of routes) {
    if (matchRoute(pathname, route.pattern)) {
      // Calculate specificity (more specific patterns have higher priority)
      const specificity = route.pattern
        .split('/')
        .filter((s) => !s.includes('*')).length;

      if (specificity > bestSpecificity) {
        bestMatch = route;
        bestSpecificity = specificity;
      }
    }
  }

  return bestMatch;
}

// ============================================================================
// Permission Checking
// ============================================================================

function hasAnyRole(user: User, requiredRoles: string[]): boolean {
  return requiredRoles.some((role) => user.roles.includes(role));
}

function hasAllPermissions(user: User, requiredPermissions: string[]): boolean {
  return requiredPermissions.every((perm) => user.permissions.includes(perm));
}

function hasAllFeatures(
  event: MiddlewareEvent,
  requiredFeatures: string[]
): boolean {
  const tenantFeatures = event.locals.tenant?.features ?? [];
  return requiredFeatures.every((feature) => tenantFeatures.includes(feature));
}

async function checkCondition(
  event: MiddlewareEvent,
  condition: RoutePermission['condition']
): Promise<boolean> {
  if (!condition) return true;
  return await condition(event);
}

// ============================================================================
// Route Guard Factory
// ============================================================================

/**
 * Create a route guard middleware
 */
export function createRouteGuard(config: RouteGuardConfig): Middleware {
  return async (event: MiddlewareEvent): Promise<MiddlewareResult> => {
    const { url, locals } = event;
    const pathname = url.pathname;

    // Find matching route permission
    const routePermission = findMatchingRoute(pathname, config.routes);

    // No matching route - use default behavior
    if (!routePermission) {
      const defaultAllow = config.defaultAllow ?? DEFAULT_CONFIG.defaultAllow;
      return { continue: defaultAllow };
    }

    // Get user (may be null for unauthenticated)
    const user = locals.user;

    // Check roles (if specified)
    if (routePermission.roles?.length) {
      if (!user) {
        // No user, roles required - deny
        return createForbiddenResult(event, routePermission, config);
      }

      if (!hasAnyRole(user, routePermission.roles)) {
        return createForbiddenResult(event, routePermission, config);
      }
    }

    // Check permissions (if specified)
    if (routePermission.permissions?.length) {
      if (!user) {
        // No user, permissions required - deny
        return createForbiddenResult(event, routePermission, config);
      }

      if (!hasAllPermissions(user, routePermission.permissions)) {
        return createForbiddenResult(event, routePermission, config);
      }
    }

    // Check features (if specified)
    if (routePermission.features?.length) {
      if (!hasAllFeatures(event, routePermission.features)) {
        return createForbiddenResult(event, routePermission, config);
      }
    }

    // Check custom condition (if specified)
    if (routePermission.condition) {
      const conditionMet = await checkCondition(event, routePermission.condition);
      if (!conditionMet) {
        return createForbiddenResult(event, routePermission, config);
      }
    }

    // All checks passed
    return { continue: true };
  };
}

function createForbiddenResult(
  event: MiddlewareEvent,
  route: RoutePermission,
  config: RouteGuardConfig
): MiddlewareResult {
  // Check for custom handler
  if (config.onForbidden) {
    return config.onForbidden(event, route);
  }

  // Default: redirect to forbidden page
  const forbiddenPath = config.forbiddenPath ?? DEFAULT_CONFIG.forbiddenPath;
  return {
    continue: false,
    redirect: forbiddenPath,
  };
}

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Check if user has a specific role
 */
export function hasRole(event: MiddlewareEvent, role: string): boolean {
  return event.locals.user?.roles.includes(role) ?? false;
}

/**
 * Check if user has any of the specified roles
 */
export function hasRoles(event: MiddlewareEvent, roles: string[]): boolean {
  const user = event.locals.user;
  if (!user) return false;
  return roles.some((role) => user.roles.includes(role));
}

/**
 * Check if user has a specific permission
 */
export function hasPermission(event: MiddlewareEvent, permission: string): boolean {
  return event.locals.user?.permissions.includes(permission) ?? false;
}

/**
 * Check if user has all specified permissions
 */
export function hasPermissions(event: MiddlewareEvent, permissions: string[]): boolean {
  const user = event.locals.user;
  if (!user) return false;
  return permissions.every((perm) => user.permissions.includes(perm));
}

/**
 * Require a specific role (throw if not met)
 */
export function requireRole(event: MiddlewareEvent, role: string): void {
  if (!hasRole(event, role)) {
    throw new Error(`Role required: ${role}`);
  }
}

/**
 * Require a specific permission (throw if not met)
 */
export function requirePermission(event: MiddlewareEvent, permission: string): void {
  if (!hasPermission(event, permission)) {
    throw new Error(`Permission required: ${permission}`);
  }
}

// ============================================================================
// Common Route Configurations
// ============================================================================

/**
 * Admin-only routes
 */
export const ADMIN_ROUTES: RoutePermission[] = [
  { pattern: '/admin/**', roles: ['admin', 'super_admin'] },
  { pattern: '/settings/organization/**', roles: ['admin', 'super_admin'] },
  { pattern: '/settings/users/**', roles: ['admin', 'super_admin'] },
  { pattern: '/settings/billing/**', roles: ['admin', 'super_admin'] },
];

/**
 * Manager-level routes
 */
export const MANAGER_ROUTES: RoutePermission[] = [
  { pattern: '/reports/**', roles: ['manager', 'admin', 'super_admin'] },
  { pattern: '/approvals/**', roles: ['manager', 'admin', 'super_admin'] },
  { pattern: '/team/**', roles: ['manager', 'admin', 'super_admin'] },
];

/**
 * Feature-gated routes
 */
export const FEATURE_ROUTES: RoutePermission[] = [
  { pattern: '/inventory/**', features: ['inventory'] },
  { pattern: '/manufacturing/**', features: ['manufacturing'] },
  { pattern: '/pos/**', features: ['point_of_sale'] },
  { pattern: '/crm/**', features: ['crm'] },
  { pattern: '/hrms/**', features: ['hrms'] },
  { pattern: '/payroll/**', features: ['payroll'] },
];
