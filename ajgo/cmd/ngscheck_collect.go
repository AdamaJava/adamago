package cmd

import (
	"ajgo/ngc"
	"bufio"
	"encoding/json"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
    "github.com/grendeloz/cmdh"
	"github.com/spf13/cobra"
)

// ngscheckCollectCmd represents the summary command
var ngscheckCollectCmd = &cobra.Command{
	Use:   "collect",
	Short: "collect NGScheck info for a list of BAM files",
	Long: `Collects metrics from NGScheck basic mode JSON files.

The --file-list file is a plain-text (unix line-endings) file that
contains a list of filenames, one per line - absolute pathnames are
recommended. If the filename ends in .bam, the string from --suffix
(default: .qp2.xml.ngcbas.json) is appended to the filename to create an
extrapolated JSON name, otherwise the file is assumed to be a JSON file.

If you are going to be matching data from the NGScheck files against the
relevant BAM files, it can be advantageous to list BAM files in
--file-list. The filename from --file-list is always the one listed in the
report regardless of whether the software used --suffix to extrapolate
the JSON name. The --suffix system only works if you have a consistent
pattern for naming your NGScheck basic mode JSON files and it is
directly based on (and is an extension of) the BAM file name.

Each JSON file is assumed to be the output from NGScheck basic mode. If
other types of valid JSON file are supplied, the process should still
work but no data will be extracted. If a file cannot be opened, it will
still be listed in the output report but the line will contain the
filename and a string "false" to show that the JSON file was not found,
and the rest of the line will consist of empty fields. If a file cannot
be opened, a warning will also be written to the log file.

NGScheck basic mode extracts metrics from qprofiler2 reports on BAM
files so some of the reported information is about qprofiler2. All
fields have "Ngc_" as a prefix so the provenance of these columns is
clear even if this data file is merged with other files.

The report contains the following columns:

  Ngc_File - the name of the file as found in --file-list
  Ngc_JsonFound - was the JSON file openable [true|false]
  Ngc_Uuid - NGScheck report UUID
  Ngc_Version - NGScheck software version
  Ngc_Qp2_Uuid - qprofiler2 report UUID
  Ngc_Qp2_Version - qprofiler2 software version
  Ngc_Qp2_Bam - BAM file that the qprofiler2 report related to
  Ngc_PredGenome - predicted genome
  Ngc_PredMolec - for predicted genome, how many molecules matched
  Ngc_PredGender - predicted gender
  Ngc_PredSeqType - predicted sequencing type, e.g. PE_150_150
  Ngc_PredSeqPlatform - predicted sequencing platform
  Ngc_PredSampleType - predicted sample type [Normal|Tumour]
  Ngc_TotalReads - total reads
  Ngc_ReadGroupCount - count of ReadGroups found
  Ngc_AvgLenR1 - average length of the first read in pair
  Ngc_AvgLenR2 - average length of the second read in pair
  Ngc_Q10Pct - percentage of reads mapped at Q10 or higher
  Ngc_Q20Pct - percentage of reads mapped at Q20 or higher
  Ngc_Q30Pct - percentage of reads mapped at Q30 or higher
  Ngc_Q40Pct - percentage of reads mapped at Q40 or higher
  Ngc_UnmappedReadPct - percentage of reads unmapped
  Ngc_DuplicateReadPct - percentage of reads that are marked duplicate
  Ngc_BasesLostPct - percentage of bases lost to analysis
  Ngc_NotProperPairPct - percentage of pairs that are "not proper"
  Ngc_ClippedPct - precentage of reads that are clipped
  Ngc_OverlapBasesPct - percentage of bases that are overlapping
  Ngc_BasesForAnalysis - bases available for analysis
  Ngc_RefSeqLength - length of reference sequence (from BAM header)
  Ngc_AvgReadDepth - average read depth: BasesForAnalysis/RefSeqLength
  Ngc_CyclesErrorAbove1Pct - count of cycles with error rate > 1%
  Ngc_AvgMismatchR1Pct - average mismatch % for first read in pair
  Ngc_AvgMismatchR2Pct - average mismatch % for second read in pair
  Ngc_CorrelnReadDepthGC - correlation between read depth and GC%
  Ngc_RG0_Name - for first readgroup, name
  Ngc_RG0_Instrument - for first readgroup, instrument name
  Ngc_RG0_RunId - for first readgroup, run id
  Ngc_RG0_FlowCellId - for first readgroup, flow cell id
`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdh.StartLogging()
		ngscheckCollectCmdRun(cmd, args)
		cmdh.FinishLogging()
	},
}

