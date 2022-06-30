package cmd

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/grendeloz/cmdh"
	"github.com/grendeloz/ngs/genome"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// submode qmotif > motif
var qmotifMotifCmd = &cobra.Command{
	Use:   "motif",
	Short: "search for one or more motifs in a genome",
	Long: `Search for motifs (as a regular expression) in a genome. The
code does not search the reverse complement so if you want the search to
work on both strands, you should specify a second motif that matches the
reverse complement.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdh.StartLogging()
		qmotifMotifCmdRun(cmd, args)
		cmdh.FinishLogging()
	},
}

func init() {
	qmotifCmd.AddCommand(qmotifMotifCmd)

	qmotifMotifCmd.Flags().StringSliceVar(&flagFastaFiles, "fasta", []string{},
		"FASTA file to be added to genome")
	qmotifMotifCmd.MarkFlagRequired("fasta")

	qmotifMotifCmd.Flags().StringVar(&flagOutfile, "outfile", "",
		"text output file for motif locations")
	qmotifMotifCmd.MarkFlagRequired("outfile")

	// We must use StringArrayVar *not StringSliceVar for this flag
	// because the Slice behaviour allows for the pattern "--x a,b"
	// where a and b are considered to be 2 parts of a composite list of
	// --x flags. This is interesting *but* regexes need to be able to
	// contain the ',' char without pflags messing with it.

	qmotifMotifCmd.Flags().StringArrayVar(&flagRegexps, "regex", []string{},
		"regular expression to be seached for")
	qmotifMotifCmd.MarkFlagRequired("regex")
}

func qmotifMotifCmdRun(cmd *cobra.Command, args []string) {

	// Check that we can compile all of the patterns - this is cheap so
	// we should test this potential point of failure *before* the
	// expensive FASTA file MD5 and reading.
	log.Infof("Search terms (%d): %s", len(flagRegexps), strings.Join(flagRegexps, " ; "))
	var searches []*Search
	for _, p := range flagRegexps {
		r, err := regexp.Compile(p)
		if err != nil {
			log.Fatal(err)
		}
		s := NewSearch()
		s.Pattern = p
		s.Regexp = r
		//s := Search(Pattern:p, Regexp:r)
		searches = append(searches, s)
		log.Infof("compiling regular expression engine for pattern: %s", p)
	}

	// Make sure we can open the output file - no point in doing all the
	// work and then discovering that we can't write the output
	f, err := os.Create(flagOutfile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()

	// MD5 and read the FASTA files(s)
	genome := genome.NewGenome("JP_test_genome")
	for _, file := range flagFastaFiles {
		log.Info("reading FASTA file:", file)

		// log MD5 before processing
		md5, err := md5sum(file)
		if err != nil {
			log.Fatalf("error calculating md5sum: %v", err)
		}
		log.Info("  MD5 checksum: ", md5)

		err = genome.AddFastaFile(file)
		if err != nil {
			log.Fatalf("error adding FASTA file: %v", err)
		}
		log.Infof("  genome %v now contains %v sequences", genome.Name, len(genome.Sequences))
	}

	log.Info("Number of FASTA files parsed: ", len(genome.FastaFiles))
	log.Info("Number of sequences: ", len(genome.Sequences))
	var bctr int
	for _, s := range genome.Sequences {
		bctr = bctr + len(s.Sequence)
	}
	log.Info("Total bases in sequences: ", bctr)

	// ghetto parallelism by search pattern.
	//
	// N.B. The way we set "s" *inside* the loop is ABSOLUTELY CRITICAL. If
	// we set "s" as part of the range statement then there is only one "s"
	// across all of the loops and because the go func is a closure, the
	// various instances of the closure ALL SEE THE SAME "s" so they all
	// execute on the last value of "s".  Putting the "s" inside the loop
	// makes it local to each instance of the loop so each closure sees it's
	// own "s".  All good! This problem cost me hours of investigation
	// and I'm still not sure I 100% understand it but it works
	// correctly now.

	var wg sync.WaitGroup
	for i, _ := range searches {
		wg.Add(1)
		s := searches[i]
		log.Infof("launching goroutine for pattern %d: %s (%p)", i, s.Pattern, s)
		go func() {
			defer wg.Done()
			err := searchGenomeForMotif(genome, s)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	// Wait until all go funcs have completed and then continue.
	wg.Wait()

	// Construct column names string - will be written for each pattern
	colNames := []string{"Sequence", "Start", "End", "Match"}
	colNamesString := strings.Join(colNames, "\t") + "\n"

	// Write search lines
	for _, s := range searches {
		log.Infof("%d matches found for pattern %s", len(s.Results), s.Pattern)
		_, err = w.WriteString("###  Pattern: " + s.Pattern +
			" MatchCount: " + strconv.Itoa(len(s.Results)) + "  ###\n" +
			colNamesString)
		if err != nil {
			log.Fatal(err)
		}
		for _, r := range s.Results {
			fields := []string{r.SeqName,
				strconv.Itoa(r.Start),
				strconv.Itoa(r.End),
				r.Match}
			line := strings.Join(fields, "\t") + "\n"
			_, err = w.WriteString(line)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

type Search struct {
	Pattern string
	Regexp  *regexp.Regexp
	Results []Match
}

func NewSearch() *Search {
	return &Search{}
}

type Match struct {
	SeqName string
	Start   int
	End     int    // 1 past last char in the match
	Match   string // the actual match
}

// searchGenomeForMotif will pattern match through a genome. All results
// go straight into the Results field of the input *Search so there is
// nothing to return except an error.
func searchGenomeForMotif(g *genome.Genome, search *Search) error {
	// Traverse sequences
	for _, seq := range g.Sequences {
		matches := search.Regexp.FindAllStringSubmatchIndex(seq.Sequence, -1)
		for _, m := range matches {
			n := Match{SeqName: seq.Header,
				Start: m[0],
				End:   m[1],
				Match: seq.Sequence[m[0]:m[1]],
			}
			search.Results = append(search.Results, n)
		}
		log.Infof("  found %d matches for pattern %s in sequence %s (%d bases)",
			len(matches), search.Pattern, seq.Header, len(seq.Sequence))
	}

	log.Info("completed genome search for: ", search.Pattern)
	return nil
}
