package cmd

import (
	"ajgo/gtf"

	"github.com/grendeloz/cmdh"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// submode genemodel > exons
var genemodelExonsCmd = &cobra.Command{
	Use:   "exons",
	Short: "prune and consolidate exon features from GTF gene model",
	Long: `Prune and consolidate exon features from gene model. Sequences
that are not chromosomes or GLxxxxx are removed, duplicate exons (from
multiple transcripts are removed), and where exons overlap they are
consolidated into a single feature that spans the overlapping exons.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdh.StartLogging()
		genemodelExonsCmdRun(cmd, args)
		cmdh.FinishLogging()
	},
}

func init() {
	genemodelCmd.AddCommand(genemodelExonsCmd)

	// pkg global flagGtfFiles is defined elsewhere
	genemodelExonsCmd.Flags().StringSliceVar(&flagGtfFiles, "in-gtf", []string{},
		"GTF file to be added to gene model")
	genemodelExonsCmd.MarkFlagRequired("gtf")

	genemodelExonsCmd.Flags().StringVar(&flagOutfileExons, "exons", "",
		"text file containing consolidated exon features")
	genemodelExonsCmd.MarkFlagRequired("exons")

	genemodelExonsCmd.Flags().StringArrayVar(&flagDeleteSeqPatterns,
		"delete-sequence-pattern", []string{},
		"regular expression(s) of sequences to be deleted from gene model")
}

func genemodelExonsCmdRun(cmd *cobra.Command, args []string) {
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

	// Prune out unwanted SeqFeats
	if len(flagDeleteSeqPatterns) != 0 {
		var deletedSeqFeats []string
		for _, pattern := range flagDeleteSeqPatterns {
			log.Infof("deleting sequences that match pattern: /%s/",
				pattern)
			names, err := g.DeleteSeqFeats(pattern)
			if err != nil {
				log.Fatalf("error deleting SeqFeats: %v", err)
			}
			deletedSeqFeats = append(deletedSeqFeats, names...)
			log.Info("  Sequences deleted: ", names)
			log.Info("  Number of Sequences with features remaining: ",
				len(g.SeqFeats))
		}
	}

	// Establish consistent ordering for SeqFeats

	names := g.SeqFeatNames()
	log.Info("Sequences:")
	for _, name := range names {
		fs := g.SeqFeats[name]
		_, _ = fs.Sort()
		log.Infof("  %s  count:%d  sorted:%t", fs.SeqName, len(fs.Features), fs.IsSorted)
	}
	log.Info("Number of features: ", g.FeatureCount())

	log.Info("removing any features that are not exons:")
	for _, name := range names {
		fs := g.SeqFeats[name]
		i := fs.KeepByFeatures([]string{`exon`})
		log.Infof("  %s  kept:%d  deleted:%d", fs.SeqName, len(fs.Features), i)
	}
	log.Info("Number of features: ", g.FeatureCount())

	log.Info("consolidating features:")
	for _, name := range names {
		fs := g.SeqFeats[name]
		i, err := fs.Consolidate()
		if err != nil {
			log.Fatalf("error consolidating SeqFeat %s: %v", name, err)
		}
		log.Infof("  %s  kept:%d  merged:%d", fs.SeqName, len(fs.Features), i)
	}
	log.Info("Number of features: ", g.FeatureCount())
	// TO DO - this is not working yet

	log.Info("Writing output file: ", flagOutfileExons)
	err := g.Write(flagOutfileExons)
	if err != nil {
		log.Fatalf("error writing file: %v", err)
	}
}
