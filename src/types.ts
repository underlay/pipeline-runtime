import { JsonObject, Schemas } from "@underlay/pipeline"

import { Schema, Instance } from "@underlay/apg"

export interface Context {
	host: string
	token: string
	key: string
}

export type Source = { id: string; output: string }

export type EvaluateInput<S extends Schema.Schema> = {
	source: Source
	schema: S
	instance: Instance.Instance<S>
	schemaURI: string
	instanceURI: string
}

export type EvaluateOutput<S extends Schema.Schema> = {
	schema: S
	instance: Instance.Instance<S>
}

export type Evaluate<
	State extends JsonObject,
	Inputs extends Schemas,
	Outputs extends Schemas
> = (
	state: State,
	inputs: { [Input in keyof Inputs]: EvaluateInput<Inputs[Input]> },
	context: Context
) => Promise<{ [Output in keyof Outputs]: EvaluateOutput<Outputs[Output]> }>
