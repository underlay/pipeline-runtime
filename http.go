package main

import (
	"context"
	"net/http"
)

// Server is a local web server
type Server struct {
	moduleDirectory string
}

// ServeHTTP handles HTTP requests using the database and core API
func (server *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	if req.Method == "GET" {
		server.Get(ctx, res, req)
	} else if req.Method == "POST" {
		server.Post(ctx, res, req)
	} else {
		res.WriteHeader(405)
	}

	return
}
