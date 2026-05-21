// Rename finds and rewrites patterns in files.
//
// Usage:
//
//	rename regexp replacement [file ... ]
//
// where the regexp can have captures, and replacement can reference those captures with \n syntax
// starting with \1.
//
// Example to rename identifiers of the form kX... to KX... where X is some upper case letter:
//
//	rename '([^a-zA-Z0-9_])k([A-Z])' '\1K\2' myfile.xyz
//
// Files are rewritten in place (temp file generated, then renamed on success), one after the other.
// If there are no files, then input is read from stdin and written to stdout.
//
// # BUGS
//
// There's no way to express a backslash in the replacement string.
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: rename regex replacement [file ...]")
		os.Exit(2)
	}
	re, err := regexp.Compile(os.Args[1])
	if err != nil {
		log.Fatalf("ERROR: Regexp: %v", err)
	}
	replacement, err := compile(os.Args[2])
	if err != nil {
		log.Fatalf("ERROR: Replacement: %v", err)
	}
	if len(os.Args) > 3 {
		for _, fn := range os.Args[3:] {
			f, err := os.Open(fn)
			if err != nil {
				log.Fatal(err)
			}
			tmp, err := os.CreateTemp(".", "rename")
			if err != nil {
				log.Fatal(err)
			}
			tmpname := tmp.Name()
			defer os.Remove(tmpname)
			err = process(f, tmp, re, replacement)
			f.Close()
			tmp.Close()
			if err != nil {
				log.Fatal(err)
			}
			err = os.Rename(tmpname, fn)
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {
		err := process(os.Stdin, os.Stdout, re, replacement)
		if err != nil {
			log.Fatal(err)
		}
	}
}

type repl = any // string | int

func compile(replacement string) (result []repl, err error) {
	i := 0
	for i < len(replacement) {
		start := i
		for i < len(replacement) && replacement[i] != '\\' {
			i++
		}
		if i > start {
			result = append(result, replacement[start:i])
		}
		if i == len(replacement) {
			break
		}
		if replacement[i] != '\\' {
			panic("Unexpected")
		}
		i++
		ns := ""
		for i < len(replacement) && replacement[i] >= '0' && replacement[i] <= '9' {
			ns += string(replacement[i])
			i++
		}
		if ns == "" {
			err = fmt.Errorf("Empty replacement sequence")
			return
		}
		var nn int64
		nn, err = strconv.ParseInt(ns, 10, 32)
		if err != nil {
			return
		}
		if nn == 0 {
			err = fmt.Errorf("Invalid replacement index")
			return
		}
		result = append(result, int(nn))
	}
	return
}

func process(in io.Reader, out io.Writer, re *regexp.Regexp, replacement []repl) error {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		l := scanner.Text()
		s := ""
		for l != "" {
			ms := re.FindStringSubmatchIndex(l)
			if ms == nil {
				s += l
				break
			}
			if ms[0] == ms[1] {
				return fmt.Errorf("Empty match")
			}
			if ms[0] > 0 {
				s += l[:ms[0]]
			}
			for _, r := range replacement {
				switch rr := r.(type) {
				case string:
					s += rr
				case int:
					if rr >= len(ms)/2 {
						return fmt.Errorf("Out of range submatch")
					}
					s += l[ms[rr*2]:ms[rr*2+1]]
				default:
					panic("Unexpected")
				}
			}
			l = l[ms[1]:]
		}
		_, err := fmt.Fprintln(out, s)
		if err != nil {
			return err
		}
	}
	return nil
}
