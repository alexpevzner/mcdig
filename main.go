// MCDIG - DIG for MDNS (Multicast DNS lookup utility)
//
// Copyright (C) 2023 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// The main function

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// Program options are global, but located and initialized here
var (
	// optDomain specifies queried domain name
	OptDomain = ""

	// optQType specifies query type
	OptQType = dns.TypeA

	// optQClass specifies query class
	OptQClass uint16 = dns.ClassINET

	// optIface specifies query interface
	OptIface = ""

	// opt4/opt6 specifies IPv4/IPv6 transport. If none is
	// set, the default is used
	Opt4 = false
	Opt6 = false

	// OptTxPeriod specifies MDNS query retransmit interval
	OptTxPeriod = 250 * time.Millisecond

	// OptTxCount specifies how many MDNS queries will be
	// sent before completion
	OptTxCount = 10

	// OptQueryTime specifies the whole query wait time
	OptQueryTime = 2500 * time.Millisecond

	// OptDebug enables debugging
	OptDebug = false

	// OptVerbose enables verbose debugging
	// It implies OptDebug
	OptVerbose = false
)

// usage prints detailed usage and exits
func usage() {
	const help = "" +
		"Usage:\n" +
		"    mcdig [@interface] [options] domain [q-type] [q-class]\n" +
		"\n" +
		"Options may be intermixed with other parameters.\n" +
		"Use -- to terminate options list.\n" +
		"\n" +
		"The @interface specifies network interface (by name)\n" +
		"If missed, all active interfaces are used\n" +
		"\n" +
		"Options are:\n" +
		"    -4         use IPv4 (the default, may be combined with -6)\n" +
		"    -6         use IPv6 (may be combined with -4)\n" +
		"    -d         enable debugging\n" +
		"    -v         enable verbose debugging\n" +
		"    -p period  MDNS query period, milliseconds (default is %d)\n" +
		"    -c count   MDNS query count, before exit (default is %d)\n" +
		"    -h         print help screen and exit\n" +
		""

	fmt.Printf(help, OptTxPeriod/time.Millisecond, OptTxCount)
	os.Exit(0)
}

// usageError prints usage error and exits
func usageError(format string, args ...interface{}) {
	fmt.Printf(format+"\n"+"Try mcdig -h for more information\n", args...)
	os.Exit(1)
}

// optParse parses command-line options.
// This function doesn't return in a case of errors
func optParse() {
	// If no arguments are given, just print usage screen and exit
	if len(os.Args) == 1 {
		usage()
	}

	// Split command line into position arguments and options
	type option struct{ Name, Val string }

	args := []string{}
	opts := []option{}
	endOfOptions := false

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		switch {
		case endOfOptions:
			args = append(args, arg)

		case arg == "--":
			endOfOptions = true

		case arg == "-h":
			usage()

		case arg == "-p" || arg == "-c":
			if i+1 == len(os.Args) {
				usageError("option %s requires argument", arg)
			}
			opts = append(opts,
				option{Name: arg, Val: os.Args[i+1]})
			i++

		case strings.HasPrefix(arg, "-") || strings.HasPrefix(arg, "@"):
			opts = append(opts, option{Name: arg})

		default:
			args = append(args, arg)
		}
	}

	// Handle positional arguments
	switch len(args) {
	default:
		usageError("invalid argument: %q", args[3])

	case 3:
		if v, ok := dns.StringToClass[strings.ToUpper(args[2])]; ok {
			OptQClass = v
		} else {
			usageError("invalid class: %q", args[2])
		}
		fallthrough

	case 2:
		if v, ok := dns.StringToType[strings.ToUpper(args[1])]; ok {
			OptQType = v
		} else {
			usageError("invalid type: %q", args[1])
		}
		fallthrough

	case 1:
		OptDomain = args[0]

	case 0:
		usageError("missed domain")
	}

	// Handle options
	for _, opt := range opts {
		switch {
		case opt.Name == "-4":
			Opt4 = true

		case opt.Name == "-6":
			Opt6 = true

		case opt.Name == "-d":
			OptDebug = true

		case opt.Name == "-v":
			OptVerbose = true

		case opt.Name == "-p" || opt.Name == "-c":
			val, err := strconv.ParseUint(opt.Val, 0, 31)
			if err != nil {
				usageError("invalid argument: %s %s",
					opt.Name, opt.Val)
			}

			switch opt.Name {
			case "-p":
				OptTxPeriod = time.Duration(val) *
					time.Millisecond
			case "-c":
				OptTxCount = int(val)

			default:
				panic("internal error")
			}

		case strings.HasPrefix(opt.Name, "@"):
			if OptIface == "" {
				OptIface = opt.Name[1:]
			} else {
				usageError("Duplicated @interface")
			}

		default:
			usageError("invalid option: %q", opt)
		}
	}

	// Fixup options
	if !Opt4 && !Opt6 {
		Opt4 = true // The default if none set
	}
}

// The main function
func main() {
	optParse()
	q := QueryRun()
	ResponseGetAndPrint(os.Stdout, q)
}
