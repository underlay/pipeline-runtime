import { isLeft, left, right } from "fp-ts/lib/Either.js";
import { validateInstance } from "@underlay/apg";
import { sortGraph, blocks, isBlockKind, makeGraphError, makeNodeError, makeEdgeError, domainEqual, makeStartEvent, makeFailureEvent, makeResultEvent, makeSuccessEvent, } from "@underlay/pipeline";
import { runtimes } from "./blocks/index.js";
import { getInstanceURI, getOutput, getSchemaURI, putOutput } from "./utils.js";
export default async function* evaluate(context, graph) {
    yield makeStartEvent();
    const order = sortGraph(graph);
    if (order === null) {
        const error = makeGraphError("Cycle detected");
        return yield makeFailureEvent(error);
    }
    const schemas = {};
    for (const nodeId of order) {
        const node = graph.nodes[nodeId];
        if (!isBlockKind(node.kind)) {
            const message = `Invalid node kind: ${JSON.stringify(node.kind)}`;
            const error = makeNodeError(nodeId, message);
            return yield makeFailureEvent(error);
        }
        const block = blocks[node.kind];
        if (!block.state.is(node.state)) {
            const error = makeNodeError(nodeId, "Invalid state");
            return yield makeFailureEvent(error);
        }
        else if (!domainEqual(block.inputs, node.inputs)) {
            const error = makeNodeError(nodeId, "Missing or extra inputs");
            return yield makeFailureEvent(error);
        }
        else if (!domainEqual(block.outputs, node.outputs)) {
            const error = makeNodeError(nodeId, "Missing or extra outputs");
            return yield makeFailureEvent(error);
        }
        const inputs = {};
        for (const [input, codec] of Object.entries(block.inputs)) {
            const edgeId = node.inputs[input];
            const schema = schemas[edgeId];
            if (!codec.is(schema)) {
                const error = makeEdgeError(edgeId, "Input failed validation");
                return yield makeFailureEvent(error);
            }
            const { source } = graph.edges[edgeId];
            inputs[input] = getInput(context, source, schema);
        }
        // TS doesn't know that node.kind and node.state are coordinated,
        // or else this would typecheck without coersion
        const evaluate = runtimes[node.kind];
        const result = await evaluate(node.state, inputs, context).then((result) => right(result), (err) => left(makeNodeError(nodeId, err.toString())));
        if (isLeft(result)) {
            return yield makeFailureEvent(result.left);
        }
        for (const [output, codec] of Object.entries(block.outputs)) {
            const { schema, instance } = result.right[output];
            if (!codec.is(schema)) {
                const message = `Node produced an invalid schema for output ${output}`;
                const error = makeNodeError(nodeId, message);
                return yield makeFailureEvent(error);
            }
            if (!validateInstance(schema, instance)) {
                const message = `Node produced an invalid instance for output ${output}`;
                const error = makeNodeError(nodeId, message);
                return yield makeFailureEvent(error);
            }
            for (const edgeId of node.outputs[output]) {
                schemas[edgeId] = schema;
            }
            await putOutput(context, { id: nodeId, output }, schema, instance);
        }
        yield makeResultEvent(nodeId);
    }
    return yield makeSuccessEvent();
}
function getInput(context, source, schema) {
    return {
        schemaURI: getSchemaURI(context, source),
        instanceURI: getInstanceURI(context, source),
        source,
        schema,
        getInstance: () => getOutput(context, source, schema),
    };
}
