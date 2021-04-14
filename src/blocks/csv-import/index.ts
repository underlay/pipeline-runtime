import { Buffer } from "buffer"
import Papa from "papaparse"

import { Instance, Schema, signalInvalidType, zip } from "@underlay/apg"

import { OptionalProperty, Property } from "@underlay/apg-codec-table"
import { ul } from "@underlay/namespaces"

import { encodeLiteral, decodeLiteral } from "@underlay/apg-format-binary"

import block, { State, Inputs, Outputs } from "@underlay/pipeline/csv-import"

import type { Evaluate } from "../../types"

import { resolveText } from "../../utils"

const unit = Instance.unit(Schema.unit())

// parseProperty validates a literal by serializing it and deserializing it again 🙃
function parseProperty(
	type: Property,
	value: string
): Instance.Value<Property> {
	if (type.kind === "uri") {
		return Instance.uri(type, value)
	} else if (type.kind === "literal") {
		const literal = Instance.literal(Schema.literal(type.datatype), value)
		const data = Buffer.concat(Array.from(encodeLiteral(type, literal)))
		const result = decodeLiteral({ data, offset: 0 }, type.datatype)
		return Instance.literal(type, result)
	} else {
		signalInvalidType(type)
	}
}

const evaluate: Evaluate<State, Inputs, Outputs> = async (state, {}) => {
	const {
		output: { schema },
	} = await block.validate(state, {})

	if (state.uri === null) {
		throw new Error("No file")
	} else {
		const {} = new URL(state.key)
	}

	const product = schema[state.key]
	if (product === undefined) {
		throw new Error("Invalid class key")
	}

	const values: Instance.Value<Outputs["output"][string]>[] = []

	const config = { skipEmptyLines: false, header: false }

	const input = await resolveText(state.uri)

	const result = Papa.parse<string[]>(input, config)

	if (result.errors.length > 0) {
		const message = result.errors
			.map((error) => `${error.type} in row ${error.row}: ${error.message}`)
			.join("\n")
		throw new Error(message)
	}

	const rows = state.header ? result.data.slice(1) : result.data
	for (const row of rows) {
		const components: Record<string, Instance.Value<OptionalProperty>> = {}
		for (const [value, column] of zip(row, state.columns)) {
			if (column === null) {
				continue
			}

			const { key, nullValue } = column
			const property = product.components[key]
			if (property.kind === "coproduct") {
				if (value === nullValue) {
					const none = Instance.coproduct(property, ul.none, unit)
					components[key] = none
				} else {
					const type = property.options[ul.some]
					components[key] = Instance.coproduct(
						property,
						ul.some,
						parseProperty(type, value)
					)
				}
			} else {
				components[key] = parseProperty(property, value)
			}
		}
		values.push(
			Instance.product<Record<string, OptionalProperty>>(product, components)
		)
	}

	const instance = Instance.instance(schema, { [state.key]: values })

	return { output: { schema, instance } }
}

export default evaluate
