package cmd

import (
	"github.com/grendeloz/cmdh"
	"github.com/grendeloz/ngs/genome"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// cmd globals
var (
	flagOutDir string
	flagSeeds  []string
)

// submode make > seed
var createSeedCmd = &cobra.Command{
	Use:   "seed",
	Short: "read ajgo genome and apply spaced seeds",
	Long: `
Read an ajgo-format binary genome (see mode genome > create), apply one
or more spaced seeds and write out each seeded genome. Seeds are a string
of characters where 1 indicates a base to be used in the seed and _
indicates a base to be skipped in the seed.

For example, the seed '11_11__111' will interrogate 10mers with 7 of the
bases checked and 3 skipped. Because there are 2 adjacent skipped bases,
this seed would be able to perfectly match against a read that had 2 
adjacent mismatches. The diagram below shows that when the seed is placed
against the read and bases with a 1 are checked against the reference,
the position where both mismatching bases line up with the skipped bases
in the seed ('__') all 7 interrogated bases in the read match the 
reference.

          111111 
 123456789012345  position
 AGCTCAGCTCTTTGC  reference sequence
 AGCTCAGCcaTTTGC  read containing two mismatches (lowercase 'ca')

    11_11__111    seed - 7 matches, 0 mismatches
   11_11__111     seed - 6 matches, 1 mismatches
  11_11__111      seed - 5 matches, 2 mismatches
 11_11__111       seed - 5 matches, 2 mismatches`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdh.StartLogging()
		createSeedCmdRun(cmd, args)
		cmdh.FinishLogging()
	},
}

func init() {
	seedCmd.AddCommand(createSeedCmd)

	createSeedCmd.Flags().StringVar(&flagGobFile, "gob", "",
		"file containing genome serialised as gob")
	createSeedCmd.MarkFlagRequired("gob")

	createSeedCmd.Flags().StringVar(&flagOutDir, "outdir", "",
		"directory for output")
	createSeedCmd.MarkFlagRequired("outdir")

	createSeedCmd.Flags().StringArrayVar(&flagSeeds, "seed", []string{},
		"spaced seed in '1_1' format ")
	createSeedCmd.MarkFlagRequired("mask")
}

func createSeedCmdRun(cmd *cobra.Command, args []string) {

	log.Infof("  --seed: %v", flagSeeds)
	log.Info("  --outdir: ", flagOutDir)
	log.Info("reading serialised genome: ", flagGobFile)
	g, err := genome.GenomeFromGob(flagGobFile)
	if err != nil {
		log.Fatalf("error reading serialised genome: %v", err)
	}
	log.Info("reading genome complete")

	for _, seed := range flagSeeds {
		log.Infof("applying seed %s to genome %s", seed, g.Name)
		gs, err := g.NewSeed(seed)
		if err != nil {
			log.Fatalf("error applying seed: %v", err)
		}
		log.Info("serialising genome seed")
		file, err := gs.WriteAsGob(flagOutDir)
		if err != nil {
			log.Fatalf("error serialising seed to %s: %v", file, err)
		}
		log.Info("writing genome seed as text")
		file, err = gs.WriteAsText(flagOutDir)
		if err != nil {
			log.Fatalf("error writing seed to text file %s: %v", file, err)
		}
	}
}
