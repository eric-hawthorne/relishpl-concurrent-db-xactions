// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package implements data persistence in the relish language environment.

package persist

/*
   persist_collection.go - sqlite persistence of relish collections

	Specific methods of the database abstraction layer for persisting relish collections.


    `CREATE TABLE robject(
       id INTEGER PRIMARY KEY,
       id2 INTEGER, 
       idReversed BOOLEAN, --???
       typeName TEXT   -- Should be typeId because type should be another RObject!!!!
    )`

*/

import (
	"fmt"
	. "relish/dbg"
	. "relish/runtime/data"
   "io"
)

// 
// TODO !!! HAVE A TABLE called [rlist] [rset] [rsortedset] etc for independent collections.
//

/*
   Persist the adding of a value to a multi-valued attribute.
   Assumes that the the obj is already persisted, but does not assume that the value is.

  TODO create a map of prepared statements and look up the statement to use.
*/
func (db *SqliteDBThread) PersistAddToAttr(th InterpreterThread, obj RObject, attr *AttributeSpec, val RObject, insertIndex int) (err error) {

	table := db.db.TableNameIfy(attr.ShortName())

	if attr.Part.Type.IsPrimitive {


		
		valCols,valVars := attr.Part.Type.DbCollectionColumnInsert()
				
      switch attr.Part.CollectionType {
      case "set": // id,val
      
      
         stmt := Stmt(fmt.Sprintf("INSERT INTO %s(id,%s) VALUES(%v,%s)", table, valCols, obj.DBID(), valVars))		

         valParts := db.db.primitiveValSQL(val) 
         stmt.Args(valParts)						

         err = db.ExecStatements(stmt)      
  


      case "list": // id, val, ord1
      

			stmt := Stmt(fmt.Sprintf("INSERT INTO %s(id,%s,ord1) VALUES(%v,%s,%v)", table, valCols, obj.DBID(), valVars, insertIndex))
				
			valParts := db.db.primitiveValSQL(val) 
			stmt.Args(valParts)
				
			err = db.ExecStatements(stmt)



      case "sortedlist", "sortedset": //, "sortedmap": // id, val, ord1
      
   
      	err = db.ExecStatement(
            fmt.Sprintf("UPDATE %s SET ord1 = ord1 + 1 WHERE id0=? AND ord1 >= ?",table), 
            obj.DBID(), 
            insertIndex)
      	if err != nil {
            return
         }
      	stmt := Stmt(fmt.Sprintf("INSERT INTO %s(id,%s,ord1) VALUES(%v,%s,%v)", table, valCols, obj.DBID(), valVars, insertIndex))
		
      	valParts := db.db.primitiveValSQL(val) 
      	stmt.Args(valParts)
		
      	err = db.ExecStatements(stmt)      
      
      
      	//	     case "sortedstringmap":	// id0,id1,key1		
      }		

	} else { // Non-Primitive part type

		err = db.EnsurePersisted(th, val)
		if err != nil {
			return
		}

		switch attr.Part.CollectionType {
		case "set": // id0,id1
			err = db.ExecStatement(fmt.Sprintf("INSERT INTO %s(id0,id1) VALUES(?,?)", table), obj.DBID(), val.DBID()) 
          

		case "list": // id0, id1, ord1
			err = db.ExecStatement(fmt.Sprintf("INSERT INTO %s(id0,id1,ord1) VALUES(?,?,?)", table), obj.DBID(), val.DBID(), insertIndex)	
			//	     case "map": // id0, id1, ord1

			//	     case "stringmap": // id0,id1,key1

		case "sortedlist", "sortedset": //, "sortedmap": // id0, id1, ord1
			err = db.ExecStatement(fmt.Sprintf("UPDATE %s SET ord1 = ord1 + 1 WHERE id0=? AND ord1 >= ?",table), obj.DBID(), insertIndex)
			if err != nil {
            return
         }
         err = db.ExecStatement(fmt.Sprintf("INSERT INTO %s(id0,id1,ord1) VALUES(?,?,?)", table), obj.DBID(), val.DBID(), insertIndex) 
			//	     case "sortedstringmap":	// id0,id1,key1		
		}

		// stmt = fmt.Sprintf("UPDATE %s SET id1=%v WHERE id0=%v",table,obj.DBID(),val.DBID())   // Ensure DBID?                                       

	}
	return
}

/*
   Persist the removing of a value from a multi-valued attribute.
   Assumes that the the removal has happened from the in-memory collection.

   TODO create a map of prepared statements and look up the statement to use.
*/
func (db *SqliteDBThread) PersistRemoveFromAttr(obj RObject, attr *AttributeSpec, val RObject, removedIndex int) (err error) {     

	table := db.db.TableNameIfy(attr.ShortName())

	if attr.Part.Type.IsPrimitive {

		// TODO Have to handle different types, string, bool, int, float in different clauses
		
		if removedIndex == -1 {	
		   
		   sqlFragment := attr.Part.Type.DbCollectionRemove() 
      	stmt := Stmt(fmt.Sprintf("DELETE FROM %s WHERE id=? AND %s", table, sqlFragment))
		   stmt.Arg(obj.DBID()) 
      	valParts := db.db.primitiveValSQL(val) 
      	stmt.Args(valParts)
		
      	err = db.ExecStatements(stmt)			
			
		} else {
			err = db.ExecStatement(fmt.Sprintf("DELETE FROM %s WHERE id0=? AND id1=? AND ord1=?", table), obj.DBID(), val.DBID(), removedIndex)
			if err != nil {
            return
         }
         err = db.ExecStatement(fmt.Sprintf("UPDATE %s SET ord1 = ord1 - 1 WHERE id0=? AND ord1 > ?",  table), obj.DBID(), removedIndex)
		}
			
	} else { // Non-Primitive part type

		//	  fmt.Printf("id1 %v",val.DBID())	
		//	  fmt.Printf("removedIndex %v",removedIndex)

		if removedIndex == -1 {
			err = db.ExecStatement(fmt.Sprintf("DELETE FROM %s WHERE id0=? AND id1=?", table), obj.DBID(), val.DBID())
		} else {
			err = db.ExecStatement(fmt.Sprintf("DELETE FROM %s WHERE id0=? AND id1=? AND ord1=?", table), obj.DBID(), val.DBID(), removedIndex)
         if err != nil {
            return
         }			
         err = db.ExecStatement(fmt.Sprintf("UPDATE %s SET ord1 = ord1 - 1 WHERE id0=? AND ord1 > ?",  table), obj.DBID(), removedIndex)
		}
	}
	return

}

