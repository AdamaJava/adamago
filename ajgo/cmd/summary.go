/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	"adamago/xml"
	"github.com/spf13/cobra"
)

var qmotifXmlFile string

// summaryCmd represents the summary command
var qmotifSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "print summary values from qmotif XML files",
	Long: `Parse qmotif XML files and write out parameters from the summary
section.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("summary called")
		summaryQmotifCmdRun(cmd, args)
	},
}

func init() {
	qmotifCmd.AddCommand(qmotifSummaryCmd)

	qmotifSummaryCmd.Flags().StringVar(&qmotifXmlFile, "xmlfile", "",
		"qmotif XML file to be parsed")
	qmotifSummaryCmd.MarkFlagRequired("xmlfile")
}

func summaryQmotifCmdRun(cmd *cobra.Command, args []string) {
	c, err := gnqcreds.Read()
	if err != nil {
		log.Fatal(err)
	}
	j, err := c.ToJson()
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Printing credentials file as JSON:")
	fmt.Printf("%v \n", j)

	// Open our xmlFile
	xmlFile, err := os.Open(qmotifXmlFile)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully Opened", qmotifXmlFile)
	defer xmlFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(xmlFile)

	// initialize our qmotif structure
	var qmotif xml.Qmotif
	// unmarshal our byteArray into the qmotif data structure
	xml.Unmarshal(byteValue, &qmotif)
}
