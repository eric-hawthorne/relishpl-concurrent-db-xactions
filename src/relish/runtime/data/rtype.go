// Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.

package data

/*
   rtype.go - types attribute specifications and relation specifications, and 
              also various runtime-internal indexes on these.
*/

import (
	"fmt"
	"hash"
	"hash/adler32"
	. "relish/dbg"
)

///////////////////////////////////////////////////////////////////////////
////////// DATA TYPES
///////////////////////////////////////////////////////////////////////////

/*
The datatype of a Relish object.
Should it also have a direct map by occurrence pos to type tuples it occurs in?
So far no need for that identified.
*/
type RType struct {
	Name                    string
	shortName               string
	IsPrimitive             bool
	Parents                 []*RType
	Children                []*RType
	Up                      []*RType                  // chain of supertypes in dispatch order. (Python-esque?) 
	SpecializationPathNodes []*SpecializationPathNode // All the specialization-generalization chains this
	// type belongs to, with a reference to this type's node in each chain.
	Package                *RPackage
	Attributes             []*AttributeSpec
	ForwardRelations       []*RelationSpec
	ReverseRelations       []*RelationSpec
	AttributesByName       map[string]*AttributeSpec
	ForwardRelationsByName map[string]*RelationSpec
	ReverseRelationsByName map[string]*RelationSpec
	NumPrimitiveAttributes int
}

/*
   Returns true iff t is a strict subtype of t2.
*/
func (t *RType) Less(t2 *RType) bool {
	for _, ti := range t.Up {
		if ti == t2 {
			return true
		}
	}
	return false
}

/*
   Returns true iff t is identical to or a strict subtype of t2.
   This is the type assignment compatibility predicate.
*/
func (t *RType) LessEq(t2 *RType) bool {
	return (t == t2) || t.Less(t2)
}

/*
   Traverses the supertypes lattice to produce a list of all the supertypes in a
   partial order.
   Order is left to right breadth first.
*/
func computeUpChain(parents []*RType, ups []*RType, visited map[*RType]bool) []*RType {

	grandParents := make([]*RType, 0, len(parents)*3)
	for _, parent := range parents {
		if !visited[parent] {
			ups = append(ups, parent)
			grandParents = append(grandParents, parent.Parents...)
			visited[parent] = true
		}
	}
	if len(grandParents) > 0 {
		ups = computeUpChain(grandParents, ups, visited)
	}
	return ups
}

/*
   Notation:

   // List

   []Byte

   [ 19 213 54 77 ]

   // Set 

   {}Point

   { albatross b c d }

   // Map

   {}String=>Point

   { 
      albatross => b
      c         => d
   }

   // Ordered Set 

   {>}String
   {<}String
   {<lastName<firstName}Person
   {>seniority}Person                  // a unary function
   {?outRanks}Person                   // a binary function

   // Key-ordered Map

   {<}String=>Point

   { 
      albatross => b
      c         => d
      albatross => b
   }

For tests, the following are FALSE:
FALSE bool
0 Bit,Byte,Int8,Uint16,Int16,Uint32,Int32,Uint64,Int
0.0 Float32,Float 
"" String     utf8 code units




        f := *(*float)(unsafe.Pointer(&i)) 

		type word uint32 
		func (w word) float32() float32 { return math.Float32frombits(uint32(w)) } 
		func (w word) int32() int32 { return int32(w) } 
		func (w word) uint32() uint32 { return uint32(w) }

TODO The {} and [] names are not correct. These types are called List_of_Float,Set_of_Float etc.

*/
func newRType(name string, shortName string, parents []*RType) *RType {
	var isPrimitive bool
	switch name {
	//
	// Should we define Int for Int64 Uint for Uint64 and Float for Float64
	//
	case "Int", "Uint", "Int32", "Uint32", "Int16", "Uint16", "Int8", "Uint8", "Bit", "Bool", "CodePoint":
		fallthrough
	case "[]Int", "[]Uint", "[]Int32", "[]Uint32", "[]Int16", "[]Uint16", "[]Int8", "[]Byte", "[]Bit", "[]Bool", "[]CodePoint":
		fallthrough
	case "{}Int", "{}Uint", "{}Int32", "{}Uint32", "{}Int16", "{}Uint16", "{}Int8", "{}Byte", "{}CodePoint":
		fallthrough
	case "Float", "Float32", "Complex", "Complex32":
		fallthrough
	case "[]Float", "[]Float32", "[]Complex", "[]Complex32":
		fallthrough
	case "{}Float", "{}Float32", "{}Complex", "{}Complex32":
		fallthrough		
	case "Time":
		fallthrough		
	case "{}Time", "[]Time":		
		fallthrough
	case "String":
		isPrimitive = true

	default:
		isPrimitive = false
	}

	upChain := computeUpChain(parents, make([]*RType, 0, len(parents)+5), make(map[*RType]bool))
	Logln(GENERATE_, fmt.Sprintf("UPCHAIN for %s:", name))
	for _, typ := range upChain {
		Logln(GENERATE_, fmt.Sprint(typ))
	}
	Logln(GENERATE_, fmt.Sprintf("END OF UPCHAIN for %s", name))

    if shortName == "" {
	   shortName = name
    }

	typ := &RType{Name: name,
		shortName: shortName,
		IsPrimitive: isPrimitive,
		Parents:     parents,
		Children:    make([]*RType, 0, 5),
		Up:          upChain,
		SpecializationPathNodes: make([]*SpecializationPathNode, 0, 10),
		Attributes:              make([]*AttributeSpec, 0, 10),
		ForwardRelations:        make([]*RelationSpec, 0, 10),
		ReverseRelations:        make([]*RelationSpec, 0, 10),
		AttributesByName:        make(map[string]*AttributeSpec),
		ForwardRelationsByName:  make(map[string]*RelationSpec),
		ReverseRelationsByName:  make(map[string]*RelationSpec),
	}
	return typ
}

