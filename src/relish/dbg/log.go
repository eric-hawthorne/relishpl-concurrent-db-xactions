// Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// This file implements configurable, multi-concern, multi-level logging.

package dbg

import (
	"fmt"
)	
	
var indent uint  // Current call nesting level
	
// These single-bit-set flags should be used (possibly |'d together explicitly) as the 
// first arg to Log,Logln,Logging,Trace functions.
// Each bit-flag represents a program aspect (or aspect x verbosity-level) to log.
//	
const (
	STACK_      = 1 << iota   // == 1 (iota has been reset)
	AST_        = 1 << iota   // == 2
	PARSE_	    = 1 << iota   // == 4
	INTERP_     = 1 << iota   // Log INTERPreter aspect of program when logging at minimum (or above) verbosity level
	INTERP2_    = 1 << iota	  // Log INTERPreter aspect of program only when logging at medium (or above) verbosity level 
	INTERP3_    = 1 << iota	  // Log INTERPreter aspect of program only when logging at max verbosity level	
	INTERP_TR   = 1 << iota	
	INTERP_TR2  = 1 << iota	
	INTERP_TR3  = 1 << iota	
	PERSIST_    = 1 << iota	
	PERSIST2_   = 1 << iota	
	PERSIST_TR  = 1 << iota	
	PERSIST_TR2 = 1 << iota
	GENERATE_   = 1 << iota			
	ALWAYS_     = 1 << iota		
	WEB_        = 1 << iota	
	WEB2_       = 1 << iota	
	ANY_        = 0xFFFFFFFFFFFFFFFF	
)

// These multi-bit flags should be used to set the DEBUG_FLAGS. They are shortcuts
// which represent which other bits should also be set when a particular bit is set. 
// i.e. They implement level-hierarchy such as INTERP_,INTERP2_,INTERP3_ in such a way that
// INTERP3_ implies INTERP2_ and INTERP1_   etc.
//
const (
	STACK__     = STACK_
	AST__       = AST_
	PARSE__	    = PARSE_
	INTERP__    = INTERP_                   // Log INTERPreter aspect of program at minimum verbosity level
	INTERP2__   = INTERP2_ | INTERP__	    // Log INTERPreter aspect of program at medium verbosity level
	INTERP3__   = INTERP3_ | INTERP2__    	// Log INTERPreter aspect of program at max verbosity level	
	INTERP__TR  = INTERP_TR	
	INTERP__TR2 = INTERP_TR2 | INTERP__TR
	INTERP__TR3 = INTERP_TR3 | INTERP__TR2	
	PERSIST__ = PERSIST_
	PERSIST2__ = PERSIST2_ | PERSIST__
	PERSIST__TR = PERSIST_TR
	PERSIST__TR2 = PERSIST_TR2 | PERSIST__TR	
	WEB__       = WEB_
	WEB2__       = WEB2_ | WEB__	
	ALL_        = ANY_
)
	
	
//const DEBUG_FLAGS = INTERP__TR3 | INTERP3__ | STACK__ | AST__ | PARSE__
//const DEBUG_FLAGS = INTERP__TR | INTERP__ | STACK__ | AST__ | PARSE__	
//const DEBUG_FLAGS = INTERP__TR | INTERP__ | AST__ | PARSE__	

// Was using this one for a long time.
//const SOME_DEBUG_FLAGS =  AST__ | PARSE__ | PERSIST2__ | PERSIST__TR2| INTERP__TR2 | INTERP2__	
// const SOME_DEBUG_FLAGS =   PERSIST2__ | PERSIST__TR2 | INTERP__TR2 | INTERP2__ | WEB__ | ALWAYS_
//const SOME_DEBUG_FLAGS =   PERSIST__ | PERSIST__TR | INTERP__TR | INTERP__ | WEB__ | ALWAYS_
const SOME_DEBUG_FLAGS =  WEB__ | ALWAYS_

// const FULL_DEBUG_FLAGS = SOME_DEBUG_FLAGS | GENERATE_

const FULL_DEBUG_FLAGS = SOME_DEBUG_FLAGS | GENERATE_ | PARSE__ | AST__
	
const NO_DEBUG_FLAGS = ALWAYS_

var DEBUG_FLAGS uint64 = NO_DEBUG_FLAGS

/*
   Set the overall debug level to 0 (none), 1 (some), or 2 (full). Debug level defaults to 0 if this method is not called.
*/
func InitLogging(masterLevel int32) {
	switch masterLevel {
	case 0: // No debugging
	   DEBUG_FLAGS = NO_DEBUG_FLAGS
	case 1: // Some
	   DEBUG_FLAGS = SOME_DEBUG_FLAGS	
    case 2: // Full 
	   DEBUG_FLAGS = FULL_DEBUG_FLAGS
	}
}

	
func Log(flags uint64,s string,args ...interface{}) {
   if flags & DEBUG_FLAGS != 0 {
      printDots()		
      fmt.Printf(s,args...)
   }
}

func Logln(flags uint64,s ...interface{}) {
   if flags & DEBUG_FLAGS != 0 {
      printDots()		
      fmt.Println(s...)
   }
}

/*
Whether we are currently logging aspects/levels identified by the specified flags.
Usage pattern: if Logging(INTERP2_) { t.Dump() }
*/
func Logging(flags uint64) bool {
   return flags & DEBUG_FLAGS != 0 	
}



func Trace(flags uint64, msg string, args ...interface{}) string {
	if flags & DEBUG_FLAGS != 0 {
		
	   args2 := []interface{}{msg,"("}
	   args2 = append(args2,args...)
	   printTrace(args2...)		
	   //printTrace(msg, "(")
 	   indent++
	   return msg
    }
    return ""
}

// Usage pattern: defer Un(Trace("..."))
func Un(msg string) {
	if msg != "" {
	   indent--
	   printTrace(")",msg)
    }
}

func printTrace(a ...interface{}) {
	printDots()
	fmt.Println(a...)
}

func printDots() {
	const dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . " +
		". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . "
	const n = uint(len(dots))
	i := 2 * indent
	for ; i > n; i -= n {
		fmt.Print(dots)
	}
	fmt.Print(dots[0:i])
}