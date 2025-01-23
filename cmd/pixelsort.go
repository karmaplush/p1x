package cmd

import (
	"path/filepath"

	"github.com/karmaplush/p1x/io"
	"github.com/karmaplush/p1x/processing"
	"github.com/spf13/cobra"
)

var pixelsortCmd = &cobra.Command{
	Use:   "pixelsort [PATH]",
	Short: "[IMG -> IMG] Generate an image whose pixels are sorted by a luminance mask",
	Long: `This command generates an image whose pixels are sorted by a pixelated luminosity
mask for the specified image file based on the specified luminosity thresholds and pixelation rate.`,
	Args: cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		if outputFlag == "" {
			outputFlag = filepath.Dir(args[0])
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		srcFilePath := args[0]

		if err := IsValidLuminanceRange(luminanceRange); err != nil {
			return err
		}

		if err := IsValidPixelationRate(pixelationRate); err != nil {
			return err
		}

		img := io.MustOpenImageAsNRGBA(srcFilePath)

		luminanceMaskImg := processing.CreatePixelatedLuminanceMask(
			img,
			luminanceRange[0],
			luminanceRange[1],
			pixelationRate,
			false,
		)

		result := processing.PixelSortingBasedOnMask(img, luminanceMaskImg)

		if err := io.SaveImageAsJPG(result, outputFlag, 100); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	pixelsortCmd.Flags().
		Float64SliceVarP(&luminanceRange, "luminance", "l", []float64{0.3, 0.7},
			`Set luminance range as [min,max]
[0...1, 0...1]. Pixels whose normalized luminance is inside this area will
be painted white, the rest of the pixels will be painted black.`)
	pixelsortCmd.Flags().
		IntVarP(&pixelationRate, "pixrate", "p", 4,
			`Set pixelation rate as integer >= 1. The higher the pixelation rate,
the more "pixelated" the output image will be.`)
	rootCmd.AddCommand(pixelsortCmd)
}
