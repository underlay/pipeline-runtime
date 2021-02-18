package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalln("Missing arguments")
	}

	inputFile, err := filepath.Abs(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}

	input, err := ioutil.ReadFile(inputFile)
	if err != nil {
		log.Fatalln(err)
	}

	var graph Graph
	err = json.Unmarshal(input, &graph)
	if err != nil {
		log.Fatalln(err)
	}

	moduleDirectory, err := filepath.Abs("node_modules")
	if err != nil {
		log.Fatalln(err)
	}

	if outputDirectory, err := filepath.Abs(os.Args[2]); err != nil {
		log.Fatalln(err)
	} else if stat, err := os.Stat(outputDirectory); os.IsNotExist(err) {
	} else if err != nil {
		log.Fatalln(err)
	} else if stat.IsDir() {
	} else {
		log.Fatalln("Output path must be a directory")
	}

	failures := graph.Evaluate(moduleDirectory, os.Args[2])

	log.Println("got the array in main :-)", failures)
	for _, failure := range failures {
		log.Println(failure.name)
	}
}