func (t RType) String() string {
	return t.Name
}

/*
A local runtime and local db unique name for the type.
*/
func (t RType) ShortName() string {
    return t.shortName
}

/*
A universally unique name for the type.

func (t RType) FullName() string {
	return t.Package.Name + "/" + t.Name
}
*/

/*
   Return the sqlite column type name of a primitive type.
   Note that we are using SQL types that might work in something like MySQL, PostgreSQL etc too.
*/
func (t *RType) DbColumnType() (sqlType string) {
	switch t {
	case IntType, UintType:
		sqlType = "BIGINT"
	case Int32Type, Uint32Type, CodePointType:
		sqlType = "INT"
	case Int16Type, Uint16Type:
		sqlType = "SMALLINT"
	case Int8Type, ByteType, BitType:
		sqlType = "TINYINT"
	case BoolType:
		sqlType = "BOOLEAN"
	case FloatType, ComplexType, Complex32Type:
		sqlType = "DOUBLE"
	case Float32Type:
		sqlType = "FLOAT"
	case StringType:
		sqlType = "TEXT"
	default:
		sqlType = ""
	}
	return
}

// DEPRECATED
func (t *RType) addForwardRelation(rel *RelationSpec) (err error) {
	_, found := t.ForwardRelationsByName[rel.End[1].Name]
	if found {
		err = fmt.Errorf("Attempt to redefine forward relation '%s' of type '%s'.", rel.End[1].Name, t.Name)
		return
	}
	t.ForwardRelations = append(t.ForwardRelations, rel)
	t.ForwardRelationsByName[rel.End[1].Name] = rel
	return
}

// DEPRECATED
func (t *RType) addReverseRelation(rel *RelationSpec) (err error) {
	_, found := t.ReverseRelationsByName[rel.End[0].Name]
	if found {
		err = fmt.Errorf("Attempt to redefine reverse relation '%s' of type '%s'.", rel.End[0].Name, t.Name)
		return
	}
	t.ReverseRelations = append(t.ReverseRelations, rel)
	t.ReverseRelationsByName[rel.End[0].Name] = rel
	return
}

/*
   Note that this does not check for attribute name overriding (hiding) of supertype attributes.
   We have to decide what is allowed here and what the semantics is.
   Do we do Eiffel-style aliasing of attribute names? Do we disallow overriding
   of an attribute? Do we allow creating an alternate getter/setter in subtype? If so,
   how do we combine the application of the getter/setters?
*/
func (t *RType) addAttribute(attr *AttributeSpec) (err error) {
/*	if len(t.Attributes) == 10 {
		err = fmt.Errorf("A type '%s' cannot have more than 10 attributes.", t.Name)
		return
	}
*/

    existingAttr, found := t.GetAttribute(attr.Part.Name)

// fmt.Printf("=== adding Attribute %s to type %s. Already found = %v\n",attr.Part.Name,t.Name,found)

//	_, found := t.AttributesByName[attr.Part.Name]
	if found {
		if existingAttr.WholeType == t {
		   err = fmt.Errorf("Attempt to redefine attribute '%s' of type '%s'.", attr.Part.Name, t.Name)			
		} else {
		   err = fmt.Errorf("Can't add attribute '%s' to type '%s'. Supertype '%s' has an attribute or relation of same name.", attr.Part.Name, t.Name, existingAttr.WholeType.Name)
	    }
		return
	}
	t.Attributes = append(t.Attributes, attr)
	t.AttributesByName[attr.Part.Name] = attr

	if attr.Part.Type.IsPrimitive {
		t.NumPrimitiveAttributes++
	}

	return
}

