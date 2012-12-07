// Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.
// Abstraction of persistence service for relish data.

package data

type DB interface {
	EnsureTypeTable(typ *RType) (err error)
	QueueStatements(statementGroup string)
	PersistSetAttr(obj RObject, attr *AttributeSpec, val RObject, attrHadValue bool) (err error)
	PersistAddToAttr(obj RObject, attr *AttributeSpec, val RObject, insertedIndex int) (err error)
	PersistRemoveFromAttr(obj RObject, attr *AttributeSpec, val RObject, removedIndex int) (err error)
    PersistRemoveAttr(obj RObject, attr *AttributeSpec) (err error) 	
    PersistClearAttr(obj RObject, attr *AttributeSpec) (err error)
	EnsurePersisted(obj RObject) (err error)
	EnsureAttributeAndRelationTables(t *RType) (err error)
	ObjectNameExists(name string) (found bool, err error)
	NameObject(obj RObject, name string)
	RecordPackageName(name string, shortName string)
	FetchByName(name string, radius int) (obj RObject, err error)
	Fetch(id int64, radius int) (obj RObject, err error)
	FetchAttribute(objId int64, obj RObject, attr *AttributeSpec, radius int) (val RObject, err error)

	/*
	
	Given an object type and an OQL selection criteria clause in a string, set the argument collection to contain 
	the matching objects from the the database.

	e.g. of first two arguments: vehicles/Car, "speed > 60 order by speed desc"   
	*/
	
    FetchN(typ *RType, oqlSelectionCriteria string, radius int, objs *[]RObject) (err error) 

    /*
    Close the connection to the database.
    */
	Close()


	/*
    Begins an immediate-mode database transaction.	
    Implementations may also first lock program access to the database to ensure a single goroutine at a time
    runs a database transaction and no other goroutines interact with the database at all during the transaction.
	*/
	BeginTransaction() (err error)

	/*
    Implementations may also unlock program access to the database to allow other goroutines to use the database.
	*/
	CommitTransaction() (err error)

	/*
    Implementations may also unlock program access to the database to allow other goroutines to use the database.
	*/
	RollbackTransaction() (err error)
	
	/*
	Lock the dbMutex.
	Used to ensure exlusive access to db for single db reads / writes 
	for which we don't want to manually start a long-running transaction.
	(Or may also be used in multi-threaded extensions of the Begin,Commit,RollbackTransaction methods.)

	This method will block until no other goroutine is using the database.
	*/
	UseDB()
	
	/*
	Unlock the dbMutex.
	*/	
	ReleaseDB()
}


/*
A single "thread" of execution of the relish interpreter, complete with its own relish data stack (hidden).
This interface provides access to the package whose method is currently executing, 
to the actual RMethod that is currently executing, and to the DBThread which can execute
database queries in a multi-threaded context. 

Note that each InterpreterThread actually corresponds to one goroutine. Goroutines are lightweight userspace
threads in Go which may be cooperative coroutines or may be mapped onto separate processor threads and/or cores.
Typically, multiple goroutines will execute in a single OS thread, but if multiple cores are available and
Go is configured to use them, then groups of goroutines may be apportioned across multiple OS-threads and cores. 
*/
type InterpreterThread interface {
	/*
	The package context of the executing method.
	*/
	Package() *RPackage
	
	/*
	The executing method.
	*/
	Method() *RMethod

    /*
    A db connection thread. Used to serialize access to the database in a multi-threaded environment,
    and to manage database transactions.
    */
	DB() DB
}

type DBThread interface {
	DB

	/*
	Grabs the dbMutex when it can (blocking until then) then executes a BEGIN IMMEDIATE TRANSACTION sql statement.
	Does not unlock the dbMutex or release this thread's ownership of the mutex. 
	Use CommitTransaction or RollbackTransaction to do that.
	*/
	BeginTransaction() (err error)

	/*
	Executes a COMMIT TRANSACTION sql statement. If it succeeds, unlocks the dbMutex and releases this thread's ownership
	of the mutex.
	If it fails (returns a non-nil error), does not unlock the dbMutex or release this thread's ownership of the mutex.

	In the error case, the correct behaviour is to either retry the commit, do a rollback, or just call ReleaseDB to
	unlock the dbMutex and release this thread's ownership of the mutex.
	*/
	CommitTransaction() (err error)

	/*
	Executes a ROLLBACK TRANSACTION sql statement. If it succeeds, unlocks the dbMutex and releases this thread's ownership
	of the mutex.
	If it fails (returns a non-nil error), does not unlock the dbMutex or release this thread's ownership of the mutex.

	In the error case, the correct behaviour is to either retry the rollback, or just call ReleaseDB to
	unlock the dbMutex and release this thread's ownership of the mutex.
	*/
	RollbackTransaction() (err error)
	
	/*
	If the thread does not already own the dbMutex, lock the mutex and
	flag that this thread owns it.
	Used to ensure exlusive access to db for single db reads / writes 
	for which we don't want to manually start a long-running transaction.

	This method will block until no other DBThread is using the database.
	*/
	UseDB()
	
	/*
	If the thread owns the dbMutex, unlock the mutex and
	flag that this thread no longer owns it.
	*/	
	ReleaseDB()
}