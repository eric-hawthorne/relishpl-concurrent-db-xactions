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
    "crypto/sha256"
    "encoding/base64"
	"os"
	"bufio"
	"net/smtp"
	"net/http"
	"net/url"	
	"io/ioutil"
	"encoding/csv"
)

// Reader for reading from standard input
var buf *bufio.Reader = bufio.NewReader(os.Stdin)


func InitBuiltinFunctions() {

	dbgMethod, err := RT.CreateMethod("",nil,"dbg", []string{"p"}, []string{"Any"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	dbgMethod.PrimitiveCode = builtinDbg

	debugMethod, err := RT.CreateMethod("",nil,"debug", []string{"p"}, []string{"Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	debugMethod.PrimitiveCode = builtinDebug

	printMethod, err := RT.CreateMethod("",nil,"print", []string{"p"}, []string{"RelishPrimitive"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	printMethod.PrimitiveCode = builtinPrint

	print2Method, err := RT.CreateMethod("",nil,"print", []string{"p1", "p2"}, []string{"RelishPrimitive", "RelishPrimitive"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	print2Method.PrimitiveCode = builtinPrint

	print3Method, err := RT.CreateMethod("",nil,"print", []string{"p1", "p2", "p3"}, []string{"RelishPrimitive", "RelishPrimitive", "RelishPrimitive"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	print3Method.PrimitiveCode = builtinPrint

	print4Method, err := RT.CreateMethod("",nil,"print", []string{"p1", "p2", "p3", "p4"}, []string{"RelishPrimitive", "RelishPrimitive", "RelishPrimitive", "RelishPrimitive"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	print4Method.PrimitiveCode = builtinPrint

	print5Method, err := RT.CreateMethod("",nil,"print", []string{"p1", "p2", "p3", "p4", "p5"}, []string{"RelishPrimitive", "RelishPrimitive", "RelishPrimitive", "RelishPrimitive", "RelishPrimitive"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	print5Method.PrimitiveCode = builtinPrint

	print6Method, err := RT.CreateMethod("",nil,"print", []string{"p1", "p2", "p3", "p4", "p5", "p6"}, []string{"RelishPrimitive", "RelishPrimitive", "RelishPrimitive", "RelishPrimitive", "RelishPrimitive", "RelishPrimitive"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	print6Method.PrimitiveCode = builtinPrint

	print7Method, err := RT.CreateMethod("",nil,"print", []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7"}, []string{"RelishPrimitive", "RelishPrimitive", "RelishPrimitive", "RelishPrimitive", "RelishPrimitive", "RelishPrimitive", "RelishPrimitive"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	print7Method.PrimitiveCode = builtinPrint



	inputMethod, err := RT.CreateMethod("",nil,"input", []string{"message"}, []string{"String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	inputMethod.PrimitiveCode = builtinInput



	email1Method, err := RT.CreateMethod("",nil,
	                                     "sendEmail", 
	                                     []string{"smtpServerAddr","from","recipient","subject","messageBody"}, 
	                                     []string{"String","String","String","String","String"}, 
	                                     []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	email1Method.PrimitiveCode = builtinSendEmail



	email2Method, err := RT.CreateMethod("",nil,
	                                     "sendEmail", 
	                                     []string{"smtpServerAddr","from","recipients","subject","messageBody"}, 
	                                     []string{"String","String","List_of_String","String","String"}, 
	                                     []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	email2Method.PrimitiveCode = builtinSendEmail


	email3Method, err := RT.CreateMethod("",nil,
	                                     "sendEmail", 
	                                     []string{"smtpServerAddr","user","password","from","recipient","subject","messageBody"}, 
	                                     []string{"String","String","String","String","String","String","String"}, 
	                                     []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	email3Method.PrimitiveCode = builtinSendEmail


	email4Method, err := RT.CreateMethod("",nil,
	                                     "sendEmail", 
	                                     []string{"smtpServerAddr","user","password","from","recipients","subject","messageBody"}, 
	                                     []string{"String","String","String","String","List_of_String","String","String"}, 
	                                     []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	email4Method.PrimitiveCode = builtinSendEmail




    ////////////////////////////////////////////////////////////
    // http client functions

    // httpGet url String > responseBody String err String
    //
    //
	httpGetMethod, err := RT.CreateMethod("relish/pkg/http",nil,"httpGet", []string{"url"}, []string{"String"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	httpGetMethod.PrimitiveCode = builtinHttpGet

    // httpPost 
    //    url String 
    //    keysVals {} String > [] String
    // > 
    //    responseBody String 
    //    err String
    //
	httpPostMethod, err := RT.CreateMethod("relish/pkg/http",nil,"httpPost", []string{"url","keysVals"}, []string{"String","Map"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	httpPostMethod.PrimitiveCode = builtinHttpPost





	eqMethod, err := RT.CreateMethod("",nil,"eq", []string{"p1", "p2"}, []string{"RelishPrimitive", "RelishPrimitive"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	eqMethod.PrimitiveCode = builtinEq
	
	eq1Method, err := RT.CreateMethod("",nil,"eq", []string{"p1", "p2"}, []string{"Any", "RelishPrimitive"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	eq1Method.PrimitiveCode = builtinEqNo
	
	eq2Method, err := RT.CreateMethod("",nil,"eq", []string{"p1", "p2"}, []string{"RelishPrimitive", "Any"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	eq2Method.PrimitiveCode = builtinEqNo	
	
	eqObjMethod, err := RT.CreateMethod("",nil,"eq", []string{"p1", "p2"}, []string{"Any", "Any"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	eqObjMethod.PrimitiveCode = builtinEqObj
			

	ltNumMethod, err := RT.CreateMethod("",nil,"lt", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	ltNumMethod.PrimitiveCode = builtinLtNum

	ltTimeMethod, err := RT.CreateMethod("",nil,"lt", []string{"p1", "p2"}, []string{"Time","Time"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	ltTimeMethod.PrimitiveCode = builtinLtTime	

	ltStrMethod, err := RT.CreateMethod("",nil,"lt", []string{"p1", "p2"}, []string{"String", "String"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	ltStrMethod.PrimitiveCode = builtinLtStr
	
	
	
	lessNumMethod, err := RT.CreateMethod("",nil,"less", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	lessNumMethod.PrimitiveCode = builtinLtNum

	lessTimeMethod, err := RT.CreateMethod("",nil,"less", []string{"p1", "p2"}, []string{"Time","Time"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	lessTimeMethod.PrimitiveCode = builtinLtTime	

	lessStrMethod, err := RT.CreateMethod("",nil,"less", []string{"p1", "p2"}, []string{"String", "String"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	lessStrMethod.PrimitiveCode = builtinLtStr	
	
	beforeTimeMethod, err := RT.CreateMethod("",nil,"before", []string{"p1", "p2"}, []string{"Time","Time"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	beforeTimeMethod.PrimitiveCode = builtinLtTime	
	
	
	

	lteNumMethod, err := RT.CreateMethod("",nil,"lte", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	lteNumMethod.PrimitiveCode = builtinLteNum

	lteTimeMethod, err := RT.CreateMethod("",nil,"lte", []string{"p1", "p2"}, []string{"Time","Time"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	lteTimeMethod.PrimitiveCode = builtinLteTime	

	lteStrMethod, err := RT.CreateMethod("",nil,"lte", []string{"p1", "p2"}, []string{"String", "String"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	lteStrMethod.PrimitiveCode = builtinLteStr
	
	
	
	lessEqNumMethod, err := RT.CreateMethod("",nil,"lessEq", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	lessEqNumMethod.PrimitiveCode = builtinLteNum

	lessEqTimeMethod, err := RT.CreateMethod("",nil,"lessEq", []string{"p1", "p2"}, []string{"Time","Time"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	lessEqTimeMethod.PrimitiveCode = builtinLteTime	

	lessEqStrMethod, err := RT.CreateMethod("",nil,"lessEq", []string{"p1", "p2"}, []string{"String", "String"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	lessEqStrMethod.PrimitiveCode = builtinLteStr	
	
	
	
	

	
	gtNumMethod, err := RT.CreateMethod("",nil,"gt", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	gtNumMethod.PrimitiveCode = builtinGtNum

	gtTimeMethod, err := RT.CreateMethod("",nil,"gt", []string{"p1", "p2"}, []string{"Time", "Time"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	gtTimeMethod.PrimitiveCode = builtinGtTime	

	gtStrMethod, err := RT.CreateMethod("",nil,"gt", []string{"p1", "p2"}, []string{"String", "String"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	gtStrMethod.PrimitiveCode = builtinGtStr	
	
	
	
	
	greaterNumMethod, err := RT.CreateMethod("",nil,"greater", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	greaterNumMethod.PrimitiveCode = builtinGtNum

	greaterTimeMethod, err := RT.CreateMethod("",nil,"greater", []string{"p1", "p2"}, []string{"Time", "Time"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	greaterTimeMethod.PrimitiveCode = builtinGtTime	

	greaterStrMethod, err := RT.CreateMethod("",nil,"greater", []string{"p1", "p2"}, []string{"String", "String"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	greaterStrMethod.PrimitiveCode = builtinGtStr
		
	afterTimeMethod, err := RT.CreateMethod("",nil,"after", []string{"p1", "p2"}, []string{"Time", "Time"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	afterTimeMethod.PrimitiveCode = builtinGtTime		




	gteNumMethod, err := RT.CreateMethod("",nil,"gte", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	gteNumMethod.PrimitiveCode = builtinGteNum

	gteTimeMethod, err := RT.CreateMethod("",nil,"gte", []string{"p1", "p2"}, []string{"Time", "Time"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	gteTimeMethod.PrimitiveCode = builtinGteTime	

	gteStrMethod, err := RT.CreateMethod("",nil,"gte", []string{"p1", "p2"}, []string{"String", "String"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	gteStrMethod.PrimitiveCode = builtinGteStr	
	
	
	
	
	
	greaterEqNumMethod, err := RT.CreateMethod("",nil,"greaterEq", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	greaterEqNumMethod.PrimitiveCode = builtinGteNum

	greaterEqTimeMethod, err := RT.CreateMethod("",nil,"greaterEq", []string{"p1", "p2"}, []string{"Time", "Time"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	greaterEqTimeMethod.PrimitiveCode = builtinGteTime	

	greaterEqStrMethod, err := RT.CreateMethod("",nil,"greaterEq", []string{"p1", "p2"}, []string{"String", "String"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	greaterEqStrMethod.PrimitiveCode = builtinGteStr
		












    /////////////////////////////////////////////////////////
    // Numeric arithmetic functions
    //
    // Note: We have an issue here about return type
    //
    // Is the language going to allow co-variant return types in different methods of same multimethod?
    //
    // What does that imply for static type checking of return values?
    //

	timesMethod, err := RT.CreateMethod("",nil,"times", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, []string{"Numeric"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timesMethod.PrimitiveCode = builtinTimes


	plusMethod, err := RT.CreateMethod("",nil,"plus", []string{"p1", "p2"}, []string{"Numeric", "Numeric"},  []string{"Numeric"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	plusMethod.PrimitiveCode = builtinPlus	

	sumNumericMethod, err := RT.CreateMethod("",nil,"sum", []string{"nums"}, []string{"List_of_Numeric"},  []string{"Numeric"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	sumNumericMethod.PrimitiveCode = builtinSum	

	sumFloatMethod, err := RT.CreateMethod("",nil,"sum", []string{"nums"}, []string{"List_of_Float"},  []string{"Float"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	sumFloatMethod.PrimitiveCode = builtinSum

	sumIntMethod, err := RT.CreateMethod("",nil,"sum", []string{"nums"}, []string{"List_of_Int"},  []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	sumIntMethod.PrimitiveCode = builtinSum

	sumUintMethod, err := RT.CreateMethod("",nil,"sum", []string{"nums"}, []string{"List_of_Uint"},  []string{"Uint"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	sumUintMethod.PrimitiveCode = builtinSum	

	sumNumericSetMethod, err := RT.CreateMethod("",nil,"sum", []string{"nums"}, []string{"Set_of_Numeric"},  []string{"Numeric"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	sumNumericSetMethod.PrimitiveCode = builtinSum	

	sumFloatSetMethod, err := RT.CreateMethod("",nil,"sum", []string{"nums"}, []string{"Set_of_Float"},  []string{"Float"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	sumFloatSetMethod.PrimitiveCode = builtinSum

	sumIntSetMethod, err := RT.CreateMethod("",nil,"sum", []string{"nums"}, []string{"Set_of_Int"},  []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	sumIntSetMethod.PrimitiveCode = builtinSum

	sumUintSetMethod, err := RT.CreateMethod("",nil,"sum", []string{"nums"}, []string{"Set_of_Uint"},  []string{"Uint"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	sumUintSetMethod.PrimitiveCode = builtinSum	



	minusMethod, err := RT.CreateMethod("",nil,"minus", []string{"p1", "p2"}, []string{"Numeric", "Numeric"},  []string{"Numeric"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	minusMethod.PrimitiveCode = builtinMinus	

	divMethod, err := RT.CreateMethod("",nil,"div", []string{"p1", "p2"}, []string{"Numeric", "Numeric"},  []string{"Numeric"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	divMethod.PrimitiveCode = builtinDiv	

	modMethod, err := RT.CreateMethod("",nil,"mod", []string{"p1", "p2"}, []string{"Integer", "Integer"},  []string{"Integer"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	modMethod.PrimitiveCode = builtinMod	

	negMethod, err := RT.CreateMethod("",nil,"neg", []string{"p1"}, []string{"Numeric"},  []string{"Numeric"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	negMethod.PrimitiveCode = builtinNeg		


    ////////////////////////////////////////////////////////////
    // Time and Duration functions

    // now location String > Time
    //
    // t = now "America/Los_Angeles"       t = now "Local"      t = now "UTC"
    //
	timeNowMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"now", []string{"loc"}, []string{"String"}, []string{"Time"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeNowMethod.PrimitiveCode = builtinTimeNow	

    // sleep durationNs Int
    //
	sleepMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"sleep", []string{"ns"}, []string{"Int"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	sleepMethod.PrimitiveCode = builtinSleep

    // tick durationNs Int > InChannel of Time
    //
	tickMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"tick", []string{"ns"}, []string{"Int"},  []string{"InChannel"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	tickMethod.PrimitiveCode = builtinTick


    // plus t Time durationNs Int > Time
    //
	timePlusMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"plus", []string{"t", "ns"}, []string{"Time", "Int"}, []string{"Time"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timePlusMethod.PrimitiveCode = builtinTimePlus	

    // addDate t Time years Int months Int days Int > Time
    //
	timeAddDateMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"addDate", []string{"t", "years","months","days"}, []string{"Time", "Int", "Int", "Int"}, []string{"Time"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeAddDateMethod.PrimitiveCode = builtinTimeAddDate

    // minus t Time durationNs Int > Time
    //
	timeMinusDurationMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"minus", []string{"t", "ns"}, []string{"Time", "Int"}, []string{"Time"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeMinusDurationMethod.PrimitiveCode = builtinTimeMinusDuration	

    // minus t2 Time t2 Time > durationNs Int
    //
	timeMinusTimeMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"minus", []string{"t", "t"}, []string{"Time", "Time"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeMinusTimeMethod.PrimitiveCode = builtinTimeMinusTime	

    // since t Time > durationNs Int
    //
	timeSinceMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"since", []string{"t"}, []string{"Time"},  []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeSinceMethod.PrimitiveCode = builtinTimeSince	

    // timeIn t Time location String > Time
    //
	timeInMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"timeIn", []string{"t","location"}, []string{"Time","String"}, []string{"Time"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeInMethod.PrimitiveCode = builtinTimeIn

    // hours n Int > durationNs Int
    // 
	hoursMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"hours", []string{"n"}, []string{"Int"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	hoursMethod.PrimitiveCode = builtinHours

    // minutes n Int > durationNs Int
    // 
	minutesMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"minutes", []string{"n"}, []string{"Int"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	minutesMethod.PrimitiveCode = builtinMinutes	

    // seconds n Int > durationNs Int
    // 
	secondsMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"seconds", []string{"n"}, []string{"Int"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	secondsMethod.PrimitiveCode = builtinSeconds

    // milliseconds n Int > durationNs Int
    // 
	millisecondsMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"milliseconds", []string{"n"}, []string{"Int"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	millisecondsMethod.PrimitiveCode = builtinMilliseconds	






// date t Time > year Int month Int day Int
//
	timeDateMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"date", []string{"t"}, []string{"Time"}, []string{"Int","Int","Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeDateMethod.PrimitiveCode = builtinTimeDate

// clock t Time > hour Int min Int sec Int
//
	timeClockMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"clock", []string{"t"}, []string{"Time"}, []string{"Int","Int","Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeClockMethod.PrimitiveCode = builtinTimeClock

// day t Time > dayOfMonth Int  
// """
//  0..31
// """
//
	timeDayMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"day", []string{"t"}, []string{"Time"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeDayMethod.PrimitiveCode = builtinTimeDay


// hour t Time >  Int  
// """
//  0..23
// """
//
	timeHourMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"hour", []string{"t"}, []string{"Time"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeHourMethod.PrimitiveCode = builtinTimeHour


// minute t Time >  Int  
// """
//  0..59
// """
//
	timeMinuteMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"minute", []string{"t"}, []string{"Time"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeMinuteMethod.PrimitiveCode = builtinTimeMinute


// second t Time >  Int  
// """
//  0..59
// """
//
	timeSecondMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"second", []string{"t"}, []string{"Time"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeSecondMethod.PrimitiveCode = builtinTimeSecond

// nanosecond t Time >  Int  
// """
//  0..999999999
// """
//
	timeNanosecondMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"nanosecond", []string{"t"}, []string{"Time"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeNanosecondMethod.PrimitiveCode = builtinTimeNanosecond


// weekday t Time >  Int  
// """
//  0..6  Sunday = 0 
// """
//
	timeWeekdayMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"weekday", []string{"t"}, []string{"Time"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeWeekdayMethod.PrimitiveCode = builtinTimeWeekday


// year t Time >  Int  
//
	timeYearMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"year", []string{"t"}, []string{"Time"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeYearMethod.PrimitiveCode = builtinTimeYear


// month t Time >  Int  
// """
//  1..12  January = 1
// """
//
	timeMonthMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"month", []string{"t"}, []string{"Time"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeMonthMethod.PrimitiveCode = builtinTimeMonth
	

// zone t Time > name String offset Int
// """
//  EST secondsEastOfUTC
// """  
//
	timeZoneMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"zone", []string{"t"}, []string{"Time"}, []string{"String","Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeZoneMethod.PrimitiveCode = builtinTimeZone


// format t Time layout String > String
// """
//  Like Go time layouts (http://golang.org/pkg/time/)
// """  
//
	timeFormatMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"format", []string{"t","layout"}, []string{"Time","String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeFormatMethod.PrimitiveCode = builtinTimeFormat








    // duration hours Int minutes Int > durationNs Int
    // 
	durationHoursMinutesMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"duration", []string{"h","m"}, []string{"Int","Int"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	durationHoursMinutesMethod.PrimitiveCode = builtinDuration

    // duration hours Int minutes Int seconds Int > durationNs Int
    // 
	durationHoursMinutesSecondsMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"duration", []string{"h","m","s"}, []string{"Int","Int","Int"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	durationHoursMinutesSecondsMethod.PrimitiveCode = builtinDuration	

    // duration hours Int minutes Int seconds Int nanoseconds Int > durationNs Int
    // 
	durationHoursMinutesSecondsNsMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"duration", []string{"h","m","s","ns"}, []string{"Int","Int","Int","Int"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	durationHoursMinutesSecondsNsMethod.PrimitiveCode = builtinDuration		


    // hoursEquivalentOf durationNs Int > Float
    //
    // Returns the duration as a floating point number of hours.
    // 
	hoursEquivalentOfMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"hoursEquivalentOf", []string{"ns"}, []string{"Int"}, []string{"Float"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	hoursEquivalentOfMethod.PrimitiveCode = builtinHoursEquivalentOf

    // minutesEquivalentOf durationNs Int > Float
    //
    // Returns the duration as a floating point number of minutes.
    // 
	minutesEquivalentOfMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"minutesEquivalentOf", []string{"ns"}, []string{"Int"}, []string{"Float"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	minutesEquivalentOfMethod.PrimitiveCode = builtinMinutesEquivalentOf

    // secondsEquivalentOf durationNs Int > Float
    //
    // Returns the duration as a floating point number of seconds.
    // 
	secondsEquivalentOfMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"secondsEquivalentOf", []string{"ns"}, []string{"Int"}, []string{"Float"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	secondsEquivalentOfMethod.PrimitiveCode = builtinSecondsEquivalentOf	

    // timeParts durationNs Int > h Int m Int s Int ns Int
    //
    // Returns the duration as the number of hours, remaining minutes, remaining seconds, 
    // and remaining nanoseconds
    // 
	timePartsMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"timeParts", []string{"ns"}, []string{"Int"}, []string{"Int","Int","Int","Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timePartsMethod.PrimitiveCode = builtinTimeParts	




    ///////////////////////////////////////////////////////////
    // Persistence functions

	dubMethod, err := RT.CreateMethod("",nil,"dub", []string{"obj", "name"}, []string{"NonPrimitive", "String"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	dubMethod.PrimitiveCode = builtinDub

	summonMethod, err := RT.CreateMethod("",nil,"summon", []string{"name"}, []string{"String"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	summonMethod.PrimitiveCode = builtinSummon

	existsMethod, err := RT.CreateMethod("",nil,"exists", []string{"name"}, []string{"String"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	existsMethod.PrimitiveCode = builtinExists


    ///////////////////////////////////////////////////////////////////
    // Boolean Logical functions - not and or

	notMethod, err := RT.CreateMethod("",nil,"not", []string{"p"}, []string{"Any"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	notMethod.PrimitiveCode = builtinNot    



/*
   TODO Replace these with proper variadic function creations.
*/
	and2Method, err := RT.CreateMethod("",nil,"and", []string{"p1", "p2"}, []string{"Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	and2Method.PrimitiveCode = builtinAnd

	and3Method, err := RT.CreateMethod("",nil,"and", []string{"p1", "p2", "p3"}, []string{"Any", "Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	and3Method.PrimitiveCode = builtinAnd

	and4Method, err := RT.CreateMethod("",nil,"and", []string{"p1", "p2", "p3", "p4"}, []string{"Any", "Any", "Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	and4Method.PrimitiveCode = builtinAnd

	and5Method, err := RT.CreateMethod("",nil,"and", []string{"p1", "p2", "p3", "p4", "p5"}, []string{"Any", "Any", "Any", "Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	and5Method.PrimitiveCode = builtinAnd

	and6Method, err := RT.CreateMethod("",nil,"and", []string{"p1", "p2", "p3", "p4", "p5", "p6"}, []string{"Any", "Any", "Any", "Any", "Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	and6Method.PrimitiveCode = builtinAnd

	and7Method, err := RT.CreateMethod("",nil,"and", []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7"}, []string{"Any", "Any", "Any", "Any", "Any", "Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	and7Method.PrimitiveCode = builtinAnd

	

	or2Method, err := RT.CreateMethod("",nil,"or", []string{"p1", "p2"}, []string{"Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	or2Method.PrimitiveCode = builtinOr

	or3Method, err := RT.CreateMethod("",nil,"or", []string{"p1", "p2", "p3"}, []string{"Any", "Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	or3Method.PrimitiveCode = builtinOr

	or4Method, err := RT.CreateMethod("",nil,"or", []string{"p1", "p2", "p3", "p4"}, []string{"Any", "Any", "Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	or4Method.PrimitiveCode = builtinOr

	or5Method, err := RT.CreateMethod("",nil,"or", []string{"p1", "p2", "p3", "p4", "p5"}, []string{"Any", "Any", "Any", "Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	or5Method.PrimitiveCode = builtinOr

	or6Method, err := RT.CreateMethod("",nil,"or", []string{"p1", "p2", "p3", "p4", "p5", "p6"}, []string{"Any", "Any", "Any", "Any", "Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	or6Method.PrimitiveCode = builtinOr

	or7Method, err := RT.CreateMethod("",nil,"or", []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7"}, []string{"Any", "Any", "Any", "Any", "Any", "Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	or7Method.PrimitiveCode = builtinOr

    









	//////////////////////////////////////////////////////////////////////////////////////
	// String methods - TODO Put them in their own package
	
	// contains s String substr String > Bool	
	//
	stringContainsMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"contains", []string{"s","substr"}, []string{"String","String"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringContainsMethod.PrimitiveCode = builtinStringContains	




/*
replace s String old String new String > String
"""
  Returns a copy of the string s with all non-overlapping instances of old replaced by new. 
"""

replace s String old String new String n Int > String
"""
  Returns a copy of the string s with the first n non-overlapping instances of old replaced by new. 
  If n < 0, there is no limit on the number of replacements.
"""
*/


    stringReplace3Method, err := RT.CreateMethod("relish/pkg/strings",nil,"replace", []string{"s","old","new"}, []string{"String","String","String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringReplace3Method.PrimitiveCode = builtinStringReplace

    stringReplace4Method, err := RT.CreateMethod("relish/pkg/strings",nil,"replace", []string{"s","old","new","n"}, []string{"String","String","String","Int"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringReplace4Method.PrimitiveCode = builtinStringReplace




// join a []String
// """
//  Concatenates the elements of a to create a single string. T
// """
//
// join a []String sep String > String
//
// Should be 
// join a []Any sep String > String
// but we don't have covariant collection type compatibility in input args yet. should have, for immutable input args	


    stringJoin1Method, err := RT.CreateMethod("relish/pkg/strings",nil,"join", []string{"a"}, []string{"List_of_String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringJoin1Method.PrimitiveCode = builtinStringJoin



    stringJoin2Method, err := RT.CreateMethod("relish/pkg/strings",nil,"join", []string{"a","sep"}, []string{"List_of_String","String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringJoin2Method.PrimitiveCode = builtinStringJoin


    stringJoin2AnyMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"join", []string{"a","sep"}, []string{"List_of_Any","String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringJoin2AnyMethod.PrimitiveCode = builtinStringJoin




	
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
    stringFill2Method, err := RT.CreateMethod("relish/pkg/strings",nil,"fill", []string{"s1", "s2"}, []string{"String", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill2Method.PrimitiveCode = builtinStringFill

	stringFill3Method, err := RT.CreateMethod("relish/pkg/strings",nil,"fill", []string{"s1", "s2", "s3"}, []string{"String", "Any", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill3Method.PrimitiveCode = builtinStringFill

	stringFill4Method, err := RT.CreateMethod("relish/pkg/strings",nil,"fill", []string{"s1", "s2", "s3", "s4"}, []string{"String", "Any", "Any", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill4Method.PrimitiveCode = builtinStringFill

	stringFill5Method, err := RT.CreateMethod("relish/pkg/strings",nil,"fill", []string{"s1", "s2", "s3", "s4", "s5"}, []string{"String", "Any", "Any", "Any", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill5Method.PrimitiveCode = builtinStringFill

	stringFill6Method, err := RT.CreateMethod("relish/pkg/strings",nil,"fill", []string{"s1", "s2", "s3", "s4", "s5", "s6"}, []string{"String", "Any", "Any", "Any", "Any", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill6Method.PrimitiveCode = builtinStringFill

	stringFill7Method, err := RT.CreateMethod("relish/pkg/strings",nil,"fill", []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7"}, []string{"String", "Any", "Any", "Any", "Any", "Any", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill7Method.PrimitiveCode = builtinStringFill	

	stringFill8Method, err := RT.CreateMethod("relish/pkg/strings",nil,"fill", []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8"}, []string{"String", "Any", "Any", "Any", "Any", "Any", "Any", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill8Method.PrimitiveCode = builtinStringFill	

		stringFill9Method, err := RT.CreateMethod("relish/pkg/strings",nil,"fill", []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8", "s9"}, []string{"String", "Any", "Any", "Any", "Any", "Any", "Any", "Any", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill9Method.PrimitiveCode = builtinStringFill		


    // length of a string
    //
    // in bytes
	stringLenMethod, err := RT.CreateMethod("",nil,"len", []string{"c"}, []string{"String"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringLenMethod.PrimitiveCode = builtinStringLen
	
    // in characters (codepoints)
    //
	stringNumCodePointsMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"numCodePoints", []string{"c"}, []string{"String"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringNumCodePointsMethod.PrimitiveCode = builtinStringNumCodePoints
	
	// concatenation of strings
	
    // s = cat s1 String s2 String s3 String s4 String s5 String s6 String s7 String s8 String s9 String > String  
	
    stringCat2Method, err := RT.CreateMethod("relish/pkg/strings",nil,"cat", []string{"s1", "s2"}, []string{"Any", "Any"},  []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat2Method.PrimitiveCode = builtinStringCat

	stringCat3Method, err := RT.CreateMethod("relish/pkg/strings",nil,"cat", []string{"s1", "s2", "s3"}, []string{"Any", "Any", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat3Method.PrimitiveCode = builtinStringCat

	stringCat4Method, err := RT.CreateMethod("relish/pkg/strings",nil,"cat", []string{"s1", "s2", "s3", "s4"}, []string{"Any", "Any", "Any", "Any"},  []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat4Method.PrimitiveCode = builtinStringCat

	stringCat5Method, err := RT.CreateMethod("relish/pkg/strings",nil,"cat", []string{"s1", "s2", "s3", "s4", "s5"}, []string{"Any", "Any", "Any", "Any", "Any"},  []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat5Method.PrimitiveCode = builtinStringCat

	stringCat6Method, err := RT.CreateMethod("relish/pkg/strings",nil,"cat", []string{"s1", "s2", "s3", "s4", "s5", "s6"}, []string{"Any", "Any", "Any", "Any", "Any", "Any"},  []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat6Method.PrimitiveCode = builtinStringCat

	stringCat7Method, err := RT.CreateMethod("relish/pkg/strings",nil,"cat", []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7"}, []string{"Any", "Any", "Any", "Any", "Any", "Any", "Any"},  []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat7Method.PrimitiveCode = builtinStringCat	

	stringCat8Method, err := RT.CreateMethod("relish/pkg/strings",nil,"cat", []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8"}, []string{"Any", "Any", "Any", "Any", "Any", "Any", "Any", "Any"},  []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat8Method.PrimitiveCode = builtinStringCat	

		stringCat9Method, err := RT.CreateMethod("relish/pkg/strings",nil,"cat", []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8", "s9"}, []string{"Any", "Any", "Any", "Any", "Any", "Any", "Any", "Any", "Any"},  []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat9Method.PrimitiveCode = builtinStringCat	
	
	
    stringHasPrefixMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"hasPrefix", []string{"s1", "s2"}, []string{"String", "String"},  []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringHasPrefixMethod.PrimitiveCode = builtinStringHasPrefix	
	
    stringHasSuffixMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"hasSuffix", []string{"s1", "s2"}, []string{"String", "String"}, []string{"Bool"}, false, 0, false)
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
    stringIndexMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"index", []string{"s1", "s2"}, []string{"String", "String"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringIndexMethod.PrimitiveCode = builtinStringIndex
	
    stringLastIndexMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"lastIndex", []string{"s1", "s2"}, []string{"String", "String"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringLastIndexMethod.PrimitiveCode = builtinStringLastIndex	


    // substring or slice
/*
	slice s String start Int end Int > String
*/
    stringSlice2Method, err := RT.CreateMethod("relish/pkg/strings",nil,"slice", []string{"s", "start"}, []string{"String", "Int"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringSlice2Method.PrimitiveCode = builtinStringSlice	
	
    stringSlice3Method, err := RT.CreateMethod("relish/pkg/strings",nil,"slice", []string{"s", "start", "end"}, []string{"String", "Int", "Int"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringSlice3Method.PrimitiveCode = builtinStringSlice	
	
    stringFirstMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"first", []string{"s", "n"}, []string{"String", "Int"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringFirstMethod.PrimitiveCode = builtinStringFirst	
	
    stringLastMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"last", []string{"s", "n"}, []string{"String", "Int"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringLastMethod.PrimitiveCode = builtinStringLast	

    stringLowerMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"lower", []string{"s"}, []string{"String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringLowerMethod.PrimitiveCode = builtinStringLower	

    stringUpperMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"upper", []string{"s"}, []string{"String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringUpperMethod.PrimitiveCode = builtinStringUpper	

    stringTitleMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"title", []string{"s"}, []string{"String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringTitleMethod.PrimitiveCode = builtinStringTitle	

    stringTrimSpaceMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"trimSpace", []string{"s"}, []string{"String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringTrimSpaceMethod.PrimitiveCode = builtinStringTrimSpace			

	






	
	
	/*
	    returns a String - the base64-encoded sha25 hash of the input argument String.
	*/
	stringBase64HashMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"base64Hash", []string{"s"}, []string{"String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringBase64HashMethod.PrimitiveCode = builtinStringHashSha256Base64		

	/*
	    returns a String - the hexadecimal-encoded sha25 hash of the input argument String.
	*/
	stringHexHashMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"hexHash", []string{"s"}, []string{"String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringHexHashMethod.PrimitiveCode = builtinStringHashSha256Hex			
	
    ///////////////////////////////////////////////////////////////////
    // Collection functions	
	
	// len coll Collection > Int	
	//
	lenMethod, err := RT.CreateMethod("",nil,"len", []string{"c"}, []string{"Collection"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	lenMethod.PrimitiveCode = builtinLen


	// cap coll Collection > Int	
	//
	capMethod, err := RT.CreateMethod("",nil,"cap", []string{"c"}, []string{"Collection"},  []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	capMethod.PrimitiveCode = builtinCap	


	// contains coll Collection val Any > Bool	
	//
	containsMethod, err := RT.CreateMethod("",nil,"contains", []string{"c","v"}, []string{"Collection","Any"},  []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	containsMethod.PrimitiveCode = builtinContains


	// clear coll Collection 
	//
	clearMethod, err := RT.CreateMethod("",nil,"clear", []string{"c"}, []string{"Collection"},  nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	clearMethod.PrimitiveCode = builtinClear


/*
	slice s List of T start Int end Int > List of T

	One idea: Can we have shared-array slice operators AND copy slice operators?

	Cannot proceed with these things overall til sort our parameterized types.

    listSlice2Method, err := RT.CreateMethod("",nil,"slice", []string{"s", "start"}, []string{"List_of_Any", "Int"}, []string{"List_of_Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	listSlice2Method.PrimitiveCode = builtinListSlice	
	
    listSlice3Method, err := RT.CreateMethod("",nil,"slice", []string{"s", "start", "end"}, []string{"List_of_Any", "Int", "Int"}, []string{"List_of_Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	listSlice3Method.PrimitiveCode = builtinListSlice	
	
    listFirstMethod, err := RT.CreateMethod("",nil,"first", []string{"s", "n"}, []string{"List_of_Any", "Int"}, []string{"List_of_Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	listFirstMethod.PrimitiveCode = builtinListFirst	
	
    listLastMethod, err := RT.CreateMethod("",nil,"last", []string{"s", "n"}, []string{"List_of_Any", "Int"}, []string{"List_of_Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	listLastMethod.PrimitiveCode = builtinListLast		

*/


    ///////////////////////////////////////////////////////////////////
    // Concurrency functions	

	// len c Channel > Int	
	//
	channelLenMethod, err := RT.CreateMethod("",nil,"len", []string{"c"}, []string{"Channel"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	channelLenMethod.PrimitiveCode = builtinChannelLen    

	// cap c Channel > Int	
	//
	channelCapMethod, err := RT.CreateMethod("",nil,"cap", []string{"c"}, []string{"Channel"},  []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	channelCapMethod.PrimitiveCode = builtinChannelCap    	


	// from c InChannel of T > T	
	//
	channelFromMethod, err := RT.CreateMethod("",nil,"<-", []string{"c"}, []string{"InChannel"},  []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	channelFromMethod.PrimitiveCode = builtinFrom   	

/* Deprecated in favour of ch <- operator

    // to c OutChannel of T obj T 	
	//
	channelToMethod, err := RT.CreateMethod("",nil,"to", []string{"c","v"}, []string{"OutChannel","Any"}, 1, false, 0, false)
	if err != nil {
		panic(err)
	}
	channelToMethod.PrimitiveCode = builtinTo 
*/


	mutexLockMethod, err := RT.CreateMethod("",nil,"lock", []string{"m"}, []string{"Mutex"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	mutexLockMethod.PrimitiveCode = builtinMutexLock

	mutexUnlockMethod, err := RT.CreateMethod("",nil,"unlock", []string{"m"}, []string{"Mutex"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	mutexUnlockMethod.PrimitiveCode = builtinMutexUnlock

	rwmutexLockMethod, err := RT.CreateMethod("",nil,"lock", []string{"m"}, []string{"RWMutex"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	rwmutexLockMethod.PrimitiveCode = builtinRWMutexLock

	rwmutexUnlockMethod, err := RT.CreateMethod("",nil,"unlock", []string{"m"}, []string{"RWMutex"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	rwmutexUnlockMethod.PrimitiveCode = builtinRWMutexUnlock

	rwmutexRLockMethod, err := RT.CreateMethod("",nil,"rlock", []string{"m"}, []string{"RWMutex"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	rwmutexRLockMethod.PrimitiveCode = builtinRWMutexRLock

	rwmutexRUnlockMethod, err := RT.CreateMethod("",nil,"runlock", []string{"m"}, []string{"RWMutex"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	rwmutexRUnlockMethod.PrimitiveCode = builtinRWMutexRUnlock




    /////////////////////////////////////////////////////////////////////
    // Type init functions

	timeInit1Method, err := RT.CreateMethod("",nil,"initTime", []string{"t","timeString"}, []string{"Time","String"},  []string{"Time","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeInit1Method.PrimitiveCode = builtinInitTimeParse    
	
	timeInit2Method, err := RT.CreateMethod("",nil,"initTime", []string{"t","timeString","layout"}, []string{"Time","String","String"},  []string{"Time","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeInit2Method.PrimitiveCode = builtinInitTimeParse	

	timeInit3Method, err := RT.CreateMethod("",nil,"initTime", []string{"t","year","month","day","hour","min","sec","nsec","loc"}, []string{"Time","Int","Int","Int","Int","Int","Int","Int","String"},  []string{"Time","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeInit3Method.PrimitiveCode = builtinInitTimeDate	



	channelInit0Method, err := RT.CreateMethod("",nil,"initChannel", []string{"c"}, []string{"Channel"}, []string{"Channel"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	channelInit0Method.PrimitiveCode = builtinInitChannel 

	channelInit1Method, err := RT.CreateMethod("",nil,"initChannel", []string{"c","n"}, []string{"Channel","Int"}, []string{"Channel"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	channelInit1Method.PrimitiveCode = builtinInitChannel	



	stringInitMethod, err := RT.CreateMethod("",nil,"initString", []string{"s","o"}, []string{"String","Any"},  []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringInitMethod.PrimitiveCode = builtinInitString  
}


///////////////////////////////////////////////////////////////////////////////////////////
// I/O functions

func builtinPrint(th InterpreterThread, objects []RObject) []RObject {
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

func builtinInput(th InterpreterThread, objects []RObject) []RObject {
   message := string(objects[0].(String))	
   fmt.Print(message)
   result,err := buf.ReadString('\n')
   if err != nil {
      result = err.Error()
   }
   result = result[:len(result)-1]
   return []RObject{String(result)}
}

func builtinDbg(th InterpreterThread, objects []RObject) []RObject {
   fmt.Println(objects[0].Debug())
   return []RObject{}
}

func builtinDebug(th InterpreterThread, objects []RObject) []RObject {
   s := objects[0].Debug()
   return []RObject{String(s)}
}

//////////////////////////////////////////////////////////////////////////////////////////////
// Comparison functions

/*
Equality operator.
TODO !!!
Really have to study which kinds of equality we want to support and with which function names,
by studying other languages. LISP, Java, Go, Python etc.
*/
func builtinEq(th InterpreterThread, objects []RObject) []RObject {
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
A primitive is not eq to a non-primitive.
*/
func builtinEqNo(th InterpreterThread, objects []RObject) []RObject {	
	return []RObject{Bool(false)}
}

/*
Non-primitives evaluate to eq if their RObject interfaces are == in Go.
This is probably not what we want eventually.
*/
func builtinEqObj(th InterpreterThread, objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	val := Bool(obj1 == obj2)
	return []RObject{val}
}



/*
Less-than operator for numeric types.
TODO !!!
Really have to study which kinds of order we want to support and with which function names,
by studying other languages. LISP, Java, Go, Python etc.
*/
func builtinLtNum(th InterpreterThread, objects []RObject) []RObject {
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


func builtinLteNum(th InterpreterThread, objects []RObject) []RObject {
	if bool( (builtinEq(th,objects))[0].(Bool)) {
		return []RObject{Bool(true)}
	}
	return builtinLtNum(th,objects) 
}





/*
Less-than operator for time types.
*/
func builtinLtTime(th InterpreterThread, objects []RObject) []RObject {
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


func builtinLteTime(th InterpreterThread, objects []RObject) []RObject {
	if bool( (builtinEq(th,objects))[0].(Bool)) {
		return []RObject{Bool(true)}
	}
	return builtinLtTime(th,objects) 
}


/*
Less-than operator (strings).
TODO !!!
Really have to study which kinds of order we want to support and with which function names,
by studying other languages. LISP, Java, Go, Python etc.
*/
func builtinLtStr(th InterpreterThread, objects []RObject) []RObject {
	obj1 := objects[0].(String)
	obj2 := objects[1].(String)
	var val RObject
	val = Bool(string(obj1) < string(obj2))
	return []RObject{val}
}


func builtinLteStr(th InterpreterThread, objects []RObject) []RObject {
	if bool( (builtinEq(th,objects))[0].(Bool)) {
		return []RObject{Bool(true)}
	}
	return builtinLtStr(th,objects) 
}


/*
Less-than operator for numeric types.
TODO !!!
Really have to study which kinds of order we want to support and with which function names,
by studying other languages. LISP, Java, Go, Python etc.
*/
func builtinGtNum(th InterpreterThread, objects []RObject) []RObject {
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


func builtinGteNum(th InterpreterThread, objects []RObject) []RObject {
	if bool( (builtinEq(th,objects))[0].(Bool)) {
		return []RObject{Bool(true)}
	}
	return builtinGtNum(th,objects) 
}


/*
Greater-than operator for time types.
*/
func builtinGtTime(th InterpreterThread, objects []RObject) []RObject {
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


func builtinGteTime(th InterpreterThread, objects []RObject) []RObject {
	if bool( (builtinEq(th,objects))[0].(Bool)) {
		return []RObject{Bool(true)}
	}
	return builtinGtTime(th,objects) 
}


/*
Greater-than operator (strings).
TODO !!!
Really have to study which kinds of order we want to support and with which function names,
by studying other languages. LISP, Java, Go, Python etc.
*/
func builtinGtStr(th InterpreterThread, objects []RObject) []RObject {
	obj1 := objects[0].(String)
	obj2 := objects[1].(String)
	var val RObject
	val = Bool(string(obj1) > string(obj2))
	return []RObject{val}
}


func builtinGteStr(th InterpreterThread, objects []RObject) []RObject {
	if bool( (builtinEq(th,objects))[0].(Bool)) {
		return []RObject{Bool(true)}
	}
	return builtinGtStr(th,objects) 
}


//////////////////////////////////////////////////////////////////////////////////////////////
// Arithmetic functions

/*
Multiply operator. Note: Result type varies (is covariant with argument type variation). All results are Numeric but result is a different
subtype of numeric depending on the argument types.
*/
func builtinTimes(th InterpreterThread, objects []RObject) []RObject {
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
func builtinPlus(th InterpreterThread, objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	return []RObject{plus(obj1,obj2)}
}

/*
Arithmetic Addition operator. Note: Result type varies (is covariant with argument type variation). All results are Numeric but result is a different
subtype of numeric depending on the argument types.
*/
func plus(obj1, obj2 RObject) RObject {
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
	return val
}

/*
Arithmetic sum (of a collection of numbers) operator. Note: Result type varies (is covariant with argument type variation). All results are Numeric but result is a different
subtype of numeric depending on the argument types.

TODO  IMPLEMENT

var ListOfNumericType *RType
var SetOfNumericType *RType

var ListOfFloatType *RType
var SetOfFloatType *RType

var ListOfIntType *RType
var SetOfIntType *RType

var ListOfUIntType *RType
var SetOfUIntType *RType

*/
func builtinSum(th InterpreterThread, objects []RObject) []RObject {
	collection := objects[0].(RCollection)
	t := collection.ElementType()
	var n RObject	
	if collection.Length() == 0 {
		switch t {
		   case FloatType:
			  n = Float(0.)
		   case IntType, Int32Type:
			  n = Int(0)
		   case UintType, Uint32Type:
			  n = Uint(0)
		   case NumericType:
			  n = Float(0.)	
	  	   default:
		     rterr.Stop("wrong element type for sum")				
		}
	} else {
		first := true
		for val := range collection.Iter(th) {
			if first {
				n = val
				first = false
			} else {
			   n = plus(n, val)
		    }
		}
	}	
	
	return []RObject{n}
}


/*
Arithmetic Subtraction operator. Note: Result type varies (is covariant with argument type variation). All results are Numeric but result is a different
subtype of numeric depending on the argument types.
*/
func builtinMinus(th InterpreterThread, objects []RObject) []RObject {
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
func builtinDiv(th InterpreterThread, objects []RObject) []RObject {
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
func builtinMod(th InterpreterThread, objects []RObject) []RObject {
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
func builtinNeg(th InterpreterThread, objects []RObject) []RObject {
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



////////////////////////////////////////////////////////////
// Time and Duration functions

// now location String > Time
//
// t = now "America/Los_Angeles"       t = now "Local"      t = now "UTC"
//
// Will throw a runtime error and halt the relish program if the location string is invalid.
//
func builtinTimeNow(th InterpreterThread, objects []RObject) []RObject {
	loc := string(objects[0].(String))
    location,err := time.LoadLocation(loc) 
    if err != nil {
    	rterr.Stop(err.Error())
    }
    t := time.Now().In(location)
	return []RObject{RTime(t)}
}	


// sleep durationNs Int
//
func builtinSleep(th InterpreterThread, objects []RObject) []RObject {
	d := time.Duration(int64(objects[0].(Int)))
    time.Sleep(d)
	return []RObject{}
}	


// tick durationNs Int > Inchannel of Time
//
func builtinTick(th InterpreterThread, objects []RObject) []RObject {
	d := time.Duration(int64(objects[0].(Int)))
    ch := time.Tick(d)

    c := &TimeChannel{Ch: ch}
 
	return []RObject{c}
}	







// plus t Time durationNs Int > Time
//
func builtinTimePlus(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
	d := time.Duration(int64(objects[1].(Int)))
    t2 := t.Add(d)
	return []RObject{RTime(t2)}
}	

// addDate t Time years Int months Int days Int > Time
//
func builtinTimeAddDate(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
	years := int(objects[1].(Int))
	months := int(objects[2].(Int))
	days := int(objects[3].(Int))		
    t2 := t.AddDate(years,months,days)
	return []RObject{RTime(t2)}
}	

// minus t Time durationNs Int > Time
//
func builtinTimeMinusDuration(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))	
	d := time.Duration(int64(objects[1].(Int)))
    t2 := t.Add(-d)
	return []RObject{RTime(t2)}	
}	

// minus t2 Time t1 Time > durationNs Int
//
func builtinTimeMinusTime(th InterpreterThread, objects []RObject) []RObject {
	t2 := time.Time(objects[0].(RTime))
	t1 := time.Time(objects[1].(RTime))	
    d := t2.Sub(t1)
	return []RObject{Int(int64(d))}			
}		

// since t Time > durationNs Int
//
func builtinTimeSince(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    d := time.Since(t)
	return []RObject{Int(int64(d))}		
}	

// timeIn t Time location String > Time 
// 
// Will throw a runtime error and halt the relish program if the location string is invalid.
//
func builtinTimeIn(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
	loc := string(objects[1].(String))	
    location,err := time.LoadLocation(loc) 
    if err != nil {
    	rterr.Stop(err.Error())
    }
    t2 := t.In(location)
	return []RObject{RTime(t2)}	
}	

// hours n Int > durationNs Int
// 
func builtinHours(th InterpreterThread, objects []RObject) []RObject {
	h := int64(objects[0].(Int))	
	d := h * 3600 * 1000000000 
	return []RObject{Int(d)}		
}	

// minutes n Int > durationNs Int
// 
func builtinMinutes(th InterpreterThread, objects []RObject) []RObject {
	m := int64(objects[0].(Int))	
	d := m * 60 * 1000000000 
	return []RObject{Int(d)}		
}		

// seconds n Int > durationNs Int
// 
func builtinSeconds(th InterpreterThread, objects []RObject) []RObject {
	s := int64(objects[0].(Int))	
	d := s * 1000000000 	
	return []RObject{Int(d)}		
}		

// milliseconds n Int > durationNs Int
// 
func builtinMilliseconds(th InterpreterThread, objects []RObject) []RObject {
	ms := int64(objects[0].(Int))	
	d := ms * 1000000 	
	return []RObject{Int(d)}		
}	
	


// date t Time > year Int month Int day Int
//
func builtinTimeDate(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    year, month, day := t.Date()
	return []RObject{Int(year),Int(month),Int(day)}
}

// clock t Time > hour Int min Int sec Int
//
func builtinTimeClock(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    hour, min, sec := t.Clock()
	return []RObject{Int(hour),Int(min),Int(sec)}
}

// day t Time > dayOfMonth Int  
// """
//  0..31
// """
//
func builtinTimeDay(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    day := t.Day()
	return []RObject{Int(day)}
}

// hour t Time >  Int  
// """
//  0..23
// """
//
func builtinTimeHour(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    hour := t.Hour()
	return []RObject{Int(hour)}
}

// minute t Time >  Int  
// """
//  0..59
// """
//
func builtinTimeMinute(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    min := t.Minute()
	return []RObject{Int(min)}
}

// second t Time >  Int  
// """
//  0..59
// """
//
func builtinTimeSecond(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    sec := t.Second()
	return []RObject{Int(sec)}
}

// nanosecond t Time >  Int  
// """
//  0..999999999
// """
//
func builtinTimeNanosecond(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    nsec := t.Nanosecond()
	return []RObject{Int(nsec)}
}

// weekday t Time >  Int  
// """
//  0..6  Sunday = 0 
// """
//
func builtinTimeWeekday(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    wd := t.Weekday()
	return []RObject{Int(int64(int(wd)))}
}

// year t Time >  Int  
//
func builtinTimeYear(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    year := t.Year()
	return []RObject{Int(year)}
}

// month t Time >  Int  
// """
//  1..12  January = 1
// """
//
func builtinTimeMonth(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    month := t.Month()
	return []RObject{Int(int64(int(month)))}
}


// zone t Time > name String offset Int
// """
//  EST secondsEastOfUTC
// """  
//
func builtinTimeZone(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    zoneName,offset := t.Zone()
	return []RObject{String(zoneName),Int(offset)}
}

// format t Time layout String > String
// """
//  Like Go time layouts (http://golang.org/pkg/time/)
// """  
//
func builtinTimeFormat(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
	layout := string(objects[1].(String))	
    s := t.Format(layout)
	return []RObject{String(s)}
}



// duration hours Int minutes Int > durationNs Int
// 
// duration hours Int minutes Int seconds Int > durationNs Int
// 	
// duration hours Int minutes Int seconds Int nanoseconds Int > durationNs Int
// 
func builtinDuration(th InterpreterThread, objects []RObject) []RObject {
	var h,m,s,ns int64
	h = int64(objects[0].(Int))
	m = int64(objects[1].(Int))
	if len(objects) > 2 {
	   s = int64(objects[0].(Int))
	}
	if len(objects) > 3 {
	   ns = int64(objects[0].(Int))
	}		
	d := h * 3600 * 1000000000 + m * 60 * 1000000000 + s * 1000000000 + ns
	return []RObject{Int(d)}	
}		

func builtinHoursEquivalentOf(th InterpreterThread, objects []RObject) []RObject {
	d := time.Duration(int64(objects[0].(Int)))	
	h := d.Hours()
    return []RObject{Float(h)}	
}

func builtinMinutesEquivalentOf(th InterpreterThread, objects []RObject) []RObject {
	d := time.Duration(int64(objects[0].(Int)))	
	m := d.Minutes()	
    return []RObject{Float(m)}		
}

func builtinSecondsEquivalentOf(th InterpreterThread, objects []RObject) []RObject {
	d := time.Duration(int64(objects[0].(Int)))	
	s := d.Seconds()
    return []RObject{Float(s)}		
}

func builtinTimeParts(th InterpreterThread, objects []RObject) []RObject {
	d := int64(objects[0].(Int))
    h := d / (3600 * 1000000000)	
    excess := d % (3600 * 1000000000)	
    m := excess / (60 * 1000000000)   
    excess = excess % (60 * 1000000000)
    s := excess / 1000000000
    ns := excess % 1000000000
    return []RObject{Int(h),Int(m),Int(s),Int(ns)}	    
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
func builtinDub(th InterpreterThread, objects []RObject) []RObject {

	relish.EnsureDatabase()
	obj := objects[0]
	name := objects[1].String()

	if obj.IsStoredLocally() {
		rterr.Stopf("Object %v is already persistent. Cannot re-dub (re-persist) as '%s'.", obj, name)
	}

	// Ensure that the name is not already used for a persistent object in this database.

	found, err := th.DB().ObjectNameExists(name)
	if err != nil {
		panic(err)
	} else if found {
		rterr.Stopf("The object name '%s' is already in use.", name)
	}

	// Ensure that the object is persisted.

	err = th.DB().EnsurePersisted(obj)
	if err != nil {
		panic(err)
	}

	// Now we have to associate the object with its name in the database. 

	th.DB().NameObject(obj, name)

	return nil
}

/*
summon String > NonPrimitive

car1 = Car: summon "FEC 092"

TODO Add an optional radius argument controlling fetch eagerness of non-primitive attributes and objects related to the object fetched. 
Right now is using radius=0 meaning fetch the object with its primitive (and non-multi) valued attribute values. Other attributes are
currently fetched from DB lazily.

*/
func builtinSummon(th InterpreterThread, objects []RObject) []RObject {
	relish.EnsureDatabase()
	name := objects[0].String()

	obj, err := th.DB().FetchByName(name, 0)
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
func builtinExists(th InterpreterThread, objects []RObject) []RObject {
	relish.EnsureDatabase()
	name := objects[0].String()

	found, err := th.DB().ObjectNameExists(name) 
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
func builtinNot(th InterpreterThread, objects []RObject) []RObject {
	return []RObject{Bool(objects[0].IsZero())}
}

/*
Boolean logic and. Note THIS IS NOT LAZY evaluator of arguments, 
unlike LISP and. Darn. Maybe change that someday.  
Returns false if any argument is zero,
otherwise returns the last argument.
Return type is Any
*/
func builtinAnd(th InterpreterThread, objects []RObject) []RObject {

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
func builtinOr(th InterpreterThread, objects []RObject) []RObject {

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

func builtinLen(th InterpreterThread, objects []RObject) []RObject {
	coll := objects[0].(RCollection)
	var val RObject
	val = Int(coll.Length())
	return []RObject{val}
}

func builtinCap(th InterpreterThread, objects []RObject) []RObject {
	coll := objects[0].(RCollection)
	var val RObject
	val = Int(coll.Cap())
	return []RObject{val}
}

func builtinContains(th InterpreterThread, objects []RObject) []RObject {
	coll := objects[0].(RCollection)
    found := coll.Contains(th, objects[1])
	return []RObject{Bool(found)}
}

func builtinClear(th InterpreterThread, objects []RObject) []RObject {
	coll,isRemovableCollection := objects[0].(RemovableCollection)
    if ! isRemovableCollection {
    	rterr.Stop("Can only apply clear to a mutable,clearable collection or map.")
    }
    coll.ClearInMemory()
	return []RObject{}
}

///////////////////////////////////////////////////////////////////////
// Concurrency functions

/*

from ch InChannel of T > T

val = from ch

TODO DUMMY Implementation 
*/
func builtinFrom(th InterpreterThread, objects []RObject) []RObject {
	c := objects[0].(RInChannel)
    val := c.From()
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

func builtinChannelLen(th InterpreterThread, objects []RObject) []RObject {
	c := objects[0].(*Channel)
	var val RObject
	val = Int(c.Length())
	return []RObject{val}
}

func builtinChannelCap(th InterpreterThread, objects []RObject) []RObject {
	c := objects[0].(*Channel)
	var val RObject
	val = Int(c.Cap())
	return []RObject{val}
}

func builtinMutexLock(th InterpreterThread, objects []RObject) []RObject {
	c := objects[0].(*Mutex)
	c.Lock()
	return []RObject{}
}

func builtinMutexUnlock(th InterpreterThread, objects []RObject) []RObject {
	c := objects[0].(*Mutex)
	c.Unlock()
	return []RObject{}
}

func builtinRWMutexLock(th InterpreterThread, objects []RObject) []RObject {
	c := objects[0].(*RWMutex)
	c.Lock()
	return []RObject{}
}

func builtinRWMutexUnlock(th InterpreterThread, objects []RObject) []RObject {
	c := objects[0].(*RWMutex)
	c.Unlock()
	return []RObject{}
}

func builtinRWMutexRLock(th InterpreterThread, objects []RObject) []RObject {
	c := objects[0].(*RWMutex)
	c.RLock()
	return []RObject{}
}

func builtinRWMutexRUnlock(th InterpreterThread, objects []RObject) []RObject {
	c := objects[0].(*RWMutex)
	c.RUnlock()
	return []RObject{}
}

/////////////////////////////////////////////////////////////// 
// String functions

// length in bytes

func builtinStringLen(th InterpreterThread, objects []RObject) []RObject {
	obj := objects[0].(String)
    s := string(obj)
	var val RObject
	val = Int(int64(len(s)))
	return []RObject{val}
}

// count of unicode codepoints

func builtinStringNumCodePoints(th InterpreterThread, objects []RObject) []RObject {
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
func builtinStringContains(th InterpreterThread, objects []RObject) []RObject {
	s := objects[0].(String)
	substr := objects[1].(String)
	var val RObject
	val = Bool( strings.Contains(string(s), string(substr)) )
	return []RObject{val}
}

func builtinStringHasPrefix(th InterpreterThread, objects []RObject) []RObject {
	s := objects[0].(String)
	substr := objects[1].(String)
	var val RObject
	val = Bool( strings.HasPrefix(string(s), string(substr)) )
	return []RObject{val}
}

func builtinStringHasSuffix(th InterpreterThread, objects []RObject) []RObject {
	s := objects[0].(String)
	substr := objects[1].(String)
	var val RObject
	val = Bool( strings.HasSuffix(string(s), string(substr)) )
	return []RObject{val}
}


/*
replace s String old String new String > String
"""
  Returns a copy of the string s with all non-overlapping instances of old replaced by new. 
"""

replace s String old String new String n Int > String
"""
  Returns a copy of the string s with the first n non-overlapping instances of old replaced by new. 
  If n < 0, there is no limit on the number of replacements.
"""
*/

func builtinStringReplace(th InterpreterThread, objects []RObject) []RObject {
    var n int = -1
	s := string(objects[0].(String))
	oldPiece := string(objects[1].(String))
	newPiece := string(objects[2].(String))	
	if len(objects) == 4 {
		n = int(objects[3].(Int))
	}
	s2 := strings.Replace(s, oldPiece, newPiece, n)
	return []RObject{String(s2)}
}


/*
join a []String
"""
 Concatenates the elements of a to create a single string. T
"""

join a []Any sep String > String
"""
 Concatenates the elements of a to create a single string. The separator string sep
 is placed between elements in the resulting string.
"""
*/

func builtinStringJoin(th InterpreterThread, objects []RObject) []RObject {
    list := objects[0].(List)
    objSlice := list.AsSlice(th)
    stringSlice := make([]string,len(objSlice))
    for i,obj := range objSlice {
    	stringSlice[i] = obj.String()
    }
    sep := ""
	if len(objects) == 2 {
		sep = string(objects[1].(String))
    }
	s2 := strings.Join(stringSlice, sep)
	return []RObject{String(s2)}
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


func builtinStringFill(th InterpreterThread, objects []RObject) []RObject {
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

func builtinStringCat(th InterpreterThread, objects []RObject) []RObject {
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
	
func builtinStringIndex(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	substr := string(objects[1].(String))
	var val RObject
    val = Int(int64(strings.Index(s, substr)))	
	return []RObject{val}
}


func builtinStringLastIndex(th InterpreterThread, objects []RObject) []RObject {
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
func builtinStringSlice(th InterpreterThread, objects []RObject) []RObject {
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
func builtinStringFirst(th InterpreterThread, objects []RObject) []RObject {
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
func builtinStringLast(th InterpreterThread, objects []RObject) []RObject {
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




	
func builtinStringLower(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	lowered := strings.ToLower(s)  
    return []RObject{String(lowered)}	
}	

func builtinStringUpper(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	capitalized := strings.ToUpper(s)      
    return []RObject{String(capitalized)}	
}	

func builtinStringTitle(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	t := strings.Title(strings.ToLower(s))   
    return []RObject{String(t)}	
}	

func builtinStringTrimSpace(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	trimmed := strings.TrimSpace(s)  
    return []RObject{String(trimmed)}	
}	



func builtinStringHashSha256Base64(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	hasher := sha256.New()
	hasher.Write([]byte(s))
	sha := hasher.Sum(nil)
	b64 := base64.URLEncoding.EncodeToString(sha)     
    return []RObject{String(b64)}	
}

func builtinStringHashSha256Hex(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	hasher := sha256.New()
	hasher.Write([]byte(s))
	sha := hasher.Sum(nil)
	hex := fmt.Sprintf("%x",sha)   
    return []RObject{String(hex)}	
}

/*


sendEmail 
   smtpServerAddr String 
   from String 
   recipient String 
   subject String 
   messageBody String 
> 
   error String
"""
 Send an email message to a single recipient, without attempting username+password authentication 
 with the sending smtp mail server.
 Returns an empty String "" if it succeeds, otherwise a mail protocol error message.
"""


sendEmail 
   smtpServerAddr String 
   from String 
   recipients [] String 
   subject String 
   messageBody String 
> 
   error String
"""
 Send an email message to possibly multiple recipients, 
 without attempting username+password authentication with the sending smtp mail server.
 Returns an empty String "" if it succeeds, otherwise a mail protocol error message.
"""


sendEmail 
   smtpServerAddr String 
   sendingMailAccountUserName String
   sendingMailAccountPassword String
   from String 
   recipient String 
   subject String 
   messageBody String 
> 
   error String
"""
 Send an email message to a single recipient, first authenticating with the sending smtp mail server
 to establish that we are a permitted sender by using "plain auth" username+password authentication.
 Returns an empty String "" if it succeeds, otherwise a mail protocol error message.
"""


sendEmail 
   smtpServerAddr String 
   sendingMailAccountUserName String
   sendingMailAccountPassword String
   from String 
   recipients [] String 
   subject String 
   messageBody String 
> 
   error String
"""
 Send an email message to possibly multiple recipients, first authenticating with the sending smtp mail server
 to establish that we are a permitted sender by using "plain auth" username+password authentication.
 Returns an empty String "" if it succeeds, otherwise a mail protocol error message.
"""


email4Method, err := RT.CreateMethod("",nil,
                                     "sendEmail", 
                                     []string{"smtpServerAddr","user","password","from","recipients","subject","messageBody"}, 
                                     []string{"String","String","String","String","List_of_String","String","String"}, 
                                     []string{"String"}, false, 0, false)
*/
func builtinSendEmail(th InterpreterThread, objects []RObject) []RObject {
   var auth smtp.Auth = nil
   var serverAddr string 
   var serverName string // without the :port part
   var recipients [] string
   var errStr string
   var offset int = 0
   serverAddr = string(objects[0].(String))
   colonPos := strings.LastIndex(serverAddr,":")
   if colonPos > 0 {
      serverName = serverAddr[:colonPos]	
   } else {
      serverName = serverAddr	
   }
  
   if len(objects) == 7 { // username and password authentication 
	  userName := string(objects[1].(String))
	  password := string(objects[2].(String))
	  if userName != "" {
	     auth = smtp.PlainAuth("", userName, password, serverName)
	  }
	  offset = 2
   } 
   from := string(objects[1 + offset].(String))
   switch objects[2 + offset].(type) {
      case String:
	     recipients = []string{string(objects[2 + offset].(String))}
	  default:	
	     coll :=  objects[2 + offset].(RCollection)
	     for val := range coll.Iter(th) {
		     recipients = append(recipients, string(val.(String)))
	     }
   }
   subject := string(objects[3 + offset].(String)) 
   body := string(objects[4 + offset].(String))

   fromHeader := "From: " + from
   toHeader := "To: "
   sep := ""
   for _,recipient := range recipients {
	  toHeader += sep + recipient
	  sep = ","
   }
   headers := "Subject: " + subject + "\r\n" + fromHeader + "\r\n" + toHeader +  "\r\n\r\n"

   err := smtp.SendMail(serverAddr, 
	                    auth,
                        from,
                        recipients,
                        []byte(headers + body))
   
   if err != nil {
      	errStr = err.Error()
   }

   return []RObject{String(errStr)}
}




// httpGet url String > responseBody String err String
//
//
func builtinHttpGet(th InterpreterThread, objects []RObject) []RObject {
	
	urlStr := string(objects[0].(String))
	
    contentStr := ""
    errStr := ""

	res, err := http.Get(urlStr)
	if err != nil {
		errStr = err.Error()
	} else {
		content, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			errStr = err.Error()
		}
		contentStr = string(content)
    }	
	return []RObject{String(contentStr),String(errStr)}
}



// httpPost 
//    url String 
//    keysVals {} String > Any
// > 
//    responseBody String 
//    err String
// """
//  The keysVals map should be a map from String to (value or collection of values)
//  The values are converted to their default String representation to be used as
//  http post argument values.
//   
// """
//
func builtinHttpPost(th InterpreterThread, objects []RObject) []RObject {
	
	urlStr := string(objects[0].(String))
	
	theMap := objects[1].(Map)
	
    contentStr := ""
    errStr := ""

	values := url.Values{}
    for key := range theMap.Iter(th) {
	   keyStr := string(key.(String))
	   val, _ := theMap.Get(key)	
       if val.IsCollection() {
		    coll := val.(RCollection)
		    for obj := range coll.Iter(th) {
			   valStr := obj.String()
		       values.Add(keyStr, valStr)			
		    }
	    } else {
	       values.Set(keyStr, val.String())
	   }
    }

	res, err := http.PostForm(urlStr, values)
	if err != nil {
		errStr = err.Error()
	} else {
		content, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			errStr = err.Error()
		}
		contentStr = string(content)
    }	
	return []RObject{String(contentStr),String(errStr)}
}



// csvRead
//    content String 
// > 
//    records [] [] String
//    err String
// """
//  
// """
//
func builtinCsvRead(th InterpreterThread, objects []RObject) []RObject {
   content := string(objects[0].(String))
   stringReader := strings.NewReader(content) 
   csvReader := csv.NewReader(stringReader)
   csvReader.TrailingComma = true
   errStr := ""
   recordsList, err := RT.Newrlist(ListOfStringType, 0, -1, nil, nil)
   if err != nil {
      errStr = err.Error()
   } else {
	   records, err := csvReader.ReadAll()
	   if err != nil {
	      errStr = err.Error()
	   } 
	   recordsList, err := RT.Newrlist(ListOfStringType, 0, -1, nil, nil)
	   if err != nil {
	      errStr = err.Error()	
	   } else {
		   for _,record := range records {
		      recordList, err := RT.Newrlist(StringType, 0, -1, nil, nil)
		      if err != nil {
			     errStr = err.Error()
			     break
		      }
		      for _,field := range record {
			     recordList.AddSimple(String(field))
		      }
		      recordsList.AddSimple(recordList)
		   }
	   }
   }
   return []RObject{recordsList, String(errStr)}
}
	
	
// csvWrite
//    records [] [] String 
// > 
//    content String 
//    err String
// """
//  
// """
//
func builtinCsvWrite(th InterpreterThread, objects []RObject) []RObject {
	return nil
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

func builtinInitTimeParse(th InterpreterThread, objects []RObject) []RObject {
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
func builtinInitTimeDate(th InterpreterThread, objects []RObject) []RObject {

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



func builtinInitChannel(th InterpreterThread, objects []RObject) []RObject {
   
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


/*
Convert any object to a String.
Note: Ignoring string methods defined for programmer-defined datatypes for the moment,
but should probably call that if it exists.
For now, call the builtin string-ification of object Go method.
*/
func builtinInitString(th InterpreterThread, objects []RObject) []RObject {
   
    // ignore the first String argument
    obj := objects[1]
	return []RObject{String(obj.String())}
}