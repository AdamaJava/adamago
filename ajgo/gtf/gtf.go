package gtf

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/grendeloz/interval"
)

type Gtf struct {
	Name     string
	SeqFeats map[string]*SeqFeat
	Files    []string
}

// SeqFeat is a collection of Features with the same SeqName.
type SeqFeat struct {
	SeqName  string
	Features []*Feature
	IsSorted bool
}

// This abstraction is pure future-proofing - it will be useful if we
// ever want to change this type.
type FeatCoord int64

// GTF files have a few idiosyncrasies.
// 1. All of the required fields must be present and missing values
//    are represented by "." char.  This means that you can't convert
//    fields into numeric types without making some sort of decision
//    about missing - is it OK to turn missing values into zero or not?
type Feature struct {
	SeqName    string
	Source     string
	Feature    string
	Start      FeatCoord
	End        FeatCoord
	Score      string // should be float but missing is "."
	Strand     string
	Frame      string // should be int but missing is "."
	Attributes map[string]string
	Original   string
	LineNumber int
}

type sortFeat struct {
	Ff []*Feature
}

// Satisfy interval.Interval interface
func (f *Feature) Low() int {
	return int(f.Start)
}
func (f *Feature) High() int {
	return int(f.End)
}

func New(name string) *Gtf {
	feats := make(map[string]*SeqFeat)
	return &Gtf{
		Name:     name,
		SeqFeats: feats}
}

func (g *Gtf) KeepGenes(genes []string) error {
	// TO DO - we need to implement this so we can produce masks for
	// particular gene panels.
	return nil
}

// AddGtfFile parses a GTF file containing a gene model and adds it to
// the Gtf. In general, gene models come in single files but this
// way of handling initialising a gene model provides future proofing.
//
// Once a GTF has been parsed into memory, the Features are separated by
// SeqName. If a SeqName already exists in Gtf.SeqFeats the
// features are appended and if the SeqName does not already exist, it is
// added. Any SeqFeat that has Features appended also has IsSorted set
// to false.
func (g *Gtf) AddGtfFile(file string) error {
	feats, err := ParseGtfModelFile(file)
	if err != nil {
		return err
	}

	// We will sort the Features into a local struct and then do a
	// single set of Add ops on the Gtf.

	lfs := make(map[string]*SeqFeat)
	// Create SeqFeats as required and distribute Features
	for _, f := range feats {
		if _, ok := lfs[f.SeqName]; !ok {
			// Add SeqFeat with new name
			lfs[f.SeqName] = &SeqFeat{SeqName: f.SeqName}
		}
		lfs[f.SeqName].Features = append(lfs[f.SeqName].Features, f)
	}

	// Add lfs to the Gtf SeqFeats
	for _, fs := range lfs {
		if _, ok := g.SeqFeats[fs.SeqName]; !ok {
			// Add SeqFeat if it does not exist
			g.SeqFeats[fs.SeqName] = &SeqFeat{SeqName: fs.SeqName}
		}
		g.SeqFeats[fs.SeqName].Add(fs.Features)
	}

	g.Files = append(g.Files, file)
	return nil
}

func (g *Gtf) FeatureCount() int {
	var ctr int
	for _, fs := range g.SeqFeats {
		ctr += len(fs.Features)
	}
	return ctr
}

// DeletSeqFeats removes any SeqFeats in a Gtf that match a given
// regexp pattern. It returns the names of SeqFeats that were deleted.
func (g *Gtf) DeleteSeqFeats(pattern string) ([]string, error) {
	var deleted []string

	deleter, err := regexp.Compile(pattern)
	if err != nil {
		return deleted, fmt.Errorf("DeleteFeatSeqs: error compiling pattern %s: %w", pattern, err)
	}

	// Delete any SeqFeats that match by name
	for _, fs := range g.SeqFeats {
		if deleter.MatchString(fs.SeqName) {
			delete(g.SeqFeats, fs.SeqName)
			deleted = append(deleted, fs.SeqName)
		}
	}

	return deleted, nil
}

// Sort sorts each SeqFeat within the gene model. The bool tells you
// whether or not a sort was done - if all of the SeqFeats IsSorted
// properties were true, no sort will be done and false will be
// returned. But if any SeqFeat needed to be sorted then true
// is returned.
func (g *Gtf) Sort() (bool, error) {
	var sortWasDone bool
	for _, sf := range g.SeqFeats {
		sorted, err := sf.Sort()
		if err != nil {
			return sortWasDone, fmt.Errorf("Sort: error sorting SeqFeat: %s", sf.SeqName)
		}
		if sorted {
			sortWasDone = true
		}
	}
	return sortWasDone, nil
}

