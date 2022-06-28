package gff3

import (
	"fmt"
	"testing"
)

func TestNewGff3FromFile(t *testing.T) {
	_, err := NewFromFile(`testdata/no_headers.gff3`)
	if err == nil {
		t.Fatalf("NewGff3FromFile should have failed - file has no headers: %v", err)
	}

	_, err = NewFromFile(`testdata/no_records.gff3`)
	if err == nil {
		t.Fatalf("NewGff3FromFile should have failed - file has no records: %v", err)
	}

	_, err = NewFromFile(`testdata/no_version.gff3`)
	if err == nil {
		t.Fatalf("NewGff3FromFile should have failed - file has no gff version: %v", err)
	}

	_, err = NewFromFile(`testdata/wrong_version.gff3`)
	if err == nil {
		t.Fatalf("NewGff3FromFile should have failed - file has wrong gff version: %v", err)
	}
}

func TestGff3File1(t *testing.T) {
	f1 := `testdata/test1.gff3.gz`
	gff3, err := NewFromFile(f1)
	if err != nil {
		t.Fatalf("error reading %s: %v", f1, err)
	}

	e1 := 1150
	g1 := gff3.Features.Count()
	if e1 != g1 {
		t.Fatalf("%s should have %v Feature but has %v", f1, e1, g1)
	}

	eSeqids := []string{`1`, `2`, `3`}
	gSeqids := gff3.SeqIds()

	e2 := len(eSeqids)
	g2 := len(gSeqids)
	if e2 != g2 {
		t.Fatalf("%s should have %v SeqIds but has %v", f1, e2, g2)
	}

	e3 := eSeqids[0]
	g3 := gSeqids[0]
	if e3 != g3 {
		t.Fatalf("%s 0'th SeqId should be %v but is %v", f1, e3, g3)
	}

	e4 := eSeqids[1]
	g4 := gSeqids[1]
	if e4 != g4 {
		t.Fatalf("%s 1'th SeqId should be %v but is %v", f1, e4, g4)
	}

	e5 := eSeqids[2]
	g5 := gSeqids[2]
	if e5 != g5 {
		t.Fatalf("%s 2'th SeqId should be %v but is %v", f1, e5, g5)
	}

	// Test that we get an error when we expect an error
	e6 := eSeqids[1]
	g6 := gSeqids[2]
	if e6 == g6 {
		t.Fatalf("%s SeqIds 1 and 2 should be different but are %v and %v", f1, e6, g6)
	}

	// Test that .Features is as expected
	e7 := `file`
	g7 := gff3.Features.Key
	if e7 != g7 {
		t.Fatalf("Features.Key should be %v but is %v", e7, g7)
	}

	e8 := `testdata/test1.gff3.gz`
	g8 := gff3.Features.Value
	if e8 != g8 {
		t.Fatalf("Features.Value should be %v but is %v", e8, g8)
	}

}

func TestGff3KeepBySeqId(t *testing.T) {
	f1 := `testdata/test1.gff3.gz`
	gff3, err := NewFromFile(f1)
	if err != nil {
		t.Fatalf("error reading %s: %v", f1, err)
	}

	e1 := 1150
	g1 := gff3.FeatureCount()
	if e1 != g1 {
		t.Fatalf("%s should have %v Feature but has %v", f1, e1, g1)
	}

	pattern := `^[23]{1}$`
	gff3.KeepBySeqId(pattern)
	if err != nil {
		t.Fatalf("error in KeepBySeqID for %s with pattern %s: %v",
			f1, pattern, err)
	}

	e2 := 685
	g2 := gff3.FeatureCount()
	if e2 != g2 {
		t.Fatalf("%s should have %v Feature but has %v", f1, e2, g2)
	}
}

func TestGff3DeleteBySeqId(t *testing.T) {
	f1 := `testdata/test1.gff3.gz`
	gff3, err := NewFromFile(f1)
	if err != nil {
		t.Fatalf("error reading %s: %v", f1, err)
	}

	e1 := 1150
	g1 := gff3.FeatureCount()
	if e1 != g1 {
		t.Fatalf("%s should have %v Feature but has %v", f1, e1, g1)
	}

	pattern := `^[23]{1}$`
	gff3.KeepBySeqId(pattern)
	if err != nil {
		t.Fatalf("error in DeleteBySeqID for %s with pattern %s: %v",
			f1, pattern, err)
	}

	e2 := 685
	g2 := gff3.FeatureCount()
	if e2 != g2 {
		t.Fatalf("%s should have %v Feature but has %v", f1, e2, g2)
	}
}

