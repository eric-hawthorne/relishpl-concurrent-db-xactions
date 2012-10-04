// Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU LESSER GPL v3 license, found in the LICENSE_LGPL3 file.

package builtin

/*
   builtin_functions.go - functions which require no package specification. They are visible in every package.
*/

import (
	//"os"
	"fmt"
	"relish"
	. "relish/runtime/data"
	"strconv"
	"strings"
	"relish/rterr"
)

func InitBuiltinFunctions() {

	printMethod, err := RT.CreateMethod("","print", []string{"p"}, []string{"RelishPrimitive"}, 0, 0, false)
	if err != nil {
		panic(err)
	}
	printMethod.PrimitiveCode = builtinPrint

	print2Method, err := RT.CreateMethod("","print", []string{"p1", "p2"}, []string{"RelishPrimitive", "RelishPrimitive"}, 0, 0, false)
	if err != nil {
		panic(err)
	}
	print2Method.PrimitiveCode = builtinPrint

	print3Method, err := RT.CreateMethod("","print", []string{"p1", "p2", "p3"}, []string{"RelishPrimitive", "RelishPrimitive", "RelishPrimitive"}, 0, 0, false)
	if err != nil {
		panic(err)
	}
	print3Method.PrimitiveCode = builtinPrint

	print4Method, err := RT.CreateMethod("","print", []string{"p1", "p2", "p3", "p4"}, []string{"RelishPrimitive", "RelishPrimitive", "RelishPrimitive", "RelishPrimitive"}, 0, 0, false)
	if err != nil {
		panic(err)
	}
	print4Method.PrimitiveCode = builtinPrint

	print5Method, err := RT.CreateMethod("","print", []string{"p1", "p2", "p3", "p4", "p5"}, []string{"RelishPrimitive", "RelishPrimitive", "RelishPrimitive", "RelishPrimitive", "RelishPrimitive"}, 0, 0, false)
	if err != nil {
		panic(err)
	}
	print5Method.PrimitiveCode = builtinPrint

	print6Method, err := RT.CreateMethod("","print", []string{"p1", "p2", "p3", "p4", "p5", "p6"}, []string{"RelishPrimitive", "RelishPrimitive", "RelishPrimitive", "RelishPrimitive", "RelishPrimitive", "RelishPrimitive"}, 0, 0, false)
	if err != nil {
		panic(err)
	}
	print6Method.PrimitiveCode = builtinPrint

	print7Method, err := RT.CreateMethod("","print", []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7"}, []string{"RelishPrimitive", "RelishPrimitive", "RelishPrimitive", "RelishPrimitive", "RelishPrimitive", "RelishPrimitive", "RelishPrimitive"}, 0, 0, false)
	if err != nil {
		panic(err)
	}
	print7Method.PrimitiveCode = builtinPrint

	eqMethod, err := RT.CreateMethod("","eq", []string{"p1", "p2"}, []string{"RelishPrimitive", "RelishPrimitive"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	eqMethod.PrimitiveCode = builtinEq

	ltNumMethod, err := RT.CreateMethod("","lt", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	ltNumMethod.PrimitiveCode = builtinLtNum

	ltStrMethod, err := RT.CreateMethod("","lt", []string{"p1", "p2"}, []string{"String", "String"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	ltStrMethod.PrimitiveCode = builtinLtStr
	
	gtNumMethod, err := RT.CreateMethod("","gt", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	gtNumMethod.PrimitiveCode = builtinGtNum

	gtStrMethod, err := RT.CreateMethod("","gt", []string{"p1", "p2"}, []string{"String", "String"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	gtStrMethod.PrimitiveCode = builtinGtStr	

	timesMethod, err := RT.CreateMethod("","times", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	timesMethod.PrimitiveCode = builtinTimes

	dubMethod, err := RT.CreateMethod("","dub", []string{"obj", "name"}, []string{"NonPrimitive", "String"}, 0, 0, false)
	if err != nil {
		panic(err)
	}
	dubMethod.PrimitiveCode = builtinDub

	summonMethod, err := RT.CreateMethod("","summon", []string{"name"}, []string{"String"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	summonMethod.PrimitiveCode = builtinSummon
	
	//////////////////////////////////////////////////////////////////////////////////////
	// String methods - TODO Put them in their own package
	
	// contains s String substr String > Bool	
	//
	stringContainsMethod, err := RT.CreateMethod("","contains", []string{"s","substr"}, []string{"String","String"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringContainsMethod.PrimitiveCode = builtinStringContains	
	
	/*

	containsAny s String chars String > Bool

	containsRune s String r Rune > bool

	count s String sep String > Int  // or Int32 or UInt or UInt32 ? Which? Why?

	equalFold s String t String > Bool

	fields s String > []String

	// func FieldsFunc(s string, f func(rune) bool) []string

	hasPrefix s String prefix String > Bool

	hasSuffix s String suffix String > Bool

	index s String t String > Int
	"""
	 index of first occurrence of t in s or -1 if t not found in s 
	"""

	indexAny s String chars String > Int

	// func IndexFunc(s string, f func(rune) bool) int

	indexRune s string r Rune > Int

	join a []string sep String > String	
	
	*/
	
	
	
}

func builtinPrint(objects []RObject) []RObject {
	for i, obj := range objects {
		if i > 0 {
			fmt.Print(" ")
		}
		switch obj.(type) {
		case Int, Int32, Float, Bool, String:
			fmt.Print(obj.String())
		default:
			fmt.Print(obj.String()) // Do something else here. TODO
		}
	}
	fmt.Println()
	return nil
}

/*
Equality operator.
TODO !!!
Really have to study which kinds of equality we want to support and with which function names,
by studying other languages. LISP, Java, Go, Python etc.
*/
func builtinEq(objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	var val RObject
	switch obj1.(type) {
	case Int:
		switch obj2.(type) {
		case Int:
			val = Bool(obj1.(Int) == obj2.(Int))
		case Int32:
			val = Bool(int64(obj1.(Int)) == int64(obj2.(Int32)))
		case String:
			val = Bool(string(obj2.(String)) == strconv.FormatInt(int64(obj1.(Int)), 10))
		default:
			val = Bool(false)
		}
	case Int32:
		switch obj2.(type) {
		case Int32:
			val = Bool(obj1.(Int32) == obj2.(Int32))
		case Int:
			val = Bool(int64(obj1.(Int32)) == int64(obj2.(Int)))
		case String:
			val = Bool(string(obj2.(String)) == strconv.Itoa(int(obj1.(Int32))))
		default:
			val = Bool(false)
		}
	case String:
		switch obj2.(type) {
		case String:
			val = Bool(string(obj1.(String)) == string(obj2.(String)))
		case Int:
			val = Bool(string(obj1.(String)) == strconv.FormatInt(int64(obj2.(Int)), 10))
		case Int32:
			val = Bool(string(obj1.(String)) == strconv.Itoa(int(obj2.(Int32))))
		default:
			val = Bool(false)
		}
	}
	return []RObject{val}
}

/*
Less-than operator for numeric types.
TODO !!!
Really have to study which kinds of order we want to support and with which function names,
by studying other languages. LISP, Java, Go, Python etc.
*/
func builtinLtNum(objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	var val RObject
	switch obj1.(type) {
	case Int:
		switch obj2.(type) {
		case Int:
			val = Bool(int64(obj1.(Int)) < int64(obj2.(Int)))
		case Int32:
			val = Bool(int64(obj1.(Int)) < int64(obj2.(Int32)))
		case Float:
			val = Bool(float64(obj2.(Float)) < float64(obj1.(Int)))
		default:
			rterr.Stop("lt is not defined for argument types")
		}
	case Int32:
		switch obj2.(type) {
		case Int32:
			val = Bool(int32(obj1.(Int32)) < int32(obj2.(Int32)))
		case Int:
			val = Bool(int64(obj1.(Int32)) < int64(obj2.(Int)))
		case Float:
			val = Bool(float64(obj2.(Float)) < float64(obj1.(Int32)))
		default:
			rterr.Stop("lt is not defined for argument types")
		}
	case Float:
		switch obj2.(type) {
		case Float:
			val = Bool(float64(obj1.(Float)) < float64(obj2.(Float)))
		case Int:
			val = Bool(float64(obj1.(Float)) < float64(obj2.(Int)))
		case Int32:
			val = Bool(float64(obj1.(Float)) < float64(obj2.(Int32)))
		default:
			rterr.Stop("lt is not defined for argument types")
		}
	default: rterr.Stop("lt is not defined for argument types")
	}
	return []RObject{val}
}



/*
Less-than operator (strings).
TODO !!!
Really have to study which kinds of order we want to support and with which function names,
by studying other languages. LISP, Java, Go, Python etc.
*/
func builtinLtStr(objects []RObject) []RObject {
	obj1 := objects[0].(String)
	obj2 := objects[1].(String)
	var val RObject
	val = Bool(string(obj1) < string(obj2))
	return []RObject{val}
}

/*
Less-than operator for numeric types.
TODO !!!
Really have to study which kinds of order we want to support and with which function names,
by studying other languages. LISP, Java, Go, Python etc.
*/
func builtinGtNum(objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	var val RObject
	switch obj1.(type) {
	case Int:
		switch obj2.(type) {
		case Int:
			val = Bool(int64(obj1.(Int)) > int64(obj2.(Int)))
		case Int32:
			val = Bool(int64(obj1.(Int)) > int64(obj2.(Int32)))
		case Float:
			val = Bool(float64(obj2.(Float)) > float64(obj1.(Int)))
		default:
		rterr.Stop("gt is not defined for argument types")
		}
	case Int32:
		switch obj2.(type) {
		case Int32:
			val = Bool(int32(obj1.(Int32)) > int32(obj2.(Int32)))
		case Int:
			val = Bool(int64(obj1.(Int32)) > int64(obj2.(Int)))
		case Float:
			val = Bool(float64(obj2.(Float)) > float64(obj1.(Int32)))
		default:
		rterr.Stop("gt is not defined for argument types")
		}
	case Float:
		switch obj2.(type) {
		case Float:
			val = Bool(float64(obj1.(Float)) > float64(obj2.(Float)))
		case Int:
			val = Bool(float64(obj1.(Float)) > float64(obj2.(Int)))
		case Int32:
			val = Bool(float64(obj1.(Float)) > float64(obj2.(Int32)))
		default:
		rterr.Stop("gt is not defined for argument types")
		}
		default: rterr.Stop("gt is not defined for argument types")	
	}
	return []RObject{val}
}

/*
Greater-than operator (strings).
TODO !!!
Really have to study which kinds of order we want to support and with which function names,
by studying other languages. LISP, Java, Go, Python etc.
*/
func builtinGtStr(objects []RObject) []RObject {
	obj1 := objects[0].(String)
	obj2 := objects[1].(String)
	var val RObject
	val = Bool(string(obj1) > string(obj2))
	return []RObject{val}
}


/*
Multiply operator.
TODO !!!
Really have to study which kinds of equality we want to support and with which function names,
by studying other languages. LISP, Java, Go, Python etc.
*/
func builtinTimes(objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	var val RObject
	switch obj1.(type) {
	case Int:
		switch obj2.(type) {
		case Int:
			val = Int(int64(obj1.(Int)) * int64(obj2.(Int)))
		case Int32:
			val = Int(int64(obj1.(Int)) * int64(obj2.(Int32)))
		case Float:
			val = Float(float64(obj2.(Float)) * float64(obj1.(Int)))
		default:
		    rterr.Stop("times is not defined for argument types")
		}
	case Int32:
		switch obj2.(type) {
		case Int32:
			val = Int(int32(obj1.(Int32)) * int32(obj2.(Int32)))
		case Int:
			val = Int(int64(obj1.(Int32)) * int64(obj2.(Int)))
		case Float:
			val = Float(float64(obj2.(Float)) * float64(obj1.(Int32)))
		default:
		    rterr.Stop("times is not defined for argument types")
		}
	case Float:
		switch obj2.(type) {
		case Float:
			val = Float(float64(obj1.(Float)) * float64(obj2.(Float)))
		case Int:
			val = Float(float64(obj1.(Float)) * float64(obj2.(Int)))
		case Int32:
			val = Float(float64(obj1.(Float)) * float64(obj2.(Int32)))
		default:
		    rterr.Stop("times is not defined for argument types")
		}
	default:
		rterr.Stop("times is not defined for argument types")
	}
	return []RObject{val}
}

/*
dub NonPrimitive String 

dub car1 "FEC 092"

Gives the object a name in the local object-persistence database.
This confers persistence on the object.

Maybe should return the DBID eventually. Does not right now. TODO

Q: Is it an error to dub the same object more than once? Even if same name? Maybe yes because
it indicates that the programmer did not know the object was already persistent.

*/
func builtinDub(objects []RObject) []RObject {

	relish.EnsureDatabase()
	obj := objects[0]
	name := objects[1].String()

	if obj.IsStoredLocally() {
		rterr.Stopf("Object %v is already persistent. Cannot re-dub (re-persist) as '%s'.", obj, name)
	}

	// Ensure that the name is not already used for a persistent object in this database.

	found, err := RT.DB().ObjectNameExists(name)
	if err != nil {
		panic(err)
	} else if found {
		rterr.Stopf("The object name '%s' is already in use.", name)
	}

	// Ensure that the object is persisted.

	err = RT.DB().EnsurePersisted(obj)
	if err != nil {
		panic(err)
	}

	// Now we have to associate the object with its name in the database. 

	RT.DB().NameObject(obj, name)

	return nil
}

/*
summon String > NonPrimitive

car1 = Car summon "FEC 092"


*/
func builtinSummon(objects []RObject) []RObject {
	relish.EnsureDatabase()
	name := objects[0].String()

	obj, err := RT.DB().FetchByName(name, 0)
	if err != nil {
		panic(err)
	}

	return []RObject{obj}
}


/*

from ch InChannel of T > T

val = from ch

TODO DUMMY Implementation 
*/
func builtinFrom(objects []RObject) []RObject {
	name := objects[0].String()

	obj, err := RT.DB().FetchByName(name, 0)
	if err != nil {
		panic(err)
	}

	return []RObject{obj}
}


/*

to 
   ch InChannel of T 
   T

to ch val


TODO DUMMY Implementation 
*/
func builtinTo(objects []RObject) []RObject {
	name := objects[0].String()

	obj, err := RT.DB().FetchByName(name, 0)
	if err != nil {
		panic(err)
	}

	return []RObject{obj}
}


/*
contains operator (strings).

contains s String substr String > Bool	

*/
func builtinStringContains(objects []RObject) []RObject {
	s := objects[0].(String)
	substr := objects[1].(String)
	var val RObject
	val = Bool( strings.Contains(string(s), string(substr)) )
	return []RObject{val}
}
