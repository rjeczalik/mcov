// Package covmerge takes the results from multiple `go test -coverprofile` runs and
// merges them into one profile
package covmerge

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"golang.org/x/tools/cover"
)

func mergeProfiles(p *cover.Profile, merge *cover.Profile) (err error) {
	if p.Mode != merge.Mode {
		return fmt.Errorf("cannot merge profiles with different modes: %s != %s", p.Mode, merge.Mode)
	}
	// Since the blocks are sorted, we can keep track of where the last block
	// was inserted and only look at the blocks after that as targets for merge
	startIndex := 0
	for _, b := range merge.Blocks {
		startIndex, err = mergeProfileBlock(p, b, startIndex)
		if err != nil {
			return err
		}
	}
	return nil
}

func mergeProfileBlock(p *cover.Profile, pb cover.ProfileBlock, startIndex int) (int, error) {
	sortFunc := func(i int) bool {
		pi := p.Blocks[i+startIndex]
		return pi.StartLine >= pb.StartLine && (pi.StartLine != pb.StartLine || pi.StartCol >= pb.StartCol)
	}

	i := 0
	if !sortFunc(i) {
		i = sort.Search(len(p.Blocks)-startIndex, sortFunc)
	}
	i += startIndex
	if i < len(p.Blocks) && p.Blocks[i].StartLine == pb.StartLine && p.Blocks[i].StartCol == pb.StartCol {
		if p.Blocks[i].EndLine != pb.EndLine || p.Blocks[i].EndCol != pb.EndCol {
			return 0, fmt.Errorf("conflicting marge overlap: %v %v %v", p.FileName, p.Blocks[i], pb)
		}
		switch p.Mode {
		case "set":
			p.Blocks[i].Count |= pb.Count
		case "count", "atomic":
			p.Blocks[i].Count += pb.Count
		default:
			return 0, fmt.Errorf("unsupported covermode: %s", p.Mode)
		}
	} else {
		if i > 0 {
			pa := p.Blocks[i-1]
			if pa.EndLine >= pb.EndLine && (pa.EndLine != pb.EndLine || pa.EndCol > pb.EndCol) {
				return 0, fmt.Errorf("conflicting start of merge overlap: %v %v %v", p.FileName, pa, pb)
			}
		}
		if i < len(p.Blocks)-1 {
			pa := p.Blocks[i+1]
			if pa.StartLine <= pb.StartLine && (pa.StartLine != pb.StartLine || pa.StartCol < pb.StartCol) {
				return 0, fmt.Errorf("conflicting end of merge overlap: %v %v %v", p.FileName, pa, pb)
			}
		}
		p.Blocks = append(p.Blocks, cover.ProfileBlock{})
		copy(p.Blocks[i+1:], p.Blocks[i:])
		p.Blocks[i] = pb
	}
	return i + 1, nil
}

func addProfile(profiles []*cover.Profile, p *cover.Profile) ([]*cover.Profile, error) {
	i := sort.Search(len(profiles), func(i int) bool { return profiles[i].FileName >= p.FileName })
	if i < len(profiles) && profiles[i].FileName == p.FileName {
		if err := mergeProfiles(profiles[i], p); err != nil {
			return nil, err
		}
	} else {
		profiles = append(profiles, nil)
		copy(profiles[i+1:], profiles[i:])
		profiles[i] = p
	}
	return profiles, nil
}

func dumpProfiles(profiles []*cover.Profile, out io.Writer) error {
	if len(profiles) == 0 {
		return nil
	}
	fmt.Fprintf(out, "mode: %s\n", profiles[0].Mode)
	for _, p := range profiles {
		for _, b := range p.Blocks {
			_, err := fmt.Fprintf(out, "%s:%d.%d,%d.%d %d %d\n", p.FileName, b.StartLine, b.StartCol, b.EndLine, b.EndCol, b.NumStmt, b.Count)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// MergeFiles merges coverage profiles specified by the input list
// into single profile file - the output one.
//
// If it's not possible to either merge profiles, e.g. due to
// conflicting profile block overlaps, or filesystem errors,
// the function returns non-nil error.
//
// TODO(rjeczalik): research recovery for "conflicting merge overlaps".
func MergeFiles(input []string, output string) error {
	dir, file := filepath.Split(output)
	if dir == "" {
		dir = "."
	}

	// Write to a temporary file in case output file is in input slice.
	f, err := ioutil.TempFile(dir, file)
	if err != nil {
		return err
	}

	var merged []*cover.Profile

	for _, file := range input {
		var profiles []*cover.Profile
		profiles, err = cover.ParseProfiles(file)
		if err != nil {
			break
		}
		for _, p := range profiles {
			merged, err = addProfile(merged, p)
			if err != nil {
				break
			}
		}
		if err != nil {
			break
		}
	}

	if err == nil {
		err = dumpProfiles(merged, f)
	}

	if err = nonil(err, f.Close()); err != nil {
		return nonil(err, os.Remove(f.Name()))
	}

	for _, file := range input {
		_ = os.Remove(file)
	}

	return os.Rename(f.Name(), output)
}

func nonil(err ...error) error {
	for _, e := range err {
		if e != nil {
			return e
		}
	}
	return nil
}
