package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"
)

type Config struct {
	Mode               int
	SrcPath            string
	OutPath            string
	LowLuminanceLimit  float64
	HighLuminanceLimit float64
	PRate              int
	OutQuality         int
}

func parseConfig() (*Config, error) {
	// MODE (outputs):
	// 1 - Luminance mask
	// 2 - Pixel Sorting (horizontal ->) based on Luminance mask

	mode := flag.Int("mode", 1, "Mode for image processing")
	srcPath := flag.String("src", "", "Source file path")
	outPath := flag.String("out", "", "Output file dir. Default same with source file")

	lowLuminanceLimit := flag.Float64("lll", 0.2, "Low luminance limit")
	highLuminanceLimit := flag.Float64("hll", 0.6, "High luminance limit")

	pRate := flag.Int("prate", 1, "Pixelation Rate for luminance mask")

	outQuality := flag.Int("outq", 100, "Output .jpg quality. Must be between 1 and 100")

	flag.Parse()

	if *srcPath == "" {
		return nil, fmt.Errorf("-src argument is required")
	}

	if *outPath == "" {
		dir := filepath.Dir(*srcPath)
		*outPath = dir
	}

	if *pRate < 1 {
		return nil, fmt.Errorf(
			"invalid -prate value. It must be greater than or equal to 1 (1 == default, no pixelation)",
		)
	}

	if *outQuality < 1 || *outQuality > 100 {
		return nil, fmt.Errorf(
			"invalid -outq value. Value must be in range [1...100] (100 == default, best quality)",
		)
	}

	return &Config{
		Mode:               *mode,
		SrcPath:            *srcPath,
		OutPath:            *outPath,
		LowLuminanceLimit:  *lowLuminanceLimit,
		HighLuminanceLimit: *highLuminanceLimit,
		PRate:              *pRate,
		OutQuality:         *outQuality,
	}, nil
}

func openImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("file opening error: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("file decoding error: %w", err)
	}

	return img, nil
}

func saveImageAsJPG(img *image.NRGBA, dir string, quality int) error {

	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf(
		"%s/frame_%d.jpg",
		dir,
		timestamp,
	)

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("file creation error: %w", err)
	}
	defer file.Close()

	err = jpeg.Encode(file, img, &jpeg.Options{Quality: quality})
	if err != nil {
		return fmt.Errorf("image encoding error: %w", err)
	}

	return nil

}

func convertSRGBToNRGBA(srgbImg *image.RGBA) *image.NRGBA {

	nrgba := image.NewNRGBA(srgbImg.Rect)
	for y := srgbImg.Rect.Min.Y; y < srgbImg.Rect.Max.Y; y++ {
		for x := srgbImg.Rect.Min.X; x < srgbImg.Rect.Max.X; x++ {
			r, g, b, _ := srgbImg.At(x, y).RGBA()

			nrgba.Set(
				x,
				y,
				// TODO: Do I really need to dealing with transparency?
				// For now alpha constant 255
				color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: 255},
			)
		}
	}

	return nrgba
}

func convertYCbCrToNRGBA(ycbcr *image.YCbCr) *image.NRGBA {

	nrgba := image.NewNRGBA(ycbcr.Rect)
	for y := ycbcr.Rect.Min.Y; y < ycbcr.Rect.Max.Y; y++ {
		for x := ycbcr.Rect.Min.X; x < ycbcr.Rect.Max.X; x++ {
			c := ycbcr.YCbCrAt(x, y)
			r, g, b := color.YCbCrToRGB(c.Y, c.Cb, c.Cr)

			nrgba.Set(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
		}
	}

	return nrgba
}

func convertToNRGBA(img image.Image) *image.NRGBA {
	switch img := img.(type) {
	case *image.NRGBA:
		return img
	case *image.YCbCr:
		return convertYCbCrToNRGBA(img)
	case *image.RGBA:
		return convertSRGBToNRGBA(img)
	default:
		return nil
	}
}

func getLuminance(r uint8, g uint8, b uint8) float64 {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0
	return 0.2126*rf + 0.7152*gf + 0.0722*bf
}

func _createLuminanceMask(
	src *image.NRGBA,
	leftBoundary float64,
	rightBoundary float64,
) *image.NRGBA {
	processingStart := time.Now()

	img := image.NewNRGBA(src.Bounds())
	img.Pix = make([]uint8, len(src.Pix))
	copy(img.Pix, src.Pix)

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	stride := img.Stride

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {

			idx := y*stride + x*4

			r := &img.Pix[idx]
			g := &img.Pix[idx+1]
			b := &img.Pix[idx+2]
			// a := img.Pix[idx+3]

			luminance := getLuminance(*r, *g, *b)
			if luminance > leftBoundary && luminance < rightBoundary {
				*r, *g, *b = 255, 255, 255
			} else {
				*r, *g, *b = 0, 0, 0
			}
		}
	}

	processingDuration := time.Since(processingStart)
	fmt.Println("[Luminance Mask (no greenlets)] Processing done in", processingDuration)

	return img
}

