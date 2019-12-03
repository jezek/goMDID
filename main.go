package main

import (
	"fmt"
	"image"
	"log"
	"math"
	"os"
	"path/filepath"

	_ "golang.org/x/image/bmp"
)

// MDID dataset (https://www.sz.tsinghua.edu.cn/labs/vipl/mdid.html) image similarity metrics, rewritten to go (golang)

// Loads MDID dataset from directory, performs metrics and writes output.
func main() {
	log.SetFlags(0)
	log.SetPrefix("log: ")
	datasetDir := "dataset/MDID"
	// Load dataset from diretory.
	dataset, err := LoadMDID(datasetDir)
	if err != nil {
		log.Fatalf("Loading MDID dataset from \"%s\" error: %v\n", datasetDir, err)
	}

	dataset = dataset[:2]
	dataset[0].Distorted = dataset[0].Distorted[:10]
	dataset[1].Distorted = dataset[1].Distorted[:10]
	// Print provided dataset evaluations.
	//fmt.Printf("%v\n", dataset)
	evaluators := map[string]func([]float64, []float64) float64{
		"SROCC":   SROCC,
		"SROCCos": SROCConlinestats,
		"SROCCgs": SROCCgostats,
		"KROCC":   KROCC,
		"KROCCgn": KROCCgonum,
		"KROCCgs": KROCCgostats,
		"PLCC":    PLCC,
		"PLCCgn":  PLCCgonum,
		"PLCCgs":  PLCCgostats,
		"RMSE":    RMSE,
	}

	evaluatorsList := []string{"SROCC", "KROCC", "PLCC", "RMSE"}
	providedMetricsList := []string{"PSNR", "SSIM", "VIF", "IWSSIM", "FSIMc", "GMSD"}
	fmt.Println("Comparing MDID MOS to provided metrics (pm) rankings using different evaluators (ev):")
	fmt.Println()
	fmt.Printf("%10s", "pm\\ev")
	for _, em := range evaluatorsList {
		fmt.Printf("%10s", em)
	}
	fmt.Println()

	mos := dataset.ProvidedMetricsByName("mos")
	for _, pm := range providedMetricsList {
		fmt.Printf("%10s", pm)
		m := dataset.ProvidedMetricsByName(pm)
		for _, em := range evaluatorsList {
			fmt.Printf("%10f", math.Abs(evaluators[em](mos, m)))
		}
		fmt.Println()
	}

	//s := []float64{0.1, 0.2, 0.2, 0.3, 0.4, 0.4, 0.4, 0.7}
	//fmt.Printf("Ranking %v [StandartCompetition]: %v\n", s, Rank(s, StandartCompetition))
	//fmt.Printf("Ranking %v [ModifiedCompetition]: %v\n", s, Rank(s, ModifiedCompetition))
	//fmt.Printf("Ranking %v [Ordinal]            : %v\n", s, Rank(s, Ordinal))
	//fmt.Printf("Ranking %v [Dense]              : %v\n", s, Rank(s, Dense))
	//fmt.Printf("Ranking %v [Fractional]         : %v\n", s, Rank(s, Fractional))

	// Compute metrics.
	imageFromPath := func(filepath string) (image.Image, error) {
		f, err := os.Open(filepath)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		img, _, err := image.Decode(f)
		if err != nil {
			return nil, err
		}

		return img, nil
	}

	metrics := map[string]func(image.Image, image.Image) float64{
		"MSEg":  MSE,
		"PSNRg": PSNR,
		"MSE":   MSErgb,
		"PSNR":  PSNRrgb,
		"SSIM":  SSIM,
	}
	fmt.Println()
	computeMetricsList := []string{"PSNRg", "PSNR", "SSIM"}
	for _, ref := range dataset {
		refImg, err := imageFromPath(ref.Path)
		if err != nil {
			log.Printf("Could not load image: %v", err)
			continue
		}

		for _, dis := range ref.Distorted {
			disImg, err := imageFromPath(dis.Path)
			if err != nil {
				log.Printf("Could not load image: %v", err)
				continue
			}

			for _, m := range computeMetricsList {
				dis.ComputedMetrics[m] = metrics[m](refImg, disImg)
				fmt.Printf("\rReference: %10s, Distorted: %10s, Metrics: %6s", filepath.Base(ref.Path), filepath.Base(dis.Path), m)
			}
		}
	}
	fmt.Printf("\rMetrics computed: %v%30s\n", computeMetricsList, "")

	// Print computed dataset evaluations.
	fmt.Println()
	fmt.Println("Comparing MDID MOS to computed metrics (cm) rankings using different evaluators (ev):")
	fmt.Println()
	fmt.Printf("%10s", "cm\\ev")
	for _, em := range evaluatorsList {
		fmt.Printf("%10s", em)
	}
	fmt.Println()

	for _, cm := range computeMetricsList {
		fmt.Printf("%10s", cm)
		m := dataset.ComputedMetricsByName(cm)
		for _, em := range evaluatorsList {
			fmt.Printf("%10f", math.Abs(evaluators[em](mos, m)))
		}
		fmt.Println()
	}

	fmt.Println()
	fmt.Println("Comparing provided metrics to computed metrics (m) using different evaluators (ev):")
	fmt.Println()
	fmt.Printf("%10s", "m\\ev")
	for _, em := range evaluatorsList {
		fmt.Printf("%10s", em)
	}
	fmt.Println()

	for _, m := range []string{"PSNR", "SSIM"} {
		fmt.Printf("%10s", m)
		pm, cm := dataset.ProvidedMetricsByName(m), dataset.ComputedMetricsByName(m)
		for _, em := range evaluatorsList {
			fmt.Printf("%10f", math.Abs(evaluators[em](pm, cm)))
		}
		fmt.Println()
		//fmt.Println(pm)
		//fmt.Println(cm)
	}
}