/*
   Persist the removing of all values from a multi-valued attribute.
   Assumes that the the removal has happened from the in-memory collection or that this will happen shortly.

   TODO create a map of prepared statements and look up the statement to use.
*/
func (db *SqliteDBThread) PersistClearAttr(obj RObject, attr *AttributeSpec) (err error) {

	table := db.db.TableNameIfy(attr.ShortName())

	if attr.Part.Type.IsPrimitive {

		// TODO Have to handle different types, string, bool, int, float in different clauses
		err = db.ExecStatement(fmt.Sprintf("DELETE FROM %s WHERE id=?", table), obj.DBID())

	} else { // Non-Primitive part type

		err = db.ExecStatement(fmt.Sprintf("DELETE FROM %s WHERE id0=?", table), obj.DBID())
      if err != nil {
         return
      }
		if attr.Inverse != nil {
		   inverseTable := db.db.TableNameIfy(attr.Inverse.ShortName())		
		   err = db.ExecStatement(fmt.Sprintf("; DELETE FROM %s WHERE id1=?", inverseTable), obj.DBID())
		}	
	}
	return

}







func (db *SqliteDBThread) PersistSetAttrElement(th InterpreterThread, obj RObject, attr *AttributeSpec, val RObject, index int) (err error) {

	table := db.db.TableNameIfy(attr.ShortName())
	
   if attr.Part.Type.IsPrimitive {

      valColSettings := attr.Part.Type.DbCollectionUpdate()   
    
 		stmt := Stmt(fmt.Sprintf("UPDATE %s SET %s WHERE id=? AND ord1=?", table, valColSettings))

 		valParts := db.db.primitiveValSQL(val) 
 		stmt.Args(valParts)
 		stmt.Arg(obj.DBID())
 		stmt.Arg(index)

 		err = db.ExecStatements(stmt)						

   } else { // non-primitive element values    

 		err = db.EnsurePersisted(th, val)
 		if err != nil {
 			return
 		}

 		stmt := Stmt(fmt.Sprintf("UPDATE %s SET id1=? WHERE id0=? AND ord1=?", table))
 		stmt.Arg(val.DBID())
 		stmt.Arg(obj.DBID())
 		stmt.Arg(index)		

 		err = db.ExecStatements(stmt)			

 	}	
	
   return   
}

      
func (db *SqliteDBThread) PersistSetCollectionElement(th InterpreterThread, coll IndexSettable, val RObject, index int) (err error) {

   table,_,_,_,elementType,err := db.EnsureCollectionTable(coll.(RCollection))
   if err != nil {
      return
   }
   
   valColSettings := elementType.DbCollectionUpdate()   
   
   if elementType.IsPrimitive {
				   
		stmt := Stmt(fmt.Sprintf("UPDATE %s SET %s WHERE id=? AND ord1=?", table, valColSettings))

		valParts := db.db.primitiveValSQL(val) 
		stmt.Args(valParts)
		stmt.Arg(coll.(RObject).DBID())
		stmt.Arg(index)

		err = db.ExecStatements(stmt)						
     
  } else { // non-primitive element values    

		err = db.EnsurePersisted(th, val)
		if err != nil {
			return
		}

		stmt := Stmt(fmt.Sprintf("UPDATE %s SET id1=? WHERE id0=? AND ord1=?", table))
		stmt.Arg(val.DBID())
		stmt.Arg(coll.(RObject).DBID())
		stmt.Arg(index)		
		
		err = db.ExecStatements(stmt)			

	} 
   
   return   
}


