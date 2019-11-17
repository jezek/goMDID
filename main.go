package main

import (
	"fmt"
	"log"
)

// MDID dataset (https://www.sz.tsinghua.edu.cn/labs/vipl/mdid.html) image similarity metrics, rewritten to go (golang)

// Loads MDID dataset from directory, performs metrics and writes output.
func main() {
	// Load dataset from diretory.
	dataset, error := LoadMDID("dataset/MDID")
	if error != nil {
		log.Fatalf("Error loading MDID dataset: %v\n", error)
	}

	//TODO Compute metrics.

	// Print dataset
	fmt.Printf("%v\n", dataset)

}
