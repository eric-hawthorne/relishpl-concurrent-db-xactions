// Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// This package and file contains definitions needed by several relish packages 
// in the compiler and interpreter.

package defs

// Lookup table of relish language builtin type names

var BuiltinTypeName map[string]bool = map[string]bool {
"RelishPrimitive" : true,
"Numeric" : true,
"Integer" : true,
"Int" : true,
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
"CodePoint" : true,
"Real" : true,
"Float" : true,
"Float32" : true,
"ComplexNumber" : true,
"Complex" : true,
"Complex32" : true,
"String" : true,
"Proxy" : true,
"Callable" : true,
"MultiMethod" : true,
"Method" : true,
"Any" : true,
"NonPrimitive" : true,
"Struct" : true,
"Collection" : true,
"InChannel" : true,
"OutChannel" : true,
"Channel" : true,
}