/*
NOTE NOTE NOTE We don't have persist remove from map yet !!!!! 
Does it remove by key? It should.
*/
func (db *SqliteDBThread) PersistMapPut(th InterpreterThread, theMap Map, key RObject,val RObject, isNewKey bool) (err error) {

   table,_,_,keyType,elementType,err := db.EnsureCollectionTable(theMap)
   if err != nil {
      return
   }	
		
	if elementType.IsPrimitive {		

   	if keyType == StringType {

        keyStr := SqlStringValueEscape(string(key.(String)))          
   	  
   	  if isNewKey {
   	      valCols,valVars := elementType.DbCollectionColumnInsert()   	     
   			
   			stmt := Stmt(fmt.Sprintf("INSERT INTO %s(id,%s,key1) VALUES(?,%s,?)", table, valCols, valVars)) 
		      stmt.Arg(theMap.DBID())
   		   valParts := db.db.primitiveValSQL(val) 
   		   stmt.Args(valParts)								
            stmt.Arg(keyStr)
      
   			err = db.ExecStatements(stmt)   	  
   	   
         } else {  // replacing value of an existing key
            
            valColSettings := elementType.DbCollectionUpdate()    
                       
      		stmt := Stmt(fmt.Sprintf("UPDATE %s SET %s WHERE id=? AND key1=?", table, valColSettings))

      		valParts := db.db.primitiveValSQL(val) 
      		stmt.Args(valParts)
      		stmt.Arg(theMap.DBID())
      		stmt.Arg(keyStr)        	

          	err = db.ExecStatements(stmt)         
         }   	   
		} else { // not a stringmap
			// TODO - We do not know if the key is persisted. We don't know if the key is an integer!!!
			// !!!!!!!!!!!!!!!!!!!!!!!!
			// !!!! NOT DONE YET !!!!!!
			// !!!!!!!!!!!!!!!!!!!!!!!!
			if isNewKey {
   	      valCols,valVars := elementType.DbCollectionColumnInsert()			   
			   
            stmt := Stmt(fmt.Sprintf("INSERT INTO %s(%s,id,ord1) VALUES(%s,?,?)", table, valCols, valVars))
	
   		   valParts := db.db.primitiveValSQL(val) 		   
   		   stmt.Args(valParts)
            stmt.Arg(theMap.DBID())		   
            switch keyType {
   	   	case UintType:
   		   	stmt.Arg(int64(uint64(key.(Uint))))   // val is actually the map key
   	   	case Uint32Type:
   		   	stmt.Arg(int(uint32(key.(Uint32))))   // val is actually the map key			   	
   	   	case IntType:
   		   	stmt.Arg(int64(key.(Int)))   // val is actually the map key
   	   	case Uint32Type:
   		   	stmt.Arg(int(key.(Int32)))   // val is actually the map key                         
            default:
               stmt.Arg(key.DBID())       						
 		      }
   			err = db.ExecStatements(stmt)
   			
         } else { // replacing value of an existing key
            
            valColSettings := elementType.DbCollectionUpdate() 
                 
         	stmt := Stmt(fmt.Sprintf("UPDATE %s SET %s WHERE id=? AND ord1=?", table, valColSettings))

            valParts := db.db.primitiveValSQL(val) 
            stmt.Args(valParts)
            
            stmt.Arg(theMap.DBID())
            
            switch keyType {
   	   	case UintType:
   		   	stmt.Arg(int64(uint64(key.(Uint))))   // val is actually the map key
   	   	case Uint32Type:
   		   	stmt.Arg(int(uint32(key.(Uint32))))   // val is actually the map key			   	
   	   	case IntType:
   		   	stmt.Arg(int64(key.(Int)))   // val is actually the map key
   	   	case Uint32Type:
   		   	stmt.Arg(int(key.(Int32)))   // val is actually the map key                         
            default:
               stmt.Arg(key.DBID())       						
 		      }                   	
         	err = db.ExecStatements(stmt)          
         }								
   	} 	
	} else { // non-primitive element type
      
		err = db.EnsurePersisted(th, val)
		if err != nil {
			return
		}      
      
      if keyType == StringType {
         
         keyStr := SqlStringValueEscape(string(key.(String)))   
           
         if isNewKey {      
   			stmt := Stmt(fmt.Sprintf("INSERT INTO %s(id0,id1,key1) VALUES(?,?,?)", table)) 
            stmt.Arg(theMap.DBID())
            stmt.Arg(val.DBID())
            stmt.Arg(keyStr)
   			err = db.ExecStatements(stmt)               
               
         } else { // replacing value of an existing key

      		stmt := Stmt(fmt.Sprintf("UPDATE %s SET id1=? WHERE id0=? AND key1=?", table))

      		stmt.Arg(val.DBID())
      		stmt.Arg(theMap.DBID())
      		stmt.Arg(keyStr)

         	err = db.ExecStatements(stmt)               
         }   
		} else {  // not a stringmap
				// TODO - We do not know if the key is persisted. We don't know if the key is an integer!!!
				// !!!!!!!!!!!!!!!!!!!!!!!!
				// !!!! NOT DONE YET !!!!!!
				// !!!!!!!!!!!!!!!!!!!!!!!!

         if isNewKey { 
            
            stmt := Stmt(fmt.Sprintf("INSERT INTO %s(id0,id1,ord1) VALUES(?,?,?)"))
            stmt.Arg(theMap.DBID())
            stmt.Arg(val.DBID())
            
            
   		   valParts := db.db.primitiveValSQL(val) 		   
   		   stmt.Args(valParts)
            stmt.Arg(theMap.DBID())		   
            switch keyType {
   	   	case UintType:
   		   	stmt.Arg(int64(uint64(key.(Uint))))   // val is actually the map key
   	   	case Uint32Type:
   		   	stmt.Arg(int(uint32(key.(Uint32))))   // val is actually the map key			   	
   	   	case IntType:
   		   	stmt.Arg(int64(key.(Int)))   // val is actually the map key
   	   	case Uint32Type:
   		   	stmt.Arg(int(key.(Int32)))   // val is actually the map key                         
            default:
               stmt.Arg(key.DBID())       						
		      }            	
   			err = db.ExecStatements(stmt)
			
         } else { // replacing value of an existing key
         
         	stmt := Stmt(fmt.Sprintf("UPDATE %s SET id1=? WHERE id0=? AND ord1=?", table))
            stmt.Arg(val.DBID())
            stmt.Arg(theMap.DBID())
            switch keyType {
   	   	case UintType:
   		   	stmt.Arg(int64(uint64(key.(Uint))))   // val is actually the map key
   	   	case Uint32Type:
   		   	stmt.Arg(int(uint32(key.(Uint32))))   // val is actually the map key			   	
   	   	case IntType:
   		   	stmt.Arg(int64(key.(Int)))   // val is actually the map key
   	   	case Uint32Type:
   		   	stmt.Arg(int(key.(Int32)))   // val is actually the map key                         
            default:
               stmt.Arg(key.DBID())       						
 		      }                    	
         	err = db.ExecStatements(stmt)                
         }				
		}
	} 
   return   
}

  
func (db *SqliteDBThread) PersistAddToCollection(th InterpreterThread, coll AddableCollection, val RObject, insertIndex int) (err error) {

   table,_,isOrdered,_,elementType,err := db.EnsureCollectionTable(coll)
   if err != nil {
      return
   }
   
	if elementType.IsPrimitive {
	
		valCols,valVars := elementType.DbCollectionColumnInsert()	
      
      if coll.IsSet() && ! isOrdered {      
         stmt := Stmt(fmt.Sprintf("INSERT INTO %s(id,%s) VALUES(?,%s)", table, valCols, valVars))		
         stmt.Arg(coll.DBID())
         valParts := db.db.primitiveValSQL(val) 
         stmt.Args(valParts)						

         err = db.ExecStatements(stmt)      
         
      } else if coll.IsList() && ! coll.IsSorting() { // id, val, ord1
      
			stmt := Stmt(fmt.Sprintf("INSERT INTO %s(id,%s,ord1) VALUES(?,%s,%v)", table, valCols, valVars))
         stmt.Arg(coll.DBID())				
			valParts := db.db.primitiveValSQL(val) 
			stmt.Args(valParts)
	      stmt.Arg(insertIndex)			
			err = db.ExecStatements(stmt)

      } else if (coll.IsList() || coll.IsSet()) && coll.IsSorting() { // id, val, ord1
   
      	err = db.ExecStatement(fmt.Sprintf("UPDATE %s SET ord1 = ord1 + 1 WHERE id0=? AND ord1 >= ?",table), coll.DBID(), insertIndex)
      	if err != nil {
            return
         }      
      	stmt := Stmt(fmt.Sprintf("INSERT INTO %s(id,%s,ord1) VALUES(?,%s,?)", table, valCols, valVars))
         stmt.Arg(coll.DBID())		
      	valParts := db.db.primitiveValSQL(val) 
      	stmt.Args(valParts)
         stmt.Arg(insertIndex)   		
      	err = db.ExecStatements(stmt)      	
      }		

	} else { // Non-Primitive part type

		err = db.EnsurePersisted(th, val)
		if err != nil {
			return
		}

      if coll.IsSet() && ! isOrdered {   // id0,id1
			err = db.ExecStatement(fmt.Sprintf("INSERT INTO %s(id0,id1) VALUES(?,?)", table), coll.DBID(), val.DBID())
          
      } else if coll.IsList() && ! coll.IsSorting() { // id0, id1, ord1

			err = db.ExecStatement(fmt.Sprintf("INSERT INTO %s(id0,id1,ord1) VALUES(?,?,?)", table), coll.DBID(), val.DBID(), insertIndex)
			//	     case "map": // id0, id1, ord1

			//	     case "stringmap": // id0,id1,key1

      } else if (coll.IsList() || coll.IsSet()) && coll.IsSorting() {  // id0, id1, ord1
			err = db.ExecStatement(fmt.Sprintf("UPDATE %s SET ord1 = ord1 + 1 WHERE id0=? AND ord1 >= ?",table), coll.DBID(), insertIndex)
         if err != nil {
            return
         }
			err = db.ExecStatement(fmt.Sprintf("INSERT INTO %s(id0,id1,ord1) VALUES(?,?,?)", table), coll.DBID(), val.DBID(), insertIndex)
			//	     case "sortedstringmap":	// id0,id1,key1		
		}
		// stmt = fmt.Sprintf("UPDATE %s SET id1=%v WHERE id0=%v",table,obj.DBID(),val.DBID())   // Ensure DBID?                                       
	}   
   return   
}


