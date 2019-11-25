package main

import (
	"fmt"
	"log"
	"math"
	"sort"
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
		"SROCC": SROCC,
		"KROCC": KROCC,
		"PLCC":  PLCC,
		"RMSE":  RMSE,
	}

	evaluatorsList := []string{"SROCC", "KROCC", "PLCC", "RMSE"}
	providedMetricsList := []string{"PSNR", "SSIM", "VIF", "IWSSIM", "FSIMc", "GMSD"}
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
			fmt.Printf("%10f", evaluators[em](mos, m))
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

// KROCC returns Kendall’s rank order correlation coefficient of inputs a, b
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
	return math.Abs(float64(cn-dn) / (0.5 * float64(n*(n-1))))
}

// SROCC returns Spearman’s rank order correlation coefficient of inputs a, b
func SROCC(a, b []float64) float64 {
	if len(a) != len(b) {
		panic("Unexpected input lenghts")
	}

	mrA, mrB := Quantile(a, 0.5), Quantile(b, 0.5)
	sumNum, sumDenA, sumDenB := 0.0, 0.0, 0.0
	for i := range a {
		cA, cB := a[i]-mrA, b[i]-mrB
		sumNum += (cA) * (cB)
		sumDenA += cA * cA
		sumDenB += cB * cB
	}
	return math.Abs(sumNum / (math.Sqrt(sumDenA) * math.Sqrt(sumDenB)))
}

// PLCC returns Pearson’s linear correlation coefficient of inputs a, b
func PLCC(a, b []float64) float64 {
	if len(a) != len(b) {
		panic("Unexpected input lenghts")
	}

	avgA, avgB := MeanA(a), MeanA(b)
	sumNum, sumDenA, sumDenB := 0.0, 0.0, 0.0
	for i := range a {
		cA, cB := a[i]-avgA, b[i]-avgB
		sumNum += (cA) * (cB)
		sumDenA += cA * cA
		sumDenB += cB * cB
	}
	return math.Abs(sumNum / (math.Sqrt(sumDenA) * math.Sqrt(sumDenB)))
}

// RMSE returns root mean square error.
func RMSE(a, b []float64) float64 {
	if len(a) != len(b) {
		panic("Unexpected input lenghts")
	}

	sum := 0.0
	for i := range a {
		sum += (b[i] - a[i]) * (b[i] - a[i])
	}
	return math.Sqrt(sum / float64(len(a)))
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

// Sd returns the standart deviation
func Sd(a []float64) float64 {
	mean := MeanA(a)

	sd := 0.0
	for _, v := range a {
		sd += (v - mean) * (v - mean)
	}
	sd = math.Sqrt(sd)
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
	sorted := make([]float64, len(a))
	copy(sorted, a)
	if !sort.Float64sAreSorted(sorted) {
		sort.Float64s(sorted)
	}

	ifl := q * float64(len(sorted)-1)
	i, d := int(ifl), ifl-float64(int(ifl))

	if i == len(sorted)-1 {
		return sorted[i]
	}

	return (1-d)*sorted[i] + d*sorted[i+1]

}
