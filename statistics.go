package main

import (
	"math"
	"sort"
)

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

// MeanA returns arithmetic mean (average) value of input.
func MeanA(a []float64) float64 {
	if len(a) == 0 {
		return math.NaN()
	}
	return Sum(a) / float64(len(a))
}

func meanSd(a []float64) (float64, float64) {
	mean := MeanA(a)

	sd := 0.0
	for _, v := range a {
		sd += (v - mean) * (v - mean)
	}
	sd = math.Sqrt(sd / float64(len(a)-1))
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
	s := a
	if !sort.Float64sAreSorted(a) {
		s = make([]float64, len(a))
		copy(s, a)
		sort.Float64s(s)
	}

	ifl := q * float64(len(s)-1)
	i, d := int(ifl), ifl-float64(int(ifl))

	if i == len(s)-1 {
		return s[i]
	}

	return (1-d)*s[i] + d*s[i+1]

}

type RankingMetod int

const (
	StandartCompetition RankingMetod = iota // 1224
	ModifiedCompetition                     // 1334
	Dense                                   // 1223
	Ordinal                                 // 1234 (if equal, first in input position, first in ranking)
	Fractional                              // 1 2.5 2.5 4
)

// Rank returns ranks for input a
func Rank(a []float64, m RankingMetod) []float64 {
	rs := make([]struct {
		v float64
		i int
	}, len(a))
	for i, v := range a {
		rs[i].v, rs[i].i = v, i
	}
	sort.SliceStable(rs, func(i, j int) bool {
		return rs[i].v < rs[j].v
	})

	res := make([]float64, len(a))
	switch m {
	case StandartCompetition:
		for i := range rs {
			if i-1 < 0 || rs[i].v != rs[i-1].v {
				res[rs[i].i] = float64(i + 1)
			} else {
				res[rs[i].i] = res[rs[i-1].i]
			}
		}
	case ModifiedCompetition:
		for i := len(rs) - 1; i >= 0; i-- {
			if i+1 >= len(rs) || rs[i].v != rs[i+1].v {
				res[rs[i].i] = float64(i + 1)
			} else {
				res[rs[i].i] = res[rs[i+1].i]
			}
		}
	case Dense:
		cr := 1
		for i := range rs {
			res[rs[i].i] = float64(cr)
			if i+1 < len(rs) && rs[i].v != rs[i+1].v {
				cr++
			}
		}
	case Fractional:
		ef, efi := false, 0
		for i := range rs {
			if ef {
				if i+1 >= len(rs) || rs[i].v != rs[i+1].v {
					v := float64((i-efi+1)*(i+1+efi+1)/2) / float64(i-efi+1)
					for j := efi; j <= i; j++ {
						res[rs[j].i] = float64(v)
					}
					ef, efi = false, 0
				}
			} else {
				if i+1 < len(rs) && rs[i].v == rs[i+1].v {
					ef, efi = true, i
				} else {
					res[rs[i].i] = float64(i + 1)
				}
			}
		}
	case Ordinal:
		for i, r := range rs {
			res[r.i] = float64(i + 1)
		}
	default:
		panic("unknown ranking method")
	}
	return res
}
