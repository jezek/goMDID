package main

import (
	"math"

	"github.com/dgryski/go-onlinestats"
	gostats "github.com/mcgrew/gostats"
	gonumstat "gonum.org/v1/gonum/stat"
)

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
// Using: https://web.archive.org/web/20181008171919/https://docs.scipy.org/doc/scipy/reference/generated/scipy.stats.kendalltau.html
func KROCC(a, b []float64) float64 {
	if len(a) != len(b) {
		panic("Unexpected input lenghts")
	}

	ra, rb := Rank(a, Fractional), Rank(b, Fractional)

	cn, dn, tan, tbn := 0, 0, 0, 0
	for i := 0; i < len(ra); i++ {
		for j := i + 1; j < len(ra); j++ {
			da, db := ra[j]-ra[i], rb[j]-rb[i]
			if da == 0 || db == 0 {
				if da != db {
					if da == 0 {
						tan++
					} else {
						tbn++
					}
				}
			} else {
				if concordantPair(ra[i], rb[i], ra[j], rb[j]) {
					cn++
				} else if discordantPair(ra[i], rb[i], ra[j], rb[j]) {
					dn++
				}
			}
		}
	}
	return float64(cn-dn) / math.Sqrt(float64((cn+dn+tan)*(cn+dn+tbn)))
}

// KROCCgostats returns Kendall’s rank order correlation coefficient of inputs a, b using gostats implementation.
func KROCCgostats(a, b []float64) float64 {
	ca, cb := make([]float64, len(a)), make([]float64, len(b))
	copy(ca, a)
	copy(cb, b)
	return gostats.KendallCorrelation(ca, cb)
}

// KROCCgonum returns Kendall’s rank order correlation coefficient of inputs a, b using gonum implementation.
func KROCCgonum(a, b []float64) float64 {
	return gonumstat.Kendall(a, b, nil)
}

// SROCC returns Spearman’s rank order correlation coefficient of inputs a, b
// Using: https://en.wikipedia.org/wiki/Spearman%27s_rank_correlation_coefficient#Definition_and_calculation
func SROCC(a, b []float64) float64 {
	if len(a) != len(b) {
		panic("Unexpected input lenghts")
	}

	return PLCC(Rank(a, Fractional), Rank(b, Fractional))
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
	return gostats.SpearmanCorrelation(ca, cb)
}

// PLCC returns Pearson’s linear correlation coefficient of inputs a, b
// Using: https://en.wikipedia.org/wiki/Pearson_correlation_coefficient#For_a_sample
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
	return gonumstat.Correlation(a, b, nil)
}

// PLCCgostats returns Pearson’s linear correlation coefficient of inputs a, b using gostats implementation
func PLCCgostats(a, b []float64) float64 {
	return gostats.PearsonCorrelation(a, b)
}
