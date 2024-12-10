package main

import (
	_ "image/png"
	"log/slog"
	"os"
	"time"
)

func setupLogger() *slog.Logger {
	stdOutHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	return slog.New(stdOutHandler)

}

// func compareColors(c1, c2 color.RGBA) bool {
// 	if c1.R != c2.R {
// 		return c1.R < c2.R
// 	}
// 	if c1.G != c2.G {
// 		return c1.G < c2.G
// 	}
// 	return c1.B < c2.B
// }

// func sortPixels(pixels []color.RGBA) {
// 	sort.Slice(pixels, func(i, j int) bool {
// 		return compareColors(pixels[i], pixels[j])
// 	})
// }

// func sortWithinWhiteArea(pixels [][]color.RGBA, mask [][]uint8) [][]color.RGBA {

// 	sortedPixels := make([][]color.RGBA, len(pixels))

// 	for i := range pixels {
// 		sortedPixels[i] = make([]color.RGBA, len(pixels[i]))
// 		copy(sortedPixels[i], pixels[i])
// 	}

// 	for i := 0; i < len(pixels); i++ {

// 		var whiteIndices []int
// 		for j, v := range mask[i] {
// 			if v == 255 {
// 				whiteIndices = append(whiteIndices, j)
// 			}
// 		}

// 		if len(whiteIndices) > 1 {

// 			l := 0
// 			r := 1

// 			for l < len(whiteIndices)-2 && r < len(whiteIndices)-1 {

// 				if whiteIndices[r]-whiteIndices[r-1] == 1 {
// 					r++
// 				} else {
// 					sortPixels(sortedPixels[i][whiteIndices[l]:whiteIndices[r]])
// 					l = r
// 					r = r + 1
// 				}
// 			}

// 			if l < len(whiteIndices)-1 {
// 				sortPixels(sortedPixels[i][whiteIndices[l]:whiteIndices[r]])
// 			}
// 		}
// 	}

// 	return sortedPixels
// }

func processPixelatedLuminanceMaskImg(log *slog.Logger, config *Config) {

	srcImg, err := loadImg(config.SrcPath)
	if err != nil {
		return
	}

	luminanceMaskImg := createPixelatedLuminanceMaskImg(
		log,
		srcImg,
		config.PRate,
		config.LowLuminanceLimit,
		config.HighLuminanceLimit,
	)

	saveJpgImg(config.SrcPath, luminanceMaskImg, config.OutQuality)

}

func main() {

	log := setupLogger()

	config, err := parseConfig()
	if err != nil {
		panic(err)
	}

	processingStart := time.Now()
	switch config.Mode {
	case 1:
		processPixelatedLuminanceMaskImg(log, config)
		log.Info("process of creating a pixelated mask based on luminance has begun")
	}

	processingDuration := time.Since(processingStart)
	log.Info("done", slog.Duration("duration", processingDuration))

	/////////

	// bounds := srcImg.Bounds()
	// w, h := bounds.Dx(), bounds.Dy()

	// pixels := make([][]color.RGBA, h)
	// for y := 0; y < h; y++ {
	// 	pixels[y] = make([]color.RGBA, w)
	// 	for x := 0; x < w; x++ {
	// 		pixels[y][x] = color.RGBAModel.Convert(srcImg.At(x, y)).(color.RGBA)
	// 	}
	// }

	// mask := make([][]uint8, h)
	// for y := 0; y < h; y++ {
	// 	mask[y] = make([]uint8, w)
	// 	for x := 0; x < w; x++ {
	// 		c := luminanceMaskImg.At(x, y).(color.RGBA)
	// 		if c.R == 255 && c.G == 255 && c.B == 255 {
	// 			mask[y][x] = 255
	// 		} else {
	// 			mask[y][x] = 0
	// 		}
	// 	}
	// }

	// sortedPixels := sortWithinWhiteArea(pixels, mask)

	// outputFile, err := os.Create("res.jpg")
	// if err != nil {
	// 	fmt.Println("Error creating output file:", err)
	// 	return
	// }
	// defer outputFile.Close()

	// outputImg := image.NewRGBA(bounds)
	// for y := 0; y < h; y++ {
	// 	for x := 0; x < w; x++ {
	// 		outputImg.Set(x, y, sortedPixels[y][x])
	// 	}
	// }

	// err = jpeg.Encode(outputFile, outputImg, &jpeg.Options{Quality: 20})
	// if err != nil {
	// 	fmt.Println("Error encoding output image:", err)
	// 	return
	// }

}
