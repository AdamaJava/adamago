// The qpv1 package contains code related to a particular report format from
// qpileup view mode - qpileup view 1. qpv1 is used in the ajgo application
// for determination of uncallable regions of genomes including those with
// low mapping quality and unusually high and low average read depths.
//
// qpv1 code was rolled into a separate package for reusability but also
// because the global constants made it difficult to use the same
// approach to parsing other reports because the same fields might appear
// but in different positions in the report.

package qpv1

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"regexp"
    "strings"
)

// These are convenience constants for referencing fields within a
// []string split from a qpileup view file.
const (
	Reference            int = iota // 0
	Position                        // 1
	Ref_base                        // 2
	A_for                           // 3
	C_for                           // 4
	G_for                           // 5
	T_for                           // 6
	N_for                           // 7
	Aqual_for                       // 8
	Cqual_for                       // 9
	Gqual_for                       // 10
	Tqual_for                       // 11
	Nqual_for                       // 12
	MapQual_for                     // 13
	ReferenceNo_for                 // 14
	NonreferenceNo_for              // 15
	HighNonreference_for            // 16
	LowReadCount_for                // 17
	A_rev                           // 18
	C_rev                           // 19
	G_rev                           // 20
	T_rev                           // 21
	N_rev                           // 22
	Aqual_rev                       // 23
	Cqual_rev                       // 24
	Gqual_rev                       // 25
	Tqual_rev                       // 26
	Nqual_rev                       // 27
	MapQual_rev                     // 28
	ReferenceNo_rev                 // 29
	NonreferenceNo_rev              // 30
	HighNonreference_rev            // 31
	LowReadCount_rev                // 32
)

func ExpectedHeaderFields() []string {
	return []string{`Reference`, `Position`, `Ref_base`,
		`A_for`, `C_for`, `G_for`, `T_for`, `N_for`,
		`Aqual_for`, `Cqual_for`, `Gqual_for`, `Tqual_for`, `Nqual_for`,
		`MapQual_for`, `ReferenceNo_for`, `NonreferenceNo_for`,
		`HighNonreference_for`, `LowReadCount_for`,
		`A_rev`, `C_rev`, `G_rev`, `T_rev`, `N_rev`,
		`Aqual_rev`, `Cqual_rev`, `Gqual_rev`, `Tqual_rev`, `Nqual_rev`,
		`MapQual_rev`, `ReferenceNo_rev`, `NonreferenceNo_rev`,
		`HighNonreference_rev`, `LowReadCount_rev`}
}

func ExpectedHeader() string {
	return `## ` + strings.Join(ExpectedHeaderFields(), "\t")
}

// CheckFileHeader reads the header (comments) from the top of a file
// and checks whether the expected list of column names is there and that
// the names match what we would expect for a qpv1 format view report.
// It will cope with text and gzip'd text files.
func CheckFileHeader(file string) error {
	// Open file
	ff, err := os.Open(file)
	if err != nil {
		return err
	}
	defer ff.Close()

	// We need to define this before we handle gzip
	var scanner *bufio.Scanner

	// Based on file extension, handle gzip files
	found, err := regexp.MatchString(`\.[gG][zZ]$`, file)
	if err != nil {
		return fmt.Errorf("error matching gzip file pattern against %s: %w", file, err)
	}
	if found {
		// For gzip files, put a gzip.Reader into the chain
		reader, err := gzip.NewReader(ff)
		if err != nil {
			return fmt.Errorf("error opening gzip file %s: %w", file, err)
		}
		defer reader.Close()
		scanner = bufio.NewScanner(reader)
	} else {
		// For non gzip files, go straight to bufio.Reader
		scanner = bufio.NewScanner(ff)
	}

	// Unnecessary but explicit
	scanner.Split(bufio.ScanLines)
	expected := ExpectedHeader()

	// Scan forward through all lines starting with '#'. If we get a
	// match then we exit successfully, otherwise if we hit-non comment
	// lines without getting a hit then we exit unsuccessfully.
	if scanner.Scan() {
		line := scanner.Text()
		if line[0:1] == `#` {
			if line == expected {
				return nil
			}
		}
	}

	// If we got here then we fell through without hitting the expected
	// header line we want so exit unsuccessfully.
	return fmt.Errorf("wanted header line not found: %s", expected)
}