// TODO StringMaps !!
//
func (db *SqliteDBThread) PersistRemoveFromCollection(coll RemovableCollection, val RObject, removedIndex int) (err error) {

   table,isMap,_,keyType,elementType,err := db.EnsureCollectionTable(coll)
   if err != nil {
      return
   }
   
   if elementType.IsPrimitive {
      
	   if isMap { 
 		   stmt := Stmt(fmt.Sprintf("DELETE FROM %s WHERE id=? AND ord1=?", table))	      
	      switch keyType {  
	      case StringType:
  		   	stmt = Stmt(fmt.Sprintf("DELETE FROM %s WHERE id=? AND key1=?", table))
  		      keyStr := SqlStringValueEscape(string(val.(String)))   
  		   	stmt.Arg(coll.DBID())
  		   	stmt.Arg(keyStr)
	   	case UintType:
		   	stmt.Arg(coll.DBID())
		   	stmt.Arg(int64(uint64(val.(Uint))))   // val is actually the map key
	   	case Uint32Type:
		   	stmt.Arg(coll.DBID())
		   	stmt.Arg(int(uint32(val.(Uint32))))   // val is actually the map key			   	
	   	case IntType:
		   	stmt.Arg(coll.DBID())
		   	stmt.Arg(int64(val.(Int)))   // val is actually the map key
	   	case Uint32Type:
		   	stmt.Arg(coll.DBID())
		   	stmt.Arg(int(val.(Int32)))   // val is actually the map key		   	   		          
         default:
		   	stmt.Arg(coll.DBID())
		   	stmt.Arg(val.DBID())   // val is actually the map key 
         }
		   err = db.ExecStatements(stmt) 		         
	   } else if removedIndex == -1 {	

		   sqlFragment := elementType.DbCollectionRemove() 
      	stmt := Stmt(fmt.Sprintf("DELETE FROM %s WHERE id=%v AND %s", table, sqlFragment))
	      stmt.Arg(coll.DBID())
      	valParts := db.db.primitiveValSQL(val) 
      	stmt.Args(valParts)
	
      	err = db.ExecStatements(stmt)			
		
		} else {
			err = db.ExecStatement(fmt.Sprintf("DELETE FROM %s WHERE id0=? AND id1=? AND ord1=?", table), coll.DBID(), val.DBID(), removedIndex)
			if err != nil {
            return
         }
         err = db.ExecStatement(fmt.Sprintf("UPDATE %s SET ord1 = ord1 - 1 WHERE id0=? AND ord1 > ?",  table), coll.DBID(), removedIndex)
		}

	} else { // Non-Primitive element type

		//	  fmt.Printf("id1 %v",val.DBID())	
		//	  fmt.Printf("removedIndex %v",removedIndex)

	   if isMap {  
		  stmt := Stmt(fmt.Sprintf("DELETE FROM %s WHERE id0=? AND ord1=?", table)) 		     
	      switch keyType {  
	      case StringType:
		   	stmt = Stmt(fmt.Sprintf("DELETE FROM %s WHERE id0=? AND key1=?", table))  		   	
  		    keyStr := SqlStringValueEscape(string(val.(String)))   
  		   	stmt.Arg(coll.DBID())
  		   	stmt.Arg(keyStr)
	   	case UintType:
		   	stmt.Arg(coll.DBID())
		   	stmt.Arg(int64(uint64(val.(Uint))))   // val is actually the map key
	   	case Uint32Type:
		   	stmt.Arg(coll.DBID())
		   	stmt.Arg(int(uint32(val.(Uint32))))   // val is actually the map key			   	
	   	case IntType:
		   	stmt.Arg(coll.DBID())
		   	stmt.Arg(int64(val.(Int)))   // val is actually the map key
	   	case Uint32Type:
		   	stmt.Arg(coll.DBID())
		   	stmt.Arg(int(val.(Int32)))   // val is actually the map key		   	   		          
         default:
		   	stmt.Arg(coll.DBID())
		   	stmt.Arg(val.DBID())   // val is actually the map key 
         }
		   err = db.ExecStatements(stmt) 

	   } else if removedIndex == -1 {
			err = db.ExecStatement(fmt.Sprintf("DELETE FROM %s WHERE id0=? AND id1=?", table), coll.DBID(), val.DBID())
		} else {
			err = db.ExecStatement(fmt.Sprintf("DELETE FROM %s WHERE id0=? AND id1=? AND ord1=?", table), coll.DBID(), val.DBID(), removedIndex)
         if err != nil {
            return
         }			
         err = db.ExecStatement(fmt.Sprintf("UPDATE %s SET ord1 = ord1 - 1 WHERE id0=? AND ord1 > ?",  table), coll.DBID(), removedIndex)
		}
	}
   return   
}


