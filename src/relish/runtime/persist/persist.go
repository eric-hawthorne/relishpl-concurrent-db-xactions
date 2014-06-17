// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package implements data persistence in the relish language environment.

package persist

/*
   persist.go - sqlite-specific implementation of generic persistence operations

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
	sqlite "code.google.com/p/go-sqlite/go1/sqlite3"
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
Notes on sqlite db use:
TODO Must reset or finalise all open statements before rollback or commit.

 (c *Conn) Exec(cmd string, args ...interface{}) error
does a finalize at end.

(s *Stmt) Exec(args ...interface{}) error
does a reset before other stuff.

func (s *Stmt) Reset() error

func (s *Stmt) Finalize() error

func (c *Conn) Prepare(cmd string) (*Stmt, error) 

*/



/*
 Should only be used for sql statements that are not dynamically constructed.
 That is, use this only for a small set of completely predefined queries, which may nonetheless have ? parameters
 in them.
 Tests if the sql statement (cmd) is a key in a prepared statements map. If so, returns the already prepared
 statement. If not found in the map, asks the db connection to prepare the statement.
 Returns the prepared statement.

 TODO maybe should do stmt.Reset() if found in hashtable
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
Execute multiple sql statments, each possibly with arguments.
Return a non-nil error and do not continue to process statements, if 
a database error occurs.
*/
func (db *SqliteDB) ExecStatements(statementGroup *StatementGroup) (err error) {

	// Replace everything  with the statementQueue insertion below

	Logln(PERSIST_, fmt.Sprintf("Executing statement:\n%s\n", statementGroup))

	for _, sqlStatement := range statementGroup.Statements {
			
		err = db.ExecStatement(sqlStatement.Statement,sqlStatement.Args...)	
        if err != nil {
        	break
        }
	}
	return
}


/*
Execute a single SQL statement.
Return the db error if any.
*/
func (db *SqliteDB) ExecStatement(statement string, args ...interface{}) (err error) {

    Logln(PERSIST_, fmt.Sprintf("Executing statement:\n%s\n", statement))
		
    stmt, prepareErr := db.Prepare(statement)
    if prepareErr != nil {
	err = fmt.Errorf("DB ERROR preparing statement:\n%s\nDetail: %s\n\n", statement, prepareErr) 	   
	    return
    }

    defer stmt.Reset()

    err = stmt.Exec(args...)
    if err != nil {
       err = fmt.Errorf("DB ERROR executing statement:\n%s\nDetail: %s\n\n", statement, err) 	
	   return
    }   
    return
}


/*
OBSOLETE
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
