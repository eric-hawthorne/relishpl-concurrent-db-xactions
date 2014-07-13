// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package implements data persistence in the relish language environment.

package persist

/*
   connection_pool.go - pool of connections to sql databases

   Because each Connection caches prepared statements (client-side), we want to 
   re-use each existing Connection as much as possible, so that the most frequently used
   connections have the most cached prepared statements.

   Therefore the connection pool is a thread-safe LIFO stack.
*/


import (
   . "util/thread_safe_stack"
   "sync"
   "fmt"
)


type ConnectionPool struct {
	  maxConnections int
	  numConnections int
    unusedConnections *Stack
    connectionCreationMutex sync.Mutex
	  dbName string
    newConn ConnectionFactory
}



func NewConnectionPool(dbName string, maxConnections int, newConnFnc ConnectionFactory ) *ConnectionPool {
   pool := &ConnectionPool{maxConnections:maxConnections,
                           unusedConnections: NewStack(),
                           dbName: dbName,
                           newConn: newConnFnc,
                          }
   return pool
}


func (pool *ConnectionPool) GrabConnection() (conn Connection) {
    val := pool.unusedConnections.PopIf()
    if val != nil {
        conn = val.(Connection)
    } else {
        var err error
        conn, err = pool.createConnection()
        if err != nil {
            panic(fmt.Sprintf("Unable to open the database '%s': %s", pool.dbName, err))
        }
        if conn == nil {
            val = pool.unusedConnections.Pop()  // really have to wait this time
            conn = val.(Connection)
        }         
    }
    return
}


func (pool *ConnectionPool) ReleaseConnection(conn Connection) {
    pool.unusedConnections.Push(conn)
}


func (pool *ConnectionPool) createConnection() (conn Connection, err error) {
    pool.connectionCreationMutex.Lock() 
    defer pool.connectionCreationMutex.Unlock()        
    if pool.numConnections < pool.maxConnections {
       conn, err = pool.newConn(pool.dbName)
       if err == nil {
          pool.numConnections++
       }
   }
   return
}



