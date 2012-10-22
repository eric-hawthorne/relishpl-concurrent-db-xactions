// Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package implements data persistence in the relish language environment.

package persist

/*
    sqlite persistence of relish objects

	Specific methods of the database abstraction layer for persisting relish objects and attribute assignments.


    `CREATE TABLE robject(
       id INTEGER PRIMARY KEY,
       id2 INTEGER, 
       idReversed BOOLEAN, --???
       typeName TEXT   -- Should be typeId because type should be another RObject!!!!
    )`

    Note that collection and mult-valued attribute persistence are handled by methods in persist_collection.go
*/

import (
	"errors"
	"fmt"
	. "relish/dbg"
	. "relish/runtime/data"
	"relish/rterr"
	"strconv"
	"strings"
	"time"
)

const TIME_LAYOUT = "2006-01-02 15:04:05.000"

/*
   Persist the setting of an attribute to a value.
   Only applies to single-valued attributes.
   Assumes that the the obj is already persisted, but does not assume that the value is.
*/
func (db *SqliteDB) PersistSetAttr(obj RObject, attr *AttributeSpec, val RObject, attrHadValue bool) (err error) {

	var stmt string

	if attr.Part.Type.IsPrimitive {

		table := db.TableNameIfy(attr.WholeType.ShortName())

		if val.Type() == StringType {
			stmt = fmt.Sprintf("UPDATE %s SET %s='%v' WHERE id=%v", table, attr.Part.Name, val, obj.DBID()) 
		} else if val.Type() == TimeType {
			attrName := attr.Part.Name
			attrLocName := attrName + "_loc"
			t := time.Time(val.(RTime))
			timeString := t.UTC().Format(TIME_LAYOUT)
			locationName := t.Location().String()
			stmt = fmt.Sprintf("UPDATE %s SET %s='%s', %s='%s' WHERE id=%v", table, attrName, timeString, attrLocName, locationName, obj.DBID()) 					
		} else {
			stmt = fmt.Sprintf("UPDATE %s SET %s=%v WHERE id=%v", table, attr.Part.Name, val, obj.DBID())
		}
	} else { // non-primitive value type

		err = db.EnsurePersisted(val)
		if err != nil {
			return
		}

		table := db.TableNameIfy(attr.ShortName())

		if attrHadValue {

			// TODO create a map of prepared statements and look up the statement to use.

			stmt = fmt.Sprintf("UPDATE %s SET id1=%v WHERE id0=%v", table, val.DBID(), obj.DBID()) // Ensure DBID?                                       
		} else {
			stmt = fmt.Sprintf("INSERT INTO %s(id0,id1) VALUES(%v,%v)", table, obj.DBID(), val.DBID()) // Ensure DBID?   
		}
	}
	db.QueueStatements(stmt)
	return

}


/*
   Persist the setting of an attribute to nil.
   Only applies to single-valued attributes.
   Assumes that the the obj is already persisted.
   Ok to call this even if there is no value of the attribute for the object.
*/
func (db *SqliteDB) PersistRemoveAttr(obj RObject, attr *AttributeSpec) (err error) {

	var stmt string

	if attr.Part.Type.IsPrimitive {

		table := db.TableNameIfy(attr.WholeType.ShortName())
		
		if attr.Part.Type == TimeType {
		   attrName := attr.Part.Name
		   attrLocName := attrName + "_loc"			
		   stmt = fmt.Sprintf("UPDATE %s SET %s=NULL,%s=NULL WHERE id=%v", table, attrName, attrLocName, obj.DBID())
        } else	{
		   stmt = fmt.Sprintf("UPDATE %s SET %s=NULL WHERE id=%v", table, attr.Part.Name, obj.DBID()) 
	    }
	} else { // non-primitive value type

		table := db.TableNameIfy(attr.ShortName())

		// TODO create a map of prepared statements and look up the statement to use.

		stmt = fmt.Sprintf("DELETE FROM %s WHERE id0=%v", table, obj.DBID()) // Ensure DBID?                                       
	}
	db.QueueStatements(stmt)
	return

}

/*
   Store the object  in the local sqlite database for the first time.
   If object is stored in local db already, does nothing.
   The object may already have a uuid, in which case it was a remote object (or an object whose persisting failed), 
   or it may not, in which case it is a new object that has only lived in memory in this process.
   Asynchronous. DB errors will not happen and be reported til later. <- to be true eventually
   NOT HAPPY WITH THE UNCERTAINTY OF STORAGE IN THE FLAG VALUE !!!
*/
func (db *SqliteDB) EnsurePersisted(obj RObject) (err error) {
	if obj.IsStoredLocally() {
		return
	}
	obj.SetStoredLocally() // Not necessarily true yet !!!!!! Failure to persist could happen after this. Can I move it down?
	if obj.HasUUID() {
		err = db.persistRemoteObject(obj)
	} else {
		err = db.persistNewObject(obj)
	}
	if err != nil {
		return
	}

	RT.Cache(obj) // Put in an in-memory object cache so that the runtime will only contain one object instance for each uuid.

	err = db.PersistAttributesAndRelations(obj)
	return
}

