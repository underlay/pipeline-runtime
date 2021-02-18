package main

import (
	"fmt"
	"path/filepath"
)

var failures = make(chan string)

// ModulePath returns the path to the node kind's nodejs module
func ModulePath(moduleDirectory string, kind string) string {
	return filepath.Join(
		moduleDirectory,
		"@underlay",
		"pipeline-runtime",
		"lib",
		kind,
	)
}

// ProcessName uniquely identifies a node process within a given workflow
func ProcessName(id ID) string { return fmt.Sprintf("node-%d", id) }

// StateProcessName uniquely identifies a state node process within a given workflow
func StateProcessName(id ID) string { return fmt.Sprintf("node-%d-state", id) }

// StateInPort identifies a node state input port
func StateInPort(id ID) string { return fmt.Sprintf("%d-state-in", id) }

// StateOutPort identifieds a node state output port
func StateOutPort(id ID) string { return fmt.Sprintf("%d-state-out", id) }

// StateOutputPath is the state file path
func StateOutputPath(outputDirectory string, id ID) string {
	return filepath.Join(
		outputDirectory,
		fmt.Sprintf("%d.state.json", id),
	)
}

// SchemaInPort identifies a schema input port
func SchemaInPort(id ID, input string) string {
	return fmt.Sprintf("%d-%s.input.schema", id, input)
}

// SchemaOutPort identifies a schema output port
func SchemaOutPort(id ID, output string) string {
	return fmt.Sprintf("%d-%s.output.schema", id, output)
}

// SchemaOutputPath is the output schema file path
func SchemaOutputPath(outputDirectory string, id ID, output string) string {
	return filepath.Join(
		outputDirectory,
		fmt.Sprintf("%d-%s.schema", id, output),
	)
}

// InstanceInPort identifies an instance input port
func InstanceInPort(id ID, input string) string {
	return fmt.Sprintf("%d-%s.input.instance", id, input)
}

// InstanceOutPort identifies an instance output port
func InstanceOutPort(id ID, output string) string {
	return fmt.Sprintf("%d-%s.output.instance", id, output)
}

// InstanceOutputPath is the output instance file path
func InstanceOutputPath(moduleDirectory string, id ID, output string) string {
	return filepath.Join(
		moduleDirectory,
		fmt.Sprintf("%d-%s.instance", id, output),
	)
}
