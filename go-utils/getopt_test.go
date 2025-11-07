package utils

import (
	"fmt"
	"reflect"
	"testing"
)

func TestOpts(t *testing.T) {
	var (
		namesFromStdin     bool
		printHelp          bool
		printVersion       bool
		verbose            bool
		outname            string
		systemEtagsCommand string
		inputFiles = make([]string, 0)
		oneArgs = make([]string, 0)
		pushInputFile = func(s string) error {
			inputFiles = append(inputFiles, s)
			return nil
		}
		opts = []Option{
			Option{
				Short:   'h',
				Long:    "help",
				Help:    "Print help",
				Handler: SetFlag(&printHelp),
			},
			Option{
				Short:   'o',
				Help:    fmt.Sprintf(`Filename of output file, "-" for stdout, default "%s"`, outname),
				Value:   true,
				Handler: SetString(&outname),
			},
			Option{
				Short:   'v',
				Help:    "Enable verbose output (for debugging)",
				Handler: SetFlag(&verbose),
			},
			Option{
				Short:   'V',
				Long:    "version",
				Help:    "Print version information",
				Handler: SetFlag(&printVersion),
			},
			Option{
				Long:  "etags",
				Value: true,
				Help: fmt.Sprintf(
					`Path of the native etags program, "" to disable this functionality, default "%s"`,
					systemEtagsCommand,
				),
				Handler: SetString(&systemEtagsCommand),
			},
			Option{
				Long: "onearg",
				Short: 'x',
				Value: true,
				Repeatable: true,
				Handler: func(s string) error {
					oneArgs = append(oneArgs, s)
					return nil
				},
			},
			Option{
				Short:   '-',
				Handler: SetFlag(&namesFromStdin),
			},
			Option{
				Repeatable: true,
				Handler: pushInputFile,
			},
		}
	)

	rest, err := GetOpts(opts, []string{
		"-",
		"-o",
		"-",
		"hello.c",
		"--help",
		"--etags=zappa",
		"-Vxzappa",
		"--onearg",
		"true",
		"--onearg",
		"false",
		"oneMoreFile.c",
		"-xv",
		"zupp",
		"--",
		"abra",
		"cadabra",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rest, []string{"abra","cadabra"}) {
		t.Fatal(rest)
	}
	if !printHelp {
		t.Fatal("printHelp")
	}
	if !verbose {
		t.Fatal("verbose")
	}
	if !printVersion {
		t.Fatal("printVersion")
	}
	if !namesFromStdin {
		t.Fatal("namesFromStdin")
	}
	if outname != "-" {
		t.Fatal("outname <" + outname + ">")
	}
	if !reflect.DeepEqual(inputFiles, []string{"hello.c", "oneMoreFile.c"}) {
		t.Fatal(inputFiles)
	}
	if !reflect.DeepEqual(oneArgs, []string{"zappa", "true", "false", "zupp"}) {
		t.Fatal(oneArgs)
	}
}
