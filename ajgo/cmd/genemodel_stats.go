package cmd

import (
	"ajgo/gtf"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// cmd globals
var (
	flagGtfFiles []string
)

// submode genome > stats
var genemodelStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "derive stats from gene model",
	Long:  `Write report statistics derived from a gene model.`,
	Run: func(cmd *cobra.Command, args []string) {
		startLogging()
		genemodelStatsCmdRun(cmd, args)
		finishLogging()
	},
}

func init() {
	genemodelCmd.AddCommand(genemodelStatsCmd)

	genemodelStatsCmd.Flags().StringSliceVar(&flagGtfFiles, "gtf", []string{},
		"GTF file to be added to gene model")
	genemodelStatsCmd.MarkFlagRequired("gtf")
}

func genemodelStatsCmdRun(cmd *cobra.Command, args []string) {

	g := gtf.New("JP_test_gene_model")

	for _, file := range flagGtfFiles {
		log.Info("reading Gtf file:", file)

		// log MD5 before processing
		md5, err := md5sum(file)
		if err != nil {
			log.Fatalf("error calculating md5sum: %v", err)
		}
		log.Info("  MD5 checksum: ", md5)

		err = g.AddGtfFile(file)
		if err != nil {
			log.Fatalf("error adding GTF: %v", err)
		}
		log.Infof("  gene model %v now contains features for %v sequences",
			g.Name, len(g.SeqFeats))
	}

	log.Info("Number of GTF files parsed: ", len(g.Files))
	log.Info("Number of Sequences with features: ", len(g.SeqFeats))
	log.Info("Sequences:")
	for _, fs := range g.SeqFeats {
		_, _ = fs.Sort()
		log.Infof("  %s  count:%d  sorted:%t", fs.SeqName, len(fs.Features), fs.IsSorted)
	}
	log.Info("Number of features: ", g.FeatureCount())
}
