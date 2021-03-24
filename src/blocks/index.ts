// Type-only import
import { Blocks } from "@underlay/pipeline"

import { Evaluate } from "../types.js"

import CsvImport from "./csv-import/index.js"
import CollectionExport from "./collection-export/index.js"

export type Runtimes = {
	[k in keyof Blocks]: Evaluate<
		Blocks[k]["state"],
		Blocks[k]["inputs"],
		Blocks[k]["outputs"]
	>
}

export const runtimes: Runtimes = {
	"csv-import": CsvImport,
	"collection-export": CollectionExport,
}
