// The output GFF will need a SOFA term for column 3, type. I can't find
// a term that exactly suits our use case but I feel that "remark"
// SO:0000700 is a good fit - see link below.
// http://www.sequenceontology.org/browser/current_release/term/SO:0000700
//
// We should also include '##' header lines that specify:
//  1. The qpileup view file(s)
//  2. the threshold for mapping quality
//  3. the minimum region length.

package cmd

import (
	"ajgo/qpv1"
	"bufio"
	"compress/gzip"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/grendeloz/cmdh"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// submode qpileup > low-mapq
var qpileupLowmapqCmd = &cobra.Command{
	Use:   "low-mapq",
	Short: "identifies regions of low mapping quality",
	Long: `Takes one or more qpileup view files and looks for regions
that have low average mapping quality scores. The qpileup view reports
must be of a particular format and all inputs will be checked that they
have the required columns in the required order.

The user can specify an average mapping quality threshold below which 
positions will be reported, and a minimum reportable length for the 
regions. Both of these parameters have defaults. The minimum size for
region length is important otherwise the GFF3 report can become noisy
with very short regions.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdh.StartLogging()
		qpileupLowmapqCmdRun(cmd, args)
		cmdh.FinishLogging()
	},
}

var (
	flagThresholdy int
)

func init() {
	qpileupCmd.AddCommand(qpileupLowmapqCmd)

	qpileupLowmapqCmd.Flags().StringSliceVar(&flagViewFiles, "view", []string{},
		"qpileup view mode data extract file(s)")
	qpileupLowmapqCmd.Flags().StringVar(&flagFilelistFile, "viewlist", "",
		"text file listing qpileup view fles for analysis")

	qpileupLowmapqCmd.Flags().IntVar(&flagRegionLength, "region-min", 100,
		"minimum reportable region size")
	qpileupLowmapqCmd.Flags().IntVar(&flagThresholdy, "threshold", 10,
		"positions with average mapq below this number will be reported")

	qpileupLowmapqCmd.Flags().StringVar(&flagOutfile, "gff3", "",
		"output GFF3 file containing regions")
	qpileupLowmapqCmd.MarkFlagRequired("gff3")
}

func qpileupLowmapqCmdRun(cmd *cobra.Command, args []string) {

	// log key parameters
	log.Infof("  --threshold %v", flagThresholdy)
	log.Infof("  --region-min %d", flagRegionLength)

	// Consolidate Qpileup View Files to be processed including
	// removing any duplicates
	var ViewFiles []string
	ViewFiles, err := ConsolidateFilesList(flagFilelistFile, flagViewFiles)
	if err != nil {
		log.Fatalf("error consolidating list of qpileup view files: %v", err)
	}

	// Before doing anything, we check will check every file to make
	// sure the first line is *exactly* what we expect. Processing is
	// tres expensive and we don't want to get to the last file and
	// then crap out so we check all files up front - fail early
	// costs a little time but it's good design.
	log.Info("checking for required data columns in --view files")
	for _, file := range ViewFiles {
		err := qpv1.CheckFileHeader(file)
		if err != nil {
			log.Fatalf("error with columns in %s: %v", file, err)
		}
	}
	log.Info("Number of view files checked: ", len(ViewFiles))

	err = ProcessLowmapq(ViewFiles)
	if err != nil {
		log.Fatal(err)
	}
}

// ProcessLowmapq is a beast and way bigger than I'd like but there it is.
// Because the files are so huge and there are potentially so many regions,
// we can't afford to save up "hits" and report at the end - we must report
// as we go.
func ProcessLowmapq(files []string) error {
	// Open GFF3 files for reporting
	log.Info("writing low-mapq GFF3 file: ", flagOutfile)
	of, err := os.Create(flagOutfile)
	if err != nil {
		return err
	}
	defer of.Close()
	ow := bufio.NewWriter(of)
	defer ow.Flush()

	// Write GFF3 header including report files
	header := "##gff-version 3\n"
	header += "##content low average mapping quality report from qpileup view file(s)\n"
	header += "##threshold " + strconv.Itoa(flagThresholdy) + "\n"
	header += "##region-min " + strconv.Itoa(flagRegionLength) + "\n"
	header += gffHeaderFromRunParameters()
	_, err = ow.WriteString(header)
	if err != nil {
		return err
	}
	for _, f := range files {
		_, err = ow.WriteString("##qpileup-view-file " + f + "\n")
		if err != nil {
			return err
		}
	}

	// We will keep a tally of the average mapq.
	lmqTally := make(map[int]int)

	// Main loop through all of the qpileup view files
	for _, file := range files {
		log.Infof("processing file: %s", file)

		// Open file
		ff, err := os.Open(file)
		if err != nil {
			return err
		}
		defer ff.Close()

		// We need to define this before we handle gzip
		var scanner *bufio.Scanner

		// Based on file extension, handle gzip files
		found, err := regexp.MatchString(`\.[gG][zZ]$`, file)
		if err != nil {
			return fmt.Errorf("error matching gzip file pattern against %s: %w", file, err)
		}
		if found {
			// For gzip files, put a gzip.Reader into the chain
			reader, err := gzip.NewReader(ff)
			if err != nil {
				return fmt.Errorf("error opening gzip file %s: %w", file, err)
			}
			defer reader.Close()
			scanner = bufio.NewScanner(reader)
		} else {
			// For non gzip files, go straight to bufio.Reader
			scanner = bufio.NewScanner(ff)
		}

		// Unnecessary but explicit
		scanner.Split(bufio.ScanLines)

		// Skip header lines and manually process first data line. This bit
		// of extra work will give us a cleaner loop for the rest of the data.
		var inLowMapQ bool
		var lowMapQStart int
		var lowMapQTotal int
		var prevRef string
		var fields []string

		rex := regexp.MustCompile(`^#`)
		for scanner.Scan() {
			line := strings.TrimSuffix(scanner.Text(), "\n")
			if rex.MatchString(line) {
				// Skip headers
			} else {
				// Process first data record
				fields = strings.Split(line, "\t")
				if fields[qpv1.Ref_base] == `N` {
					break
				}
				// Unavoidable ugliness - math with strings
				mqf, err := strconv.Atoi(fields[qpv1.MapQual_for])
				if err != nil {
					return fmt.Errorf("error converting MapQual_for: %s",
						fields[qpv1.MapQual_for])
				}
				mqr, err := strconv.Atoi(fields[qpv1.MapQual_rev])
				if err != nil {
					return fmt.Errorf("error converting MapQual_rev: %s",
						fields[qpv1.MapQual_rev])
				}
				rf, err := strconv.Atoi(fields[qpv1.ReferenceNo_for])
				if err != nil {
					return fmt.Errorf("error converting ReferenceNo_for: %s",
						fields[qpv1.ReferenceNo_for])
				}
				nrf, err := strconv.Atoi(fields[qpv1.NonreferenceNo_for])
				if err != nil {
					return fmt.Errorf("error converting NonreferenceNo_for: %s",
						fields[qpv1.NonreferenceNo_for])
				}
				rr, err := strconv.Atoi(fields[qpv1.ReferenceNo_rev])
				if err != nil {
					return fmt.Errorf("error converting ReferenceNo_rev: %s",
						fields[qpv1.ReferenceNo_rev])
				}
				nrr, err := strconv.Atoi(fields[qpv1.NonreferenceNo_rev])
				if err != nil {
					return fmt.Errorf("error converting NonreferenceNo_rev: %s",
						fields[qpv1.NonreferenceNo_rev])
				}
				mapqAvg := 0
				// Without this readdepth check, the divide by zero does NOT throw
				// an error but does give a result of -9223372036854775808 !!!
				readdepth := rf + nrf + rr + nrr
				if readdepth > 0 {
					mapqAvg = int(math.Round(float64(mqf+mqr) /
						float64(rf+nrf+rr+nrr)))
				}
				lmqTally[mapqAvg]++
				if mapqAvg < flagThresholdy {
					inLowMapQ = true
					pos, err := strconv.Atoi(fields[qpv1.Position])
					if err != nil {
						return fmt.Errorf("error converting Position: %s",
							fields[qpv1.Position])
					}
					lowMapQTotal = mapqAvg
					lowMapQStart = pos
				}
				prevRef = fields[qpv1.Reference]
				break
			}
		}

		// Read the file
		lctr := 1
		sctr := 0
		rctr := 0
		for scanner.Scan() {
			if lctr%10000000 == 0 {
				log.Infof("  %d lines processed", lctr)
			}
			lctr++
			line := strings.TrimSuffix(scanner.Text(), "\n")
			fields = strings.Split(line, "\t")
			// If too fields are present then the referencing later will
			// cause a panic so skip (but count) any short lines.
			if len(fields) < 33 {
				//log.Warnf("  line %d has only %d fields: %s", lctr, len(fields), line)
				sctr++
				continue
			}
			// Unavoidable ugliness - math with strings
			mqf, err := strconv.Atoi(fields[qpv1.MapQual_for])
			if err != nil {
				return fmt.Errorf("error converting MapQual_for: %s",
					fields[qpv1.MapQual_for])
			}
			mqr, err := strconv.Atoi(fields[qpv1.MapQual_rev])
			if err != nil {
				return fmt.Errorf("error converting MapQual_rev: %s",
					fields[qpv1.MapQual_rev])
			}
			rf, err := strconv.Atoi(fields[qpv1.ReferenceNo_for])
			if err != nil {
				return fmt.Errorf("error converting ReferenceNo_for: %s",
					fields[qpv1.ReferenceNo_for])
			}
			nrf, err := strconv.Atoi(fields[qpv1.NonreferenceNo_for])
			if err != nil {
				return fmt.Errorf("error converting NonreferenceNo_for: %s",
					fields[qpv1.NonreferenceNo_for])
			}
			rr, err := strconv.Atoi(fields[qpv1.ReferenceNo_rev])
			if err != nil {
				return fmt.Errorf("error converting ReferenceNo_rev: %s",
					fields[qpv1.ReferenceNo_rev])
			}
			nrr, err := strconv.Atoi(fields[qpv1.NonreferenceNo_rev])
			if err != nil {
				return fmt.Errorf("error converting NonreferenceNo_rev: %s",
					fields[qpv1.NonreferenceNo_rev])
			}
			mapqAvg := 0
			// Without this readdepth check, the divide by zero does NOT throw
			// an error but does give a result of -9223372036854775808 !!!
			readdepth := rf + nrf + rr + nrr
			if readdepth > 0 {
				mapqAvg = int(math.Round(float64(mqf+mqr) /
					float64(rf+nrf+rr+nrr)))
			}
			lmqTally[mapqAvg]++
			if fields[qpv1.Ref_base] == `N` {
				// Hitting an N always stops a region
				if inLowMapQ {
					// Exiting a LowMapQ region so save previous Region
					pos, err := strconv.Atoi(fields[qpv1.Position])
					if err != nil {
						return fmt.Errorf("error converting Position: %s",
							fields[qpv1.Position])
					}
					if pos-lowMapQStart > flagRegionLength {
						rctr++
						_, err =
							ow.WriteString(makeLowMapqGffRecord(fields[qpv1.Reference],
								lowMapQStart, pos, rctr, lowMapQTotal) + "\n")
						if err != nil {
							return err
						}
					}
					inLowMapQ = false
					lowMapQStart = 0
					lowMapQTotal = 0
				}
			} else if mapqAvg < flagThresholdy {
				// Start tracking LowMapQ region if not already doing so
				if !inLowMapQ {
					inLowMapQ = true
					pos, err := strconv.Atoi(fields[qpv1.Position])
					if err != nil {
						return fmt.Errorf("error converting Position: %s",
							fields[qpv1.Position])
					}
					lowMapQStart = pos
					lowMapQTotal = mapqAvg
				} else {
					lowMapQTotal += mapqAvg
				}

			} else {
				if inLowMapQ {
					// Exiting a LowMapQ region so save previous Region if
					// it passes the minimum region size
					pos, err := strconv.Atoi(fields[qpv1.Position])
					if err != nil {
						return fmt.Errorf("error converting Position: %s",
							fields[qpv1.Position])
					}
					if pos-lowMapQStart > flagRegionLength {
						rctr++
						_, err =
							ow.WriteString(makeLowMapqGffRecord(fields[qpv1.Reference],
								lowMapQStart, pos, rctr, lowMapQTotal) + "\n")
						if err != nil {
							return err
						}
					}

					inLowMapQ = false
					lowMapQStart = 0
					lowMapQTotal = 0
				}
			}

			// Sanity check - no file can contain multiple references
			if prevRef != "" && prevRef != fields[qpv1.Reference] {
				return fmt.Errorf("file %s contains multiple references at line %d : %s, %s",
					file, lctr, prevRef, fields[qpv1.Reference])
			}
			prevRef = fields[qpv1.Reference]
		}

		// If we were in a LowMapQ region check to see if it passes the
		// the minimum region size
		if inLowMapQ {
			// Exiting a LowMapQ region so save previous Region
			pos, err := strconv.Atoi(fields[qpv1.Position])
			if err != nil {
				return fmt.Errorf("error converting Position: %s",
					fields[qpv1.Position])
			}
			if pos-lowMapQStart > flagRegionLength {
				rctr++
				_, err =
					ow.WriteString(makeLowMapqGffRecord(fields[qpv1.Reference],
						lowMapQStart, pos, rctr, lowMapQTotal) + "\n")
				if err != nil {
					return err
				}
			}
		}

		if sctr != 0 {
			log.Warnf("  %d lines of %d were short - fewer than 33 fields", sctr, lctr)
		}
	}

	// Log Mapq tallys
	var sorted []int
	var total int
	for k, v := range lmqTally {
		sorted = append(sorted, k)
		total += v
	}
	log.Infof("Tally of average mapping quality score (total=%d):", total)
	log.Info("  MapQ\tCount\tPercent")
	sort.Ints(sorted)
	for _, i := range sorted {
		log.Infof("  %d\t%d\t%.3f", i, lmqTally[i], float64(lmqTally[i])*100/float64(total))
	}

	return nil
}

// Note that what is passed in is the current base in the sequence, not
// the start - it's only when process the first base PAST the
// homopolymer that we know it has ended. So we need to do the math
// based on the current location and the repeat length and remembering
// that we are using half-open numbering and that GFF3 coordinates are
// supposed to be is 1-based and we are working from a 0-based loop.
func makeLowMapqGffRecord(seq string, low, high, ctr, mapqTotal int) string {
	avgmapq := mapqTotal / (high - low)
	gff3fields := []string{
		seq,
		`ajgo:low-mapq`,
		`remark`,
		strconv.Itoa(low),
		strconv.Itoa(high),
		`.`,
		`.`,
		`.`,
		`ID=lowmpaq` + strconv.Itoa(ctr) +
			`;length=` + strconv.Itoa(high-low) +
			`;avgmapq=` + strconv.Itoa(avgmapq)}
	line := strings.Join(gff3fields, "\t")
	return line
}