func TestGff3FeaturesBySeqId(t *testing.T) {
	f1 := `testdata/test1.gff3.gz`
	gff3, err := NewFromFile(f1)
	if err != nil {
		t.Fatalf("error reading %s: %v", f1, err)
	}

	feats := gff3.FeaturesBySeqId()

	e1 := 3
	g1 := len(feats)
	if e1 != g1 {
		t.Fatalf("sequence count in %s should be %v but is %v", f1, e1, g1)
	}

	// Test whether the 0'th Feature of seq "1" looks like what we expect
	seq := `1`
	fidx := 0
	feat := feats[seq].Features[fidx]

	e2 := 16
	g2 := feat.LineNumber
	if e2 != g2 {
		t.Fatalf("LineNumber of seq `%s` feature `%d` in %s should be %v but is %v", seq, fidx, f1, e2, g2)
	}
	e3 := `1`
	g3 := feat.SeqId
	if e3 != g3 {
		t.Fatalf("SeqId of seq `%s` feature `%d` in %s should be %v but is %v", seq, fidx, f1, e3, g3)
	}
	e4 := `GRCh37`
	g4 := feat.Source
	if e4 != g4 {
		t.Fatalf("Source of seq `%s` feature `%d` in %s should be %v but is %v", seq, fidx, f1, e4, g4)
	}
	e5 := `chromosome`
	g5 := feat.Type
	if e5 != g5 {
		t.Fatalf("Type of seq `%s` feature `%d` in %s should be %v but is %v", seq, fidx, f1, e5, g5)
	}
	e6 := 1
	g6 := feat.Start
	if e6 != g6 {
		t.Fatalf("Start of seq `%s` feature `%d` in %s should be %v but is %v", seq, fidx, f1, e6, g6)
	}
	e7 := 249250621
	g7 := feat.End
	if e7 != g7 {
		t.Fatalf("End of seq `%s` feature `%d` in %s should be %v but is %v", seq, fidx, f1, e7, g7)
	}
	e8 := `.`
	g8 := feat.Score
	if e8 != g8 {
		t.Fatalf("Score of seq `%s` feature `%d` in %s should be %v but is %v", seq, fidx, f1, e8, g8)
	}
	e9 := `.`
	g9 := feat.Strand
	if e9 != g9 {
		t.Fatalf("Strand of seq `%s` feature `%d` in %s should be %v but is %v", seq, fidx, f1, e9, g9)
	}
	e10 := `.`
	g10 := feat.Strand
	if e10 != g10 {
		t.Fatalf("Phase of seq `%s` feature `%d` in %s should be %v but is %v", seq, fidx, f1, e10, g10)
	}

	// Blunt test for whole-of-Feature
	e11 := `1~GRCh37~chromosome~1~249250621~.~.~.~Alias=CM000663.1,NC_000001.10;ID=chromosome:1`
	g11 := feat.debugString()
	if e11 != g11 {
		t.Fatalf("seq `%s` feature `%d` in %s should be %v but is %v", seq, fidx, f1, e11, g11)
	}

	// Test whether the 1'th Feature of seq "1" looks like what we expect
	seq = `1`
	fidx = 1
	feat = feats[seq].Features[fidx]

	// Blunt test for whole-of-Feature
	e20 := `1~cpg~biological_region~10469~11240~1.3e+03~.~.~external_name=oe %3D 0.79;logic_name=cpg`
	g20 := feat.debugString()
	if e20 != g20 {
		t.Fatalf("seq `%s` feature `%d` in %s should be %v but is %v", seq, fidx, f1, e20, g20)
	}

	e21 := 18
	g21 := feat.LineNumber
	if e21 != g21 {
		t.Fatalf("LineNumber of seq `%s` feature `%d` in %s should be %v but is %v", seq, fidx, f1, e20, g20)
	}

	// Test whether the 311'th Feature of seq "3" looks like what we expect
	seq = `3`
	fidx = 311
	feat = feats[seq].Features[fidx]

	// Blunt test for whole-of-Feature
	e31 := `3~ensembl~exon~324372~324475~.~-~.~Name=ENSE00002088485;Parent=transcript:ENST00000516208;constitutive=1;ensembl_end_phase=-1;ensembl_phase=-1;exon_id=ENSE00002088485;rank=1;version=1`
	g31 := feat.debugString()
	if e31 != g31 {
		t.Fatalf("seq `%s` feature `%d` in %s should be %v but is %v", seq, fidx, f1, e31, g31)
	}
}

