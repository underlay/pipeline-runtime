import type { Blocks } from "@underlay/pipeline"

import type { Evaluate } from "../types"

import CsvImport from "./csv-import"
import CollectionExport from "./collection-export"

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