// Add adds Features and sets IsSorted to false. Checking sortedness can be
// expensive so we don't do it by default. The SeqFeat *may* still be
// sorted but without a check, we must assume it is not.
//
// If you have lots of Features to add, it will be quickest to use this
// version of Add and do a single CheckSort() on the SeqFeat once all of
// the adding is complete.
func (fs *SeqFeat) Add(feats []*Feature) error {
	fs.Features = append(fs.Features, feats...)
	fs.IsSorted = false
	return nil
}

// AddWithCheckSort adds Features and checks sortedness. Checking
// sortedness can be expensive so this function is best used when you
// have all of the Features you want to add and you can pass them in as
// single slice. If you are going to add Features one at a time or in
// small quantities, it will be more efficient to use Add() instead and
// call CheckSorted() once adding is complete.
func (fs *SeqFeat) AddWithCheckSort(feats []*Feature) error {
	fs.Features = append(fs.Features, feats...)
	fs.CheckSorted()
	return nil
}

// CheckSorted checks and if necessary updates the IsSorted property.
func (fs *SeqFeat) CheckSorted() {
	// Check sortedness
	var IsSorted bool = true
	for i := 0; i < len(fs.Features)-1; i++ {
		if fs.Features[i].Start > fs.Features[i+1].Start {
			IsSorted = false
			// Once we know it's unsorted,we can skip checking
			break
		}
	}
	fs.IsSorted = IsSorted
}

// Sort sorts Features smallest to largest based on the Start position.
// The bool tells you whether or not a sort was done - if the SeqFeat
// IsSorted property is true, the sort will not be done. If you wish to
// force a sort, set IsSorted to false and than call Sort.
//
// Features with the same Start position will be sorted smallest to
// largest based on the End position. The ordering of Features with the
// same Start and End position is unspecified and although the sort
// appears to be stable, this is not specifically part of the design and
// is not guaranteed.
func (fs *SeqFeat) Sort() (bool, error) {
	// Check sortedness
	if fs.IsSorted {
		return false, nil
	}

	// We are going to use a map to do our sorting. Once all Features have
	// been placed into the map by start position, the observed starts
	// are sorted and the map is walked doing by-End sorting for any cases
	// where there are multiple Features with the same start position.

	// Walk Features putting them into the map by start position
	sorter := make(map[int]*sortFeat)
	for i := 0; i < len(fs.Features); i++ {
		// We need to convert FeatCoord (.Start) to an int
		start := int(fs.Features[i].Start)
		//log.Infof("  idx:%d start:%d feat:%+v", i, start, fs.Features[i])
		if _, ok := sorter[start]; !ok {
			sorter[start] = &sortFeat{}
		}
		sorter[start].Ff = append(sorter[start].Ff, fs.Features[i])
	}

	// Now we walk the map by start position, do any required by-End
	// sorting and write the Features to sorted in their final order.
	var sorted []*Feature

	// Sort the starts
	starts := []int{}
	for k := range sorter {
		starts = append(starts, k)
	}
	sort.Ints(starts)

	// Walk the map by start
	for _, start := range starts {
		if len(sorter[start].Ff) == 1 {
			// If there's only one Feature, append it
			sorted = append(sorted, sorter[start].Ff[0])
		} else {
			// If there's more than one Feature, we sort by Feature.End
			endSorter := make(map[int]*sortFeat)
			for _, f := range sorter[start].Ff {
				// We need to convert FeatCoord (.End) to an int
				end := int(f.End)
				if _, ok := endSorter[end]; !ok {
					endSorter[end] = &sortFeat{}
				}
				endSorter[end].Ff = append(endSorter[end].Ff, f)
			}

			// Now order the ends
			ends := []int{}
			for k := range endSorter {
				ends = append(ends, k)
			}
			sort.Ints(ends)

			// Finally write out the Features by ends
			for _, end := range ends {
				for _, f := range endSorter[end].Ff {
					sorted = append(sorted, f)
				}
			}
		}
	}

	// That's a lot of trouble just so we can do this
	fs.Features = sorted
	fs.IsSorted = true
	return true, nil
}

