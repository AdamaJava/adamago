package gff3

import (
	"strings"
)

// We have intentionally kept Tree as a separate structure derived
// from (but not embedded within) a Gff3. This is because maintaining a
// Tree inside a Gff3 during the application of Selectors is way too
// hard. The model we have adopted means that if you use Selectors on a
// Gff3, use Gff3.NewTree() after the selecting is done and you will
// have a nice clean Tree that only contains Nodes and
// Features that appear in the post-Selector Gff3.
//
// This structure does its best to implement the Gff3 spec at:
// https://github.com/The-Sequence-Ontology/Specifications/blob/master/gff3.md
//
// There are 2 attributes of GFF3 Feature records that provide grouping
// - ID and Parent. A record may only have one ID although multiple
// Feature records may share a single ID as long as they collectively
// constitute a single Feature - e.g. a CDS which is a discontiguous feature
// defined via multiple ranges. A Feature may only have one Parent
// attribute but it may contain multiple comma-separated parent IDs
// because a Feature is allowed to belong to multiple groups - e.g. an
// exon Feature may be part of multiple transcripts and so will list
// each transcript it is part of as a Parent.
//
// Of Nodes, Leaves and Orphans.
//
// We will be dividing Features into three classes depending on how they
// relate to other Features, i.e. which attributes they have.
//
// Node.
// Features with ID attributes are Nodes because they may have child
// Features. The Parent attribute always refers to an ID so only
// Features with IDs (Nodes) can be Parents. Examples include gene.
//
// Leaf.
// We call Features that do not have an ID but do have a Parent, Leaves.
// Leaves belong to at least one group (Parent) so they link to at least
// one Parent Node. Examples include exon (in Ensembl GFF3). In
// practice, Leaves are easily turned into Nodes by giving them an ID. In
// Ensembl GFF3 gene models, exons have a Name which is the unique ENSEx
// identifier so this could just have easily been set up as an ID.
//
// Orphan Leaf.
// Some Feature types (e.g. biological_region) do not have ID or Parent
// attributes so they exist in splendid isolation. Because they don't
// relate to anything else, we stick them in their own map under Tree.
//
// We will keep the Feature tree as a simple ID-based map (as well as
// having the Node<->Node links) because a single-layer map lookup is
// (probably) going to be quicker (and simpler to code) to
// find nodes than a tree search/walk.

// Tree is a representation of the Features from a Gff3.
type Tree struct {
	Nodes   map[string]*TreeNode
	Orphans []*Feature
}

// TreeNode is a collection of Features with the same IdString. Child
// Features are kept in two different slices depending on whether they
// themselves are a Node or a Leaf. A Feature may list multiple Parents
// so the tree is more of a web.
//
// Also note that a Gff3 may contain multiple Feature with the same ID as
// long as they collectively form one logical Feature. An example is CDS
// which is a discontiguous Feature, i.e. there is one CDS per transcript
// BUT across multiple genomic ranges so a single CDS appears as multiple
// lines in the Gff3. Because of this, Self must be a slice so it can
// contain multiple Feature.
type TreeNode struct {
	IdString    string
	Self        []*Feature
	ChildLeaves []*Feature
	Parents     []*TreeNode
	ChildNodes  []*TreeNode
}

func NewTree() *Tree {
	nodes := make(map[string]*TreeNode)
	return &Tree{Nodes: nodes}
}

// NodeById takes an ID string and finds the matching TreeNode. If no
// such node exists, it is created. This behaviour is needed because we
// are not certain that a Feature always appears in a GFF3 file before any
// of it's children, so even if we haven't seen a Node yet, we need to be
// able to store its children and once we find the Node Feature, we can
// fill in the Self and Parent pointers.
func (t *Tree) NodeById(id string) *TreeNode {
	if _, ok := t.Nodes[id]; !ok {
		n := &TreeNode{}
		t.Nodes[id] = n
	}
	return t.Nodes[id]
}

// Features recursively extracts a list of all Features and sub-Features
// of a TreeNode.
func (n *TreeNode) Features() []*Feature {
	var feats []*Feature
	// Self
	feats = append(feats, n.Self...)
	// ChildLeaves
	feats = append(feats, n.ChildLeaves...)
	// ChildNodes - recursive
	for _, c := range n.ChildNodes {
		feats = append(feats, c.Features()...)
	}
	return feats
}

// NewGffTree builds a tree structure from a Gff3. Note that we will
// link TreeNodes so they point to their child nodes as well as their
// parent. This will let us go gene->transcript as well as
// gene<-transcript.
func (g *Gff3) NewTree() *Tree {
	t := NewTree()

	// TO DO
	// This entire function should probably move to features.go and
	// become a receiver on *Features. This function can be kept but it
	// becomes a minimalist wrapper.

	//var ctr int = 0
	for _, f := range g.Features.Features {
		//if ctr > 50 {
		//	log.Fatal("I'm goin'")
		//}
		//ctr++
		//log.Infof("ID: %s  Parent: %s", f.Attributes[`ID`], f.Attributes[`Parent`])

		// Nodes, Leaves and Orphan Leaves are treated differently
		if _, ok := f.Attributes[`ID`]; ok {
			//log.Info("node:  ", f.AttributesString())
			// Has ID: Node
			n := t.NodeById(f.Attributes[`ID`])
			n.Self = append(n.Self, f)
			if _, ok := f.Attributes[`Parent`]; ok {
				parents := strings.Split(f.Attributes[`Parent`], `,`)
				for _, parent := range parents {
					p := t.NodeById(parent)
					n.Parents = append(n.Parents, p)
					p.ChildNodes = append(p.ChildNodes, n)
				}
			}
		} else if _, ok := f.Attributes[`Parent`]; ok {
			//log.Info("leaf:  ", f.AttributesString())
			// No ID but has Parent: Leaf
			parents := strings.Split(f.Attributes[`Parent`], `,`)
			for _, parent := range parents {
				p := t.NodeById(parent)
				p.ChildLeaves = append(p.ChildLeaves, f)
			}
		} else {
			//log.Info("orphan:  ", f.AttributesString())
			// No ID and no Parent: Orphan Leaf
			t.Orphans = append(t.Orphans, f)
		}
	}
	return t
}
