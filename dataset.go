package main

import "fmt"

// Dataset type holds more reference images for metrics.
type Dataset []Reference

func (d Dataset) String() string {
	str := ""
	for i, ref := range d {
		str += fmt.Sprintf("%2d. %v\n", i+1, ref)
	}
	return str
}

// Reference type holds path to reference image and more distorted images with their metrics.
type Reference struct {
	Path      string
	Distorted []Distortion
}

func (r Reference) String() string {
	str := fmt.Sprintf("%v\n", r.Path)
	for i, ref := range r.Distorted {
		str += fmt.Sprintf("\t%2d. %v\n", i+1, ref)
	}
	return str
}

// Distortion holds a path to distorted image, its dissortion info and metrics against it's parent reference image.
type Distortion struct {
	Path            string
	DissortionsInfo string
	ProvidedMetrics Metrics
	ComputedMetrics Metrics
}

// Metrics (plural) is an map containing name => value pairs.
type Metrics map[string]float64
