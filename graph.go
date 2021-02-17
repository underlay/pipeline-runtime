package main

import (
	"encoding/json"
	"fmt"
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

// Module returns the path to the node kind's nodejs module
func (node *Node) Module(modules string) string {
	return filepath.Join(modules, "@underlay", fmt.Sprintf("block-%s-runtime", node.Kind))
}

// ProcessName uniquely identifies a node process within a given workflow
func (node *Node) ProcessName() string { return fmt.Sprintf("node-%d", node.ID) }

// StateProcessName uniquely identifies a state node process within a given workflow
func (node *Node) StateProcessName() string { return fmt.Sprintf("node-%d-state", node.ID) }

// StateInPort identifies the node's state in port
func (node *Node) StateInPort() string { return fmt.Sprintf("%d-state-in", node.ID) }

// StateOutPort identifieds the node's state out port
func (node *Node) StateOutPort() string { return fmt.Sprintf("%d-state-out", node.ID) }

// StateOutputPath is the state file path
func (node *Node) StateOutputPath() string { return fmt.Sprintf("node-%d.state.json", node.ID) }

// SchemaOutputPath is the output schema file path
func (node *Node) SchemaOutputPath(output string) string {
	return fmt.Sprintf("node-%d-%s.schema", node.ID, output)
}

// InstanceOutputPath is the output instance file path
func (node *Node) InstanceOutputPath(output string) string {
	return fmt.Sprintf("node-%d-%s.instance", node.ID, output)
}

// ValidateCommand assembles the validate shell command
func (node *Node) ValidateCommand(modules string) string {
	path := filepath.Join(node.Module(modules), "lib", "validate.js")

	state := fmt.Sprintf("--state {i:%s}", node.StateInPort())

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

// EvaluateCommand assembles the evaluate shell command
func (node *Node) EvaluateCommand(workflow string, modules string) string {
	path := filepath.Join(node.Module(modules), "lib", "evaluate.js")

	state := fmt.Sprintf("--state {i:%s}", node.StateInPort())

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

// ParseGraph parses a graph out of serialized JSON
func ParseGraph(input []byte) *Graph {
	var graph Graph
	json.Unmarshal(input, &graph)
	return &graph
}

// ValidateWorkflow creates a validate workflow
func (graph *Graph) ValidateWorkflow(workflow string, modules string) *scipipe.Workflow {
	wf := scipipe.NewWorkflow(workflow, 4)

	processes := map[ID]*scipipe.Process{}
	for _, node := range graph.Nodes {
		validateProc := wf.NewProc(
			node.ProcessName(),
			node.ValidateCommand(modules),
		)
		for output := range node.Outputs {
			schemaOutPort := SchemaOutPort(node.ID, output)
			schemaPath := node.SchemaOutputPath(output)
			validateProc.SetOut(schemaOutPort, schemaPath)
		}
		processes[node.ID] = validateProc

		stateOutPort := node.StateOutPort()
		stateProc := wf.NewProc(
			node.StateProcessName(),
			fmt.Sprintf("{o:%s}", stateOutPort),
		)
		statePath := node.StateOutputPath()
		stateProc.SetOut(stateOutPort, statePath)
		stateValue := node.State
		stateProc.CustomExecute = func(task *scipipe.Task) { task.OutIP(stateOutPort).Write(stateValue) }

		stateInPort := node.StateInPort()
		validateProc.In(stateInPort).From(stateProc.Out(stateOutPort))
	}

	for _, edge := range graph.Edges {
		target := processes[edge.Target.ID]
		schemaInPort := SchemaInPort(edge.Target.ID, edge.Target.Input)
		source := processes[edge.Source.ID]
		schemaOutPort := SchemaOutPort(edge.Source.ID, edge.Source.Output)
		target.In(schemaInPort).From(source.Out(schemaOutPort))
	}

	return wf
}

// EvaluateWorkflow creates an evaluate workflow
func (graph *Graph) EvaluateWorkflow(workflow string, modules string) *scipipe.Workflow {
	wf := scipipe.NewWorkflow(workflow, 4)

	processes := map[ID]*scipipe.Process{}
	for _, node := range graph.Nodes {
		evaluateProc := wf.NewProc(
			node.ProcessName(),
			node.EvaluateCommand(workflow, modules),
		)

		for output := range node.Outputs {
			schemaOutPort := SchemaOutPort(node.ID, output)
			schemaPath := node.SchemaOutputPath(output)
			evaluateProc.SetOut(schemaOutPort, schemaPath)
			instanceOutPort := InstanceOutPort(node.ID, output)
			instancePath := node.InstanceOutputPath(output)
			evaluateProc.SetOut(instanceOutPort, instancePath)
		}
		processes[node.ID] = evaluateProc

		stateOutPort := node.StateOutPort()
		stateProc := wf.NewProc(
			node.StateProcessName(),
			fmt.Sprintf("{o:%s}", stateOutPort),
		)
		statePath := node.StateOutputPath()
		stateProc.SetOut(stateOutPort, statePath)
		stateValue := node.State
		stateProc.CustomExecute = func(task *scipipe.Task) { task.OutIP(stateOutPort).Write(stateValue) }

		stateInPort := node.StateInPort()
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

	return wf
}

// SchemaInPort identifies a schema input port
func SchemaInPort(id ID, input string) string {
	return fmt.Sprintf("%d-%s.input.schema", id, input)
}

// InstanceInPort identifies an instance input port
func InstanceInPort(id ID, input string) string {
	return fmt.Sprintf("%d-%s.input.instance", id, input)
}

// SchemaOutPort identifies a schema output port
func SchemaOutPort(id ID, output string) string {
	return fmt.Sprintf("%d-%s.output.schema", id, output)
}

// InstanceOutPort identifies an instance output port
func InstanceOutPort(id ID, output string) string {
	return fmt.Sprintf("%d-%s.output.instance", id, output)
}
