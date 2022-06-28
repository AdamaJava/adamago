package cmd

import (
	"github.com/spf13/cobra"
)

// seedCmd collects seed-related submodes
var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Operations on genome seeds",
	Long: `
This mode is a collection of operations related to spaced seeds used for
matching reads against reference sequences. This is not a current
priority and so is a work-in-progress.`,
}

func init() {
	rootCmd.AddCommand(seedCmd)
}
