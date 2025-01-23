package converter

import (
	"fmt"
	"image"
	"image/color"
	"runtime"
	"sync"
)

func convertRGBAToNRGBA(rbga *image.RGBA) *image.NRGBA {

	nrgba := image.NewNRGBA(rbga.Rect)

	numWorkers := runtime.NumCPU()
	lines := make(chan int, rbga.Rect.Max.Y)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for y := range lines {
				for x := rbga.Rect.Min.X; x < rbga.Rect.Max.X; x++ {
					r, g, b, _ := rbga.At(x, y).RGBA()

					nrgba.Set(
						x,
						y,
						color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: 255},
					)
				}
			}
		}()
	}

	for y := 0; y < rbga.Rect.Max.Y; y++ {
		lines <- y
	}
	close(lines)
	wg.Wait()

	return nrgba
}

func convertYCbCrToNRGBA(ycbcr *image.YCbCr) *image.NRGBA {

	nrgba := image.NewNRGBA(ycbcr.Rect)

	numWorkers := runtime.NumCPU()
	lines := make(chan int, ycbcr.Rect.Max.Y)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for y := range lines {
				for x := ycbcr.Rect.Min.X; x < ycbcr.Rect.Max.X; x++ {
					c := ycbcr.YCbCrAt(x, y)
					r, g, b := color.YCbCrToRGB(c.Y, c.Cb, c.Cr)

					nrgba.Set(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
				}
			}
		}()
	}

	for y := 0; y < ycbcr.Rect.Max.Y; y++ {
		lines <- y
	}
	close(lines)

	wg.Wait()
	return nrgba
}

func ConvertToNRGBA(img image.Image) (*image.NRGBA, error) {
	switch img := img.(type) {
	case *image.NRGBA:
		return img, nil
	case *image.YCbCr:
		return convertYCbCrToNRGBA(img), nil
	case *image.RGBA:
		return convertRGBAToNRGBA(img), nil
	default:
		return nil, fmt.Errorf("cannot convert image to NRGBA: unsupported image type %T", img)
	}
}
