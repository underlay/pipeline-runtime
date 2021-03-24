import { readFileSync } from "fs"
import { Kafka } from "kafkajs"
import { v4 as uuid } from "uuid"

const kafka = new Kafka({ brokers: ["localhost:9092"] })

const producer = kafka.producer()

const key = uuid()
console.log("key", key)
const value = readFileSync("sample.graph.json")

console.log("connecting...")
await producer.connect()
console.log("connected. sending...")
await producer.send({
	topic: "pipeline-evaluate",
	messages: [{ key, value }],
})
console.log("sent. disconnecting...")
await producer.disconnect()
