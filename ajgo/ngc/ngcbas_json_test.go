package ngc

import (
	"encoding/json"
	"io"
	"os"
	"testing"
)

// TestNgscheckBasic test unmarshalling of NGScheck basic mode JSON
func TestNgscheckBasic(t *testing.T) {
	// Open our JSON
	f :=
		`testdata/e9d6b832-1046-439d-9915-405d812e9712.bam.qp2.xml.ngcbas.json`
	jsonFile, err := os.Open(f)
	if err != nil {
		t.Fatalf(`error opening file %s: %v`, f, err)
	}
	defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
	byteValue, _ := io.ReadAll(jsonFile)

	// initialize our NgscheckBasic structure and unmarshal
	var ngb NgscheckBasic
	err = json.Unmarshal(byteValue, &ngb)
	if err != nil {
		t.Fatalf(`error while unmarshalling JSON: %v`, err)
	}
	(&ngb).Finalise()

	e1 := `ab72943f-43a4-4430-8198-24e62696d60d`
	g1 := ngb.Ngscheck.ReportUuid
	if e1 != g1 {
		t.Fatalf(`ngb.Ngscheck.ReportUuid should be %v but is %v`, e1, g1)
	}

	e5 := `CE772AF3862F03B1B13E4410E54BD01A`
	g5 := ngb.Qprofiler2.BamMd5sum
	if e5 != g5 {
		t.Fatalf(`ngb.Qprofiler2.BamMd5sum should be %v but is %v`, e5, g5)
	}
	e6 := `Linux`
	g6 := ngb.Qprofiler2.OS
	if e6 != g6 {
		t.Fatalf(`ngb.Qprofiler2.OS should be %v but is %v`, e6, g6)
	}

	e10 := `Human GRCh37`
	g10 := ngb.Scorecard.PredictedGenome
	if e10 != g10 {
		t.Fatalf(`ngb.Scorecard.PredictedGenome should be %v but is %v`, e10, g10)
	}
	e11 := 88.14
	g11 := ngb.Scorecard.Q30Pct
	if e11 != g11 {
		t.Fatalf(`ngb.Scorecard.Q30Pct should be %v but is %v`, e11, g11)
	}
	e12 := `Pass`
	g12 := ngb.Scorecard.Q30PctScore
	if e12 != g12 {
		t.Fatalf(`ngb.Scorecard.Q30PctScore should be %v but is %v`, e12, g12)
	}

	e100 := 368606.7680769231
	g100 := ngb.ReadDepth.Mean
	if e100 != g100 {
		t.Fatalf(`ngb.ReadDepth.Mean should be %v but is %v`, e100, g100)
	}
}
