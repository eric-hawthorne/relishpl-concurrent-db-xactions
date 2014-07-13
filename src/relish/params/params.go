// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// Adjustable parameters for relish compilation and execution environment.

package params

// Garbage Collection Check Interval (seconds). Checks to see if allocated unfreed memory has doubled
// or increased by 10 MB since last check time. 
// If so, does a relish garbage collection and Go garbage collection.
//
var GcIntervalSeconds = 20

// If the last relish GC was longer than this long ago, relish GC is run for sure.
var GcForceIntervalSeconds = 600  // 10 minutes


// Maximum connection pool size
var DbMaxConnections = 20