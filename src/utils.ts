import { Readable } from "stream"

import fetch from "node-fetch"

import { PutObjectCommand, GetObjectCommand } from "@aws-sdk/client-s3"

import { Instance, Schema } from "@underlay/apg"
import { encode } from "@underlay/apg-format-binary"
import schemaSchema, { fromSchema } from "@underlay/apg-schema-schema"

import type { Source, Context } from "./types"

import { S3, bucket } from "./s3"

const getSchemaKey = ({ key }: Context, { id, output }: Source) =>
	`executions/${key}/${id}.${output}.schema`

const getInstanceKey = ({ key }: Context, { id, output }: Source) =>
	`executions/${key}/${id}.${output}.instance`

export const getSchemaURI = (context: Context, source: Source) =>
	`s3://${bucket}/${getSchemaKey(context, source)}`

export const getInstanceURI = (context: Context, source: Source) =>
	`s3://${bucket}/${getInstanceKey(context, source)}`

export async function putOutput<S extends Schema.Schema>(
	context: Context,
	source: Source,
	schema: S,
	instance: Instance.Instance<S>
) {
	const schemaData = encode(schemaSchema, fromSchema(schema))
	const schemaKey = getSchemaKey(context, source)
	const schemaCommand = new PutObjectCommand({
		Bucket: bucket,
		Key: schemaKey,
		Body: schemaData,
	})
	await S3.send(schemaCommand)

	const instanceData = encode(schema, instance)

	const instanceKey = getInstanceKey(context, source)
	const command = new PutObjectCommand({
		Bucket: bucket,
		Key: instanceKey,
		Body: instanceData,
	})

	await S3.send(command)
}

interface Handler {
	pattern: RegExp
	resolve(match: RegExpExecArray): Promise<NodeJS.ReadableStream>
}

const handlers: Handler[] = [
	{
		pattern: /^https?:\/\//,
		async resolve({ input }) {
			const res = await fetch(input)
			if (res.ok && res.body !== null) {
				return res.body
			} else {
				throw new Error("Invalid response")
			}
		},
	},
	{
		pattern: /^s3:\/\/([a-z0-9\-\.]+)\/(.+)$/,
		async resolve([_, bucket, key]) {
			const command = new GetObjectCommand({ Bucket: bucket, Key: key })
			const { Body } = await S3.send(command)
			if (Body instanceof Readable) {
				return Body
			} else {
				throw new Error("Unexpected response from S3")
			}
		},
	},
]

export async function resolveURI(uri: string): Promise<NodeJS.ReadableStream> {
	for (const { pattern, resolve } of handlers) {
		const match = pattern.exec(uri)
		if (match !== null) {
			return resolve(match)
		}
	}

	throw new Error("No URI matching handlers found")
}

export async function resolveText(uri: string): Promise<string> {
	const stream = await resolveURI(uri)
	let result = ""
	for await (const chunk of stream) {
		if (Buffer.isBuffer(chunk)) {
			result += chunk.toString("utf-8")
		} else {
			result += chunk
		}
	}
	return result
}