var flagSuffix string

func init() {
	ngscheckCmd.AddCommand(ngscheckCollectCmd)

	ngscheckCollectCmd.Flags().StringVar(&flagFilelistFile, "file-list", "",
		"text file containing BAM/JSON absolute pathnames")
	ngscheckCollectCmd.MarkFlagRequired("file-list")

	ngscheckCollectCmd.Flags().StringVar(&flagSuffix, "suffix", ".qp2.xml.ngcbas.json",
		"bam suffix for json")

	ngscheckCollectCmd.Flags().StringVar(&flagOutfile, "out", "",
		"output text file")
	ngscheckCollectCmd.MarkFlagRequired("out")
}

func ngscheckCollectCmdRun(cmd *cobra.Command, args []string) {
	log.Info("--suffix ", flagSuffix)
	// Get list of BAM or JSON files
	files, err := LinesFromFile(flagFilelistFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Files to process: ", len(files))

	// Open output file
	f, err := os.Create(flagOutfile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	sep := "\t"

	header := strings.Join(FieldNames(), sep)
	_, err = w.WriteString(header + "\n")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		jfile := file
		// If file is a BAM, add suffix to get JSON name
		isBam, err := regexp.MatchString(`(?i).bam$`, file)
		if err != nil {
			log.Fatal(err)
		}
		if isBam {
			jfile = jfile + flagSuffix
		}

		// If we can't open the JSON file for any reason, write a
		// (mostly) empty line to the report and skip to next file
		log.Info("processing: ", jfile)
		j, err := os.Open(jfile)
		if err != nil {
			log.Warnf("error: %v", err)
			fields := []string{file, `false`}
			// Append 40 empty strings to pad empty line
			for i := 1; i < 41; i++ {
				fields = append(fields, ``)
			}
			_, err = w.WriteString(strings.Join(fields, sep) + "\n")
			if err != nil {
				log.Fatal(err)
			}
			continue
		}
		byteValue, _ := io.ReadAll(j)

		// unmarshall our NgscheckBasic structure
		var ngb ngc.NgscheckBasic
		err = json.Unmarshal(byteValue, &ngb)
		if err != nil {
			log.Fatal("error while unmarshalling: ", err)
		}
		(&ngb).Finalise()

		// log.Infof("data: %+v",ngb)

		// Write parsed fields
		fields := []string{file, `true`}
		fields = append(fields, AssembleFields(&ngb)...)
		_, err = w.WriteString(strings.Join(fields, sep) + "\n")
		if err != nil {
			log.Fatal(err)
		}

		// No defer or we get too many files open
		j.Close()
	}
}

// If you add any extra fields to this list, don't forget to also change
// the code above (around line 156) that pads empty lines so they still
// contain the right number of (empty) fields - some CSV readers will not
// cope if some lines have fewer fields. You will also need to modify
// FieldNames() to add appropriate column names in matching order.
func AssembleFields(n *ngc.NgscheckBasic) []string {
	var fs []string
	fs = append(fs, n.Ngscheck.ReportUuid) // 1
	fs = append(fs, n.Ngscheck.Version)
	fs = append(fs, n.Qprofiler2.BamReportUuid)
	fs = append(fs, n.Qprofiler2.Version)
	fs = append(fs, n.Qprofiler2.Bam)
	fs = append(fs, n.Scorecard.PredictedGenome)
	fs = append(fs, n.Scorecard.PredictedMolecMatch)
	fs = append(fs, n.Scorecard.PredictedGender)
	fs = append(fs, n.Scorecard.PredictedSeqType)
	fs = append(fs, n.Scorecard.PredictedSeqPlatform) // 10
	fs = append(fs, n.Scorecard.PredictedSampleType)
	fs = append(fs, iToS(n.SummaryMetrics.TotalReads))
	fs = append(fs, iToS(n.SummaryMetrics.ReadGroupCount))
	fs = append(fs, iToS(n.SummaryMetrics.AvgLengthFirstReadInPair))
	fs = append(fs, iToS(n.SummaryMetrics.AvgLengthSecondReadInPair))
	fs = append(fs, fToS(n.Scorecard.UnmappedReadPct))
	fs = append(fs, fToS(n.Scorecard.DuplicateReadPct))
	fs = append(fs, fToS(n.Scorecard.BasesLostAnalysisPct))
	fs = append(fs, fToS(n.Scorecard.NotProperPairPct))
	fs = append(fs, fToS(n.Scorecard.ClippedPct)) // 20
	fs = append(fs, fToS(n.Scorecard.OverlapBasesPct))
	fs = append(fs, iToS(n.Scorecard.BasesForAnalysis))
	fs = append(fs, iToS(n.Scorecard.RefSequenceLength))
	fs = append(fs, fToS(n.Scorecard.AverageReadDepth))
	fs = append(fs, n.Scorecard.CyclesErrorAbove1Pct)
	fs = append(fs, fToS(n.Scorecard.AverageMismatchR1Pct))
	fs = append(fs, fToS(n.Scorecard.AverageMismatchR2Pct))
	fs = append(fs, fToS(n.Scorecard.CorrelnReadDepthGC)) // 28
	if len(n.Qnames) > 0 {
		fs = append(fs, n.Qnames[0].ReadGroup,
			n.Qnames[0].Instrument,
			n.Qnames[0].RunId,
			n.Qnames[0].FlowCellId) // 32
	} else {
		fs = append(fs, ``, ``, ``, ``)
	}
	if n.BaseQualities[0].Read == `First` {
		fs = append(fs, fToS(n.BaseQualities[0].Q10),
			fToS(n.BaseQualities[0].Q20),
			fToS(n.BaseQualities[0].Q30),
			fToS(n.BaseQualities[0].Q40)) // 36
	} else {
		fs = append(fs, ``, ``, ``, ``)
	}
	if n.BaseQualities[1].Read == `Second` {
		fs = append(fs, fToS(n.BaseQualities[1].Q10),
			fToS(n.BaseQualities[1].Q20),
			fToS(n.BaseQualities[1].Q30),
			fToS(n.BaseQualities[1].Q40)) // 40
	} else {
		fs = append(fs, ``, ``, ``, ``)
	}
	return fs
}

func fToS(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 32)
}

