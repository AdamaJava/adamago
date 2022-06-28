package cmd

import (
	"time"

	"github.com/grendeloz/ngs/genome"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// submode genome > readgob
var genomeReadGobCmd = &cobra.Command{
	Use:   "readgob",
	Short: "read gob-serialized genome",
	Long:  `Test reading of gob file created by mode genome>writegob.`,
	Run: func(cmd *cobra.Command, args []string) {
		startLogging()
		genomeReadGobCmdRun(cmd, args)
		finishLogging()
	},
}

func init() {
	genomeCmd.AddCommand(genomeReadGobCmd)

	genomeReadGobCmd.Flags().StringVar(&flagInfileGenome, "genome", "",
		"ajgo serialised genome")
	genomeReadGobCmd.MarkFlagRequired("genome")
}

func genomeReadGobCmdRun(cmd *cobra.Command, args []string) {

	log.Info("Reading gob: ", flagInfileGenome)
	g, err := genome.GenomeFromGob(flagInfileGenome)
	if err != nil {
		log.Fatalf("error reading gob file: %v", err)
	}
	log.Info("Reading complete, genome read: ", g.Name)
	log.Info("Number of FASTA files: ", len(g.FastaFiles))
	log.Info("Number of sequences: ", len(g.Sequences))
	var bctr int
	for _, s := range g.Sequences {
		log.Infof("  %d - %s", len(s.Sequence), s.Header)
		bctr = bctr + len(s.Sequence)
	}
	log.Info("Total bases in sequences: ", bctr)

	// At the moment this mode does nothing except read the genome so we
	// will just put in a delay so folks can see what the memory use is.
	time.Sleep(10)
}
