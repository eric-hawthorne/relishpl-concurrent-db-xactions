// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.

package data

/*
   gc.go -  garbage collection - relish needs its own gc on top of Go gc, because relish has global in-memory caches
                                 (maps) of relish data objects, and presence in such a cache would
                                 prevent an object which is unreachable in the relish program from being
                                 collected by Go gc. Currently, there are two caches, a memcache of persistent objects
                                 used to ensure there is only one copy of each identified object fetched into memory from db,
                                 and a map from object to non-persistent-object id, which is used in debugging
                                 printouts of objects.
*/


import (
   . "relish/dbg"
   "sync"
   "fmt"
   "sort"
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


// var nRlocking int32
// var nRlocked int32
// var nLocking int32
// var nLocked int32


func GCMutexRLock(msg string) {
//   nRlocking++	
//   if nRlocked > 0 || (len(msg) > 1 && msg[0] == 'r' && msg[1] == 'e') {
//      Logln(GC_,">>> GCMutex.RLock'ing() ",msg, "nRlocking =", nRlocking, "nRlocked =", nRlocked)		
//   }
   GCMutex.RLock()
//   nRlocking--
//   nRlocked++
//   if nRlocked > 1 || (len(msg) > 1 && msg[0] == 'r' && msg[1] == 'e') {   
//     Logln(GC_,">>> GCMutex.RLock() rlocked! ",msg, "nRlocking =", nRlocking, "nRlocked =", nRlocked)   
//   }
}

func GCMutexRUnlock(msg string) {
//   nRlocked--	
//   if nRlocked > 0 || (len(msg) > 1 && msg[0] == 'D' && msg[1] == 'e') {   
//      Logln(GC_,"<<< GCMutex.RUnlock()",msg, "nRlocked =", nRlocked)	
//   }
   GCMutex.RUnlock()
}

func GCMutexLock(msg string) {
//   nLocking++
//   Logln(GC_,">>>>>> GCMutex.Lock'ing()",msg,"nLocked =",nLocked,"nRlocked =", nRlocked)		
   GCMutex.Lock()
//   nLocking--
//   nLocked++
//   Logln(GC_,">>> GCMutex.RLock() rlocked! ",msg, "nLocked =", nLocked)      
}

func GCMutexUnlock(msg string) {
//   nLocked--	
//   Logln(GC_,"<<<<<< GCMutex.Unlock()",msg, "nLocked =", nLocked)	
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
Mark the global variables (known as the context) as reachable.
*/
func (rt *RuntimeEnv) MarkContext() {
    defer Un(Trace(GC2_,"MarkContext"))
	for _,obj := range rt.context {
		obj.Mark()
	}
    Logln(GC2_,"Marked",len(rt.context),"context-map objects (global variables) and their associates.")		
}



// 
// Objects in-transit are objects currently in a channel. They are in-transit between
// relish-program "threads" (go-routines). When placed into a channel, an object (if the channel is buffered)
// may be no longer referenced except for its presence in the channel buffer. 
// Therefore, when doing a "Mark", such objects (in their interstitial "wormholes") would not get marked as
// reachable, so would be erroneously collected in the next relish gc sweep.
// So we have to maintain a global list of in-transit (channel-traveling) objects and explicitly mark
// all of those as well as the objects that get marked because of reachability from a relish program-stack.
//

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


/*
Remove from runtime-global maps, those objects which are unreachable from thread stacks or from constants.
By removing these relish-unreachable objects from the maps, the objects become garbage-collectable
by Go.

NOTE: Factors we could be counting to decide whether to copy maps or delete from them:

1. Every nth GC   n = 10 say  (but may not be often enough)

2. # of elements formerly in map (N0)

3. Fraction of elements deleted F  (per attr or total)

4. Statistics of F over up to the last n GCs  (Fbar)

5. # of elements still in map N1

6. # of elements removed R (last time, or this time) (total not per attr)

7. Rbar over up to the last nGCs (total not per attr)

*/

var nObjs0, nObjects0, nIds0, nIdents0, nAttrs0, nAtt0 int  // from last collection 

var nObjsCumRemoved, nIdsCumRemoved, nAttCumRemoved int // since last copying of maps


func (rt *RuntimeEnv) Sweep2() {

  
  var nObjs, nObjects, nIds, nIdents int //, nAtt, nAttrs, nAttrs1, nAtt1 int
  var nObjsLeft, nIdsLeft int //, nAtt1Left int

  var copiedObjectsCache, copiedObjectIds bool //, copiedAttrs bool

  nObjects = len(rt.objects)

  if (nObjsCumRemoved > 50000) || (nObjs0 > 1000) || (nObjects0 > 0 && nObjs0 * 100 / nObjects0 > 30) {  // copy the persistent objects cache
    freshObjectsMap := make(map[int64]RObject)

    for key, obj := range rt.objects {
       if obj.IsMarked() == markSense {  // Reachable
           freshObjectsMap[key] = obj
           nObjsLeft++
       }  
    } 
    nObjs = nObjects - nObjsLeft
    rt.objects = freshObjectsMap
    copiedObjectsCache = true

  } else {  // delete unreachable objects from existing persistent objects cache
    for key, obj := range rt.objects {
       if obj.IsMarked() != markSense {  // Not reachable
           delete(rt.objects,key)
           nObjs++
       }  
    } 
  }

  nIdents = len(rt.objectIds) 

  if  (nIdsCumRemoved > 50000) || (nIds0 > 1000) || (nIdents0 > 0 && nIds0 * 100 / nIdents0 > 30) {  // copy the non-persistent object ids map
    freshObjectIdsMap := make(map[RObject]uint64)

    for obj,oid := range rt.objectIds {
       if obj.IsMarked() == markSense {  // Reachable
         freshObjectIdsMap[obj] = oid
         nIdsLeft++
       }  
    }
    nIds = nIdents - nIdsLeft
    rt.objectIds = freshObjectIdsMap
    copiedObjectIds = true

  } else { // delete unreachable objects from existing non-persistent object ids map
    for obj := range rt.objectIds {
       if obj.IsMarked() != markSense {  // Not reachable
         delete(rt.objectIds,obj)
         nIds++
       }  
    }
  }


  nObjs0 = nObjs
  nObjects0 = nObjects
  nIds0 = nIds
  nIdents0 = nIdents
  // nAtt0 = nAtt
  // nAttrs0 = nAttrs

  if copiedObjectsCache {
       Logln(GC2_,"Copied persistent objects cache map (prevent Go map memory leak).")   
       nObjsCumRemoved = 0    
  } else {
       nObjsCumRemoved += nObjs
  }

  if copiedObjectIds {
       Logln(GC2_,"Copied non-persistent object ids map (prevent Go map memory leak).") 
       nIdsCumRemoved = 0    

  } else {
       nIdsCumRemoved += nIds
  } 

  markSense = ! markSense 
  
 
  Logln(GC2_,"Swept",nObjs,"of",nObjects,"from cache,\n",nIds,"of",nIdents,"from non-persistent ids.")   
}


/*
Remove from runtime-global maps, those objects which are unreachable from thread stacks or from constants.
By removing these relish-unreachable objects from the maps, the objects become garbage-collectable
by Go.
*/
func (rt *RuntimeEnv) Sweep() {
	
	var nObjs, nObjects, nIds, nIdents int 
	
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

	markSense = ! markSense	
	
  Logln(GC2_,"Swept",nObjs,"of",nObjects,"from cache,\n",nIds,"of",nIdents,"from non-persistent ids.")		
}


// MEMORY LEAK DEBUGGING SUPPORT
// Note: Requires uncommenting some code in the Mark() method in robject.go at around line 302.
// That code may have been commented out for speed improvement.

var n_runits_ever uint64 = 0
var n_rsets_ever uint64 = 0
var n_rsortedsets_ever uint64 = 0
var n_rlists_ever uint64 = 0
var n_rbiglists_ever uint64 = 0
var n_maps_ever uint64 = 0

var mappingMemory bool = false

var instanceMap map[*RType]int

func StartMemoryMapping() {
	mappingMemory = true
    instanceMap = make(map[*RType]int)
}

/*
Called by ReportMemoryMap below.
*/
func stopMemoryMapping() {
	mappingMemory = false
	instanceMap = nil
}

/*
   Return a multi-line string that constitutes a report on how many instances of each
   non-primitive, non-native type of relish object are reachable in memory.
   The report is ordered from most instances to least.
*/
func ReportMemoryMap() string {

   // First, report how many unit objects and collections have ever been allocated.
   report := fmt.Sprintf("\n%d runits\n%d rsets\n%d rsortedsets\n%d rlists\n%d maps ever allocated\n\n", 
                         n_runits_ever,n_rsets_ever,n_rsortedsets_ever,n_rlists_ever,n_maps_ever)

   // invert the map so we can sort it.
   
   typesByNumInstances := make(map[int][]*RType)

   for typ, n := range instanceMap {
      typs := typesByNumInstances[n]
      typesByNumInstances[n] = append(typs,typ)
   }

   var numsOfInstances []int
   for n := range typesByNumInstances {
   	  numsOfInstances = append(numsOfInstances,n)
   }

   // now sort by numbers of instances descending

   sort.Sort(sort.Reverse(sort.IntSlice(numsOfInstances)))
   
   // traverse the reversed map in order and add to report string

   for _,n := range numsOfInstances {
      types := typesByNumInstances[n]
      for _,typ := range types {
         report += fmt.Sprintf("%8d  %s\n",n, typ.Name)
      }
   }

   stopMemoryMapping()
   return report
}