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
		"my":    compare,
		"myg":   compareg,
	}
	fmt.Println()
	computeMetricsList := []string{"PSNRg", "myg", "PSNR", "my", "SSIM"}
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

func mean(img *image.Gray) float64 {
	return meanRect(img, img.Bounds())
}

func meanRect(img *image.Gray, rect image.Rectangle) float64 {
	acc := uint(0)
	for y := rect.Canon().Min.Y; y < rect.Canon().Max.Y; y++ {
		for x := rect.Canon().Min.X; x < rect.Canon().Max.X; x++ {
			acc += uint(img.GrayAt(x, y).Y)
		}
	}
	ret := float64(acc) / float64(rect.Bounds().Dx()*rect.Bounds().Dy())
	//fmt.Printf("mean %dx%d on %d,%d: %f\n", rect.Dx(), rect.Dy(), rect.Min.X, rect.Min.Y, ret)
	return ret
}

var wasRect = map[image.Rectangle]bool{}

// Returns 0 if identical, > 0 otherwise
func compare(a, b image.Image) float64 {
	if !a.Bounds().Eq(b.Bounds()) {
		panic("images have to have equal bounds")
	}
	wasRect = map[image.Rectangle]bool{}
	//log.Printf("ci")
	aR, aG, aB := rgbExplodeToGray8(a)
	bR, bG, bB := rgbExplodeToGray8(b)
	cR, cG, cB := compareRectRecursive(aR, bR, a.Bounds()), compareRectRecursive(aG, bG, a.Bounds()), compareRectRecursive(aB, bB, a.Bounds())

	return (cR + cG + cB) / 3
}

// Returns 0 if identical, > 0 otherwise
func compareg(a, b image.Image) float64 {
	if !a.Bounds().Eq(b.Bounds()) {
		panic("images have to have equal bounds")
	}
	wasRect = map[image.Rectangle]bool{}
	return compareRectRecursive(gray8(a), gray8(b), a.Bounds())
}

func compareRectRecursive(i1, i2 *image.Gray, rect image.Rectangle) float64 {
	//log.Printf("cr %v", rect)
	if rect.Dx()*rect.Dy() == 0 {
		return 0
	}
	if wasRect[rect] {
		return 0
	}
	wasRect[rect] = true
	mi1, mi2 := meanRect(i1, rect), meanRect(i2, rect)

	cm := (math.Abs(mi1-mi2) * math.Abs(mi1-mi2)) * math.Pow((float64(rect.Bounds().Dx()*rect.Bounds().Dy())/float64(i1.Bounds().Dx()*i1.Bounds().Dy())), 1)
	//fmt.Printf("weighted mean %dx%d on %d,%d: %f\n", rect.Dx(), rect.Dy(), rect.Min.X, rect.Min.Y, cm)
	if rect.Dx()*rect.Dy() == 1 {
		return cm
	}

	max := func(a, b int) int {
		if a > b {
			return a
		}
		return b
	}

	return cm +
		compareRectRecursive(i1, i2, image.Rectangle{rect.Min, image.Pt(rect.Min.X+max(1, (rect.Dx()/2)+(rect.Dx()%2)), rect.Min.Y+max(1, (rect.Dy()/2)+(rect.Dy()%2)))}) +
		compareRectRecursive(i1, i2, image.Rectangle{image.Pt(rect.Max.X-max(1, rect.Dx()/2), rect.Min.Y), image.Pt(rect.Max.X, rect.Min.Y+max(1, (rect.Dy()/2)+(rect.Dy()%2)))}) +
		compareRectRecursive(i1, i2, image.Rectangle{image.Pt(rect.Min.X, rect.Max.Y-max(1, rect.Dy()/2)), image.Pt(rect.Min.X+max(1, (rect.Dx()/2)+(rect.Dx()%2)), rect.Max.Y)}) +
		compareRectRecursive(i1, i2, image.Rectangle{image.Pt(rect.Max.X-max(1, rect.Dx()/2), rect.Max.Y-max(1, rect.Dy()/2)), rect.Max})

}
