// Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// Utility function(s) for handling relish runtime errors.

package rterr

import (
   "os"
   "fmt"
)

/*
Prints the error message on os.Stdout, prefixed with the phrase: Runtime Error: 
then exits the relish process with status 1 (meaning abnormal exit as opposed to status 0 which would mean an exit with success.)
*/
func Stop(errorMessage string) {
	fmt.Fprintln(os.Stdout, "Runtime Error:", errorMessage)
	os.Exit(1)
}

/*
Prints the formatted error message on os.Stdout, prefixed with the phrase: Runtime Error: 
Formats the message just like Printf except also adds a carriage return after the formatted message. 
then exits the relish process with status 1 (meaning abnormal exit as opposed to status 0 which would mean an exit with success.)
*/
func Stopf(errorMessage string, args ...interface{}) {
	errorMessage = "Runtime Error: " + errorMessage 
	fmt.Fprintf(os.Stdout, errorMessage, args...)
	fmt.Fprintln(os.Stdout)
	os.Exit(1)
}