/*
   Persists (if necessary) each object which is an attribute of the argument object, or is related to the object.
   Then persists the attribute linkage table or relation table entry.
   Assumes that obj is persisted but has only just now been persisted.
   Note. Only persists relation entries in the forward direction.

   TODO WE STILL HAVE TO CATCH REFERENCE LOOPS !!! Or do we avoid them already?

   TODO NOT DOING RELATIONS YET !! JUST ATTRIBUTES !!!! !!!!!!!!! !!!!!!!!!!!!!!!!!!! !!!! !
*/
func (db *SqliteDB) PersistAttributesAndRelations(obj RObject) (err error) {
	defer Un(Trace(PERSIST_TR2, "PersistAttributesAndRelations", obj))

	objTyp := obj.Type()

	var stmt string

	for _, attr := range objTyp.Attributes {

		if !attr.Part.Type.IsPrimitive {

			table := db.TableNameIfy(attr.ShortName())

			val, found := RT.AttrValue(obj, attr, false, true)
			if !found {
				break
			}
			if attr.Part.CollectionType != "" { // a collection of non-primitives
				collection := val.(RCollection)
				isMap := collection.IsMap()
				if isMap {
					theMap := collection.(Map)
					for key := range theMap.Iter() {
						val, _ := theMap.Get(key)
						err = db.EnsurePersisted(val)
						if err != nil {
							return
						}
						if attr.Part.CollectionType == "stringmap" || attr.Part.CollectionType == "orderedstringmap" {
							stmt = fmt.Sprintf("INSERT INTO %s(id0,id1,key1) VALUES(%v,%v,%s)", table, obj.DBID(), val.DBID(), key) // Ensure DBID?   
						} else {
							// TODO - We do not know if the key is persisted. We don't know if the key is an integer!!!
							// !!!!!!!!!!!!!!!!!!!!!!!!
							// !!!! NOT DONE YET !!!!!!
							// !!!!!!!!!!!!!!!!!!!!!!!!
							stmt = fmt.Sprintf("INSERT INTO %s(id0,id1,ord1) VALUES(%v,%v,%v)", table, obj.DBID(), val.DBID(), key.DBID()) // Ensure DBID?   					 
						}
						db.QueueStatements(stmt)
					}
				} else {
					i := 0
					for val := range collection.Iter() {
						err = db.EnsurePersisted(val)
						if err != nil {
							return
						}
						if collection.IsOrdered() {
							stmt = fmt.Sprintf("INSERT INTO %s(id0,id1,ord1) VALUES(%v,%v,%v)", table, obj.DBID(), val.DBID(), i) // Ensure DBID?  
						} else { // unordered set 
							stmt = fmt.Sprintf("INSERT INTO %s(id0,id1) VALUES(%v,%v)", table, obj.DBID(), val.DBID()) // Ensure DBID?  		
						}
						db.QueueStatements(stmt)
						i++
					}
				}
			} else { // a single non-primitive value

				err = db.EnsurePersisted(val)
				if err != nil {
					return
				}
				stmt = fmt.Sprintf("INSERT INTO %s(id0,id1) VALUES(%v,%v)", table, obj.DBID(), val.DBID()) // Ensure DBID?   
				db.QueueStatements(stmt)
			}
		} else if attr.Part.CollectionType != "" { // a collection of primitive-type objects
			// TODO
			// !!!!!!!!!!!!!!!!!!!!!!!!
			// !!!! NOT DONE YET !!!!!!
			// !!!!!!!!!!!!!!!!!!!!!!!!	   
		}
	}

	for _, typ := range objTyp.Up {
		for _, attr := range typ.Attributes {
			if !attr.Part.Type.IsPrimitive {

				table := db.TableNameIfy(attr.ShortName())

				val, found := RT.AttrVal(obj, attr)
				if !found {
					break
				}
				if attr.Part.CollectionType != "" { // a collection of non-primitives
					collection := val.(RCollection)
					isMap := collection.IsMap()
					if isMap {
						theMap := collection.(Map)
						for key := range theMap.Iter() {
							val, _ := theMap.Get(key)
							err = db.EnsurePersisted(val)
							if err != nil {
								return
							}
							if attr.Part.CollectionType == "stringmap" || attr.Part.CollectionType == "orderedstringmap" {
								stmt = fmt.Sprintf("INSERT INTO %s(id0,id1,key1) VALUES(%v,%v,%s)", table, obj.DBID(), val.DBID(), key) // Ensure DBID?   
							} else {
								// TODO - We do not know if the key is persisted. We don't know if the key is an integer!!!
								// !!!!!!!!!!!!!!!!!!!!!!!!
								// !!!! NOT DONE YET !!!!!!
								// !!!!!!!!!!!!!!!!!!!!!!!!
								stmt = fmt.Sprintf("INSERT INTO %s(id0,id1,ord1) VALUES(%v,%v,%v)", table, obj.DBID(), val.DBID(), key.DBID()) // Ensure DBID?   					 
							}
							db.QueueStatements(stmt)
						}
					} else {
						i := 0
						for val := range collection.Iter() {
							err = db.EnsurePersisted(val)
							if err != nil {
								return
							}
							if collection.IsOrdered() {
								stmt = fmt.Sprintf("INSERT INTO %s(id0,id1,ord1) VALUES(%v,%v,%v)", table, obj.DBID(), val.DBID(), i) // Ensure DBID?  
							} else { // unordered set 
								stmt = fmt.Sprintf("INSERT INTO %s(id0,id1) VALUES(%v,%v)", table, obj.DBID(), val.DBID()) // Ensure DBID?  		
							}
							db.QueueStatements(stmt)
							i++
						}
					}
				} else { // a single non-primitive value

					err = db.EnsurePersisted(val)
					if err != nil {
						return
					}
					stmt = fmt.Sprintf("INSERT INTO %s(id0,id1) VALUES(%v,%v)", table, obj.DBID(), val.DBID()) // Ensure DBID?   
					db.QueueStatements(stmt)
				}
			} else if attr.Part.CollectionType != "" { // a collection of primitive-type objects
				// TODO
				// !!!!!!!!!!!!!!!!!!!!!!!!
				// !!!! NOT DONE YET !!!!!!
				// !!!!!!!!!!!!!!!!!!!!!!!!	   
			}
		}
	}
	return
}