/*
This is a HIGHLY UNOPTIMIZED way of getting the attribute spec from the type given the attribute name.
It searches the type's AttributesByName hashtable, and if not found there, traverses the type's Up chain
of supertypes, searching each's  AttributesByName hashtable in turn until an attribute of the specified name
is found. Returns nil if not found

We have to see how we can cache this dispatch, similar to how we cache multi-method dispatch. TODO TODO
*/
func (t *RType) GetAttribute(attrName string) (attr *AttributeSpec, found bool) {
	attr, found = t.AttributesByName[attrName]
	if !found {
		for _, supertype := range t.Up {
			attr, found = supertype.AttributesByName[attrName]
			if found {
				return
			}
		}
	}
	return
}


/*
DEPRECATED

This is a HIGHLY UNOPTIMIZED way of getting the attribute spec from the type given the attribute name.
It searches the type's AttributesByName hashtable, and if not found there, traverses the type's Up chain
of supertypes, searching each's  AttributesByName hashtable in turn until an attribute of the specified name
is found. Returns nil if not found

We have to see how we can cache this dispatch, similar to how we cache multi-method dispatch. TODO TODO
*/
func (t *RType) GetRelation(relName string) (rel *RelationSpec, found bool, isForward bool) {
	rel, found = t.ForwardRelationsByName[relName]
	if !found {
		for _, supertype := range t.Up {
			rel, found = supertype.ForwardRelationsByName[relName]
			if found {
				isForward = true
				return
			}
		}
	}
	
	rel, found = t.ReverseRelationsByName[relName]
	if !found {
		for _, supertype := range t.Up {
			rel, found = supertype.ReverseRelationsByName[relName]
			if found {
				return
			}
		}
	}	
	
	return
}


/*
A specification of a type of object having an attribute whose value is a given allowable number of a given type
of objects.
*/
type AttributeSpec struct {
	WholeType   *RType
	Part        RelEnd
	IsTransient bool
	IsForwardRelation bool
	IsReverseRelation bool
	Inverse *AttributeSpec
}

func (attr *AttributeSpec) IsRelation() bool {
	return attr.IsForwardRelation || attr.IsReverseRelation
}

func (attr *AttributeSpec) IsOneWay() bool {
	return ! attr.IsForwardRelation && ! attr.IsReverseRelation
}


/*
One end of a relation - specifies arity and type constraints and a few other details.
*/
type RelEnd struct {
	Name           string
	Type           *RType
	ArityLow       int32
	ArityHigh      int32
	CollectionType string         ///  "list", "sortedlist","set", "sortedset", "map", "stringmap", "sortedmap","sortedstringmap" ""
	OrderAttr      *AttributeSpec // which primitive attribute of other is it ordered by when retrieving? nil if none

	OrderMethod      *RMultiMethod
	OrderMethodArity int32 // 1 or 2 if applicable
	IsAscending      bool  // ascending order if ordered collection? or descending order

	Protection    string // "public" "protected" "package" "private"
	DependentPart bool   // delete of parent results in delete of attribute value
}

/*
   Return the sqlite column definition of a primitive-type attribute.
*/
func (end RelEnd) DbColumnDef() (colDef string) {
	if end.Type == ComplexType {
		colDef = end.Name + "_r DOUBLE,\n" + end.Name + "_i DOUBLE"
	} else if end.Type == TimeType {
		colDef = end.Name + " TEXT,\n" + end.Name + "_loc TEXT"
	} else {
		colDef = end.Name + " "
		if end.ArityHigh == 1 {
			colDef += end.Type.DbColumnType()
		} else {
			colDef += "BLOB"
		}
	}
	return
}

/*
   The name for this attribute that will become the db table name.
   (only applicable if the attribute's Part Type is non-primitive.)

   Currently the name is e.g. "Cart___wheel__Wheel"
*/
func (attr *AttributeSpec) ShortName() string {
	return fmt.Sprintf("%s___%s__%s", attr.WholeType.ShortName(),
		attr.Part.Name, attr.Part.Type.ShortName())
}

/*

DEPRECATED

A specification of a relationship between two types.
Note that probably it is constrained to have set or sortedset as the collectiontype (why not list with order manually managed?)


Cart 0 1 cart -- horses 1 N Horse

end[0] -- end[1]

*/
type RelationSpec struct {
	End         [2]RelEnd
	IsTransient bool
}

