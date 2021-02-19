package main

import (
	"context"
	"net/http"
)

// Get handles HTTP GET requests
func (server *Server) Get(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(501)
}
