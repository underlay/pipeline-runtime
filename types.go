package main

import (
	"encoding/json"
)

// ID is used for Node and Edge identifiers
type ID = string

// Node represents an instance of a block
type Node struct {
	Kind    string
	State   json.RawMessage
	Inputs  map[string]interface{}
	Outputs map[string][]ID
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
	Nodes map[string]*Node
	Edges map[string]*Edge
}

// Failure represents a failed command
type Failure struct {
	id     ID
	output string
}