// Consolidate looks at all of the Features in a (sorted) SeqFeat and if
// they are immediately adjacent or overlap in any way, they are
// consolidated into a single new feature. If any Features are equal to
// each other or contained within another Feature, the duplicate or
// contained Feature is deleted. The returned boolean tells whether any
// work was done - if no records were merged (or the list of Features
// was empty) then false is returned.
//
// This process is destructive! Genes with many transcripts often share
// exons across multiple transcripts so removal of the duplicates means it
// is no longer possible to work on transcripts. It will also merge
// overlapping exons from different genes on opposite strands.
//
// Despite the caveats, there are use cases where this behaviour is
// exactly what is needed, for example if you are constructing a mask to
// work out which genomic positions are part within the exons of a gene
// or set of genes. For that use case, the Prune* family of functions
// may also be useful.
func (fs *SeqFeat) Consolidate() (int, error) {
	var count int
	if !fs.IsSorted {
		return count, fmt.Errorf("Consolidate: cannot call on an unsorted SeqFeat")
	}

	// Consolidating an empty list of Features is legal but obviously
	// there are no records to be consolidated.
	if len(fs.Features) == 0 {
		return 0, nil
	}

	// This is a bit tricky but we will always be comparing the last
	// Feature in the keepers list against the next Feature on the full
	// list. This will let the keeper Feature merge with as many records
	// as are required from the main list. Once we get a disjoint Compare,
	// that Feature from the main list is copied onto the keeper list and
	// away we go again merging onto the new "last" keeper Feature.

	var keepers []*Feature
	keepers = append(keepers, fs.Features[0])

	for i := 1; i < len(fs.Features); i++ {
		keepidx := len(keepers) - 1
		allen := interval.Compare(keepers[keepidx], fs.Features[i])

		//log.Infof("%s[%d,%d]%d vs %s[%d,%d]%d = %d",
		//	keepers[keepidx].SeqName, keepers[keepidx].Start,
		//    keepers[keepidx].End, keepidx,
		//	fs.Features[i].SeqName, fs.Features[i].Start,
		//    fs.Features[i].End, i,
		//	allen)

		// 1. Return error on AllenR of Unknown
		// 2. Return error if b starts before a because that means that
		//    the lists are not sorted.
		// 3. Append to the keepers list if PrecedesB
		// 2. Otherwise merge.
		if allen == interval.Unknown {
			return count, fmt.Errorf("Consolidate: Allen Relationship is Unknown for {%+v} vs {%+v}",
				keepers[keepidx], fs.Features[i])
		} else if allen == interval.FinishesB ||
			allen == interval.IsContainedByB ||
			allen == interval.IsOverlappedByB ||
			allen == interval.IsMetByB ||
			allen == interval.IsPrecededByB {
			return count, fmt.Errorf("Consolidate: {%+v} vs {%+v} means SeqFeat %s is unsorted",
				keepers[keepidx], fs.Features[i], fs.SeqName)
		} else if allen == interval.PrecedesB {
			keepers = append(keepers, fs.Features[i])
		} else {
			keepers[keepidx].Merge(fs.Features[i])
			count++
		}
	}

	// Attach the list of keepers to fs
	fs.Features = keepers
	return count, nil
}

// KeepByFeatures takes a list of strings and all Features in the SeqFeat
// will be deleted unless Feature.Feature *exactly* matches one of the
// supplied strings. An example use case would be to only keep Features
// that were exons and drop CDS, start_codon, stop_codon etc.
func (fs *SeqFeat) KeepByFeatures(keeps []string) int {
	keep := make(map[string]int)
	for _, s := range keeps {
		keep[s]++
	}

	var keepers []*Feature
	var lost int
	for _, f := range fs.Features {
		if _, ok := keep[f.Feature]; ok {
			keepers = append(keepers, f)
		} else {
			lost++
		}
	}

	fs.Features = keepers
	return lost
}

