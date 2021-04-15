import { resolve } from "path"

export default {
	entry: resolve("lib", "index.js"),
	output: {
		filename: "index.js",
		library: { type: "commonjs" },
	},
	resolve: { extensions: [".js"] },

	target: "node",
	module: {
		rules: [
			{
				test: /\.js$/,
				exclude: /(node_modules)\//,
				loader: "babel-loader",
			},
		],
	},
}
