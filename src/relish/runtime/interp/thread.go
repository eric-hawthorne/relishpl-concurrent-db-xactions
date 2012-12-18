// Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the operation of the intermediate-code interpreter
// in the relish language.

package interp

/*
   thread.go - One Thread is created per goroutine. The Thread maintains an execution-stack of RObjects
*/

import (
	//"os"
	"fmt"
	"relish/compiler/ast"
	. "relish/dbg"
	. "relish/runtime/data"
	"errors"
)


const DEFAULT_STACK_DEPTH = 50 // DEFAULT INITIAL STACK DEPTH PER THREAD

/*
If parent is nil, something else must take care of initializing 
the ExecutingMethod and ExecutingPackage attributes of the new thread.
*/
func (i *Interpreter) NewThread(parent *Thread) *Thread {
	defer UnM(parent,TraceM(parent,INTERP_TR, "NewThread"))
	return newThread(DEFAULT_STACK_DEPTH, i, parent)
}

/*
If parent is nil, something else must take care of initializing 
the ExecutingMethod and ExecutingPackage attributes of the new thread.
*/
func newThread(initialStackDepth int, i *Interpreter, parent *Thread) *Thread {
	dbt := &dbThread{db:i.rt.DB()}
	t := &Thread{Pos: -1, Base: -1, Stack: make([]RObject, initialStackDepth), EvalContext: nil, dbConnectionThread: dbt}
	if parent != nil {
		t.ExecutingMethod = parent.ExecutingMethod
		t.ExecutingPackage = parent.ExecutingPackage
	} 
	t.EvalContext = &methodEvaluationContext{i, t}

	i.registerThread(t)
	
	return t
}


func (i *Interpreter) registerThread(t *Thread) {
    gcMutex.RLock()	
    defer gcMutex.RUnlock()	

	defer UnM(t, TraceM(t, GC2_, "RegisterThread"))
	
	i.threads[t] = true
}

/*
Remember to call this when thread execution is done!
*/
func (i *Interpreter) DeregisterThread(t *Thread) {
    gcMutex.RLock()
    defer gcMutex.RUnlock()	

    LoglnM(t, GC2_, "DeregisterThread(")

	delete(i.threads,t)
    RemoveContext(t) 	

    Logln(GC2_, ")DeregisterThread")
}

/*
Not really a thread. Actually equivalent to a goroutine.

Each routine that is called pushes its context onto the stack, thus has a stack "frame".
The first element in the stack frame is an Int32 which holds the position on the stack of the beginning of the 
previous routine stack frame, or holds -1 if this is the first routine call pushed on the stack.
The value of a local variable or a routine parameter is found by taking the variable's Offset and adding it
to the Base, then dereferencing that stack item.
*/
type Thread struct {
	Pos         int // position of the top of the stack
	Base        int // position on stack of the beginning of the currently executing routine's stack frame
	Stack       []RObject
	EvalContext *methodEvaluationContext // Basically a self-reference, but includes a reference to the interpreter

	ExecutingMethod *RMethod       // Shortcut for dispatch efficiency
	ExecutingPackage *RPackage   // Shortcut for dispatch efficiency

	dbConnectionThread *dbThread  // Manages serialized and transactional access to the database in
	                              // multi-threaded environment

    // This may be temporary - it is used for generators inside collection constructors, but may
    // be replaced by proper go-routine-and-channel generators or closures.
    Objs       []RObject   // A list of objects that will be built up then become owned by a proper collection object 
                           // and detached from the Thread.
	YieldCardinality int   // How many objects per iteration the current generator is expected to append to Objs
	
	err string  // Will be "" unless we are in a stack-unrolling panic
}




func (t *Thread) Push(obj RObject) {
    gcMutex.RLock()		
	defer UnM(t, TraceM(t, INTERP_TR3, "Push", obj))
	t.Pos++
	if len(t.Stack) <= t.Pos {
		oldStack := t.Stack
		t.Stack = make([]RObject, len(t.Stack)*2)
		copy(t.Stack, oldStack)
	}
	t.Stack[t.Pos] = obj
	if Logging(STACK_) && Logging(INTERP_TR3) {
		t.Dump()
	}
	gcMutex.RUnlock()	
}

