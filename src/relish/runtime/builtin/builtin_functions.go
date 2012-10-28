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
	"unicode/utf8"
	"relish/rterr"
    "time"
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

	ltTimeMethod, err := RT.CreateMethod("","lt", []string{"p1", "p2"}, []string{"Time","Time"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	ltTimeMethod.PrimitiveCode = builtinLtTime	

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

	gtTimeMethod, err := RT.CreateMethod("","gt", []string{"p1", "p2"}, []string{"Time", "Time"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	gtTimeMethod.PrimitiveCode = builtinGtTime	

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

	negMethod, err := RT.CreateMethod("","neg", []string{"p1"}, []string{"Numeric"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	negMethod.PrimitiveCode = builtinNeg		



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
	
    // substituting values for %s 
    //
    stringFill2Method, err := RT.CreateMethod("","fill", []string{"s1", "s2"}, []string{"String", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill2Method.PrimitiveCode = builtinStringFill

	stringFill3Method, err := RT.CreateMethod("","fill", []string{"s1", "s2", "s3"}, []string{"String", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill3Method.PrimitiveCode = builtinStringFill

	stringFill4Method, err := RT.CreateMethod("","fill", []string{"s1", "s2", "s3", "s4"}, []string{"String", "Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill4Method.PrimitiveCode = builtinStringFill

	stringFill5Method, err := RT.CreateMethod("","fill", []string{"s1", "s2", "s3", "s4", "s5"}, []string{"String", "Any", "Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill5Method.PrimitiveCode = builtinStringFill

	stringFill6Method, err := RT.CreateMethod("","fill", []string{"s1", "s2", "s3", "s4", "s5", "s6"}, []string{"String", "Any", "Any", "Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill6Method.PrimitiveCode = builtinStringFill

	stringFill7Method, err := RT.CreateMethod("","fill", []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7"}, []string{"String", "Any", "Any", "Any", "Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill7Method.PrimitiveCode = builtinStringFill	

	stringFill8Method, err := RT.CreateMethod("","fill", []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8"}, []string{"String", "Any", "Any", "Any", "Any", "Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill8Method.PrimitiveCode = builtinStringFill	

		stringFill9Method, err := RT.CreateMethod("","fill", []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8", "s9"}, []string{"String", "Any", "Any", "Any", "Any", "Any", "Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill9Method.PrimitiveCode = builtinStringFill		


    // length of a string
    //
    // in bytes
	stringLenMethod, err := RT.CreateMethod("","len", []string{"c"}, []string{"String"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringLenMethod.PrimitiveCode = builtinStringLen
	
    // in characters (codepoints)
    //
	stringNumCodePointsMethod, err := RT.CreateMethod("","numCodePoints", []string{"c"}, []string{"String"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringNumCodePointsMethod.PrimitiveCode = builtinStringNumCodePoints
	
	// concatenation of strings
	
    // s = cat s1 String s2 String s3 String s4 String s5 String s6 String s7 String s8 String s9 String > String  
	
    stringCat2Method, err := RT.CreateMethod("","cat", []string{"s1", "s2"}, []string{"Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat2Method.PrimitiveCode = builtinStringCat

	stringCat3Method, err := RT.CreateMethod("","cat", []string{"s1", "s2", "s3"}, []string{"Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat3Method.PrimitiveCode = builtinStringCat

	stringCat4Method, err := RT.CreateMethod("","cat", []string{"s1", "s2", "s3", "s4"}, []string{"Any", "Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat4Method.PrimitiveCode = builtinStringCat

	stringCat5Method, err := RT.CreateMethod("","cat", []string{"s1", "s2", "s3", "s4", "s5"}, []string{"Any", "Any", "Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat5Method.PrimitiveCode = builtinStringCat

	stringCat6Method, err := RT.CreateMethod("","cat", []string{"s1", "s2", "s3", "s4", "s5", "s6"}, []string{"Any", "Any", "Any", "Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat6Method.PrimitiveCode = builtinStringCat

	stringCat7Method, err := RT.CreateMethod("","cat", []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7"}, []string{"Any", "Any", "Any", "Any", "Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat7Method.PrimitiveCode = builtinStringCat	

	stringCat8Method, err := RT.CreateMethod("","cat", []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8"}, []string{"Any", "Any", "Any", "Any", "Any", "Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat8Method.PrimitiveCode = builtinStringCat	

		stringCat9Method, err := RT.CreateMethod("","cat", []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8", "s9"}, []string{"Any", "Any", "Any", "Any", "Any", "Any", "Any", "Any", "Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat9Method.PrimitiveCode = builtinStringCat	
	
	
    stringHasPrefixMethod, err := RT.CreateMethod("","hasPrefix", []string{"s1", "s2"}, []string{"String", "String"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringHasPrefixMethod.PrimitiveCode = builtinStringHasPrefix	
	
    stringHasSuffixMethod, err := RT.CreateMethod("","hasSuffix", []string{"s1", "s2"}, []string{"String", "String"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringHasSuffixMethod.PrimitiveCode = builtinStringHasSuffix	
	
	// index of substring in string
	
/*	
	index s String t String > Int
	"""
	 index of first occurrence of t in s or -1 if t not found in s 
	"""
*/
    stringIndexMethod, err := RT.CreateMethod("","index", []string{"s1", "s2"}, []string{"String", "String"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringIndexMethod.PrimitiveCode = builtinStringIndex
	
    stringLastIndexMethod, err := RT.CreateMethod("","lastIndex", []string{"s1", "s2"}, []string{"String", "String"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringLastIndexMethod.PrimitiveCode = builtinStringLastIndex	


    // substring or slice
/*
	slice s String start Int end Int > String
*/
    stringSlice2Method, err := RT.CreateMethod("","slice", []string{"s", "start"}, []string{"String", "Int"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringSlice2Method.PrimitiveCode = builtinStringSlice	
	
    stringSlice3Method, err := RT.CreateMethod("","slice", []string{"s", "start", "end"}, []string{"String", "Int", "Int"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringSlice3Method.PrimitiveCode = builtinStringSlice	
	
    stringFirstMethod, err := RT.CreateMethod("","first", []string{"s", "n"}, []string{"String", "Int"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringFirstMethod.PrimitiveCode = builtinStringFirst	
	
    stringLastMethod, err := RT.CreateMethod("","last", []string{"s", "n"}, []string{"String", "Int"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	stringLastMethod.PrimitiveCode = builtinStringLast	
	
	
	
	
    ///////////////////////////////////////////////////////////////////
    // Collection functions	
	
	// len coll Collection > Int	
	//
	lenMethod, err := RT.CreateMethod("","len", []string{"c"}, []string{"Collection"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	lenMethod.PrimitiveCode = builtinLen


	// cap coll Collection > Int	
	//
	capMethod, err := RT.CreateMethod("","cap", []string{"c"}, []string{"Collection"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	capMethod.PrimitiveCode = builtinCap	


    ///////////////////////////////////////////////////////////////////
    // Concurrency functions	

	// len c Channel > Int	
	//
	channelLenMethod, err := RT.CreateMethod("","len", []string{"c"}, []string{"Channel"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	channelLenMethod.PrimitiveCode = builtinChannelLen    

	// cap c Channel > Int	
	//
	channelCapMethod, err := RT.CreateMethod("","cap", []string{"c"}, []string{"Channel"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	channelCapMethod.PrimitiveCode = builtinChannelCap    	


	// from c InChannel of T > T	
	//
	channelFromMethod, err := RT.CreateMethod("","<-", []string{"c"}, []string{"InChannel"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	channelFromMethod.PrimitiveCode = builtinFrom   	

/* Deprecated in favour of ch <- operator

    // to c OutChannel of T obj T 	
	//
	channelToMethod, err := RT.CreateMethod("","to", []string{"c","v"}, []string{"OutChannel","Any"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	channelToMethod.PrimitiveCode = builtinTo 
*/


	mutexLockMethod, err := RT.CreateMethod("","lock", []string{"m"}, []string{"Mutex"}, 0, 0, false)
	if err != nil {
		panic(err)
	}
	mutexLockMethod.PrimitiveCode = builtinMutexLock

	mutexUnlockMethod, err := RT.CreateMethod("","unlock", []string{"m"}, []string{"Mutex"}, 0, 0, false)
	if err != nil {
		panic(err)
	}
	mutexUnlockMethod.PrimitiveCode = builtinMutexUnlock

	rwmutexLockMethod, err := RT.CreateMethod("","lock", []string{"m"}, []string{"RWMutex"}, 0, 0, false)
	if err != nil {
		panic(err)
	}
	rwmutexLockMethod.PrimitiveCode = builtinRWMutexLock

	rwmutexUnlockMethod, err := RT.CreateMethod("","unlock", []string{"m"}, []string{"RWMutex"}, 0, 0, false)
	if err != nil {
		panic(err)
	}
	rwmutexUnlockMethod.PrimitiveCode = builtinRWMutexUnlock

	rwmutexRLockMethod, err := RT.CreateMethod("","rlock", []string{"m"}, []string{"RWMutex"}, 0, 0, false)
	if err != nil {
		panic(err)
	}
	rwmutexRLockMethod.PrimitiveCode = builtinRWMutexRLock

	rwmutexRUnlockMethod, err := RT.CreateMethod("","runlock", []string{"m"}, []string{"RWMutex"}, 0, 0, false)
	if err != nil {
		panic(err)
	}
	rwmutexRUnlockMethod.PrimitiveCode = builtinRWMutexRUnlock




    /////////////////////////////////////////////////////////////////////
    // Type init functions

	timeInit1Method, err := RT.CreateMethod("","initTime", []string{"t","timeString"}, []string{"Time","String"}, 2, 0, false)
	if err != nil {
		panic(err)
	}
	timeInit1Method.PrimitiveCode = builtinInitTimeParse    
	
	timeInit2Method, err := RT.CreateMethod("","initTime", []string{"t","timeString","layout"}, []string{"Time","String","String"}, 2, 0, false)
	if err != nil {
		panic(err)
	}
	timeInit2Method.PrimitiveCode = builtinInitTimeParse	

	timeInit3Method, err := RT.CreateMethod("","initTime", []string{"t","year","month","day","hour","min","sec","nsec","loc"}, []string{"Time","Int","Int","Int","Int","Int","Int","Int","String"}, 2, 0, false)
	if err != nil {
		panic(err)
	}
	timeInit3Method.PrimitiveCode = builtinInitTimeDate	



	channelInit0Method, err := RT.CreateMethod("","initChannel", []string{"c"}, []string{"Channel"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	channelInit0Method.PrimitiveCode = builtinInitChannel 

	channelInit1Method, err := RT.CreateMethod("","initChannel", []string{"c","n"}, []string{"Channel","Int"}, 1, 0, false)
	if err != nil {
		panic(err)
	}
	channelInit1Method.PrimitiveCode = builtinInitChannel	

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
		case Float:
			val = Bool(float64(obj2.(Float)) == float64(obj1.(Int)))			
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
		case Float:
			val = Bool(float64(obj2.(Float)) == float64(obj1.(Int32)))			
		case String:
			val = Bool(string(obj2.(String)) == strconv.Itoa(int(obj1.(Int32))))
		default:
			val = Bool(false)
		}
	case Float:
		switch obj2.(type) {
		case Int32:
			val = Bool(float64(obj1.(Float)) == float64(obj2.(Int32)))		
		case Int:
			val = Bool(float64(obj1.(Float)) == float64(obj2.(Int)))	
		case Float:
			val = Bool(float64(obj2.(Float)) == float64(obj1.(Float)))			
		case String:
			val = Bool(string(obj2.(String)) == strconv.FormatFloat(float64(obj1.(Float)), 'G', -1, 64)) 
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
		case Float:
			val = Bool(string(obj1.(String)) == strconv.FormatFloat(float64(obj2.(Float)), 'G', -1, 64)) 		
		default:
			val = Bool(false)
		}
	case RTime:
		switch obj2.(type) {
		case RTime:
			val = Bool(time.Time(obj1.(RTime)).Equal(time.Time(obj2.(RTime))))				
//		case String:
// Add a Parse of an ISO8601 time string, and allow String, RTime combination as well
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
Less-than operator for time types.
*/
func builtinLtTime(objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	var val RObject
	switch obj1.(type) {
	case RTime:
		switch obj2.(type) {
		case RTime:
			val = Bool(time.Time(obj1.(RTime)).Before(time.Time(obj2.(RTime))))
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
Greater-than operator for time types.
*/
func builtinGtTime(objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	var val RObject
	switch obj1.(type) {
	case RTime:
		switch obj2.(type) {
		case RTime:
			val = Bool(time.Time(obj1.(RTime)).After(time.Time(obj2.(RTime))))
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



/*
Numeric negation operator. Note: Result type varies (is covariant with argument type variation). All results are Numeric but result is a different
subtype of numeric depending on the argument types.
*/
func builtinNeg(objects []RObject) []RObject {
	obj1 := objects[0]
	var val RObject
	switch obj1.(type) {
	case Int:
			val = Int(-int64(obj1.(Int)))
	case Int32:
			val = Int32(-int32(obj1.(Int32)))
	case Float:
			val = Float(-float64(obj1.(Float)))
	default:
		    rterr.Stop("neg is not defined for argument type")
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

func builtinCap(objects []RObject) []RObject {
	coll := objects[0].(RCollection)
	var val RObject
	val = Int(coll.Cap())
	return []RObject{val}
}

///////////////////////////////////////////////////////////////////////
// Concurrency functions

/*

from ch InChannel of T > T

val = from ch

TODO DUMMY Implementation 
*/
func builtinFrom(objects []RObject) []RObject {
	c := objects[0].(*Channel)
    val  := <- c.Ch
	return []RObject{val}
}


/*

to 
   ch OutChannel of T 
   val T

to ch val


func builtinTo(objects []RObject) []RObject {
	c := objects[0].(*Channel)
    val := objects[1]
    // TODO do a runtime type-compatibility check of val's type with c.ElementType
	c.Ch <- val

	return []RObject{}
}
*/

func builtinChannelLen(objects []RObject) []RObject {
	c := objects[0].(*Channel)
	var val RObject
	val = Int(c.Length())
	return []RObject{val}
}

func builtinChannelCap(objects []RObject) []RObject {
	c := objects[0].(*Channel)
	var val RObject
	val = Int(c.Cap())
	return []RObject{val}
}

func builtinMutexLock(objects []RObject) []RObject {
	c := objects[0].(*Mutex)
	c.Lock()
	return []RObject{}
}

func builtinMutexUnlock(objects []RObject) []RObject {
	c := objects[0].(*Mutex)
	c.Unlock()
	return []RObject{}
}

func builtinRWMutexLock(objects []RObject) []RObject {
	c := objects[0].(*RWMutex)
	c.Lock()
	return []RObject{}
}

func builtinRWMutexUnlock(objects []RObject) []RObject {
	c := objects[0].(*RWMutex)
	c.Unlock()
	return []RObject{}
}

func builtinRWMutexRLock(objects []RObject) []RObject {
	c := objects[0].(*RWMutex)
	c.RLock()
	return []RObject{}
}

func builtinRWMutexRUnlock(objects []RObject) []RObject {
	c := objects[0].(*RWMutex)
	c.RUnlock()
	return []RObject{}
}

/////////////////////////////////////////////////////////////// 
// String functions

// length in bytes

func builtinStringLen(objects []RObject) []RObject {
	obj := objects[0].(String)
    s := string(obj)
	var val RObject
	val = Int(int64(len(s)))
	return []RObject{val}
}

// count of unicode codepoints

func builtinStringNumCodePoints(objects []RObject) []RObject {
	obj := objects[0].(String)
    s := string(obj)
	var val RObject
	val = Int(int64(utf8.RuneCountInString(s)))
	return []RObject{val}
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

func builtinStringHasPrefix(objects []RObject) []RObject {
	s := objects[0].(String)
	substr := objects[1].(String)
	var val RObject
	val = Bool( strings.HasPrefix(string(s), string(substr)) )
	return []RObject{val}
}

func builtinStringHasSuffix(objects []RObject) []RObject {
	s := objects[0].(String)
	substr := objects[1].(String)
	var val RObject
	val = Bool( strings.HasSuffix(string(s), string(substr)) )
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


func builtinStringFill(objects []RObject) []RObject {
	s := string(objects[0].(String))

	nFillers := len(objects) - 1
    nSlots := strings.Count(s, "%s")
	if nSlots != nFillers {
		snippet := s
		if len(snippet) > 70 {
			chars := strings.Split(s,"")
			if len(chars) > 25 {
			   firstChars := chars[:25]
			   snippet = strings.Join(firstChars,"") + "..."
		    }
		}
        rterr.Stopf("fill: String '%s' contains %d %%s slots but there are %d objects to fill slots.",snippet, nSlots, nFillers)
	} 

	for i := 1; i <= nSlots; i++ {
       filler := objects[i].String()
       s = strings.Replace(s,"%s", filler, 1)
	}
    return []RObject{String(s)}
}



		// concatenation of strings

	    // s = cat s1 String s2 String s3 String s4 String s5 String s6 String s7 String s8 String s9 String > String  

func builtinStringCat(objects []RObject) []RObject {
	s := objects[0].String()
    n := len(objects) 
	for i := 1; i < n; i++ {
		s += objects[i].String()
	}
    return []RObject{String(s)}
}

		// index of substring in string

	/*	
		index s String t String > Int
		"""
		 index of first occurrence of t in s or -1 if t not found in s 
		"""
	*/
	
func builtinStringIndex(objects []RObject) []RObject {
	s := string(objects[0].(String))
	substr := string(objects[1].(String))
	var val RObject
    val = Int(int64(strings.Index(s, substr)))	
	return []RObject{val}
}


func builtinStringLastIndex(objects []RObject) []RObject {
	s := string(objects[0].(String))
	substr := string(objects[1].(String))
	var val RObject
    val = Int(int64(strings.LastIndex(s, substr)))	
	return []RObject{val}
}


	    // substring or slice
	/*
		slice s String start Int end Int > String
		
		slice s String start Int end Int > String		
	*/

/*
Counts in bytes.
*/
func builtinStringSlice(objects []RObject) []RObject {
	s := string(objects[0].(String))
	start := int(objects[1].(Int))
	length := len(s)
	end := length
	if len(objects) == 3 {
		end = int(objects[2].(Int))
		if end < 0 {
			end = length + end
		}
	}
	substr := s[start:end]
    return []RObject{String(substr)}
}

/*
Counts in utf-8 encoded codepoints.
Does not mind if the actual string is shorter than n codepoints.
*/
func builtinStringFirst(objects []RObject) []RObject {
	s := string(objects[0].(String))
	n := int(objects[1].(Int))
    i := 0
    end := len(s)
    var pos int
	for pos, _ = range s {
	    if i >= n {
		   end = pos
		   break
	    }
	    i++
	}
	substr := s[0:end]
    return []RObject{String(substr)}	
}


/*
Counts in utf-8 encoded codepoints.
Does not mind if the actual string is shorter than n codepoints.
*/
func builtinStringLast(objects []RObject) []RObject {
	s := string(objects[0].(String))
	n := int(objects[1].(Int))
    if n == 0 {
      return []RObject{String("")}	
    }

    nRunes := utf8.RuneCountInString(s)
    n = nRunes - n
    if n < 0 {
	   n = 0
    }
	
    i := 0
    end := len(s)
    var pos int
    var start int
	for pos, _ = range s {
	    if i >= n {
		   start = pos
		   break
	    }
	    i++
	}
	substr := s[start:end]
    return []RObject{String(substr)}	
}




/* ISO 8601 format as in W3C

t err = Time "2012-09-23T14:23Z"

t err = Time "2012-09-23T14:23:00Z"

t err = Time "2012-09-23T14:23:00.000Z"

t err = Time "2012-09-23 14:23 UTC"

t err = Time "2012-09-23 14:23:00 America/New York"

t err = Time "2012-09-23 14:23:00.000 Local" 

t err = Time "2012-09-23 14:23:00.000 " "2006-01-02 15:04:05.999 "




initTime t0 Time location String > t Time err String

initTime t0 Time location String > t Time err String

initTime t0 Time inKey String location String > t Time err String

*/

func builtinInitTimeParse(objects []RObject) []RObject {
// ignore first Time arg
	if len(objects) == 2 {
		return initTimeParse1(objects[1].String())
	} 
	return initTimeParse2(objects[1].String(),objects[2].String())
}

func initTimeParse1(timeString string) []RObject {

   layout := "2006-01-02T15:04Z"		
	
   t, err := time.Parse(layout, timeString)
   if err != nil {
      layout = "2006-01-02T15:04:05Z"
      t, err = time.Parse(layout, timeString)	
	  if err != nil {
	     layout = "2006-01-02T15:04:05.999Z"
	     t, err = time.Parse(layout, timeString)
	     if err != nil {
             pieces := strings.SplitN(timeString, " " , 3) 
             if len(pieces) < 3 {
			    var tZero time.Time 
			    t := RTime(tZero)
		        return []RObject{t, String("Invalid date-time string format - see relish language reference")}	
             }
             timeString = pieces[0] + " " + pieces[1]
		     locationName := pieces[2]
		     loc, err := time.LoadLocation(locationName) 
		     if err != nil {
			    var tZero time.Time 
			    t := RTime(tZero)
		        return []RObject{t, String(err.Error())}	
		     }	

	         layout = "2006-01-02 15:04"	
             tUtc, err := time.Parse(layout, timeString)		
		     if err != nil {
	            layout = "2006-01-02 15:04:05"				
                tUtc, err = time.Parse(layout, timeString)		
		        if err != nil {			
	               layout = "2006-01-02 15:04:05.999"
                   tUtc, err = time.Parse(layout, timeString)					
		           if err != nil {				  
			          var tZero time.Time 
			          t := RTime(tZero)
		              return []RObject{t, String(err.Error())}	
		           }
		        }
		     }
		     y := tUtc.Year()
		     m := tUtc.Month()
		     d := tUtc.Day()
		     hh := tUtc.Hour()
		     mm := tUtc.Minute()
		     ss := tUtc.Second()
		     ns := tUtc.Nanosecond() 

		     t = time.Date(y, m, d, hh, mm, ss, ns, loc) 
	     }		  
	  }  
   }
   return []RObject{RTime(t),String("")}
}


func initTimeParse2(timeString string, layout string) []RObject {
   t, err := time.Parse(layout, timeString)
   if err != nil {
	  var tZero time.Time 
	  t := RTime(tZero)
      return []RObject{t, String(err.Error())}	
   }		
   return []RObject{RTime(t),String("")}
}



/*
t err = Time 2012 12 30 18 36 29 0 "UTC"

initTime t0 Time year Int month Int day Int hour Int min Int sec Int nsec Int location String > t Time err String
*/
func builtinInitTimeDate(objects []RObject) []RObject {

   // ignore first Time argument
  
   year := int(objects[1].(Int))
   monthInt := int(objects[2].(Int))
   month := time.Month(monthInt)
   day := int(objects[3].(Int))
   hour := int(objects[4].(Int))
   min := int(objects[5].(Int))
   sec := int(objects[6].(Int))
   nsec := int(objects[7].(Int))
   locationName := objects[8].String()
   loc, err := time.LoadLocation(locationName) 
   if err != nil {
	  var tZero time.Time 
	  t := RTime(tZero)
      return []RObject{t, String(err.Error())}	
   }

   t := time.Date(year, month, day, hour, min, sec, nsec, loc) 

   return []RObject{RTime(t),String("")}
}



func builtinInitChannel(objects []RObject) []RObject {
   
    c := objects[0].(*Channel)

	var n int
	if len(objects) == 2 {
		n = int(objects[1].(Int))
		if n < 0 {
			rterr.Stop("Channel capacity cannot be specified to be less than zero.")
		}
	} 
	c.Ch = make(chan RObject, n)
	return []RObject{c}
}
