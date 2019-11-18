package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Dataset type holds more reference images for metrics.
type Dataset []Reference

func (d Dataset) String() string {
	strs := []string{}
	for i, ref := range d {
		strs = append(strs, fmt.Sprintf("%2d. %v", i+1, ref))
	}
	return strings.Join(strs, "\n")
}

// Reference type holds path to reference image and more distorted images with their metrics.
type Reference struct {
	Path      string
	Distorted []Distortion
}

func (r Reference) String() string {
	lines := []string{fmt.Sprintf("%v", filepath.Base(r.Path))}

	for i, dis := range r.Distorted {
		lines = append(lines, fmt.Sprintf("\t%2d. %v", i+1, dis))
	}
	return strings.Join(lines, "\n")
}

// Distortion holds a path to distorted image, its dissortion info and metrics against it's parent reference image.
type Distortion struct {
	Path            string
	DissortionsInfo string
	ProvidedMetrics Metrics
	ComputedMetrics Metrics
}

func (d Distortion) String() string {
	lines := []string{fmt.Sprintf("%v", filepath.Base(d.Path))}

	pmlines := []string(nil)
	if len(d.ProvidedMetrics) > 0 {
		pmlines = append(pmlines, "\t\tProvided metrics:")
		for m, v := range d.ProvidedMetrics {
			pmlines = append(pmlines, fmt.Sprintf("\t\t%8s: %f", m, v))
		}
		lines = append(lines, pmlines...)
	}
	return strings.Join(lines, "\n")
}

// Metrics (plural) is an map containing name => value pairs.
type Metrics map[string]float64
