// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package implements data persistence in the relish language environment.

package persist

/*
    sqlite persistence of relish collections

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
)

// 
// TODO !!! HAVE A TABLE called [rlist] [rset] [rsortedset] etc for independent collections.
//

/*
   Persist the adding of a value to a multi-valued attribute.
   Assumes that the the obj is already persisted, but does not assume that the value is.

  TODO create a map of prepared statements and look up the statement to use.
*/
func (db *SqliteDB) PersistAddToAttr(obj RObject, attr *AttributeSpec, val RObject, insertIndex int) (err error) {

	table := db.TableNameIfy(attr.ShortName())

	if attr.Part.Type.IsPrimitive {

		// TODO Have to handle different types, string, bool, int, float in different clauses

	} else { // Non-Primitive part type

		err = db.EnsurePersisted(val)
		if err != nil {
			return
		}

		switch attr.Part.CollectionType {
		case "set": // id0,id1
			db.QueueStatement(fmt.Sprintf("INSERT INTO %s(id0,id1) VALUES(%v,%v)", table, obj.DBID(), val.DBID())) 
          

		case "list": // id0, id1, ord1
			db.QueueStatement(fmt.Sprintf("INSERT INTO %s(id0,id1,ord1) VALUES(%v,%v,%v)", table, obj.DBID(), val.DBID(), insertIndex))	
			//	     case "map": // id0, id1, ord1

			//	     case "stringmap": // id0,id1,key1

		case "sortedlist", "sortedset": //, "sortedmap": // id0, id1, ord1
			db.QueueStatement(fmt.Sprintf("UPDATE %s SET ord1 = ord1 + 1 WHERE id0=%v AND ord1 >= %v",table, obj.DBID(), insertIndex))
			db.QueueStatement(fmt.Sprintf("INSERT INTO %s(id0,id1,ord1) VALUES(%v,%v,%v)", table, obj.DBID(), val.DBID(), insertIndex)) 
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
func (db *SqliteDB) PersistRemoveFromAttr(obj RObject, attr *AttributeSpec, val RObject, removedIndex int) (err error) {     

	table := db.TableNameIfy(attr.ShortName())

	if attr.Part.Type.IsPrimitive {

		// TODO Have to handle different types, string, bool, int, float in different clauses

	} else { // Non-Primitive part type

		//	  fmt.Printf("id1 %v",val.DBID())	
		//	  fmt.Printf("removedIndex %v",removedIndex)

		if removedIndex == -1 {
			db.QueueStatement(fmt.Sprintf("DELETE FROM %s WHERE id0=%v AND id1=%v", table, obj.DBID(), val.DBID()))
		} else {
			db.QueueStatement(fmt.Sprintf("DELETE FROM %s WHERE id0=%v AND id1=%v AND ord1=%v", table, obj.DBID(), val.DBID(), removedIndex))
			db.QueueStatement(fmt.Sprintf("UPDATE %s SET ord1 = ord1 - 1 WHERE id0=%v AND ord1 > %v",  table, obj.DBID(), removedIndex))
		}
	}
	return

}

/*
   Persist the removing of all values from a multi-valued attribute.
   Assumes that the the removal has happened from the in-memory collection or that this will happen shortly.

   TODO create a map of prepared statements and look up the statement to use.
*/
func (db *SqliteDB) PersistClearAttr(obj RObject, attr *AttributeSpec) (err error) {

	table := db.TableNameIfy(attr.ShortName())

	if attr.Part.Type.IsPrimitive {

		// TODO Have to handle different types, string, bool, int, float in different clauses

	} else { // Non-Primitive part type

		db.QueueStatement(fmt.Sprintf("DELETE FROM %s WHERE id0=%v", table, obj.DBID()))
		if attr.Inverse != nil {
		   inverseTable := db.TableNameIfy(attr.Inverse.ShortName())		
		   db.QueueStatement(fmt.Sprintf("; DELETE FROM %s WHERE id1=%v", inverseTable, obj.DBID()))
		}	
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
func (db *SqliteDB) fetchMultiValuedNonPrimitiveAttributeValues(id int64, obj RObject, radius int) (err error) {
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
func (db *SqliteDB) fetchNonPrimitiveAttributeValueCollection(objId int64, obj RObject, attr *AttributeSpec, radius int) (err error) {
	defer Un(Trace(PERSIST_TR2, "fetchUnaryNonPrimitiveAttributeValue", objId, attr.ShortName()))

	// first, determine if the collection exists in memory already as the value of the attribute of the object.
	// If not, create it.

	collection, err := RT.EnsureMultiValuedAttributeCollection(obj, attr)
	if err != nil {
		return
	}

	err = db.fetchCollection(collection, objId, db.TableNameIfy(attr.ShortName()), radius)

	return
}

/*
Fetch attributea which are multi-valued but the values in the collection are primitives. 

TODO

*/
func (db *SqliteDB) fetchMultiValuedPrimitiveAttributeValues(id int64, obj RObject, radius int) (err error) {
	defer Un(Trace(PERSIST_TR2, "fetchMultiValuedPrimitiveAttributeValues", id))
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
func (db *SqliteDB) fetchCollection(collection RCollection, collectionOrOwnerId int64, collectionTableName string, radius int) (err error) {
	defer Un(Trace(PERSIST_TR2, "fetchCollection", collectionOrOwnerId, collectionTableName))

	remColl := collection.(RemovableMixin)
	remColl.ClearInMemory()

	orderClause := ""
	if collection.IsOrdered() {
		orderClause = " ORDER BY ord1"
	}

	query := fmt.Sprintf("SELECT id1 FROM %s WHERE id0=%v%s", collectionTableName, collectionOrOwnerId, orderClause)

	selectStmt, err := db.conn.Prepare(query)
	if err != nil {
		return
	}

	defer selectStmt.Finalize()

	err = selectStmt.Exec()
	if err != nil {
		return
	}

    collection.SetMayContainProxies(radius <= 0) 

	var val RObject

	for selectStmt.Next() {
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
	return
}
