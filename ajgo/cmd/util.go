package cmd

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"github.com/google/uuid"
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// md5sum returns the MD5 hash of a file.  The MD5 provides a signature
// the file, allowing us to check whether two versions are the same.
func md5sum(file string) (string, error) {
	var chk string
	f, err := os.Open(file)
	if err != nil {
		return chk, err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return chk, err
	}

	chk = fmt.Sprintf("%x", h.Sum(nil))
	return chk, nil
}

// gffHeaderFromRunParameters returns a multi-line string that captures
// run parameters.
func gffHeaderFromRunParameters() string {
	return strings.Join(gffHeadersFromRunParameters(), ``)
}

// gffHeadersFromRunParameters returns a multi-line string that captures
// run parameters.
func gffHeadersFromRunParameters() []string {
	run := NewRunParameters()
	return []string{
		"##uuid " + uuid.New().String(),
		"##version " + run.Version,
		fmt.Sprintf("##creation-date %v", run.StartTime),
		fmt.Sprintf("##created-by-user %s (%d)", run.UserName, run.UserId),
		fmt.Sprintf("##created-by-group %s (%d)", run.GroupName, run.GroupId),
		fmt.Sprintf("##created-on %s", run.HostName),
		fmt.Sprintf("##invocation %v", run.Args),
	}
}

// LinesFromFile reads a file and returns the trimmed lines.
func LinesFromFile(file string) ([]string, error) {
	var lines []string

	// Open file
	ff, err := os.Open(file)
	if err != nil {
		return lines, err
	}
	defer ff.Close()

	scanner := bufio.NewScanner(ff)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\n")
		lines = append(lines, line)
	}

	return lines, nil
}

func ConsolidateFilesList(file string, files []string) ([]string, error) {
	var TmpFiles, ViewFiles []string

	if flagFilelistFile != "" {
		files, err := LinesFromFile(flagFilelistFile)
		if err != nil {
			return ViewFiles, fmt.Errorf("problem parsing file: %s", flagFilelistFile)
		}
		TmpFiles = append(TmpFiles, files...)
	}
	if len(flagViewFiles) != 0 {
		TmpFiles = append(TmpFiles, flagViewFiles...)
	}

	// Put the files into a hash and drop any that are duplicates
	chkDups := make(map[string]int)
	for _, file := range TmpFiles {
		if _, ok := chkDups[file]; !ok {
			chkDups[file]++
			ViewFiles = append(ViewFiles, file)
		} else {
			log.Warnf("duplicate file specified: %s", file)
		}
	}

	return ViewFiles, nil
}

func reverseString(s string) string {
	rns := []rune(s) // convert to rune
	for i, j := 0, len(rns)-1; i < j; i, j = i+1, j-1 {
		rns[i], rns[j] = rns[j], rns[i]
	}
	return string(rns)
}

func reverseBytes(b []byte) []byte {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return b
}