func TestGff3Feature(t *testing.T) {
	f1 := `testdata/test1.gff3.gz`
	gff3, err := NewFromFile(f1)
	if err != nil {
		t.Fatalf("error reading %s: %v", f1, err)
	}

	// Test whether the 0'th Feature looks like what we expect
	fidx := 0
	feat := gff3.Features.Features[fidx]

	e2 := 16
	g2 := feat.LineNumber
	if e2 != g2 {
		t.Fatalf("LineNumber of feature `%d` in %s should be %v but is %v", fidx, f1, e2, g2)
	}
	e3 := `1`
	g3 := feat.SeqId
	if e3 != g3 {
		t.Fatalf("SeqId of Feature `%d` in %s should be %v but is %v", fidx, f1, e3, g3)
	}
	e4 := `GRCh37`
	g4 := feat.Source
	if e4 != g4 {
		t.Fatalf("Source of feature `%d` in %s should be %v but is %v", fidx, f1, e4, g4)
	}
	e5 := `chromosome`
	g5 := feat.Type
	if e5 != g5 {
		t.Fatalf("Type of Feature `%d` in %s should be %v but is %v", fidx, f1, e5, g5)
	}
	e6 := 1
	g6 := feat.Start
	if e6 != g6 {
		t.Fatalf("Start of Feature `%d` in %s should be %v but is %v", fidx, f1, e6, g6)
	}
	e7 := 249250621
	g7 := feat.End
	if e7 != g7 {
		t.Fatalf("End of Feature `%d` in %s should be %v but is %v", fidx, f1, e7, g7)
	}
	e8 := `.`
	g8 := feat.Score
	if e8 != g8 {
		t.Fatalf("Score of Feature `%d` in %s should be %v but is %v", fidx, f1, e8, g8)
	}
	e9 := `.`
	g9 := feat.Strand
	if e9 != g9 {
		t.Fatalf("Strand of Feature `%d` in %s should be %v but is %v", fidx, f1, e9, g9)
	}
	e10 := `.`
	g10 := feat.Phase
	if e10 != g10 {
		t.Fatalf("Phase of Feature `%d` in %s should be %v but is %v", fidx, f1, e10, g10)
	}

	// Blunt test for whole-of-Feature
	e11 := `1~GRCh37~chromosome~1~249250621~.~.~.~Alias=CM000663.1,NC_000001.10;ID=chromosome:1`
	g11 := feat.debugString()
	if e11 != g11 {
		t.Fatalf("Feature `%d` in %s should be %v but is %v", fidx, f1, e11, g11)
	}

	// Test whether the 1'th Feature of looks like what we expect
	fidx = 1
	feat = gff3.Features.Features[fidx]

	// Blunt test for whole-of-Feature
	e20 := `1~cpg~biological_region~10469~11240~1.3e+03~.~.~external_name=oe %3D 0.79;logic_name=cpg`
	g20 := feat.debugString()
	if e20 != g20 {
		t.Fatalf("Feature `%d` in %s should be %v but is %v", fidx, f1, e20, g20)
	}

	e21 := 18
	g21 := feat.LineNumber
	if e21 != g21 {
		t.Fatalf("LineNumber of Feature `%d` in %s should be %v but is %v", fidx, f1, e21, g21)
	}

	// Test whether the last (1149'th) Feature like what we expect
	fidx = 1149
	feat = gff3.Features.Features[fidx]

	// Blunt test for whole-of-Feature
	e31 := `3~ensembl~exon~324372~324475~.~-~.~Name=ENSE00002088485;Parent=transcript:ENST00000516208;constitutive=1;ensembl_end_phase=-1;ensembl_phase=-1;exon_id=ENSE00002088485;rank=1;version=1`
	g31 := feat.debugString()
	if e31 != g31 {
		t.Fatalf("Feature `%d` in %s should be %v but is %v", fidx, f1, e31, g31)
	}

	e32 := 1184
	g32 := feat.LineNumber
	if e32 != g32 {
		t.Fatalf("LineNumber of Feature `%d` in %s should be %v but is %v", fidx, f1, e32, g32)
	}
}

func TestNewFeatureFromLine(t *testing.T) {
	f1 := "1\tGRCh37\tchromosome\t1\t249250621\t.\t.\t.\tAlias=CM000663.1,NC_000001.10;ID=chromosome:1"
	feat, err := NewFeatureFromLine(f1)
	if err != nil {
		t.Fatalf("Error calling NewFeatureFromLine on %s: %v", f1, err)
	}

	e3 := `1`
	g3 := feat.SeqId
	if e3 != g3 {
		t.Fatalf("SeqId of Feature should be %v but is %v", e3, g3)
	}
	e4 := `GRCh37`
	g4 := feat.Source
	if e4 != g4 {
		t.Fatalf("Source of Feature should be %v but is %v", e4, g4)
	}
	e5 := `chromosome`
	g5 := feat.Type
	if e5 != g5 {
		t.Fatalf("Type of Feature should be %v but is %v", e5, g5)
	}
	e6 := 1
	g6 := feat.Start
	if e6 != g6 {
		t.Fatalf("Start of Feature should be %v but is %v", e6, g6)
	}
	e7 := 249250621
	g7 := feat.End
	if e7 != g7 {
		t.Fatalf("End of Feature should be %v but is %v", e7, g7)
	}
	e8 := `.`
	g8 := feat.Score
	if e8 != g8 {
		t.Fatalf("Score of Feature should be %v but is %v", e8, g8)
	}
	e9 := `.`
	g9 := feat.Strand
	if e9 != g9 {
		t.Fatalf("Strand of Feature should be %v but is %v", e9, g9)
	}
	e10 := `.`
	g10 := feat.Phase
	if e10 != g10 {
		t.Fatalf("Phase of Feature should be %v but is %v", e10, g10)
	}
}

