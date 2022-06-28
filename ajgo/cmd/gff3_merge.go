package cmd

import (
	"ajgo/gff3"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// submode gff3 > merge
var gff3MergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "merge and consolidate multiple GFF3 files",
	Long: `Read two or more GFF3 files, consolidate records within each
GFF3 and them sequentially merge them in the order they were specified
on the command line, i.e. the second GFF3 is merged onto the first and
then any subsequent GFF3s are merged one at a time onto the new composite.

This mode does not make use of any information from the individual
Feature records apart from the SeqId, Type, Start and End. The Start
and End are used to determine the Allen Relationships between intervals
to determine when merging is appropriate. Merging only happens within
Feature of the same SeqId because that's the nature of genomes - regions
on different sequences do not, by definition overlap.

The use of Type is more nuanced. In order to keep some sort of stats
about the merges, where Feature with different types are merged, instead
of creating one Feature, we create multiple Feature - one Feature is
created with any bases that are in both Feature being merged (i.e. the
overlap) and separate Feature are created for any bases that are only in
one Feature. We call this prudent merging, meaning it is careful to 
preserve as much as possible of the original Feature and to only merge 
those bases where we must merge. Using this strategy it is not unusual
to merge two intervals and get back three intervals.

For a more detailed explanation of both simple and prudent merging, run

 ajgo merge-gff3

Depending on your use case, the gff3 > select mode may be helpful
before or after the merge.`,
	Run: func(cmd *cobra.Command, args []string) {
		startLogging()
		gff3MergeCmdRun(cmd, args)
		finishLogging()
	},
}

func init() {
	gff3Cmd.AddCommand(gff3MergeCmd)

	gff3MergeCmd.Flags().StringSliceVar(&flagGff3Files, "gff3", []string{},
		"GFF3 files to be merged")
	gff3MergeCmd.MarkFlagRequired("gff3")

	gff3MergeCmd.Flags().StringVar(&flagOutfileGeneModel, "out-gff3", "",
		"gene model after consolidation - GFF3 format")
	gff3MergeCmd.MarkFlagRequired("out-gff3")
}

func gff3MergeCmdRun(cmd *cobra.Command, args []string) {
	if len(flagGff3Files) < 2 {
		log.Fatal("must supply at least 2 GFF3 files to merge")
	}

	// TO DO - we need to create a clean GFF3 with new headers
	// etc and then modify the loop so all flagGff3Files are merged onto
	// the clean GFF3. We also need to update the headers from the
	// merged GFF3 so they all get a `-1` suffix to each key, e.g.
	//  ##gff-version 3  -> ##gff-version-1 3
	// This will let is keep all of the headers from the merged GFF3s in
	// the new GFF3.

	gMerged := gff3.NewGff3()
	headers := []string{
		"##gff-version 3",
		"##created-by ajgo mode: gff3 > merge",
	}
	rpHeaders := gffHeadersFromRunParameters()
	gMerged.Header = append(gMerged.Header, headers...)
	gMerged.Header = append(gMerged.Header, rpHeaders...)

	var vHeaders []string

	// Now merge in the files
	for i := 0; i < len(flagGff3Files); i++ {
		file := flagGff3Files[i]
		log.Infof("merging GFF3 file %d: %s", i, file)
		header := fmt.Sprintf("##merged-gff3-file %d %s", i, file)
		gMerged.Header = append(gMerged.Header, header)

		// log MD5 before processing
		md5, err := md5sum(file)
		if err != nil {
			log.Fatalf("error calculating md5sum: %v", err)
		}
		log.Info("  MD5 checksum: ", md5)

		g, err := gff3.NewFromFile(file)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("  this GFF3 file contains %d features from %d SeqIds",
			g.FeatureCount(), len(g.SeqIds()))

		// Merge Features & Headers
		fs1 := gff3.MergeFeatures(gMerged.Features, g.Features)
		gMerged.Features = fs1
		vHeaders = append(vHeaders, "###") // visual separator
		vHeaders = append(vHeaders, g.VersionedHeaders(strconv.Itoa(i))...)

		log.Infof("  merged GFF3 file contains %d features from %d SeqIds",
			gMerged.FeatureCount(), len(gMerged.SeqIds()))
	}
	gMerged.Header = append(gMerged.Header, vHeaders...)

	log.Info("Number of GFF3 files merged: ", len(flagGff3Files))
	log.Info("Number of Sequences with features: ", len(gMerged.SeqIds()))
	log.Info("Sequences:")
	log.Info("  Name\tCount\tSumIntvl\tConsIntvl\tIsSorted")
	seqs := gMerged.Features.BySeqId()
	seqids := gMerged.SeqIds()
	for _, seqid := range seqids {
		fs := seqs[seqid]
		fs.Sort() // This is a requirement for consolidation
		sci, err := fs.SumConsolidatedIntervals()
		if err != nil {
			log.Warning(err)
		}
		log.Infof("  %s\t%d\t%d\t%d\t%t", seqid,
			fs.Count(),
			fs.SumIntervals(),
			sci,
			fs.IsSorted)
	}
	log.Info("Number of features: ", gMerged.FeatureCount())

	// Write out the merged GFF3
	err := gMerged.Write(flagOutfileGeneModel)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("writing complete: %s", flagOutfileGeneModel)
}
