// This file collects package globals used as commandline flags.
// Centralising this list helps us maintain a consistent naming scheme
// and facilitates reuse of flags wherever possible rather than creating
// new ones (with increasingly baroque names) for each mode and submode.

package cmd

var (
	flagOutfileGenome string
	flagInfileGenome  string

	flagOutfileGeneModel string
	flagInfileGeneModel  string

	flagFilelistFile string
	flagFastaFiles   []string
	flagGff3Files    []string
	flagViewFiles    []string
	flagInfile       string
	flagOutfile      string
	flagGobFile      string
	flagName         string

	flagRegionLength int
	flagThreshold    int

	flagGenomeRegion string

	flagOutfileHomopoly string
	flagOutfileExons    string

	flagSelectors []string

	flagDeleteSeqPatterns []string
	flagRegexps           []string
)