func (t *Thread) Reserve(n int) {
    gcMutex.RLock()		
	defer UnM(t, TraceM(t, INTERP_TR3, "Reserve", n))
	for len(t.Stack) <= t.Pos+n {
		oldStack := t.Stack
		t.Stack = make([]RObject, len(t.Stack)*2)
		copy(t.Stack, oldStack)
	}
	t.Pos += n
	gcMutex.RUnlock()	
}




func (t *Thread) PushNoLock(obj RObject) {	
	defer UnM(t, TraceM(t, INTERP_TR3, "Push", obj))
	t.Pos++
	if len(t.Stack) <= t.Pos {
		oldStack := t.Stack
		t.Stack = make([]RObject, len(t.Stack)*2)
		copy(t.Stack, oldStack)
	}
	t.Stack[t.Pos] = obj
	if Logging(STACK_) && Logging(INTERP_TR3) {
		t.Dump()
	}	
}

func (t *Thread) ReserveNoLock(n int) {	
	defer UnM(t, TraceM(t, INTERP_TR3, "Reserve", n))
	for len(t.Stack) <= t.Pos+n {
		oldStack := t.Stack
		t.Stack = make([]RObject, len(t.Stack)*2)
		copy(t.Stack, oldStack)
	}
	t.Pos += n
}





/*
Call this at the beginning of a method call, to push the previous (outer) method's stack frame base onto the stack.
When finished evaluating arguments to the call, then complete the context store by calling setBase with the
value that has been returned by PushBase().
The numReturnArgs argument says how many stack positions to leave free below the base-pointer position for
results (return values) of the method call.

Returns the position of the base pointer which holds the stack-pointer to the previous stack-frame base position.
*/
func (t *Thread) PushBase(numReturnArgs int) int {
    gcMutex.RLock()		
	defer UnM(t, TraceM(t, INTERP_TR3, "PushBase"))

	if numReturnArgs > 0 {
		t.ReserveNoLock(numReturnArgs)
	}
	t.PushNoLock(Int32(t.Base))
	t.ReserveNoLock(2) // Reserve space for the currently-executing-method reference and code offset in current method
	gcMutex.RUnlock()		
	return t.Pos - 2
	
}

func (t *Thread) SetBase(newBase int) {
	defer UnM(t, TraceM(t, INTERP_TR3, "SetBase", newBase))
	t.Base = newBase
}

/*
Call this to return from a method call.
Pops the stack to just before the beginning of the current routine's stack frame, and
sets the thread's Base pointer to the beginning of the previous (outer) routine's stack frame.


                            B       P
0   1   2   3   4   5   6   7   8   9
-1  v1  v2  0   v1  v2  v3  3   v1  v2

*/
func (t *Thread) PopBase() {
	defer UnM(t, TraceM(t, INTERP_TR3, "PopBase"))
	obj := t.PopN(t.Pos - t.Base + 1) // 9 - 7 + 1 = 3
	t.Base = int(obj.(Int32))
	t.ExecutingMethod = t.Stack[t.Base+1].(*RMethod)
	t.ExecutingPackage = t.ExecutingMethod.Pkg
	LogM(t, INTERP3_, "---Base = %d\n", t.Base)
}

/*
Return the value of the local variable (or parameter) with the given offset from the current routine's 
stack base.
*/
func (t *Thread) GetVar(offset int) (obj RObject, err error) {
	defer UnM(t, TraceM(t, INTERP_TR3, "GetVar", "offset", offset, "stack index", t.Base+offset))
	if offset == -99 {
		err = errors.New("Unassigned variable.")
		return
	}
	obj = t.Stack[t.Base+offset]
	return
}

func (t *Thread) Pop() RObject {
	gcMutex.RLock()
	obj := t.Stack[t.Pos]
	defer UnM(t, TraceM(t, INTERP_TR3, "Pop", "==>", obj))
	t.Stack[t.Pos] = nil // ensure var/param value is garbage-collectable if not otherwise referred to.
	t.Pos--
	if Logging(STACK_) && Logging(INTERP_TR3) {
		t.Dump()
	}
	gcMutex.RUnlock()	
	return obj
}