/*
   The name for this relation that will become the db table name.
   TODO Consider only using the end[1].Name so that we can turn an attribute table into
   a relation table with no change.

   Currently the name is e.g. "Cart__cart__horse__Horse"
*/
func (rel *RelationSpec) ShortName() string {
	return fmt.Sprintf("%s__%s__%s__%s", rel.End[0].Type.ShortName(),
		rel.End[0].Name, rel.End[1].Name, rel.End[1].Type.ShortName())
}

/*
   Create a new relish type.
   The parent types are the direct supertypes of the type.
   The parent types are assumed to exist.
   Returns an error if a type with the name already exists. (What scope?)
   TODO How do we handle incremental compilation that includes redefinitions of types
*/
func (rt *RuntimeEnv) CreateType(typeName string, typeShortName string, parentTypeNames []string) (*RType, error) {
	if _, found := rt.Types[typeName]; found {
		return nil, fmt.Errorf("Attempt to redefine type '%s'.", typeName)
	}
	var parentTypes []*RType = make([]*RType, len(parentTypeNames))
	for i, parentTypeName := range parentTypeNames {
		if parentType, found := rt.Types[parentTypeName]; found {
			parentTypes[i] = parentType
		} else {
			return nil, fmt.Errorf("Defining type '%s' but parent type '%s' does not exist.", typeName, parentTypeName)
		}
	}
	typ := newRType(typeName, typeShortName, parentTypes)

	// Make or extend specialization paths.   
	if len(parentTypes) == 0 {
		// If the new type is a top (most general) type in the type lattice, give it a
		// top-most specialization-path node.
		//
		typ.startSpecializationPath()

	} else { // The new type has parent types

		for _, parentType := range parentTypes {
			if len(parentType.Children) == 0 {

				// If this new type is the first child of the parent type, 
				// then existing specialization-paths involving the parent type
				// will currently end at the parent type. i.e. bottom node in each
				// specialization-path is the parent type.
				// Extend each specialization-path that ends at the parent type
				// down to reach this new sub-type.
				// 
				for _, parentNode := range parentType.SpecializationPathNodes {
					parentNode.extendPath(typ)
				}
			} else {
				// The parent type of the new node already has other children types.
				// So that means the parent type's specialization paths do not
				// end at the parent type but continue down to each of its existing
				// children types.
				//
				// To accommodate the new sibling subtype, we need to create
				// copies of the parent type's specialization paths, or
				// more precisely, of the prefixes of the paths that end at the
				// parent type, then extend each copied prefix path down to reach
				// this new sub-type. 
				// Detail: See discussion below of how many copies
				// need to be made of the parent type's specialization paths.
				//

				// a map from "checksum" hash of specialization-path to specialization-path
				// used to determine the set of upwards-distinct specialization paths that
				// must thus be copied and extended down to the new subtype.
				//
				pathMap := make(map[uint32]*SpecializationPathNode)
				for _, parentNode := range parentType.SpecializationPathNodes {

					copiedNode, pathHash := parentNode.copyUpwards()

					// Now it may be the case that there are different paths up from 
					// the parent type which are identical to each other. They only
					// begin to differ lower down the specialization hierarchy than
					// the parent type. 
					// We only want to keep a copy of one path from each group of
					// upwards-equivalent paths. That one copy will be the new 
					// specialization path that goes down to the new type.
					//
					_, found := pathMap[pathHash]
					if !found {
						pathMap[pathHash] = copiedNode
					}
				}
				for _, copiedNode := range pathMap {
					copiedNode.glue()
					copiedNode.extendPath(typ)
				}
			}
		}
	}
	rt.Types[typeName] = typ
	rt.Typs[typ.ShortName()] = typ
	return typ, nil

}

/*
   Create if necessary and return the type representing a set of some element type.
*/
func (rt *RuntimeEnv) getSetType(elementType *RType) (typ *RType, err error) {
	typeName := "Set_of_" + elementType.Name
	typeShortName := "Set_of_" + elementType.ShortName()	
	typ, found := rt.Types[typeName]
	if !found {
		typ, err = rt.CreateType(typeName, typeShortName, []string{"Set"})
	}
	return
}

/*
   Create if necessary and return the type representing a list of some element type.
*/
func (rt *RuntimeEnv) getListType(elementType *RType) (typ *RType, err error) {
	typeName := "List_of_" + elementType.Name
	typeShortName := "List_of_" + elementType.ShortName()
	typ, found := rt.Types[typeName]
	if !found {
		typ, err = rt.CreateType(typeName, typeShortName, []string{"List"})
	}
	return
}

