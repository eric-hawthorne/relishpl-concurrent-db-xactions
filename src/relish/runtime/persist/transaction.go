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
)


// A mutual-exclusion lock which ensures single-goroutine-at-a-time serialized access to the database.
var dbMutex sync.Mutex

 
/*
Begins an immediate-mode database transaction.
*/
func (db *SqliteDBThread) BeginTransaction() (err error) {
   err = db.ExecStatement("BEGIN IMMEDIATE TRANSACTION")
   return
}

/*
Commits the in-effect database transaction.
*/
func (db *SqliteDBThread) CommitTransaction() (err error) {
   err = db.ExecStatement("COMMIT TRANSACTION")
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

