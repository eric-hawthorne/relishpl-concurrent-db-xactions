// Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package implements data persistence in the relish language environment.

package persist

/*
   This file implements transaction management and database access serialization aspects of
   sqlite persistence of relish objects

   SEE ALSO the dbThread type in interp/interpreter.go for an extended implementation of the methods
   below that fully implements safe multi-threaded access to the database.
*/

import (
	"sync"
	. "relish/dbg"
)


// A mutual-exclusion lock which ensures single-goroutine-at-a-time serialized access to the database.
var dbMutex sync.Mutex

 
/*
Begins an immediate-mode database transaction.
TODO: Should get this out of QueueStatements so we can properly handle errors.
*/
func (db *SqliteDB) BeginTransaction() (err error) {
   db.QueueStatements("BEGIN IMMEDIATE TRANSACTION")
   return
}

/*
Commits the in-effect database transaction.
TODO: Should get this out of QueueStatements so we can properly handle errors.
*/
func (db *SqliteDB) CommitTransaction() (err error) {
   db.QueueStatements("COMMIT TRANSACTION")
   return
}

/*
Rolls back the in-effect database transaction.
TODO: Should get this out of QueueStatements so we can properly handle errors.
*/
func (db *SqliteDB) RollbackTransaction() (err error) {
   db.QueueStatements("ROLLBACK TRANSACTION")
   return
}

/*
Lock the dbMutex.
Used to ensure exlusive access to db for single db reads / writes 
for which we don't want to manually start a long-running transaction.
(Or may also be used in multi-threaded extensions of the Begin,Commit,RollbackTransaction methods.)

This method will block until no other goroutine is using the database.
*/
func (db *SqliteDB) UseDB() {	
   Logln(PERSIST2_,"About to lock the dbMutex") 	
   dbMutex.Lock()
   Logln(PERSIST2_,"Locked the dbMutex")
}

/*
Unlock the dbMutex.
*/	
func (db *SqliteDB) ReleaseDB() bool{
   dbMutex.Unlock()
   Logln(PERSIST2_,"Unlocked the dbMutex") 	
   return true
}



