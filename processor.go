package main

import (
	"image"
	"image/color"
	"log/slog"
	"math"
)

func getLuminance(r uint32, g uint32, b uint32) float64 {
	rf := float64(r>>8) / 255.0
	gf := float64(g>>8) / 255.0
	bf := float64(b>>8) / 255.0
	return 0.2126*rf + 0.7152*gf + 0.0722*bf
}

func createPixelatedLuminanceMaskImg(
	log *slog.Logger,
	img image.Image,
	pixelationRate int,
	lowLuminanceLimit float64,
	highLuminanceLimit float64,
) image.Image {

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	downscaledWidth := int(math.Ceil(float64(width) / float64(pixelationRate)))
	downscaledHeight := int(math.Ceil(float64(height) / float64(pixelationRate)))

	downscaledImage := image.NewRGBA(image.Rect(0, 0, downscaledWidth, downscaledHeight))

	log.Info(
		"Empty downscaled luminance image mask created",
		slog.Group("maskDimensions",
			slog.Int("width", downscaledWidth),
			slog.Int("height", downscaledHeight),
		),
	)

	for dy := 0; dy < downscaledHeight; dy++ {
		for dx := 0; dx < downscaledWidth; dx++ {

			var totalBlockLuminance float64
			var totalPixelsInBlockCount int

			for y := dy * pixelationRate; y < (dy+1)*pixelationRate && y < height; y++ {
				for x := dx * pixelationRate; x < (dx+1)*pixelationRate && x < width; x++ {

					srcPixelColor := img.At(x, y)
					r, g, b, _ := srcPixelColor.RGBA()

					srcPixelLuminance := getLuminance(r, g, b)

					totalBlockLuminance += srcPixelLuminance
					totalPixelsInBlockCount++
				}
			}

			averageBlockLuminance := totalBlockLuminance / float64(totalPixelsInBlockCount)

			var blockColor color.Color
			if averageBlockLuminance >= lowLuminanceLimit &&
				averageBlockLuminance <= highLuminanceLimit {
				blockColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
			} else {
				blockColor = color.RGBA{R: 0, G: 0, B: 0, A: 255}
			}
			downscaledImage.Set(dx, dy, blockColor)
		}
	}

	upscaledImage := image.NewRGBA(bounds)
	for dy := 0; dy < downscaledHeight; dy++ {
		for dx := 0; dx < downscaledWidth; dx++ {
			blockColor := downscaledImage.At(dx, dy)

			for y := dy * pixelationRate; y < (dy+1)*pixelationRate && y < height; y++ {
				for x := dx * pixelationRate; x < (dx+1)*pixelationRate && x < width; x++ {
					upscaledImage.Set(x, y, blockColor)
				}
			}
		}
	}

	return upscaledImage
}
