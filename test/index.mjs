import { readFileSync } from "fs"
import { handler } from "../lib/index.js"

const data = JSON.parse(readFileSync("sample.graph.json", "utf-8"))
const events = await handler(data, {})

console.log(events)
