import { Graph, EvaluateEvent } from "@underlay/pipeline";
import { Context } from "./types.js";
export default function evaluate(context: Context, graph: Graph): AsyncGenerator<EvaluateEvent, void, undefined>;
