// This file centralises initialisation and logging for GRZ-style cobra
// commander applications.
//
// To make this system work, we require a number of global cobra
// PersistentFlags to be defined in the root cobra.Command (rootCmd
// in our example). The content below can appear in any init() but they
// must be defined against the root cobra.Command.

/*
func init() {
    // Persistent flags, global for the application.
    rootCmd.PersistentFlags().StringVar(&globals.cfgFile, "config",
        "", "config file (default is $HOME/.ajgo.yaml)")
    rootCmd.PersistentFlags().StringVar(&globals.logFile, "logfile",
        "", "log file (defaults to STDERR if no file specified)")
    rootCmd.PersistentFlags().StringVar(&globals.logLevel, "loglevel",
        "INFO", "log level")
    rootCmd.PersistentFlags().BoolVar(&globals.verbose, "verbose",
        false, "turn on verbose messaging")
}
*/

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// defined as a var to force early execution, even before init()s
	run     = NewRunParameters()
	globals globalFlags
)

// global flags which must be initialised via the init() in rootCmd.
type globalFlags struct {
	cfgFile  string
	logFile  string
	logLevel string
	verbose  bool
}

// execution information
type RunParameters struct {
	Version   string
	StartTime time.Time
	Args      []string
	UserId    int
	UserName  string
	GroupId   int
	GroupName string
	HostName  string
}

// Initialise and start logging. Note that this can not happen until after
// cobra flags have been parsed, assuming that we are allowing users to
// set values for logfile and loglevel.
func startLogging() {
	// Use our custom formatter
	formatter := LogFormat{}
	formatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(&formatter)

	// Should fail if user-supplied logfile already exists
	if globals.logFile != "" {
		file, err := os.OpenFile(globals.logFile,
			os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
		if err == nil {
			log.SetOutput(file)
		} else {
			// Using fmt and os.Exit here because logging is not established.
			fmt.Println("unable to log to file", globals.logFile, ":", err)
			os.Exit(1)
		}
	}

	// cobra.PersistentFlags() handles the defaulting so globals.logLevel
	// will be set to INFO if no level was supplied by the user.
	switch strings.ToUpper(globals.logLevel) {
	case "FATAL":
		log.SetLevel(log.FatalLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	default:
		// This can only happen if the user sets a loglevel and it's not
		// one of the expected values.
		log.Fatalf("%v is not a recognised loglevel", globals.logLevel)
	}

	// Log key execution parameters
	log.Info("Tool: ", run.Version)
	log.Info("Cmdline: ", run.Args)
	log.Info("Host: ", run.HostName)
	log.Infof("User: %d (%s)", run.UserId, run.UserName)
	log.Infof("Group: %d (%s)", run.GroupId, run.GroupName)

	// Read config file (default or user-supplied)
	grzInitConfig()
	log.Infof("Config file: %v", viper.ConfigFileUsed())

	//return true
}

// Finishes logging including elapsed time
func finishLogging() {
	end := time.Now()
	elapsed := end.Sub(run.StartTime)
	log.Info("Elapsed time: ", elapsed)
}

// The LogFormat struct and Format function below are based on info from:
// stackoverflow questions/48971780/change-format-of-log-output-logrus

// LogFormat is a custom format for log messages (via logrus)
type LogFormat struct {
	TimestampFormat string
}

// Format method (on LogFormat) implements our custom logrus log format
func (f *LogFormat) Format(entry *log.Entry) ([]byte, error) {
	var b *bytes.Buffer

	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	b.WriteString(entry.Time.Format(f.TimestampFormat))
	b.WriteString(" [")
	b.WriteString(strings.ToUpper(entry.Level.String()))
	b.WriteString("]")

	if entry.Message != "" {
		b.WriteString(" - ")
		b.WriteString(entry.Message)
	}

	if len(entry.Data) > 0 {
		b.WriteString(" || ")
	}
	for key, value := range entry.Data {
		b.WriteString(key)
		b.WriteByte('=')
		b.WriteByte('{')
		fmt.Fprint(b, value)
		b.WriteString("}, ")
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

// Return a record with execution parameters
func NewRunParameters() RunParameters {
	userId := os.Getuid()
	groupId := os.Getgid()

	// Systems that use LDAP for user management (e.g. Avalon) bork when
	// trying to get the names to match the UID/GID numbers so we are
	// going to silently ignore errors on those functions.
	userName := ""
	tmpUserName, err := user.LookupId(strconv.Itoa(userId))
	if err == nil {
		userName = tmpUserName.Name
	}
	groupName := ""
	tmpGroupName, err := user.LookupGroupId(strconv.Itoa(groupId))
	if err == nil {
		groupName = tmpGroupName.Name
	}

	hostName, err := os.Hostname()
	if err != nil {
		log.Fatal("Error:", err)
	}

	// Setup and return RunParameters
	var run RunParameters
	run.StartTime = time.Now()
	run.Version = "ajgo v0.3.0-dev"
	run.Args = os.Args
	run.UserId = userId
	run.UserName = userName
	run.GroupId = groupId
	run.GroupName = groupName
	run.HostName = hostName
	return run
}

// grzInitConfig reads in config file and ENV variables if set. It is
// called from startLogging() so users do not need to call it themselves.
func grzInitConfig() {
	if globals.cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(globals.cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".c2md" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".c2md")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Info("Using config file:", viper.ConfigFileUsed())
	}
}