func (t *Thread) PopNoLock() RObject {
	obj := t.Stack[t.Pos]
	defer UnM(t, TraceM(t, INTERP_TR3, "Pop", "==>", obj))
	t.Stack[t.Pos] = nil // ensure var/param value is garbage-collectable if not otherwise referred to.
	t.Pos--
	if Logging(STACK_) && Logging(INTERP_TR3) {
		t.Dump()
	}
	return obj
}

/*
Efficiently pop n items off the stack.
*/
func (t *Thread) PopN(n int) RObject {
	gcMutex.RLock()	
	lastPopped := t.Pos - n + 1
	obj := t.Stack[lastPopped]
	defer UnM(t, TraceM(t, INTERP_TR3, "PopN", n, "==>", obj))
	for i := t.Pos; i >= lastPopped; i-- {
		t.Stack[i] = nil // ensure var/param value is garbage-collectable if not otherwise referred to.
	}
	t.Pos -= n
	if Logging(STACK_) && Logging(INTERP_TR3) {
		t.Dump()
	}
	gcMutex.RUnlock()	
	return obj
}

/*
Return a slice which represents the top n objects on the thread's stack.
In order from bottom most (oldest pushed) to top most (most recently pushed).
*/
func (t *Thread) TopN(n int) []RObject {
	return t.Stack[t.Pos-n+1 : t.Pos+1]
}

func (t *Thread) copyStackFrameFrom(parent *Thread, numReturnVals int) {
   gcMutex.RLock()		
   n := parent.Pos - parent.Base + numReturnVals + 1	
   src := parent.Stack[parent.Base - numReturnVals:parent.Pos+1]
   if copy(t.Stack, src) != n {
   	   panic("stack copy range exception during go-routine spawn.")
   }
   t.Base = numReturnVals
   t.Pos = n - 1
   gcMutex.RUnlock()	   
}

/*
Mark all structured objects on the stack as reachable and safe from garbage collection.
*/
func (t *Thread) Mark() {
	defer UnM(t, TraceM(t, GC_, "Mark"))
	

    
	for i := t.Pos; i >= 0; i-- {
		if t.Stack[i] != nil {
			t.Stack[i].Mark()
		}
	}
	
    for _,obj := range t.Objs {
	   obj.Mark()
    }	

    LoglnM(t,GC_,"  Marked",t.Pos + 1 + len(t.Objs),"objects on stack and their associates.")
   
	// When finished, signal somehow.
	// The philosophy should be each stack has to signal when it is finished marking,
	// then it has to wait on an RLock() of an RWMutex for the RT.GC() to finish and unlock the RWMutex.

	// So interpreter.GC() should 
	// a) Know how many threads are active
	// b) lightweight flag each thread (set a bool in the thread) to begin marking
	//     - It's ok if the flag is missed for a while
	// - Flags actually have to wait for all of them to rendezvous before
	//   commencing to mark, so that it is safe to traverse a collection and mark its members
	// c) wait for all threads to finish marking
	// d) proceed with the sweep 
}

/*
DEBUG printout of stack
*/
func (t *Thread) Dump() {
	LogMutex.Lock()
	fmt.Println("------STACK----------")
	for i := t.Pos; i >= 0; i-- {
		fmt.Printf("%3d: %v\n", i, t.Stack[i])
	}
	fmt.Printf("Pos : %d\n", t.Pos)
	fmt.Printf("Base : %d\n", t.Base)
	fmt.Println("---------------------")
	LogMutex.Unlock()
}

func (t *Thread) CodeFile() *ast.File {
   return t.ExecutingMethod.CodeFile()
}

/*
The package context of the executing method.
*/
func (t *Thread) Package() *RPackage {
	return t.ExecutingPackage
}
	
/*
The executing method.
*/
func (t *Thread) Method() *RMethod {
	return t.ExecutingMethod
}



/*
The DBThread which can execute db queries in a serialized fashion in a multi-threaded environment.
*/
func (t *Thread) DB() DB {
   return t.dbConnectionThread
}

func (t *Thread) Err() string {
	return t.err
}



