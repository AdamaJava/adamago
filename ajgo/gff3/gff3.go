package gff3

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"ajgo/selector"
)

var (
	ErrNoGff3Headers = errors.New("File has no header lines - not compliant with GFF3 spec")
	ErrNoGff3Records = errors.New("File has no records - is this really what you wanted?")
)

// A GFF3 file is serialised in text as a list of features.  GFF3
// is designed to capture relationships between features by explicitly
// defining parent-child relationships via the Attributes column.
//
// The relationships make GFF3 somewhat more difficult to use than GTF
// because a GFF3 record only carries information specific to itself and
// not information from higher-level nodes in the relationship tree.
// For example, in Ensembl gene model GFF3 files, gene records contain
// the gene Name but exon and transcript records do not so if you want
// to filter exons by gene-name, you need to model the relationships
// between GFF3 records in your software.
//
// It's also worth noting that relationships as defined in the file are
// are unidirectional child->parent. This is very inconvenient because
// you must always build the tree from the leaves toward the root which
// is contrary to the way we think.  For example, transcripts contain
// the ID of the gene they relate to BUT gene records have no
// information on which transcripts they contain. So you can't take
// a list of genes and find the transcripts by following links, instead
// you need a full traversal of the transcript features looking at each
// one to see if it relates to any of the genes of interest.
//
// In order to work with the relationships between features, we have the
// Tree type and the Gff3 function NewTree.
type Gff3 struct {
	Name   string
	File   string
	Header []string
	//Features []*Feature
	Features *Features
}

func NewGff3() *Gff3 {
	return &Gff3{Features: NewFeatures()}
}

func (g *Gff3) FeatureCount() int {
	return g.Features.Count()
}

// Clone creates a deep copy of a Gff3, i.e. the new Gff3 shares no
// pointers with the original Gff3. After calling Clone you can change
// the original Gff3 or the copy without any concern that the change
// will affect the other Gff3. Obviously your memory use doubles.
func (g *Gff3) Clone() *Gff3 {
	ng := NewGff3()
	ng.Name = g.Name
	ng.File = g.File
	ng.Header = append(ng.Header, g.Header...)
	nfs := g.Features.Clone()
	ng.Features = nfs
	return ng
}

// DeleteBySeqId removes any Feature that have SeqId that match the
// regexp pattern. It returns the SeqId that were deleted.
func (g *Gff3) DeleteBySeqId(pattern string) ([]string, error) {
	var deleted []string

	deleted, err := g.Features.DeleteBySeqId(pattern)
	if err != nil {
		return deleted, fmt.Errorf("Gff3.DeleteBySeqId: error deleting Features: %w", err)
	}

	return deleted, nil
}

// KeepBySeqId keeps any Feature that have SeqId that match the
// regexp pattern. It returns the SeqId that were kept.
func (g *Gff3) KeepBySeqId(pattern string) ([]string, error) {
	var kept []string

	kept, err := g.Features.KeepBySeqId(pattern)
	if err != nil {
		return kept, fmt.Errorf("Gff3.KeepBySeqId: error keeping Features: %w", err)
	}

	return kept, nil
}

// NewFromFile reads from a file and returns a pointer to a Gff3.
func NewFromFile(file string) (*Gff3, error) {
	// Open file
	ff, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer ff.Close()

	// We need to define this before we handle gzip
	var scanner *bufio.Scanner

	// Based on file extension, handle gzip files
	found, err := regexp.MatchString(`\.[gG][zZ]$`, file)
	if err != nil {
		return nil, fmt.Errorf("NewFromFile: error matching gzip file pattern against %s: %w", file, err)
	}
	if found {
		// For gzip files, put a gzip.Reader into the chain
		reader, err := gzip.NewReader(ff)
		if err != nil {
			return nil, fmt.Errorf("NewFromFile: error opening gzip file %s: %w", file, err)
		}
		defer reader.Close()
		scanner = bufio.NewScanner(reader)
	} else {
		// For non gzip files, go straight to bufio.Reader
		scanner = bufio.NewScanner(ff)
	}

	gff3, err := NewFromScanner(scanner)
	if err != nil {
		return gff3, fmt.Errorf("NewFromFile: error scanning: %w", err)
	}
	gff3.File = file
	gff3.Features.Key = `file`
	gff3.Features.Value = file
	return gff3, nil
}