// We need to cope with different integer types
func iToS(i interface{}) string {
	switch v := i.(type) {
	case int64:
		return strconv.FormatInt(v, 10)
	case int:
		return strconv.FormatInt(int64(v), 10)
	default:
		return ""
	}
}

func FieldNames() []string {
	return []string{
		`Ngc_File`,
		`Ngc_JsonFound`,
		`Ngc_Uuid`, // 1
		`Ngc_Version`,
		`Ngc_Qp2_Uuid`,
		`Ngc_Qp2_Version`,
		`Ngc_Qp2_Bam`,
		`Ngc_PredGenome`,
		`Ngc_PredMolec`,
		`Ngc_PredGender`,
		`Ngc_PredSeqType`,
		`Ngc_PredSeqPlatform`, // 10
		`Ngc_PredSampleType`,
		`Ngc_TotalReads`,
		`Ngc_ReadGroupCount`,
		`Ngc_AvgLenFirstReadInPair`,
		`Ngc_AvgLenSecondReadInPair`,
		`Ngc_UnmappedReadPct`,
		`Ngc_DuplicateReadPct`,
		`Ngc_BasesLostPct`,
		`Ngc_NotProperPairPct`,
		`Ngc_ClippedPct`, // 20
		`Ngc_OverlapBasesPct`,
		`Ngc_BasesForAnalysis`,
		`Ngc_RefSeqLength`,
		`Ngc_AvgReadDepth`,
		`Ngc_CyclesErrorAbove1Pct`,
		`Ngc_AvgMismatchRead1Pct`,
		`Ngc_AvgMismatchRead2Pct`,
		`Ngc_CorrelnReadDepthGC`,
		`Ngc_RG0_Name`,
		`Ngc_RG0_Instrument`, // 30
		`Ngc_RG0_RunId`,
		`Ngc_RG0_FlowCelId`,
		`Ngc_Q10Read1Pct`,
		`Ngc_Q20Read1Pct`,
		`Ngc_Q30Read1Pct`,
		`Ngc_Q40Read1Pct`,
		`Ngc_Q10Read2Pct`,
		`Ngc_Q20Read2Pct`,
		`Ngc_Q30Read2Pct`,
		`Ngc_Q40Read2Pct`} // 40
}