func (db *SqliteDBThread) PersistClearCollection(coll RemovableCollection) (err error) {
	
   table,_,_,_,elementType,err := db.EnsureCollectionTable(coll)
   if err != nil {
      return
   }	

	if elementType.IsPrimitive {

		// TODO Have to handle different types, string, bool, int, float in different clauses
		err = db.ExecStatement(fmt.Sprintf("DELETE FROM %s WHERE id=?", table), coll.DBID())		

	} else { // Non-Primitive part type

		err = db.ExecStatement(fmt.Sprintf("DELETE FROM %s WHERE id0=?", table), coll.DBID())

	}
   return   
}














/*
JUST HERE FOR EXAMPLE
func (db *SqliteDB) fetchPrimitiveAttributeValues(id int64, obj RObject) (err os.Error) {
   defer Un(Trace(PERSIST_TR2,"fetchPrimitiveAttributeValues",id))	
	// TODO

   objTyp := obj.Type()

   // Create the select clause

   selectClause := "SELECT "
   sep := ""
   for _,attr := range objTyp.Attributes {
      if attr.Part.Type.IsPrimitive {	
         selectClause += sep + attr.Part.Name
         sep = ","
      }
   }	

   for _,typ := range objTyp.Up {
      for _,attr := range typ.Attributes {
         if attr.Part.Type.IsPrimitive {
	         selectClause += sep + attr.Part.Name
	         sep = ","        
		 }	
      }
   }

   // Figure out the total number of primitive attributes in all the types combined.

   numPrimitiveAttrs := objTyp.NumPrimitiveAttributes

   // Now for the object's type and all types in the upchain, we need to 
   // create a join statement on the type tables, then
   // collect the primitive attributes, in the order they were put into the type tables.  

   specificTypeTable := db.TableNameIfy(objTyp.ShortName())
   from := " FROM " + specificTypeTable

   for _,typ := range objTyp.Up {
      from += " JOIN " + db.TableNameIfy(typ.ShortName()) + " USING (id)"
      numPrimitiveAttrs += typ.NumPrimitiveAttributes
   }

   where := fmt.Sprintf(" WHERE %s.id=%v",specificTypeTable,id)

   stmt := selectClause + from + where

   Logln(PERSIST2_,"query:",stmt)

   selectStmt,err := db.conn.Prepare(stmt)
   if err != nil {
      return
   }

   defer selectStmt.Finalize()

   err = selectStmt.Exec()
   if err != nil {
      return
   }
   if ! selectStmt.Next() {
      panic(fmt.Sprintf("No object found in database with id=%v",id))
   }
*/

// Now construct a Scan call with [] *[]byte
/*
   attrValsBytes1 := make([]*[]byte,numPrimitiveAttrs)

   attrValsBytes := make([]interface{},numPrimitiveAttrs)

   for i := 0; i < len(attrValsBytes1); i++ {
      attrValsBytes[i] = attrValsBytes1[i]	
   }
*/

/*
   attrValsBytes1 := make([][]byte,numPrimitiveAttrs)

   attrValsBytes := make([]interface{},numPrimitiveAttrs)

   for i := 0; i < len(attrValsBytes1); i++ {
      attrValsBytes[i] = &attrValsBytes1[i]	
   }

   // have to make a slice whose length = the total number of attributes in all the types combined.
   // 
   err = selectStmt.Scan(attrValsBytes...)
   if err != nil {
      return
   }

   // Now go through the attrValsBytes and interpret each according to the datatype of each primitive
   // attribute, and set the primitive attributes using the runtime SetAttrVal method.

   i := 0
   var val RObject

   for _,attr := range objTyp.Attributes {
      if attr.Part.Type.IsPrimitive {
	     valByteSlice := *(attrValsBytes[i].(*[]byte))
         if convertAttrVal(valByteSlice, attr, &val) {
	        RT.RestoreAttr(obj, attr, val)
         }
	     i ++
      }	
   }

   for _,typ := range objTyp.Up {
      for _,attr := range typ.Attributes {
         if attr.Part.Type.IsPrimitive {
	        valByteSlice := *(attrValsBytes[i].(*[]byte))
		    if convertAttrVal(valByteSlice, attr, &val)	{
			   RT.RestoreAttr(obj, attr, val)
			}
			i ++
		 }	
      }
   }

   return 
}
0000000000000000000
*/

/////////////////////////////////////////
// FETCHING collections from the database
/////////////////////////////////////////

/*
Fetch attributes which are multi-valued and the collection is of non-primitive objects.
*/
func (db *SqliteDBThread) fetchMultiValuedNonPrimitiveAttributeValues(id int64, obj RObject, radius int) (err error) {
	defer Un(Trace(PERSIST_TR2, "fetchMultiValuedNonPrimitiveAttributeValues", id))

	objTyp := obj.Type()

	for _, attr := range objTyp.Attributes {
		if attr.Part.CollectionType != "" && !attr.Part.Type.IsPrimitive {
			err = db.fetchNonPrimitiveAttributeValueCollection(id, obj, attr, radius)
			if err != nil {
				return
			}
		}
	}

	for _, typ := range objTyp.Up {
		for _, attr := range typ.Attributes {
			if attr.Part.CollectionType != "" && !attr.Part.Type.IsPrimitive {
				err = db.fetchNonPrimitiveAttributeValueCollection(id, obj, attr, radius)
				if err != nil {
					return
				}
			}
		}
	}
	return
}

