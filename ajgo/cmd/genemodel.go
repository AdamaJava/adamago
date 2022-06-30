// Copyright (c) QIMR Berghofer Medical Research Institute

package cmd

import (
	"github.com/spf13/cobra"
)

// genemodelCmd is a submode which is a collection of sub-submodes
var genemodelCmd = &cobra.Command{
	Use:   "genemodel",
	Short: "Operations on gene models",
	Long: `Operations on gene models which are defined as collections of
gene features, defined by a sequence, start and end bases and some
information. They are typically distributed in GTF or GFF3 format.`,
}

func init() {
	rootCmd.AddCommand(genemodelCmd)
}
