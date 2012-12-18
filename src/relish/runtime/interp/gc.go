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
   "time"
   "runtime"
   "relish/runtime/data"
)

const GC_INTERVAL_MINUTES = 10 

var m runtime.MemStats

/*
Run the garbage collector loop.
*/
func (i *Interpreter) GCLoop() {
	
    defer Un(Trace(GC_,"GCLoop"))

    var prevA uint64

    for {
//	    time.Sleep(GC_INTERVAL_MINUTES * time.Minute)
	    time.Sleep(GC_INTERVAL_MINUTES * time.Second * 2)
		    
	    runtime.ReadMemStats(&m)
//	    if m.Alloc > prevA * 2 {
	    if m.Alloc >= prevA  {		
		   Logln(GC_,"Prev Alloc",prevA,", Alloc",m.Alloc)
		
	       i.GC()
	
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
    data.GCMutex.Lock()
    defer data.GCMutex.Unlock()    
    defer Un(Trace(GC_,"GC"))    
    i.mark()
    i.sweep()

}


/*
Mark all reachable objects as being reachable.
*/
func (i *Interpreter) mark() {
    defer Un(Trace(GC_,"mark"))

    nThreads := len(i.threads)
    Logln(GC_,nThreads,"interpreter threads active.")

	for t := range i.threads {
	   t.Mark()
	}
	
    i.rt.MarkConstants()
}


func (i *Interpreter) sweep() {
    defer Un(Trace(GC_,"sweep"))

    i.rt.Sweep()
}


