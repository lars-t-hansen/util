// On each line, replace runs of whitespace with a single comma except at the beginning or end
//
// Usage: go run awk2csv.go < input > output
package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	r, _ := regexp.Compile(`\s+`)
	for scanner.Scan() {
		fmt.Println(r.ReplaceAllString(strings.TrimSpace(scanner.Text()), ","))
	}
}
