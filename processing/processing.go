package processing

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"runtime"
	"sort"
	"sync"
	"time"
)

func getLuminance(r uint8, g uint8, b uint8) float64 {
	rf, gf, bf := float64(r)/255.0, float64(g)/255.0, float64(b)/255.0
	return 0.2126*rf + 0.7152*gf + 0.0722*bf
}

func CreateLuminanceMask(
	src *image.NRGBA,
	leftThreshold float64,
	rightThreshold float64,
	isReversedRange bool,
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

					luminance := getLuminance(*r, *g, *b)

					if isReversedRange { // ===---=== (outer range)
						if luminance < leftThreshold || luminance > rightThreshold {
							*r, *g, *b = 255, 255, 255
						} else {
							*r, *g, *b = 0, 0, 0
						}
					} else { // ---===--- (inner range)
						if luminance > leftThreshold && luminance < rightThreshold {
							*r, *g, *b = 255, 255, 255
						} else {
							*r, *g, *b = 0, 0, 0
						}
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

func CreatePixelatedLuminanceMask(
	src *image.NRGBA,
	leftThreshold float64,
	rightThreshold float64,
	pixelationRate int,
	isReversedRange bool,
) *image.NRGBA {

	if pixelationRate == 1 {
		return CreateLuminanceMask(src, leftThreshold, rightThreshold, isReversedRange)
	}

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

					if isReversedRange { // ===---=== (outer range)
						if avgBlockLuminance < leftThreshold || avgBlockLuminance > rightThreshold {
							blockColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
						} else {
							blockColor = color.RGBA{R: 0, G: 0, B: 0, A: 255}
						}
					} else { // ---===--- (inner range)
						if avgBlockLuminance > leftThreshold && avgBlockLuminance > rightThreshold {
							blockColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
						} else {
							blockColor = color.RGBA{R: 0, G: 0, B: 0, A: 255}
						}
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

func sortPixelsInRange(img *image.NRGBA, line, start, end int) {

	stride := img.Stride

	var pixels []color.RGBA
	for x := start; x < end; x++ {
		idx := line*stride + x*4
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
		idx := line*stride + (start+x)*4
		img.Pix[idx] = px.R
		img.Pix[idx+1] = px.G
		img.Pix[idx+2] = px.B
		img.Pix[idx+3] = px.A
	}
}

func PixelSortingBasedOnMask(src *image.NRGBA, mask *image.NRGBA) *image.NRGBA {
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

				l := 0
				r := 0

				for l < width {

					lIdx := y*stride + l*4

					if mask.Pix[lIdx] == 255 && mask.Pix[lIdx+1] == 255 &&
						mask.Pix[lIdx+2] == 255 {

						r = l
						for r < width {
							rIdx := y*stride + r*4
							if !(mask.Pix[rIdx] == 255 && mask.Pix[rIdx+1] == 255 && mask.Pix[rIdx+2] == 255) {
								break
							}
							r++
						}

						sortPixelsInRange(img, y, l, r)

						l = r

					} else {
						l++
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
	fmt.Println("[Pixel Sorting] Processing done in", processingDuration)

	return img
}
