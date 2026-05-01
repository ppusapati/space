const load = async ({ locals }) => {
  return {
    user: locals.user,
    tenant: locals.tenant
  };
};

var _layout_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  load: load
});

const index = 0;
let component_cache;
const component = async () => component_cache ??= (await import('./_layout.svelte-CycCPfnH.js')).default;
const server_id = "src/routes/+layout.server.ts";
const imports = ["_app/immutable/nodes/0.HncFrOXu.js","_app/immutable/chunks/2ca6FVAA.js","_app/immutable/chunks/C_cpAzru.js","_app/immutable/chunks/CJVlgsoI.js","_app/immutable/chunks/CB37v9Ks.js","_app/immutable/chunks/BoMof6nN.js","_app/immutable/chunks/CSVygvZK.js","_app/immutable/chunks/D2C-QB8q.js","_app/immutable/chunks/Dn20OIXa.js","_app/immutable/chunks/DWL_wfiQ.js","_app/immutable/chunks/C0OCMEl6.js","_app/immutable/chunks/CCOeGC2r.js","_app/immutable/chunks/DbqS2JPX.js","_app/immutable/chunks/DCQA5CKL.js","_app/immutable/chunks/ADasbYQY.js","_app/immutable/chunks/CJGPL90U.js"];
const stylesheets = ["_app/immutable/assets/ChartTypePicker.BvclAt2H.css","_app/immutable/assets/0.CFLsfKQV.css"];
const fonts = [];

export { component, fonts, imports, index, _layout_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=0-DPgcS5ub.js.map
