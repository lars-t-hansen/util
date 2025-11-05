// SPDX-License-Identifier: MIT
// From https://github.com/lars-t-hansen/util/go-utils

// Work in progress!

package utils

type Option struct {
	Short rune
	Long string
	Help string
	Value bool
	Repeatable bool
	Handler func(string) error
}

// GetOpts is a simple table-driven argument parser that is superficially compatible with GNU
// getopt-style parsing.  The parser takes a table of options and handlers, a program name and a
// list of arguments, and parses the arguments, invoking handlers for the various option names for
// each argument.  Argument parsing stops at '--', any remaining arguments are returned as a slice
// from the call.
//
// An option can have a long and/or a short name.  If it has neither it is the default handler.
//
// An option can have the short name '-' (and no long name and take no value).  This is recognized
// as a special case.
//
// Long names with values can use the syntax '--opt=value' or '--opt value'.
//
// Short names can be run together and can also be run together with a value for at most one of the
// options, as in '-nkr2' for 'sort(1)', equivalent to -n -k 2 -r; also '-nkr 2' is accepted.  The
// boundary between option letters and value in the former case is where a letter is not a valid
// option letter, so '-nk2r' won't work.
//
// Failure to validate the option table results in a panic.  Any error returned from GetOpts is an
// error returned from the parser due to an unknown option or from one of the handlers as a
// validation error, the latter wrapped in a parser error explaining the context.
//
// Other than the default option, an option is not repeatably unless its Repeatable attribute is
// set.
//
// The help text is documentation that is not used by GetOpts, but locating it in the options table
// may be useful for other parts of the program.
func GetOpts(options []Option, progname string, args []string) (rest []string, parseErr error) {
	// These maps map the option name or character without the leading '-' or '--'.
	short := make(map[rune]*Option)
	long := make(map[string]*Option)
	handled := make(map[any]bool)
	var defaultHandler func(string) error
	var loneDash bool

	// Parse the table.
	for i := range options {
		o := &options[i]
		if o.Handler == nil {
			panic("Option without handler at loc %d: short=%c long=%s", i, o.Short, o.Long)
		}
		if o.Short == 0 && o.Long == "" {
			if defaultHandler != nil {
				panic("Multiple default handlers")
			}
			defaultHandler = o
		}
		if o.Short == '-' {
			if o.Long != "" {
				panic("Long name defined for lone dash")
			}
			if o.Value {
				panic("Lone dash cannot take a value")
			}
			loneDash = true
			continue
		}
		if o.Short != 0 {
			if short[o.Short] != nil {
				panic("Multiple definitions for short option %c", o.Short)
			}
			short[i.Short] = o
		}
		if o.Long != "" {
			if long[o.Long] != nil {
				panic("Multiple definitions for long option %s", o.Long)
			}
			long[o.Long] = o
		}
	}
	if defaultHandler == nil {
		// TODO: assign a handler that causes an error
	}

	// Parse the options.
	i := 0
	for i < len(args) {
		a := args[i]
		i++
		if len(a) > 0 && a[0] == '-' {
			if len(a) > 1 && a[1] == '-' {
				if len(a) == 2 {
					rest = args[i:]
					return
				}
				a = a[2:]
				optname, value, matched := strings.Cut(a, "=")
				opt := long[optname]
				if opt == nil {
					parseErr = fmt.Errorf("Unknown option \"%s\"", optname)
					return
				}
				if !opt.Value && matched {
					parseErr = fmt.Errorf("Option \"%s\" does not take a value", optname)
					return
				}
				if opt.Value && !matched {
					// This will allow eg --arg -value which could also be considered an error
					if i == len(args) {
						parseErr = fmt.Errorf("Missing value for option \"%s\"", optname)
						return
					}
					value = args[i]
					i++
				}
				if !opt.Repeatable && handled[optname] {
					parseErr = fmt.Errorf("Repeated but unrepeatable option \"%s\"", optname)
					return
				}
				handled[optname] = true
				if opt.Value {
					parseErr = opt.Handler(value)
				} else {
					parseErr = opt.Handler("")
				}
				if parseErr != nil {
					parseErr = fmt.Errorf("Rejected option \"%s\": %w", r)
					return
				}
			} else {
				if len(a) == 1 {
					if loneDash {
						// special case
					} else {
						// error
					}
				} else {
					// Now must parse the option string.  This is a mess.  In principle it is a
					// string of option characters followed immediately by a value.  At most one of
					// the option chars can take a value.  The parsing of option characters stops
					// when a character encountered is not an option char.  If there is not a value
					// in the option string then one can follow in the next arg.
				}
			}
		} else {
			if !defaultHandler(a) {
				parsErr = fmt.Errorf("Rejected non-argument value \"%s\"", a)
				return
			}
		}
	}
	return
}

// A simple handler that will set a flag to true and always succeed
func SetFlag(flagp *bool) func(string) error {
	return func(_ string) error {
		*flagp = true
		return nil
	}
}

// A simple handler that will set a string to a given value and always succeed
func SetString(stringp *string) func(string) error {
	return func(s string) error {
		*stringp = s
		return nil
	}
}
