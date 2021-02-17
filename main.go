package main

import (
	"io/ioutil"
	"log"
	"path/filepath"
)

func main() {
	// fmt.Println("Hello world")
	// Init workflow with a name, and a number for max concurrent tasks, so we
	// don't overbook our CPU (it is recommended to set it to the number of CPU
	// cores of your computer)
	// wf := scipipe.NewWorkflow("hello_world", 4)

	// // Initialize processes and set output file paths
	// hello := wf.NewProc("hello", "echo 'Hello ' > {o:out}")
	// hello.SetOut("out", "hello.txt")

	// world := wf.NewProc("world", "echo $(cat {i:in}) World >> {o:out}")
	// world.SetOut("out", "{i:in|%.txt}_world.txt")

	// // Connect network
	// world.In("in").From(hello.Out("out"))

	// // Run workflow
	// wf.Run()

	// input :=
	input, err := ioutil.ReadFile("sample.graph.json")
	if err != nil {
		log.Fatalln(err)
	}
	graph := ParseGraph(input)
	// log.Println(graph)

	modules, err := filepath.Abs("node_modules")
	if err != nil {
		log.Fatalln(err)
	}

	wf := graph.EvaluateWorkflow("foo", modules)
	wf.Run()
}
