// Package mcov provides utilities for gathering multiple coverage
// profiles for either the same or multiple different packages.
package mcov

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/rjeczalik/mcov/covmerge"
)

var counter int32

func seq() int {
	return int(atomic.AddInt32(&counter, 1))
}

// SubprofileMerge merges coverage subprofile files produced
// by go test.
//
// If coverprofile is empty, the function reads main profile
// file from -test.coverprofile flag instead.
//
// If coverprofile is empty and either no coverprofile is available
// in os.Args or coverprofile file does not exist, the function is a nop.
//
// A subprofile file is a coverage profile that was produced
// in a separate process, which was invoked within the test
// executed by go test.
//
// For example if go test was executed with -coverprofile=coverage.txt
// flag, it may produce the following subprofile files:
//
//   coverage.txt.1
//   coverage.txt.2
//   coverage.txt.3
//
// SubprofileMerge merges the above files back with the original
// coverage.txt profile. The function is best called within TestMain,
// after m.Run() when coverprofile file is saved.
//
// If subprofiles failed to merged or there were filesystem errors
// while reading / writing the files, the function returns non-nil
// error.
func SubprofileMerge(coverprofile string) error {
	if coverprofile == "" {
		for _, arg := range os.Args {
			if arg == "--" {
				break
			}
			if strings.HasPrefix(arg, "-test.coverprofile=") {
				coverprofile = filepath.Base(arg[len("-test.coverprofile="):])
				break
			}
		}
	}

	if coverprofile == "" {
		return nil // nop
	}

	if _, err := os.Stat(coverprofile); os.IsNotExist(err) {
		return nil // nop
	} else if err != nil {
		return err
	}

	input := []string{coverprofile}

	for i := 1; ; i++ {
		subprofile := fmt.Sprintf("%s.%d", coverprofile, i)

		if _, err := os.Stat(subprofile); os.IsNotExist(err) {
			break
		} else if err != nil {
			return err
		}

		input = append(input, subprofile)
	}

	return covmerge.MergeFiles(input, coverprofile)
}

// SubprofileFlags appends coverage flags to the given flags.
//
// If os.Args contains non-empty value for a -test.coverprofile flag,
// the same value suffixed with a sequence number (a subprofile path)
// will be appended to flags.
//
// If os.Args contains -test.covermode flag, the same flag will be
// appended to flags.
//
// If os.Args contains -test.coverpkg flag, the same flag will be
// appended to flags.
//
// If os.Args contained none of the above, the function is a nop
// and untouched flags will be returned.
//
// The function respects end-of-options tag (--).
func SubprofileFlags(flags []string) []string {
	for _, arg := range os.Args {
		if arg == "--" {
			break
		}
		if strings.HasPrefix(arg, "-test.coverprofile=") {
			subprofile := fmt.Sprintf("%s.%d", filepath.Base(arg[len("-test.coverprofile="):]), seq())
			flags = append(flags, "-test.coverprofile="+subprofile)
		}
		if strings.HasPrefix(arg, "-test.covermode=") {
			flags = append(flags, arg)
		}
		if strings.HasPrefix(arg, "-test.coverpkg=") {
			flags = append(flags, arg)
		}
	}
	return flags
}
