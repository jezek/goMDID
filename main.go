package main

import (
	"fmt"
	"log"
	"math"
	"sort"

	"github.com/dgryski/go-onlinestats"
	statistics "github.com/mcgrew/gostats"
	"gonum.org/v1/gonum/stat"
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

	// Print dataset evaluations
	//fmt.Printf("%v\n", dataset)
	evaluators := map[string]func([]float64, []float64) float64{
		"SROCC":   SROCC,
		"SROCCos": SROCConlinestats,
		"SROCCgs": SROCCgostats,
		"KROCC":   KROCC,
		"KROCCgs": KROCCgostats,
		"PLCC":    PLCC,
		"PLCCgn":  PLCCgonum,
		"PLCCgs":  PLCCgostats,
		"RMSE":    RMSE,
	}

	evaluatorsList := []string{"SROCC", "SROCCos", "SROCCgs", "KROCC", "KROCCgs", "PLCC", "PLCCgn", "PLCCgs", "RMSE"}
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

}

func sgn(a float64) int {
	switch {
	case a < 0:
		return -1
	case a > 0:
		return +1
	}
	return 0
}

func concordantPair(x1, y1, x2, y2 float64) bool {
	return sgn(x2-x1) == sgn(y2-y1)
}

func discordantPair(x1, y1, x2, y2 float64) bool {
	return sgn(x2-x1) == -1*sgn(y2-y1)
}

// KROCC returns Kendall’s rank order correlation coefficient of inputs a, b.
func KROCC(a, b []float64) float64 {
	if len(a) != len(b) {
		panic("Unexpected input lenghts")
	}

	n, cn, dn := len(a), 0, 0
	for i := 0; i < len(a); i++ {
		for j := i + 1; j < len(a); j++ {
			if concordantPair(a[i], b[i], a[j], b[j]) {
				cn++
			}
			if discordantPair(a[i], b[i], a[j], b[j]) {
				dn++
			}
		}
	}
	return float64(cn-dn) / (0.5 * float64(n*(n-1)))
}

// KROCCgostats returns Kendall’s rank order correlation coefficient of inputs a, b using gostats implementation.
func KROCCgostats(a, b []float64) float64 {
	ca, cb := make([]float64, len(a)), make([]float64, len(b))
	copy(ca, a)
	copy(cb, b)
	return statistics.KendallCorrelation(ca, cb)
}

// SROCC returns Spearman’s rank order correlation coefficient of inputs a, b
// using: https://www.groundai.com/project/empirical-evaluation-of-full-reference-image-quality-metrics-on-mdid-database/1
func SROCC(a, b []float64) float64 {
	if len(a) != len(b) {
		panic("Unexpected input lenghts")
	}

	mrA, mrB := Quantile(a, 0.5), Quantile(b, 0.5)
	cenMulSum, mrdASum, mrdBSum := 0.0, 0.0, 0.0
	for i := range a {
		cA, cB := a[i]-mrA, b[i]-mrB
		cenMulSum += (cA) * (cB)
		mrdASum += cA * cA
		mrdBSum += cB * cB
	}
	return cenMulSum / (math.Sqrt(mrdASum) * math.Sqrt(mrdBSum))
}

// SROCConlinestats returns Spearman’s rank order correlation coefficient of inputs a, b using go-onlinestats implementation
func SROCConlinestats(a, b []float64) float64 {
	sr, _ := onlinestats.Spearman(a, b)
	return sr
}

// SROCCgostats returns Spearman’s rank order correlation coefficient of inputs a, b using gostats implementation
func SROCCgostats(a, b []float64) float64 {
	ca, cb := make([]float64, len(a)), make([]float64, len(b))
	copy(ca, a)
	copy(cb, b)
	return statistics.SpearmanCorrelation(ca, cb)
}

