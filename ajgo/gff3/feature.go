package gff3

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/grendeloz/interval"
)

// The names here are based on http://gmod.org/wiki/GFF3
type Feature struct {
	SeqId      string
	Source     string
	Type       string
	Start      int
	End        int
	Score      string // should be float but missing is "."
	Strand     string
	Phase      string // should be int but missing is "."
	Attributes map[string]string
	LineNumber int // Line number within the Gff3 file
}

// Satisfy interval.Interval interface
func (f *Feature) Low() int {
	return f.Start
}
func (f *Feature) High() int {
	return f.End
}

// NewFeature returns a new instance of type Feature.
//
// Feature fields have whatever the default value is for their type
// with the exception of Source which is set to `ajgo`, Type which is
// set to the root SOFA accession `SO:0000110`, and Score/Strand/Phase
// which are all set to their missing value `.`.
func NewFeature() *Feature {
	// Set defaults where useful
	attrs := make(map[string]string)
	return &Feature{Source: `ajgo`,
		Type:       `SO:0000110`,
		Score:      `.`,
		Strand:     `.`,
		Phase:      `.`,
		Attributes: attrs}
}

// NewFeatureFromLine takes a single line of text, strips the line
// endings, if any, creates a GFF3Feature, and returns a pointer to it.
func NewFeatureFromLine(line string) (*Feature, error) {
	var feat Feature

	line = strings.TrimSuffix(line, "\n")
	fields := strings.Split(line, "\t")
	if len(fields) != 8 && len(fields) != 9 {
		return nil, fmt.Errorf("NewFeatureFromLine: %d fields supplied - 8 or 9 are required", len(fields))
	}

	feat.SeqId = fields[0]
	feat.Source = fields[1]
	feat.Type = fields[2]
	if i, err := strconv.ParseInt(fields[3], 10, 64); err != nil {
		return nil, fmt.Errorf("NewFeatureFromLine: Feature.Start error converting %s to int64: %w", fields[3], err)
	} else {
		feat.Start = int(i)
	}
	if i, err := strconv.ParseInt(fields[4], 10, 64); err != nil {
		return nil, fmt.Errorf("NewFeatureFromLine: Feature.End error converting %s to int64: %w", fields[4], err)
	} else {
		feat.End = int(i)
	}
	feat.Score = fields[5]
	feat.Strand = fields[6]
	feat.Phase = fields[7]

	if len(fields) == 9 {
		// Attributes is full of leading/trailing spaces which upsets splitting
		// on space so brace yourself for profligate use of TrimSpace.
		feat.Attributes = make(map[string]string)
		splitable := strings.TrimSpace(fields[8])
		if splitable != "" {
			attributes := strings.Split(splitable, ";")
			var key, val string
			for _, a := range attributes {
				// There is often (always?) a trailing empty attribute and we don't
				// want empty stuff in the map so skip any empty attributes
				if a == "" {
					continue
				}
				// Much more space trimming required here
				a := strings.TrimSpace(a)
				subs := strings.SplitN(a, "=", 2)
				// The split doesn't remove white space
				if len(subs) == 2 {
					key = strings.TrimSpace(subs[0])
					val = strings.TrimSpace(subs[1])
					feat.Attributes[key] = val
				} else if len(subs) == 1 {
					key = strings.TrimSpace(subs[0])
					feat.Attributes[key] = ""
				} else {
					// This should be impossible because we already checked
					// that field[8] exists and != ""
				}
			}
		}
	}
	return &feat, nil
}

// Merge will merge the supplied Feature (b) on top of the current
// Feature. Note that this is a very blunt tool and is almost certainly
// NOT what you want to do unless you have already done an Allen
// Relationship-aware comparison on the two Features and determined that
// they should be merged.
//
// The only check done before the Merge is that the SeqId of both Features
// is identical or an error is returned.
//
// The Start of the merged Feature is the lesser of the Starts of the
// two Features and the End is the greater of the Ends of the two
// Features. All other fields are set to missing if the Features have
// different values, otherwise the values are left alone, i.e. they
// will stay set to the value that the Features share.
//
// Attributes are deleted unless they are present and identical in
// both Features.
//
// Merge is destructive - it directly modifies the Feature (a) it is
// called against.
//
// You may also find MergeWithSplit useful.
//
func (a *Feature) Merge(b *Feature) error {
	// This should *never* happen but it's still worth a sanity check.
	if a.SeqId != b.SeqId {
		return fmt.Errorf("Merge: cannot merge Feature with different SeqId's: %s %s",
			a.SeqId, b.SeqId)
	}

	// Do the merge.
	a.merge(b)

	return nil
}

