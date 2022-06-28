package cmd

import (
	"github.com/spf13/cobra"
)

var selectorCmd = &cobra.Command{
	Use:   "selector",
	Short: "help on selectors and their use",
	Long: `
A Selector is a string of the form operation:subject:pattern which can
be used to select/filter records in the ajgo system. For example, valid
selectors that can be used against GFF3 records include:

  keep:seqid:^GL
  delete:type:.*_UTR

In general, the effect of all selectors is to drop records from some
data structure. Selectors with delete operations drop any record that
matches the pattern while keep operations drop any record that does
not match the pattern - in both cases records are dropped.

As a general rule, a keep operation is a blunter tool than a delete
operation because every record you wish to keep must match the keep
pattern. Because delete only drops matching records, multiple delete 
selectors, applied sequentially, with tight patterns can be used to
selectively prune away records that you don't wish to retain.

In cases where multiple selectors are allowed, they are applied
sequentially in the order in which they appeared on the command line.

Every ajgo mode that uses selectors has its own list of valid operations
and subjects but pattern is always a regex. The colon character ':' must
not be used in the subject, operation or pattern of a selector - it is
strictly reserved as a separator for the selector. ajgo will exit with
an error if an invalid subject or operation is specified. Selectors are
used in modes including:

  ajgo genemodel select`,
}

func init() {
	rootCmd.AddCommand(selectorCmd)
}