func NewFeatureFromFields(fields []string) (*Feature, error) {
	var feat Feature
	if len(fields) < 8 {
		return nil, fmt.Errorf("NewFeatureFromFields: only %d fields supplied - 8 or 9 are required", len(fields))
	}

	feat.Original = strings.Join(fields, "\t")
	feat.SeqName = fields[0]
	feat.Source = fields[1]
	feat.Feature = fields[2]
	if i, err := strconv.ParseInt(fields[3], 10, 64); err != nil {
		return nil, fmt.Errorf("NewFeatureFromFields: Feature.Start error converting %s to int64: %w", fields[3], err)
	} else {
		feat.Start = FeatCoord(i)
	}
	if i, err := strconv.ParseInt(fields[4], 10, 64); err != nil {
		return nil, fmt.Errorf("NewFeatureFromFields: Feature.End error converting %s to int64: %w", fields[4], err)
	} else {
		feat.End = FeatCoord(i)
	}
	feat.Score = fields[5]
	feat.Strand = fields[6]
	feat.Frame = fields[7]

	// This format is full of spaces and leading and trailing spaces
	// upset everything including splitting on space so get ready for
	// what appears to be an inordinate level of TrimSpace use.
	feat.Attributes = make(map[string]string)
	splitable := strings.TrimSpace(fields[8])
	if splitable != "" {
		attributes := strings.Split(splitable, ";")
		for _, a := range attributes {
			// There is often a trailing empty attribute and we don't want
			// empty stuff in the map so skip any empty attributes
			if a == "" {
				continue
			}
			// Much more space trimming required here
			a := strings.TrimSpace(a)
			subs := strings.SplitN(a, " ", 2)
			// The split doesn't remove white space
			if len(subs) == 2 {
				key := strings.TrimSpace(subs[0])
				val := strings.TrimSpace(subs[1])
				feat.Attributes[key] = val
			} else if len(subs) == 1 {
				key := strings.TrimSpace(subs[0])
				feat.Attributes[key] = ""
			} else {
				// This should be impossible because we already checked
				// that field[8] != ""
			}
		}
	}
	return &feat, nil
}

func ParseGtfModelFile(file string) ([]*Feature, error) {
	var feats []*Feature

	// Open file
	ff, err := os.Open(file)
	if err != nil {
		return feats, err
	}
	defer ff.Close()

	// We need to define this before we handle gzip
	var scanner *bufio.Scanner

	// Based on file extension, handle gzip files
	found, err := regexp.MatchString(`\.[gG][zZ]$`, file)
	if err != nil {
		return feats, fmt.Errorf("ParseGtfModelFile: error matching gzip file pattern against %s: %w", file, err)
	}
	if found {
		// For gzip files, put a gzip.Reader into the chain
		reader, err := gzip.NewReader(ff)
		if err != nil {
			return feats, fmt.Errorf("ParseGtfModelFile: error opening gzip file %s: %w", file, err)
		}
		defer reader.Close()
		scanner = bufio.NewScanner(reader)
	} else {
		// For non gzip files, go straight to bufio.Reader
		scanner = bufio.NewScanner(ff)
	}

	// Unnecessary but explicit
	scanner.Split(bufio.ScanLines)

	// Pattern for track lines
	rex := regexp.MustCompile(`^track`)

	// Read the file
	lctr := 0
	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\n")
		lctr++
		if rex.MatchString(line) {
			// We are skipping track lines, if any
		} else {
			fields := strings.Split(line, "\t")
			f, err := NewFeatureFromFields(fields)
			if err != nil {
				return feats, fmt.Errorf("ParseGtfModelFile: error creating new Feature from fields: %w", err)
			}
			f.LineNumber = lctr
			feats = append(feats, f)
		}
	}

	return feats, nil
}

/*

// This logic relies upon a feature being a half-open range, i.e. Low
// is within the range but High is the first position past the end of
// the range. For example, an Interval with Low:High of 0:1 would
// include 0 but not 1 and an interval 6:8 would include 6 and 7 but
// not 8.
func (a *Feature) Compare(b *Feature) AllenRelationship {
	if a.End < b.Start {
		return PrecedesB // AR=1
	} else if a.End == b.Start {
		return MeetsB // AR=2
	} else if a.Start < b.Start && a.End < b.End && a.End > b.Start {
		return OverlapsB // AR=3
	} else if a.Start == b.Start && a.End < b.End {
		return StartsB // AR=4
	} else if a.Start > b.Start && a.End == b.End {
		return FinishesB // AR=5
	} else if a.Start < b.Start && a.End > b.End {
		return ContainsB // AR=6
	} else if a.Start == b.Start && a.End == b.End {
		return EqualsB // AR=7
	} else if a.Start > b.Start && a.End < b.End {
		return IsContainedByB // AR=8
	} else if a.Start < b.Start && a.End == b.End {
		return IsFinishedByB // AR=9
	} else if a.Start == b.Start && a.End > b.End {
		return IsStartedByB // AR=10
	} else if a.Start > b.Start && a.End > b.End && a.Start < b.End {
		return IsOverlappedByB // AR=11
	} else if a.Start == b.End {
		return IsMetByB // AR=12
	} else if a.Start > b.End {
		return IsPrecededByB // AR=13
	} else {
		return Error // AR=0
	}
}
*/

