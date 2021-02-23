package workflow

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/scipipe/scipipe"
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
func ProcessName(id ID, node *Node) string { return fmt.Sprintf("%s-%s", node.Kind, id) }

// StateProcessName uniquely identifies a state node process within a given workflow
func StateProcessName(id ID, node *Node) string { return fmt.Sprintf("%s-%s-state", node.Kind, id) }

// StateInPort identifies a node state input port
func StateInPort(id ID) string { return fmt.Sprintf("%s-state-in", id) }

// StateOutPort identifieds a node state output port
func StateOutPort(id ID) string { return fmt.Sprintf("%s-state-out", id) }

// StateOutputPath is the state file path
func StateOutputPath(outputDirectory string, id ID, node *Node) string {
	return filepath.Join(
		outputDirectory,
		fmt.Sprintf("%s-%s.state.json", node.Kind, id),
	)
}

// SchemaInPort identifies a schema input port
func SchemaInPort(id ID, input string) string {
	return fmt.Sprintf("%s-%s.input.schema", id, input)
}

// SchemaOutPort identifies a schema output port
func SchemaOutPort(id ID, output string) string {
	return fmt.Sprintf("%s-%s.output.schema", id, output)
}

// SchemaOutputPath is the output schema file path
func SchemaOutputPath(outputDirectory string, id ID, node *Node, output string) string {
	return filepath.Join(
		outputDirectory,
		fmt.Sprintf("%s-%s-%s.schema", node.Kind, id, output),
	)
}

// InstanceInPort identifies an instance input port
func InstanceInPort(id ID, input string) string {
	return fmt.Sprintf("%s-%s.input.instance", id, input)
}

// InstanceOutPort identifies an instance output port
func InstanceOutPort(id ID, output string) string {
	return fmt.Sprintf("%s-%s.output.instance", id, output)
}

// InstanceOutputPath is the output instance file path
func InstanceOutputPath(moduleDirectory string, id ID, node *Node, output string) string {
	return filepath.Join(
		moduleDirectory,
		fmt.Sprintf("%s-%s-%s.instance", node.Kind, id, output),
	)
}

func bufferFailures(failures chan *Failure, buffer chan map[ID]string) {
	b := map[ID]string{}
	for f := range failures {
		b[f.id] = f.output
	}
	buffer <- b
}

func executeTask(id ID, task *scipipe.Task, failures chan *Failure) {
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
		failures <- &Failure{id, string(out)}
	default:
		scipipe.Failf(
			"Could not run command.\nCommand: %s\nOutput: \n%s\nError: %s\n",
			task.Command, string(out), err.Error(),
		)
	}
}
