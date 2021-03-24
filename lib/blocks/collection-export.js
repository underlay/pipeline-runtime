import { createHash } from "crypto";
import tar from "tar-stream";
import fetch from "node-fetch";
import StatusCodes from "http-status-codes";
import { encode } from "@underlay/apg-format-binary";
import schemaSchema, { fromSchema } from "@underlay/apg-schema-schema";
const evaluate = async ({ etag, url, readme }, { input: { schema, instance } }, { key }) => {
    if (url === null || readme === null) {
        throw new Error("Invalid state");
    }
    const pack = tar.pack();
    pack.entry({ name: "README.md" }, readme);
    pack.entry({ name: "index.schema" }, encode(schemaSchema, fromSchema(schema)));
    const buffer = encode(schema, instance);
    const hash = createHash("sha256");
    hash.update(buffer);
    const name = `instances/${hash.digest("hex")}`;
    pack.entry({ name }, buffer, (err) => pack.finalize());
    const headers = { "Content-Type": "application/x-tar" };
    if (etag !== null) {
        headers["If-Match"] = `"${etag}"`;
    }
    const res = await fetch(`${url}?key=${key}`, {
        method: "POST",
        headers,
        body: pack,
    });
    if (res.status === StatusCodes.CREATED) {
        return {};
    }
    else {
        throw new Error(`Publishing collection failed with status code ${res.status}`);
    }
};
export default evaluate;
