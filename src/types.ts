import {
	JsonObject,
	Schemas,
	graphError,
	nodeError,
	edgeError,
	ValidateError,
} from "@underlay/pipeline"

import { Instance } from "@underlay/apg"

import * as t from "io-ts"

export interface Context {
	key: string
	directory: string
}

export type Evaluate<
	State extends JsonObject,
	Inputs extends Schemas,
	Outputs extends Schemas
> = (
	state: State,
	inputs: {
		[input in keyof Inputs]: {
			schema: Inputs[input]
			instance: Instance.Instance<Inputs[input]>
		}
	},
	context: Context
) => Promise<
	{
		[output in keyof Outputs]: {
			schema: Outputs[output]
			instance: Instance.Instance<Outputs[output]>
		}
	}
>

// "start" is the start state; one start event is sent for every execution.
// Transitions:
// start -> result
// start -> failure
// result -> result
// result -> failure
// result -> success
// Every evaluate terminates in either a success of failure event.

export type EvaluateEventStart = t.TypeOf<typeof evaluateEventStart>
const evaluateEventStart = t.type({ event: t.literal("start") })

export type EvaluateEventResult = t.TypeOf<typeof evaluateEventResult>
const evaluateEventResult = t.type({ event: t.literal("result"), id: t.string })

export type EvaluateEventFailure = t.TypeOf<typeof evaluateEventFailure>
const evaluateEventFailure = t.type({
	event: t.literal("failure"),
	error: t.union([graphError, nodeError, edgeError]),
})

export type EvaluateEventSuccess = t.TypeOf<typeof evaluateEventSuccess>
const evaluateEventSuccess = t.type({ event: t.literal("success") })

export type EvaluateEvent = t.TypeOf<typeof evaluateEvent>
export const evaluateEvent = t.union([
	evaluateEventStart,
	evaluateEventResult,
	evaluateEventFailure,
	evaluateEventSuccess,
])

export function makeResultEvent(id: string): EvaluateEventResult {
	return { event: "result", id }
}

export function makeFailureEvent(error: ValidateError): EvaluateEventFailure {
	return { event: "failure", error }
}

export function makeSuccessEvent(): EvaluateEventSuccess {
	return { event: "success" }
}
