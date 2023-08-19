// MDIG - DIG for MDNS (Multicast DNS lookup utility)
//
// Copyright (C) 2023 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// The main function

package main

import (
	"fmt"
	"os"
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

	// OptRetransmitInterval specifies MDNS query retransmit interval
	OptRetransmitInterval = 250 * time.Millisecond

	// OptQueryTime specifies the whole query wait time
	OptQueryTime = 2500 * time.Millisecond

	// OptDebug enables debugging
	OptDebug = true

	// OptVerbose enables verbose debugging
	// It implies OptDebug
	OptVerbose = true
)

// usage prints detailed usage and exits
func usage() {
	const help = "" +
		"Usage:\n" +
		"    mdig [@interface] [options] domain [q-type] [q-class]\n" +
		"\n" +
		"Options may be intermixed with other parameters.\n" +
		"Use -- to terminate options list.\n" +
		"\n" +
		"The @interface specifies network interface (by name)\n" +
		"If missed, all active interfaces are used\n" +
		"\n" +
		"Options are:\n" +
		"    -4 use IPv4 (the default, may be combined with -6)\n" +
		"    -6 use IPv6 (may be combined with -4)\n" +
		"    -h print help screen and exit\n" +
		""

	fmt.Print(help)
	os.Exit(0)
}

// usageError prints usage error and exits
func usageError(format string, args ...interface{}) {
	fmt.Printf(format+"\n"+"Try mdig -h for more information\n", args...)
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
	args := []string{}
	opts := []string{}
	endOfOptions := false

	for _, arg := range os.Args[1:] {
		switch {
		case endOfOptions:
			args = append(args, arg)

		case arg == "--":
			endOfOptions = true

		case arg == "-h":
			usage()

		case strings.HasPrefix(arg, "-") || strings.HasPrefix(arg, "@"):
			opts = append(opts, arg)

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
		case opt == "-4":
			Opt4 = true

		case opt == "-6":
			Opt6 = true

		case strings.HasPrefix(opt, "@"):
			if OptIface == "" {
				OptIface = opt[1:]
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
	QueryRun()
	ResponseGetAndPrint(os.Stdout)
}
