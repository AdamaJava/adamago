// Copyright (c) QIMR Berghofer Medical Research Institute

package cmd

import (
	"ajgo/gff3"
	"github.com/grendeloz/cmdh"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// submode genemodel > ensembl-gff3 > gene-cds
var genemodelEnsemblGff3GeneCdsCmd = &cobra.Command{
	Use:   "gene-cds",
	Short: "select and consolidate CDS",
	Long: `Read a GFF3 file, group and only keep those groupings that
have at least one element with a Type of cds.  Extract the CDS elements
and consolidate within the groupings.cdsand only keep features related to a list of
gene names or ids read from a file. Write out the derived GFF3 file.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdh.StartLogging()
		genemodelEnsemblGff3GeneCdsCmdRun(cmd, args)
		cmdh.FinishLogging()
	},
}

func init() {
	genemodelEnsemblGff3Cmd.AddCommand(genemodelEnsemblGff3GeneCdsCmd)

	genemodelEnsemblGff3GeneCdsCmd.Flags().StringVar(&flagInfileGeneModel, "in-gff3", "",
		"gene model input - GFF3 format")
	genemodelEnsemblGff3GeneCdsCmd.MarkFlagRequired("in-gff3")
	genemodelEnsemblGff3GeneCdsCmd.Flags().StringVar(&flagOutfileGeneModel, "out-gff3", "",
		"gene model after consolidation - GFF3 format")
	genemodelEnsemblGff3GeneCdsCmd.MarkFlagRequired("out-gff3")
}

func genemodelEnsemblGff3GeneCdsCmdRun(cmd *cobra.Command, args []string) {
	// Read source GFF3
	log.Info("reading GFF3: ", flagInfileGeneModel)
	gIn, err := gff3.NewFromFile(flagInfileGeneModel)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("  Number of Features: ", gIn.FeatureCount())

	// Clone the input Gff3 - we could do this in place but this is
	// cleaner and leaves gIn unchanged in case we want to use it.
	log.Info("consolidating CDS by gene")
	gOut := gIn.Clone()

	// Create tree
	log.Info("creating Gff3Tree")
	t := gOut.NewTree()
	log.Info("  Number of Nodes: ", len(t.Nodes))
	log.Info("  Number of Orphans: ", len(t.Orphans))

	// TO DO - the consolidation!!!
	log.Info("  Number of Features: ", gOut.FeatureCount())

	// Write out the new post-selection Gff3
	err = gOut.Write(flagOutfileGeneModel)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("writing complete: %s", flagOutfileGeneModel)
}
