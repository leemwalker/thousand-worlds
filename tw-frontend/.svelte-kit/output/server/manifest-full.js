export const manifest = (() => {
function __memo(fn) {
	let value;
	return () => value ??= (value = fn());
}

return {
	appDir: "_app",
	appPath: "_app",
	assets: new Set(["favicon.ico","manifest.json","offline.html"]),
	mimeTypes: {".json":"application/json",".html":"text/html"},
	_: {
		client: {start:"_app/immutable/entry/start.Boall3f-.js",app:"_app/immutable/entry/app.BofA7a1k.js",imports:["_app/immutable/entry/start.Boall3f-.js","_app/immutable/chunks/BVL9hfkl.js","_app/immutable/chunks/CNXArMli.js","_app/immutable/chunks/B0GIxsE8.js","_app/immutable/entry/app.BofA7a1k.js","_app/immutable/chunks/CNXArMli.js","_app/immutable/chunks/BWH_Fj_B.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
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