func createLuminanceMask(
	src *image.NRGBA,
	leftBoundary float64,
	rightBoundary float64,
) *image.NRGBA {
	processingStart := time.Now()

	img := image.NewNRGBA(src.Bounds())
	img.Pix = make([]uint8, len(src.Pix))
	copy(img.Pix, src.Pix)

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	stride := img.Stride

	numWorkers := runtime.NumCPU()
	lines := make(chan int, height)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for y := range lines {
				for x := 0; x < width; x++ {
					idx := y*stride + x*4

					r := &img.Pix[idx]
					g := &img.Pix[idx+1]
					b := &img.Pix[idx+2]
					// a := img.Pix[idx+3]

					luminance := getLuminance(*r, *g, *b)
					if luminance > leftBoundary && luminance < rightBoundary {
						*r, *g, *b = 255, 255, 255
					} else {
						*r, *g, *b = 0, 0, 0
					}
				}
			}
		}()
	}

	for y := 0; y < height; y++ {
		lines <- y
	}
	close(lines)

	wg.Wait()

	processingDuration := time.Since(processingStart)
	fmt.Println("[Luminance Mask] Processing done in", processingDuration)

	return img
}

func createPixelatedLuminanceMask(
	src *image.NRGBA,
	leftBoundary float64,
	rightBoundary float64,
	pixelationRate int,
) *image.NRGBA {
	processingStart := time.Now()

	var wg sync.WaitGroup

	img := image.NewNRGBA(src.Bounds())
	img.Pix = make([]uint8, len(src.Pix))
	copy(img.Pix, src.Pix)

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	stride := img.Stride

	downscaledWidth := int(math.Ceil(float64(width) / float64(pixelationRate)))
	downscaledHeight := int(math.Ceil(float64(height) / float64(pixelationRate)))

	downscaledImg := image.NewNRGBA(image.Rect(0, 0, downscaledWidth, downscaledHeight))

	numWorkers := runtime.NumCPU()
	downscaledLines := make(chan int, downscaledHeight)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for dy := range downscaledLines {
				for dx := 0; dx < downscaledWidth; dx++ {

					var totalBlockLuminance float64
					var totalPixelsInBlockCount int

					for y := dy * pixelationRate; y < (dy+1)*pixelationRate && y < height; y++ {
						for x := dx * pixelationRate; x < (dx+1)*pixelationRate && x < width; x++ {
							idx := y*stride + x*4
							r := &img.Pix[idx]
							g := &img.Pix[idx+1]
							b := &img.Pix[idx+2]

							luminance := getLuminance(*r, *g, *b)
							totalBlockLuminance += luminance
							totalPixelsInBlockCount++
						}
					}

					avgBlockLuminance := totalBlockLuminance / float64(totalPixelsInBlockCount)

					var blockColor color.Color
					if avgBlockLuminance > leftBoundary && avgBlockLuminance < rightBoundary {
						blockColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
					} else {
						blockColor = color.RGBA{R: 0, G: 0, B: 0, A: 255}
					}

					downscaledImg.Set(dx, dy, blockColor)
				}
			}
		}()
	}

	for y := 0; y < downscaledHeight; y++ {
		downscaledLines <- y
	}
	close(downscaledLines)
	wg.Wait()

	downscaledLines = make(chan int, downscaledHeight)
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for dy := range downscaledLines {
				for dx := 0; dx < downscaledWidth; dx++ {

					blockColor := downscaledImg.At(dx, dy)

					for y := dy * pixelationRate; y < (dy+1)*pixelationRate && y < height; y++ {
						for x := dx * pixelationRate; x < (dx+1)*pixelationRate && x < width; x++ {
							img.Set(x, y, blockColor)
						}
					}
				}
			}
		}()
	}

	for y := 0; y < downscaledHeight; y++ {
		downscaledLines <- y
	}
	close(downscaledLines)
	wg.Wait()

	processingDuration := time.Since(processingStart)
	fmt.Println("[Pixelated Luminance Mask] Processing done in", processingDuration)

	return img
}

