package cmd

import (
	"github.com/spf13/cobra"
)

// ngscheckCmd is a submode which is a collection of sub-submodes
var ngscheckCmd = &cobra.Command{
	Use:   "ngscheck",
	Short: "modes related to proprietary NGScheck tool",
	Long:  `Operations on NGScheck output files including in JSON
format. NGScheck is a proprietary tool that operates on qprofiler2
output XML files so these modes are only relevant if you have NGScheck.`,
}

func init() {
	rootCmd.AddCommand(ngscheckCmd)
}