/*
Fetch from the db the values in a collection of a non-primitive-typed unary attribute of the object.
Set the attribute to the fetched collection.  
*/
func (db *SqliteDBThread) fetchNonPrimitiveAttributeValueCollection(objId int64, obj RObject, attr *AttributeSpec, radius int) (err error) {
	defer Un(Trace(PERSIST_TR2, "fetchNonPrimitiveAttributeValueCollection", objId, attr.ShortName()))

	// first, determine if the collection exists in memory already as the value of the attribute of the object.
	// If not, create it.

	collection, err := RT.EnsureMultiValuedAttributeCollection(obj, attr)
	if err != nil {
		return
	}

	err = db.fetchCollection(collection, objId, db.db.TableNameIfy(attr.ShortName()), radius)

	return
}

/*
Fetch attributea which are multi-valued but the values in the collection are primitives. 

TODO

*/
func (db *SqliteDBThread) fetchMultiValuedPrimitiveAttributeValues(id int64, obj RObject) (err error) {
	defer Un(Trace(PERSIST_TR2, "fetchMultiValuedPrimitiveAttributeValues", id))

	objTyp := obj.Type()

	for _, attr := range objTyp.Attributes {
		if attr.Part.CollectionType != "" && attr.Part.Type.IsPrimitive {
			err = db.fetchPrimitiveAttributeValueCollection(id, obj, attr)
			if err != nil {
				return
			}
		}
	}

	for _, typ := range objTyp.Up {
		for _, attr := range typ.Attributes {
			if attr.Part.CollectionType != "" && attr.Part.Type.IsPrimitive {
				err = db.fetchPrimitiveAttributeValueCollection(id, obj, attr)
				if err != nil {
					return
				}
			}
		}
	}
	return
}


/*
Fetch from the db the values in a collection of a non-primitive-typed unary attribute of the object.
Set the attribute to the fetched collection.  
*/
func (db *SqliteDBThread) fetchPrimitiveAttributeValueCollection(objId int64, obj RObject, attr *AttributeSpec) (err error) {
	defer Un(Trace(PERSIST_TR2, "fetchPrimitiveAttributeValueCollection", objId, attr.ShortName()))

	// first, determine if the collection exists in memory already as the value of the attribute of the object.
	// If not, create it.

	collection, err := RT.EnsureMultiValuedAttributeCollection(obj, attr)
	if err != nil {
		return
	}

	err = db.fetchPrimitiveValueCollection(collection, objId, db.db.TableNameIfy(attr.ShortName()))

	return
}


/*
   This is a lower-level function.

   Usable for rlist,rset,rsortedset - not for maps

   Populates a collection with members fetched from the database.
   If the collection has members already, they are removed from the in-memory collection. (The collection is emptied)
   Then the collection members are fetched afresh from the database.
   If radius is zero, fetches proxies into the collection.
   Otherwise, fetches proper objects.

   The collectionTableName is either [rlist],[rset],[rsortedset] or e.g. [Car___wheels__Wheel]

   Usage: db.fetchCollection(collection, collectionOrOwnerId, collectionTableName, radius)
*/
func (db *SqliteDBThread) fetchCollection(collection RCollection, collectionOrOwnerId int64, collectionTableName string, radius int) (err error) {
	defer Un(Trace(PERSIST_TR2, "fetchCollection", collectionOrOwnerId, collectionTableName))

	remColl := collection.(RemovableMixin)
	remColl.ClearInMemory()

	orderClause := ""
	if collection.IsOrdered() {
		orderClause = " ORDER BY ord1"
	}

   if collection.IsMap() {
      theMap := collection.(Map)
      if theMap.KeyType() == StringType {
   	   if collection.IsOrdered() {
   		   orderClause = " ORDER BY key1"
      	}         
      	query := fmt.Sprintf("SELECT id1,key1 FROM %s WHERE id0=?%s", collectionTableName, orderClause)
   		
      	selectStmt, queryErr := db.Prepare(query)
      	if queryErr != nil {
   			err = queryErr      	   
      		return
      	}

      	defer selectStmt.Reset() // Ensure statement does not remain open

      	err = selectStmt.Query(collectionOrOwnerId)
      	if err != nil && err != io.EOF {
      		return
      	}

         collection.SetMayContainProxies(radius <= 0) 

      	var val RObject
      	var key RObject
   		var id1 int64
   		var keyStr string
   		
      	// for selectStmt.Next() {

         for ; err == nil ; err = selectStmt.Next() {

      		err = selectStmt.Scan(&id1,&keyStr)
      		if err != nil {
      			return
      		}
      		key = String(keyStr)      		
      		if radius > 0 { // fetch the full objects
      			val, err = db.Fetch(id1, radius-1)
      			if err != nil {
      				return
      			}
      		} else { // Just put proxy objects into the collection.
      			val = Proxy(id1)
      		}
      		
	         theMap.PutSimple(key, val)     		
      	}
         if err == io.EOF {
            err = nil 
         } else {
            err = fmt.Errorf("DB ERROR on query:\n%s\nDetail: %s\n\n", query, err)   
            return  
         }         
              
         
      } else {  // An object-keyed map
      	query := fmt.Sprintf("SELECT id1,ord1 FROM %s WHERE id0=?%s", collectionTableName, orderClause)         
   
      	selectStmt, queryErr := db.Prepare(query)
      	if queryErr != nil {
   			err = queryErr      	   
      		return
      	}

         defer selectStmt.Reset() // Ensure statement does not remain open

      	err = selectStmt.Query(collectionOrOwnerId)
         if err != nil && err != io.EOF {
            return
         }


         collection.SetMayContainProxies(radius <= 0) 

      	var val RObject
      	var key RObject
   		var id1 int64
   		var ord1 int64

         for ; err == nil ; err = selectStmt.Next() {

      		err = selectStmt.Scan(&id1,&ord1)
      		if err != nil {
      			return
      		}
      		
   			key, err = db.Fetch(ord1, 1) // Should this just be a proxy?
   			if err != nil {
   				return
   			}      		
      		  		
      		if radius > 0 { // fetch the full objects
      			val, err = db.Fetch(id1, radius-1)
      			if err != nil {
      				return
      			}
      		} else { // Just put proxy objects into the collection.
      			val = Proxy(id1)
      		}
      		
	         theMap.PutSimple(key, val)     		
      	}   
         if err == io.EOF {
            err = nil 
         } else {
            err = fmt.Errorf("DB ERROR on query:\n%s\nDetail: %s\n\n", query, err)   
            return  
         } 
      }
      
   } else { // Not a map
      
   	query := fmt.Sprintf("SELECT id1 FROM %s WHERE id0=?%s", collectionTableName, orderClause)

   	selectStmt, queryErr := db.Prepare(query)
   	if queryErr != nil {
			err = queryErr      	   
   		return
   	}

      defer selectStmt.Reset() // Ensure statement does not remain open

   	err = selectStmt.Query(collectionOrOwnerId)
      if err != nil && err != io.EOF {
         return
      }

       collection.SetMayContainProxies(radius <= 0) 

   	var val RObject

      for ; err == nil; err = selectStmt.Next() {
   		var id1 int64
   		err = selectStmt.Scan(&id1)
   		if err != nil {
   			return
   		}
   		if radius > 0 { // fetch the full objects
   			val, err = db.Fetch(id1, radius-1)
   			if err != nil {
   				return
   			}
   		} else { // Just put proxy objects into the collection.
   			val = Proxy(id1)
   		}
   		addColl := collection.(AddableMixin)
   		addColl.AddSimple(val)
   	}

      if err == io.EOF {
         err = nil 
      } else {
         err = fmt.Errorf("DB ERROR on query:\n%s\nDetail: %s\n\n", query, err)   
         return  
      } 
	}
	return
}


