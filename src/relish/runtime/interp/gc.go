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
)

var m runtime.MemStats

/*
Run the garbage collector loop.

Checks every CG_INTERVAL_MINUTES to see if it should run the GC.
Runs it if the current allocated-and-not-yet-freed memory is greater than twice 
the lowest memory allocated-and-not-freed level since GC was last run.
*/
func (i *Interpreter) GCLoop() {
	
    defer Un(Trace(GC2_,"GCLoop"))

    var prevA uint64

    for {
	    time.Sleep(time.Duration(params.GcIntervalSeconds) * time.Second)
	    // time.Sleep(4 * time.Second)    
		    
	    runtime.ReadMemStats(&m)
	    if m.Alloc > prevA * 2  {	
		   Logln(GC_,"GC because Prev Alloc",prevA,", Alloc",m.Alloc)
		
	       i.GC()
	
	       runtime.ReadMemStats(&m)	
	       prevA = m.Alloc
	
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

    i.rt.Sweep()
}


