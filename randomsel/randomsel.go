// Print a random selection of lines from a file, in the original order.  Reads from stdin, writes
// to stdout.
//
// The only parameter is a floating point value in the range [0,100] indicating the percentage of
// lines we want.
//
// This is not intended to be clever.  Extremely Huge (tm) files may defeat it.

package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s percentage", os.Args[0])
	}
	pct, err := strconv.ParseFloat(os.Args[1], 64)
	if err != nil || pct < 0 || pct > 100 {
		log.Fatalf("Usage: %s percentage", os.Args[0])
	}
	scanner := bufio.NewScanner(os.Stdin)
	ls := make([]string, 0, 1000)
	for scanner.Scan() {
		ls = append(ls, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal("Scanner failed", err)
	}

	// Number of selections
	toPick := min(int(float64(len(ls))*pct/100), len(ls))
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
