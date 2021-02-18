package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/scipipe/scipipe"
)

// ID is used for Node and Edge identifiers
type ID = uint64

// Node represents an instance of a block
type Node struct {
	ID      ID
	Kind    string
	State   json.RawMessage
	Inputs  map[string]interface{}
	Outputs map[string][]ID
}

// MakeValidateCommand assembles the validate shell command
func (node *Node) MakeValidateCommand(moduleDirectory string) string {
	path := filepath.Join(ModulePath(moduleDirectory, node.Kind), "validate.js")

	state := fmt.Sprintf("--state {i:%s}", StateInPort(node.ID))

	inputSchemas := "--input-schemas"
	for input := range node.Inputs {
		inputSchemas += fmt.Sprintf(" %s={i:%s}", input, SchemaInPort(node.ID, input))
	}

	outputSchemas := "--output-schemas"
	for output := range node.Outputs {
		outputSchemas += fmt.Sprintf(" %s={o:%s}", output, SchemaOutPort(node.ID, output))
	}

	return fmt.Sprintf(
		"node %s %s %s %s",
		path, state,
		inputSchemas,
		outputSchemas,
	)
}

// MakeEvaluateCommand assembles the evaluate shell command
func (node *Node) MakeEvaluateCommand(moduleDirectory string) string {
	path := filepath.Join(ModulePath(moduleDirectory, node.Kind), "evaluate.js")

	state := fmt.Sprintf("--state {i:%s}", StateInPort(node.ID))

	inputSchemas := "--input-schemas"
	inputInstances := "--input-instances"
	for input := range node.Inputs {
		inputSchemas += fmt.Sprintf(" %s={i:%s}", input, SchemaInPort(node.ID, input))
		inputInstances += fmt.Sprintf(" %s={i:%s}", input, InstanceInPort(node.ID, input))
	}

	outputSchemas := "--output-schemas"
	outputInstances := "--output-instances"
	for output := range node.Outputs {
		outputSchemas += fmt.Sprintf(" %s={o:%s}", output, SchemaOutPort(node.ID, output))
		outputInstances += fmt.Sprintf(" %s={o:%s}", output, InstanceOutPort(node.ID, output))
	}

	return fmt.Sprintf(
		"node %s %s %s %s %s %s",
		path, state,
		inputSchemas, inputInstances,
		outputSchemas, outputInstances,
	)
}

// Edge represents a connection between nodes
type Edge struct {
	Source struct {
		ID     ID
		Output string
	}
	Target struct {
		ID    ID
		Input string
	}
}

// Graph represents a graph of nodes and edges
type Graph struct {
	Nodes []*Node
	Edges []*Edge
}

// Validate creates and runs a validate workflow
func (graph *Graph) Validate(
	moduleDirectory string, outputDirectory string,
) []*Failure {
	failures := make(chan *Failure)
	buffer := make(chan []*Failure)
	go bufferFailures(failures, buffer)

	wf := scipipe.NewWorkflow("validate", 4)

	processes := map[ID]*scipipe.Process{}
	for _, node := range graph.Nodes {
		validateProc := wf.NewProc(
			ProcessName(node.ID),
			node.MakeValidateCommand(moduleDirectory),
		)

		validateProc.CustomExecute = func(task *scipipe.Task) { executeTask(task, failures) }

		for output := range node.Outputs {
			schemaOutPort := SchemaOutPort(node.ID, output)
			schemaPath := SchemaOutputPath(outputDirectory, node.ID, output)
			validateProc.SetOut(schemaOutPort, schemaPath)
		}
		processes[node.ID] = validateProc

		stateOutPort := StateOutPort(node.ID)
		stateProc := wf.NewProc(
			StateProcessName(node.ID),
			fmt.Sprintf("{o:%s}", stateOutPort),
		)
		statePath := StateOutputPath(outputDirectory, node.ID)

		stateProc.SetOut(stateOutPort, statePath)
		stateValue := node.State
		stateProc.CustomExecute = func(task *scipipe.Task) { task.OutIP(stateOutPort).Write(stateValue) }

		stateInPort := StateInPort(node.ID)
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

// Evaluate creates and runs an evaluate workflow
func (graph *Graph) Evaluate(
	moduleDirectory string, outputDirectory string,
) []*Failure {
	failures := make(chan *Failure)
	buffer := make(chan []*Failure)
	go bufferFailures(failures, buffer)

	wf := scipipe.NewWorkflow("evaluate", 4)

	processes := map[ID]*scipipe.Process{}
	for _, node := range graph.Nodes {
		evaluateProc := wf.NewProc(
			ProcessName(node.ID),
			node.MakeEvaluateCommand(moduleDirectory),
		)

		evaluateProc.CustomExecute = func(task *scipipe.Task) { executeTask(task, failures) }

		for output := range node.Outputs {
			schemaOutPort := SchemaOutPort(node.ID, output)
			schemaPath := SchemaOutputPath(outputDirectory, node.ID, output)
			evaluateProc.SetOut(schemaOutPort, schemaPath)
			instanceOutPort := InstanceOutPort(node.ID, output)
			instancePath := InstanceOutputPath(outputDirectory, node.ID, output)
			evaluateProc.SetOut(instanceOutPort, instancePath)
		}
		processes[node.ID] = evaluateProc

		stateOutPort := StateOutPort(node.ID)
		stateProc := wf.NewProc(
			StateProcessName(node.ID),
			fmt.Sprintf("{o:%s}", stateOutPort),
		)
		statePath := StateOutputPath(outputDirectory, node.ID)
		stateProc.SetOut(stateOutPort, statePath)
		stateValue := node.State
		stateProc.CustomExecute = func(task *scipipe.Task) { task.OutIP(stateOutPort).Write(stateValue) }

		stateInPort := StateInPort(node.ID)
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

// Failure represents a failed command
type Failure struct {
	err    error
	name   string
	output string
}

func executeTask(task *scipipe.Task, failures chan *Failure) {
	// If any of the input files don't exist, just silently continue
	for input, inputIP := range task.InIPs {
		if _, err := os.Stat(inputIP.Path()); os.IsNotExist(err) {
			scipipe.LogAuditf(task.Name, "Skipping: missing input %s", input)
			return
		} else if err != nil {
			scipipe.Fail(err.Error())
		}
	}

	cmd := fmt.Sprintf("cd %s && %s && cd ..", task.TempDir(), task.Command)
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	switch err := err.(type) {
	case nil:
	case *exec.ExitError:
		scipipe.Error.Printf("Command failed: %s\n\n%s\n", err.Error(), string(out))
		failures <- &Failure{err, task.Name, string(out)}
	default:
		scipipe.Failf(
			"Could not run command.\nCommand: %s\nOutput: \n%s\nError: %s\n",
			task.Command, string(out), err.Error(),
		)
	}
}

func bufferFailures(failures chan *Failure, buffer chan []*Failure) {
	b := []*Failure{}
	for f := range failures {
		b = append(b, f)
	}
	buffer <- b
}