/*
   Persist a remote object which has not yet been persisted locally.
   TODO This does not check whether the object is already in the local db, and if it is,
   problems will occur!
   What should we do in that case?
   Check in this method and fetch and return the old object?
   Throw an exception?
   Right now it will probably create a second stored copy of the object with id reversed!!!
*/
func (db *SqliteDB) persistRemoteObject(obj RObject) (err error) {

	obj.ClearIdReversed()
	id, id2 := obj.UUIDuint64s()

	dbid := int64(id)
	dbid2 := int64(id2)

	found, err := db.exists(dbid)
	if err != nil {
		return
	}
	if found {
		dbid, dbid2 = dbid2, dbid
		obj.SetIdReversed()
		found, err = db.exists(dbid)
		if err != nil {
			return
		}
		if found {
			rterr.Stop("Can't store remote object locally because of db id conflict.") // Chance of this is extremely small. Two 64 bit int random ids had to be present twice. 
		}
	}
	db.insert(obj, dbid, dbid2)
	return
}

/*
   Persist a newly created object. Creates a uuid for it.
*/
func (db *SqliteDB) persistNewObject(obj RObject) (err error) {
	for { // loop til we create a first half of uuid which is unique as an object id in the local db. 
		// Should almost never require more than one attempt.
		var id,id2 uint64
		id, id2, err = obj.EnsureUUIDuint64s()
		if err != nil {
			return
		}
		dbid := int64(id)
		var found bool
		found, err = db.exists(dbid)
		if err != nil {
			return
		}
		if found {
			obj.RemoveUUID()
		} else {
			db.insert(obj, dbid, int64(id2))

			obj.SetStoredLocally() // We don't actually know if this is correct yet. The db statements may have failed. TODO!!! FIX

			break
		}
	}
	return
}

/*
   Returns true if the object is found in the database.
*/
func (db *SqliteDB) exists(id int64) (found bool, err error) {
	stmt := fmt.Sprintf("SELECT count(*) FROM RObject where id=%v", id)
	selectStmt, err := db.conn.Prepare(stmt)
	if err != nil {
		return
	}

	defer selectStmt.Finalize()

	err = selectStmt.Exec()
	if err != nil {
		return
	}
	if selectStmt.Next() {
		var n int64

		err = selectStmt.Scan(&n)
		if err != nil {
			return
		}
		if n >= 1 {
			found = true
			if n > 1 {
				rterr.Stop("More than one object with same id in sqlite db!")
			}
		}
	} else {
		panic("Why is there no result of a count(*) statement???")
	}
	return
}

/*
   Gives the (already persisted) object an official name in the database.
*/
func (db *SqliteDB) NameObject(obj RObject, name string) {
	id := obj.DBID()
	stmt := fmt.Sprintf("INSERT INTO RName(name,id) VALUES('%s',%v);", name, id)
	db.QueueStatements(stmt)
}

/*
Returns true if an object has been named in the database with the argument name.
*/
func (db *SqliteDB) ObjectNameExists(name string) (found bool, err error) {
	stmt := fmt.Sprintf("SELECT count(*) FROM RName where name='%s'", name)
	selectStmt, err := db.conn.Prepare(stmt)
	if err != nil {
		return
	}

	defer selectStmt.Finalize()

	err = selectStmt.Exec()
	if err != nil {
		return
	}
	if selectStmt.Next() {
		var n int64

		err = selectStmt.Scan(&n)
		if err != nil {
			return
		}
		found = (n > 0)
	} else {
		panic("Why is there no result of a count(*) statement???")
	}
	return
}

/*
   Given the local dbid of an object, retrieve and return the object stored in the database.
   If the object is a unit (structure), the radius determines how many related objects in the tree are also fetched.
   0 means fetch object only. 1 means referred-to objects, 2 means referred-to by referred to etc.

   When a collection of non-primitives is fetched, it is lazily fetched. Currently that means that proxy
   objects are created in memory for each element of the collection.
   So for example, if you have an object with a multi-valued attribute, and you want the actual other objects fetched
   along with the object, use a depth of 2.

   TODO THIS METHOD NEEDS TO HAVE A MUTEX LOCK!!!!!!!

   TODO THIS METHOD NEEDS TO FLUSH THE SQL STATEMENT QUEUE BEFORE IT RUNS THE SELECT QUERY!!!!!!!
*/
func (db *SqliteDB) Fetch(id int64, radius int) (obj RObject, err error) {
	defer Un(Trace(PERSIST_TR, "Fetch", id, radius))
	var found bool
	obj, found = RT.GetObject(id)
	if found {
		return
	}
	stmt := fmt.Sprintf("SELECT * FROM RObject where id=%v", id)
	return db.fetch1(stmt, radius, fmt.Sprintf("id=%v", id), false)
}

