package gff3

import (
	"bufio"
	"strings"
	"testing"
)

// 2 SeqIds, 5 identical Feature each
var feature_1 = `##gff-version	3
1	ensembl	exon	1	10	.	.	.	ID=1
1	ensembl	exon	5	20	.	.	.	ID=2
1	ensembl	exon	21	23	.	.	.	ID=3
1	ensembl	exon	25	27	.	.	.	ID=4
1	ensembl	exon	30	40	.	.	.	ID=5
2	lbmesne	noxe	1	10	.	.	.	ID=6
2	ensembl	exon	5	20	.	.	.	ID=7
2	ensembl	exon	21	23	.	.	.	ID=8
2	ensembl	exon	25	27	.	.	.	ID=9
2	ensembl	exon	30	40	.	.	.	ID=10
`

// fs2 has the same content as fs1 but is read in a different order
var feature_2 = `##gff-version	3
2	ensembl	exon	30	40	.	.	.	ID=10
1	ensembl	exon	5	20	.	.	.	ID=2
2	ensembl	exon	5	20	.	.	.	ID=7
1	ensembl	exon	21	23	.	.	.	ID=3
1	ensembl	exon	1	10	.	.	.	ID=1
1	ensembl	exon	25	27	.	.	.	ID=4
2	ensembl	exon	25	27	.	.	.	ID=9
2	lbmesne	noxe	1	10	.	.	.	ID=6
2	ensembl	exon	21	23	.	.	.	ID=8
1	ensembl	exon	30	40	.	.	.	ID=5
`

func TestFeaturePrivateNewOverlapFeature(t *testing.T) {
	s := strings.NewReader(feature_1)
	b := bufio.NewScanner(s)
	g, err := NewFromScanner(b)
	if err != nil {
		t.Fatalf("NewGff3FromScanner should not have failed: %v", err)
	}

	e1 := 10
	g1 := g.FeatureCount()
	if e1 != g1 {
		t.Fatalf("Featurecount be %v but is %v", e1, g1)
	}

	O1 := newOverlapFeature(g.Features.Features[4],
		g.Features.Features[5])

	e2 := `1`
	g2 := O1.SeqId
	if e2 != g2 {
		t.Fatalf("SeqId should be %v but is %v", e2, g2)
	}

	e3 := `ajgo-merge`
	g3 := O1.Source
	if e3 != g3 {
		t.Fatalf("Source should be %v but is %v", e3, g3)
	}

	e4 := `SO:0000110`
	g4 := O1.Type
	if e4 != g4 {
		t.Fatalf("Type should be %v but is %v", e4, g4)
	}

	e5 := 0
	g5 := O1.Start
	if e5 != g5 {
		t.Fatalf("Start should be %v but is %v", e5, g5)
	}

	e6 := 0
	g6 := O1.End
	if e6 != g6 {
		t.Fatalf("End should be %v but is %v", e6, g6)
	}

	e7 := `.`
	g7 := O1.Score
	if e7 != g7 {
		t.Fatalf("Score should be %v but is %v", e7, g7)
	}

	e8 := `.`
	g8 := O1.Strand
	if e8 != g8 {
		t.Fatalf("Strand should be %v but is %v", e8, g8)
	}

	e9 := `.`
	g9 := O1.Phase
	if e9 != g9 {
		t.Fatalf("Phase should be %v but is %v", e9, g9)
	}

	e10 := `ensembl,lbmesne`
	g10 := O1.Attributes[`Sources`]
	if e10 != g10 {
		t.Fatalf("Attributes[Sources] should be %v but is %v", e10, g10)
	}

	e11 := `exon,noxe`
	g11 := O1.Attributes[`Types`]
	if e11 != g11 {
		t.Fatalf("Attributes[Sources] should be %v but is %v", e11, g11)
	}
}
