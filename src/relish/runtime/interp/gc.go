// Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the operation of the intermediate-code interpreter
// in the relish language.

package interp

/*
   gc.go -  garbage collection
*/


import (
   . "relish/dbg"
   // "sync"
)


/*
Run the garbage collector.
*/
func (i *Interpreter) GC() {
    defer Un(Trace(INTERP_TR,"GC"))
	/*
	GCMutex.Lock()
	defer GCMutex.Unlock()
	
	rt.sweep()
	markSense = ! markSense	
	*/
}

/*
func (rt *RuntimeEnv) sweep() {
	for key, obj := range rt.objects {
	   if obj.IsMarked() != markSense {  // Not reachable
	       delete(rt.objects,key)
	   }	
	}  
	for obj := range rt.objectIds {
	   if obj.IsMarked() != markSense {  // Not reachable
		   delete(rt.objectIds,obj)
	   }	
	}
	for _,attrMap := range rt.attributes {
		for obj := range attrMap {
		   if obj.IsMarked() != markSense {  // Not reachable
			   delete(attrMap,obj)
		   }			
		}
	}
}
*/
