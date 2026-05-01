/**
 * Middleware Composition
 *
 * Utilities for composing multiple middleware functions
 */

import type { Middleware, MiddlewareEvent, MiddlewareResult } from './types.js';

// ============================================================================
// Composition
// ============================================================================

/**
 * Compose multiple middleware functions into a single middleware
 *
 * Middleware are executed in order. If any middleware returns `continue: false`,
 * execution stops and that result is returned.
 *
 * @example
 * ```ts
 * const middleware = compose(
 *   createAuthGuard(),
 *   createTenantGuard(),
 *   createRouteGuard(routeConfig)
 * );
 * ```
 */
export function compose(...middlewares: Middleware[]): Middleware {
  return async (event: MiddlewareEvent): Promise<MiddlewareResult> => {
    for (const middleware of middlewares) {
      const result = await middleware(event);

      // Merge locals if provided
      if (result.locals) {
        Object.assign(event.locals, result.locals);
      }

      // Stop if middleware says not to continue
      if (!result.continue) {
        return result;
      }
    }

    // All middleware passed
    return { continue: true };
  };
}

/**
 * Compose middleware with error handling
 *
 * Wraps each middleware in try-catch and provides error handling
 */
export function composeWithErrorHandling(
  middlewares: Middleware[],
  onError?: (error: Error, event: MiddlewareEvent) => MiddlewareResult
): Middleware {
  return async (event: MiddlewareEvent): Promise<MiddlewareResult> => {
    for (const middleware of middlewares) {
      try {
        const result = await middleware(event);

        // Merge locals if provided
        if (result.locals) {
          Object.assign(event.locals, result.locals);
        }

        // Stop if middleware says not to continue
        if (!result.continue) {
          return result;
        }
      } catch (error) {
        console.error('[Middleware] Error:', error);

        if (onError) {
          return onError(error as Error, event);
        }

        // Default error handling
        return {
          continue: false,
          error: {
            status: 500,
            message: 'Internal server error',
          },
        };
      }
    }

    return { continue: true };
  };
}

// ============================================================================
// Conditional Middleware
// ============================================================================

/**
 * Run middleware only if condition is met
 */
export function when(
  condition: (event: MiddlewareEvent) => boolean | Promise<boolean>,
  middleware: Middleware
): Middleware {
  return async (event: MiddlewareEvent): Promise<MiddlewareResult> => {
    const shouldRun = await condition(event);

    if (shouldRun) {
      return middleware(event);
    }

    return { continue: true };
  };
}

/**
 * Run middleware only for specific routes
 */
export function forRoutes(routes: (string | RegExp)[], middleware: Middleware): Middleware {
  return when((event) => {
    const pathname = event.url.pathname;

    return routes.some((route) => {
      if (typeof route === 'string') {
        return pathname === route || pathname.startsWith(route + '/');
      }
      return route.test(pathname);
    });
  }, middleware);
}

/**
 * Skip middleware for specific routes
 */
export function exceptRoutes(routes: (string | RegExp)[], middleware: Middleware): Middleware {
  return when((event) => {
    const pathname = event.url.pathname;

    return !routes.some((route) => {
      if (typeof route === 'string') {
        return pathname === route || pathname.startsWith(route + '/');
      }
      return route.test(pathname);
    });
  }, middleware);
}

/**
 * Run middleware only for specific HTTP methods
 */
export function forMethods(methods: string[], middleware: Middleware): Middleware {
  return when((event) => {
    const method = event.request.method.toUpperCase();
    return methods.map((m) => m.toUpperCase()).includes(method);
  }, middleware);
}

// ============================================================================
// Fallback Middleware
// ============================================================================

/**
 * Run first middleware, fall back to second if first doesn't handle
 */
export function fallback(primary: Middleware, fallbackMiddleware: Middleware): Middleware {
  return async (event: MiddlewareEvent): Promise<MiddlewareResult> => {
    const result = await primary(event);

    // If primary handled (didn't continue), return its result
    if (!result.continue) {
      return result;
    }

    // Otherwise run fallback
    return fallbackMiddleware(event);
  };
}

// ============================================================================
// Parallel Middleware
// ============================================================================

/**
 * Run middleware in parallel and combine results
 *
 * All middleware must return `continue: true` for the combined result to continue.
 * First error or redirect wins.
 */
export function parallel(...middlewares: Middleware[]): Middleware {
  return async (event: MiddlewareEvent): Promise<MiddlewareResult> => {
    const results = await Promise.all(
      middlewares.map((m) => m(event))
    );

    // Merge all locals
    for (const result of results) {
      if (result.locals) {
        Object.assign(event.locals, result.locals);
      }
    }

    // Check for any failures
    for (const result of results) {
      if (!result.continue) {
        return result;
      }
    }

    return { continue: true };
  };
}

// ============================================================================
// Utility Middleware
// ============================================================================

/**
 * Middleware that always continues (no-op)
 */
export const passThrough: Middleware = () => ({ continue: true });

/**
 * Middleware that always stops
 */
export function block(error?: { status: number; message: string }): Middleware {
  return () => ({
    continue: false,
    error: error ?? { status: 403, message: 'Forbidden' },
  });
}

/**
 * Middleware that redirects
 */
export function redirect(url: string): Middleware {
  return () => ({
    continue: false,
    redirect: url,
  });
}

/**
 * Middleware that logs requests (for debugging)
 */
export function logger(prefix = '[Middleware]'): Middleware {
  return (event) => {
    const { url, request, locals } = event;
    console.log(
      `${prefix} ${request.method} ${url.pathname}`,
      { user: locals.user?.email, tenant: locals.tenant?.code }
    );
    return { continue: true };
  };
}

// ============================================================================
// SvelteKit Integration
// ============================================================================

/**
 * Convert middleware to SvelteKit handle function
 *
 * @example
 * ```ts
 * // hooks.server.ts
 * import { toSvelteKitHandle, compose, createAuthGuard } from '@samavāya/core/middleware';
 *
 * const middleware = compose(
 *   createAuthGuard(),
 *   createTenantGuard()
 * );
 *
 * export const handle = toSvelteKitHandle(middleware);
 * ```
 */
export function toSvelteKitHandle(middleware: Middleware) {
  return async ({ event, resolve }: { event: any; resolve: (event: any) => Promise<Response> }) => {
    // Create middleware event from SvelteKit event
    const middlewareEvent: MiddlewareEvent = {
      url: event.url,
      request: event.request,
      cookies: event.cookies,
      locals: event.locals,
      params: event.params,
      route: event.route,
    };

    // Run middleware
    const result = await middleware(middlewareEvent);

    // Handle redirect
    if (result.redirect) {
      return new Response(null, {
        status: 302,
        headers: { Location: result.redirect },
      });
    }

    // Handle error
    if (result.error) {
      return new Response(result.error.message, {
        status: result.error.status,
      });
    }

    // Continue to resolve
    return resolve(event);
  };
}
