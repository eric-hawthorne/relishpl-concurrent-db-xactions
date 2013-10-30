// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// This package and file contains definitions needed by several relish packages 
// in the compiler and interpreter.

package defs


// Lookup table of the origin and artifact of each relish standard library package pathname.
// In theory, the standard library can thus be split into several different artifacts such
// as a core standard library and various extension (but still considered standard) libraries.
// Also, the special value "relish" is returned for standard library packages which 
// consist of only a set of inbuilt methods, so they need no actual relish artifact loaded.
//
var StandardLibPackageArtifact map[string]string = map[string]string {
	"strings" : "relish",
	"datetime" : "relish",
	"http" : "relish",
	"csv" : "relish",	
	"json" : "relish",	
	"io" : "shared.relish.pl2012/relish_lib",		
	"files" : "shared.relish.pl2012/relish_lib",	
	"http_srv" : "shared.relish.pl2012/relish_lib",		
	"crypto" : "shared.relish.pl2012/relish_lib",			
}


// Lookup table of relish language builtin type names

var BuiltinTypeName map[string]bool = map[string]bool {
"RelishPrimitive" : true,
"Numeric" : true,
"Integer" : true,
"Int" : true,
"IntOrString" : true,
"Int32" : true,
"Int16" : true,
"Int8" : true,
"Uint" : true,
"Uint32" : true,
"Uint16" : true,
"Byte" : true,
"Bit" : true,
"Bool" : true,
"Text" : true,
"CodePoint" : true,  // TODO - Change to Rune
"Real" : true,
"Float" : true,
"Float32" : true,
"ComplexNumber" : true,
"Complex" : true,
"Complex32" : true,
"String" : true,
"Time" : true,
"Proxy" : true,
"Callable" : true,
"MultiMethod" : true,
"Method" : true,
"Any" : true,
"NonPrimitive" : true,
"Struct" : true,
"Collection" : true,
"List" : true,
"Set" : true,
"Map" : true,
"InChannel" : true,
"OutChannel" : true,
"Channel" : true,
"Mutex" : true,
"RWMutex" : true,
"Bytes": true,
"Bits": true,
"CodePoints": true,  // TODO - Change to Runes
}



// Implementation detail. Separator used in a map key composed of multiple parts we want to be able to separate later.
const KEY_PART_SEPARATOR = "_|_"  