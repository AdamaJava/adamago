package cmd

import (
	"encoding/json"
	"io"
	"os"

	ngc "ajgo/ngc"
	log "github.com/sirupsen/logrus"
    "github.com/grendeloz/cmdh"
	"github.com/spf13/cobra"
)

// summaryCmd represents the summary command
var ngscheckDebugCmd = &cobra.Command{
	Use:   "debug",
	Short: "check that .ngcbas.json file can be parsed",
	Long: `Parse an NGScheck basic mode JSON output file into generic
data structure for downstream use.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdh.StartLogging()
		ngscheckDebugCmdRun(cmd, args)
		cmdh.FinishLogging()
	},
}

func init() {
	ngscheckCmd.AddCommand(ngscheckDebugCmd)

	ngscheckDebugCmd.Flags().StringVar(&flagInfile, "ngcbas-json", "",
		"NGScheck basic mode JSON report file to be parsed")
	ngscheckDebugCmd.MarkFlagRequired("ngcbas-json")
}

func ngscheckDebugCmdRun(cmd *cobra.Command, args []string) {
	// Open our JSON
	log.Info("processing: ", flagInfile)
	jsonFile, err := os.Open(flagInfile)
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
	byteValue, _ := io.ReadAll(jsonFile)

	// initialize our NgcBasicGrafliD structure and unmarshal
	var ngb ngc.NgscheckBasic
	err = json.Unmarshal(byteValue, &ngb)
	if err != nil {
		log.Fatal("error while unmarshalling: ", err)
	}
	(&ngb).Finalise()

	log.Infof("Parsed struct: %+v", ngb)
}
