package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"path/filepath"
	"regexp"
	"strconv"
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

	metricsDir := filepath.Join(path, "metrics_results")
	metricsFiles, err := ioutil.ReadDir(metricsDir)
	if err != nil {
		log.Printf("reading distorted images direcory error: %v", err)
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

		distorted := Distortion{
			Path:            filepath.Join(distortionsDir, fi.Name()),
			ProvidedMetrics: make(Metrics, len(metricsFiles)),
		}
		dataset[refIndex].Distorted = append(dataset[refIndex].Distorted, distorted)
	}
	for _, fi := range metricsFiles {
		if fi.IsDir() {
			log.Printf("\tis dir %s", fi.Name())
			continue
		}

		if strings.ToLower(filepath.Ext(fi.Name())) != ".txt" {
			log.Printf("\tno .txt file %s", fi.Name())
			continue
		}

		metricsName := strings.TrimSuffix(fi.Name(), filepath.Ext(fi.Name()))

		metricsFilePath := filepath.Join(metricsDir, fi.Name())
		content, err := ioutil.ReadFile(metricsFilePath)
		if err != nil {
			log.Printf("\treading metrics file error: %v", err)
			continue
		}

		lines := strings.Split(string(content), "\r\n")

		for i, line := range lines {
			if line == "" {
				break
			}

			value, err := strconv.ParseFloat(line, 64)
			if err != nil {
				log.Printf("\tconverting %s metrics value %s to float64 error: %v", metricsName, line, err)
				value = math.NaN()
			}
			ri, di := i/len(dataset[0].Distorted), i%len(dataset[0].Distorted)
			dataset[ri].Distorted[di].ProvidedMetrics[metricsName] = value
		}
	}

	mosFileContent, err := ioutil.ReadFile(filepath.Join(path, "mos.txt"))
	if err != nil {
		log.Printf("\treading mos metrics file error: %v", err)
	} else {
		metricsName := "mos"
		lines := strings.Split(string(mosFileContent), "\r\n")
		for i, line := range lines {
			if line == "" {
				break
			}
			value, err := strconv.ParseFloat(line, 64)
			if err != nil {
				log.Printf("\tconverting %s metrics value %s to float64 error: %v", metricsName, line, err)
				value = math.NaN()
			}
			ri, di := i/len(dataset[0].Distorted), i%len(dataset[0].Distorted)
			dataset[ri].Distorted[di].ProvidedMetrics[metricsName] = value
		}
	}

	mosStdFileContent, err := ioutil.ReadFile(filepath.Join(path, "mos.txt"))
	if err != nil {
		log.Printf("\treading mos metrics file error: %v", err)
	} else {
		metricsName := "mos_std"
		lines := strings.Split(string(mosStdFileContent), "\r\n")
		for i, line := range lines {
			if line == "" {
				break
			}
			value, err := strconv.ParseFloat(line, 64)
			if err != nil {
				log.Printf("\tconverting %s metrics value %s to float64 error: %v", metricsName, line, err)
				value = math.NaN()
			}
			ri, di := i/len(dataset[0].Distorted), i%len(dataset[0].Distorted)
			dataset[ri].Distorted[di].ProvidedMetrics[metricsName] = value
		}
	}
	return dataset, nil
}
