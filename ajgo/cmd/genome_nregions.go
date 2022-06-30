package cmd

// We will write out in GFF3 format http://gmod.org/wiki/GFF3.
// This requires that column 3 contain a SOFA term or accession.
//
// Using the Sequence Ontology Browser, we found that there is a SO term
// for regions of N's. The term is N_region and the accession is SO:0001835
// http://www.sequenceontology.org/browser/current_release/term/SO:0001835

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

// submode genome > stats
var genomeNregionsCmd = &cobra.Command{
	Use:   "n-regions",
	Short: "find all contiguous runs of N in a genome",
	Long: `For an ajgo serialised genome, identify all contiguous runs
of N bases and output to a GFF3 file.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdh.StartLogging()
		genomeNregionsCmdRun(cmd, args)
		cmdh.FinishLogging()
	},
}

func init() {
	genomeCmd.AddCommand(genomeNregionsCmd)

	genomeNregionsCmd.Flags().StringVar(&flagInfileGenome, "in-genome", "",
		"ajgo serialised genome")
	genomeNregionsCmd.MarkFlagRequired("in-genome")

	genomeNregionsCmd.Flags().StringVar(&flagOutfile, "gff3", "",
		"output file in GFF3")
	genomeNregionsCmd.MarkFlagRequired("gff3")
}

func genomeNregionsCmdRun(cmd *cobra.Command, args []string) {
	// Read in base genome
	log.Info("reading serialised genome: ", flagInfileGenome)
	g, err := genome.GenomeFromGob(flagInfileGenome)
	if err != nil {
		log.Fatal(err)
	}

	err = identifyNRegions(g, flagOutfile)
	if err != nil {
		log.Fatal(err)
	}
}

func identifyNRegions(g *genome.Genome, file string) error {
	log.Info("searching for contiguous runs of N")

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	// Write GFF3 header
	header := "##gff-version 3\n"
	header += "##content genomic N regions\n"
	header += "##format 1-based half-open\n"
	header += "##genome " + flagInfileGenome + "\n"
	header += gffHeaderFromRunParameters()
	_, err = w.WriteString(header)
	if err != nil {
		return err
	}

	// Inner slice is seqName, start, end, length
	var nregions [][]string

	// Traverse sequences
	// See also for dev: https://go.dev/play/p/vbJLvvY7I7v
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
					// Report completed N homopolymer
					if prev == `N` || prev == `n` {
						nregions = append(nregions, []string{
							s.Header,
							strconv.Itoa(i - repeatCtr),
							strconv.Itoa(repeatCtr)})
					}
					inRepeat = false
					repeatCtr = 0
				}
			}
		}
		// If we were inRepeat when the traverse ended, report the final repeat
		if inRepeat {
			// Report completed homopolymer
			max := len(s.Sequence)
			if s.Sequence[max-1:max] == `N` ||
				s.Sequence[max-1:max] == `n` {
				//fmt.Printf("%s %d %d\n", s.Sequence[max-2:max-1], max-repeatCtr, repeatCtr)
				nregions = append(nregions, []string{
					s.Header,
					strconv.Itoa(max - repeatCtr),
					strconv.Itoa(repeatCtr)})
			}
			inRepeat = false
		}
	}

	// Construct and write GFF3 records
	log.Info("writing N regions file: ", flagOutfile)
	var lines []string
	for i, n := range nregions {
		start, err := strconv.Atoi(n[1])
		if err != nil {
			return err
		}
		length, err := strconv.Atoi(n[2])
		if err != nil {
			return err
		}
		gff3fields := []string{
			n[0],
			`ajgo:n-regions`,
			`N_region`,
			n[1],
			strconv.Itoa(start + length),
			`.`,
			`.`,
			`.`,
			`ID=nregion` + strconv.Itoa(i+1) +
				`;length=` + strconv.Itoa(length)}
		line := strings.Join(gff3fields, "\t")
		lines = append(lines, line)
	}
	output := strings.Join(lines, "\n") + "\n"
	_, err = w.WriteString(output)
	if err != nil {
		return err
	}

	return nil
}