// NewFromScanner reads from a *bufio.Scanner and returns a pointer
// to a Gff3. It is an alternative to NewFromFile and is useful when
// you have Gff3 records as a block of text in memory.
func NewFromScanner(scanner *bufio.Scanner) (*Gff3, error) {
	gff3 := NewGff3()
	gff3.Features.Key = `source`
	gff3.Features.Value = `gff3.NewFromScanner()`

	// Unnecessary but explicit
	scanner.Split(bufio.ScanLines)

	// Pattern for track lines
	rex := regexp.MustCompile(`^#`)

	// Read the file
	lctr := 0
	fs := NewFeatures()
	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\n")
		lctr++
		if rex.MatchString(line) {
			// Ensembl seems to use ### as a visual divider line in
			// GFF3 files so we are going to drop these lines.
			if line != `###` {
				gff3.Header = append(gff3.Header, line)
			}
		} else {
			f, err := NewFeatureFromLine(line)
			if err != nil {
				return nil, fmt.Errorf("NewFromScanner: error creating Feature: %w", err)
			}
			f.LineNumber = lctr
			//if _, ok := gff3.Seqs[f.SeqId]; !ok {
			//	gff3.Seqs[f.SeqId] = &FeatureCollection{Id: f.SeqId}
			//}
			//gff3.Seqs[f.SeqId].Features = append(gff3.Seqs[f.SeqId].Features, f)
			fs.Features = append(fs.Features, f)
		}
	}

	// If there are no Headers then it can't be a gff3 because the
	// gff-version on the first line is mandatory.
	if len(gff3.Header) == 0 {
		return nil, ErrNoGff3Headers
	}

	// If there are no Features then something is wrong
	if fs.Count() == 0 {
		return nil, ErrNoGff3Records
	}

	// If the very first line is not the gff3 identifier then we exit
	// immediately. Unfortunately, some folks (Ensembl at least) seem to
	// put an arbitrary number of spaces in as separator so we will need
	// to use a pattern rather than a simple string equality test.
	ok, err := regexp.Match(`^##gff-version\s+3`, []byte(gff3.Header[0]))
	if err != nil {
		return nil, fmt.Errorf("NewFromScanner: error pattern matching gff-version line: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("NewFromScanner: file is not a gff3, first line is: %s", gff3.Header[0])
	}

	gff3.Features = fs
	return gff3, nil
}

func (g *Gff3) Write(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	// Write Headers (remember they still have their ##/#! prefixes)
	for _, h := range g.Header {
		_, err = w.WriteString(h + "\n")
		if err != nil {
			return err
		}
	}

	// TO DO - Features should probably have a Write() of its own - this
	//         is a pretty dirty way tot do this.

	// Write Features
	for _, f := range g.Features.Features {
		_, err = w.WriteString(f.String() + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}

// SeqIds returns a sorted list of SeqId strings. This is
// useful anywhere that you want consistent ordering. Note that since
// SeqId is always a string the ordering is by string so chromosome
// names may not sort as you'd expect/hope.
func (g *Gff3) SeqIds() []string {
	return g.Features.SeqIds()
}

// FeatureAttributes will look at all Features across all Seqs and
// tally which attributes are present and how often.
func (g *Gff3) FeatureAttributes() map[string]int {
	return g.Features.Attributes()
}

func (g *Gff3) ApplySelector(sel *selector.Selector) error {
	return g.Features.ApplySelector(sel)
}

// FeaturesBySeqId creates a map of Features structs where each Features
// contain Feature with the same SeqId. This can simplify a lot of other
// operations such as Merge and Consolidate because it removes the
// possibility that overlapping ranges are from different sequences.
func (g *Gff3) FeaturesBySeqId() map[string]*Features {
	return g.Features.BySeqId()
}

// VersionedHeaders returns the header lines from a GFF3 but with
// an identifier string inserted into to each header line. This allows us
// to merge headers from multiple GFF3 files while retaining information
// about which file the headers originally came from. Obviously enough
// this also means we need to record which file is referred to by
// each identifier. In the following example, we used "2" as the
// identifier which transformed these headers:
//
//  ##gff-version   3
//  ##sequence-region   1 1 249250621
//  #!genome-build  GRCh37.p13
//  #!genome-version GRCh37
//  # An ensemble gene model
//
// into these:
//
//  ##gff-version-2   3
//  ##sequence-region-2   1 1 249250621
//  #!genome-build-2  GRCh37.p13
//  #!genome-version-2 GRCh37
//  #-2 An ensemble gene model
func (g *Gff3) VersionedHeaders(suffix string) []string {
	var versioned []string

	// We will be splitting on the first white space but we must capture
	// the whitespace and newline so we can remake the line accurately.
	re := regexp.MustCompile("(?s)(^#[^\t ]*)(.*)")

	for _, h := range g.Header {
		submatches := re.FindStringSubmatch(h)
		if len(submatches) == 3 {
			versioned = append(versioned, submatches[1]+`-`+suffix+submatches[2])
		} else {
			// If the matching failed for some reason
			versioned = append(versioned, h)
		}
	}

	return versioned
}
