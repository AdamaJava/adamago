package cmd

import (
	"github.com/spf13/cobra"
)

// qmotifCmd is a mode to contain submodes related to qmotif
var qmotifCmd = &cobra.Command{
	Use:   "qmotif",
	Short: "operations related to the AdamaJava qmotif tool",
	Long: `
This mode is a collection of operations related to the AdamaJava qmotif
tool. qmotif searches for genomic motifs in next-generation sequencing 
reads. A common use case for qmotif is to find reads that contain 
telomeric motifs.`,
}

func init() {
	rootCmd.AddCommand(qmotifCmd)
}
