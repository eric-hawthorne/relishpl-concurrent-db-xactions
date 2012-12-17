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
   "sync"
   "time"
   "runtime"
)

const GC_INTERVAL_MINUTES = 10 

var m runtime.MemStats

var gcMutex sync.RWMutex

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
	    if m.Alloc > prevA * 2 {
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
    defer Un(Trace(GC_,"GC"))
    gcMutex.Lock()
    i.mark()
    i.sweep()
    gcMutex.Unlock()
}


var threadCreationMutex sync.RWMutex

var gcMarkStartWaitGroup sync.WaitGroup

var gcMarkFinishedWaitGroup sync.WaitGroup

var gcSweepFinishedWaitGroup sync.WaitGroup

/*
Mark all reachable objects as being reachable.
*/
func (i *Interpreter) mark() {
    defer Un(Trace(GC_,"mark"))

    // Lock so we don't create more threads.
    threadCreationMutex.Lock()

    nThreads := len(i.threads)
    Logln(GC_,nThreads,"interpreter threads active.")

    gcMarkStartWaitGroup.Add(nThreads)
    gcMarkFinishedWaitGroup.Add(nThreads)   

    gcSweepFinishedWaitGroup.Add(1)

	for t := range i.threads {
	   t.doGC = true
	   runtime.Gosched()
	}
	
    threadCreationMutex.Unlock()

	runtime.Gosched()
    gcMarkStartWaitGroup.Wait()
	runtime.Gosched()
    i.rt.MarkConstants()
    runtime.Gosched()
    gcMarkFinishedWaitGroup.Wait()
}


func (i *Interpreter) sweep() {
    defer Un(Trace(GC_,"sweep"))

    i.rt.Sweep()

    gcSweepFinishedWaitGroup.Done()
}