/*
Deprecated comment:
Each type should have a list (map from opposite end name?) of forwardRels (I am end 0) and backRels (where I am end 1)
*/
func (rt *RuntimeEnv) CreateRelation(typeName1 string,
	endName1 string,
	arityLow1 int32,
	arityHigh1 int32,
	collectionType1 string,
	orderFuncOrAttrName1 string,
	isAscending1 bool,	
	typeName2 string,
	endName2 string,
	arityLow2 int32,
	arityHigh2 int32,
	collectionType2 string,
	orderFuncOrAttrName2 string,
	isAscending2 bool,	
	isTransient bool,
	dispatcher Dispatcher) (type1 *RType, type2 *RType,err error) {
				


	forwardRelAttr,err := rt.CreateAttribute(typeName1,
									 	 typeName2,
										 endName2,
										 arityLow2,
										 arityHigh2,   // Is the -1 meaning N respected in here???? TODO
										 collectionType2,
				                         orderFuncOrAttrName2,
				                         isAscending2,
										 isTransient,
										 true,
										 false,
										 dispatcher)

   if err != nil {
       return
   }

   reverseRelAttr,err := rt.CreateAttribute(typeName2,
								 	 typeName1,
									 endName1,
									 arityLow1,
									 arityHigh1,   // Is the -1 meaning N respected in here???? TODO
									 collectionType1,
			                         orderFuncOrAttrName1,
			                         isAscending1,
									 isTransient,
									 false,
									 true,
									 dispatcher)


   if err != nil {
       return
   }

   forwardRelAttr.Inverse = reverseRelAttr
   reverseRelAttr.Inverse = forwardRelAttr

   type1 = forwardRelAttr.WholeType
   type2 = reverseRelAttr.WholeType

   return
}





/*
Each type should have a list (map from name?) of attributes.

I need to create several auto-generated primitive-implementation methods for each attribute
(getters, setters etc, and make them multimethods!!!!!)
Is this needed in order to get dispatch, or not yet?

*/
func (rt *RuntimeEnv) CreateAttribute(typeName1 string,
	typeName2 string,
	endName2 string,
	arityLow2 int32,
	arityHigh2 int32,
	collectionType2 string,
	orderFuncOrAttrName string,
	isAscending bool,
	isTransient bool,
	isForwardRelation bool,
	isReverseRelation bool,
	dispatcher Dispatcher) (attr *AttributeSpec, err error) {

	//fmt.Println(rt.Types)	
	typ1, found := rt.Types[typeName1]
	if !found {
		err = fmt.Errorf("Type '%s' not found.", typeName1)
		return
	}
	
	typ2, found := rt.Types[typeName2]
	if !found {
		err = fmt.Errorf("Type '%s' not found.", typeName2)
		return
	}

	var orderAttr *AttributeSpec
	var attrFound bool
	var orderMethod *RMultiMethod

	// TODO Find out which of attribute or method it is. How????
	// Should be along the lines of: 
	// If there is a method of that name defined which takes only typ2 as arg signature, then use that function,
	// Otherwise could check if there is an attribute of that name on typ2 or its supertypes, and if not,
	// throw a type compatibility panic.

	var orderMethodArity int32 = 0

	if orderFuncOrAttrName != "" {
		orderMethod, methodFound := rt.MultiMethods[orderFuncOrAttrName]

		if methodFound {

			// Find out if there is an arity 1 method
			unaryMethod, _ := dispatcher.GetMethodForTypes(orderMethod, typ2)
			// Find out if there is an arity 2 method
			binaryMethod, _ := dispatcher.GetMethodForTypes(orderMethod, typ2, typ2)

			if unaryMethod != nil && binaryMethod != nil {
				panic(fmt.Sprintf("Can't order collection because ambiguous unary and binary method '%s'.", orderFuncOrAttrName))
			}

			if unaryMethod != nil {
				orderMethodArity = 1
			} else if binaryMethod != nil {
				orderMethodArity = 2
			} else {
				panic(fmt.Sprintf("Can't order collection. No unary or binary method '%s' found.", orderFuncOrAttrName))
			}
		} else {
			orderAttr, attrFound = typ2.GetAttribute(orderFuncOrAttrName)
			if !attrFound {
				panic(fmt.Sprintf("Can't order collection. No method or attribute '%s' found.", orderFuncOrAttrName))
			}
		}
	}

	attr = &AttributeSpec{typ1,
		RelEnd{
			Name:             endName2,
			Type:             typ2,
			ArityLow:         arityLow2,
			ArityHigh:        arityHigh2,
			CollectionType:   collectionType2,
			OrderAttr:        orderAttr,
			OrderMethod:      orderMethod,
			OrderMethodArity: orderMethodArity, // 1 or 2
			IsAscending:      isAscending,
		},
		isTransient,
		isForwardRelation,
		isReverseRelation,
		nil,
	}
	err = typ1.addAttribute(attr)

	return
}

