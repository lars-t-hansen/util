// Print a random selection of lines from a file, in the original order.  Reads from stdin, writes
// to stdout.
//
// This is not intended to be clever.  Extremely Huge (tm) files may defeat it.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
)

var (
	atLeast = flag.Uint("atleast", 0, "Print at least this many lines (up to file length)")
	pct = flag.Float64("pct", 0, "Print this percentage of lines")
)

func main() {
	flag.Parse()
	if *atLeast == 0 && *pct == 0 {
		fmt.Fprintln(os.Stderr, "At least one of -atleast and -pct is required.\n")
		flag.Usage()
		os.Exit(2)
	}
	if *pct < 0 || *pct > 100 {
		fmt.Fprintln(os.Stderr, "Percentage out of range")
		os.Exit(2)
	}

	scanner := bufio.NewScanner(os.Stdin)
	ls := make([]string, 0, 1000)
	for scanner.Scan() {
		ls = append(ls, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Scanner failed", err)
		os.Exit(1)
	}
	*atLeast = min(*atLeast, uint(len(ls)))

	// Number of selections
	toPick := max(int(*atLeast), min(int(float64(len(ls))*(*pct)/100), len(ls)))
	if toPick == 0 {
		return
	}

	// Bag of candidates, indices into `ls`
	cand := make([]int, len(ls))
	for i := 0 ; i < len(cand); i++ {
		cand[i] = i
	}

	// Permute the candidates
	for i := 0 ; i < len(cand) ; i++ {
		r := rand.Intn(len(cand))
		cand[i], cand[r] = cand[r], cand[i]
	}

	// Print the prefix of the permutation in the original order
	sel := cand[:toPick]
	sort.Sort(sort.IntSlice(sel))
	for _, k := range sel {
		fmt.Println(ls[k])
	}
}
