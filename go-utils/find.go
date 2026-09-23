// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026 Sonar developers
//
// From https://github.com/NordicHPC/sonar/util/anonymize on 23-Sep-2026; since modified.

// Utilities for enumerating files.

package utils

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"
)

const (
	maxDepth = 8
)

// FindFiles finds the files named by args, returning a list of cleaned, sorted, deduplicated file
// names.
//
// An argument can name a directory, in which case files in the directory and its subdirectories are
// enumerated with the given glob.  The glob should not have a directory component.
//
// An argument of the form @filename names an indirection file, a text file wherein non-blank lines
// are treated in the same way as the components of args.  Relative paths in the indirection file
// are interpreted relative to the location of the indirection file itself.  Any files found are
// returned as full file names relative to the current working directory.
func FindFiles(args []string, glob string) ([]string, error) {
	filenames, err := findFilesRaw(args, glob, 0)
	if err != nil {
		return nil, err
	}
	for i := range filenames {
		filenames[i] = path.Clean(filenames[i])
	}
	slices.Sort(filenames)
	return slices.Compact(filenames), nil
}

func findFilesRaw(args []string, glob string, depth int) ([]string, error) {
	filenames := make([]string, 0, len(args))
	for _, f := range args {
		files, err := findNamed(f, glob, depth)
		if err != nil {
			return nil, err
		}
		filenames = append(filenames, files...)
	}
	return filenames, nil
}

func findNamed(f, glob string, depth int) ([]string, error) {
	if f == "" {
		return nil, fmt.Errorf("Empty file name")
	}
	if f[0] == '@' {
		indirname := f[1:]
		if indirname == "" {
			return nil, fmt.Errorf("Empty indirection name '@'")
		}
		if depth == maxDepth {
			return nil, fmt.Errorf("Indirection files nested too deeply")
		}
		indir, err := os.Open(indirname)
		if err != nil {
			return nil, err
		}
		defer indir.Close()
		dirname := path.Dir(indirname)
		names := make([]string, 0)
		scanner := bufio.NewScanner(indir)
		for scanner.Scan() {
			t := strings.TrimSpace(scanner.Text())
			if t == "" {
				continue
			}
			var fn string
			if t != "" && t[0] == '@' {
				fn = "@" + path.Join(dirname, t[1:])
			} else {
				fn = path.Join(dirname, t)
			}
			names = append(names, fn)
		}
		return findFilesRaw(names, glob, depth+1)
	}

	info, err := os.Stat(f)
	if err == nil {
		if info.Mode()&fs.ModeType == 0 {
			return []string{f}, nil
		}
		if info.Mode()&fs.ModeDir != 0 {
			return GlobDir(f, glob)
		}
	}
	return nil, fmt.Errorf("The item %s is neither file nor directory", f)
}

// Walk the directory tree at dir (ignoring symlinks) and match all plain files against glob, which
// should have no pathname component.  Any returned paths are prefixed by dir.  File names are
// supposedly cleaned and sorted (and naturally deduplicated).
func GlobDir(dir, glob string) ([]string, error) {
	files := make([]string, 0)
	err := fs.WalkDir(os.DirFS(dir), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Mode()&fs.ModeType == 0 {
			matched, err := path.Match(glob, d.Name())
			if err != nil {
				panic("Bad glob: " + glob)
			}
			if matched {
				files = append(files, path.Join(dir, p))
			}
		}
		return nil
	})
	return files, err
}
