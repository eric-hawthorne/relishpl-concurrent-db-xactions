// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package implements data persistence in the relish language environment.

package persist

/*
   sqlite persistence of relish objects

   This file handles definition of the database abstraction layer type, and basic generic persistence operations.
   See also persist_schema.go, and persist_data.go, which add specific methods to the database abstraction layer
   (for persisting relish types and attribute and relation specifications, 
   and for persisting relish objects and attribute assignments, respectively)


   `CREATE TABLE robject(
      id INTEGER PRIMARY KEY,
      id2 INTEGER, 
      idReversed BOOLEAN, --???
      typeName TEXT   -- Should be typeId because type should be another RObject!!!!
   )`

*/

import (
	"code.google.com/p/gosqlite/sqlite"
	"fmt"
	. "relish/dbg"
	"strings"
	. "relish/runtime/data"
)

/*
   A handle to the sqlite database that holds the open connection	   
   and has lots of specific methods for creating and working with relish
   data and metadata in the database.
*/
type SqliteDB struct {
	dbName         string
	conn           *sqlite.Conn
	statementQueue chan string
	preparedStatements map[string]*sqlite.Stmt
}

/*
   Opens a connection to a sqlite database of the specified name.
   Creates the database file if does not already exist.
   Returns a handle to the database that holds the open connection
   and has lots of specific methods for creating and working with relish
   data and metadata in the database.
*/
func NewDB(dbName string) *SqliteDB {
	db := &SqliteDB{dbName: dbName, statementQueue: make(chan string, 1000)}
	conn, err := sqlite.Open(dbName)
	if err != nil {
		panic(fmt.Sprintf("Unable to open the database '%s': %s", dbName, err))
	}
	db.conn = conn

    db.preparedStatements = make(map[string]*sqlite.Stmt)
	db.EnsureObjectTable()
	db.EnsureObjectNameTable()
	db.EnsurePackageTable()	
	
	

	go db.executeStatements()

	return db
}

/*
 Should only be used for sql statements that are not dynamically constructed.
 That is, use this only for a small set of completely predefined queries, which may nonetheless have ? parameters
 in them.
 Tests if the sql statement (cmd) is a key in a prepared statements map. If so, returns the already prepared
 statement. If not found in the map, asks the db connection to prepare the statement.
 Returns the prepared statement.
*/
func (db *SqliteDB) Prepare(cmd string) (stmt *sqlite.Stmt, err error) {
   stmt,found := db.preparedStatements[cmd]
   if ! found {
	  stmt,err = db.conn.Prepare(cmd)
	  if err == nil {
		 db.preparedStatements[cmd] = stmt
	  }
   }	
   return
}

/*
Execute asynchronously a statementGroup consisting of one or more semicolon separated INSERT or UPDATE statements.
A statementGroup is a string with semi-colon separated SQL statements in it.
Blocks only if 10,000 statementGroups are pending execution.
*/
func (db *SqliteDB) QueueStatements(statementGroup *StatementGroup) {

	// Replace everything  with the statementQueue insertion below

	Logln(PERSIST_, fmt.Sprintf("Executing statement:\n%s\n", statementGroup))

	for _, sqlStatement := range statementGroup.Statements {
			
        Logln(PERSIST_, fmt.Sprintf("Executing statement:\n%s\n", sqlStatement.Statement))
			
		err := db.conn.Exec(sqlStatement.Statement,sqlStatement.Args...)
		if err != nil {
			fmt.Printf("DB ERROR on statement:\n%s\nDetail: %s\n\n", sqlStatement.Statement, err)
		}	
	}
	// replace from here up with statementQueue statememt below.

	// db.statementQueue <- statementGroup
}


/*
Execute asynchronously a statementGroup consisting of one or more semicolon separated INSERT or UPDATE statements.
A statementGroup is a string with semi-colon separated SQL statements in it.
Blocks only if 10,000 statementGroups are pending execution.
*/
func (db *SqliteDB) QueueStatement(statement string) {

	// Replace everything  with the statementQueue insertion below

	Logln(PERSIST_, fmt.Sprintf("Executing statement:\n%s\n", statement))

	err := db.conn.Exec(statement)
	if err != nil {
		fmt.Printf("DB ERROR on statement:\n%s\nDetail: %s\n\n", statement, err)
	}
}


/*
Runs in its own goroutine, taking statements out of the queue and executing them in the sqlite database.
Loops forever.
TODO add good error handling.
*/
func (db *SqliteDB) executeStatements() {
	for {
		statementGroup := <-db.statementQueue

		// if strings.HasPrefix(statementGroup,"SELECT") {}

		Logln(PERSIST_, fmt.Sprintf("Executing statements:\n%s\n", statementGroup))

		statements := strings.SplitAfter(statementGroup, ";")
		for _, statement := range statements {
			if statement != "" {
				err := db.conn.Exec(statement)
				if err != nil {
					fmt.Printf("DB ERROR on statement:\n%s\nDetail: %s\n\n", statement, err)
				}
			}
		}
	}
}

func (db *SqliteDB) Close() {
	db.conn.Close()
}
