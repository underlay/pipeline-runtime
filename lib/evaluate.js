import { readFileSync, writeFileSync } from "fs";
import { resolve } from "path";
import { validateInstance } from "@underlay/apg";
import { encode, decode } from "@underlay/apg-format-binary";
import schemaSchema, { fromSchema, } from "@underlay/apg-schema-schema";
import { sortGraph, blocks, isBlockKind, makeGraphError, makeNodeError, makeEdgeError, domainEqual, } from "@underlay/pipeline";
import { makeFailureEvent, makeResultEvent, makeSuccessEvent, } from "./types.js";
import { runtimes } from "./runtimes.js";
export default async function* evaluate(context, graph) {
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
        const inputSchemas = {};
        for (const [input, edgeId] of Object.entries(node.inputs)) {
            inputSchemas[input] = schemas[edgeId];
        }
        for (const [input, codec] of Object.entries(block.inputs)) {
            const edgeId = node.inputs[input];
            if (!codec.is(inputSchemas[input])) {
                const error = makeEdgeError(edgeId, "Input failed validation");
                return yield makeFailureEvent(error);
            }
        }
        // This is probably the most likely part of evaluation to fail,
        // adding some kind of error handling here would be smart
        const inputs = readInputInstances(context, graph, nodeId, inputSchemas);
        // TS doesn't know that node.kind and node.state are coordinated,
        // or else this would typecheck without coersion
        const evaluate = runtimes[node.kind];
        const event = await evaluate(node.state, inputs, context)
            .then((result) => {
            for (const [output, codec] of Object.entries(block.outputs)) {
                const { schema, instance } = result[output];
                if (!codec.is(schema)) {
                    const message = `Node produced an invalid schema for output ${output}`;
                    const error = makeNodeError(nodeId, message);
                    return makeFailureEvent(error);
                }
                else if (!validateInstance(schema, instance)) {
                    const message = `Node produced an invalid instance for output ${output}`;
                    const error = makeNodeError(nodeId, message);
                    return makeFailureEvent(error);
                }
                else {
                    for (const edgeId of node.outputs[output]) {
                        schemas[edgeId] = schema;
                    }
                    writeSchema(context, { id: nodeId, output }, schema);
                    writeInstance(context, { id: nodeId, output }, schema, instance);
                }
            }
            return makeResultEvent(nodeId);
        })
            .catch((err) => {
            const error = makeNodeError(nodeId, err.toString());
            return makeFailureEvent(error);
        });
        if (event.event === "failure") {
            return yield event;
        }
        else {
            yield event;
        }
    }
    return yield makeSuccessEvent();
}
const getSchemaPath = ({ directory }, source) => resolve(directory, `${source.id}.${source.output}.schema`);
const writeSchema = (context, source, schema) => writeFileSync(getSchemaPath(context, source), encode(schemaSchema, fromSchema(schema)));
const getInstancePath = ({ directory }, source) => resolve(directory, `${source.id}.${source.output}.instance`);
const readInstance = (context, source, schema) => decode(schema, readFileSync(getInstancePath(context, source)));
const writeInstance = (context, source, schema, instance) => writeFileSync(getInstancePath(context, source), encode(schema, instance));
const readInputInstances = (context, graph, id, inputSchemas) => Object.fromEntries(Object.entries(inputSchemas).map(([input, schema]) => {
    const edgeId = graph.nodes[id].inputs[input];
    const { source } = graph.edges[edgeId];
    const instance = readInstance(context, source, schema);
    return [input, { schema, instance }];
}));
