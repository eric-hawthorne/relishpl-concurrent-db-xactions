// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.

package data

/*
   gc.go -  garbage collection
*/


import (
   . "relish/dbg"
   "sync"
)


var GCMutex sync.RWMutex

var DeferGC int32  // if > 0 means do not GC

var markSense bool = true  // What value of object Marked flag means "is reachable" - if true, 1, if false, 0

/*
For debugging of GC.
At all times except in GC between mark and sweep, objects encountered in the program 
should have opposite IsMarked as the current GCMarkSense. During the mark phase, if reachable they
will be given the same IsMarked value as the GCMarkSense, then those with opposite will be swept.
*/
func GCMarkSense() bool { 
   return markSense
}

func GCMutexRLock(msg string) {
   //Logln(GC_,">>> GCMutex.RLock()")		
   GCMutex.RLock()
}

func GCMutexRUnlock(msg string) {
   //Logln(GC_,"<<< GCMutex.RUnlock()")	
   GCMutex.RUnlock()
}

func GCMutexLock(msg string) {
   //Logln(GC_,">>>>>> GCMutex.Lock()")		
   GCMutex.Lock()
}

func GCMutexUnlock(msg string) {
   //Logln(GC_,"<<<<<< GCMutex.Unlock()")	
   GCMutex.Unlock()
}

/*
Mark the constants as reachable.
*/
func (rt *RuntimeEnv) MarkConstants() {
    defer Un(Trace(GC2_,"MarkConstants"))
	for _,obj := range rt.constants {
		obj.Mark()
	}
    Logln(GC2_,"Marked",len(rt.constants),"constants and their associates.")		
}


/*
Mark the reflected DataTypes as reachable.
*/
func (rt *RuntimeEnv) MarkDataTypes() {
    defer Un(Trace(GC2_,"MarkDataTypes"))
	for _,obj := range rt.ReflectedDataTypes {
		obj.Mark()
	}
    Logln(GC2_,"Marked",len(rt.ReflectedDataTypes),"reflected DataTypes.")		
}

/*
Mark the reflected Attributes as reachable.
*/
func (rt *RuntimeEnv) MarkAttributes() {
    defer Un(Trace(GC2_,"MarkAttributes"))
	for _,obj := range rt.ReflectedAttributes {
		obj.Mark()
	}
    Logln(GC2_,"Marked",len(rt.ReflectedAttributes),"reflected Attributes.")		
}

/*
Mark the gloabal variables (known as the context) as reachable.
*/
func (rt *RuntimeEnv) MarkContext() {
    defer Un(Trace(GC2_,"MarkContext"))
	for _,obj := range rt.context {
		obj.Mark()
	}
    Logln(GC2_,"Marked",len(rt.context),"context-map objects (global variables) and their associates.")		
}


var inTransitMutex sync.Mutex

func (rt *RuntimeEnv) IncrementInTransitCount(obj RObject) {
   inTransitMutex.Lock()
  
   rt.inTransit[obj]++
   
   inTransitMutex.Unlock() 
}

func (rt *RuntimeEnv) DecrementInTransitCount(obj RObject) {
   inTransitMutex.Lock()

   if rt.inTransit[obj] == 1 {
       delete(rt.inTransit,obj)	
   } else {
       rt.inTransit[obj]--	
   }

   inTransitMutex.Unlock() 
}

/*
Mark the objects in transit as reachable.
*/
func (rt *RuntimeEnv) MarkInTransit() {
    defer Un(Trace(GC2_,"MarkInTransit"))
	for obj := range rt.inTransit {
		obj.Mark()
	}
    Logln(GC2_,"Marked",len(rt.inTransit),"objects in channels, and their associates.")		
}


func (rt *RuntimeEnv) MarkAttributeVals() {
   n := 0
	for rt.markAttributeValsRound() {  Logln(GC2_,n) }
}

func (rt *RuntimeEnv) markAttributeValsRound() bool {
   newMarks := 0
	for attr,attrMap := range rt.attributes {
	   if ! attr.Part.Type.IsPrimitive {
   		for obj,val := range attrMap {
   		   if obj.IsMarked() == markSense {  // Object is marked as reachable
   			   if val.IsMarked() != markSense {
   			      val.Mark() 
   			      newMarks++
			      }
   		   }			
   		}
	   }
	}
	return newMarks > 0
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
	
    Logln(GC2_,"Swept",nObjs,"of",nObjects,"from cache,\n",nIds,"of",nIdents,"from non-persistent ids,\n",nAtt,"of",nAttrs,"attribute associations.")		
}