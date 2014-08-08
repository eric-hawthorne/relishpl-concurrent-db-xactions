// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.

package data

/*
   transaction.go - metadata for tracking database transactions and their effect on in-memory objects.
                    Used as part of a mechanism which keeps the state of in-memory relish data objects
                    consistent with the committed database state of those objects.

                    LIMITATION: Currently, each relish application serializes each database query/update and each
                    transaction until committed/rolled back.
                    Also, if an application both reads and modifies relish persistent objects, then only
                    one instance of the application should be run at a time on a given computer with a local db.

                    These restrictions may be lifted once even better mechanisms are built to ensure
                    memory vs database object state consistency. The restrictions are currently in place because
                    it would be horribly bad in relish if the in-memory state of an object got inconsistent with 
                    the db state unbeknownst to the relish program, and then further updates were made by the program
                    to the object state.
*/


import (
   "sync"
 )


/*
Represents (complements) a database transaction.
*/
type RTransaction struct {
   DirtyObjects map[Persistable]bool  // Objects dubbed in the scope of this transaction 
                                  // or persistent and having attributes changed 
                                  // within the scope of this transaction
   mutex sync.RWMutex
   IsRolledBack bool
}

func NewTransaction() (tx *RTransaction) {
	tx = &RTransaction{DirtyObjects: make(map[Persistable]bool)}
  tx.mutex.Lock()
  return tx
}

/*
// A special singleton transaction that stands in for any transaction that has been
// rolled back. Persistent objects which were dirtied (affected) within a transaction
// that was subsequently rolled back get their transaction reference set to this 
// transaction, which indicates that their attributes need to be refreshed from DB.
*/ 
var RolledBackTransaction *RTransaction = &RTransaction{}

/*
Call this after the database rollback occurs.
When this has been called, object.IsRolledBack() is true, until object.SetTransaction(nil)
is called.

*/
func (tx *RTransaction) RollBack() {
	for object := range tx.DirtyObjects {
//       fmt.Println("Rolling back",object)

       object.RollBack()
	}
  tx.IsRolledBack = true
  tx.mutex.Unlock()
}

func (tx *RTransaction) RLock() {
  tx.mutex.RLock()
}

func (tx *RTransaction) RUnlock() {
  tx.mutex.RUnlock()
}


/*
Call this after the database commit occurs.
*/
func (tx *RTransaction) Commit() {
	for object := range tx.DirtyObjects {
       object.This().SetStoredLocally()
       object.SetTransaction(nil)
	}
  tx.mutex.Unlock()
}
