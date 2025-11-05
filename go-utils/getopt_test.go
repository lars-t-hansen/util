package utils

import (
	"fmt"
	"os"
)

func pushInputFile(_ bool, n string) error {
	return nil
}

var (
	namesFromStdin     bool
	printHelp          bool
	printVersion       bool
	verbose            bool
	outname            string
	systemEtagsCommand string
)

func example() {
	opts := []Option[bool]{
		Option[bool]{
			Short:   'h',
			Long:    "help",
			Help:    "Print help",
			Handler: SetFlag[bool](&printHelp),
		},
		Option[bool]{
			Short:   'o',
			Help:    fmt.Sprintf(`Filename of output file, "-" for stdout, default "%s"`, outname),
			Value:   true,
			Handler: SetString[bool](&outname),
		},
		Option[bool]{
			Short:   'v',
			Help:    "Enable verbose output (for debugging)",
			Handler: SetFlag[bool](&verbose),
		},
		Option[bool]{
			Short:   'V',
			Long:    "version",
			Help:    "Print version information",
			Handler: SetFlag[bool](&printVersion),
		},
		Option[bool]{
			Long:  "etags",
			Value: true,
			Help: fmt.Sprintf(
				`Path of the native etags program, "" to disable this functionality, default "%s"`,
				systemEtagsCommand,
			),
			Handler: SetString[bool](&systemEtagsCommand),
		},
		Option[bool]{
			Short:   '-',
			Handler: SetFlag[bool](&namesFromStdin),
		},
		Option[bool]{
			// Implicitly repeatable
			Handler: pushInputFile,
		},
	}
	rest, err := GetOpts(opts, true, os.Args[1:])
	if err != nil {
		panic(err)
		os.Exit(2)
	}
	_ = rest
	if printHelp {
		// ...;
		os.Exit(0)
	}
	if printVersion {
		//...;
		os.Exit(0)
	}
	for _, r := range rest {
		pushInputFile(true, r)
	}
}
