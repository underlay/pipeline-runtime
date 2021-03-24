import { Blocks } from "@underlay/pipeline";
import { Evaluate } from "./types.js";
export declare type Runtimes = {
    [k in keyof Blocks]: Evaluate<Blocks[k]["state"], Blocks[k]["inputs"], Blocks[k]["outputs"]>;
};
export declare const runtimes: Runtimes;
