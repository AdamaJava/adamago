package cmd

import (
	"os"
	"strconv"
	"strings"

	"github.com/grendeloz/cmdh"
	"github.com/grendeloz/ngs/genome"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// submode genome > stats
var genomeStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "write summary stats about sequences in a serialised genome",
	Long: `
Write summary stats about sequences in a serialised genome.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdh.StartLogging()
		genomeStatsCmdRun(cmd, args)
		cmdh.FinishLogging()
	},
}

func init() {
	genomeCmd.AddCommand(genomeStatsCmd)

	genomeStatsCmd.Flags().StringVar(&flagInfileGenome, "in-genome", "",
		"ajgo serialised genome")
	genomeStatsCmd.MarkFlagRequired("in-genome")

	genomeStatsCmd.Flags().StringVar(&flagOutfile, "outfile", "",
		"output file")
	genomeStatsCmd.MarkFlagRequired("outfile")
}

func genomeStatsCmdRun(cmd *cobra.Command, args []string) {
	// Read in base genome
	log.Info("reading serialised genome: ", flagInfileGenome)
	g, err := genome.GenomeFromGob(flagInfileGenome)
	if err != nil {
		log.Fatal(err)
	}

	writeInfoTsv(g)
	if err != nil {
		log.Fatal(err)
	}
}

func writeInfoTsv(g *genome.Genome) error {
	log.Info("calculating stats")

	type seqi struct {
		id       int
		name     string
		length   int
		genomepc float64
		gccount  int
	}

	f, err := os.Create(flagOutfile)
	if err != nil {
		return err
	}
	defer f.Close()

	// Count gc and assemble per-sequence stats
	var seqs []seqi
	glength := 0
	gctotal := 0
	for i, s := range g.Sequences {
		si := seqi{id: i + 1, name: s.Header, length: s.Length()}
		glength += s.Length()

		// count GC
		ucs := strings.ToUpper(s.Sequence)
		for i := 0; i < len(ucs); i++ {
			if ucs[i] == 'G' || ucs[i] == 'C' {
				si.gccount += 1
			}
		}
		gctotal += si.gccount

		seqs = append(seqs, si)
	}

	log.Info("Genome Length: ", glength)
	log.Info("GC total count: ", gctotal)

	// Write per-sequence stats
	headers := []string{`Id`,
		`Name`,
		`Length (basepairs)`,
		`% of genome`,
		`GC%`}
	f.Write([]byte(strings.Join(headers, "\t") + "\n"))

	for _, s := range seqs {
		gcpc := float64(s.gccount) / float64(s.length) * 100
		glpc := float64(s.length) / float64(glength) * 100
		gcpcS := strconv.FormatFloat(gcpc, 'f', 5, 64)
		glpcS := strconv.FormatFloat(glpc, 'f', 5, 64)
		n := strings.TrimLeft(s.name, ">")
		l := strconv.Itoa(s.length)
		i := strconv.Itoa(s.id)

		vals := []string{i, n, l, glpcS, gcpcS}
		f.Write([]byte(strings.Join(vals, "\t") + "\n"))
	}

	return nil
}
