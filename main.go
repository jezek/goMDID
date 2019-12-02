package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"

	_ "golang.org/x/image/bmp"

	"github.com/dgryski/go-onlinestats"
	statistics "github.com/mcgrew/gostats"
	"gonum.org/v1/gonum/stat"
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
	return statistics.KendallCorrelation(ca, cb)
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
	return statistics.SpearmanCorrelation(ca, cb)
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

// Returns gray image using default go (luminance) method to convert from rgb.
func gray8(img image.Image) *image.Gray {
	grayImg := image.NewGray(img.Bounds())

	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			grayImg.Set(x, y, img.At(x, y))
			//continue

			//R, G, B, _ := img.At(x, y).RGBA()
			// Luminance: Y = 0.299R + 0.587G + 0.114B
			//Y := 0.299*float64(R) + 0.587*float64(G) + 0.1148*float64(B)
			// Mean: Y = (R + G + B) / 3
			//Y := (float64(R) + float64(G) + float64(B)) / 3.0
			// Luma: Y = 0.2126R + 0.7152G + 0.0722B
			//Y := 0.2126*float64(R) + 0.7152*float64(G) + 0.0722*float64(B)
			// Luster: Y = (min(R, G, B) + max(R, G, B))/2
			//grayImg.Set(x, y, color.Gray{uint8(Y / 257)})
		}
	}

	return grayImg
}

// Returns Mean-Squared Error of the two input 8-bit grayscale images.
// Using: http://homepages.inf.ed.ac.uk/rbf/CVonline/LOCAL_COPIES/VELDHUIZEN/node18.html
func MSEGray(a, b *image.Gray) float64 {
	if !a.Bounds().Eq(b.Bounds()) {
		panic("images have to have equal bounds")
	}

	values := make([]float64, a.Bounds().Dx()*a.Bounds().Dy())
	for i := range values {
		x, y := i%a.Bounds().Dx(), i/a.Bounds().Dx()
		values[i] = float64(int(b.GrayAt(x, y).Y) - int(a.GrayAt(x, y).Y)) // error
		values[i] *= values[i]                                             // square error
	}
	return stat.Mean(values, nil) // mean square error
}

// Returns Mean-Squared Error of the two input color images, converting to gray images (using gray8(...)) and then computing MSEGray(...)
// Using: http://homepages.inf.ed.ac.uk/rbf/CVonline/LOCAL_COPIES/VELDHUIZEN/node18.html
func MSE(a, b image.Image) float64 {
	if !a.Bounds().Eq(b.Bounds()) {
		panic("images have to have equal bounds")
	}

	return MSEGray(gray8(a), gray8(b))

}

// Returns Peak Signal-to-Noise Ratio of the two input color images, using MSE(...) in calculation.
// Using: http://homepages.inf.ed.ac.uk/rbf/CVonline/LOCAL_COPIES/VELDHUIZEN/node18.html
func PSNR(i1, i2 image.Image) float64 {
	return -10 * math.Log10(MSE(i1, i2)/65025.0) // 65025 = 255*255
}

func vector(img *image.Gray) []float64 {
	res := make([]float64, img.Bounds().Dx()*img.Bounds().Dy())
	for i := range res {
		x, y := i%img.Bounds().Dx(), i/img.Bounds().Dx()
		res[i] = float64(img.GrayAt(x, y).Y)
	}
	return res
}

func merge(a, b []float64, f func(i int) float64) []float64 {
	if len(a) != len(b) {
		panic("vectors have different lenghts")
	}
	res := make([]float64, len(a))
	for i := range res {
		res[i] = f(i)
	}
	return res
}

// Returns an gray image for every color component of input image.
func rgbExplodeToGray8(img image.Image) (R *image.Gray, G *image.Gray, B *image.Gray) {
	R, G, B = image.NewGray(img.Bounds()), image.NewGray(img.Bounds()), image.NewGray(img.Bounds())

	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA() //uint16 represented as uint32
			R.Set(x, y, color.Gray{uint8(r / 257)})
			G.Set(x, y, color.Gray{uint8(g / 257)})
			B.Set(x, y, color.Gray{uint8(b / 257)})
		}
	}

	return
}

// Returns Mean-Squared Error of the two input color images, by decompositing RGB values to separate images and return average of them.
func MSErgb(a, b image.Image) float64 {
	if !a.Bounds().Eq(b.Bounds()) {
		panic("images have to have equal bounds")
	}

	aR, aG, aB := rgbExplodeToGray8(a)
	bR, bG, bB := rgbExplodeToGray8(b)
	mseR, mseG, mseB := MSEGray(aR, bR), MSEGray(aG, bG), MSEGray(aB, bB)

	return (mseR + mseG + mseB) / 3
}

// Returns Peak Signal-to-Noise Ratio of the two input color images using MSErgb(...).
// Using: http://homepages.inf.ed.ac.uk/rbf/CVonline/LOCAL_COPIES/VELDHUIZEN/node18.html
func PSNRrgb(i1, i2 image.Image) float64 {
	return -10 * math.Log10(MSErgb(i1, i2)/65025.0) // 65025 = 255*255
}

// Default SSIM constants.
const (
	L  = 255.0
	K1 = 0.01
	K2 = 0.03
)

// Calculated SSIM coeficients.
var (
	C1 = math.Pow((K1 * L), 2.0)
	C2 = math.Pow((K2 * L), 2.0)
)

func SSIM(a, b image.Image) float64 {
	if !a.Bounds().Eq(b.Bounds()) {
		panic("images dimensions not equal")
	}

	ga, gb := gray8(a), gray8(b)

	avgA, stdA, avgB, stdB, covAB := 0.0, 0.0, 0.0, 0.0, 0.0
	for y := ga.Bounds().Min.Y; y < ga.Bounds().Max.Y; y++ {
		for x := ga.Bounds().Min.X; x < ga.Bounds().Max.X; x++ {
			avgA += float64(ga.GrayAt(x, y).Y)
			avgB += float64(gb.GrayAt(x, y).Y)
		}
	}
	n := float64(ga.Bounds().Dx() * ga.Bounds().Dy())
	avgA, avgB = avgA/n, avgA/n
	for y := ga.Bounds().Min.Y; y < ga.Bounds().Max.Y; y++ {
		for x := ga.Bounds().Min.X; x < ga.Bounds().Max.X; x++ {
			vA, vB := float64(ga.GrayAt(x, y).Y), float64(gb.GrayAt(x, y).Y)
			stdA += (vA - avgA) * (vA - avgA)
			stdB += (vB - avgB) * (vB - avgB)
			covAB += (vA - avgA) * (vB - avgB)
		}
	}
	stdA, stdB = math.Sqrt(stdA/(n-1)), math.Sqrt(stdB/(n-1))
	covAB = covAB / (n - 1)

	return (((2.0 * avgA * avgB) + C1) * ((2.0 * covAB) + C2)) / ((math.Pow(avgA, 2.0) + math.Pow(avgB, 2.0) + C1) * (math.Pow(stdA, 2.0) + math.Pow(stdB, 2.0) + C2))
}