/*
RelEnd
   ...
   CollectionType string //  "list", "sortedlist","set", "sortedset", "map", "stringmap", "sortedmap","sortedstringmap" ""
   OrderAttrName string   // which primitive attribute of other is it ordered by when retrieving? "" if none

   OrderMethod *RMultiMethod
*/

/*
   Given a list of typeNames, return the unique RTypeTuple object representing that
   list of types.
   First has to look up the type objects by name in the runtime environment's Types map.

   This function is used during the definition of a method implementation for a particular
   argument-types signature.
*/
func (rt *RuntimeEnv) getTypeTupleFromTypes(typeNames []string) (*RTypeTuple, error) {

	mTypes := make([]*RType, len(typeNames))
	for i, typeName := range typeNames {
		typ, found := rt.Types[typeName]
		if !found {
			return nil, fmt.Errorf("Type '%v' not found.", typeName)
		}
		mTypes[i] = typ
	}
	return rt.TypeTupleTree.GetTypeTupleFromTypes(mTypes), nil
}

////////////////////////////////////////////////////////////////////////////////////////////////
//////// SPECIALIZATION PATHS - Used in the calculation of the absolute and relative specificity 
//////// of types and type tuples.
//////// Each specialization path is a single unique path (chain) from general to specific types 
//////// in the multiple-inheritance type lattice.
////////////////////////////////////////////////////////////////////////////////////////////////

/*
   Nodes in a chain (a doubly-linked list) representing a single path from a most general datatype
   down to a most specific datatype.
*/
type SpecializationPathNode struct {
	// DimID uint64 // May not need this.
	Up    *SpecializationPathNode
	Down  *SpecializationPathNode
	Level uint32
	Type  *RType
}

/*
   Return a copy of the specialization path from the node upwards.
   Also return a hash of the names of the types in the path from the
   node upwards.
   Note. The copy of the path is not yet "glued" to its types, in that
   while the path refers to the types, the types do not yet have a reference
   to the copied path. This is  because some copied paths produced when adding
   a new type into the inheritance structure are redundant and are thrown away.
*/
func (node *SpecializationPathNode) copyUpwards() (*SpecializationPathNode, uint32) {
	a32Hash := adler32.New()
	copiedNode := node.copyUpwards1(nil, a32Hash)
	pathHash := a32Hash.Sum32()
	return copiedNode, pathHash
}

/*
   Helper function for copyUpwards. Recursive.
*/
func (node *SpecializationPathNode) copyUpwards1(downNode *SpecializationPathNode, h hash.Hash32) *SpecializationPathNode {
	h.Write([]byte(node.Type.Name))
	copiedNode := &SpecializationPathNode{Up: nil, Down: downNode, Level: node.Level, Type: node.Type}
	if node.Up != nil {
		copiedUpNode := node.Up.copyUpwards1(copiedNode, h)
		copiedNode.Up = copiedUpNode
	}
	return copiedNode
}

/*
   Glues the specialization path to the types by giving the types a reference
   to their node in the path. Second stage of specializeation-path creation.
*/
func (node *SpecializationPathNode) glue() {
	for nd := node; nd != nil; nd = nd.Up {
		node.Type.SpecializationPathNodes = append(node.Type.SpecializationPathNodes, nd)
	}
}

/*
   Call this when creating a type with no parent types.
   Gives the type a top-most specialization-path-node.
   The node has specialization-level 0, thus declaring that the type is
   at the general top of the specialization lattice.
*/
func (t *RType) startSpecializationPath() {
	node := &SpecializationPathNode{Up: nil, Down: nil, Level: 0, Type: t}
	t.SpecializationPathNodes = append(t.SpecializationPathNodes, node)
}

/*
Extend the given specialization path downward to reach the new type.
This function should be called on the bottom-most node in the specialization path. 

For a newly copied specialization-path, glue it first then extend it.
*/
func (node *SpecializationPathNode) extendPath(typ *RType) {
	newNode := &SpecializationPathNode{Up: node, Down: nil, Level: node.Level + 1, Type: typ}
	node.Down = newNode
	typ.SpecializationPathNodes = append(typ.SpecializationPathNodes, newNode)
}

