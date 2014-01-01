// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// Adjustable parameters for relish compilation and execution environment.

package params

// Garbage Collection Interval (seconds). Checks to see if allocated unfreed memory has doubled
// since last check time. If so, does a relish garbage collection and Go garbage collection.
//
var GcIntervalSeconds = 600