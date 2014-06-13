// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.

package data

/*
   transaction.go - metadata for tracking database transactions and their effect on in-memory objects.
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
