package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var outputFlag string

var rootCmd = &cobra.Command{
	Use:   "p1x",
	Short: "CLI for (some kind of) digital art image & video processing",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().
		StringVarP(&outputFlag, "output", "o", "", "Directory to save output")
}
