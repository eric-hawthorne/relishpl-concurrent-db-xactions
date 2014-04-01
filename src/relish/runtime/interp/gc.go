// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the operation of the intermediate-code interpreter
// in the relish language.

package interp

/*
   gc.go -  garbage collection
*/


import (
   . "relish/dbg"
   "time"
   "runtime"
   "relish/runtime/data"
   "relish/params"
   "runtime/pprof"   
   "os"
   "fmt"
)

var m runtime.MemStats

/*
Run the garbage collector loop.

Checks every params.GcIntervalSeconds to see if it should run the GC.
Runs it if the current allocated-and-not-yet-freed memory is greater than twice 
the lowest memory allocated-and-not-freed level since GC was last run.

!!!!!!!!!!!!!!!!
!!!!!!!!!!!!!!!!
TODO : !!!!!!!!! : Need to add a params.GcForcedIntervalSeconds e.g. 3 minutes worth 
where it will definitely GC if that amount of time has elapsed.
*/
func (i *Interpreter) GCLoop() {
	
    defer Un(Trace(GC2_,"GCLoop"))

    var prevA uint64

    var prevGcTime time.Time = time.Now()
    var currentGcTime time.Time
    var forced bool
    var nGCs int64

    for {
	    time.Sleep(time.Duration(params.GcIntervalSeconds) * time.Second)
	    // time.Sleep(4 * time.Second)    
		    
      // See if too much time has passed since last relish GC  
      forced = false  
      currentGcTime = time.Now()
      if currentGcTime.Sub(prevGcTime) > time.Duration(params.GcForceIntervalSeconds) * time.Second {
         forced = true
      }

	    runtime.ReadMemStats(&m)
	    if forced || (m.Alloc > prevA * 2) || (m.Alloc > prevA + 10000000) {	// if grew double or by more than 10 MB approx
         if forced {
		        Logln(GC_,"GC because more than",params.GcForceIntervalSeconds,"seconds since last GC. Prev Alloc",prevA,", Alloc",m.Alloc)
		     } else {
            Logln(GC_,"GC because Prev Alloc",prevA,", Alloc",m.Alloc)          
         }

	       i.GC()
	
         prevGcTime = currentGcTime
         nGCs ++
	       runtime.ReadMemStats(&m)	
	       prevA = m.Alloc
	
         Logln(GC_,"Sys",m.Sys,"Mallocs",m.Mallocs,"Frees",m.Frees)
         Logln(GC_,"HeapAlloc",m.HeapAlloc,"HeapInuse",m.HeapInuse,"HeapIdle",m.HeapIdle,"HeapReleased",m.HeapReleased,"HeapObjects",m.HeapObjects,"HeapSys",m.HeapSys)
         Logln(GC_,"StackInuse",m.StackInuse,"StackSys",m.StackSys)
         if Logging(GC_) {
            i.rt.DebugAttributesMemory()    
         }      
         if Logging(GC2_) {
            memProfileFilename := fmt.Sprintf("memory%d.prof", nGCs)
            f, err := os.Create(memProfileFilename)
            if err == nil {
              pprof.WriteHeapProfile(f)
              f.Close()
            } else {
              Logln(GC2_,"Unable to create memory profile file", memProfileFilename)
            }

         }               
	    } else if m.Alloc < prevA {
		   
		     prevA = m.Alloc   
		
	    }
    }
}

/*
Run the garbage collector.
*/
func (i *Interpreter) GC() {
    data.GCMutexLock("GC")
    defer data.GCMutexUnlock("GC")        
    defer Un(Trace(GC2_,"GC"))    
    
    if Logging(GC_) {
       data.StartMemoryMapping()
    }

    i.mark()
    i.sweep()

    Logln(GC_,data.ReportMemoryMap()) 

    runtime.GC()  // Go garbage collector.
}


/*
Mark all reachable objects as being reachable.
*/
func (i *Interpreter) mark() {
    defer Un(Trace(GC2_,"mark"))

    nThreads := len(i.threads)
    Logln(GC2_,nThreads,"interpreter threads active.")

	for t := range i.threads {
	   t.Mark()
	}
	
    i.rt.MarkConstants()

    i.rt.MarkInTransit()

    i.rt.MarkDataTypes()

    i.rt.MarkAttributes()
    
    i.rt.MarkAttributeVals()    

    i.rt.MarkContext()
}


func (i *Interpreter) sweep() {
    defer Un(Trace(GC2_,"sweep"))

    i.rt.Sweep2()  // Rename back to Sweep later.
}


