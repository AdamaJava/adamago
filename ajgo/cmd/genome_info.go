package cmd

import (
	"github.com/grendeloz/cmdh"
	"github.com/grendeloz/ngs/genome"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// submode genome > info
var genomeInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "describe a serialised genome",
	Long:  `For an ajgo serialised genome, write out key info.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdh.StartLogging()
		genomeInfoCmdRun(cmd, args)
		cmdh.FinishLogging()
	},
}

func init() {
	genomeCmd.AddCommand(genomeInfoCmd)

	genomeInfoCmd.Flags().StringVar(&flagInfileGenome, "in-genome", "",
		"ajgo serialised genome")
	genomeInfoCmd.MarkFlagRequired("in-genome")
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
