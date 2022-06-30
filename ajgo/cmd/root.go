// Copyright (c) QIMR Berghofer Medical Research Institute

package cmd

import (
	"github.com/grendeloz/cmdh"
	"github.com/spf13/cobra"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ajgo",
	Short: "A go tool for operations related to AdamaJava",
	Long: `ajgo is a centralised toolbox of operations related to
AdamaJava including parsing XML reports.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cmdh.Initialise(rootCmd, "ajgo", "v0.4.0-dev")
}
