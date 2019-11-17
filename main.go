package main

import (
	"fmt"
	"log"
)

// MDID dataset (https://www.sz.tsinghua.edu.cn/labs/vipl/mdid.html) image similarity metrics, rewritten to go (golang)

// Loads MDID dataset from directory, performs metrics and writes output.
func main() {
	datasetDir := "dataset/MDID"
	// Load dataset from diretory.
	dataset, err := LoadMDID(datasetDir)
	if err != nil {
		log.Fatalf("Loading MDID dataset from \"%s\" error: %v\n", datasetDir, err)
	}

	//TODO Compute metrics.

	// Print dataset
	fmt.Printf("%v\n", dataset)

}
