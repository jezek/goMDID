package main

// Dataset type holds more reference images for metrics.
type Dataset []Reference

// Refernce type holds path to reference image and more distorted images with their metrics.
type Reference struct {
	Path      string
	Distorged []Distortion
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
