import type { Handle } from '@sveltejs/kit';

/**
 * Server-side hook for authentication and tenant context
 */
export const handle: Handle = async ({ event, resolve }) => {
  // Get session from cookie
  const sessionId = event.cookies.get('session');

  // Initialize locals
  event.locals.user = null;
  event.locals.tenant = null;
  event.locals.sessionId = sessionId ?? null;

  // If session exists, validate and fetch user
  if (sessionId) {
    try {
      // TODO: Validate session with backend API
      // const session = await validateSession(sessionId);
      // event.locals.user = session.user;
      // event.locals.tenant = session.tenant;
    } catch {
      // Invalid session, clear cookie
      event.cookies.delete('session', { path: '/' });
    }
  }

  // Resolve the request
  const response = await resolve(event, {
    transformPageChunk: ({ html }) => {
      // Can inject theme or other data into HTML
      return html;
    },
  });

  return response;
};
