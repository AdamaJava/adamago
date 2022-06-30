package cmd

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/grendeloz/cmdh"
	"github.com/grendeloz/ngs/genome"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// submode genome > homopolymer
var genomeHomopolymerCmd = &cobra.Command{
	Use:   "homopolymer",
	Short: "create GFF3 of homopolymers in genome",
	Long: `Create a GFF3 dwifile that lists all of the homopolymers in
a genome of a given length or greater.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdh.StartLogging()
		genomeHomopolymerCmdRun(cmd, args)
		cmdh.FinishLogging()
	},
}

func init() {
	genomeCmd.AddCommand(genomeHomopolymerCmd)

	genomeHomopolymerCmd.Flags().StringVar(&flagInfileGenome, "in-genome", "",
		"ajgo serialised genome")
	genomeHomopolymerCmd.MarkFlagRequired("in-genome")

	genomeHomopolymerCmd.Flags().IntVar(&flagThreshold, "min-length", 5,
		"minimum length for reporting homopolymers")

	genomeHomopolymerCmd.Flags().StringVar(&flagOutfile, "gff3", "",
		"gff3 file of homopolymer regions")
	genomeHomopolymerCmd.MarkFlagRequired("gff3")
}

func genomeHomopolymerCmdRun(cmd *cobra.Command, args []string) {
	// Read in base genome
	log.Info("reading serialised genome: ", flagInfileGenome)
	g, err := genome.GenomeFromGob(flagInfileGenome)
	if err != nil {
		log.Fatal(err)
	}

	err = identifyHomopolymerRegions(g)
	if err != nil {
		log.Fatal(err)
	}
}

func identifyHomopolymerRegions(g *genome.Genome) error {
	log.Info("identifying homopolymers")

	f, err := os.Create(flagOutfile)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	// Write GFF3 header
	header := "##gff-version 3\n"
	header += "##content homopolymer regions\n"
	header += "##min-length " + strconv.Itoa(flagThreshold) + "\n"
	header += "##format 1-based half-open\n"
	header += "##genome " + flagInfileGenome + "\n"
	header += gffHeaderFromRunParameters()
	_, err = w.WriteString(header)
	if err != nil {
		return err
	}

	// Traverse sequences searching for homopolymers
	rctr := 0
	for _, s := range g.Sequences {
		var inRepeat bool
		var this, prev string
		var repeatCtr int
		// i starts at one because we are always looking at previous char
		for i := 1; i < len(s.Sequence); i++ {
			// ranges within substring are half-open - 0:1 is the first character
			this = s.Sequence[i : i+1]
			prev = s.Sequence[i-1 : i]
			if this == prev {
				if !inRepeat {
					inRepeat = true
					repeatCtr = 2
				} else {
					repeatCtr++
				}
			} else {
				if inRepeat {
					// Report completed homopolymer
					if repeatCtr >= flagThreshold {
						rctr++
						rec := makeHomopolymerGffRecord(s.Header, prev, i, repeatCtr, rctr)
						_, err = w.WriteString(rec + "\n")
						if err != nil {
							return err
						}
					}
					inRepeat = false
					repeatCtr = 0
				}
			}
		}
		// If we were inRepeat when the traverse ended, report the final repeat
		if inRepeat {
			// Report completed homopolymer
			if repeatCtr >= flagThreshold {
				max := len(s.Sequence)
				rctr++
				rec := makeHomopolymerGffRecord(s.Header, prev, max, repeatCtr, rctr)
				_, err = w.WriteString(rec + "\n")
				if err != nil {
					return err
				}
			}
			inRepeat = false
		}
	}

	return nil
}

// Note that what is passed in is the current base in the sequence, not
// the start - it's only when process the first base PAST the
// homopolymer that we know it has ended. So we need to do the math
// based on the current location and the repeat length and remembering
// that we are using half-open numbering and that GFF3 coordinates are
// supposed to be is 1-based and we are working from a 0-based loop.
func makeHomopolymerGffRecord(seq, base string, current, length, ctr int) string {
	gff3fields := []string{
		seq,
		`ajgo:homopolymer`,
		`remark`,
		strconv.Itoa(current + 1 - length),
		strconv.Itoa(current + 1),
		`.`,
		`.`,
		`.`,
		`ID=hpoly` + strconv.Itoa(ctr) +
			`;base=` + base +
			`;length=` + strconv.Itoa(length)}
	line := strings.Join(gff3fields, "\t")
	return line
}