// This private function does the merge without any checks. If checks
// are needed, you can implement those in public wrapper functions  -
// see the code for Merge as an example.
func (a *Feature) merge(b *Feature) {
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
	if a.Type != b.Type {
		a.Type = `.`
	}
	if a.Score != b.Score {
		a.Score = `.`
	}
	if a.Strand != b.Strand {
		a.Strand = `.`
	}
	if a.Phase != b.Phase {
		a.Phase = `.`
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
}

func (f *Feature) String() string {
	// Attributes are ;-separated
	attrString := f.AttributesString()

	// Fields are tab-separated
	output := strings.Join([]string{
		f.SeqId,
		f.Source,
		f.Type,
		strconv.Itoa(int(f.Start)),
		strconv.Itoa(int(f.End)),
		f.Score,
		f.Strand,
		f.Phase,
		attrString}, "\t")

	return output
}

// Makes a deep copy of a Feature, i.e. the new Feature shares no
// pointers with the original Feature. After Clone, you can change
// either Feature without any concern that the change will affect
// the other copy.
//
// This is especially useful for destructive operations that change a
// Feature (e.g. Merge) and you are not sure what other data structures
// might have a pointer to, and be relying upon, that Feature.
func (f *Feature) Clone() *Feature {
	// Caveat Emptor! I used to have a lazy cloning method that used
	// a round-trip through string:
	//   nf, _ := NewFeatureFromLine(f.String())
	// Moving to the custom cloning code below made Clone 3 times faster.
	n := &Feature{
		SeqId:      f.SeqId,
		Source:     f.Source,
		Type:       f.Type,
		Start:      f.Start,
		End:        f.End,
		Score:      f.Score,
		Strand:     f.Strand,
		Phase:      f.Phase,
		Attributes: make(map[string]string, len(f.Attributes)),
		LineNumber: f.LineNumber,
	}
	for k, v := range f.Attributes {
		n.Attributes[k] = v
	}
	return n
}

// debugString is a private function for use in testing an debugging. It
// converts a Feature to a string but with ~ as the separator instead of
// tab. This can make it easier to do visual debugging since whitespace
// within field values is easy to spot when the separator is not also
// whitespace.
func (f *Feature) debugString() string {
	// Attributes are ;-separated
	attrString := f.AttributesString()

	// Fields are ~-separated
	output := strings.Join([]string{
		f.SeqId,
		f.Source,
		f.Type,
		strconv.Itoa(int(f.Start)),
		strconv.Itoa(int(f.End)),
		f.Score,
		f.Strand,
		f.Phase,
		attrString}, `~`)

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
			attrStrings = append(attrStrings, s+"="+f.Attributes[s])
		}
	}
	// Attributes are ;-separated
	attrString := strings.Join(attrStrings, ";")

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

	// Assemble attributes string
	var attrStrings []string
	for _, k := range keys {
		attrStrings = append(attrStrings, k+"="+f.Attributes[k])
	}
	attrString := strings.Join(attrStrings, ";")

	return attrString
}

