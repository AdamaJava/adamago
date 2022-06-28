package cmd

import (
	"github.com/spf13/cobra"
)

// submode genemodel > ensembl-gff3
var genemodelEnsemblGff3Cmd = &cobra.Command{
	Use:   "ensembl-gff3",
	Short: "operations on ensembl gene models in gff3 format",
	Long:  "Operations on ensembl gene models in gff3 format.",
}

func init() {
	genemodelCmd.AddCommand(genemodelEnsemblGff3Cmd)
}
