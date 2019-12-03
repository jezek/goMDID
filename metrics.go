package main

import (
	"image"
	"image/color"
	"math"

	"gonum.org/v1/gonum/stat"
)

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

//TODO https://ece.uwaterloo.ca/~z70wang/research/ssim/
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
