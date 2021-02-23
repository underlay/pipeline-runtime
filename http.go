package main

import (
	"context"
	"net/http"
)

func (p *pipelinePlugin) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	if req.Method == "GET" {
		p.Get(ctx, res, req)
	} else if req.Method == "POST" {
		p.Post(ctx, res, req)
	} else {
		res.WriteHeader(405)
	}

	return
}
