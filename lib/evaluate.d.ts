import { Graph } from "@underlay/pipeline";
import { EvaluateEventResult, EvaluateEventFailure, EvaluateEventSuccess, Context } from "./types.js";
export default function evaluate(context: Context, graph: Graph): AsyncGenerator<EvaluateEventResult | EvaluateEventFailure | EvaluateEventSuccess, void, undefined>;
