package cmd

import (
	"ajgo/gff3"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// submode gff3 > stats
var gff3StatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "basic by-SeqId stats",
	Long:  `Simple summary statistics by-SeqId for a GFF3.`,
	Run: func(cmd *cobra.Command, args []string) {
		startLogging()
		gff3StatsCmdRun(cmd, args)
		finishLogging()
	},
}

func init() {
	gff3Cmd.AddCommand(gff3StatsCmd)

	gff3StatsCmd.Flags().StringVar(&flagInfile, "gff3", "",
		"GFF3 file")
	gff3MergeCmd.MarkFlagRequired("gff3")
}

func gff3StatsCmdRun(cmd *cobra.Command, args []string) {
	log.Info("reading: ", flagInfile)
	g, err := gff3.NewFromFile(flagInfile)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Total number of features: ", g.FeatureCount())
	log.Info("Sequences:")
	log.Info("  Name\tCount\tSumIntvl\tConsIntvl\tIsSorted")
	seqs := g.Features.BySeqId()
	seqids := g.SeqIds()
	var totalSi, totalSci int
	for _, seqid := range seqids {
		fs := seqs[seqid]
		sorted := fs.IsSorted
		fs.Sort() // This is a requirement for consolidation
		si := fs.SumIntervals()
		sci, err := fs.SumConsolidatedIntervals()
		if err != nil {
			log.Warning(err)
		}
		log.Infof("  %s\t%d\t%d\t%d\t%t", seqid,
			fs.Count(),
			si,
			sci,
			sorted)
		totalSi += si
		totalSci += sci
	}
	log.Infof("  Totals\t%d\t%d\t%d", g.FeatureCount, totalSi, totalSci)
}
