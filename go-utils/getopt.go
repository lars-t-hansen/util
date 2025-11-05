// SPDX-License-Identifier: MIT
// From https://github.com/lars-t-hansen/util/go-utils

// GNU-ish getopt-style option parser.

package utils

type Option[Cx any] struct {
	Short rune
	Long string
	Help string
	Value bool
	Repeatable bool
	Handler func(cx Cx, string) error
}

// GetOpts is a simple table-driven argument parser that is superficially compatible with GNU
// getopt-style parsing.  The parser takes a table of options and handlers, a context argument to
// pass to the handlers, a program name and a list of arguments, and parses the arguments, invoking
// handlers for the various option names for each argument.  Argument parsing stops at '--', any
// remaining arguments are returned as a slice from the call.
//
// An option can have a long and/or a short name.  If it has neither it is the default handler.
//
// An option can have the short name '-' (and no long name and take no value).  This is recognized
// as a special case.
//
// Long names with values can use the syntax '--opt=value' or '--opt value'.
//
// Short names with values can use the syntax '-n value' or '-nvalue', in the latter case provided
// only that the first letter of the value is not a valid short option letter.  Thus '-k 2' can be
// written '-k2'.
//
// Short names can be run together, eg '-nr' is equivalent to '-n -r'.
//
// Short names be run together with a value for at most one of the options, as in '-nkr2' for
// 'sort(1)', equivalent to -n -k 2 -r; also '-nkr 2' is accepted.  Again the boundary between
// option letters and value in the former case is where a letter is not a valid option letter, so
// '-nk2r' won't work.
//
// Failure to validate the option table results in a panic.  Any error returned from GetOpts is an
// error returned from the parser due to an unknown option or from one of the handlers as a
// validation error, the latter wrapped in a parser error explaining the context.
//
// Other than the default option, an option is not repeatable unless its Repeatable attribute is
// set.
//
// The help text is documentation that is not used by GetOpts, but locating it in the options table
// may be useful for other parts of the program.
func GetOpts[Cx any](options []Option[Cx], cx Cx, progname string, args []string) (rest []string, parseErr error) {
	// These maps map the option name or character without the leading '-' or '--'.
	short := make(map[rune]*Option)
	long := make(map[string]*Option)
	handled := make(map[any]bool)
	var defaultHandler func(Cx, string) error

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
		if o.Short != 0 {
			if short[o.Short] != nil {
				panic("Multiple definitions for short option %c", o.Short)
			}
			if o.Short == '-' {
				if o.Long != "" {
					panic("Long name defined for lone dash")
				}
				if o.Value {
					panic("Lone dash cannot take a value")
				}
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
		defaultHandler = func(_ Cx, _ string) error {
			return fmt.Errorf("Additional arguments not allowed")
		}
	}

	// Parse the options.
	argIx := 0
	for argIx < len(args) {
		arg := args[argIx]
		argIx++
		if len(arg) > 0 && arg[0] == '-' {
			if len(arg) > 1 && arg[1] == '-' {
				if len(arg) == 2 {
					rest = args[argIx:]
					return
				}
				arg = arg[2:]
				optname, value, matched := strings.Cut(arg, "=")
				opt := long[optname]
				if opt == nil {
					parseErr = fmt.Errorf("Unknown option \"%s\"", optname)
					return
				}
				if !opt.Repeatable && handled[optname] {
					parseErr = fmt.Errorf("Repeated but unrepeatable option \"%s\"", optname)
					return
				}
				handled[optname] = true
				if !opt.Value && matched {
					parseErr = fmt.Errorf("Option \"%s\" does not take a value", optname)
					return
				}
				if opt.Value && !matched {
					if argIx == len(args) {
						parseErr = fmt.Errorf("Missing value for option \"%s\"", optname)
						return
					}
					value = args[argIx]
					argIx++
				}
				parseErr = opt.Handler(value)
				if parseErr != nil {
					parseErr = fmt.Errorf("Rejected option \"%s\": %w", r)
					return
				}
			} else {
				i := 0
				if len(arg) > 1 {
					i++
				}
				var needValue func(Cx, string) error
				loop {
					if i == len(arg) {
						break
					}
					optname := arg[i]
					opt := short[optname]
					i++
					if opt == nil {
						if i < 2 {
							parseErr = fmt.Errorf("Illegal option")
							return
						}
						break
					}
					if !opt.Repeatable && handled[optname] {
						parseErr = fmt.Errorf("Repeated but unrepeatable option \"%c\"", optname)
						return
					}
					handled[optname] = true
					if opt.Value {
						if needValue != nil {
							parseErr = fmt.Errorf("Multiple short options compete for a value")
							return
						}
						needValue = opt.Handler
					} else {
						parseErr = opt.Handler(cx, "")
						if parseErr != nil {
							return
						}
					}
				}
				if needValue != nil {
					var value string
					if i < len(arg) {
						value = arg[i:]
					} else {
						if argIx == len(args) {
							parseErr = fmt.Errorf("Missing value for option \"%s\"", optname)
							return
						}
						value = args[argIx]
						argIx++
					}
					parseErr = needValue(cx, value)
					if parseErr != nil {
						return
					}
				}
			}
		} else {
			parseErr = defaultHandler(a)
			if parseErr != nil {
				return
			}
		}
	}
	return
}

func PrintOptions[Cx any](output io.Writer, options []Option[Cx]) {
	// Print a line with short and long options
	// Print help indented on next line
	// Don't be fancy
	for _, o := range options {
	}
}

// We can have more of these...

// A simple handler that will set a flag to true and always succeed
func SetFlag[Cx any](flagp *bool) func(Cx, string) error {
	return func(_ Cx, _ string) error {
		*flagp = true
		return nil
	}
}

// A simple handler that will set a string to a given value and always succeed
func SetString[Cx any](stringp *string) func(Cx, string) error {
	return func(_ Cx, s string) error {
		*stringp = s
		return nil
	}
}
