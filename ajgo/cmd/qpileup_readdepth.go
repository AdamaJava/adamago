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
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/grendeloz/cmdh"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	flagBamCount int
	flagAbove    bool
	flagBelow    bool
)

// submode qpileup > read-depth
var qpileupReaddepthCmd = &cobra.Command{
	Use:   "read-depth",
	Short: "identifies regions of unusual read-depth",
	Long: `
Takes one or more qpileup view files and looks for regions
that have read depths above or below a specified threshold. The qpileup
view reports must be of a particular format and all inputs will be 
checked that they have the required columns in the required order.

The user can specify threshold read depth (above or below which 
positions will be reported) and a minimum reportable length for the 
regions. The minimum size for region length is important otherwise
the GFF3 report can become noisy with very short regions.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdh.StartLogging()
		qpileupReaddepthCmdRun(cmd, args)
		cmdh.FinishLogging()
	},
}

func init() {
	qpileupCmd.AddCommand(qpileupReaddepthCmd)

	qpileupReaddepthCmd.Flags().StringSliceVar(&flagViewFiles, "view", []string{},
		"qpileup view mode data extract file(s)")
	qpileupReaddepthCmd.Flags().StringVar(&flagFilelistFile, "viewlist", "",
		"text file listing qpileup view fles for analysis")

	qpileupReaddepthCmd.Flags().IntVar(&flagBamCount, "bam-count", 0,
		"count of BAM files in qpileup h5 that we have view files from")
	qpileupReaddepthCmd.MarkFlagRequired("bam-count")

	qpileupReaddepthCmd.Flags().IntVar(&flagThreshold, "threshold", 0,
		"positions with read depth above/below this number will be reported")
	qpileupReaddepthCmd.Flags().BoolVar(&flagAbove, "above", false,
		"set to true for reporting above --threshold")
	qpileupReaddepthCmd.Flags().BoolVar(&flagBelow, "below", false,
		"set to true for reporting below --threshold")
	qpileupReaddepthCmd.Flags().IntVar(&flagRegionLength, "region-min", 100,
		"minimum reportable region size")

	qpileupReaddepthCmd.Flags().StringVar(&flagOutfile, "gff3", "",
		"output GFF3 file containing regions")
}

func qpileupReaddepthCmdRun(cmd *cobra.Command, args []string) {

	// One and only one of --above and --below
	if flagAbove && flagBelow {
		log.Fatal("only one of --above and --below can be specified")
	}
	if !flagAbove && !flagBelow {
		log.Fatal("one of --above or --below must be specified")
	}

	// log key parameters
	if flagAbove {
		log.Infof("  --above %d", flagThreshold)
	} else if flagBelow {
		log.Infof("  --below %d", flagThreshold)
	}
	log.Infof("  --region-min %d", flagRegionLength)
	log.Infof("  --bam-count %d", flagBamCount)

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

	err = ProcessReaddepth(ViewFiles)
	if err != nil {
		log.Fatal(err)
	}
}

// Process is a beast and way bigger than I'd like but there it is.
// Because the files are so huge and there are potentially so many regions,
// we can't afford to save up "hits" and report at the end - we must report
// as we go.
func ProcessReaddepth(files []string) error {
	// Open GFF3 files for reporting
	log.Info("writing read-depth GFF3 file: ", flagOutfile)
	of, err := os.Create(flagOutfile)
	if err != nil {
		return err
	}
	defer of.Close()
	ow := bufio.NewWriter(of)
	defer ow.Flush()

	// Setup our test function
	testFunc := func(rd int) bool { return float64(rd/flagBamCount) > float64(flagThreshold) }
	if flagBelow {
		testFunc = func(rd int) bool { return float64(rd/flagBamCount) < float64(flagThreshold) }
	}

	// Write GFF3 header including report files
	header := "##gff-version 3\n"
	if flagAbove {
		header += "##content regions where average read depth is above threshold - from qpileup view file(s)\n"
	} else if flagBelow {
		header += "##content regions where average read depth is below threshold - from qpileup view file(s)\n"
	}
	header += "##threshold " + strconv.Itoa(flagThreshold) + "\n"
	header += "##region-min " + strconv.Itoa(flagRegionLength) + "\n"
	header += "##bam-count " + strconv.Itoa(flagBamCount) + "\n"
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
		var inRegion bool
		var regionStart int
		var regionDepth int
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
				readdepth := rf + nrf + rr + nrr
				if testFunc(readdepth) {
					inRegion = true
					pos, err := strconv.Atoi(fields[qpv1.Position])
					if err != nil {
						return fmt.Errorf("error converting Position: %s",
							fields[qpv1.Position])
					}
					regionDepth += readdepth
					regionStart = pos
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
			// If too few fields are present then the referencing later will
			// cause a panic so skip (but count) any short lines.
			if len(fields) < 33 {
				//log.Warnf("  line %d has only %d fields: %s", lctr, len(fields), line)
				sctr++
				continue
			}
			// Unavoidable ugliness - math with strings
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
			readdepth := rf + nrf + rr + nrr
			if fields[qpv1.Ref_base] == `N` {
				// Hitting an N always stops a region
				if inRegion {
					// Exiting a region so save if it passes the minimum region size
					pos, err := strconv.Atoi(fields[qpv1.Position])
					if err != nil {
						return fmt.Errorf("error converting Position: %s",
							fields[qpv1.Position])
					}
					if pos-regionStart > flagRegionLength {
						rctr++
						_, err =
							ow.WriteString(makeReaddepthGffRecord(fields[qpv1.Reference],
								regionStart, pos, rctr, regionDepth) + "\n")
						if err != nil {
							return err
						}
					}
					inRegion = false
					regionDepth = 0
					regionStart = 0
				}
			} else if testFunc(readdepth) {
				// Start tracking LowMapQ region if not already doing so
				if !inRegion {
					inRegion = true
					pos, err := strconv.Atoi(fields[qpv1.Position])
					if err != nil {
						return fmt.Errorf("error converting Position: %s",
							fields[qpv1.Position])
					}
					regionDepth = readdepth
					regionStart = pos
				} else {
					regionDepth += readdepth
				}
			} else {
				if inRegion {
					// Exiting a region so save if it passes the minimum region size
					pos, err := strconv.Atoi(fields[qpv1.Position])
					if err != nil {
						return fmt.Errorf("error converting Position: %s",
							fields[qpv1.Position])
					}
					if pos-regionStart > flagRegionLength {
						rctr++
						_, err =
							ow.WriteString(makeReaddepthGffRecord(fields[qpv1.Reference],
								regionStart, pos, rctr, regionDepth) + "\n")
						if err != nil {
							return err
						}
					}

					inRegion = false
					regionDepth = 0
					regionStart = 0
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
		if inRegion {
			// Exiting a LowMapQ region so save previous Region
			pos, err := strconv.Atoi(fields[qpv1.Position])
			if err != nil {
				return fmt.Errorf("error converting Position: %s",
					fields[qpv1.Position])
			}
			if pos-regionStart > flagRegionLength {
				rctr++
				_, err =
					ow.WriteString(makeReaddepthGffRecord(fields[qpv1.Reference],
						regionStart, pos, rctr, regionDepth) + "\n")
				if err != nil {
					return err
				}
			}
		}

		if sctr != 0 {
			log.Warnf("  %d lines of %d were short - fewer than 33 fields", sctr, lctr)
		}
	}

	return nil
}

// Remember that we are using half-open numbering and that GFF3 coords
// are supposed to be 1-based and we are working from a 0-based loop.
func makeReaddepthGffRecord(seq string, low, high, ctr, depth int) string {
	test := ""
	if flagAbove {
		test = "above-" + strconv.Itoa(flagThreshold)
	} else if flagBelow {
		test = "below-" + strconv.Itoa(flagThreshold)
	} else {
		test = "bug-please-report"
	}
	// depth is the sum of all bases at all positions in this region so
	// we must divide by the length of the region and the number of BAMs.
	avgdepth := depth / (high - low) / flagBamCount
	gff3fields := []string{
		seq,
		`ajgo:read-depth`,
		`remark`,
		strconv.Itoa(low),
		strconv.Itoa(high),
		`.`,
		`.`,
		`.`,
		`ID=readdepth` + strconv.Itoa(ctr) +
			`;test=` + test +
			`;length=` + strconv.Itoa(high-low) +
			`;avgdepth=` + strconv.Itoa(avgdepth)}
	line := strings.Join(gff3fields, "\t")
	return line
}
