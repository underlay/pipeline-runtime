package main

import (
	"context"
	"net/http"
)

func (p *pipelinePlugin) Get(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(501)
}
