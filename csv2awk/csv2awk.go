// Csv2awk converts CSV files to AWK-friendly space-separated-column files.  Empty
// input columns are given the synthetic value ".".  Spaces in values are translated
// to underscore.
//
// Usage: csv2awk
//
// Reads stdin, writes to stdout
package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	input := os.Stdin
	output := bufio.NewWriter(os.Stdout)
	defer output.Flush()

	r := csv.NewReader(input)
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	for _, r := range records {
		for i, f := range r {
			if i > 0 {
				fmt.Fprint(output, " ")
			}
			if f == "" {
				f = "."
			}
			fmt.Fprint(output, strings.ReplaceAll(f, " ", "_"))
		}
		fmt.Fprintln(output)
	}
}
