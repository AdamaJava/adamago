package cmd

import (
	"github.com/spf13/cobra"
)

// gff3Cmd is a submode which is a collection of sub-submodes
var gff3Cmd = &cobra.Command{
	Use:   "gff3",
	Short: "Operations on GFF3 files",
	Long: `
Operations on GFF3 files. Note that these are generic operations and
there may be more specific submodes elsewhere in ajgo that work on
particular datasets in GFF3 format and are aware of the specific
properties of those datasets.

For example under genemodel > ensembl-gff3 there are modes that account
for the relationship, in Ensembl GFF3-format gene model files, between 
genes and transcripts and can create appropriate subsets of the gene
models, e.g. gene-by-exon or gene-by-CDS.`,
}

func init() {
	rootCmd.AddCommand(gff3Cmd)
}
