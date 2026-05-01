/**
 * Authentication Guard Middleware
 *
 * Handles:
 * - Session validation
 * - User authentication
 * - Redirect to login for protected routes
 * - Redirect away from login when authenticated
 */

import type {
  Middleware,
  MiddlewareEvent,
  MiddlewareResult,
  AuthGuardConfig,
  Session,
  User,
} from './types.js';

// ============================================================================
// Default Configuration
// ============================================================================

const DEFAULT_CONFIG: Required<Omit<AuthGuardConfig, 'validateSession' | 'fetchUser' | 'onUnauthorized'>> = {
  publicRoutes: ['/login', '/register', '/forgot-password', '/reset-password', '/verify-email'],
  publicPatterns: [/^\/api\/public/, /^\/health/, /^\/_app/],
  loginPath: '/login',
  afterLoginPath: '/',
};

// ============================================================================
// Helpers
// ============================================================================

function isPublicRoute(url: URL, config: AuthGuardConfig): boolean {
  const pathname = url.pathname;

  // Check exact matches
  const publicRoutes = config.publicRoutes ?? DEFAULT_CONFIG.publicRoutes;
  if (publicRoutes.some((route) => pathname === route || pathname.startsWith(route + '/'))) {
    return true;
  }

  // Check patterns
  const publicPatterns = config.publicPatterns ?? DEFAULT_CONFIG.publicPatterns;
  if (publicPatterns.some((pattern) => pattern.test(pathname))) {
    return true;
  }

  return false;
}

function isLoginRoute(url: URL, config: AuthGuardConfig): boolean {
  const loginPath = config.loginPath ?? DEFAULT_CONFIG.loginPath;
  return url.pathname === loginPath || url.pathname.startsWith(loginPath + '/');
}

function getRedirectUrl(url: URL, config: AuthGuardConfig): string {
  const loginPath = config.loginPath ?? DEFAULT_CONFIG.loginPath;
  const returnTo = encodeURIComponent(url.pathname + url.search);
  return `${loginPath}?returnTo=${returnTo}`;
}

// ============================================================================
// Mock Session Store (replace with real implementation)
// ============================================================================

const sessionCache = new Map<string, { session: Session; user: User }>();

async function defaultValidateSession(sessionId: string): Promise<Session | null> {
  const cached = sessionCache.get(sessionId);
  if (!cached) return null;

  // Check expiration
  if (new Date(cached.session.expiresAt) < new Date()) {
    sessionCache.delete(sessionId);
    return null;
  }

  return cached.session;
}

async function defaultFetchUser(userId: string): Promise<User | null> {
  // Find user in cache
  for (const [, data] of sessionCache) {
    if (data.user.id === userId) {
      return data.user;
    }
  }
  return null;
}

// ============================================================================
// Auth Guard Factory
// ============================================================================

/**
 * Create an authentication guard middleware
 */
export function createAuthGuard(config: AuthGuardConfig = {}): Middleware {
  const validateSession = config.validateSession ?? defaultValidateSession;
  const fetchUser = config.fetchUser ?? defaultFetchUser;

  return async (event: MiddlewareEvent): Promise<MiddlewareResult> => {
    const { url, cookies, locals } = event;

    // Initialize auth state
    locals.user = null;
    locals.session = null;
    locals.sessionId = null;

    // Get session ID from cookie
    const sessionId = cookies.get('session') ?? cookies.get('sessionId');

    if (sessionId) {
      locals.sessionId = sessionId;

      try {
        // Validate session
        const session = await validateSession(sessionId);

        if (session) {
          locals.session = session;

          // Fetch user
          const user = await fetchUser(session.userId);

          if (user) {
            locals.user = user;

            // Update last active time (side effect)
            session.lastActiveAt = new Date();
          } else {
            // User not found, invalidate session
            cookies.delete('session', { path: '/' });
            cookies.delete('sessionId', { path: '/' });
            locals.session = null;
            locals.sessionId = null;
          }
        } else {
          // Session expired or invalid
          cookies.delete('session', { path: '/' });
          cookies.delete('sessionId', { path: '/' });
        }
      } catch (error) {
        console.error('[AuthGuard] Session validation error:', error);
        // Clear invalid session
        cookies.delete('session', { path: '/' });
        cookies.delete('sessionId', { path: '/' });
      }
    }

    // Check if route is public
    if (isPublicRoute(url, config)) {
      return { continue: true };
    }

    // Redirect authenticated users away from login
    if (isLoginRoute(url, config) && locals.user) {
      const afterLoginPath = config.afterLoginPath ?? DEFAULT_CONFIG.afterLoginPath;
      const returnTo = url.searchParams.get('returnTo');
      return {
        continue: false,
        redirect: returnTo ?? afterLoginPath,
      };
    }

    // Require authentication for protected routes
    if (!locals.user) {
      // Check for custom handler
      if (config.onUnauthorized) {
        return config.onUnauthorized(event);
      }

      // Default: redirect to login
      return {
        continue: false,
        redirect: getRedirectUrl(url, config),
      };
    }

    // User is authenticated
    return { continue: true };
  };
}

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Set session in cache (for testing/development)
 */
export function setSession(sessionId: string, session: Session, user: User): void {
  sessionCache.set(sessionId, { session, user });
}

/**
 * Clear session from cache
 */
export function clearSession(sessionId: string): void {
  sessionCache.delete(sessionId);
}

/**
 * Clear all sessions
 */
export function clearAllSessions(): void {
  sessionCache.clear();
}

/**
 * Check if user is authenticated
 */
export function isAuthenticated(event: MiddlewareEvent): boolean {
  return event.locals.user !== null;
}

/**
 * Require authentication (throw if not authenticated)
 */
export function requireAuth(event: MiddlewareEvent): User {
  if (!event.locals.user) {
    throw new Error('Authentication required');
  }
  return event.locals.user;
}
