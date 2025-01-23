package cmd

import (
	"path/filepath"

	"github.com/karmaplush/p1x/io"
	"github.com/karmaplush/p1x/processing"
	"github.com/spf13/cobra"
)

var lmaskCmd = &cobra.Command{
	Use:   "lmask [PATH]",
	Short: "[IMG -> IMG] Generate a luminance mask",
	Long: `This command generates a luminance mask for the given image file
	based on specified luminance thresholds`,
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

		img := io.MustOpenImageAsNRGBA(srcFilePath)
		result := processing.CreateLuminanceMask(img, luminanceRange[0], luminanceRange[1], false)
		if err := io.SaveImageAsJPG(result, outputFlag, 100); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	lmaskCmd.Flags().
		Float64SliceVarP(&luminanceRange, "luminance", "l", []float64{0.3, 0.7},
			`Set luminance range as [min,max]
[0-1.0,0-1.0]. Pixels whose normalized luminance is inside this area will
be painted white, the rest of the pixels will be painted black.`)
	rootCmd.AddCommand(lmaskCmd)
}
