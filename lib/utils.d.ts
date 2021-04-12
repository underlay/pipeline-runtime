/// <reference types="node" />
import { Instance, Schema } from "@underlay/apg";
import type { Source, Context } from "./types.js";
export declare const getSchemaURI: (context: Context, source: Source) => string;
export declare const getInstanceURI: (context: Context, source: Source) => string;
export declare function putOutput<S extends Schema.Schema>(context: Context, source: Source, schema: S, instance: Instance.Instance<S>): Promise<void>;
export declare function getOutput<S extends Schema.Schema>({ directory }: Context, source: Source, schema: S): Promise<Instance.Instance<S>>;
export declare function resolveURI(uri: string): Promise<NodeJS.ReadableStream>;
