package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strings"
)

func LoadMDID(path string) (Dataset, error) {
	log.Printf("LoadMDID(%v)", path)

	referencesDir := filepath.Join(path, "reference_images")
	referenceFiles, err := ioutil.ReadDir(referencesDir)
	if err != nil {
		return nil, fmt.Errorf("reading reference images direcory error: %w", err)
	}

	distortionsDir := filepath.Join(path, "distortion_images")
	distortedFiles, err := ioutil.ReadDir(distortionsDir)
	if err != nil {
		return nil, fmt.Errorf("reading distorted images direcory error: %w", err)
	}

	refRegexp, err := regexp.Compile(`^img\d{2}`)
	if err != nil {
		return nil, fmt.Errorf("reference from distortion regexp compile error: %w", err)
	}

	dataset := make(Dataset, 0, len(referenceFiles))
	refMap := map[string]int{}

	log.Printf("\tloading reference images from %s", referencesDir)
	for _, fi := range referenceFiles {
		if fi.IsDir() {
			log.Printf("\tis dir %s", fi.Name())
			continue
		}

		if strings.ToLower(filepath.Ext(fi.Name())) != ".bmp" {
			log.Printf("\tno .bmp file %s", fi.Name())
			continue
		}

		reference := Reference{
			Path:      filepath.Join(referencesDir, fi.Name()),
			Distorted: make([]Distortion, 0, len(distortedFiles)/len(referenceFiles)),
		}

		dataset = append(dataset, reference)
		refMap[strings.TrimSuffix(fi.Name(), filepath.Ext(fi.Name()))] = len(dataset) - 1
	}

	log.Printf("\tloading distortion images from %s", distortionsDir)
	for _, fi := range distortedFiles {
		if fi.IsDir() {
			log.Printf("\tis dir %s", fi.Name())
			continue
		}

		if strings.ToLower(filepath.Ext(fi.Name())) != ".bmp" {
			log.Printf("\tno .bmp file %s", fi.Name())
			continue
		}

		referenceImgName := refRegexp.FindString(fi.Name())
		if referenceImgName == "" {
			log.Printf("\tcan not extract reference name from %s", fi.Name())
			continue
		}

		refIndex, ok := refMap[referenceImgName]
		if !ok {
			log.Printf("\tno reference image for %s", referenceImgName)
			continue
		}

		distorted := Distortion{Path: filepath.Join(distortionsDir, fi.Name())}
		dataset[refIndex].Distorted = append(dataset[refIndex].Distorted, distorted)
	}
	return dataset, nil
}
