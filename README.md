# pipeline-runtime

The [src/index.ts](./src/index.ts) script starts a kafka consumer on topic `pipeline-evaluate` that receives pipeline graphs and evaluates the blocks in topological order.

The `evaluate` function exported from `src/evaluate.ts` is an async generator the yields `EvaluateEvent` event objects. Each event has a `.event` property that is one of `start`, `result`, `failure`, or `success`. Every pipeline execution begins with a `start` event and ends with _either_ a `failure` or `success` event, with zero or more `result` events in-between.

![](./event-fsm.svg)

The `failure` event carries a `ValidateError` object.

```typescript
type EvaluateEvent =
	| { event: "start" }
	| { event: "result"; id: string }
	| { event: "failure"; error: ValidateError }
	| { event: "success" }

type ValidateError =
	| { type: "validate/graph"; message: string }
	| { type: "validate/node"; id: string; message: string }
	| { type: "validate/edge"; id: string; message: string }
```

## Deploy

```
tsc
npm run build
npm run zip
```

Then upload `pipeline-runtime.zip` to lambda.