// PLCC returns Pearson’s linear correlation coefficient of inputs a, b
func PLCC(a, b []float64) float64 {
	if len(a) != len(b) {
		panic("Unexpected input lenghts")
	}

	avgA, avgB := MeanA(a), MeanA(b)
	cenMulSum, stdASum, stdBSum := 0.0, 0.0, 0.0
	for i := range a {
		cA, cB := a[i]-avgA, b[i]-avgB
		cenMulSum += (cA) * (cB)
		stdASum += cA * cA
		stdBSum += cB * cB
	}
	return cenMulSum / (math.Sqrt(stdASum) * math.Sqrt(stdBSum))
}

// PLCCgonum returns Pearson’s linear correlation coefficient of inputs a, b using gonum implementation
func PLCCgonum(a, b []float64) float64 {
	return stat.Correlation(a, b, nil)
}

// PLCCgostats returns Pearson’s linear correlation coefficient of inputs a, b using gostats implementation
func PLCCgostats(a, b []float64) float64 {
	return statistics.PearsonCorrelation(a, b)
}

// RMSE returns root mean square error of inputs a, b which are centered around avgerage and normalized/divaded by standart deviation.
func RMSE(a, b []float64) float64 {
	if len(a) != len(b) {
		panic("Unexpected input lenghts")
	}

	avgA, sdA := meanSd(a)
	avgB, sdB := meanSd(a)

	errSqrSum := 0.0
	for i := range a {
		cnA, cnB := (a[i]-avgA)/sdA, (b[i]-avgB)/sdB
		errSqrSum += (cnB - cnA) * (cnB - cnA)
	}
	return math.Sqrt(errSqrSum / float64(len(a)))
}

// NRMSE_Sd returns root mean square error normalized/divided by standart deviation of a.
func NRMSE_Sd(a, b []float64) float64 {
	return RMSE(a, b) / Sd(a)
}

// NRMSE_Mean returns root mean square error normalized/divided by mean of a.
func NRMSE_Mean(a, b []float64) float64 {
	return RMSE(a, b) / MeanA(a)
}

// NRMSE_Maxmin returns root mean square error normalized/divided by difference between max and min value of a.
func NRMSE_Maxmin(a, b []float64) float64 {
	return RMSE(a, b) / (Max(a) - Min(a))
}

// NRMSE_Iq returns root mean square error normalized/divided by difference interquartile range (i.e. the difference between 25th and 75th percentile).
func NRMSE_Iq(a, b []float64) float64 {
	return RMSE(a, b) / (Quantile(a, 0.75) - Quantile(a, 0.25))
}

func meanSd(a []float64) (float64, float64) {
	mean := MeanA(a)

	sd := 0.0
	for _, v := range a {
		sd += (v - mean) * (v - mean)
	}
	sd = math.Sqrt(sd)
	return mean, sd
}

// Sd returns the standart deviation
func Sd(a []float64) float64 {
	_, sd := meanSd(a)
	return sd
}

// Sum returns summation of input values.
func Sum(a []float64) float64 {

	sum := 0.0
	for _, v := range a {
		sum += v
	}
	return sum
}

// MeanA returns arithmetic mean (average) value of input.
func MeanA(a []float64) float64 {
	if len(a) == 0 {
		return math.NaN()
	}
	return Sum(a) / float64(len(a))
}

// Max returns maximum value from inputs.
func Max(a []float64) float64 {
	if len(a) == 0 {
		return math.NaN()
	}

	max := a[0]
	for _, v := range a[1:] {
		if max < v {
			max = v
		}
	}
	return max
}

// Min returns minimum value from inputs.
func Min(a []float64) float64 {
	if len(a) == 0 {
		return math.NaN()
	}

	min := a[0]
	for _, v := range a[1:] {
		if min > v {
			min = v
		}
	}
	return min
}

// Quantile returns the q quantile from a.
func Quantile(a []float64, q float64) float64 {
	if !sort.Float64sAreSorted(a) {
		sorted := make([]float64, len(a))
		copy(sorted, a)
		sort.Float64s(sorted)
		a = sorted
	}

	ifl := q * float64(len(a)-1)
	i, d := int(ifl), ifl-float64(int(ifl))

	if i == len(a)-1 {
		return a[i]
	}

	return (1-d)*a[i] + d*a[i+1]

}
