import { resolve } from "path";
import { mkdirSync, rmdirSync } from "fs";
import * as t from "io-ts";
import { isLeft } from "fp-ts/lib/Either.js";
import { Graph, makeGraphError, makeFailureEvent, makeStartEvent, } from "@underlay/pipeline";
import evaluate from "./evaluate.js";
const host = process.env.COLLECTION_SERVER_HOST;
try {
    new URL(host);
}
catch {
    throw new Error("Invalid COLLECTION_SERVER_HOST environment variable");
}
const evaluateEvent = t.type({ key: t.string, token: t.string, graph: Graph });
const rootDirectory = resolve();
export async function handler(event, {}) {
    const result = evaluateEvent.decode(event);
    if (isLeft(result)) {
        const error = makeGraphError("Invalid graph value");
        return [makeStartEvent(), makeFailureEvent(error)];
    }
    const { key, token, graph } = result.right;
    const directory = resolve(rootDirectory, key);
    mkdirSync(directory);
    const events = [];
    const context = { host, key, token, directory };
    for await (const event of evaluate(context, graph)) {
        events.push(event);
    }
    rmdirSync(directory, { recursive: true });
    return events;
}
