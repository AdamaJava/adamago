package ngc

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// This file contains data structures and code for NGScheck basic mode
// outputs in JSON format.

type NgscheckBasic struct {
	Ngscheck         Ngscheck         `json:"ngscheck"`
	Qprofiler2       Qprofiler2       `json:"qprofiler2"`
	Temp1            TempScorecard    `json:"Scorecard"`
	Scorecard        Scorecard        // populated via Finalise()
	ExecutionLog     ExecutionLog     `json:"Execution Log"`
	SummaryMetrics   SummaryMetrics   `json:"Summary Metrics"`
	BaseQualities    []BaseQuality    `json:"Base Quality"`
	ReadGroups       []ReadGroup      `json:"Read group properties"`
	BasesLosts       []BasesLost      `json:"Bases Lost"`
	Qnames           []Qname          `json:"QNAME"`
	Pairs            []Pairs          `json:"Pairs"`
	GenderPrediction GenderPrediction `json:"Gender Prediction"`
	ReadDepth        ReadDepth        `json:"Read Depth"`
}

type Ngscheck struct {
	ReportUuid string `json:"reportUuid"`
	RunBy      string `json:"runBy"`
	ReportDate string `json:"reportDate"`
	Version    string `json:"version"`
	InputXml   string `json:"inputXmlFilename"`
	OutputPdf  string `json:"outputPdfFilename"`
	OutputXml  string `json:"outputXmlFilename"`
}

// These parameters come from the qprofile2 files that was parsed by
// NGScheck.
type Qprofiler2 struct {
	RunBy         string `json:"runBy"`
	StartTime     string `json:"startTime"`
	FinishTime    string `json:"finishTime"`
	OS            string `json:"operatingSystem"`
	Version       string `json:"version"`
	Bam           string `json:"bamFilename"`
	BamMd5sum     string `json:"bamMd5sum"`
	BamReportUuid string `json:"bamReportUuid"`
}

// The Scorecard JSON structure is unfortunate in that it is composed of
// 2-element lists but many of those lists have a numeric Value and a
// string Score. This is OK in JSON which allows a list to hold values
// of different types but go does NOT allow this so we cannot unmarshall
// the easy way. Instead we are going to have to capture all of these
// mixed lists as []interface{} and use type assertions to convert each
// list item individually into its final form.
type TempScorecard struct {
	PredictedGenome      []string      `json:"Predicted Genome"`
	PredictedMolecMatch  []string      `json:"Predicted Assembled Molecules Matched"`
	PredictedGender      []string      `json:"Predicted Gender"`
	PredictedSeqType     []string      `json:"Predicted Sequencing Type"`
	PredictedSeqPlatform []string      `json:"Predicted Sequencing Platform"`
	PredictedSampleType  []string      `json:"Predicted Tumour/Normal"`
	Q30Pct               []interface{} `json:"Q30 %"`
	UnmappedReadPct      []interface{} `json:"Unmapped Read %"`
	DuplicateReadPct     []interface{} `json:"Duplicate Read %"`
	BasesLostAnalysisPct []interface{} `json:"Bases lost to analysis %"`
	NotProperPairPct     []interface{} `json:"Not Proper Pair %"`
	ClippedPct           []interface{} `json:"Clipped %"`
	OverlapBasesPct      []interface{} `json:"Overlap Bases %"`
	BasesForAnalysis     []interface{} `json:"Bases remaining for analysis"`
	RefSequenceLength    []interface{} `json:"Reference sequence total length from BAM header"`
	AverageReadDepth     []interface{} `json:"Average read depth"`
	CyclesErrorAbove1Pct []interface{} `json:"Number of cycles > 1% mismatches"`
	AverageMismatchR1Pct []interface{} `json:"Average mismatch % - first read in pair"`
	AverageMismatchR2Pct []interface{} `json:"Average mismatch % - second read in pair"`
	CorrelnReadDepthGC   []interface{} `json:"Pearson correlation of read depth and GC Content"`
}

type ExecutionLog map[string]string

// This data represents the 3 column scorecard table on the front page
// of NGScheck. The 3 columns are effectively ParameterName, Value,
// Score. Score is a string with values like "Pass", "Warn1", etc. Not
// all parameters have a Score so it is not unusual to have an empty
// string for Score.
//
// This structure does NOT come from the JSON. Because of the way the
// Scorecard is represented in JSON, we need to unmarshall the Scorecard
// into a temporary data structure composed of fields of type interface{}.
// The Finalise function uses type switching to populate the Scorecard
// struct from the TempScorecard struct.
type Scorecard struct {
	PredictedGenome      string
	PredictedMolecMatch  string
	PredictedGender      string
	PredictedSeqType     string
	PredictedSeqPlatform string
	PredictedSampleType  string
	// These types must be populated from Temp1 via a call to Finalise()
	Q30Pct                    float64
	Q30PctScore               string
	UnmappedReadPct           float64
	UnmappedReadPctScore      string
	DuplicateReadPct          float64
	DuplicateReadPctScore     string
	BasesLostAnalysisPct      float64
	BasesLostAnalysisPctScore string
	NotProperPairPct          float64
	NotProperPairPctScore     string
	ClippedPct                float64
	ClippedPctScore           string
	OverlapBasesPct           float64
	OverlapBasesPctScore      string
	BasesForAnalysis          int64
	RefSequenceLength         int64
	AverageReadDepth          float64
	CyclesErrorAbove1Pct      string
	CyclesErrorAbove1PctScore string
	AverageMismatchR1Pct      float64
	AverageMismatchR1PctScore string
	AverageMismatchR2Pct      float64
	AverageMismatchR2PctScore string
	CorrelnReadDepthGC        float64
	CorrelnReadDepthGCScore   string
}

