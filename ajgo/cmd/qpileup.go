package cmd

import (
	"github.com/spf13/cobra"
)

// qpileupCmd implements the qpileup mode
var qpileupCmd = &cobra.Command{
	Use:   "qpileup",
	Short: "operations related to the AdamaJava qpileup tool",
	Long: `
This mode is a collection of operations related to the AdamaJava qpileup
tool. qpileup is used to aggregate data at a base-pair level across large
numbers of BAM files so various summary metrics can be created.`,
}

func init() {
	rootCmd.AddCommand(qpileupCmd)
}
