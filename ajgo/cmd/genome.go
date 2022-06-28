package cmd

import (
	"github.com/spf13/cobra"
)

// genomeCmd is a submode which is purely a collection of sub-submodes
var genomeCmd = &cobra.Command{
	Use:   "genome",
	Short: "Operations on genomes",
	Long: `Operations on genomes which are defined as collections of one
or more sequences, usually in FASTA format.`,
}

func init() {
	rootCmd.AddCommand(genomeCmd)
}