type SummaryMetrics struct {
	AvgLengthFirstReadInPair  int `json:"Average length of first-of-pair reads"`
	AvgLengthSecondReadInPair int `json:"Average length of second-of-pair reads"`
	// We cannot seem to capture this field. After some experimentation, the
	// commas are the problem - they have special meaning in JSON tags.
	// I would rather not capture it than have it (incorrectly) appear as 0.
	//DiscardedReads            int `json:"Discarded reads (FailedVendorQuality, secondary, supplementary)"`
	TotalReads     int `json:"Total reads including discarded reads"`
	ReadGroupCount int `json:"Read Group Count"`
}

type BaseQuality struct {
	Read string  `json:"Read"`
	Q10  float64 `json:"Q10"`
	Q20  float64 `json:"Q20"`
	Q30  float64 `json:"Q30"`
	Q40  float64 `json:"Q40"`
}

type ReadGroup struct {
	ReadGroup  string `json:"Read Group"`
	ReadCount  int64  `json:"Read Count"`
	ReadLength string `json:"Avg/Max Read Length"`
	TLEN       string `json:"Mode/Median TLEN"`
}

type BasesLost struct {
	ReadGroup          string  `json:"Read Group"`
	UnmappedReadsPct   float64 `json:"Unmapped Reads"`
	DuplicateReadsPct  float64 `json:"Duplicate Reads"`
	NotProperPairsPct  float64 `json:"Not Proper Pairs"`
	AdaptorTrimmingPct float64 `json:"Adaptor Trimming"`
	SoftClippedPct     float64 `json:"Soft Clipped"`
	HardClippedPct     float64 `json:"Hard Clipped"`
	OverlapBasesPct    float64 `json:"Overlap Bases"`
	TotalBasesLostPct  float64 `json:"Total Bases Lost"`
}

type Qname struct {
	ReadGroup    string `json:"Read Group"`
	Instrument   string `json:"Instrument"`
	RunId        string `json:"Run Id"`
	FlowCellId   string `json:"Flow Cell Id"`
	FlowCellLane string `json:"Flow Cell Lane"`
	TileNumber   string
	Temp1        interface{} `json:"Tile Number"`
}

type Pairs struct {
	ReadGroup        string  `json:"Read Group"`
	ProperInward     float64 `json:"Proper Inward"`
	ProperOutward    float64 `json:"Proper Outward"`
	NotProperF3F5    float64 `json:"Not Proper F3F5"`
	NotProperF5F3    float64 `json:"Not Proper F5F3"`
	NotProperInward  float64 `json:"Not Proper Inward"`
	NotProperOutward float64 `json:"Not Proper Outward"`
	NotProperOther   float64 `json:"Not Proper Other"`
	Total            float64 `json:"Total"`
}

type GenderPrediction struct {
	Gender           string  `json:"Gender"`
	Method           string  `json:"Method"`
	Threshold        int64   `json:"Threshold"`
	ChrXReadCount    int64   `json:"chrX Read Count"`
	ChrYReadCount    int64   `json:"chrY Read Count"`
	ChrXLength       int64   `json:"chrX Length"`
	ChrYLength       int64   `json:"chrY Length"`
	ChrXYReadRatio   float64 `json:"chrX chrY Read Ratio"`
	ReadsRatioGender string  `json:"Reads Ratio Gender"`
	ChrXDensityPeak  float64 `json:"chrX Density Peak"`
	ChrYDensityPeak  float64 `json:"chrY Density Peak"`
	ChrXYPeakRatio   float64 `json:"chrX/chrY Peak Ratio"`
}

type ReadDepth struct {
	GCCorrelation    float64 `json:"GC Content Pearson correlation"`
	Mean             float64 `json:"Mean"`
	StdDev           float64 `json:"Standard deviation"`
	QCoeffDispersion float64 `json:"Quartile Coefficient of Dispersion"`
	Autocorrelation  float64 `json:"Autocorrelation"`
	CoeffVariation   float64 `json:"Coefficient of Variation"`
}

