package cmd

import (
	"ajgo/gff3"
	"ajgo/selector"

	"github.com/grendeloz/cmdh"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// submode genemodel > ensembl-gff3 > select
var genemodelEnsemblGff3SelectCmd = &cobra.Command{
	Use:   "select",
	Short: "select features from gene model",
	Long: `Read a GFF3 file, apply one or more selectors and write out
the derived GFF3 file out. Currently supported selector subjects and
operations are:

  Subject    Operation(s)
  seqid      keep, delete

Subject seqid matches against the seqid string which is the first column
of each GFF3 record. 

For a general description of selectors, see

    ajgo selector --help
`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdh.StartLogging()
		genemodelEnsemblGff3SelectCmdRun(cmd, args)
		cmdh.FinishLogging()
	},
}

func init() {
	genemodelEnsemblGff3Cmd.AddCommand(genemodelEnsemblGff3SelectCmd)

	genemodelEnsemblGff3SelectCmd.Flags().StringVar(&flagInfile, "in-gff3", "",
		"gene model to be selected - GFF3 format")
	genemodelEnsemblGff3SelectCmd.MarkFlagRequired("in-gff3")
	genemodelEnsemblGff3SelectCmd.Flags().StringVar(&flagOutfile, "out-gff3", "",
		"gene model after selects - GFF3 format")
	genemodelEnsemblGff3SelectCmd.MarkFlagRequired("out-gff3")

	genemodelEnsemblGff3SelectCmd.Flags().StringArrayVar(&flagSelectors, "select", []string{},
		"selector statements (operation:subject:pattern) for filtering features")
	genemodelEnsemblGff3SelectCmd.MarkFlagRequired("select")
}

func genemodelEnsemblGff3SelectCmdRun(cmd *cobra.Command, args []string) {
	log.Info("reading GFF3: ", flagInfile)
	g, err := gff3.NewFromFile(flagInfile)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Number of Features: ", g.FeatureCount())

	// Apply selectors seriatim
	log.Info("applying selectors")
	selectors, err := selector.NewFromStrings(flagSelectors)
	if err != nil {
		log.Fatal(err)
	}
	for _, sel := range selectors {
		log.Infof("applying selector: %s - %s - %s",
			sel.Operation, sel.Subject, sel.Pattern)
		g.ApplySelector(sel)
	}
	log.Info("Number of Features: ", g.FeatureCount())

	// Write out the new post-selection Gff3
	err = g.Write(flagOutfile)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("writing complete: %s", flagOutfile)

}
