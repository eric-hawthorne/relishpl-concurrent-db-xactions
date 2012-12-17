// Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.

package data

/*
   gc.go -  garbage collection
*/


import (
   . "relish/dbg"
)



var markSense bool = true  // What value of object Marked flag means "is reachable" - if true, 1, if false, 0

/*
Mark the constants as reachable.
*/
func (rt *RuntimeEnv) MarkConstants() {
    defer Un(Trace(GC_,"MarkConstants"))
	for _,obj := range rt.constants {
		obj.Mark()
	}
    Logln(GC__,"Marked",len(rt.constants),"constants and their associates.")		
}


/*
Remove from runtime-global maps, those objects which are unreachable from thread stacks or from constants.
By removing these relish-unreachable objects from the maps, the objects become garbage-collectable
by Go.
*/
func (rt *RuntimeEnv) Sweep() {
	
	var nObjs, nObjects, nIds, nIdents, nAtt, nAttrs int
	
	nObjects = len(rt.objects)
	for key, obj := range rt.objects {
	   if obj.IsMarked() != markSense {  // Not reachable
	       delete(rt.objects,key)
	       nObjs++
	   }	
	} 
	nIdents = len(rt.objectIds) 
	for obj := range rt.objectIds {
	   if obj.IsMarked() != markSense {  // Not reachable
		   delete(rt.objectIds,obj)
		   nIds++
	   }	
	}
	for _,attrMap := range rt.attributes {
		nAttrs += len(attrMap)
		for obj := range attrMap {
		   if obj.IsMarked() != markSense {  // Not reachable
			   delete(attrMap,obj)
			   nAtt++
		   }			
		}
	}
	markSense = ! markSense	
	
    Logln(GC__,"Swept",nObjs,"of",nObjects,"from cache,\n",nIds,"of",nIdents,"from non-persistent ids,\n",nAtt,"of",nAttrs,"attribute associations.")		
}