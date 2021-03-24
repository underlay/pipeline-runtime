import { Buffer } from "buffer";
import Papa from "papaparse";
import fetch from "node-fetch";
import { Instance, Schema, signalInvalidType, zip } from "@underlay/apg";
import { ul } from "@underlay/namespaces";
import { encodeLiteral, decodeLiteral } from "@underlay/apg-format-binary";
import block from "@underlay/pipeline/lib/blocks/csv-import/index.js";
const unit = Instance.unit(Schema.unit());
// parseProperty validates a literal by serializing it and deserializing it again 🙃
function parseProperty(type, value) {
    if (type.kind === "uri") {
        return Instance.uri(type, value);
    }
    else if (type.kind === "literal") {
        const literal = Instance.literal(Schema.literal(type.datatype), value);
        const data = Buffer.concat(Array.from(encodeLiteral(type, literal)));
        const result = decodeLiteral({ data, offset: 0 }, type.datatype);
        return Instance.literal(type, result);
    }
    else {
        signalInvalidType(type);
    }
}
const evaluate = async (state, {}) => {
    const { output: { schema }, } = await block.validate(state, {});
    return new Promise(async (resolve, reject) => {
        const { file } = state;
        if (file === null) {
            throw new Error("No file");
        }
        const product = schema[state.key];
        if (product === undefined) {
            throw new Error("Invalid class key");
        }
        const values = [];
        let skip = state.header;
        const config = { skipEmptyLines: true, header: false };
        const stream = Papa.parse(Papa.NODE_STREAM_INPUT, config);
        stream.on("data", (row) => {
            if (row.length !== state.columns.length) {
                stream.end();
                throw new Error("Bad row length");
            }
            else if (skip) {
                skip = false;
                return;
            }
            const components = {};
            for (const [value, column] of zip(row, state.columns)) {
                if (column === null) {
                    continue;
                }
                const { key, nullValue } = column;
                const property = product.components[key];
                if (property.kind === "coproduct") {
                    if (value === nullValue) {
                        const none = Instance.coproduct(property, ul.none, unit);
                        components[key] = none;
                    }
                    else {
                        const type = property.options[ul.some];
                        components[key] = Instance.coproduct(property, ul.some, parseProperty(type, value));
                    }
                }
                else {
                    try {
                        components[key] = parseProperty(property, value);
                    }
                    catch (e) {
                        stream.end();
                        throw e;
                    }
                }
            }
            values.push(Instance.product(product, components));
        });
        stream.on("end", () => {
            const instance = Instance.instance(schema, { [state.key]: values });
            resolve({ output: { schema, instance } });
        });
        stream.on("error", (error) => reject(error));
        fetch(file).then((res) => res.body.pipe(stream));
    });
};
export default evaluate;
