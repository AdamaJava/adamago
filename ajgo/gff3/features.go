package gff3

import (
	"ajgo/selector"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/grendeloz/interval"
	// log "github.com/sirupsen/logrus"
)

// Features is a collection of Features based on some property
// (captured in Key) and some value of the property (captured in
// Value).
//
// The Gff3 type has functions for creating commonly used
// Features, e.g. FeaturesByGene, FeaturesByTranscript.
type Features struct {
	Key      string
	Value    string
	Features []*Feature
	IsSorted bool
}

// NewFeatures creates a pointer to a new instance of type Features.
func NewFeatures() *Features {
	return &Features{}
}

// CheckSorted checks and if necessary updates the IsSorted property.
// Note that it only checks Start and no other Feature fields so it will
// not distinguish between different SeqId, Type etc.
func (fs *Features) CheckSorted() {
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
// If the Seq.IsSorted property is true, the sort will not be done. If
// you wish to force a sort, set IsSorted to false and than call Sort.
// Sort is SeqId-aware so it will partition the Feature by SeqId and
// sort within each partition.
//
// Features with the same Start position will be sorted smallest to
// largest based on the End position. The ordering of Features with the
// same Start and End position is unspecified and although the sort may
// currently be stable, this is not by design and is not guaranteed.
func (fs *Features) Sort() {
	// Do not sort if already sorted
	if fs.IsSorted {
		return
	}

	// Sorting needs to be by SeqId
	seqs := fs.BySeqId()

	var seqids []string
	for seqid, sfs := range seqs {
		sfs.simpleSort()
		seqids = append(seqids, seqid)
	}
	sort.Strings(seqids)

	// Put humpty dumpty back together again
	var feats []*Feature
	for _, seqid := range seqids {
		feats = append(feats, seqs[seqid].Features...)
	}

	fs.Features = feats
	fs.IsSorted = true
}

// Id returns a simple string identifier based on Key & Value fields.
func (fs *Features) Id() string {
	return strings.Join([]string{fs.Key, fs.Value}, `:`)
}

// Consolidate looks at all of the Feature in a (sorted) instance of
// Features and if they are immediately adjacent or overlap in any way,
// they are consolidated into a single new Feature. If any Feature are
// equal to each other or one is contained within the other, the duplicate
// or contained Feature is deleted. Features must not contain records
// with different SeqId.
//
// This process is destructive! Genes with many transcripts often share
// exons across multiple transcripts so removal of the duplicates means it
// is no longer possible to work on transcripts. It will also merge
// overlapping exons from different genes on opposite strands.
//
// Despite the caveats, there are use cases where this behaviour is
// exactly what is needed, for example if you are constructing a mask to
// work out which genomic positions are within the exons of a gene
// or set of genes.
func (fs *Features) Consolidate() error {
	if !fs.IsSorted {
		return fmt.Errorf("Consolidate: cannot call on an unsorted Features")
	}

	// Consolidating an empty list of Features is legal but obviously
	// there are no records to be consolidated so let's not waste time.
	if len(fs.Features) == 0 {
		return nil
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

		// Check SeqId
		if keepers[keepidx].SeqId != fs.Features[i].SeqId {
			return fmt.Errorf("Consolidate: cannot call on a Features with mixed SeqId")
		}

		allen := interval.Compare(keepers[keepidx], fs.Features[i])

		// 1. Return error on AllenR of Unknown
		// 2. Return error if b starts before a because that means that
		//    the lists are not sorted.
		// 3. Append to the keepers list if PrecedesB
		// 2. Otherwise merge.
		if allen == interval.Unknown {
			return fmt.Errorf("Consolidate: Allen Relationship is Unknown for {%+v} vs {%+v}",
				keepers[keepidx], fs.Features[i])
		} else if allen == interval.FinishesB ||
			allen == interval.IsContainedByB ||
			allen == interval.IsOverlappedByB ||
			allen == interval.IsMetByB ||
			allen == interval.IsPrecededByB {
			return fmt.Errorf("Consolidate: cannot call on an unsorted Features: {%+v} vs {%+v} for %s",
				keepers[keepidx], fs.Features[i], fs.Id())
		} else if allen == interval.PrecedesB {
			keepers = append(keepers, fs.Features[i])
		} else {
			keepers[keepidx].Merge(fs.Features[i])
		}
	}

	// Attach the list of keepers
	fs.Features = keepers
	return nil
}

// KeepByType takes a list of strings and Feature in the Features
// will be deleted unless Feature.Type *exactly* matches one of the
// supplied strings. An example use case would be to only keep Feature
// that were exons and drop CDS, start_codon, stop_codon etc.
func (fs *Features) KeepByType(keeps []string) int {
	// Create a lookup list of the names we will keep
	keep := make(map[string]int)
	for _, s := range keeps {
		keep[s] = 1
	}

	var kept []*Feature
	var lost int
	for _, f := range fs.Features {
		if _, ok := keep[f.Type]; ok {
			kept = append(kept, f)
		} else {
			lost++
		}
	}

	fs.Features = kept
	return lost
}

func (fs *Features) Count() int {
	return len(fs.Features)
}

// Clone creates a deep copy of a Features, i.e. the new Features shares
// no pointers with the original Features. After calling Clone you can
// change the original Features or the copy without any concern that the
// change will affect the other Features. Obviously your memory use doubles.
func (fs *Features) Clone() *Features {
	nfs := &Features{}
	nfs.Key = fs.Key
	nfs.Value = fs.Value
	nfs.IsSorted = fs.IsSorted

	for _, ogf := range fs.Features {
        // I did try a gob encode/decode here and it was 7x slower!
		nf := ogf.Clone()
		nfs.Features = append(nfs.Features, nf)
	}

	return nfs
}

// DeleteBySeqId removes any Feature that have SeqId that match the
// regexp pattern. It returns the SeqId that were deleted.
//
// DeleteBySeqId is destructive - it removes Feature from the calling
// Features struct. Depending on your use case, you may wish to Clone
// the Features first and then call DeleteBySeqId on the copy.
func (fs *Features) DeleteBySeqId(pattern string) ([]string, error) {
	var deleted []string

	deleter, err := regexp.Compile(pattern)
	if err != nil {
		return deleted, fmt.Errorf("DeleteBySeqId: error compiling pattern %s: %w", pattern, err)
	}

	// If we group all Feature by SeqId first, we have some slice
	// manipulation but a lot less regexp so hopefully faster.
	seqs := fs.BySeqId()

	// Delete any seqs map entries that match
	for k, v := range seqs {
		if deleter.MatchString(v.Value) {
			delete(seqs, k)
			deleted = append(deleted, k)
		}
	}

	// Anything that survived the cull gets put back
	var feats []*Feature
	for _, v := range seqs {
		feats = append(feats, v.Features...)
	}
	fs.Features = feats

	// We are not sure if this messed with sorting so ...
	fs.IsSorted = false

	return deleted, nil
}

// KeepBySeqId keeps any Feature that have SeqId that match the
// regexp pattern. It returns the SeqId that were kept.
//
// KeepBySeqId is destructive - it removes Feature from the calling
// Features struct. Depending on your use case, you may wish to Clone
// the Features first and then call KeepBySeqId on the copy.
func (fs *Features) KeepBySeqId(pattern string) ([]string, error) {
	var kept []string

	keeper, err := regexp.Compile(pattern)
	if err != nil {
		return kept, fmt.Errorf("KeepBySeqId: error compiling pattern %s: %w", pattern, err)
	}

	// If we sort all Feature by SeqId first, we have some slice
	// manipulation but a lot less regexp so hopefully faster.
	seqs := fs.BySeqId()

	// Keep any seqs map entries that match
	for k, v := range seqs {
		if keeper.MatchString(v.Value) {
			kept = append(kept, k)
		} else {
			delete(seqs, k)
		}
	}

	// Anything that survived the cull gets put back
	var feats []*Feature
	for _, v := range seqs {
		feats = append(feats, v.Features...)
	}
	fs.Features = feats

	return kept, nil
}

// SeqIds returns a sorted list of SeqId strings. This is
// useful anywhere that you want consistent ordering. Note that since
// SeqId is always a string the ordering is by string so chromosome
// names may not sort as you'd expect/hope.
func (fs *Features) SeqIds() []string {
	seqids := make(map[string]int)
	for _, f := range fs.Features {
		if _, ok := seqids[f.SeqId]; !ok {
			seqids[f.SeqId] = 0
		}
		seqids[f.SeqId]++
	}
	var names []string
	for k, _ := range seqids {
		names = append(names, k)
	}
	sort.Strings(names)

	return names
}

// Attributes will look at all Features and tally which attributes are
// present and how often.
func (fs *Features) Attributes() map[string]int {
	tally := make(map[string]int)

	for _, f := range fs.Features {
		for k, _ := range f.Attributes {
			tally[k]++
		}
	}

	return tally
}

func (fs *Features) ApplySelector(sel *selector.Selector) error {
	switch sel.Subject {
	case `seqid`:
		_, err := fs.selectBySeqId(sel)
		if err != nil {
			fmt.Errorf("ApplySelector: %w", err)
		}
	default:
		return fmt.Errorf("ApplySelector: selector subject not recognised in: %s", sel)
	}

	return nil
}

// selectBySeqId applies a Selector to modify the Features list.
// It returns a list of strings that are the identifiers of the Feature
// that are lost in the process.  This process is destructive - it
// modifies the source Features.
func (fs *Features) selectBySeqId(sel *selector.Selector) ([]string, error) {
	var results []string
	switch sel.Operation {
	case `keep`:
		results, err := fs.KeepBySeqId(sel.Pattern)
		if err != nil {
			return results, err
		}
	case `delete`:
		results, err := fs.DeleteBySeqId(sel.Pattern)
		if err != nil {
			return results, err
		}
	default:
		return results, fmt.Errorf("selectByColumnSeqid: selector operation not recognised in: %s", sel)
	}

	return results, nil
}

// BySeqId creates a map of Features structs where each Features
// contain Feature with the same SeqId. This can simplify a lot of other
// operations such as Merge and Consolidate because it removes the
// possibility that overlapping ranges are from different sequences.
//
// Note that the new Features use pointers to the Feature from the
// source Features so if you change any Feature in the new map, it
// will change the Feature in the original Features too. If this is
// not what you want to happen, use Clone to get a separate deep
// copy of the source Features before calling BySeqId.
func (fs *Features) BySeqId() map[string]*Features {
	feats := make(map[string]*Features)
	for _, f := range fs.Features {
		if _, ok := feats[f.SeqId]; !ok {
			feats[f.SeqId] = &Features{Key: `seqid`, Value: f.SeqId}
		}
		feats[f.SeqId].Features = append(feats[f.SeqId].Features, f)
	}
	return feats
}

// ByAttributeIdGene is designed specifically for GFF3 files in the
// format used by Ensembl for gene models. It relies on information
// being present in a particular format within the Attributes field
// of the Feature in the GFF3. Here are some edited Feature examples
// from an Ensembl gene model GFF3 showing the Type field and a
// truncated version of the Attribute field:
//
//   Type         Attributes
//   pseudogene   ID=gene:ENSG00000223972;Name=DDX11L1;biotype=pseudogene;...
//   lincRNA_gene ID=gene:ENSG00000243485;Name=MIR1302-10;biotype=lincRNA;...
//   snRNA_gene   ID=gene:ENSG00000222623;Name=RNU6-1100P;biotype=snRNA;...
//   gene         ID=gene:ENSG00000187634;Name=SAMD11;biotype=protein_coding;...
//
//
// ByAttributeIdGene creates a map of Features types where each
// Features collects all of the Feature that relate to a single gene ID.
// It makes a set of assumptions that are specific to records in the
// format of Ensembl gene model GFF3 files:
//
//   1. All genes will have a single Feature with a Type of an Attribute
//      of the form ID=gene:... AND a Type field with value gene
//   2. ID=gene Feature have no parents
//   2. The only child nodes of ID=gene are transcripts so one level of
//      following parent-child relationships will capture all relevant
//      Feature.
func (fs *Features) ByAttrIdGene() map[string]*Features {
	feats := make(map[string]*Features)
	// TO DO - there is no logic in this function yet.
	return feats
}

// PrudentMergeByType examines all of the Feature in a (sorted) instance of
// Features and if they are immediately adjacent or overlap in any way,
// they are consolidated into new Feature. For an explanation of prudent
// merging see
//
//  ajgo merge-gff3 --help
//
// This process is destructive - it changes and deletes Feature from
// Features. For cases where this is undesirable, use Clone to make a
// copy of Features and call PrudentMergeByType on the copy.
func (fs *Features) PrudentMergeByType() error {
	if !fs.IsSorted {
		return fmt.Errorf("PrudentMergeByType: cannot call on an unsorted Features")
	}

	// We will maintain 2 lists of Feature - Keepers and Candidates.
	// Keepers starts empty and Candidate starts with all of the Feature
	// from Features. We will consume Feature from the start of Candidates
	// and append the resultant merged Feature to Keepers.

	// Prime the keepers and candidates lists
	var keepers, candidates []*Feature
	candidates = append(candidates, fs.Features...)

	for {
		// If we have exhausted the candidates or we are down to the last
		// one then we are finished
		if len(candidates) == 0 {
			break
		} else if len(candidates) == 1 {
			keepers = append(keepers, candidates[0])
			break
		}

		// A and B are always the first 2 Feature in Candidates.
		A := candidates[0]
		B := candidates[1]

		// Check SeqId
		if A.SeqId != B.SeqId {
			return fmt.Errorf("PrudentMergeByType: cannot call on a Features with mixed SeqId: {%+v} and {%+v}", A, B)
		}

		nfs, err := PrudentMerge(A, B)
		if err != nil {
			return fmt.Errorf("PrudentMergeByType: error merging {%+v} vs {%+v}: %w", A, B, err)
		}

		// The count of new Feature in nfs tells us what to do:
		// 1. If one Feature came back then there was a complete overlap
		//    - consume A and B from Candidates
		//    - insert the result Feature into Candidates
		// 2. If 2 or more Feature returned then there was some sort of
		//    partial overlap
		//    - consume A and B from Candidates
		//    - append the first result feature to Keepers
		//    - insert the rest of the result Feature into Candidates

		if len(nfs) == 1 {
			candidates = candidates[2:]
			candidates = insertFeatures(candidates, nfs[0])
		} else if len(nfs) == 2 || len(nfs) == 3 {
			candidates = candidates[2:]
			keepers = append(keepers, nfs[0])
			candidates = insertFeatures(candidates, nfs[1:]...)
		} else {
			return fmt.Errorf("PrudentMergeByType: big problem - should be impossible to fall through to here: {%+v} vs {%+v}", A, B)
		}
	}

	// Attach the list of keepers
	fs.Features = keepers
	return nil
}

// AddFeatures appends one or more *Feature. fs will be unsorted at the
// end of the operation. See also AddFeaturesWithSort.
func (fs *Features) AddFeatures(fs2 ...*Feature) {
	fs.Features = append(fs.Features, fs2...)
	fs.IsSorted = false
}

// AddFeatures appends one or more *Feature and sorts.
//
// Before using this function, consider whether it is appropriate.
// It calls Sort every time it is executed and Sort is pretty
// computationally expensive. If you are adding *Feature one at a
// time, this is not the function you should be using. For that use
// case, either gather all of the *Feature to be added and call this
// function once with the full list, OR, call AddFeature multiple times
// (it does not call Sort) and once all *Feature have been added, call
// Sort once in your code.
func (fs *Features) AddFeaturesWithSort(fs2 ...*Feature) {
	fs.Features = append(fs.Features, fs2...)
	fs.Sort()
}

// MergeFeatures merges two *Features.
//
// Under the hood, it uses PrudentMergeByType in a SeqId-safe fashion.
// The returned *Features contains only new and cloned *Feature so it
// can be changed without fear of changing the source *Features.
func MergeFeatures(f1, f2 *Features) *Features {
	// To avoid side effects, we will work with clones
	A := f1.Clone()
	B := f2.Clone()

	// The basic strategy is to smash the two sets of Feature together,
	// sort them by SeqId and then merge within each SeqId.
	nfs := NewFeatures()
	nfs.Key = `merged`
	nfs.Value = A.Id() + `+` + B.Id()

	tfs := NewFeatures()
	tfs.Features = append(A.Features, B.Features...)
	seqs := tfs.BySeqId()
	var seqids []string
	for seqid, fs := range seqs {
		fs.Sort()
		fs.PrudentMergeByType()
		seqids = append(seqids, seqid)
	}

	sort.Strings(seqids)
	for _, seqid := range seqids {
		nfs.Features = append(nfs.Features, seqs[seqid].Features...)
	}

	return nfs
}

// *****  private functions  *****************************************

// ivCountErrMsg is a convenience function to create error messages when
// PrudentMerge returns an unexpected count of subintervals.
func ivCountErrMsg(ar interval.AllenRelationship, A, B *Feature, e, o int) string {
	return fmt.Sprintf("PrudentMergeByType: %s on {%+v} vs {%+v} should have returned %d subintervals, not %d",
		ar, A, B, e, o)
}

// insertFeatures places one or more *Feature into a []*Feature in the correct
// position based on Start so that []*Feature stays sorted. This is a bit
// tricky but messing around with a few slices is way cheaper than doing
// a full sort every time we need to insert a *Feature.
func insertFeatures(fs1 []*Feature, fs2 ...*Feature) []*Feature {
	var fs []*Feature
	fs = append(fs, fs1...)
	for _, f := range fs2 {
		// We are looking for the first element in fs1 that has the same
		// or later Start value than f.
		for i, _ := range fs {
			if fs[i].Start >= f.Start {
				// tmp slices x,y stop tricksy indexing problems with
				// modifying fs on-the-fly.
				var x, y []*Feature
				y = append(y, fs[i:]...)
				x = append(fs[:i], f)
				x = append(x, y...)
				fs = x
				break // move on to splicing in next item from fs2
			}
		}
	}
	return fs
}

// simpleSort is a private function for the nitty gritty logic of
// sorting a *Features. It is used in multiple places. It is *not*
// SeqId-aware. The *Feature will all survive the sort intact.
//
// We are going to use a map to do our sorting. Once all Features have
// been placed into the map by start position, the observed starts
// are sorted and the map is walked doing by-End sorting for any cases
// where there are multiple Features with the same start position.
//
// TO DO - we do not seem to be using the by-end sorting anywhere so we
// should think about whether we should keep this logic. It adds
// computational cost possibly without adding value.
func (fs *Features) simpleSort() {
	// Walk *Feature slice putting them into the map by start position
	sorter := make(map[int][]*Feature)
	for i := 0; i < len(fs.Features); i++ {
		start := int(fs.Features[i].Start)
		if _, ok := sorter[start]; !ok {
			sorter[start] = []*Feature{}
		}
		sorter[start] = append(sorter[start], fs.Features[i])
	}

	// Walk the map by start position, do any required by-End
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
		if len(sorter[start]) == 1 {
			// If there's only one Feature, append it
			sorted = append(sorted, sorter[start]...)
		} else {
			// If there's more than one Feature, we sort by Feature.End
			endSorter := make(map[int][]*Feature)
			for _, f := range sorter[start] {
				if _, ok := endSorter[f.End]; !ok {
					endSorter[f.End] = make([]*Feature, 2)
				}
				endSorter[f.End] = append(endSorter[f.End], f)
			}

			// Order the ends
			ends := []int{}
			for k := range endSorter {
				ends = append(ends, k)
			}
			sort.Ints(ends)

			// Append the Features by end
			for _, end := range ends {
				for _, f := range endSorter[end] {
					sorted = append(sorted, f)
				}
			}
		}
	}

	fs.Features = sorted
}

// Sum Intervals adds the lengths of all of the Feature within *Features.
func (fs *Features) SumIntervals() int {
	var total int
	for _, f := range fs.Features {
		total += f.End - f.Start
	}
	return total
}

// Sum Intervals adds the lengths of all of the Feature within *Features
// but calls Consolidate first. Consolidate is called on a Clone so this
// operation is non-destructive.
func (fs *Features) SumConsolidatedIntervals() (int, error) {
	nfs := fs.Clone()
	err := nfs.Consolidate()
	if err != nil {
		return 0, fmt.Errorf("SumConsolidatedIntervals: %w", err)
	}
	return nfs.SumIntervals(), nil
}
