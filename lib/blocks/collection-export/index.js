import fetch from "node-fetch";
const evaluate = async ({ etag, id, readme }, { input }, { host, token, key }) => {
    if (id === null || readme === null) {
        throw new Error("Invalid state");
    }
    const url = `http://${host}/api/collection/${id}?key=${key}&token=${token}`;
    const headers = {
        "content-type": "text/markdown",
        "x-collection-schema": input.schemaURI,
        "x-collection-instance": input.instanceURI,
    };
    // // TODO: Uncomment this after etags are actually implemented
    // if (etag !== null) {
    // 	headers["if-match"] = `"${etag}"`
    // }
    const res = await fetch(url, { method: "POST", headers, body: readme });
    if (res.ok) {
        return {};
    }
    else {
        throw new Error(`Publishing collection failed with status code ${res.status}`);
    }
};
export default evaluate;
