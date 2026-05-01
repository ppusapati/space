export const manifest = (() => {
function __memo(fn) {
	let value;
	return () => value ??= (value = fn());
}

return {
	appDir: "_app",
	appPath: "_app",
	assets: new Set(["favicon.png","icons.svg","manifest.json","offline.html","sw.js"]),
	mimeTypes: {".png":"image/png",".svg":"image/svg+xml",".json":"application/json",".html":"text/html",".js":"text/javascript"},
	_: {
		client: {start:"_app/immutable/entry/start.COHjfsPy.js",app:"_app/immutable/entry/app.DIL8RvMP.js",imports:["_app/immutable/entry/start.COHjfsPy.js","_app/immutable/chunks/K-2Nw86o.js","_app/immutable/chunks/C_cpAzru.js","_app/immutable/chunks/Dn20OIXa.js","_app/immutable/entry/app.DIL8RvMP.js","_app/immutable/chunks/C_cpAzru.js","_app/immutable/chunks/D2C-QB8q.js","_app/immutable/chunks/2ca6FVAA.js","_app/immutable/chunks/Dn20OIXa.js","_app/immutable/chunks/DWL_wfiQ.js","_app/immutable/chunks/CB37v9Ks.js","_app/immutable/chunks/DbqS2JPX.js","_app/immutable/chunks/DCQA5CKL.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
		nodes: [
			__memo(() => import('./nodes/0.js')),
			__memo(() => import('./nodes/1.js')),
			__memo(() => import('./nodes/2.js')),
			__memo(() => import('./nodes/3.js')),
			__memo(() => import('./nodes/4.js')),
			__memo(() => import('./nodes/5.js')),
			__memo(() => import('./nodes/6.js')),
			__memo(() => import('./nodes/7.js'))
		],
		remotes: {
			
		},
		routes: [
			{
				id: "/",
				pattern: /^\/$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 4 },
				endpoint: null
			},
			{
				id: "/(app)/dashboard",
				pattern: /^\/dashboard\/?$/,
				params: [],
				page: { layouts: [0,2,], errors: [1,,], leaf: 5 },
				endpoint: null
			},
			{
				id: "/(app)/forms/[formId]",
				pattern: /^\/forms\/([^/]+?)\/?$/,
				params: [{"name":"formId","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,2,], errors: [1,,], leaf: 6 },
				endpoint: null
			},
			{
				id: "/(auth)/login",
				pattern: /^\/login\/?$/,
				params: [],
				page: { layouts: [0,3,], errors: [1,,], leaf: 7 },
				endpoint: null
			}
		],
		prerendered_routes: new Set([]),
		matchers: async () => {
			
			return {  };
		},
		server_assets: {}
	}
}
})();
