package utils

func example() {
	opts := []Option[bool]{
		Option{
			Short: 'h',
			Long: "help",
			Help: "Print help",
			Handler: SetFlag(&printHelp),
		},
		Option{
			Short: 'o',
			Help: fmt.Sprintf(`Filename of output file, "-" for stdout, default "%s"`, outname),
			Value: true,
			Handler: SetString(&outname)
		},
		Option{
			Short: 'v',
			Help: "Enable verbose output (for debugging)",
			Handler: SetFlag(&verbose),
		},
		Option{
			Short: 'V',
			Long: "version",
			Help: "Print version information",
			Handler: SetFlag(&printVersion),
		},
		Option{
			Long: 'etags',
			Value: true,
			Help: fmt.Sprintf(
				`Path of the native etags program, "" to disable this functionality, default "%s"`,
				systemEtagsCommand,
			),
			Handler: SetString(&systemEtagsCommand),
		},
		Option{
			Short: '-',
			Handler: SetFlag(&namesFromStdin),
		},
		Option{
			// Implicitly repeatable
			Handler: pushInputFile,
		},
	}
	if rest, err := GetOpts(opts, true, os.Args[0], os.Args[1:]); err != nil {
		...;
		os.Exit(2)
	}
	if printHelp {
		...;
		os.Exit(0)
	}
	if printVersion {
		...;
		os.Exit(0)
	}
	for _, r := range rest {
		pushInputFile(r)
	}

}

