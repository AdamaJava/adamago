package cmd

import (
	"regexp"

	"ajgo/selector"

	"github.com/grendeloz/cmdh"
	"github.com/grendeloz/ngs/genome"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// submode genome > select
var genomeSelectCmd = &cobra.Command{
	Use:   "select",
	Short: "create new genome with subset of sequences",
	Long: `Use selector statements to keep and delete sequences from an
existing genome to create a new genome.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdh.StartLogging()
		genomeSelectCmdRun(cmd, args)
		cmdh.FinishLogging()
	},
}

func init() {
	genomeCmd.AddCommand(genomeSelectCmd)

	genomeSelectCmd.Flags().StringVar(&flagInfileGenome, "in-genome", "",
		"ajgo serialised genome")
	genomeSelectCmd.MarkFlagRequired("in-genome")

	genomeSelectCmd.Flags().StringArrayVar(&flagSelectors, "select", []string{},
		"selector statements (operation:subject:pattern) for filtering sequences")
	genomeSelectCmd.MarkFlagRequired("select")

	genomeSelectCmd.Flags().StringVar(&flagOutfileGenome, "out-genome", "",
		"filestem name for new ajgo serialised genome")
	genomeSelectCmd.MarkFlagRequired("out-genome")
}

func genomeSelectCmdRun(cmd *cobra.Command, args []string) {
	// Read in base genome
	log.Info("reading serialised genome: ", flagInfileGenome)
	g, err := genome.GenomeFromGob(flagInfileGenome)
	if err != nil {
		log.Fatal(err)
	}

	// Get our selectors ready-to-use
	selectors, err := selector.NewFromStrings(flagSelectors)
	if err != nil {
		log.Fatal(err)
	}

	// Apply selectors seriatim
	var keepers []*genome.Sequence
	for _, sel := range selectors {
		log.Infof("applying selector: %s - %s - %s",
			sel.Operation, sel.Subject, sel.Pattern)
		// Set up regex for matching
		rex, err := regexp.Compile(sel.Pattern)
		if err != nil {
			log.Fatal(err)
		}
		for _, seq := range g.Sequences {
			match := rex.Match([]byte(seq.Header))
			switch sel.Operation {
			case `keep`:
				if match {
					keepers = append(keepers, seq)
				} else {
					log.Infof("  deleting sequence: %s", seq.Header)
				}
			case `delete`:
				if match {
					log.Infof("  deleting sequence: %s", seq.Header)
				} else {
					keepers = append(keepers, seq)
				}
			default:
				log.Fatalf("selector operation not recognised: %s", sel.Operation)
			}
		}
	}

	// Swap in the list of Sequence to keep
	g.Sequences = keepers

	// Add a new Provenance record
	g.AddProvenance()

	// Write out the new genome
	file, err := g.WriteAsGob(flagOutfileGenome)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("writing complete: %s", file)
}