/*
   Given a name that some object has been 'dubbed' with, retrieve and return the object stored in the database.
   If the object is a unit (structure), the radius determines how many related objects in the tree are also fetched.
   0 means fetch object only. 1 means referred-to objects, 2 means referred-to by referred to etc.

   When a collection of non-primitives is fetched, it is lazily fetched. Currently that means that proxy
   objects are created in memory for each element of the collection.
   So for example, if you have an object with a multi-valued attribute, and you want the actual other objects fetched
   along with the object, use a depth of 2.

   TODO THIS METHOD NEEDS TO HAVE A MUTEX LOCK!!!!!!!

   TODO THIS METHOD NEEDS TO FLUSH THE SQL STATEMENT QUEUE BEFORE IT RUNS THE SELECT QUERY!!!!!!!
*/
func (db *SqliteDB) FetchByName(name string, radius int) (obj RObject, err error) {
	defer Un(Trace(PERSIST_TR, "FetchByName", name, radius))
	stmt := fmt.Sprintf("SELECT * FROM RObject WHERE id IN (SELECT id FROM RName WHERE name='%s')", name)
	return db.fetch1(stmt, radius, fmt.Sprintf("name='%s'", name), true)
}

/*
   Give the dbid of an object, fetches the value of the specified attribute of the object from the database.
   The attribute should have a non-primitive value type or be multi-valued with primitive type of collection members.
*/
func (db *SqliteDB) FetchAttribute(objId int64, obj RObject, attr *AttributeSpec, radius int) (val RObject, err error) {
	defer Un(Trace(PERSIST_TR, "FetchAttribute", objId, radius))

	if attr.Part.CollectionType == "" && !attr.Part.Type.IsPrimitive {
		err = db.fetchUnaryNonPrimitiveAttributeValue(objId, obj, attr, radius)
		if err != nil {
			return
		}
	} else if attr.Part.CollectionType != "" && !attr.Part.Type.IsPrimitive {
		err = db.fetchNonPrimitiveAttributeValueCollection(objId, obj, attr, radius)
		if err != nil {
			return
		}
	} else if attr.Part.CollectionType != "" && attr.Part.Type.IsPrimitive {
		err = errors.New("I don't handle collections of primitives yet.")
		return
	}
	val, _ = RT.AttrVal(obj, attr)
	return
}

/*
   TODO Retrieve the object stored in the database.
   If the object is a unit (structure), the radius determines how many related objects in the tree are also fetched.
   0 means fetch object only. 1 means referred-to objects, 2 means referred-to by referred to etc.

   TODO THIS METHOD NEEDS TO HAVE A MUTEX LOCK!!!!!!!

   TODO THIS METHOD NEEDS TO FLUSH THE SQL STATEMENT QUEUE BEFORE IT RUNS THE SELECT QUERY!!!!!!!
*/
func (db *SqliteDB) fetch1(query string, radius int, errSuffix string, checkCache bool) (obj RObject, err error) {
	selectStmt, err := db.conn.Prepare(query)
	if err != nil {
		return
	}

	defer selectStmt.Finalize()

	err = selectStmt.Exec()
	if err != nil {
		return
	}
	if !selectStmt.Next() {
		panic(fmt.Sprintf("No object found in database with %s.", errSuffix))
	}
	var id int64
	var id2 int64
	var flags int
	var typeName string

	err = selectStmt.Scan(&id, &id2, &flags, &typeName)
	if err != nil {
		return
	}

	if checkCache {
		dbid := DBID(id, id2, flags)

		var found bool
		obj, found = RT.GetObject(dbid)
		if found {
			return
		}
	}

    fullTypeName := RT.Typs[typeName].Name

	obj, err = RT.NewObject(fullTypeName)
	if err != nil {
		return
	}

	// Now we have to store the unit64(id),uint64(id2),byte(flags) into the object.

	//   unit := obj.(*runit)
	//   (&(unit.robject)).RestoreIdsAndFlags(id,id2,flags)

	ob := obj.(Persistable)
	ob.RestoreIdsAndFlags(id, id2, flags)

	Logln(PERSIST2_, "id:", id, ", id2:", id2, ", flags:", flags, ", typeName:", typeName)

	oid, oid2 := obj.UUIDuint64s()

	Logln(PERSIST2_, "obj.id:", oid, ", obj.id2:", oid2, ", Flags():", obj.Flags(), ", obj.Type():", obj.Type())

	// Now fetch and restore the values of the unary primitive attributes of the object.
	// These attribute values are stored in the db rows that represent the object in the db.
	// The object in the db consists of a single row in each of several database tables.
	// There is one table for each type the object conforms to (i.e. specific type and supertypes), 
	// and a single row in each such table identified by the object's dbid.

	err = db.fetchUnaryPrimitiveAttributeValues(id, obj)
	if err != nil {
		return
	}

	// Have to set this here before confirmed in order to avoid attribute or relation reference loops causing
	// infinite looping during fetching. 
	// TODO consider replacing with SetStoringLocally and a later SetStoredLocally

	obj.SetStoredLocally()

	// Now fetch (at least proxies for) the non-primitive attributes (if we should do it now.)
	// Maybe this should be fully lazy. Wait until the attribute value is asked for.

	// TODO

	// THIS NEEDS TO DEPEND ON DEPTH

	// TODO

	if radius > 0 {
		err = db.fetchUnaryNonPrimitiveAttributeValues(id, obj, radius-1)
		if err != nil {
			return
		}

		err = db.fetchMultiValuedNonPrimitiveAttributeValues(id, obj, radius-1)
		if err != nil {
			return
		}

		err = db.fetchMultiValuedPrimitiveAttributeValues(id, obj, radius-1)
		if err != nil {
			return
		}

	}
	return
}