func sortPixelsInRange(img *image.NRGBA, y, start, end int) {
	stride := img.Stride

	var pixels []color.RGBA
	for x := start; x < end; x++ {
		idx := y*stride + x*4
		pixels = append(pixels, color.RGBA{
			R: img.Pix[idx],
			G: img.Pix[idx+1],
			B: img.Pix[idx+2],
			A: img.Pix[idx+3],
		})
	}

	sort.Slice(pixels, func(i, j int) bool {
		luminance := func(c color.RGBA) float64 {
			return 0.2126*float64(c.R) + 0.7152*float64(c.G) + 0.0722*float64(c.B)
		}
		return luminance(pixels[i]) < luminance(pixels[j])
	})

	for x, px := range pixels {
		idx := y*stride + (start+x)*4
		img.Pix[idx] = px.R
		img.Pix[idx+1] = px.G
		img.Pix[idx+2] = px.B
		img.Pix[idx+3] = px.A
	}
}

func pixelSortingBasedOnMask(src *image.NRGBA, mask *image.NRGBA) *image.NRGBA {
	processingStart := time.Now()

	img := image.NewNRGBA(src.Bounds())
	img.Pix = make([]uint8, len(src.Pix))
	copy(img.Pix, src.Pix)

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	stride := img.Stride

	numWorkers := runtime.NumCPU()
	lines := make(chan int, height)

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for y := range lines {
				lp := 0
				for lp < width {
					for lp < width {
						lpIdx := y*stride + lp*4
						maskPixel := mask.Pix[lpIdx : lpIdx+4]
						if !(maskPixel[0] == 0 && maskPixel[1] == 0 && maskPixel[2] == 0) {
							break
						}
						lp++
					}

					if lp >= width {
						break
					}
					rp := lp
					for rp < width {
						rpIdx := y*stride + rp*4
						maskPixel := mask.Pix[rpIdx : rpIdx+4]
						if maskPixel[0] == 0 && maskPixel[1] == 0 && maskPixel[2] == 0 {
							break
						}
						rp++
					}

					sortPixelsInRange(img, y, lp, rp)

					lp = rp
				}

			}
		}()

	}

	for y := 0; y < height; y++ {
		lines <- y
	}

	close(lines)
	wg.Wait()

	processingDuration := time.Since(processingStart)
	fmt.Println("[Pixel Sorting] Processing done in", processingDuration)

	return img
}

func main() {

	config, err := parseConfig()
	if err != nil {
		panic(err)
	}

	srcImg, err := openImage(config.SrcPath)
	if err != nil {
		panic(err)
	}

	img := convertToNRGBA(srcImg)
	fmt.Printf("Decoded image type value %T was converted to %T\n", srcImg, img)

	// lm := createPixelatedLuminanceMask(
	// 	img,
	// 	config.LowLuminanceLimit,
	// 	config.HighLuminanceLimit,
	// 	config.PRate,
	// )

	// res := pixelSortingBasedOnMask(img, lm)

	// saveImageAsJPG(res, config.OutPath, 5)

	for range 96 {

		randomLowLuminanceLimit := config.LowLuminanceLimit + rand.Float64()*0.025
		randomHighLuminanceLimit := config.HighLuminanceLimit + rand.Float64()*0.025
		// randomQuality := 1 + rand.Intn(2)
		randomPRate := config.PRate

		pixelatedLuminanceMaskImg := createPixelatedLuminanceMask(
			img,
			randomLowLuminanceLimit,
			randomHighLuminanceLimit,
			randomPRate,
		)

		res := pixelSortingBasedOnMask(img, pixelatedLuminanceMaskImg)

		saveImageAsJPG(res, config.OutPath, 5)

	}

}
