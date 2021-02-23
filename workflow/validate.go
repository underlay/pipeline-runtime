package workflow

import (
	"fmt"
	"path/filepath"

	"github.com/scipipe/scipipe"
)

// MakeValidateCommand assembles the validate shell command
func MakeValidateCommand(moduleDirectory string, id ID, node *Node) string {
	path := filepath.Join(ModulePath(moduleDirectory, node.Kind), "validate.js")

	state := fmt.Sprintf("--state {i:%s}", StateInPort(id))

	inputSchemas := "--input-schemas"
	for input := range node.Inputs {
		inputSchemas += fmt.Sprintf(" %s={i:%s}", input, SchemaInPort(id, input))
	}

	outputSchemas := "--output-schemas"
	for output := range node.Outputs {
		outputSchemas += fmt.Sprintf(" %s={o:%s}", output, SchemaOutPort(id, output))
	}

	return fmt.Sprintf(
		"node %s %s %s %s",
		path, state,
		inputSchemas,
		outputSchemas,
	)
}

// Validate creates and runs a validate workflow
func Validate(
	moduleDirectory string, outputDirectory string, graph *Graph,
) map[ID]string {
	failures := make(chan *Failure)
	buffer := make(chan map[ID]string)
	go bufferFailures(failures, buffer)

	wf := scipipe.NewWorkflow("validate", 4)

	processes := map[ID]*scipipe.Process{}
	for id, node := range graph.Nodes {
		validateProc := wf.NewProc(
			ProcessName(id, node),
			MakeValidateCommand(moduleDirectory, id, node),
		)

		validateProc.CustomExecute = func(task *scipipe.Task) { executeTask(id, task, failures) }

		for output := range node.Outputs {
			schemaOutPort := SchemaOutPort(id, output)
			schemaPath := SchemaOutputPath(outputDirectory, id, node, output)
			validateProc.SetOut(schemaOutPort, schemaPath)
		}

		processes[id] = validateProc

		stateOutPort := StateOutPort(id)
		stateProc := wf.NewProc(
			StateProcessName(id, node),
			fmt.Sprintf("{o:%s}", stateOutPort),
		)
		statePath := StateOutputPath(outputDirectory, id, node)
		stateProc.SetOut(stateOutPort, statePath)

		stateValue := node.State
		stateProc.CustomExecute = func(task *scipipe.Task) { task.OutIP(stateOutPort).Write(stateValue) }

		stateInPort := StateInPort(id)
		validateProc.In(stateInPort).From(stateProc.Out(stateOutPort))
	}

	for _, edge := range graph.Edges {
		target := processes[edge.Target.ID]
		schemaInPort := SchemaInPort(edge.Target.ID, edge.Target.Input)
		source := processes[edge.Source.ID]
		schemaOutPort := SchemaOutPort(edge.Source.ID, edge.Source.Output)
		target.In(schemaInPort).From(source.Out(schemaOutPort))
	}

	wf.Run()
	close(failures)
	return <-buffer
}
