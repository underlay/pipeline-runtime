import { mkdirSync } from "fs";
import { tmpdir } from "os";
import { resolve } from "path";
import { Kafka } from "kafkajs";
import { Graph, makeGraphError } from "@underlay/pipeline";
import { makeFailureEvent } from "./types.js";
import evaluate from "./evaluate.js";
const rootDirectory = resolve(tmpdir());
console.log("root directory", rootDirectory);
const kafka = new Kafka({ brokers: ["localhost:9092"] });
const producer = kafka.producer();
const consumer = kafka.consumer({ groupId: "pipeline-evaluate-runtime" });
const evaluateTopic = "pipeline-evaluate";
const evaluateEventTopic = "pipeline-evaluate-event";
await producer.connect();
await consumer.connect();
await consumer.subscribe({ topic: evaluateTopic });
await consumer.run({
    eachMessage: async ({ topic, partition, message }) => {
        if (topic === evaluateTopic) {
            const key = message.key.toString();
            console.log("message key", key);
            if (message.value === null) {
                const error = makeGraphError("No message value");
                return sendResult(key, makeFailureEvent(error));
            }
            const graph = JSON.parse(message.value.toString("utf-8"));
            if (!Graph.is(graph)) {
                const error = makeGraphError("Invalid graph value");
                return sendResult(key, makeFailureEvent(error));
            }
            const directory = resolve(rootDirectory, key);
            mkdirSync(directory);
            console.log("job directory", directory);
            const context = { key, directory };
            for await (const event of evaluate(context, graph)) {
                await sendResult(key, event);
            }
            // rmdirSync(directory)
        }
    },
});
async function sendResult(key, result) {
    await producer.send({
        topic: evaluateEventTopic,
        messages: [{ key, value: JSON.stringify(result) }],
    });
}
