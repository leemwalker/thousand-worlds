export const manifest = (() => {
function __memo(fn) {
	let value;
	return () => value ??= (value = fn());
}

return {
	appDir: "_app",
	appPath: "_app",
	assets: new Set(["favicon.ico","icons/icon-192.png","icons/icon-512.png","manifest.json","offline.html"]),
	mimeTypes: {".png":"image/png",".json":"application/json",".html":"text/html"},
	_: {
		client: {start:"_app/immutable/entry/start.CwShTl_-.js",app:"_app/immutable/entry/app.DokwMO1e.js",imports:["_app/immutable/entry/start.CwShTl_-.js","_app/immutable/chunks/Cn1zA9IX.js","_app/immutable/chunks/vUR8X68U.js","_app/immutable/chunks/DS7xxYIV.js","_app/immutable/entry/app.DokwMO1e.js","_app/immutable/chunks/vUR8X68U.js","_app/immutable/chunks/BYitGVcx.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
		nodes: [
			__memo(() => import('./nodes/0.js')),
			__memo(() => import('./nodes/1.js')),
			__memo(() => import('./nodes/2.js')),
			__memo(() => import('./nodes/3.js'))
		],
		remotes: {
			
		},
		routes: [
			{
				id: "/",
				pattern: /^\/$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 2 },
				endpoint: null
			},
			{
				id: "/game",
				pattern: /^\/game\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 3 },
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
