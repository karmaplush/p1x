package cmd

import "fmt"

var luminanceRange []float64
var pixelationRate int

func IsValidLuminanceRange(luminanceRange []float64) error {
	if len(luminanceRange) != 2 {
		return fmt.Errorf("invalid range: expected two values")
	}

	min, max := luminanceRange[0], luminanceRange[1]

	if min < 0 || min > 1 || max < 0 || max > 1 || min >= max {
		return fmt.Errorf(
			"invalid range, expected two values in 0...1 range where first < second (e.g., -l 0.2,0.5)",
		)
	}

	return nil
}

func IsValidPixelationRate(pixelationRate int) error {
	if pixelationRate < 1 {
		return fmt.Errorf("invalid pixelation rate: expected integer >= 1")
	}

	return nil
}
