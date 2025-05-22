package main

import (
	"strings"
	"testing"
)

func TestSimple(t *testing.T) {
	patterns := `
S c 1
S b 2
S z c
C ax \d+ y
`
	input := `
abracadabra hocuspozus ax10axax20ax10axas here
`
	expect := `
a2ra1ada2ra ho1uspocus axy10000axaxy10001axy10000axas here
`
	var out strings.Builder
	run(strings.NewReader(patterns), strings.NewReader(input), &out)
	output := out.String()
	if output != expect {
		t.Fatal(expect, output)
	}
}
