package cmd

import (
	"fmt"
	"github.com/grendeloz/cmdh"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print the gnqac version number",
	Long:  `Print the gnqac version number and exit.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s %s\n", cmdh.Tool(), cmdh.Version())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
