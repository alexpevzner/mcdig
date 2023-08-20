// MCDIG - DIG for MDNS (Multicast DNS lookup utility)
//
// Copyright (C) 2023 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Logging

package main

import (
	"fmt"
	"os"
)

// LogVerbose writes a verbose debug message
func LogVerbose(format string, args ...interface{}) {
	if OptVerbose {
		fmt.Printf(format+"\n", args...)
	}
}

// LogDebug writes a debug message
func LogDebug(format string, args ...interface{}) {
	if OptDebug || OptVerbose {
		fmt.Printf(format+"\n", args...)
	}
}

// LogError writes an error message
func LogError(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// LogFatal writes an error message and terminates the program
func LogFatal(format string, args ...interface{}) {
	LogError(format, args...)
	os.Exit(1)
}
