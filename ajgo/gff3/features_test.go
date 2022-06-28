package gff3

import (
	"bufio"
	"strings"
	"testing"
)

// 2 SeqIds, 5 identical Feature each
var fs1 = `##gff-version	3
1	ensembl	exon	1	10	.	.	.	ID=1
1	ensembl	exon	5	20	.	.	.	ID=2
1	ensembl	exon	21	23	.	.	.	ID=3
1	ensembl	exon	25	27	.	.	.	ID=4
1	ensembl	exon	30	40	.	.	.	ID=5
2	ensembl	exon	1	10	.	.	.	ID=6
2	ensembl	exon	5	20	.	.	.	ID=7
2	ensembl	exon	21	23	.	.	.	ID=8
2	ensembl	exon	25	27	.	.	.	ID=9
2	ensembl	exon	30	40	.	.	.	ID=10
`

// fs2 has the same content as fs1 but is read in a different order
var fs2 = `##gff-version	3
2	ensembl	exon	30	40	.	.	.	ID=10
1	ensembl	exon	5	20	.	.	.	ID=2
2	ensembl	exon	5	20	.	.	.	ID=7
1	ensembl	exon	21	23	.	.	.	ID=3
1	ensembl	exon	1	10	.	.	.	ID=1
1	ensembl	exon	25	27	.	.	.	ID=4
2	ensembl	exon	25	27	.	.	.	ID=9
2	ensembl	exon	1	10	.	.	.	ID=6
2	ensembl	exon	21	23	.	.	.	ID=8
1	ensembl	exon	30	40	.	.	.	ID=5
`

func TestFeaturesPrivateSort(t *testing.T) {
	s := strings.NewReader(fs1)
	b := bufio.NewScanner(s)
	g, err := NewFromScanner(b)
	if err != nil {
		t.Fatalf("NewGff3FromScanner should not have failed: %v", err)
	}

	e1 := 10
	g1 := g.FeatureCount()
	if e1 != g1 {
		t.Fatalf("Featurecount be %d but is %d", e1, g1)
	}

}

func TestFeaturesSort(t *testing.T) {
	s := strings.NewReader(fs1)
	b := bufio.NewScanner(s)
	g, err := NewFromScanner(b)
	if err != nil {
		t.Fatalf("NewGff3FromScanner should not have failed: %v", err)
	}

	e1 := 10
	g1 := g.FeatureCount()
	if e1 != g1 {
		t.Fatalf("Featurecount be %d but is %d", e1, g1)
	}

	// Note that this sort is not SeqId-aware
	g.Features.Sort()

	// TO DO - add some actual tests!
}
