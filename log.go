// MDIG - DIG for MDNS (Multicast DNS lookup utility)
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

// LogDebug writes a debug message
func LogDebug(format string, args ...interface{}) {
	if OptDebug {
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