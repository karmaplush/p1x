package main

import (
	"flag"
	"fmt"
)

type Config struct {
	Mode               int
	SrcPath            string
	LowLuminanceLimit  float64
	HighLuminanceLimit float64
	PRate              int
	OutQuality         int
}

// MODE (outputs):
// 1 - Luminance mask
// 2 - Pixel Sorting (horizontal ->) based on Luminance mask

func parseConfig() (*Config, error) {
	mode := flag.Int("mode", 1, "Mode for launching ")
	srcPath := flag.String("src", "", "Source file path")

	lowLuminanceLimit := flag.Float64("lll", 0.2, "Low luminance limit")
	highLuminanceLimit := flag.Float64("hll", 0.6, "High luminance limit")

	pRate := flag.Int("prate", 1, "Pixelation Rate for luminance mask")

	outQuality := flag.Int("outq", 100, "Output .jpg quality. Must be between 1 and 100")

	flag.Parse()

	if *srcPath == "" {
		return nil, fmt.Errorf("-src argument is required")
	}

	if *pRate < 1 {
		return nil, fmt.Errorf(
			"invalid -prate value. It must be greater than or equal to 1 (1 == default, no pixelation)",
		)
	}

	if *outQuality < 1 || *outQuality > 100 {
		return nil, fmt.Errorf(
			"invalid -outq value. It must be greater than 0 and less than 101 (100 == default, best quality)",
		)
	}

	return &Config{
		Mode:               *mode,
		SrcPath:            *srcPath,
		LowLuminanceLimit:  *lowLuminanceLimit,
		HighLuminanceLimit: *highLuminanceLimit,
		PRate:              *pRate,
		OutQuality:         *outQuality,
	}, nil
}
