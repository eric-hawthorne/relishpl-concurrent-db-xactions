// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package implements data persistence in the relish language environment.

package persist

/*
   transaction.go 
   
   This file implements transaction management and database access serialization aspects of
   sqlite persistence of relish objects

   SEE ALSO the dbThread type in interp/interpreter.go for an extended implementation of the methods
   below that fully implements safe multi-threaded access to the database.
*/

import (
	"sync"
   "math/rand"
   "time"
)


// A mutual-exclusion lock which ensures single-goroutine-at-a-time serialized access to the database.
var dbMutex sync.Mutex

const N_BEGIN_TRIES = 30
const N_COMMIT_TRIES = 30
const TRY_GAP_WIDENING_MS_INCREMENT = 100  // longest wait will be 6 seconds + random factor 
 
/*
Begins an immediate-mode database transaction.
*/
func (db *SqliteDBThread) BeginTransaction() (err error) {
   err = db.ExecStatement("BEGIN IMMEDIATE TRANSACTION")

   var tryGapWidth int64 = TRY_GAP_WIDENING_MS_INCREMENT
   var r int64
   for i := 0; i < N_BEGIN_TRIES && err != nil; i ++ {
      r = rand.Int63n(1000 + tryGapWidth)
      tryGapWidth += TRY_GAP_WIDENING_MS_INCREMENT
      time.Sleep( time.Duration( (tryGapWidth + r) * 1000000 ) )
      err = db.ExecStatement("BEGIN IMMEDIATE TRANSACTION")
   }
   return
}

/*
Commits the in-effect database transaction.
*/
func (db *SqliteDBThread) CommitTransaction() (err error) {
   err = db.ExecStatement("COMMIT TRANSACTION")

   var tryGapWidth int64 = TRY_GAP_WIDENING_MS_INCREMENT
   var r int64
   for i := 0; i < N_COMMIT_TRIES && err != nil; i ++ {
      r = rand.Int63n(1000 + tryGapWidth)
      tryGapWidth += TRY_GAP_WIDENING_MS_INCREMENT
      time.Sleep(  time.Duration( (tryGapWidth + r) * 1000000 ) )
      err = db.ExecStatement("COMMIT TRANSACTION")
   }
   return

   return
}

/*
Rolls back the in-effect database transaction.
*/
func (db *SqliteDBThread) RollbackTransaction() (err error) {
   err = db.ExecStatement("ROLLBACK TRANSACTION")
   return
}


func (db *SqliteDBThread) UseDB() { 
   panic("should not ever be called. Generic DBThread type calls its method instead.")
}



func (db *SqliteDBThread) ReleaseDB() bool {
   panic("should not ever be called. Generic DBThread type calls its method instead.")
   return false
}