/*
   TODO Retrieve a list of objects stored in the database.

   TODO THIS METHOD NEEDS TO HAVE A MUTEX LOCK!!!!!!!

   TODO THIS METHOD NEEDS TO FLUSH THE SQL STATEMENT QUEUE BEFORE IT RUNS THE SELECT QUERY!!!!!!!
*/
func (db *SqliteDB) fetchMultipleEager(query string, radius int, numPrimitiveAttrs int, errSuffix string, checkCache bool, objs *[]RObject) (err error) {

	Logln(PERSIST_, query)

	selectStmt, err := db.conn.Prepare(query)
	if err != nil {
		return
	}

	defer selectStmt.Finalize()

	err = selectStmt.Exec()
	if err != nil {
		return
	}
	for selectStmt.Next() {
	
	    var obj RObject
	
		var id int64
		var id2 int64
		var flags int
		var typeName string

		attrValsBytes1 := make([][]byte, numPrimitiveAttrs)

		attrValsBytes := make([]interface{}, numPrimitiveAttrs + 4)

        attrValsBytes[0] = &id
        attrValsBytes[1] = &id2
        attrValsBytes[2] = &flags
        attrValsBytes[3] = &typeName

 		for i := 0; i < len(attrValsBytes1); i++ {
			attrValsBytes[i+4] = &attrValsBytes1[i]
		}


		err = selectStmt.Scan(attrValsBytes...)
		
		if err != nil {
			return
		}
		
		if checkCache {
			dbid := DBID(id, id2, flags)

			var found bool
			obj, found = RT.GetObject(dbid)
			if found {
	        	*objs = append(*objs, obj)				
				continue    // to next object in the resultset
			}
		}

        fullTypeName := RT.Typs[typeName].Name

	    obj, err = RT.NewObject(fullTypeName)
		if err != nil {
			return
		}

		// Now we have to store the unit64(id),uint64(id2),byte(flags) into the object.

		//   unit := obj.(*runit)
		//   (&(unit.robject)).RestoreIdsAndFlags(id,id2,flags)

		ob := obj.(Persistable)
		ob.RestoreIdsAndFlags(id, id2, flags)

		Logln(PERSIST2_, "id:", id, ", id2:", id2, ", flags:", flags, ", typeName:", typeName)

		oid, oid2 := obj.UUIDuint64s()

		Logln(PERSIST2_, "obj.id:", oid, ", obj.id2:", oid2, ", Flags():", obj.Flags(), ", obj.Type():", obj.Type())

		// Now restore the values of the unary primitive attributes of the object.
		// These attribute values are stored in the db rows that represent the object in the db.
		// The object in the db consists of a single row in each of several database tables.
		// There is one table for each type the object conforms to (i.e. specific type and supertypes), 
		// and a single row in each such table identified by the object's dbid.
		
	    objTyp := obj.Type()
	    attrValsBytes = attrValsBytes[4:]
        db.restoreAttrs(obj, objTyp, attrValsBytes)		

		// Have to set this here before confirmed in order to avoid attribute or relation reference loops causing
		// infinite looping during fetching. 
		// TODO consider replacing with SetStoringLocally and a later SetStoredLocally

		obj.SetStoredLocally()

		// Now fetch (at least proxies for) the non-primitive attributes (if we should do it now.)
		// Maybe this should be fully lazy. Wait until the attribute value is asked for.

		// TODO

		// THIS NEEDS TO DEPEND ON DEPTH

		// TODO
    
		if radius > 0 {
			err = db.fetchUnaryNonPrimitiveAttributeValues(id, obj, radius-1)
			if err != nil {
				return
			}

			err = db.fetchMultiValuedNonPrimitiveAttributeValues(id, obj, radius-1)
			if err != nil {
				return
			}

			err = db.fetchMultiValuedPrimitiveAttributeValues(id, obj, radius-1)
			if err != nil {
				return
			}
		}
		
		*objs = append(*objs, obj)

    }

	return
}



