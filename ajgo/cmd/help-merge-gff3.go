package cmd

import (
	"github.com/spf13/cobra"
)

// submode gff3 > merge
var mergingCmd = &cobra.Command{
	Use:   "merge-gff3",
	Short: "principles of merging GFF3 files",
	Long: `
GFF3 files are plain text tab-separated files with structured header
comments that describe genomic features. The formal specification is at:

 https://github.com/The-Sequence-Ontology/Specifications/blob/master/gff3.md

Each Feature has 9 Columns or fields where the first 8 fields have fixed
formats and contain single items of information and the 9th Column is a
structured collection of (effectively unlimited) "Attributes" in key=value
format. The columns are:

 1  SeqId - the sequence on which this feature lies 
 2  Source - algorithm or procedure that generated the feature
 3  Type - the type of the feature (must be SOFA term or accession)
 4  Start - start of the interval
 5  End - the end of the interval
 6  Score
 7  Strand
 8  Phase
 9  Attributes

Note that Start and End assume that the coordinate system is 1-based,
and half-open, i.e. the Start position is within the feature but the End
position is the first base past the feature.

In general, merging GFF3 records must, at a minimum, take account of the
SeqID and the Start and End columns. In some cases additional logic may
be required, for example how to handle the merging of features that are
of different Types.

Depending on your use case, the gff3 > select mode may be helpful
before or after the merge.

Allen Relationships

In 1983, James F. Allen published a paper where he enumerated the
relationships between two time intervals. The paper is available at:
https://cse.unl.edu/~choueiry/Documents/Allen-CACM1983.pdf

Figure 1 shows an interval A (11-17 closed) and a variety of different
intervals (b) and what the relationship is between the intervals.

 Figure 1. Allen Relationships between intervals.

 AR | Diagrammatic Representation |  A ...
 ---|-----------------------------|-----------------
    | 000000000111111111122222222 |
    | 123456789012345678901234567 |
 ---|-----------------------------|-----------------
    |           AAAAAAA           |
  1 |                     BBBBBBB | PrecedesB
  2 |                  BBBBBBB    | MeetsB
  3 |                BBBBBBB      | OverlapsB
  4 |           BBBBBBBBB         | StartsB
  5 |         BBBBBBBBB           | FinishesB
  6 |             BBB             | ContainsB
  7 |           BBBBBBB           | EqualsB
  8 |         BBBBBBBBBBB         | IsContainedByB
  9 |             BBBBB           | IsFinishedByB
 10 |           BBBBB             | IsStartedByB
 11 |      BBBBBBB                | IsOverlappedByB
 12 |    BBBBBBB                  | IsMetByB
 13 | BBBBBBB                     | IsPrecededByB
 ---|-----------------------------|-----------------


Simple Merging

In simple merging, if two intervals overlap, they are merged into a
single new interval. There is a question about whether two immediately
adjacent intervals (relationships MeetsB, IsMetByB) should be merged but
for the purposes of simple merging, adjacency is not considered ito be
overlap and so adjacent intervals are not merged.

Figure 2 is based on Figure 1 and shows, for each of the Allen
Relationships, what intervals would be the result of simple merging with
adjacent intervals NOT merged. If a merged interval is created, it is
shown by the "." character.
 
 Figure 2. Simple Merging

 AR | Diagrammatic Representation |  A ...
 ---|-----------------------------|-----------------
    | 000000000111111111122222222 |
    | 123456789012345678901234567 |
 ---|-----------------------------|-----------------
  1 |           AAAAAAA   BBBBBBB | PrecedesB
  2 |           AAAAAAABBBBBBB    | MeetsB
  3 |           ............      | OverlapsB
  4 |           .........         | StartsB
  5 |         .........           | FinishesB
  6 |           .......           | ContainsB
  7 |           .......           | EqualsB
  8 |         ...........         | IsContainedByB
  9 |           .......           | IsFinishedByB
 10 |           .......           | IsStartedByB
 11 |      ............           | IsOverlappedByB
 12 |    BBBBBBBAAAAAAA           | IsMetByB
 13 | BBBBBBB   AAAAAAA           | IsPrecededByB
 ---|-----------------------------|-----------------


Prudent Merging

Prudent merging (our name) is a different strategy where we split
overlapping intervals so the non-overlapping sub-intervals are separate
from an interval that captures the overlap. This results in more intervals,
not less but it does make it easy to quantify the overlap. For example, if
you had two sets of intervals (X and Y) and you merged them using prudent
merging, you would be able to quantify how many positions were in intervals
just covered by set X, how many positions were in intervals just in Set Y
and how many positions were in intervals from both Set X and Set Y.

 Figure 3. Prudent Merging

 AR | Diagrammatic Representation |  A ...
 ---|-----------------------------|-----------------
    | 000000000111111111122222222 |
    | 123456789012345678901234567 |
 ---|-----------------------------|-----------------
  1 |           AAAAAAA   BBBBBBB | PrecedesB
  2 |           AAAAAAABBBBBBB    | MeetsB
  3 |           AAAAA..BBBBB      | OverlapsB
  4 |           .......BB         | StartsB
  5 |         BB.......           | FinishesB
  6 |           AA...AA           | ContainsB
  7 |           .......           | EqualsB
  8 |         BB.......BB         | IsContainedByB
  9 |           AA.....           | IsFinishedByB
 10 |           .....AA           | IsStartedByB
 11 |      BBBBB..AAAAA           | IsOverlappedByB
 12 |    BBBBBBBAAAAAAA           | IsMetByB
 13 | BBBBBBB   AAAAAAA           | IsPrecededByB
 ---|-----------------------------|-----------------

To better understand prudent merging, let's examine some of the
relationships in more detail. 

Relationship 3, AOverlapsB, would result in 3 intervals post merge.
The first new interval would be a (half-open) interval 11-16 which
contains positions that were in interval A and did not overlap with
interval B. The second new interval would be 16-18 which captures the
positions that were in both intervals A and B. The third interval would
be 18-23 which is those positions that were in interval B and did not
overlap with interval A. This would allow us to easily determine that
pre-merge, intervals A and B each covered 7 positions and post-merge 
there were 5 positions in A, 5 positions in B and 2 positions that
overlapped. With simple merging, you would end up with a single 12
position interval and in order to do the overlap math, you would need
the merged interval to store the information that it was the result of
merging a 7-position A interval with a 7 position B interval.

Relationship 8, AIsContainedByB would result in 3 intervals post-merge.
The first (8-11) and third (18-20) intervals were in interval B but did
not overlap with A whereas the second interval (11-18) shows the overlap
which is the entirety of interval A because it was entirely contained
within interval B.`,
}

func init() {
	rootCmd.AddCommand(mergingCmd)
}
