import { Kafka } from "kafkajs"

const kafka = new Kafka({ brokers: ["localhost:9092"] })

const consumer = kafka.consumer({ groupId: "test-group" })

await consumer.connect()
await consumer.subscribe({
	topic: "pipeline-evaluate-event",
	fromBeginning: true,
})

await consumer.run({
	eachMessage: async ({ topic, partition, message }) => {
		const key = message.key.toString()
		const result = JSON.parse(message.value)
		console.log("event", key, result)
	},
})
