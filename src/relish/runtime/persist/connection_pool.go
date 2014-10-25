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
   . "relish/runtime/data"
   "sync"
   "fmt"
   . "relish/dbg"
)


type ConnectionPool struct {
	  maxConnections int
	  numConnections int
    maxWriteConnections int  // if -1, means read and write connections use same pool.
    numWriteConnections int
    unusedConnections *Stack
    unusedWriteConnections *Stack
    connectionCreationMutex sync.Mutex
	  dbName string
    newConn ConnectionFactory
}



func NewConnectionPool(dbName string, maxConnections int, maxWriteConnections int, newConnFnc ConnectionFactory ) (pool *ConnectionPool) {

   if maxWriteConnections == -1 {
      pool = &ConnectionPool{maxConnections:maxConnections,
                            maxWriteConnections: maxWriteConnections,
                            unusedConnections: NewStack(),
                            dbName: dbName,
                            newConn: newConnFnc,
                          }
   } else {
      pool = &ConnectionPool{maxConnections:maxConnections,
                            maxWriteConnections: maxWriteConnections,
                            unusedConnections: NewStack(),
                            unusedWriteConnections: NewStack(),
                            dbName: dbName,
                            newConn: newConnFnc,
                          } 
   }
   return
}


func (pool *ConnectionPool) GrabConnection(doingWrite bool) (conn Connection) {
   if doingWrite && pool.maxWriteConnections != -1 {
     
      val := pool.unusedWriteConnections.PopIf()
      if val != nil {
          conn = val.(Connection)
      } else {
          var err error
          conn, err = pool.createConnection(doingWrite)
          if err != nil {
              panic(fmt.Sprintf("Unable to open the database '%s': %s", pool.dbName, err))
          }
          if conn == nil {
              val = pool.unusedWriteConnections.Pop()  // really have to wait this time
              conn = val.(Connection)
          } else {
              Log(PERSIST2_, "Created write connection %d\n",conn.Id())
          }        
      }    
      Log(PERSIST2_, "Grabbed write connection %d\n",conn.Id())          

       return
   }


    val := pool.unusedConnections.PopIf()
    if val != nil {
        conn = val.(Connection)
    } else {
        var err error
        conn, err = pool.createConnection(doingWrite)
        if err != nil {
            panic(fmt.Sprintf("Unable to open the database '%s': %s", pool.dbName, err))
        }
        if conn == nil {
            val = pool.unusedConnections.Pop()  // really have to wait this time
            conn = val.(Connection)
        } else {
            Log(PERSIST2_, "Created connection %d\n",conn.Id())
        }        
    }    
    Log(PERSIST2_, "Grabbed connection %d\n",conn.Id())     
    return
}


func (pool *ConnectionPool) ReleaseConnection(conn Connection) {
    if conn.IsReadOnly() || pool.maxWriteConnections == -1 {
       Log(PERSIST2_, "Released connection %d\n",conn.Id())  
       pool.unusedConnections.Push(conn)
    } else {
       Log(PERSIST2_, "Released write connection %d\n",conn.Id())  
       pool.unusedWriteConnections.Push(conn)      
    }
}


func (pool *ConnectionPool) createConnection(doingWrite bool) (conn Connection, err error) {
    pool.connectionCreationMutex.Lock() 
    defer pool.connectionCreationMutex.Unlock()  

    if pool.maxWriteConnections != -1 && doingWrite {

        if pool.numWriteConnections < pool.maxWriteConnections {
           conn, err = pool.newConn(pool.dbName, pool.numWriteConnections + 1)
           if err == nil {
              pool.numWriteConnections++
              conn.SetReadOnly(false)              
           }

        }
        return 
    }

    if pool.numConnections < pool.maxConnections {
       conn, err = pool.newConn(pool.dbName, pool.numConnections + 1)
       if err == nil {
           pool.numConnections++
           if pool.maxWriteConnections != -1 {
               conn.SetReadOnly(true)
           }          
       }
   }


 
   return
}





