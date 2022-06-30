package cmd

import (
	"github.com/grendeloz/cmdh"
	"github.com/grendeloz/ngs/genome"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// submode genome > create
var createGenomeCmd = &cobra.Command{
	Use:   "create",
	Short: "create binary genome from FASTA",
	Long: `Read genome as FASTA file(s) and serialise as an ajgo genome in
go encoding/gob binary format. This binary format is required by most other
ajgo modes that use a genome.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdh.StartLogging()
		createGenomeCmdRun(cmd, args)
		cmdh.FinishLogging()
	},
}

func init() {
	genomeCmd.AddCommand(createGenomeCmd)

	createGenomeCmd.Flags().StringSliceVar(&flagFastaFiles, "fasta", []string{},
		"FASTA file to be added to genome")
	createGenomeCmd.MarkFlagRequired("fasta")

	createGenomeCmd.Flags().StringVar(&flagOutfileGenome, "out-genome", "",
		"filename stem for serialised genome")
	createGenomeCmd.MarkFlagRequired("out-genome")

	createGenomeCmd.Flags().StringVar(&flagName, "name", "",
		"name to be embedded in serialised genome")
	createGenomeCmd.MarkFlagRequired("name")
}

func createGenomeCmdRun(cmd *cobra.Command, args []string) {
	gn := genome.NewGenome(flagName)

	for _, file := range flagFastaFiles {
		log.Info("reading FASTA file:", file)

		// log MD5 before processing
		md5, err := md5sum(file)
		if err != nil {
			log.Fatalf("error calculating md5sum: %v", err)
		}
		log.Info("  MD5 checksum: ", md5)

		err = gn.AddFastaFile(file)
		if err != nil {
			log.Fatalf("error adding FASTA file: %v", err)
		}
		log.Infof("  genome %v now contains %v sequences", gn.Name, len(gn.Sequences))
	}

	log.Info("Number of FASTA files parsed: ", len(gn.FastaFiles))
	log.Info("Number of sequences: ", len(gn.Sequences))
	var bctr int
	for _, s := range gn.Sequences {
		bctr = bctr + len(s.Sequence)
	}
	log.Info("Total bases in sequences: ", bctr)

	// Encode and Decode
	log.Info("writing to: ", flagOutfileGenome)
	file, err := gn.WriteAsGob(flagOutfileGenome)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("writing completed: ", file)
}