/*
    Has a reference to the DB. 
    Executes DB queries in a serialized fashion in a multi-threaded environment.
    Also manages database transactions.
*/
type dbThread struct {
	db DB  // the database connection-managing object

	acquiringDbLock bool  // This thread is in the process of acquiring and locking the dbMutex 
	                      // (but may still be blocked waiting for the mutex to be unlocked by another thread)
	
	dbLockOwnershipDepth int  // How many nested claims has this thread currently made for ownership of the dbMutex
                              // If > 0, this thread owns and has locked the dbMutex.
                              // Note: thread "ownership" of dbMutex is an abstract concept imposed by this dbThread s/w,
                              // because Go Mutexes are not inherently owned by any particular goroutine.	
}

/*
Grabs the dbMutex when it can (blocking until then) then executes a BEGIN IMMEDIATE TRANSACTION sql statement.
Does not unlock the dbMutex or release this thread's ownership of the mutex. 
Use CommitTransaction or RollbackTransaction to do that.
*/
func (dbt * dbThread) BeginTransaction() (err error) {
   Logln(PERSIST2_,"dbThread.BeginTransaction") 	
   dbt.UseDB()	
   err = dbt.db.BeginTransaction() 
   if err != nil {
   	   dbt.ReleaseDB()
   }
   return
}

/*
Executes a COMMIT TRANSACTION sql statement. If it succeeds, unlocks the dbMutex and releases this thread's ownership
of the mutex.
If it fails (returns a non-nil error), does not unlock the dbMutex or release this thread's ownership of the mutex.

In the error case, the correct behaviour is to either retry the commit, do a rollback, or just call ReleaseDB to
unlock the dbMutex and release this thread's ownership of the mutex.
*/
func (dbt * dbThread) CommitTransaction() (err error) {
    Logln(PERSIST2_,"dbThread.CommitTransaction") 		
	err = dbt.db.CommitTransaction()
	if err == nil {
	   dbt.ReleaseDB()
    }
   return
}

/*
Executes a ROLLBACK TRANSACTION sql statement. If it succeeds, unlocks the dbMutex and releases this thread's ownership
of the mutex.
If it fails (returns a non-nil error), does not unlock the dbMutex or release this thread's ownership of the mutex.

In the error case, the correct behaviour is to either retry the rollback, or just call ReleaseDB to
unlock the dbMutex and release this thread's ownership of the mutex.
*/
func (dbt * dbThread) RollbackTransaction() (err error) {
    Logln(PERSIST2_,"dbThread.RollbackTransaction") 	
	err = dbt.db.RollbackTransaction()
	if err == nil {
		dbt.ReleaseDB()
	}
	return
}

/*
If the thread does not already own the dbMutex, lock the mutex and
flag that this thread owns it.
Used to ensure exlusive access to db for single db reads / writes 
for which we don't want to manually start a long-running transaction.

This method will block until no other DBThread is using the database.
*/
func (dbt * dbThread) UseDB() {
   Logln(PERSIST2_,"dbThread.UseDB when ownership level is",dbt.dbLockOwnershipDepth) 		
   if dbt.acquiringDbLock {
      return	
   }	
   if dbt.dbLockOwnershipDepth == 0 {
   	   dbt.acquiringDbLock = true
   	   dbt.db.UseDB()
       dbt.acquiringDbLock = false      	
   }
   dbt.dbLockOwnershipDepth++
   Logln(PERSIST2_,"dbThread.UseDB: Set ownership level to",dbt.dbLockOwnershipDepth)    
}

/*

Remove one level of interest of this thread in the dbMutex.
If we have lost all interest in it, and
if the thread owns the dbMutex, unlock the mutex and
flag that this thread no longer owns it.
Returns false if this thread still has an interest in and lock on the dbMutex.
*/	
func (dbt * dbThread) ReleaseDB() bool {
    Logln(PERSIST2_,"dbThread.ReleaseDB when ownership level is",dbt.dbLockOwnershipDepth) 		
    if dbt.dbLockOwnershipDepth > 0 {
	   dbt.dbLockOwnershipDepth--
       Logln(PERSIST2_,"dbThread.ReleaseDB: Set ownership level to",dbt.dbLockOwnershipDepth)  	
	   if dbt.dbLockOwnershipDepth == 0 {
		  dbt.db.ReleaseDB()	
	   } else {
	      return false	
	   }
    }	
    return true
}



func (dbt * dbThread) EnsureTypeTable(typ *RType) (err error) {
   dbt.UseDB()	
   err = dbt.db.EnsureTypeTable(typ)
   dbt.ReleaseDB()  
   return 
}

