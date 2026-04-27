export const manifest = (() => {
function __memo(fn) {
	let value;
	return () => value ??= (value = fn());
}

return {
	appDir: "_app",
	appPath: "_app",
	assets: new Set([]),
	mimeTypes: {},
	_: {
		client: {start:"_app/immutable/entry/start.CIwXXTfO.js",app:"_app/immutable/entry/app.Dq3Ikgq6.js",imports:["_app/immutable/entry/start.CIwXXTfO.js","_app/immutable/chunks/BKDrm4Ee.js","_app/immutable/chunks/AVpQJuEG.js","_app/immutable/chunks/DIeogL5L.js","_app/immutable/entry/app.Dq3Ikgq6.js","_app/immutable/chunks/AVpQJuEG.js","_app/immutable/chunks/DIeogL5L.js","_app/immutable/chunks/piy101Fz.js","_app/immutable/chunks/Bzak7iHL.js","_app/immutable/chunks/DK-l3hE2.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
		nodes: [
			__memo(() => import('./nodes/0.js')),
			__memo(() => import('./nodes/1.js')),
			__memo(() => import('./nodes/2.js'))
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