/* 
   A specialization path is a unique path down the type specialization lattice.

   Procedure for adding a new subtype:
   1. If the parent type's Children list is empty, 
      extend each path down to the new type.
   2. If the parent has children
      a. Copy all paths, down to the parent only.
         (Two steps: copy, then glue to types. Second step done after uniqueness testing.)
         (while doing so, make a hash or concatenation of the type names,
         and throw away a fresh copy if it has the same hash value.
      b. Extend all copy paths.

*/

////////////////////////////////////////////////////////////////////////////////////////////////
//////// TYPE TUPLES - lists of types that appear as formal parameters, and actual arguments
//////// in definitions, and calls, respectively, of Relish methods.
////////
//////// Part of the Relish language interpreter's multi-method dispatch system.
//////// 
//////// For each unique list of types which has appeared so far in a method definition or
//////// a method call, there will be a single RTypeTuple object, and the runtime environment
//////// maintains a tree index to rapidly find that tuple object given a list of types (or 
//////// typed objects) that is being used a second or subsequent time in a method definition or 
//////// call.
////////
////////////////////////////////////////////////////////////////////////////////////////////////

type RTypeTuple struct {
	Types []*RType
}

func (tt *RTypeTuple) String() string {
	s := "("
	sep := ""
	for _, typ := range tt.Types {
		s += sep + typ.Name
		sep = ","
	}
	s += ")"
	return s
}

/*
   TODO 
   Returns the specialization distance 
   (Squared Euclidian distance in a multi-dimensional type-specialization space) from
   the presumed more general type tuple (superTT) to the presumed more specific type 
   tuple (subTT).
   Distances are divided by the number of dimensions.
   The second result value is the Euclidean specialization depth 
   (specificity compared to topmost general types in the ontology) of the supertype tuple.
   Third result value is set to true if there is no compatible specialization from superTT
   down to subTT.

   Assumes the two type tuples are the same length (arity).
*/
func (subTT *RTypeTuple) SpecializationDistanceFrom(superTT *RTypeTuple) (float64, float64, bool) {
	if len(superTT.Types) == 0 { // Degenerate case. No types to compare.
		return 0.0, 0.0, false
	}
	var sumSquaredLevelDiff uint32 = 0
	var sumSquaredSupertypeSpecificity uint32 = 0
	var nDimensions uint32 = 0
	for i, subType := range subTT.Types {
		superType := superTT.Types[i]
		var foundOnePathForTypeSpecialization bool = false
		for _, pathNode := range subType.SpecializationPathNodes {
			nDimensions++
			node := pathNode
			var found bool = false
			var levelDiff uint32

			if superType == AnyType { // Special case
				levelDiff = 100
				found = true
				foundOnePathForTypeSpecialization = true
			} else if superType == NonPrimitiveType {
				if !node.Type.IsPrimitive { // Special case 
					levelDiff = 99
					found = true
					foundOnePathForTypeSpecialization = true
				}
			} else {

				for levelDiff = 0; ; levelDiff++ {
					if node.Type == superType {
						found = true
						foundOnePathForTypeSpecialization = true
						break
					} else if node.Up == nil {
						break
					} else {
						node = node.Up
					}
				}
			}
			if found {
				sumSquaredLevelDiff += levelDiff * levelDiff
				sumSquaredSupertypeSpecificity += node.Level * node.Level
			}
		}
		if !foundOnePathForTypeSpecialization {
			return -1.0, -1.0, true
		}
	}
	return float64(sumSquaredLevelDiff) / float64(nDimensions), float64(sumSquaredSupertypeSpecificity) / float64(nDimensions), false
}

////////////////////////////////////////////////////////////////////////////
//////// The tree index of type tuples
//
// How do we ensure there is only one typetuple object per
// tuple-signature in the runtime? Would seem to need a tree index structure
// to point to the type tuples. Root is leftmost type.
////////////////////////////////////////////////////////////////////////////

type TypeTupleTreeNode struct {
	mType    *RType                          // nil if root node - this is just documentation
	nextType map[*RType][]*TypeTupleTreeNode // empty map if end of tree
	tuple    *RTypeTuple                     // nil if no typetuple ends here.
}

/*
   Return a reference to process-wide unique type tuple representing the 
   list of types of the argument objects.
   TBD: If the tuple is not already created, create it and insert it in the 
   TypeTupleTree (an index that helps later calls of this method find the type tuple.)

   T1 T2 T3 T4

   NIL T1
         T3
         T2 tuple
         T2
           T4
         T2 
           T3 
             T4 
               T5
         T2 
           T3 
             T4

*/
func (tttn *TypeTupleTreeNode) GetTypeTuple(mObjects []RObject) *RTypeTuple {
	return tttn.findOrCreateTypeTuple(mObjects, mObjects)
}