func (dbt * dbThread) QueueStatements(statementGroup string) {
   dbt.UseDB()
   dbt.db.QueueStatements(statementGroup)
   dbt.ReleaseDB()
   return
}

func (dbt * dbThread) PersistSetAttr(obj RObject, attr *AttributeSpec, val RObject, attrHadValue bool) (err error) {
   dbt.UseDB()	
   err = dbt.db.PersistSetAttr(obj, attr, val, attrHadValue)
   dbt.ReleaseDB()
   return
}

func (dbt * dbThread) PersistAddToAttr(obj RObject, attr *AttributeSpec, val RObject, insertedIndex int) (err error) {
   dbt.UseDB()	
   err = dbt.db.PersistAddToAttr(obj, attr, val, insertedIndex)
   dbt.ReleaseDB()
   return 
}

func (dbt * dbThread) PersistRemoveFromAttr(obj RObject, attr *AttributeSpec, val RObject, removedIndex int) (err error) {
   dbt.UseDB()	
   err = dbt.db.PersistRemoveFromAttr(obj, attr, val, removedIndex)
   dbt.ReleaseDB()
   return   
}

func (dbt * dbThread) PersistRemoveAttr(obj RObject, attr *AttributeSpec) (err error) {
   dbt.UseDB()	
   err = dbt.db.PersistRemoveAttr(obj, attr)
   dbt.ReleaseDB() 
   return  
}

func (dbt * dbThread) PersistClearAttr(obj RObject, attr *AttributeSpec) (err error) {
   dbt.UseDB()	
   err = dbt.db.PersistClearAttr(obj, attr)
   dbt.ReleaseDB()
   return
}

func (dbt * dbThread) EnsurePersisted(obj RObject) (err error) {
   dbt.UseDB()	
   err = dbt.db.EnsurePersisted(obj)
   dbt.ReleaseDB() 
   return  
}

func (dbt * dbThread) EnsureAttributeAndRelationTables(t *RType) (err error) {
   dbt.UseDB()	
   err = dbt.db.EnsureAttributeAndRelationTables(t)
   dbt.ReleaseDB()
   return 
}

func (dbt * dbThread) ObjectNameExists(name string) (found bool, err error) {
   dbt.UseDB()
   found,err = dbt.db.ObjectNameExists(name)
   dbt.ReleaseDB()
   return
}

func (dbt * dbThread) NameObject(obj RObject, name string) {
   dbt.UseDB()
   dbt.db.NameObject(obj, name)
   dbt.ReleaseDB() 
   return  
}

func (dbt * dbThread) RecordPackageName(name string, shortName string) {
   dbt.UseDB()
   dbt.db.RecordPackageName(name, shortName)
   dbt.ReleaseDB()
   return   	
}

func (dbt * dbThread) FetchByName(name string, radius int) (obj RObject, err error) {
   dbt.UseDB()
   obj, err = dbt.db.FetchByName(name, radius)   
   dbt.ReleaseDB() 
   return  
}

func (dbt * dbThread) Fetch(id int64, radius int) (obj RObject, err error) {
   dbt.UseDB()
   obj, err = dbt.db.Fetch(id, radius)
   dbt.ReleaseDB()  
   return 
}

func (dbt * dbThread) FetchAttribute(objId int64, obj RObject, attr *AttributeSpec, radius int) (val RObject, err error) {
   dbt.UseDB()
   val, err = dbt.db.FetchAttribute(objId, obj, attr, radius)
   dbt.ReleaseDB()  
   return 
}

/*
Given an object type and an OQL selection criteria clause in a string, set the argument collection to contain 
the matching objects from the the database.

e.g. of first two arguments: vehicles/Car, "speed > 60 order by speed desc"   
*/
	
func (dbt * dbThread) FetchN(typ *RType, oqlSelectionCriteria string, radius int, objs *[]RObject) (err error) {
   dbt.UseDB()
   err = dbt.db.FetchN(typ, oqlSelectionCriteria, radius, objs)
   dbt.ReleaseDB()	
   return
}

/*
Close the connection to the database.
*/
func (dbt * dbThread) Close() {
	dbt.db.Close()
}