func TestFeatureClone(t *testing.T) {
	f1 := "1\tGRCh37\tchromosome\t1\t249250621\t.\t.\t.\tAlias=CM000663.1,NC_000001.10;ID=chromosome:1"
	feat, err := NewFeatureFromLine(f1)
	if err != nil {
		t.Fatalf("Error calling NewFeatureFromLine on %s: %v", f1, err)
	}

	newFeat := feat.Clone()

	problems := HelperCompareFeatures(feat, newFeat)
	if problems != "" {
		t.Fatal(problems)
	}

	if feat.SeqId != newFeat.SeqId {
		t.Fatalf("SeqId of Clone'd Feature's do not match: %v vs %v",
			feat.SeqId, newFeat.SeqId)
	}
	if feat.Source != newFeat.Source {
		t.Fatalf("Source of Clone'd Feature's do not match: %v vs %v",
			feat.Source, newFeat.Source)
	}
	if feat.Type != newFeat.Type {
		t.Fatalf("Type of Clone'd Feature's do not match: %v vs %v",
			feat.Type, newFeat.Type)
	}
	if feat.Start != newFeat.Start {
		t.Fatalf("Start of Clone'd Feature's do not match: %v vs %v",
			feat.Start, newFeat.Start)
	}
	if feat.End != newFeat.End {
		t.Fatalf("End of Clone'd Feature's do not match: %v vs %v",
			feat.End, newFeat.End)
	}
	if feat.Score != newFeat.Score {
		t.Fatalf("Score of Clone'd Feature's do not match: %v vs %v",
			feat.Score, newFeat.Score)
	}
	if feat.Strand != newFeat.Strand {
		t.Fatalf("Strand of Clone'd Feature's do not match: %v vs %v",
			feat.Strand, newFeat.Strand)
	}
	if feat.Phase != newFeat.Phase {
		t.Fatalf("Phase of Clone'd Feature's do not match: %v vs %v",
			feat.Phase, newFeat.Phase)
	}
}

func HelperCompareFeatures(f1, f2 *Feature) string {
	problem := ""

	if f1.SeqId != f2.SeqId {
		problem += fmt.Sprintf("SeqId of Feature's do not match: %v vs %v\n",
			f1.SeqId, f2.SeqId)
	}
	if f1.Source != f2.Source {
		problem += fmt.Sprintf("Source of Feature's do not match: %v vs %v\n",
			f1.Source, f2.Source)
	}
	if f1.Type != f2.Type {
		problem += fmt.Sprintf("Type of Feature's do not match: %v vs %v\n",
			f1.Type, f2.Type)
	}
	if f1.Start != f2.Start {
		problem += fmt.Sprintf("Start of Feature's do not match: %v vs %v\n",
			f1.Start, f2.Start)
	}
	if f1.End != f2.End {
		problem += fmt.Sprintf("End of Feature's do not match: %v vs %v\n",
			f1.End, f2.End)
	}
	if f1.Score != f2.Score {
		problem += fmt.Sprintf("Score of Feature's do not match: %v vs %v\n",
			f1.Score, f2.Score)
	}
	if f1.Strand != f2.Strand {
		problem += fmt.Sprintf("Strand of Feature's do not match: %v vs %v\n",
			f1.Strand, f2.Strand)
	}
	if f1.Phase != f2.Phase {
		problem += fmt.Sprintf("Phase of Feature's do not match: %v vs %v\n",
			f1.Phase, f2.Phase)
	}

	// Check Attribute count matches
	if len(f1.Attributes) != len(f2.Attributes) {
		problem += fmt.Sprintf("Attribute count of Feature's do not match: %v vs %v\n",
			len(f1.Attributes), len(f2.Attributes))
	}

	// Check Attributes
	for k, v := range f1.Attributes {
		if _, ok := f2.Attributes[k]; !ok {
			problem += fmt.Sprintf("Attribute key %v from f1 is not in f2\n", k)
		}
		if v != f2.Attributes[k] {
			problem += fmt.Sprintf("Attribute value of Feature's do not match: %v vs %v\n",
				v, f2.Attributes[k])
		}
	}

	return problem
}
