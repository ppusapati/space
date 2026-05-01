const load = async ({ locals }) => {
  return {
    user: locals.user,
    tenant: locals.tenant
  };
};
export {
  load
};
