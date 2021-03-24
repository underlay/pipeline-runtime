declare module "papaparse" {
	const NODE_STREAM_INPUT = 1
	function parse(
		flag: typeof NODE_STREAM_INPUT,
		config: { skipEmptyLines: boolean; header: boolean }
	): NodeJS.WritableStream
}
