package main

import (
	"log"
	"net/http"
	"path/filepath"
)

func main() {
	moduleDirectory, err := filepath.Abs("node_modules")
	if err != nil {
		log.Fatalln(err)
	}

	server := &Server{moduleDirectory: moduleDirectory}
	log.Println("http://localhost:8086")
	log.Fatal(http.ListenAndServe(":8086", server))
}
