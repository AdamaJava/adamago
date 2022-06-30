package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/grendeloz/cmdh"
	"github.com/grendeloz/ngs/genome"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// submode genome > homopolymer-stats
var genomeHomopolymerStatsCmd = &cobra.Command{
	Use:   "homopolymer-stats",
	Short: "analyse homopolymers in genome",
	Long:  `Tally count of homopolymers by length and type (e.g.  AAAAAA).`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdh.StartLogging()
		genomeHomopolymerStatsCmdRun(cmd, args)
		cmdh.FinishLogging()
	},
}

func init() {
	genomeCmd.AddCommand(genomeHomopolymerStatsCmd)

	genomeHomopolymerStatsCmd.Flags().StringVar(&flagInfileGenome, "in-genome", "",
		"ajgo serialised genome")
	genomeHomopolymerStatsCmd.MarkFlagRequired("in-genome")

	genomeHomopolymerStatsCmd.Flags().StringVar(&flagOutfile, "outfile", "",
		"text output file for homopolymer tallies")
	genomeHomopolymerStatsCmd.MarkFlagRequired("outfile")
}

func genomeHomopolymerStatsCmdRun(cmd *cobra.Command, args []string) {
	// Read in base genome
	log.Info("reading serialised genome: ", flagInfileGenome)
	g, err := genome.GenomeFromGob(flagInfileGenome)
	if err != nil {
		log.Fatal(err)
	}

	hp, err := identifyHomopolymers(g)
	if err != nil {
		log.Fatal(err)
	}

	err = writeHomopolymers(hp, flagOutfile)
	if err != nil {
		log.Fatal(err)
	}
}

type HpTally struct {
	// Tally: map by base and then homopolymer length
	Counts map[string]map[int]int
}

func NewHpTally() *HpTally {
	var hp = HpTally{}
	hp.Counts = make(map[string]map[int]int)
	return &hp
}

func (hp *HpTally) Add(base string, length int) {
	if _, ok := hp.Counts[base]; !ok {
		//do something here
		hp.Counts[base] = make(map[int]int)
	}
	hp.Counts[base][length]++
}

func identifyHomopolymers(g *genome.Genome) (*HpTally, error) {
	log.Info("identifying homopolymers")
	// Traverse sequences searching for homopolymers
	hp := NewHpTally()
	for _, s := range g.Sequences {
		var inRepeat bool
		var this, prev string
		var repeatCtr int
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
					hp.Add(prev, repeatCtr)
					inRepeat = false
					repeatCtr = 0
				}
			}
		}
		// If we were inRepeat when the traverse ended, report the final repeat
		if inRepeat {
			// Report completed homopolymer
			max := len(s.Sequence)
			hp.Add(s.Sequence[max-2:max-1], repeatCtr)
			inRepeat = false
		}
	}

	return hp, nil
}

func writeHomopolymers(hp *HpTally, file string) error {
	log.Info("writing homopolymer report: ", file)

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	// Write header
	_, err = w.WriteString(strings.Join(
		[]string{"Base", "Length", "Count"},
		"\t") + "\n")
	if err != nil {
		return err
	}

	// Tallys out!

	// Sort by base
	var bases []string
	for b, _ := range hp.Counts {
		bases = append(bases, b)
	}
	sort.Strings(bases)

	var tallys []string
	for _, b := range bases {
		// Sort by length within base
		var lengths []int
		for l, _ := range hp.Counts[b] {
			lengths = append(lengths, l)
		}
		sort.Ints(lengths)
		for _, l := range lengths {
			tallys = append(tallys, fmt.Sprintf("%s\t%d\t%d", b, l, hp.Counts[b][l]))
		}
	}

	// Join and write regions lines
	output := strings.Join(tallys, "\n") + "\n"
	_, err = w.WriteString(output)
	if err != nil {
		return err
	}

	return nil
}