/*
start with 
typeTupleTree :=  &RTypeTupleTreeNode{}
Create for T1 T2 

*/
func (tttn *TypeTupleTreeNode) findOrCreateTypeTuple(mObjects []RObject, allObjects []RObject) *RTypeTuple {
	if len(mObjects) == 0 {
		if tttn.tuple == nil {
			tttn.tuple = createTypeTuple(allObjects)
		}
		return tttn.tuple
	}

	typ := mObjects[0].Type()
	if tttn.nextType != nil { // we need to explore the tree from here.
		nextNodes, found := tttn.nextType[typ]
		if found {
			for _, node := range nextNodes {
				tpl := node.findOrCreateTypeTuple(mObjects[1:], allObjects)
				if tpl != nil {
					return tpl
				}
			}
			// If got here need to create a new nextType node in the 
			// list under key typ
			newNode := &TypeTupleTreeNode{mType: typ}
			tttn.nextType[typ] = append(tttn.nextType[typ], newNode)
		}
		// If got here need to create a list of nodes with a
		// single new node for the typ, and store the list under key typ
		newNode := &TypeTupleTreeNode{mType: typ}
		tttn.nextType[typ] = []*TypeTupleTreeNode{newNode}
	}
	// If got here need to make the nextType map and
	// create a list of nodes with a
	// single new node for the typ, and store the list under key typ

	newNode := &TypeTupleTreeNode{mType: typ}
	tttn.nextType = make(map[*RType][]*TypeTupleTreeNode)
	tttn.nextType[typ] = []*TypeTupleTreeNode{newNode}

	tpl := newNode.findOrCreateTypeTuple(mObjects[1:], allObjects)
	return tpl
}

/*
   Given a list of objects, return a RTypeTuple containing a list
   of the types of the objects.
*/
func createTypeTuple(mObjects []RObject) *RTypeTuple {
	tt := &RTypeTuple{Types: make([]*RType, len(mObjects))}
	for i, obj := range mObjects {
		tt.Types[i] = obj.Type()
	}
	return tt
}

func (tttn *TypeTupleTreeNode) GetTypeTupleFromTypes(mTypes []*RType) *RTypeTuple {
	return tttn.findOrCreateTypeTupleFromTypes(mTypes, mTypes)
}

/*
start with 
typeTupleTree :=  &RTypeTupleTreeNode{}
Create for T1 T2 

*/
func (tttn *TypeTupleTreeNode) findOrCreateTypeTupleFromTypes(mTypes []*RType, allTypes []*RType) *RTypeTuple {
	//fmt.Printf("findOrCreateTypeTupleForTypes %v %v",mTypes,allTypes)
	if len(mTypes) == 0 {
		if tttn.tuple == nil {
			tttn.tuple = createTypeTupleFromTypes(allTypes)
		}
		return tttn.tuple
	}
	typ := mTypes[0]
	if tttn.nextType != nil { // we need to explore the tree from here.
		nextNodes, found := tttn.nextType[typ]
		if found {
			for _, node := range nextNodes {
				tpl := node.findOrCreateTypeTupleFromTypes(mTypes[1:], allTypes)
				if tpl != nil {
					return tpl
				}
			}
			// If got here need to create a new nextType node in the 
			// list under key typ
			newNode := &TypeTupleTreeNode{mType: typ}
			tttn.nextType[typ] = append(tttn.nextType[typ], newNode)
		}
		// If got here need to create a list of nodes with a
		// single new node for the typ, and store the list under key typ
		newNode := &TypeTupleTreeNode{mType: typ}
		tttn.nextType[typ] = []*TypeTupleTreeNode{newNode}
	}
	// If got here need to make the nextType map and
	// create a list of nodes with a
	// single new node for the typ, and store the list under key typ

	newNode := &TypeTupleTreeNode{mType: typ}
	tttn.nextType = make(map[*RType][]*TypeTupleTreeNode)
	tttn.nextType[typ] = []*TypeTupleTreeNode{newNode}

	tpl := newNode.findOrCreateTypeTupleFromTypes(mTypes[1:], allTypes)
	return tpl
}

/*
   Given a list of objects, return a RTypeTuple containing a list
   of the types of the objects.
*/
func createTypeTupleFromTypes(mTypes []*RType) *RTypeTuple {
	//fmt.Println("Making type tuple for")
	//fmt.Println(mTypes)
	tt := &RTypeTuple{Types: make([]*RType, len(mTypes))}
	for i, typ := range mTypes {
		tt.Types[i] = typ
	}
	return tt
}
