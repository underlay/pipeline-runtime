const path = require("path")

module.exports = {
	entry: path.resolve(__dirname, "src", "index.ts"),
	output: {
		filename: "index.js",
		library: { type: "commonjs" },
	},
	resolve: { extensions: [".js", ".ts"] },

	target: "node",
	module: {
		rules: [
			{ test: /\.ts$/, loader: "ts-loader" },
			// {
			// 	test: /\.js$/,
			// 	exclude: /(node_modules)\//,
			// 	loader: "babel-loader",
			// 	options: { presets: ["@babel/preset-env"] },
			// },
		],
	},
}
