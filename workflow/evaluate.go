package workflow

import (
	"fmt"
	"path/filepath"

	"github.com/scipipe/scipipe"
)

// MakeEvaluateCommand assembles the evaluate shell command
func MakeEvaluateCommand(moduleDirectory string, id ID, node *Node) string {
	path := filepath.Join(ModulePath(moduleDirectory, node.Kind), "evaluate.js")

	state := fmt.Sprintf("--state {i:%s}", StateInPort(id))

	inputSchemas := "--input-schemas"
	inputInstances := "--input-instances"
	for input := range node.Inputs {
		inputSchemas += fmt.Sprintf(" %s={i:%s}", input, SchemaInPort(id, input))
		inputInstances += fmt.Sprintf(" %s={i:%s}", input, InstanceInPort(id, input))
	}

	outputSchemas := "--output-schemas"
	outputInstances := "--output-instances"
	for output := range node.Outputs {
		outputSchemas += fmt.Sprintf(" %s={o:%s}", output, SchemaOutPort(id, output))
		outputInstances += fmt.Sprintf(" %s={o:%s}", output, InstanceOutPort(id, output))
	}

	return fmt.Sprintf(
		"node %s %s %s %s %s %s",
		path, state,
		inputSchemas, inputInstances,
		outputSchemas, outputInstances,
	)
}

// Evaluate creates and runs an evaluate workflow
func Evaluate(
	moduleDirectory string, outputDirectory string, graph *Graph,
) map[ID]string {
	failures := make(chan *Failure)
	buffer := make(chan map[ID]string)
	go bufferFailures(failures, buffer)

	wf := scipipe.NewWorkflow("evaluate", 4)

	processes := map[ID]*scipipe.Process{}
	for id, node := range graph.Nodes {
		evaluateProc := wf.NewProc(
			ProcessName(id, node),
			MakeEvaluateCommand(moduleDirectory, id, node),
		)

		evaluateProc.CustomExecute = func(task *scipipe.Task) { executeTask(id, task, failures) }

		for output := range node.Outputs {
			schemaOutPort := SchemaOutPort(id, output)
			schemaPath := SchemaOutputPath(outputDirectory, id, node, output)
			evaluateProc.SetOut(schemaOutPort, schemaPath)
			instanceOutPort := InstanceOutPort(id, output)
			instancePath := InstanceOutputPath(outputDirectory, id, node, output)
			evaluateProc.SetOut(instanceOutPort, instancePath)
		}
		processes[id] = evaluateProc

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
		evaluateProc.In(stateInPort).From(stateProc.Out(stateOutPort))
	}

	for _, edge := range graph.Edges {
		target := processes[edge.Target.ID]
		schemaInPort := SchemaInPort(edge.Target.ID, edge.Target.Input)
		instanceInPort := InstanceInPort(edge.Target.ID, edge.Target.Input)
		source := processes[edge.Source.ID]
		schemaOutPort := SchemaOutPort(edge.Source.ID, edge.Source.Output)
		instanceOutPort := InstanceOutPort(edge.Source.ID, edge.Source.Output)
		target.In(schemaInPort).From(source.Out(schemaOutPort))
		target.In(instanceInPort).From(source.Out(instanceOutPort))
	}

	wf.Run()
	close(failures)
	return <-buffer
}
