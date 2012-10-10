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

    /////////////////////////////////////////////////////////
    // Numeric arithmetic functions
    //
    // Note: We have an issue here about return type
    //
    // Is the language going to allow co-variant return types in different methods of same multimethod?
    //
    // What does that imply for static type checking of return values?
    //

	timesMethod, err := RT.CreateMethod("","times", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	timesMethod.PrimitiveCode = builtinTimes


	plusMethod, err := RT.CreateMethod("","plus", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	plusMethod.PrimitiveCode = builtinPlus	

	minusMethod, err := RT.CreateMethod("","minus", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	minusMethod.PrimitiveCode = builtinMinus	

	divMethod, err := RT.CreateMethod("","div", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	divMethod.PrimitiveCode = builtinDiv	

	modMethod, err := RT.CreateMethod("","mod", []string{"p1", "p2"}, []string{"Integer", "Integer"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	modMethod.PrimitiveCode = builtinMod	



    ///////////////////////////////////////////////////////////
    // Persistence functions

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

	existsMethod, err := RT.CreateMethod("","exists", []string{"name"}, []string{"String"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	existsMethod.PrimitiveCode = builtinExists


    ///////////////////////////////////////////////////////////////////
    // Boolean Logical functions - not and or

	notMethod, err := RT.CreateMethod("","not", []string{"p"}, []string{"Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	notMethod.PrimitiveCode = builtinNot    



/*
   TODO Replace these with proper variadic function creations.
*/
	and2Method, err := RT.CreateMethod("","and", []string{"p1", "p2"}, []string{"Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	and2Method.PrimitiveCode = builtinAnd

	and3Method, err := RT.CreateMethod("","and", []string{"p1", "p2", "p3"}, []string{"Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	and3Method.PrimitiveCode = builtinAnd

	and4Method, err := RT.CreateMethod("","and", []string{"p1", "p2", "p3", "p4"}, []string{"Any", "Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	and4Method.PrimitiveCode = builtinAnd

	and5Method, err := RT.CreateMethod("","and", []string{"p1", "p2", "p3", "p4", "p5"}, []string{"Any", "Any", "Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	and5Method.PrimitiveCode = builtinAnd

	and6Method, err := RT.CreateMethod("","and", []string{"p1", "p2", "p3", "p4", "p5", "p6"}, []string{"Any", "Any", "Any", "Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	and6Method.PrimitiveCode = builtinAnd

	and7Method, err := RT.CreateMethod("","and", []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7"}, []string{"Any", "Any", "Any", "Any", "Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	and7Method.PrimitiveCode = builtinAnd

	

	or2Method, err := RT.CreateMethod("","or", []string{"p1", "p2"}, []string{"Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	or2Method.PrimitiveCode = builtinOr

	or3Method, err := RT.CreateMethod("","or", []string{"p1", "p2", "p3"}, []string{"Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	or3Method.PrimitiveCode = builtinOr

	or4Method, err := RT.CreateMethod("","or", []string{"p1", "p2", "p3", "p4"}, []string{"Any", "Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	or4Method.PrimitiveCode = builtinOr

	or5Method, err := RT.CreateMethod("","or", []string{"p1", "p2", "p3", "p4", "p5"}, []string{"Any", "Any", "Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	or5Method.PrimitiveCode = builtinOr

	or6Method, err := RT.CreateMethod("","or", []string{"p1", "p2", "p3", "p4", "p5", "p6"}, []string{"Any", "Any", "Any", "Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	or6Method.PrimitiveCode = builtinOr

	or7Method, err := RT.CreateMethod("","or", []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7"}, []string{"Any", "Any", "Any", "Any", "Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	or7Method.PrimitiveCode = builtinOr


















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
	


    ///////////////////////////////////////////////////////////////////
    // Collection functions	
	
	// len coll Collection > Int	
	//
	lenMethod, err := RT.CreateMethod("","len", []string{"c"}, []string{"Collection"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	lenMethod.PrimitiveCode = builtinLen


	stringLenMethod, err := RT.CreateMethod("","len", []string{"c"}, []string{"String"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringLenMethod.PrimitiveCode = builtinStringLen	
	
}


///////////////////////////////////////////////////////////////////////////////////////////
// I/O functions

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

//////////////////////////////////////////////////////////////////////////////////////////////
// Comparison functions

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
	case Bool:
		switch obj2.(type) {
		case Bool:
			val = Bool(obj1.(Bool) == obj2.(Bool))	
		default:
			val = Bool(false)			
		}			
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


//////////////////////////////////////////////////////////////////////////////////////////////
// Arithmetic functions

/*
Multiply operator. Note: Result type varies (is covariant with argument type variation). All results are Numeric but result is a different
subtype of numeric depending on the argument types.
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
Arithmetic Addition operator. Note: Result type varies (is covariant with argument type variation). All results are Numeric but result is a different
subtype of numeric depending on the argument types.
*/
func builtinPlus(objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	var val RObject
	switch obj1.(type) {
	case Int:
		switch obj2.(type) {
		case Int:
			val = Int(int64(obj1.(Int)) + int64(obj2.(Int)))
		case Int32:
			val = Int(int64(obj1.(Int)) + int64(obj2.(Int32)))
		case Float:
			val = Float(float64(obj2.(Float)) + float64(obj1.(Int)))
		default:
		    rterr.Stop("plus is not defined for argument types")
		}
	case Int32:
		switch obj2.(type) {
		case Int32:
			val = Int(int32(obj1.(Int32)) + int32(obj2.(Int32)))
		case Int:
			val = Int(int64(obj1.(Int32)) + int64(obj2.(Int)))
		case Float:
			val = Float(float64(obj2.(Float)) + float64(obj1.(Int32)))
		default:
		    rterr.Stop("plus is not defined for argument types")
		}
	case Float:
		switch obj2.(type) {
		case Float:
			val = Float(float64(obj1.(Float)) + float64(obj2.(Float)))
		case Int:
			val = Float(float64(obj1.(Float)) + float64(obj2.(Int)))
		case Int32:
			val = Float(float64(obj1.(Float)) + float64(obj2.(Int32)))
		default:
		    rterr.Stop("plus is not defined for argument types")
		}
	default:
		rterr.Stop("plus is not defined for argument types")
	}
	return []RObject{val}
}


/*
Arithmetic Subtraction operator. Note: Result type varies (is covariant with argument type variation). All results are Numeric but result is a different
subtype of numeric depending on the argument types.
*/
func builtinMinus(objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	var val RObject
	switch obj1.(type) {
	case Int:
		switch obj2.(type) {
		case Int:
			val = Int(int64(obj1.(Int)) - int64(obj2.(Int)))
		case Int32:
			val = Int(int64(obj1.(Int)) - int64(obj2.(Int32)))
		case Float:
			val = Float(float64(obj2.(Float)) - float64(obj1.(Int)))
		default:
		    rterr.Stop("minus is not defined for argument types")
		}
	case Int32:
		switch obj2.(type) {
		case Int32:
			val = Int(int32(obj1.(Int32)) - int32(obj2.(Int32)))
		case Int:
			val = Int(int64(obj1.(Int32)) - int64(obj2.(Int)))
		case Float:
			val = Float(float64(obj2.(Float)) - float64(obj1.(Int32)))
		default:
		    rterr.Stop("minus is not defined for argument types")
		}
	case Float:
		switch obj2.(type) {
		case Float:
			val = Float(float64(obj1.(Float)) - float64(obj2.(Float)))
		case Int:
			val = Float(float64(obj1.(Float)) - float64(obj2.(Int)))
		case Int32:
			val = Float(float64(obj1.(Float)) - float64(obj2.(Int32)))
		default:
		    rterr.Stop("minus is not defined for argument types")
		}
	default:
		rterr.Stop("minus is not defined for argument types")
	}
	return []RObject{val}
}

/*
Arithmetic division operator. Note: Result type varies (is covariant with argument type variation). All results are Numeric but result is a different
subtype of numeric depending on the argument types.
*/
func builtinDiv(objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	var val RObject
	switch obj1.(type) {
	case Int:
		switch obj2.(type) {
		case Int:
			val = Int(int64(obj1.(Int)) / int64(obj2.(Int)))
		case Int32:
			val = Int(int64(obj1.(Int)) / int64(obj2.(Int32)))
		case Float:
			val = Float(float64(obj2.(Float)) / float64(obj1.(Int)))
		default:
		    rterr.Stop("div is not defined for argument types")
		}
	case Int32:
		switch obj2.(type) {
		case Int32:
			val = Int(int32(obj1.(Int32)) / int32(obj2.(Int32)))
		case Int:
			val = Int(int64(obj1.(Int32)) / int64(obj2.(Int)))
		case Float:
			val = Float(float64(obj2.(Float)) / float64(obj1.(Int32)))
		default:
		    rterr.Stop("div is not defined for argument types")
		}
	case Float:
		switch obj2.(type) {
		case Float:
			val = Float(float64(obj1.(Float)) / float64(obj2.(Float)))
		case Int:
			val = Float(float64(obj1.(Float)) / float64(obj2.(Int)))
		case Int32:
			val = Float(float64(obj1.(Float)) / float64(obj2.(Int32)))
		default:
		    rterr.Stop("div is not defined for argument types")
		}
	default:
		rterr.Stop("div is not defined for argument types")
	}
	return []RObject{val}
}

/*
Arithmetic modulo operator. Note: Result type varies (is covariant with argument type variation). All results are Numeric but result is a different
subtype of numeric depending on the argument types.
*/
func builtinMod(objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	var val RObject
	switch obj1.(type) {
	case Int:
		switch obj2.(type) {
		case Int:
			val = Int(int64(obj1.(Int)) % int64(obj2.(Int)))
		case Int32:
			val = Int(int64(obj1.(Int)) % int64(obj2.(Int32)))
		default:
		    rterr.Stop("mod is not defined for argument types")
		}
	case Int32:
		switch obj2.(type) {
		case Int32:
			val = Int(int32(obj1.(Int32)) % int32(obj2.(Int32)))
		case Int:
			val = Int(int64(obj1.(Int32)) % int64(obj2.(Int)))
		default:
		    rterr.Stop("mod is not defined for argument types")
		}
	default:
		rterr.Stop("mod is not defined for argument types")
	}
	return []RObject{val}
}




/////////////////////////////////////////////////////////////////////
// Persistence functions

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

car1 = Car: summon "FEC 092"

TODO Add an optional radius argument controlling fetch eagerness of non-primitive attributes and objects related to the object fetched. 
Right now is using radius=0 meaning fetch the object with its primitive (and non-multi) valued attribute values. Other attributes are
currently fetched from DB lazily.

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
exists String > Bool

if exists "FEC 092"
   car1 = Car: summon "FEC 092"


*/
func builtinExists(objects []RObject) []RObject {
	relish.EnsureDatabase()
	name := objects[0].String()

	found, err := RT.DB().ObjectNameExists(name) 
	if err != nil {
		panic(err)
	}

	return []RObject{Bool(found)}
}

//////////////////////////////////////////////////////////////////////
// Boolean logic functions

/*
Boolean logic not. 
Return type is Bool
*/
func builtinNot(objects []RObject) []RObject {
	return []RObject{Bool(objects[0].IsZero())}
}

/*
Boolean logic and. Note THIS IS NOT LAZY evaluator of arguments, 
unlike LISP and. Darn. Maybe change that someday.  
Returns false if any argument is zero,
otherwise returns the last argument.
Return type is Any
*/
func builtinAnd(objects []RObject) []RObject {

    var obj RObject
	for _,obj = range objects {
       if obj.IsZero() {
       	 return []RObject{Bool(false)}
       }
	}
	return[]RObject{obj}
}

/*
Boolean logic or. Note THIS IS NOT LAZY evaluator of arguments, 
unlike LISP or. Darn. Maybe change that someday.  
Returns false if all arguments are zero,
otherwise returns the first argument which is non-zero.
Return type is Any
*/
func builtinOr(objects []RObject) []RObject {

    var obj RObject
	for _,obj = range objects {
       if ! obj.IsZero() {
       	 return[]RObject{obj}
       }
	}

	return []RObject{Bool(false)}
}


///////////////////////////////////////////////////////////////////////  
// Collection functions

func builtinLen(objects []RObject) []RObject {
	coll := objects[0].(RCollection)
	var val RObject
	val = Int(coll.Length())
	return []RObject{val}
}


func builtinStringLen(objects []RObject) []RObject {
	obj := objects[0].(String)
    s := string(obj)
	var val RObject
	val = Int(int64(len(s)))
	return []RObject{val}
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

/*
TODO IMPORTANT

Implement a variadic builtin function called fill which is used for python/go style string variable substitution.
e.g.

s = fill "Hello, my name is %s %s and I am %d years old." firstName lastName 19


   s = fill """
"""
Hello!

My name is %s %s 
and I am %d years old.
"""
            firstName
            lastName
            19 

Also, should provide a template builtin function

result = template templateString singleArgObject
*/