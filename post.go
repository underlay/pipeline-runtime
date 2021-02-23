package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"

	tar "github.com/ipfs/go-ipfs/tar"

	workflow "github.com/underlay/pipeline-runtime/workflow"
)

const collectionPath = "/api/v0/collection"
const validatePath = "/api/v0/validate"
const evaluatePath = "/api/v0/evaluate"

func (p *pipelinePlugin) Post(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	mediaType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte(err.Error()))
		return
	}

	if req.URL.Path == "/" {
		outFile, err := os.Create("collection.tar")
		defer outFile.Close()
		if err != nil {
			res.WriteHeader(500)
		} else if _, err = io.Copy(outFile, req.Body); err != nil {
			res.WriteHeader(500)
		} else {
			res.WriteHeader(204)
		}

		return
	} else if req.URL.Path == collectionPath {
		if err := req.ParseForm(); err != nil {
			res.WriteHeader(400)
			return
		}

		host, err := url.QueryUnescape(req.Form.Get("host"))
		if err != nil {
			res.WriteHeader(400)
			return
		}

		id, err := url.QueryUnescape(req.Form.Get("id"))
		if err != nil {
			res.WriteHeader(400)
			return
		}

		if mediaType != "application/x-tar" {
			res.WriteHeader(415)
			return
		}

		var reader io.Reader = req.Body
		if req.Header.Get("Content-Encoding") == "gzip" {
			reader, err = gzip.NewReader(req.Body)
			if err != nil {
				res.WriteHeader(500)
				return
			}
		}

		node, err := tar.ImportTar(ctx, reader, p.ipfs.DAG)
		if err != nil {
			res.WriteHeader(502)
			return
		}

		cid := node.Cid()

		err = p.ipfs.DAG.Add(ctx, node)
		if err != nil {
			res.WriteHeader(502)
			return
		}

		log.Println("host", host, "id", id)
		log.Println("PINNED A CID@!!!!", cid.String())

		res.WriteHeader(204)
		return
	} else if req.URL.Path == validatePath || req.URL.Path == evaluatePath {
		if mediaType != "application/json" {
			res.WriteHeader(415)
			return
		}

		var graph workflow.Graph
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
			failures = workflow.Validate(p.moduleDirectory, outputDirectory, &graph)
		} else if req.URL.Path == evaluatePath {
			failures = workflow.Evaluate(p.moduleDirectory, outputDirectory, &graph)
		}

		if failures == nil {
			res.WriteHeader(500)
		} else {
			res.WriteHeader(200)
			_ = json.NewEncoder(res).Encode(failures)
		}
	} else {
		res.WriteHeader(404)
		return
	}
}