/*
Fetch from db and set unary primitive-valued attributes of an object.
*/
func (db *SqliteDB) fetchUnaryPrimitiveAttributeValues(id int64, obj RObject) (err error) {
	defer Un(Trace(PERSIST_TR2, "fetchPrimitiveAttributeValues", id))
	// TODO

	objTyp := obj.Type()

	// Create the select clause

    numPrimAttributeColumns := 0

	selectClause := "SELECT "
	sep := ""
	for _, attr := range objTyp.Attributes {
		if attr.Part.Type.IsPrimitive && attr.Part.CollectionType == "" {
			if attr.Part.Type == TimeType {
			   selectClause += sep + attr.Part.Name	+ "," + attr.Part.Name + "_loc" 
			   numPrimAttributeColumns += 2							
			} else if attr.Part.Type == ComplexType || attr.Part.Type == Complex32Type {
			   selectClause += sep + attr.Part.Name	+ "_r," + attr.Part.Name + "_i" 	
			   numPrimAttributeColumns += 2
			} else {
			   selectClause += sep + attr.Part.Name
			   numPrimAttributeColumns ++
		    }
			sep = ","
		}
	}

	for _, typ := range objTyp.Up {
		for _, attr := range typ.Attributes {
				if attr.Part.Type.IsPrimitive && attr.Part.CollectionType == "" {
				if attr.Part.Type == TimeType {
				   selectClause += sep + attr.Part.Name	+ "," + attr.Part.Name + "_loc" 
				   numPrimAttributeColumns += 2							
				} else if attr.Part.Type == ComplexType || attr.Part.Type == Complex32Type {
				   selectClause += sep + attr.Part.Name	+ "_r," + attr.Part.Name + "_i" 	
				   numPrimAttributeColumns += 2
				} else {
				   selectClause += sep + attr.Part.Name
				   numPrimAttributeColumns ++
			    }
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

	for _, typ := range objTyp.Up {
		from += " JOIN " + db.TableNameIfy(typ.ShortName()) + " USING (id)"
		numPrimitiveAttrs += typ.NumPrimitiveAttributes
	}

	where := fmt.Sprintf(" WHERE %s.id=%v", specificTypeTable, id)

	stmt := selectClause + from + where

	Logln(PERSIST2_, "query:", stmt)

	selectStmt, err := db.conn.Prepare(stmt)
	if err != nil {
		return
	}

	defer selectStmt.Finalize()

	err = selectStmt.Exec()
	if err != nil {
		return
	}
	if !selectStmt.Next() {
		panic(fmt.Sprintf("No object found in database with id=%v", id))
	}

	// Now construct a Scan call with [] *[]byte
	/*
	   attrValsBytes1 := make([]*[]byte,numPrimitiveAttrs)

	   attrValsBytes := make([]interface{},numPrimitiveAttrs)

	   for i := 0; i < len(attrValsBytes1); i++ {
	      attrValsBytes[i] = attrValsBytes1[i]	
	   }
	*/

	attrValsBytes1 := make([][]byte, numPrimAttributeColumns)

	attrValsBytes := make([]interface{}, numPrimAttributeColumns)

	for i := 0; i < len(attrValsBytes1); i++ {
		attrValsBytes[i] = &attrValsBytes1[i]
	}

	// have to make a slice whose length = the total number of attributes in all the types combined.
	// 
	err = selectStmt.Scan(attrValsBytes...)
	if err != nil {
		return
	}
	
    db.restoreAttrs(obj, objTyp, attrValsBytes)

	return
}

/*
Given the result of a scan of a select result row, restore object attribute values for an object in the runtime.
*/
func (db *SqliteDB) restoreAttrs(obj RObject, objTyp *RType, attrValsBytes []interface{}) {

	// Now go through the attrValsBytes and interpret each according to the datatype of each primitive
	// attribute, and set the primitive attributes using the runtime SetAttrVal method.

	i := 0
	var val RObject
	var nonNil bool 

	for _, attr := range objTyp.Attributes {
		if attr.Part.Type.IsPrimitive  && attr.Part.CollectionType == "" {
			valByteSlice := *(attrValsBytes[i].(*[]byte))
			if attr.Part.Type == TimeType || attr.Part.Type == ComplexType || attr.Part.Type == Complex32Type {
				i++
				valByteSlice2 := *(attrValsBytes[i].(*[]byte))
     		    nonNil = convertAttrValTwoFields(valByteSlice, valByteSlice2, attr, &val) 				
			} else {
			   nonNil = convertAttrVal(valByteSlice, attr, &val) 
		    }
		    if nonNil {
		   		RT.RestoreAttr(obj, attr, val)			    
		    }
		    i++
		}
	}

	for _, typ := range objTyp.Up {
		for _, attr := range typ.Attributes {
			if attr.Part.Type.IsPrimitive  && attr.Part.CollectionType == "" {
				valByteSlice := *(attrValsBytes[i].(*[]byte))
				if attr.Part.Type == TimeType || attr.Part.Type == ComplexType || attr.Part.Type == Complex32Type {
					i++
					valByteSlice2 := *(attrValsBytes[i].(*[]byte))
	     		    nonNil = convertAttrValTwoFields(valByteSlice, valByteSlice2, attr, &val) 				
				} else {
				   nonNil = convertAttrVal(valByteSlice, attr, &val) 
			    }
			    if nonNil {
			   		RT.RestoreAttr(obj, attr, val)			    
			    }
			    i++				
			}
		}
	}
}



/*
Fetch from the db the values of non-primitive-typed unary attributes of the object.
Set each attribute to the fetched object.
*/
func (db *SqliteDB) fetchUnaryNonPrimitiveAttributeValues(id int64, obj RObject, radius int) (err error) {
	defer Un(Trace(PERSIST_TR2, "fetchUnaryNonPrimitiveAttributeValues", id))

	objTyp := obj.Type()

	for _, attr := range objTyp.Attributes {
		if attr.Part.CollectionType == "" && !attr.Part.Type.IsPrimitive {
			err = db.fetchUnaryNonPrimitiveAttributeValue(id, obj, attr, radius)
			if err != nil {
				return
			}
		}
	}

	for _, typ := range objTyp.Up {
		for _, attr := range typ.Attributes {
			if attr.Part.CollectionType == "" && !attr.Part.Type.IsPrimitive {
				err = db.fetchUnaryNonPrimitiveAttributeValue(id, obj, attr, radius)
				if err != nil {
					return
				}
			}
		}
	}
	return
}

/*
Fetch from the db the object value of a non-primitive-typed unary attribute of the object.
Set the attribute to the fetched object.  
*/
func (db *SqliteDB) fetchUnaryNonPrimitiveAttributeValue(objId int64, obj RObject, attr *AttributeSpec, radius int) (err error) {
	defer Un(Trace(PERSIST_TR2, "fetchUnaryNonPrimitiveAttributeValue", objId, attr.ShortName()))

	// First query for the id of the attribute value object

	query := fmt.Sprintf("SELECT id1 FROM %s WHERE id0=%v", db.TableNameIfy(attr.ShortName()), objId)

	selectStmt, err := db.conn.Prepare(query)
	if err != nil {
		return
	}

	defer selectStmt.Finalize()

	err = selectStmt.Exec()
	if err != nil {
		return
	}

	if !selectStmt.Next() {
		// TODO NOT SURE IF THIS SHOULD BE AN ERROR!!!!!!
		err = errors.New(fmt.Sprintf("Object %s has no value for attribute %s", objId, attr.Part.Name))
		return
	}

	var id1 int64
	err = selectStmt.Scan(&id1)
	if err != nil {
		return
	}

	// Then fetch the object with the id
	val, err := db.Fetch(id1, radius)
	if err != nil {
		return
	}
	RT.RestoreAttr(obj, attr, val)

	return
}

/*
Convert the byte slice which was returned as a column-value of a databse result row into
a primitive-type RObject. Sets the val argument to the new RObject.
If the value from the database was NULL (empty string in numeric fields), does not
set the val argument, and returns false.

TODO NOT HANDLING NULLS PROPERLY HERE YET !!!!!!!!!

*/
func convertAttrVal(valByteSlice []byte, attr *AttributeSpec, val *RObject) (nonNullValueFound bool) {
	switch attr.Part.Type {
	case IntType:
		if len(valByteSlice) > 0 {
			x, err := strconv.ParseInt(string(valByteSlice), 10, 64)
			if err != nil {
				panic(errors.New("attr " + attr.Part.Name + " as int64: " + err.Error()))
			}
			*val = Int(x)
			nonNullValueFound = true
		}

	case Int32Type:
		if len(valByteSlice) > 0 {
			x, err := strconv.Atoi(string(valByteSlice))
			if err != nil {
				panic(errors.New("attr " + attr.Part.Name + " as int: " + err.Error()))
			}
			*val = Int32(x)
			nonNullValueFound = true
		}

	case FloatType:
		if len(valByteSlice) > 0 {
			x, err := strconv.ParseFloat(string(valByteSlice), 64)
			if err != nil {
				panic(errors.New("attr " + attr.Part.Name + " as float64: " + err.Error()))
			}
			*val = Float(x)
			nonNullValueFound = true
		}

	case BoolType:
		*val = Bool(string(valByteSlice) == "1")

	case StringType:
		*val = String(SqlStringValueUnescqpe(string(valByteSlice)))
		nonNullValueFound = true // Shoot!! How do we distinguish between not set and empty string?

	default:
		panic(fmt.Sprintf("I don't know how to restore a type %v attribute.", attr.Part.Type))
	}
	return
}


/*
Convert the byte slice which was returned as a column-value of a databse result row into
a primitive-type RObject. Sets the val argument to the new RObject.
If the value from the database was NULL (empty string in numeric fields), does not
set the val argument, and returns false.

TODO NOT HANDLING NULLS PROPERLY HERE YET !!!!!!!!!

*/
func convertAttrValTwoFields(valByteSlice []byte, valByteSlice2 []byte, attr *AttributeSpec, val *RObject) (nonNullValueFound bool) {
	switch attr.Part.Type {
	case TimeType:
		if len(valByteSlice) > 0 {
			timeString := string(valByteSlice)
			var locationName string
	  	    if len(valByteSlice2) > 0 {
			   locationName = string(valByteSlice2)		       
		    }		
			timeUTC, err := time.Parse(TIME_LAYOUT, timeString) 
			if err != nil {
				panic(errors.New("attr " + attr.Part.Name + " as Time: " + err.Error()))
			}
			location, err := time.LoadLocation(locationName)
			if err != nil {
				panic(errors.New("attr " + attr.Part.Name + " time.Location: " + err.Error()))
			}			
            *val = RTime(timeUTC.In(location))	
			nonNullValueFound = true
		}



	case ComplexType:
		if len(valByteSlice) > 0 {
			r, err := strconv.ParseFloat(string(valByteSlice), 64)
			if err != nil {
				panic(errors.New("attr " + attr.Part.Name + "_r as float64: " + err.Error()))
			}
		    if len(valByteSlice2) == 0 {
				panic(errors.New("attr " + attr.Part.Name + " imaginary part is null"))			
			}
			
			i, err := strconv.ParseFloat(string(valByteSlice), 64)
			if err != nil {
				panic(errors.New("attr " + attr.Part.Name + "_i as float64: " + err.Error()))
			}			
				
			*val = Complex(complex(r,i))
			nonNullValueFound = true
		}
		
	case Complex32Type:
		if len(valByteSlice) > 0 {
			r, err := strconv.ParseFloat(string(valByteSlice), 32)
			if err != nil {
				panic(errors.New("attr " + attr.Part.Name + "_r as float32: " + err.Error()))
			}
		    if len(valByteSlice2) == 0 {
				panic(errors.New("attr " + attr.Part.Name + " imaginary part is null"))			
			}
			
			i, err := strconv.ParseFloat(string(valByteSlice), 32)
			if err != nil {
				panic(errors.New("attr " + attr.Part.Name + "_i as float32: " + err.Error()))
			}			
				
			r32 := float32(r)
			i32 := float32(i)	
			*val = Complex32(complex(r32,i32))
			nonNullValueFound = true
		}		

	default:
		panic(fmt.Sprintf("I don't know how to restore a type %v attribute.", attr.Part.Type))
	}
	return
}




/*
   Persist an robject for the first time. 
   TODO need a special case for a collection.

   1. inserts a row into the RObject table.
   2. For the object's type and each type in the up chain of the type, create a row in the type's table.

*/
func (db *SqliteDB) insert(obj RObject, dbid, dbid2 int64) {

	stmt := fmt.Sprintf("INSERT INTO RObject(id,id2,flags,typeName) VALUES(%v,%v,%v,'%s');", dbid, dbid2, obj.Flags(), obj.Type().ShortName())

	stmt += db.instanceInsertStatement(obj.Type(), obj)
	for _, typ := range obj.Type().Up {
		stmt += db.instanceInsertStatement(typ, obj)
	}

	db.QueueStatements(stmt)
}

/*
   Return the SQL INSERT statement that inserts a row in the type's instance-persistence table.
   This table is named with the type's short name, and contains columns to represent the primitive-valued attributes
   that are defined for the type.

   Note: Must begin with ";"
*/
func (db *SqliteDB) instanceInsertStatement(t *RType, obj RObject) string {

	table := db.TableNameIfy(t.ShortName())
	primitiveAttrVals := db.primitiveAttrValsSQL(t, obj)
	s := fmt.Sprintf("INSERT INTO %s VALUES(%v%s);", table, obj.DBID(), primitiveAttrVals)
	return s
}

/*
Return a comma separated string representation of the values (for the argument object) of 
the primitive=valued attributes defined in the type.
*/
func (db *SqliteDB) primitiveAttrValsSQL(t *RType, obj RObject) string {

	var s string
	for _, attr := range t.Attributes {
		if attr.Part.Type.IsPrimitive {
			val, found := RT.AttrVal(obj, attr)
			s += ","
			if found {
				switch val.(type) {
				case Int:
					s += strconv.FormatInt(int64(val.(Int)), 10)
				//case Uint:
				//   s += strconv.Itoa64(int64(val.(Uint)))		
				case RTime:
					t := time.Time(val.(RTime))
					timeString := t.UTC().Format(TIME_LAYOUT)
					locationName := t.Location().String()
					s += "'" + timeString + "','" + locationName + "'"															
				case String:
					s += "'" + SqlStringValueEscqpe(string(val.(String))) + "'"
				case Float:
					s += strconv.FormatFloat(float64(val.(Float)), 'G', -1, 64)
					
				case Complex:
					c := complex128(val.(Complex))
					r := real(c)
					i := imag(c)
					s += strconv.FormatFloat(r, 'G', -1, 64) + "," + strconv.FormatFloat(i, 'G', -1, 64)
				case Complex32:
					c := complex64(val.(Complex32))
					r := float64(real(c))
					i := float64(imag(c))					
					s += strconv.FormatFloat(r, 'G', -1, 32) + "," + strconv.FormatFloat(i, 'G', -1, 32)								
				case Bool:
					boolVal := bool(val.(Bool))
					if boolVal {
						s += "1"
					} else {
						s += "0"
					}
				case Int32:
					s += strconv.Itoa(int(val.(Int32)))
				default:
					panic(fmt.Sprintf("I don't know how to create SQL for an attribute value of underlying type %v.", val.Type()))
				}
			} else if attr.Part.Type == TimeType || attr.Part.Type == ComplexType || attr.Part.Type == Complex32Type {
				s += "NULL,NULL"
			} else {
				s += "NULL"
			}
		}
	}
	return s

}

func SqlStringValueEscqpe(s string) string {
	s = strconv.Quote(s)	
    // Use QuoteToASCII(s) instead???
	
	s = strings.Replace(s, `\"`, `"`, -1)
	s = strings.Replace(s, "'", "''", -1)
	s = s[1:len(s)-1] // Strip surrounding double-quotes
	return s
}

func SqlStringValueUnescqpe(s string) string {

	
	s = strings.Replace(s, `"`, `\"`, -1)
	s, err := strconv.Unquote(`"` + s + `"`)
	if err != nil {
		panic(err)
	}
	return s
}