func (db *SqliteDBThread) fetchPrimitiveValueCollection(collection RCollection, collectionOrOwnerId int64, collectionTableName string) (err error) {
	defer Un(Trace(PERSIST_TR2, "fetchPrimitiveValueCollection", collectionOrOwnerId, collectionTableName))

	remColl := collection.(RemovableMixin)
	remColl.ClearInMemory()

	orderClause := ""
	if collection.IsOrdered() {
		orderClause = " ORDER BY ord1"
	}
	

   if collection.IsMap() {
      theMap := collection.(Map)
   	typ := theMap.ValType()

   	valCols,_ := typ.DbCollectionColumnInsert()
   	      
      if theMap.KeyType() == StringType {
   	   if collection.IsOrdered() {
   		   orderClause = " ORDER BY key1"
      	}         
      	
   	   query := fmt.Sprintf("SELECT %s,key1 FROM %s WHERE id=?%s", valCols, collectionTableName, orderClause)
      	
      	selectStmt, queryErr := db.Prepare(query)
      	if queryErr != nil {
   			err = queryErr      	   
      		return
      	}

         defer selectStmt.Reset() // Ensure statement does not remain open

         err = selectStmt.Query(collectionOrOwnerId)
         if err != nil && err != io.EOF {
            return
         }         

         collection.SetMayContainProxies(false) 

      	var val RObject
      	var keyStr string
      	var key RObject      	   	
      	
         var numColumns int
         switch typ {
         case ComplexType,Complex32Type,TimeType:
            numColumns = 2
         default: 
            numColumns = 1
         }
      	valsBytes1 := make([][]byte, numColumns)

      	valsBytes := make([]interface{}, numColumns+1)

      	for i := 0; i < len(valsBytes1); i++ {
      		valsBytes[i] = &valsBytes1[i]
      	}
      	valsBytes[numColumns] = &keyStr

         var nonNil bool

         for ; err == nil; err = selectStmt.Next() {

         	err = selectStmt.Scan(valsBytes...)
         	if err != nil {
         		return
         	}	   
      		valByteSlice := *(valsBytes[0].(*[]byte))   	
         	if numColumns == 1 {
      			nonNil = convertVal(valByteSlice, typ,"collection element val", &val)  
      			if ! nonNil { panic("nil not valid element in a primitive value collection") }  	   
   	   
      	   } else { // 2
      			valByteSlice2 := *(valsBytes[1].(*[]byte))	   
               nonNil = convertValTwoFields(valByteSlice, valByteSlice2, typ,"collection element val", &val) 	
      			if ! nonNil { panic("nil not valid element in a primitive value collection") }          			   
      	   }      	

      		key = String(keyStr)            	
      		
/*     		
      	for selectStmt.Next() {
      	      
      		var id1 int64
      		var keyStr string
      		err = selectStmt.Scan(&id1,&keyStr)
      		if err != nil {
      			return
      		}
      		key := String(keyStr)      		
      		if radius > 0 { // fetch the full objects
      			val, err = db.Fetch(id1, radius-1)
      			if err != nil {
      				return
      			}
      		} else { // Just put proxy objects into the collection.
      			val = Proxy(id1)
      		}
*/      				
      		
	         theMap.PutSimple(key, val)     		
      	}
         if err == io.EOF {
            err = nil 
         } else {
            err = fmt.Errorf("DB ERROR on query:\n%s\nDetail: %s\n\n", query, err)   
            return  
         } 



              
      } else if theMap.KeyType() == IntType || theMap.KeyType() == UintType {  // An integer keyed map

         query := fmt.Sprintf("SELECT %s,ord1 FROM %s WHERE id=?%s", valCols, collectionTableName, orderClause)         
   
         selectStmt, queryErr := db.Prepare(query)
         if queryErr != nil {
            err = queryErr          
            return
         }

         defer selectStmt.Reset() // Ensure statement does not remain open

         err = selectStmt.Query(collectionOrOwnerId)
         if err != nil && err != io.EOF {
            return
         } 

         collection.SetMayContainProxies(false) 

         var val RObject
         var key RObject   
         var ord1 int64             
         
         var numColumns int
         switch typ {
         case ComplexType,Complex32Type,TimeType:
            numColumns = 2
         default: 
            numColumns = 1
         }
         valsBytes1 := make([][]byte, numColumns)

         valsBytes := make([]interface{}, numColumns+1)

         for i := 0; i < len(valsBytes1); i++ {
            valsBytes[i] = &valsBytes1[i]
         }
         valsBytes[numColumns] = &ord1

         var nonNil bool

         for ; err == nil; err = selectStmt.Next() {         

            err = selectStmt.Scan(valsBytes...)
            if err != nil {
               return
            }     
            valByteSlice := *(valsBytes[0].(*[]byte))    
            if numColumns == 1 {
               nonNil = convertVal(valByteSlice, typ,"collection element val", &val)  
               if ! nonNil { panic("nil not valid element in a primitive value collection") }      
         
            } else { // 2
               valByteSlice2 := *(valsBytes[1].(*[]byte))      
               nonNil = convertValTwoFields(valByteSlice, valByteSlice2, typ,"collection element val", &val)   
               if ! nonNil { panic("nil not valid element in a primitive value collection") }                     
            }        
         
            if theMap.KeyType() == IntType {
               key = Int(ord1)
            } else if theMap.KeyType() == UintType {
               key = Uint(uint64(ord1))               
            }
/*       

         for selectStmt.Next() {
            var id1 int64
            var ord1 int64
            err = selectStmt.Scan(&id1,&ord1)
            if err != nil {
               return
            }
            
            key, err = db.Fetch(ord1, 1) // Should this just be a proxy?
            if err != nil {
               return
            }           
                  
            if radius > 0 { // fetch the full objects
               val, err = db.Fetch(id1, radius-1)
               if err != nil {
                  return
               }
            } else { // Just put proxy objects into the collection.
               val = Proxy(id1)
            }
*/          
            
            theMap.PutSimple(key, val)          
         }       

         if err == io.EOF {
            err = nil 
         } else {
            err = fmt.Errorf("DB ERROR on query:\n%s\nDetail: %s\n\n", query, err)   
            return  
         } 


      } else {  // An object-keyed map
         

      	query := fmt.Sprintf("SELECT %s,ord1 FROM %s WHERE id0=?%s", valCols, collectionTableName, orderClause)         
   
      	selectStmt, queryErr := db.Prepare(query)
      	if queryErr != nil {
   			err = queryErr      	   
      		return
      	}

      	defer selectStmt.Reset() // Ensure statement does not remain open

         err = selectStmt.Query(collectionOrOwnerId)
         if err != nil && err != io.EOF {
            return
         } 

         collection.SetMayContainProxies(false) 

      	var val RObject
      	var key RObject   
      	var ord1 int64      	   	
      	
         var numColumns int
         switch typ {
         case ComplexType,Complex32Type,TimeType:
            numColumns = 2
         default: 
            numColumns = 1
         }
      	valsBytes1 := make([][]byte, numColumns)

      	valsBytes := make([]interface{}, numColumns+1)

      	for i := 0; i < len(valsBytes1); i++ {
      		valsBytes[i] = &valsBytes1[i]
      	}
      	valsBytes[numColumns] = &ord1

         var nonNil bool

         for ; err == nil; err = selectStmt.Next() {   

         	err = selectStmt.Scan(valsBytes...)
         	if err != nil {
         		return
         	}	   
      		valByteSlice := *(valsBytes[0].(*[]byte))   	
         	if numColumns == 1 {
      			nonNil = convertVal(valByteSlice, typ,"collection element val", &val)  
      			if ! nonNil { panic("nil not valid element in a primitive value collection") }  	   
   	   
      	   } else { // 2
      			valByteSlice2 := *(valsBytes[1].(*[]byte))	   
               nonNil = convertValTwoFields(valByteSlice, valByteSlice2, typ,"collection element val", &val) 	
      			if ! nonNil { panic("nil not valid element in a primitive value collection") }          			   
      	   }      	
      	
   			key, err = db.Fetch(ord1, 1) // Should this just be a proxy?
   			if err != nil {
   				return
   			}      	
	
/*      	

      	for selectStmt.Next() {
      		var id1 int64
      		var ord1 int64
      		err = selectStmt.Scan(&id1,&ord1)
      		if err != nil {
      			return
      		}
      		
   			key, err = db.Fetch(ord1, 1) // Should this just be a proxy?
   			if err != nil {
   				return
   			}      		
      		  		
      		if radius > 0 { // fetch the full objects
      			val, err = db.Fetch(id1, radius-1)
      			if err != nil {
      				return
      			}
      		} else { // Just put proxy objects into the collection.
      			val = Proxy(id1)
      		}
*/      		
      		
	         theMap.PutSimple(key, val)     		
      	}   
         if err == io.EOF {
            err = nil 
         } else {
            err = fmt.Errorf("DB ERROR on query:\n%s\nDetail: %s\n\n", query, err)   
            return  
         } 


      }
   } else { // not a map

   	typ := collection.ElementType()
	
   	valCols,_ := typ.DbCollectionColumnInsert()	

   	query := fmt.Sprintf("SELECT %s FROM %s WHERE id=?%s", valCols, collectionTableName, orderClause)

   	selectStmt, queryErr := db.Prepare(query)
   	if queryErr != nil {
			err = queryErr      	   
   		return
   	}

   	defer selectStmt.Reset() // Ensure statement does not remain open

      err = selectStmt.Query(collectionOrOwnerId)
      if err != nil && err != io.EOF {
         return
      } 

   	var val RObject
      var numColumns int
      switch typ {
      case ComplexType,Complex32Type,TimeType:
         numColumns = 2
      default: 
         numColumns = 1
      }
   	valsBytes1 := make([][]byte, numColumns)

   	valsBytes := make([]interface{}, numColumns)

   	for i := 0; i < len(valsBytes1); i++ {
   		valsBytes[i] = &valsBytes1[i]
   	}

      var nonNil bool

      for ; err == nil; err = selectStmt.Next() {   

      	err = selectStmt.Scan(valsBytes...)
      	if err != nil {
      		return
      	}	   
   		valByteSlice := *(valsBytes[0].(*[]byte))   	
      	if numColumns == 1 {
   			nonNil = convertVal(valByteSlice, typ,"collection element val", &val)  
   			if ! nonNil { panic("nil not valid element in a primitive value collection") }  	   
   	   
   	   } else { // 2
   			valByteSlice2 := *(valsBytes[1].(*[]byte))	   
            nonNil = convertValTwoFields(valByteSlice, valByteSlice2, typ,"collection element val", &val) 	
   			if ! nonNil { panic("nil not valid element in a primitive value collection") }          			   
   	   }
	   
   		addColl := collection.(AddableMixin)
   		addColl.AddSimple(val)
   	}

      if err == io.EOF {
         err = nil 
      } else {
         err = fmt.Errorf("DB ERROR on query:\n%s\nDetail: %s\n\n", query, err)   
         return  
      } 

   }
	return
}
