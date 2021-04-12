import { readFileSync, writeFileSync } from "fs";
import { resolve } from "path";
import { Readable } from "stream";
import fetch from "node-fetch";
import { PutObjectCommand, GetObjectCommand } from "@aws-sdk/client-s3";
import { decode, encode } from "@underlay/apg-format-binary";
import schemaSchema, { fromSchema } from "@underlay/apg-schema-schema";
import { S3, bucket } from "./s3.js";
const getSchemaName = ({ id, output }) => `${id}.${output}.schema`;
const getSchemaKey = ({ key }, source) => `executions/${key}/${getSchemaName(source)}`;
const getInstanceName = ({ id, output }) => `${id}.${output}.instance`;
const getInstanceKey = ({ key }, source) => `executions/${key}/${getInstanceName(source)}`;
export const getSchemaURI = (context, source) => `s3://${bucket}/${getSchemaKey(context, source)}`;
export const getInstanceURI = (context, source) => `s3://${bucket}/${getInstanceKey(context, source)}`;
export async function putOutput(context, source, schema, instance) {
    const schemaData = encode(schemaSchema, fromSchema(schema));
    const schemaKey = getSchemaKey(context, source);
    const schemaCommand = new PutObjectCommand({
        Bucket: bucket,
        Key: schemaKey,
        Body: schemaData,
    });
    await S3.send(schemaCommand);
    const instanceData = encode(schema, instance);
    const path = resolve(context.directory, getInstanceName(source));
    writeFileSync(path, instanceData);
    const instanceKey = getInstanceKey(context, source);
    const command = new PutObjectCommand({
        Bucket: bucket,
        Key: instanceKey,
        Body: instanceData,
    });
    await S3.send(command);
}
export async function getOutput({ directory }, source, schema) {
    const path = resolve(directory, getInstanceName(source));
    const data = readFileSync(path);
    return decode(schema, data);
}
const handlers = [
    {
        pattern: /^https?:\/\//,
        async resolve({ input }) {
            const res = await fetch(input);
            if (res.ok && res.body !== null) {
                return res.body;
            }
            else {
                throw new Error("Invalid response");
            }
        },
    },
    {
        pattern: /^s3:\/\/([a-z0-9\-\.]+)\/(.+)$/,
        async resolve([_, bucket, key]) {
            const command = new GetObjectCommand({ Bucket: bucket, Key: key });
            const { Body } = await S3.send(command);
            if (Body instanceof Readable) {
                return Body;
            }
            else {
                throw new Error("Unexpected response from S3");
            }
        },
    },
];
export async function resolveURI(uri) {
    for (const { pattern, resolve } of handlers) {
        const match = pattern.exec(uri);
        if (match !== null) {
            return resolve(match);
        }
    }
    throw new Error("No URI matching handlers found");
}
