import * as t from "io-ts"

import { isLeft } from "fp-ts/lib/Either.js"

import {
	Graph,
	makeGraphError,
	EvaluateEvent,
	makeFailureEvent,
	makeStartEvent,
} from "@underlay/pipeline"

import evaluate from "./evaluate"

const evaluateEvent = t.type({
	host: t.string,
	key: t.string,
	token: t.string,
	graph: Graph,
})

export async function handler(event: any, {}: {}): Promise<EvaluateEvent[]> {
	const result = evaluateEvent.decode(event)
	if (isLeft(result)) {
		const error = makeGraphError("Invalid graph value")
		return [makeStartEvent(), makeFailureEvent(error)]
	}

	const { host, key, token, graph } = result.right

	const events: EvaluateEvent[] = []
	const context = { host, key, token }
	for await (const event of evaluate(context, graph)) {
		events.push(event)
	}

	return events
}
