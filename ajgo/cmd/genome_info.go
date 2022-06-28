package cmd

// We will write out in GFF3 format http://gmod.org/wiki/GFF3.
// This requires that column 3 contain a SOFA term or accession.
//
// Using the Sequence Ontology Browser, we found that there is a SO term
// for regions of N's. The term is N_region and the accession is SO:0001835
// http://www.sequenceontology.org/browser/current_release/term/SO:0001835

import (
	"github.com/grendeloz/ngs/genome"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// submode genome > info
var genomeInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "write info from a serialised genome",
	Long:  `For an ajgo serialised genome, write out key metrics.`,
	Run: func(cmd *cobra.Command, args []string) {
		startLogging()
		genomeInfoCmdRun(cmd, args)
		finishLogging()
	},
}

func init() {
	genomeCmd.AddCommand(genomeInfoCmd)

	genomeInfoCmd.Flags().StringVar(&flagInfileGenome, "in-genome", "",
		"ajgo serialised genome")
	genomeInfoCmd.MarkFlagRequired("in-genome")

	//genomeInfoCmd.Flags().StringVar(&flagOutfile, "outfile", "",
	//	"output file")
	//genomeInfoCmd.MarkFlagRequired("outfile")
}

func genomeInfoCmdRun(cmd *cobra.Command, args []string) {
	// Read in base genome
	log.Info("reading serialised genome: ", flagInfileGenome)
	g, err := genome.GenomeFromGob(flagInfileGenome)
	if err != nil {
		log.Fatal(err)
	}

	err = logGenomeInfo(g)
	if err != nil {
		log.Fatal(err)
	}
}

func logGenomeInfo(g *genome.Genome) error {
	log.Info("Name :", g.Name)
	log.Info("UUID :", g.UUID)
	log.Info("FASTA Files:")
	for i, f := range g.FastaFiles {
		log.Infof("  %d  %s %s", i, f.MD5, f.Filepath)
	}
	log.Info("Provenance records:")
	for i, p := range g.Provenance {
		log.Infof("  %d", i)
		log.Info("      Version: ", p.Version)
		log.Info("      Time:    ", p.StartTime)
		log.Info("      Host:    ", p.HostName)
		log.Infof("      User:    %s (%d)", p.UserName, p.UserId)
		log.Infof("      Group:   %s (%d)", p.GroupName, p.GroupId)
		log.Info("      Args:    ", p.Args)
	}
	log.Info("Sequences:")
	for i, s := range g.Sequences {
		log.Infof("  %d  %s %d %s", i, s.Header, s.Length, s.FastaFile)
	}

	return nil
}