// PrudentMerge does an interval.Compare on a pair of sorted *Feature
// and returns a slice of non-overlapping *Feature that cover the
// same bases as A and B but with any overlap represented as a separate
// Feature. A and B must be sorted so that A.Start <= B.Start.
//
// Note that the *Feature passed in are not altered - the *Feature
// returned are new or Clone'd from the inputs.
func PrudentMerge(a, b *Feature) ([]*Feature, error) {
	if a.Start > b.Start {
		return nil, fmt.Errorf("PrudentMerge: A.Start must be <= B.Start: {%+v} vs {%+v}", a, b)
	}

	var nfs []*Feature
	A := a.Clone()
	B := b.Clone()
	allen := interval.Compare(A, B)

	// We do not need to handle all Allen Relationships - given that A and B
	// are sorted by Start and we have already enforced that requirement at
	// the top of this func, we can be selective here. We have a sentinel else
	// clause in case it turns out we have had a failure of imagination. :)
	//
	// In the comments below ' is used to denote truncation and which end
	// of the interval was shortened. So A' is A shortened by reducing the
	// End and 'A is A shortened by increasing the Start.
	//
	// Any Feature that is an overlap has its Type set to the generic
	// SOFA accession SO:0000110 with an Attribute of Sources which
	// contains a list of the Source fields from the merged Feature.

	if allen == interval.PrecedesB || allen == interval.MeetsB {
		nfs = append(nfs, A, B)
	} else if allen == interval.OverlapsB {
		// 3 Features - A', overlap, 'B
		O := newOverlapFeature(A, B)
		O.Start = B.Start
		O.End = A.End
		A.End = B.Start
		B.Start = A.End
		nfs = append(nfs, A, O, B)
	} else if allen == interval.StartsB {
		// 2 Features - overlap, 'B
		O := newOverlapFeature(A, B)
		O.Start = A.Start
		O.End = B.Start
		B.Start = A.End
		nfs = append(nfs, O, B)
	} else if allen == interval.ContainsB {
		// 3 Features - A', overlap, 'A
		O := newOverlapFeature(A, B)
		O.Start = B.Start
		O.End = B.End
		A2 := A.Clone()
		A2.Start = B.End
		A.End = B.Start
		nfs = append(nfs, A, O, A2)
	} else if allen == interval.EqualsB {
		// 1 Feature - overlap
		O := newOverlapFeature(A, B)
		O.Start = A.Start
		O.End = A.End
		nfs = append(nfs, O)
	} else if allen == interval.IsFinishedByB {
		// 2 Features - A', overlap
		O := newOverlapFeature(A, B)
		O.Start = B.Start
		O.End = B.End
		A.End = B.Start
		nfs = append(nfs, A, O)
	} else if allen == interval.IsStartedByB {
		// 2 Features - overlap, 'A
		O := newOverlapFeature(A, B)
		O.Start = B.Start
		O.End = B.End
		A.Start = B.End
		nfs = append(nfs, O, A)
	} else if allen == interval.Unknown {
		return nil, fmt.Errorf("PrudentMerge: Allen Relationship is Unknown for {%+v} vs {%+v}", A, B)
	} else {
		return nil, errors.New("PrudentMerge: big big problem - should be impossible to fall through to here")
	}

	return nfs, nil
}

// newOverlapFeature uses A and B as the basis for a new *Feature.
// It uses two Attributes - Sources and Types - to track the Source and
// Type fields of all Feature that contributed to the merged Feature.
//
// Values of the overlap Feature are:
//
//  SeqId - taken from A with no check that A and B have the same SeqId
//  Source  - set to `ajgo-merge`
//  Attributes - Sources and Types set as noted above
//
// All other fields are set to whatever value is supplied by
// NewFeature().

func newOverlapFeature(A, B *Feature) *Feature {
	C := NewFeature()
	C.SeqId = A.SeqId
	C.Source = `ajgo-merge`
	sep := `,`

	// Create non-redundant map of Source
	sources := make(map[string]int)
	if A.Source == `ajgo-merge` {
		srcs := strings.Split(A.Attributes[`Sources`], sep)
		for _, s := range srcs {
			sources[s]++
		}
	} else {
		sources[A.Source]++
	}
	if B.Source == `ajgo-merge` {
		srcs := strings.Split(B.Attributes[`Sources`], sep)
		for _, s := range srcs {
			sources[s]++
		}
	} else {
		sources[B.Source]++
	}
	var finalSources []string
	for k, _ := range sources {
		finalSources = append(finalSources, k)
	}
	sort.Strings(finalSources)

	// Create non-redundant map of Type
	types := make(map[string]int)
	if A.Source == `ajgo-merge` {
		typs := strings.Split(A.Attributes[`Types`], sep)
		for _, t := range typs {
			types[t]++
		}
	} else {
		types[A.Type]++
	}
	if B.Source == `ajgo-merge` {
		typs := strings.Split(B.Attributes[`Types`], sep)
		for _, t := range typs {
			types[t]++
		}
	} else {
		types[B.Type]++
	}
	var finalTypes []string
	for k, _ := range types {
		finalTypes = append(finalTypes, k)
	}
	sort.Strings(finalTypes)

	// Create non-redundant map of ID
	ids := make(map[string]int)
	if A.Source == `ajgo-merge` {
		tids := strings.Split(A.Attributes[`IDs`], sep)
		for _, i := range tids {
			ids[i]++
		}
	} else {
		if _, ok := A.Attributes[`ID`]; ok {
			ids[A.Attributes[`ID`]]++
		}
	}
	if B.Source == `ajgo-merge` {
		tids := strings.Split(B.Attributes[`IDs`], sep)
		for _, i := range tids {
			ids[i]++
		}
	} else {
		if _, ok := B.Attributes[`ID`]; ok {
			ids[B.Attributes[`ID`]]++
		}
	}
	var finalIds []string
	for k, _ := range ids {
		finalIds = append(finalIds, k)
	}
	sort.Strings(finalIds)

	C.Attributes[`IDs`] = strings.Join(finalIds, sep)
	C.Attributes[`Sources`] = strings.Join(finalSources, sep)
	C.Attributes[`Types`] = strings.Join(finalTypes, sep)
	return C
}