// Finalise helps deal with the myriad problems in the JSON, especially
// in the Scorecard. As a consequence, we have had to unmarshall a
// significant number of fields into interface{} which is difficult to use.
//
// In the routine, we will use type assertion switching to work out what
// sort of value was in the JSON and turn it into something we can work
// with. You do not have to call Finalise() but it is highly
// recommended, especially if you want to use data from Scorecard which
// is where all of the predictions appear.
func (n *NgscheckBasic) Finalise() {
	// Some of this is simple assignment but a lot requires type assertions
    // and type switches to convert from interface{} and []interface{}.
	s := &n.Scorecard
	t := &n.Temp1

	s.PredictedGenome = t.PredictedGenome[0]
	s.PredictedMolecMatch = t.PredictedMolecMatch[0]
	s.PredictedGender = t.PredictedGender[0]
	s.PredictedSeqType = t.PredictedSeqType[0]
	s.PredictedSeqPlatform = t.PredictedSeqPlatform[0]
	s.PredictedSampleType = t.PredictedSampleType[0]

	s.Q30Pct = t.Q30Pct[0].(float64)
	s.Q30PctScore = t.Q30Pct[1].(string)
	s.UnmappedReadPct = t.UnmappedReadPct[0].(float64)
	s.UnmappedReadPctScore = t.UnmappedReadPct[1].(string)
	s.DuplicateReadPct = t.DuplicateReadPct[0].(float64)
	s.DuplicateReadPctScore = t.DuplicateReadPct[1].(string)
	s.BasesLostAnalysisPct = t.BasesLostAnalysisPct[0].(float64)
	s.BasesLostAnalysisPctScore = t.BasesLostAnalysisPct[1].(string)
	s.NotProperPairPctScore = t.NotProperPairPct[1].(string)
	s.ClippedPct = t.ClippedPct[0].(float64)
	s.ClippedPctScore = t.ClippedPct[1].(string)
	s.OverlapBasesPctScore = t.OverlapBasesPct[1].(string)
	s.CyclesErrorAbove1Pct = t.CyclesErrorAbove1Pct[0].(string)
	s.CyclesErrorAbove1PctScore = t.CyclesErrorAbove1Pct[1].(string)
	s.AverageMismatchR1Pct = t.AverageMismatchR1Pct[0].(float64)
	s.AverageMismatchR1PctScore = t.AverageMismatchR1Pct[1].(string)
	s.AverageMismatchR2PctScore = t.AverageMismatchR2Pct[1].(string)
	s.CorrelnReadDepthGC = t.CorrelnReadDepthGC[0].(float64)
	s.CorrelnReadDepthGCScore = t.CorrelnReadDepthGC[1].(string)

	// This looks awful but I cannot seem to directly assert .(int64) on
	// these next 2 (compiler is OK but panics at runtime) so I have to go
	// the long way round by asserting float64 and converting to int64.
	s.BasesForAnalysis = int64(math.Round(t.BasesForAnalysis[0].(float64)))
	s.RefSequenceLength = int64(math.Round(t.RefSequenceLength[0].(float64)))
	// If there is an error during ParseFloat, we will do nothing which
	// will leave s.AverageReadDepth with a sensible value - the default 0
	tmp1 := t.AverageReadDepth[0].(string)
	tmp1 = strings.TrimSuffix(tmp1, "x") // pesky "x" e.g. 31.3x
	f, err := strconv.ParseFloat(tmp1, 32)
	if err == nil {
		s.AverageReadDepth = f
	}

	// For single-end sequencing, these 3 fields appear in the JSON as ""
    // i.e. empty strings, rather than the floats they are in paired-end
    // sequencing. This screws up our unmarshall so we will have to use
    // type switching on the Temp1 values to work out what type was
    // assigned and use type assertions to turn that into float for
    // populating the Scorecard value. In practice we ignore the empty
    // string case because that will leave the default value of 0.0
    // in Scorecard.NotProperPairPct which is fine.
	switch v := t.NotProperPairPct[0].(type) {
	case float64:
		s.NotProperPairPct = v
	}

	switch v := t.OverlapBasesPct[0].(type) {
	case float64:
		s.OverlapBasesPct = v
	}

	switch v := t.AverageMismatchR2Pct[0].(type) {
	case float64:
		s.AverageMismatchR2Pct = t.AverageMismatchR2Pct[0].(float64)
	default:
		_ = fmt.Sprintf("%v", v)
	}

	// Qname.`Tile Number` can be a string or an integer. Sigh. There may
	// be 0 or more Qname in Qnames so loop
	for i, _ := range n.Qnames {
        // Do this or we will be working on a copy so changes will not stick!
        q := &n.Qnames[i]
		switch v := q.Temp1.(type) {
		case int64:
			q.TileNumber = strconv.FormatInt(v, 10)
		case float64:
			q.TileNumber = strconv.FormatFloat(v, 'f', 0, 32)
		case string:
			q.TileNumber = v
		default:
			q.TileNumber = fmt.Sprintf("Finalise: unknown type: %v", v)
		}
	}
}
