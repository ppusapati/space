const handle = async ({ event, resolve }) => {
  const sessionId = event.cookies.get("session");
  event.locals.user = null;
  event.locals.tenant = null;
  event.locals.sessionId = sessionId ?? null;
  const response = await resolve(event, {
    transformPageChunk: ({ html }) => {
      return html;
    }
  });
  return response;
};

export { handle };
//# sourceMappingURL=hooks.server-D7fCRx-J.js.map