// Merge will merge the supplied Feature (b) on top of the current
// Feature. The Allen Relationship between the Features is not checked
// so if that matters to you, you must enforce any relevant logic before
// calling Merge.
//
// The Name of both Features must be identical or an error is returned.
// The Start of the merged Feature is the lesser of the Starts of the
// two Features and the End is the greater of the Ends of the two
// Features. All other fields are set to missing if the Features have
// different values, otherwise the values are left alone, i.e. they
// will stay set to the value that the Features share.
//
// Attributes are deleted unless they are present and identical in
// both Features.
func (a *Feature) Merge(b *Feature) error {
	if a.SeqName != b.SeqName {
		return fmt.Errorf("Merge: cannot merge Features with different sequence names: %s %s",
			a.SeqName, b.SeqName)
	}

	// Set outer limits for the merged interval
	if b.Start < a.Start {
		a.Start = b.Start
	}
	if b.End > a.End {
		a.End = b.End
	}

	// If not the same, set to missing
	if a.Source != b.Source {
		a.Source = `.`
	}
	if a.Feature != b.Feature {
		a.Feature = `.`
	}
	if a.Score != b.Score {
		a.Score = `.`
	}
	if a.Strand != b.Strand {
		a.Strand = `.`
	}
	if a.Frame != b.Frame {
		a.Frame = `.`
	}

	var attrs []string
	for k, _ := range a.Attributes {
		attrs = append(attrs, k)
	}
	for _, attr := range attrs {
		if _, ok := b.Attributes[attr]; ok {
			if a.Attributes[attr] != b.Attributes[attr] {
				// a and b have different values for attr so delete
				delete(a.Attributes, attr)
			}
		} else {
			// attr is not in b so delete
			delete(a.Attributes, attr)
		}
	}

	return nil
}

func (f *Feature) String() string {
	// Attributes are ;-separated
	attrString := f.AttributesString()

	// Fields are tab-separated
	output := strings.Join([]string{
		f.SeqName,
		f.Source,
		f.Feature,
		strconv.Itoa(int(f.Start)),
		strconv.Itoa(int(f.End)),
		f.Score,
		f.Strand,
		f.Frame,
		attrString}, "\t")

	return output
}

// SelectedAttributesString is a variant of AttributesString where a
// list is supplied which determines which attributes are written and
// the order. If an attribute on the list is not present, it is skipped in
// the output. This means that different Features may have different
// Attributes written out but only if the Features had different Attributes
// in the first place.
func (f *Feature) SelectedAttributesString(attrs []string) string {
	// Write selected attributes in selected order
	var attrStrings []string
	for _, s := range attrs {
		if _, ok := f.Attributes[s]; ok {
			attrStrings = append(attrStrings, s+" "+f.Attributes[s])
		}
	}
	// Attributes are ;-separated
	attrString := strings.Join(attrStrings, "; ")

	return attrString
}

// AttributesString will create a ;-separated string of the key:value
// Attributes sorted by key name. This may not match the original order
// of attributes from the gene model file.
func (f *Feature) AttributesString() string {
	// Sort keys
	var keys []string
	for k, _ := range f.Attributes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Write all attributes in key-sorted order
	var attrStrings []string
	for _, k := range keys {
		attrStrings = append(attrStrings, k+" "+f.Attributes[k])
	}

	// Attributes are ;-separated
	attrString := strings.Join(attrStrings, "; ")

	return attrString
}

func (g *Gtf) Write(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	// Write Feature by SeqFeat
	names := g.SeqFeatNames()
	for _, name := range names {
		for _, f := range g.SeqFeats[name].Features {
			_, err = w.WriteString(f.String() + "\n")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// SeqFeatNames returns sorted list of the SeqFeat names in the
// Gtf. This is useful anywhere that you want consistent ordering.
func (g *Gtf) SeqFeatNames() []string {
	var names []string
	for _, fs := range g.SeqFeats {
		names = append(names, fs.SeqName)
	}
	sort.Strings(names)

	return names
}

// FeatureAttributes will look at all Features and tally
// which attributes are present and how often.
func (g *Gtf) FeatureAttributes() map[string]int {
	tally := make(map[string]int)

	for _, fs := range g.SeqFeats {
		for _, f := range fs.Features {
			for k, _ := range f.Attributes {
				tally[k]++
			}
		}
	}

	return tally
}
