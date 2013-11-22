// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.

package data

/*
   runtime.go - the runtime environment for a relish program - manages other entities such
   as packages, types, methods, and object instances and their attributes and relations.

   runtimeenv
*/

import (
	"fmt"
	. "relish/dbg"
	"sync"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////////////////////
//////// RUNTIME ENVIRONMENT
//////// 
//////// Used by the interpreter. Used and populated by the parser.
//////// Contains maps from names to types and methods.
//////// Contains a tree index of all unique type tuples that have been 
//////// used in the definition or a call of a method.
////////
//////// Will also contain references to Relish data objects (global variables) I imagine.
////////
////////////////////////////////////////////////////////////////////////////////////////////////

/*
   RuntimeEnv - the nexus for Relish types and methods and objects and related indexes.
   Has methods to create types and methods, which are used by the parser to create a program
   runtime.
   Also has methods to set and get values of attributes of objects.
   Note: Many of the most important methods of RuntimeEnv are defined in the other
   relish source files which deal specifically with other kinds of runtime entities.
*/
type RuntimeEnv struct {
	Types         map[string]*RType        // map from type name to RType object. 
	Typs          map[string]*RType        // map from type.ShortName() to RType object. 
	MultiMethods  map[string]*RMultiMethod // map from method name to RMultiMethod object.
	Packages      map[string]*RPackage     // by package Name
	Pkgs          map[string]*RPackage     // by package ShortName
	PkgShortNameToName map[string]string   
	PkgNameToShortName map[string]string
	InbuiltFunctionsPackage *RPackage      // shortcut to package containing inbuilt functions
	RunningArtifact string                 // Full origin-qualified name of the artifact one of whose package main functions was run in relish command
	TypeTupleTree *TypeTupleTreeNode
	objects       map[int64]RObject // persistable objects by DBID()
	// objects map[string] RObject // persistable objects by uuidstr - do we need this for distributed computing?
	objectIds map[RObject]uint64 // non-persistable object to its local numeric id (for debug printing the object)
	idGen     *IdGenerator
	db        DB
	DatabaseURI string  // filepath (or eventually some other kind of db connection uri) of the database for persisting relish objects
	
	// Note about attribute values. An attribute may have a single value or be multi-valued,
	// where the multi-valued attribute is implemented by an RCollection.
	// But how do we tell if we have a single-valued attribute whose value is an ROject that
	// happens to be a collection? The answer is that a multi-value-attribute-implementing
	// RCollection must have a reference to its owner RObject. A plain-old RCollection object
	// which is a single value always has a nil owner.
	//
	attributes map[*AttributeSpec]map[RObject]RObject

	// TODO HOW do we deal with object networks that are not fully retrieved 
	// from storage yet? i.e. that need to be lazily loaded into memory?

	// Perhaps the trick is to load a collection valued attribute (when it or a member is requested)
	// in a partial way as follows: create the collection in memory but make it a collection
	// of RObjects which are nothing but the uint64 id of the true object. They have an IsProxy =true
	// Upon retrieving from the collection, we do a coll[i] = EnsureFetched(coll[i]) which 
	// replaces the ith member with the real RObject.
	// 

	// from fully qualified constant name to value.
	constants map[string]RObject

    // The following is to make sure that objects in-transit in a channel
    // can be marked by the relish garbage collector.
    // The map contains objects which have been sent into a channel and not yet received.
    // The value is the number of times the object is currently in a channel.
    inTransit map[RObject]uint32

	// from fully qualified type name to reflect.DataType, a *GoWrapper RObject.
	ReflectedDataTypes map[string]RObject

	// from fully qualified attribute name to reflect.Attribute, a *GoWrapper RObject.
	ReflectedAttributes map[string]RObject

	// a map of contexts for the evaluation of methods.

	evalContexts map[RObject]MethodEvaluationContext

	evalContextMutex sync.Mutex
}

var RT *RuntimeEnv

func init() {
	RT = NewRuntimeEnv()
	RT.createPrimitiveTypes()
}

func (rt *RuntimeEnv) DB() DB {
	return rt.db
}

func (rt *RuntimeEnv) CreateConstant(name string, value RObject) (err error) {
	if _, found := rt.constants[name]; found {
		err = fmt.Errorf("Redefining constant '%s'", name)
		return 
	}
	rt.constants[name] = value
	return
}

func (rt *RuntimeEnv) GetConstant(name string) (val RObject, found bool) {
	val, found = rt.constants[name]
	return
}

/*
   Caches the object in the runtime's in-memory object cache. Assumes an object instance with the same
   dbid is not already in the cache.
   Assumes the object already has a dbid i.e. has been persisted locally.
*/
func (rt *RuntimeEnv) Cache(obj RObject) {
	rt.objects[obj.DBID()] = obj
}

/*
   Return the object with the given dbid from the in-memory cache, or nil if not found.
*/
func (rt *RuntimeEnv) GetObject(id int64) (obj RObject, found bool) {
	obj, found = rt.objects[id]
	return
}

func (rt *RuntimeEnv) Uncache(obj RObject) {
	delete(rt.objects, obj.DBID()) 
}

type MemCache interface {
	/*
	   Caches the object in the in-memory object cache. Assumes an object instance with the same
	   dbid is not already in the cache.
	   Assumes the object already has a dbid i.e. has been persisted locally.
	*/	
	Cache(obj RObject)
	
    GetObject(id int64) (obj RObject, found bool) 
	
    Uncache(obj RObject)	
}

/*
   A generator of unsigned 64-bit integer ids starting at 1 and incrementing by 1 each time a new one
   is requested. Call the NextID() method to get a runtime-unique id.
*/
type IdGenerator struct {
	ch chan uint64
}

func NewIdGenerator() *IdGenerator {
	gen := &IdGenerator{ch: make(chan uint64)}
	go gen.sendID()
	return gen
}

/*
   Blocks until a next id is available.
*/
func (gen *IdGenerator) NextID() uint64 {
	id := <-gen.ch
	return id
}

func (gen *IdGenerator) sendID() {
	var id uint64
	for id = 1; ; id++ {
		gen.ch <- id
	}
}

/*
Still needs SetDB(db) before is valid.
*/
func NewRuntimeEnv() *RuntimeEnv {
	rt := &RuntimeEnv{Types: make(map[string]*RType),
		Typs: make(map[string]*RType),
		MultiMethods:  make(map[string]*RMultiMethod),
		Packages:      make(map[string]*RPackage),
		Pkgs:          make(map[string]*RPackage),
		PkgShortNameToName: make(map[string]string),   
		PkgNameToShortName: make(map[string]string),		
		TypeTupleTree: &TypeTupleTreeNode{},
		objects:       make(map[int64]RObject),
		objectIds:     make(map[RObject]uint64),
		idGen:         NewIdGenerator(),
		attributes:    make(map[*AttributeSpec]map[RObject]RObject),
		constants:     make(map[string]RObject),
		inTransit:     make(map[RObject]uint32),
	    ReflectedDataTypes: make(map[string]RObject),
	    ReflectedAttributes: make(map[string]RObject),
		evalContexts:  make(map[RObject]MethodEvaluationContext),
	}
	rt.PkgNameToShortName["relish.pl2012/core/inbuilt"] = "inbuilt"
	rt.PkgShortNameToName["inbuilt"] = "relish.pl2012/core/inbuilt"	
	return rt
}

/*
   Return the next available identifier for a not-yet-persistable object which must be
   identified.
*/
func (rt *RuntimeEnv) NextLocalID() uint64 {
	return rt.idGen.NextID()
}

func (rt *RuntimeEnv) SetDB(db DB) {
	rt.db = db
}

/*
   Return the value of the specified attribute for the specified object.
   Does not currently distinguish between multi-value attributes and single-valued.
   If it is a multi-valued attribute, returns the RCollection which implements
   the multi-value attribute.
   What does it return if no value has been defined? How about a found boolean
*/
func (rt *RuntimeEnv) AttrVal(obj RObject, attr *AttributeSpec) (val RObject, found bool) {
	return rt.AttrValue(obj, attr, true, true)
}

/*
   Return the value of the specified attribute for the specified object.
   Does not currently distinguish between multi-value attributes and single-valued.
   If it is a multi-valued attribute, returns the RCollection which implements
   the multi-value attribute.
   What does it return if no value has been defined? How about a found boolean
*/
func (rt *RuntimeEnv) AttrValue(obj RObject, attr *AttributeSpec, checkPersistence bool, allowNoValue bool) (val RObject, found bool) {
	attrVals, found := rt.attributes[attr]
	if found {
		val, found = attrVals[obj]
		if found {
			return
		} else if (! checkPersistence) && (! allowNoValue) {
			panic(fmt.Sprintf("Error: attribute %s.%s has no value.", obj, attr.Part.Name))			
		}
	}

	//Logln(PERSIST_,"AttrVal ! found in mem and strdlocally=",obj.IsStoredLocally())
	//Logln(PERSIST_,"AttrVal ! found in mem and attr.Part.CollectionType=",attr.Part.CollectionType)
	//Logln(PERSIST_,"AttrVal ! found in mem and attr.Part.Type.IsPrimitive=",attr.Part.Type.IsPrimitive)
	if checkPersistence && obj.IsStoredLocally() && (attr.Part.CollectionType != "" || !attr.Part.Type.IsPrimitive) {
		var err error
		val, err = rt.db.FetchAttribute(obj.DBID(), obj, attr, 0)
		if err != nil {
			// TODO  - NOT BEING PRINCIPLED ABOUT WHAT TO DO IF NO VALUE! Should sometimes allow, sometimes not!
			
            if strings.Contains(err.Error(), "has no value for attribute") {
	           if allowNoValue {
		          return
		       } else {			
			      panic(fmt.Sprintf("Error fetching attribute %s.%s from database: %s", obj, attr.Part.Name, err))
		       }
		    } else {
		       panic(fmt.Sprintf("Error fetching attribute %s.%s from database: %s", obj, attr.Part.Name, err))
			}
		}
		if val != nil {
			Logln(PERSIST2_, "AttrVal (fetched) =", val)
			found = true
		}
	}
	return
}





/*
Version to be used in template execution.

Note: Now checks relations as well as one-way attributes.
*/
func (rt *RuntimeEnv) AttrValByName(obj RObject, attrName string) (val RObject, err error) {

	attr, found := obj.Type().GetAttribute(attrName)
	if ! found {
       err = fmt.Errorf("Attribute or relation %s not found in type %v or supertypes.", attrName, obj.Type())		
	   return	
	}	
	val, _ = RT.AttrVal(obj, attr)
	return
}






/*
Untypechecked assignment. Assumes type has been statically checked.

func (rt *RuntimeEnv) SetAttr(obj RObject, attr *AttributeSpec, val RObject, , context MethodEvaluationContext, isInverse bool) {

	attrVals, found := rt.attributes[attr]
	if found {
		_, found = attrVals[obj]
	} else {
		attrVals = make(map[RObject]RObject)
		rt.attributes[attr] = attrVals
	}
	attrVals[obj] = val
	if obj.IsStoredLocally() {
		rt.db.PersistSetAttr(obj, attr, val, found)
	}
	
    if ! isInverse && attr.Inverse != nil {
	   err = rt.SetOrAddToAttr(val, attr.Inverse, obj, context, true)
    }	
	return
}
*/

/*
Untypechecked assignment. Used in restoration (summoning) of an object from persistent storage.
*/
func (rt *RuntimeEnv) RestoreAttr(obj RObject,  attr *AttributeSpec, val RObject) {
	
	defer Un(Trace(PERSIST_TR2, "RestoreAttr", obj, attr, val))
	attrVals, found := rt.attributes[attr]
	if !found {
		attrVals = make(map[RObject]RObject)
		rt.attributes[attr] = attrVals
	}
	attrVals[obj] = val

	return
}

/*
Optionally typechecked assignment. Never used in inverse.
*/
func (rt *RuntimeEnv) SetAttr(th InterpreterThread, obj RObject, attr *AttributeSpec, val RObject, typeCheck bool, context MethodEvaluationContext, isInverse bool) (err error) {

    if obj == NIL {
		err = fmt.Errorf("nil cannot have a value assigned to '%v' attribute.", attr.Part.Name)
		return		
    }

    if val == NIL {
       err = rt.UnsetAttr(th, obj, attr, isInverse, true)
       return
    }

	if typeCheck && !val.Type().LessEq(attr.Part.Type) {
		err = fmt.Errorf("Cannot assign  '%v.%v %v' a value of type '%v'.", obj.Type(), attr.Part.Name, attr.Part.Type, val.Type())
		return
	}
	
	attrVals, found := rt.attributes[attr]
	
	var oldVal RObject
	 
	if found {
		oldVal, found = attrVals[obj]
	} else {
		attrVals = make(map[RObject]RObject)
		rt.attributes[attr] = attrVals
	} 

    // Also have to do this in UnsetAttr !!!
	if ! found {
		if obj.IsStoredLocally() && attr.IsRelation() {
			oldVal, found = rt.AttrVal(obj, attr)
		}
	}


	attrVals[obj] = val
	if obj.IsStoredLocally() {
		th.DB().PersistSetAttr(obj, attr, val, found)
	}

	if ! isInverse && attr.Inverse != nil {
	   if oldVal != nil && oldVal != NIL {
		   // Remove (one entry for) obj from oldVal's inverse attr. 
		   err = rt.RemoveAttrGeneral(th,oldVal,attr.Inverse,obj,true,true)
		   if err != nil {
			  return
		   }
	   }	
	   if val != NIL {
	   	  err = rt.SetOrAddToAttr(th, val, attr.Inverse, obj, context, true)
	   }
	}	
	return
}

// attributes map[*AttributeSpec] map[RObject] RObject

/*
Typechecked adding a member to a multi-valued attribute.
TODO TODO TODO

Create the collection on demand.

context is a context for evaluating a sorting comparison operator on the collection.

*/
func (rt *RuntimeEnv) AddToAttr(th InterpreterThread, obj RObject, attr *AttributeSpec, val RObject, typeCheck bool, context MethodEvaluationContext, isInverse bool) (err error) {

	if typeCheck && !val.Type().LessEq(attr.Part.Type) {
		err = fmt.Errorf("Cannot assign  '%v.%v %v' a value of type '%v'.", obj.Type(), attr.Part.Name, attr.Part.Type, val.Type())
		return
	}
	
	if obj == NIL {
		err = fmt.Errorf("nil cannot have a value added to '%v' attribute.", attr.Part.Name)
		return		
	}
	
	// Note: Need to put in a check here as to whether the collection accepts NIL elements, and
	// if val == NIL, reject the addition.	

	objColl, err := rt.EnsureMultiValuedAttributeCollection(obj, attr)
	if err != nil {
		return
	}

	addColl := objColl.(AddableMixin)     // Will throw an exception if collection type does not implement Add(..)
	added, newLen := addColl.Add(val, context) // returns false if is a set and val is already a member.

	/* TODO figure out efficient persistence of collection updates
	 */
	//fmt.Printf("added=%v\n",added)
	//fmt.Printf("IsStoredLocally=%v\n",obj.IsStoredLocally())

	if added {
	    if obj.IsStoredLocally() {
			var insertIndex int
			if objColl.(RCollection).IsSorting() {
				orderedColl := objColl.(OrderedCollection)
				insertIndex = orderedColl.Index(val, 0)
			} else {
				insertIndex = newLen - 1
			}
			th.DB().PersistAddToAttr(obj, attr, val, insertIndex)
		}
		if ! isInverse && attr.Inverse != nil {
			err = rt.SetOrAddToAttr(th, val, attr.Inverse, obj, context, true)
		}
	}

	return
}

/*
   Used to make val the value or a value of the attribute of obj. 
   If the attribute is multi-valued, val will be added to the collection of values.
   If the attribute is single-valued, val will be set as the value of the attribute. 
*/
func (rt *RuntimeEnv) SetOrAddToAttr(th InterpreterThread, obj RObject, attr *AttributeSpec, val RObject, context MethodEvaluationContext, isInverse bool) (err error) {
	
   if attr.Part.CollectionType == "" {	
      err = rt.SetAttr(th, obj, attr, val, false, context, isInverse) 
   } else {	
      err = rt.AddToAttr(th, obj, attr, val, false, context, isInverse)
   }	
   return
}


/*
 Remove all elements of the multivalued attribute, in memory and in the db.
 If the attribute has an inverse, also removes the inverse attribute values.
*/
func (rt *RuntimeEnv) ClearAttr(th InterpreterThread, obj RObject, attr *AttributeSpec) (err error) {

    objColl, foundCollection := rt.AttrVal(obj, attr)

 	if !foundCollection { // this object does not have the collection implementation of this multi-valued attribute	
                         // Must be already empty or unassigned?	
	  return
    }
		
	if attr.IsRelation() {
	   inverseAttr := attr.Inverse

       collection := objColl.(RCollection)
	   for val := range collection.Iter(th) {
           rt.RemoveAttrGeneral(th, val, inverseAttr, obj, true, false)		
	   }
    }

	collection := objColl.(RemovableMixin) // Will throw an exception if collection type does not implement ClearInMemory()
	collection.ClearInMemory()	
	
	
	err = th.DB().PersistClearAttr(obj, attr)
	return
}



/*
TODO Optimize this  to add all at once with a slice copy or similar, then persist in fewer
separate DB calls.
*/
func (rt *RuntimeEnv) ExtendCollectionTypeChecked(coll RCollection, vals []RObject, context MethodEvaluationContext) (err error) {

    for _,val := range vals {
       err = rt.AddToCollectionTypeChecked(coll, val, context)	
       if err != nil {
	      return
       }
    }
    return
}


func (rt *RuntimeEnv) AddToCollectionTypeChecked(coll RCollection, val RObject, context MethodEvaluationContext) (err error) {

	if !val.Type().LessEq(coll.ElementType()) {
		err = fmt.Errorf("Cannot add a value of type '%v' to a collection with element-type constraint '%v'.", val.Type(),coll.ElementType())
		return
	}
	
	addColl := coll.(AddableMixin)     // Will throw an exception if collection type does not implement Add(..)
	
	// Re-enable when we implement persisting of independent collections.
	// added, newLen := addColl.Add(val, context) // returns false if is a set and val is already a member.

	addColl.Add(val, context) // returns false if is a set and val is already a member.	

/*
Need to decide how to persist collections and check if persisted and handle persisting add

	TODO figure out efficient persistence of collection updates
	
	//fmt.Printf("added=%v\n",added)
	//fmt.Printf("IsStoredLocally=%v\n",obj.IsStoredLocally())

    // This part is COPYITIS from AddToAttrTypeChecked method.
	if added && obj.IsStoredLocally() {
		var insertIndex int
		if objColl.(RCollection).IsSorting() {
			orderedColl := objColl.(OrderedCollection)
			insertIndex = orderedColl.Index(val, 0)
		} else {
			insertIndex = newLen - 1
		}
		rt.db.PersistAddToAttr(obj, attr, val, insertIndex)
	}

	*/

	return
}

/*
TODO Optimize this  to add all at once with a slice copy or similar, then persist in fewer
separate DB calls.
*/
func (rt *RuntimeEnv) ExtendMapTypeChecked(theMap Map, keysVals []RObject, context MethodEvaluationContext) (err error) {

    n := len(keysVals)
    for i := 0; i < n; i+=2 {
       key := keysVals[i]
       val := keysVals[i+1]
       err = rt.PutInMapTypeChecked(theMap, key, val, context)	
       if err != nil {
	      return
       }
    }
    return
}

func (rt *RuntimeEnv) PutInMapTypeChecked(theMap Map, key RObject, val RObject, context MethodEvaluationContext) (err error) {

	if !key.Type().LessEq(theMap.KeyType()) {
		err = fmt.Errorf("Cannot use a key of type '%v' in a map with key-type constraint '%v'.", key.Type(),theMap.KeyType())
		return
	}
	
	if !val.Type().LessEq(theMap.ValType()) {
		err = fmt.Errorf("Cannot put a value of type '%v' in a map with value-type constraint '%v'.", val.Type(),theMap.ValType())
		return
	}	

    theMap.Put(key, val, context)
	
/*
Need to decide how to persist collections and check if persisted and handle persisting add

	TODO figure out efficient persistence of map collection updates
	
	//fmt.Printf("added=%v\n",added)
	//fmt.Printf("IsStoredLocally=%v\n",obj.IsStoredLocally())

    // This part is COPYITIS from AddToAttrTypeChecked method.
	if added && obj.IsStoredLocally() {
		var insertIndex int
		if objColl.(RCollection).IsSorting() {
			orderedColl := objColl.(OrderedCollection)
			insertIndex = orderedColl.Index(val, 0)
		} else {
			insertIndex = newLen - 1
		}
		rt.db.PersistAddToAttr(obj, attr, val, insertIndex)
	}

	*/

	return
}



/*
Helper method. Ensures that a collection exists in memory to manage the values of a multi-valued attribute of an object.
*/
func (rt *RuntimeEnv) EnsureMultiValuedAttributeCollection(obj RObject, attr *AttributeSpec) (collection RCollection, err error) {

    //defer Un(Trace(ALWAYS_,"EnsureMultiValuedAttributeCollection", attr.Part.Name))
	attrVals, found := rt.attributes[attr]
	if !found { // No assignment has ever happened (in this runtime) of a value to this attribute.
		attrVals = make(map[RObject]RObject)
		rt.attributes[attr] = attrVals
	}

	objColl, foundCollection := attrVals[obj]
	if !foundCollection { // this object does not already have the collection implementation of this multi-valued attribute
		var owner RObject
		var minCardinality, maxCardinality int64
		if attr.Part.ArityHigh == 1 { // This is a collection-valued attribute of arity 1. (1 collection)
			minCardinality = 0
			maxCardinality = -1 // largest possible collection is allowed - the attribute is not constraining it.	
			panic("Should not rt.EnsureMultiValuedAttributeCollection on a collection-valued attribute.")
		} else { // This is a multi-valued attribute. The collection is a hidden implementation detail. 
			minCardinality = int64(attr.Part.ArityLow)
			maxCardinality = int64(attr.Part.ArityHigh)
			owner = obj //  	Collection is owned by the "whole" object.

		}
		// Create the list or set collection

		var sortWith *sortOp
		var unaryMethod *RMultiMethod
		var binaryMethod *RMultiMethod
		var orderAttr *AttributeSpec

		// fmt.Println(attr.Part.CollectionType)

		if attr.Part.CollectionType == "sortedlist" || attr.Part.CollectionType == "sortedset" || attr.Part.CollectionType == "sortedmap" || attr.Part.CollectionType == "sortedstringmap" {
			if attr.Part.OrderMethodArity == 1 {
				unaryMethod = attr.Part.OrderMethod
			} else if attr.Part.OrderMethodArity == 2 {
				binaryMethod = attr.Part.OrderMethod
			} else { // must be an attribute
				orderAttr = attr.Part.OrderAttr
			}				

			if binaryMethod == nil {
				binaryMethod, _ = rt.InbuiltFunctionsPackage.MultiMethods["lt"]
			}

			sortWith = &sortOp{
				attr:          orderAttr,
				unaryFunction: unaryMethod,
				lessFunction:  binaryMethod,
				descending:    !attr.Part.IsAscending,
			}
		}

		/*
			@@@@@@@@@@@@@@@@@@@@@@@

			NEED TO FIND OUT IF THERE IS A UNARY METHOD OF THE TYP2, otherwise it must be a binary method.

			Only one of the attr or unaryFunction will be non-nil.
			If attr or unaryFunction is non-nil, then lessFunction must be the "lt" multiMethod.

			collection.sortWith.lessFunction,_ := RT.MultiMethods["lt"]

			If attr and unaryFunction are nil, lessFunction may be any binary boolean function which has a method whose
			parameter signature is compatible with a pair of values of the elementType of the collection. lessFunction MAY
			be the "lt" function in this case but need not be. The function is treated as a "less-than" predicate.

			type sortOp struct {
				attr *AttributeSpec
				unaryFunction *RMultiMethod
				lessFunction *RMultiMethod
				descending bool
			}

			$$$$$$$$$$$$$$$$$$$$$$$$


			  attr = &AttributeSpec{typ1,
			                        RelEnd{
			 									    Name:endName2,
			                                       Type:typ2,
			                                       ArityLow:arityLow2,
			                                       ArityHigh:arityHigh2,
			                                       CollectionType:collectionType2,
			                                       OrderAttr:orderAttrName,
			                                       OrderMethod: orderMethod,
			 									   OrderMethodArity: int32,
			                                       IsAscending:isAscending,
			                                      },

			@@@@@@@@@@@@@@
		*/

		switch attr.Part.CollectionType {
		case "list", "sortedlist":
			objColl, err = rt.Newrlist(attr.Part.Type, minCardinality, maxCardinality, owner, sortWith)
		case "set":
			objColl, err = rt.Newrset(attr.Part.Type, minCardinality, maxCardinality, owner)
		case "sortedset":
			objColl, err = rt.Newrsortedset(attr.Part.Type, minCardinality, maxCardinality, owner, sortWith)
		default:
			panic("I don't handle map attributes yet.")
		}

		attrVals[obj] = objColl
	}
	collection = objColl.(RCollection)
	return
}

/*
Removes val from the multi-valued attribute if val is in the collection. Does nothing and does not complain if val is not in the collection.
If removePersistent is true, also removes the value from the persistent version of the attribute association.
*/
func (rt *RuntimeEnv) RemoveFromAttr(th InterpreterThread, obj RObject, attr *AttributeSpec, val RObject, isInverse bool, removePersistent bool) (err error) { 	
	
	objColl, foundCollection := rt.AttrVal(obj, attr)

	if !foundCollection { // this object does not have the collection implementation of this multi-valued attribute		
		return
	}

	collection := objColl.(RemovableMixin) // Will throw an exception if collection type does not implement Remove(..)
	removed, removedIndex := collection.Remove(val)



   
	
	if removed  {
	   if removePersistent && obj.IsStoredLocally() {
	    	th.DB().PersistRemoveFromAttr(obj, attr, val, removedIndex)
	   }
	
	   if ! isInverse && attr.Inverse != nil {
	   	  err = rt.RemoveAttrGeneral(th, val, attr.Inverse, obj, true, removePersistent)
	   }
	}

	return
}




/*

Unsets the attribute. Used when setting to nil.
Deletes the corresponding row from the attribute database table if it exists.
Ok to call this even if attribute had no value.
If removePersistent, removes the attribute association from the database if it was persisted.
*/
func (rt *RuntimeEnv) UnsetAttr(th InterpreterThread, obj RObject, attr *AttributeSpec, isInverse bool, removePersistent bool) (err error) {

    if attr.Part.Type.IsPrimitive {
       err = fmt.Errorf("Cannot assign primitive-valued attribute '%v.%v %v' a value of nil", obj.Type(), attr.Part.Name, attr.Part.Type)
       return
    }

	attrVals, found := rt.attributes[attr]
    var val RObject

	if found {
		val, found = attrVals[obj]
    }

	if ! found {
		if obj.IsStoredLocally() {
			val, found = rt.AttrVal(obj, attr)
			attrVals = rt.attributes[attr]
		}
	}

	if found {
	   delete(attrVals,obj)

       if removePersistent && obj.IsStoredLocally() {
	          err = th.DB().PersistRemoveAttr(obj, attr) 	 
       }
       

       if ! isInverse && attr.Inverse != nil && val != NIL {
           err = rt.RemoveAttrGeneral(th, val, attr.Inverse, obj, true, removePersistent)
       }
    }	
	
    return
}

/*
If the attribute is multi-valued, removes the val from it, otherwise
if single valued, unsets the attribute.
*/
func (rt *RuntimeEnv) RemoveAttrGeneral(th InterpreterThread, obj RObject, attr *AttributeSpec, val RObject, isInverse bool, removePersistent bool) (err error) {
   if attr.Part.CollectionType == "" {	
	  err = rt.UnsetAttr(th, obj, attr, isInverse, removePersistent)
   } else {
      err = rt.RemoveFromAttr(th, obj, attr, val, isInverse, removePersistent)	
   }
   return
}


/*
type AttributeSpec struct {
   WholeType *RType
   Part RelEnd
   IsTransient bool
}


One end of a relation - specifies arity and type constraints and a few other details.

type RelEnd struct {
   Name string
   Type *RType
   ArityLow int32
   ArityHigh int32
   CollectionType string // "list", "sortedlist", "set", "sortedset", "map", "sortedmap", ""
   OrderAttrName string   // which primitive attribute of other is it ordered by when retrieving? "" if none

   OrderMethod *RMultiMethod

   Protection string // "public" "protected" "package" "private"
   DependentPart bool // delete of parent results in delete of attribute value
}
*/

/*
Can only be used in between a SetEvalContext and UnsetEvalContext call.
*/
func (rt *RuntimeEnv) GetEvalContext(obj RObject) MethodEvaluationContext {
	return rt.evalContexts[obj]
}

/*
You MUST call UnsetEvalContext after calling this!!
*/
func (rt *RuntimeEnv) SetEvalContext(obj RObject, context MethodEvaluationContext) {
	rt.evalContextMutex.Lock()
	rt.evalContexts[obj] = context
}

func (rt *RuntimeEnv) UnsetEvalContext(obj RObject) {
	delete(rt.evalContexts, obj)
	rt.evalContextMutex.Unlock()
}

// Usage
// RT.SetEvalContext(obj, context)
// defer RT.UnsetEvalContext(obj)
// context := RT.GetEvalContext(obj) 

type MethodEvaluationContext interface {

	/*
	   Evaluates a call of the (single-valued) multimethod on the argument objects.
	   Returns the result of the method call.
	*/
	EvalMultiMethodCall(mm *RMultiMethod, args []RObject) RObject
	
	/*
	   Return the interpreter thread aspect of the evaluation context.
	*/
    InterpThread() InterpreterThread 	
}

type Dispatcher interface {

	/*
	   The main dispatch function.
	   First looks up a cache (map) of method implementations keyed by typetuples.
	   a method will be found in this cache if the type tuple of the arguments
	   has had the multimethod called on it before.
	   If there is a cache miss, uses a multi-argument dynamic dispatch algorithm
	   to find the best-matching method implementation (then caches the find under
	   the type-tuple of the arguments for next time.)
	   Returns the best method implementation for the types of the argument objects,
	   or nil if the multimethod has no method signature which is compatible with
	   the types of the argument objects.
	   Also returns the type-tuple of the argument objects, which can be used to
	   report the lack of a compatible method.
	*/
	GetMethod(mm *RMultiMethod, args []RObject) (*RMethod, *RTypeTuple)

	/*
	   Same as GetMethod but for types instead of object instances.
	*/
	GetMethodForTypes(mm *RMultiMethod, types ...*RType) (*RMethod, *RTypeTuple)
}
