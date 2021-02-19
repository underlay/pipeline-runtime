package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
)

const validatePath = "/api/v0/validate"
const evaluatePath = "/api/v0/evaluate"

// Post handles HTTP POST requests
func (server *Server) Post(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	if req.URL.Path != validatePath && req.URL.Path != evaluatePath {
		res.WriteHeader(404)
		return
	}

	mediaType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		res.WriteHeader(400)
		return
	} else if mediaType != "application/json" {
		res.WriteHeader(415)
		return
	}

	var graph Graph
	if err := json.NewDecoder(req.Body).Decode(&graph); err != nil {
		res.WriteHeader(400)
		return
	}

	outputDirectory, err := ioutil.TempDir(".", "workflow-")
	if err != nil {
		res.WriteHeader(500)
		return
	}

	defer os.RemoveAll(outputDirectory)

	var failures map[string]string
	if req.URL.Path == validatePath {
		failures = Validate(server.moduleDirectory, outputDirectory, &graph)
	} else if req.URL.Path == evaluatePath {
		failures = Evaluate(server.moduleDirectory, outputDirectory, &graph)
	}

	if failures == nil {
		res.WriteHeader(500)
	} else {
		res.WriteHeader(200)
		_ = json.NewEncoder(res).Encode(failures)
	}
}
