package cmd

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"

	ajxml "github.com/adamajava/adamago/xml"
	"github.com/grendeloz/cmdh"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var qmotifXmlFile string

// summaryCmd represents the summary command
var qmotifSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "print summary values from qmotif XML files",
	Long: `
Parse qmotif XML files and write out parameters from the summary section.
The values are written as a single tab-separated line.

Note that qmotif must NOT have been run in includes-only mode. The INI
file can have the includes defined (which will trigger reporting for the
includes) but the includes-only flag must have been set to false.  The 
reason is that, even if we report by includes region, the whole BAM must 
have been traversed so that the total number of reads is recorded. This is
important because we can't compare telomeric read counts unless we scale
them according to the size of the BAM.

For example, a 60x tumour BAM typically contains twice as many reads as
a 30x normal BAM and so unsurprisingly will have approximately twice as many
telomeric reads. In order to compare the relative amount of telomeric reads
in the tumour and normal BAMs, we need to scale both raw read counts to
a common read count - for qmotif all scaling is to a nominal 1B reads.
So a BAM with 500M reads would have its raw telomeric read count doubled
during scaling and a BAM with 2B reads would have its raw telomeric read
count halved during scaling.

The columns in the report line are:

 1.  qmotif-version
 2.  TotalReads
 3.  NoOfMotifs
 4.  RawUnmapped
 5.  RawIncludes
 6.  RawGenomic
 7.  ScaledUnmapped
 8.  ScaledIncludes
 9.  ScaledGenomic
 10. BasesInMotifs
 11. BAM-name

Note that the BAM name is taken out of the qmotif XML file so it will
be the full pathname of the BAM as it was when qmotif was run against
it.
`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdh.StartLogging()
		summaryQmotifCmdRun(cmd, args)
		cmdh.FinishLogging()
	},
}

func init() {
	qmotifCmd.AddCommand(qmotifSummaryCmd)

	qmotifSummaryCmd.Flags().StringVar(&qmotifXmlFile, "xmlfile", "",
		"qmotif XML file to be parsed")
	qmotifSummaryCmd.MarkFlagRequired("xmlfile")
}

func summaryQmotifCmdRun(cmd *cobra.Command, args []string) {
	// Open our xmlFile
	xmlFile, err := os.Open(qmotifXmlFile)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("processing: ", qmotifXmlFile)
	defer xmlFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := io.ReadAll(xmlFile)

	// initialize our qmotif structure
	var qmotif ajxml.Qmotif
	// unmarshal our byteArray into the qmotif data structure
	err = xml.Unmarshal(byteValue, &qmotif)
	if err != nil {
		log.Fatal("error while unmarshalling: ", err)
	}

	vals := []string{qmotif.Version,
		qmotif.Summary.Counts.TotalReads.Count,
		qmotif.Summary.Counts.NoOfMotifs.Count,
		qmotif.Summary.Counts.RawUnmapped.Count,
		qmotif.Summary.Counts.RawIncludes.Count,
		qmotif.Summary.Counts.RawGenomic.Count,
		qmotif.Summary.Counts.ScaledUnmapped.Count,
		qmotif.Summary.Counts.ScaledIncludes.Count,
		qmotif.Summary.Counts.ScaledGenomic.Count,
		qmotif.Summary.Counts.BasesInMotifs.Count,
		qmotif.Summary.Bam}

	fmt.Println(strings.Join(vals, string('\t')))
}
