// Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.

package data

import ( 
	     "strconv"
         "errors"
         . "time"
         "fmt"
         "sync"
)

//"fmt"

/*
STRATEGY FOR PRIMITIVES

This is a rubbishy rambling comment

See if the following is possible:

Create an RUnit interface, also an RTyped interface??

Now start defining primitive types in go. Make sure they implement RObject, RTyped, RUnit

NOTE!!!! Primitives cannot be persisted by themselves.
So they cannot have a UUID.
They must be part of something else.

Is this true?

What about global (universal) URN-named constant values.

Or global  (universal) URN-named variables of primitive type:
e.g. GlobalStore.Get("Canada Debt Clock")          (a uint64 i.e. Uint value)

I think in this special case we may need a boxing wrapper type for each primitive type?
ONLY for the purpose of persisting it independently. NEED TO THINK FURTHER ABOUT THIS!!!!
*/

var PrimitiveType *RType
var NumericType *RType
var IntegerType *RType
var IntType *RType
var Int32Type *RType
var Int16Type *RType
var Int8Type *RType
var UintType *RType
var Uint32Type *RType
var Uint16Type *RType
var ByteType *RType
var BitType *RType
var BoolType *RType
var CodePointType *RType
var RealType *RType
var FloatType *RType
var Float32Type *RType
var ComplexNumberType *RType
var ComplexType *RType
var Complex32Type *RType
var TextType *RType
var StringType *RType
var TimeType *RType
var ProxyType *RType

var CallableType *RType
var MultiMethodType *RType
var MethodType *RType

var AnyType *RType
var NothingType *RType
var NonPrimitiveType *RType
var StructType *RType
var CollectionType *RType
var SetType *RType
var ListType *RType
var MapType *RType

var InChannelType *RType
var OutChannelType *RType
var ChannelType *RType
var MutexType *RType
var RWMutexType *RType


var ListOfAnyType *RType
var SetOfAnyType *RType

var ListOfStringType *RType
var SetOfStringType *RType

var ListOfNumericType *RType
var SetOfNumericType *RType

var ListOfFloatType *RType
var SetOfFloatType *RType

var ListOfIntType *RType
var SetOfIntType *RType

var ListOfUIntType *RType
var SetOfUIntType *RType


func (rt *RuntimeEnv) createPrimitiveTypes() {

	PrimitiveType, _ = rt.CreateType("RelishPrimitive", "", []string{})
	NumericType, _ = rt.CreateType("Numeric", "", []string{"RelishPrimitive"})
	IntegerType, _ = rt.CreateType("Integer", "", []string{"Numeric"})
	IntType, _ = rt.CreateType("Int", "", []string{"Integer"})
	Int32Type, _ = rt.CreateType("Int32", "", []string{"Integer"})
	Int16Type, _ = rt.CreateType("Int16", "", []string{"Integer"})
	Int8Type, _ = rt.CreateType("Int8", "", []string{"Integer"})
	UintType, _ = rt.CreateType("Uint", "", []string{"Integer"})
	Uint32Type, _ = rt.CreateType("Uint32", "", []string{"Integer"})
	Uint16Type, _ = rt.CreateType("Uint16", "", []string{"Integer"})
	ByteType, _ = rt.CreateType("Byte", "", []string{"Integer"})
	BitType, _ = rt.CreateType("Bit", "", []string{"Integer"})
	BoolType, _ = rt.CreateType("Bool", "", []string{"RelishPrimitive"})
	TextType, _ = rt.CreateType("Text", "", []string{"RelishPrimitive"})
	CodePointType, _ = rt.CreateType("CodePoint", "", []string{"Text"})
	RealType, _ = rt.CreateType("Real", "", []string{"Numeric"})
	FloatType, _ = rt.CreateType("Float", "", []string{"Real"})
	Float32Type, _ = rt.CreateType("Float32", "", []string{"Real"})
	ComplexNumberType, _ = rt.CreateType("ComplexNumber", "", []string{"Numeric"})
	ComplexType, _ = rt.CreateType("Complex", "", []string{"ComplexNumber"})
	Complex32Type, _ = rt.CreateType("Complex32", "", []string{"ComplexNumber"})
	StringType, _ = rt.CreateType("String", "", []string{"Text"})
	TimeType, _ = rt.CreateType("Time", "", []string{"RelishPrimitive"})
	ProxyType, _ = rt.CreateType("Proxy", "", []string{})

	CallableType, _ = rt.CreateType("Callable", "", []string{"Text"})
	MultiMethodType, _ = rt.CreateType("MultiMethod", "", []string{"Callable"})
	MethodType, _ = rt.CreateType("Method", "", []string{"Callable"})
	// Do I need a "Closure" type???

	AnyType, _ = rt.CreateType("Any", "", []string{})
	NothingType, _ = rt.CreateType("Nothing", "", []string{})	
	NonPrimitiveType, _ = rt.CreateType("NonPrimitive", "", []string{})
	StructType, _ = rt.CreateType("Struct", "", []string{})
	CollectionType, _ = rt.CreateType("Collection", "", []string{})
	SetType, _ = rt.CreateType("Set", "", []string{"Collection"})	
	ListType, _ = rt.CreateType("List", "", []string{"Collection"})	
	MapType, _ = rt.CreateType("Map", "", []string{"Collection"})	

	// Do a need Iterable and InChannel <: Iterable, Collection <: Iterable
	
	InChannelType, _ = rt.CreateType("InChannel", "", []string{})
	OutChannelType, _ = rt.CreateType("OutChannel", "", []string{})
	ChannelType, _ = rt.CreateType("Channel", "", []string{"InChannel","OutChannel"})	
	ChannelType.IsParameterized = true	

	MutexType, _ = rt.CreateType("Mutex", "", []string{})

	RWMutexType, _ = rt.CreateType("RWMutex", "", []string{})	


    // Primitive collection types

	ListOfAnyType, _ = rt.GetListType(AnyType)
	SetOfAnyType, _  = rt.GetSetType(AnyType)	

	ListOfStringType, _ = rt.GetListType(StringType)
	SetOfStringType, _  = rt.GetSetType(StringType)	

	ListOfNumericType, _ = rt.GetListType(NumericType)
	SetOfNumericType, _  = rt.GetSetType(NumericType)	

	ListOfFloatType, _ = rt.GetListType(FloatType)
	SetOfFloatType, _  = rt.GetSetType(FloatType)			

	ListOfIntType, _ = rt.GetListType(IntType)
	SetOfIntType, _  = rt.GetSetType(IntType)	

	ListOfUintType, _ = rt.GetListType(UintType)
	SetOfUintType, _  = rt.GetSetType(UintType)		


}


/*
Return a zero-value of the type.
For structured object types, this is currently defined to be NIL which is a Go constant in this data package.
NIL is of go-type Nil, of relish RType NothingType. 
NIL's name in relish source code is nil, and its string representation when printed is *nil* 
*/
func (t *RType) Zero() RObject {
	var z RObject
    switch t {
	case PrimitiveType:
		z = Int(0)	
	case NumericType:
		z = Int(0)	
	case IntegerType:
		z = Int(0)	
	case IntType:
		z = Int(0)
    case Int32Type:
    	z = Int32(0)
/*
    case Int16Type:
    	z = Int16(0)
    case Int8Type:
    	z = Int8(0)
*/
    case UintType:
    	z = Uint(0)
    case Uint32Type:
        z = Uint(0)	
/*
    case Uint16Type:
    	z = Uint16(0)
    case ByteType:
    	z = Byte(0)    	
    case BitType:
    	z = Bit(0)
*/
    case BoolType:
    	z = Bool(false)
    case TextType:
    	z = String("")
/*
    case CodePointType:
    	z = CodePoint(0)
*/
    case RealType:
        z = Float(0.)
    case FloatType:
        z = Float(0.)    	
/*
    case Float32Type:
        z = Float32(0.)    	
*/
    case ComplexNumberType:
    case ComplexType:
        z = Complex(0+0i)    	
    case Complex32Type:
        z = Complex32(0+0i)       	
    case StringType:
    	z = String("")
    case TimeType:
    	var t Time
    	z = RTime(t)
    case AnyType:
		z = Int(0)	    	
    default:
    	z = NIL   // Hmmm. Do I need one Nil per RType???? With a KnownType attribute?
    }   
    return z
}









type RInChannel interface {
	RObject

	From() RObject
}



type Channel struct {
	Ch chan RObject
	ElementType *RType
}

func (c *Channel) From() RObject {
   return <- c.Ch
}

func (c *Channel) To(val RObject) {
   c.Ch <- val
}

// TODO

func (p Channel) IsZero() bool {
	return p.Ch == nil
}

func (p Channel) Type() *RType {
	return ChannelType
}

func (p Channel) This() RObject {
	return p
}

/*
Hmmm. TODO
*/
func (p Channel) IsUnit() bool {
	return true
}

/*
Hmmm. TODO Maybe a Channel should be considered a collection???
*/
func (p Channel) IsCollection() bool {
	return false
}

func (p Channel) String() string {
	var descriptor string
	if p.Ch == nil {
		descriptor = "uninitialized"
	} else if cap(p.Ch) > 0 {
		descriptor = fmt.Sprintf("cap: %d len: %d",cap(p.Ch),len(p.Ch))
	} else {
		descriptor = "synchronous"
	}
	return fmt.Sprintf("Channel (%s) of %v", descriptor, p.ElementType)
}

func (p Channel) Length() int64 {
	if p.Ch == nil {
		return 0
	}
	return int64(len(p.Ch))
}

func (p Channel) Cap() int64 {
	if p.Ch == nil {
		return 0
	}
	return int64(cap(p.Ch))
}


func (p Channel) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Channel) UUID() []byte {
	panic("A Channel cannot have a UUID.")
	return nil
}

func (p Channel) DBID() int64 {
	panic("A Channel cannot have a DBID.")
	return 0
}

func (p Channel) EnsureUUID() (theUUID []byte, err error) {
	panic("A Channel cannot have a UUID.")
	return
}

func (p Channel) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Channel cannot have a UUID.")
	return
}

func (p Channel) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Channel cannot have a UUID.")
	return
}

func (p Channel) UUIDstr() string {
	panic("A Channel cannot have a UUID.")
	return ""
}

func (p Channel) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Channel cannot have a UUID.")
	return
}

func (p Channel) UUIDabbrev() string {
	panic("A Channel cannot have a UUID.")
	return ""
}

func (p Channel) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Channel cannot have a UUID.")
	return
}

func (p Channel) RemoveUUID() {
	panic("A Channel does not have a UUID.")
	return
}

func (p Channel) Flags() int8 {
	panic("A Channel has no Flags.")
	return 0
}

func (p Channel) IsDirty() bool {
	return false
}
func (p Channel) SetDirty() {
}
func (p Channel) ClearDirty() {
}

func (p Channel) IsIdReversed() bool {
	return false
}

func (p Channel) SetIdReversed() {}

func (p Channel) ClearIdReversed() {}

func (p Channel) IsLoadNeeded() bool {
	return false
}

func (p Channel) SetLoadNeeded()   {}
func (p Channel) ClearLoadNeeded() {}

func (p Channel) IsValid() bool { return true }
func (p Channel) SetValid()     {}
func (p Channel) ClearValid()   {}

func (p Channel) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Channel) SetStoredLocally()     {}
func (p Channel) ClearStoredLocally()   {}

func (p Channel) IsProxy() bool { return false }

func (p Channel) IsTransient() bool { return true }


func (p Channel) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}





type TimeChannel struct {
	Ch <-chan Time
	ElementType *RType
}

func (c *TimeChannel) From() RObject {
   return RTime(<- c.Ch)
}


// TODO

func (p TimeChannel) IsZero() bool {
	return p.Ch == nil
}

func (p TimeChannel) Type() *RType {
	return ChannelType
}

func (p TimeChannel) This() RObject {
	return p
}

/*
Hmmm. TODO
*/
func (p TimeChannel) IsUnit() bool {
	return true
}

/*
Hmmm. TODO Maybe a TimeChannel should be considered a collection???
*/
func (p TimeChannel) IsCollection() bool {
	return false
}

func (p TimeChannel) String() string {
	var descriptor string
	if p.Ch == nil {
		descriptor = "uninitialized"
	} else if cap(p.Ch) > 0 {
		descriptor = fmt.Sprintf("cap: %d len: %d",cap(p.Ch),len(p.Ch))
	} else {
		descriptor = "synchronous"
	}
	return fmt.Sprintf(" Channel (%s) of %v", descriptor, p.ElementType)
}

func (p TimeChannel) Length() int64 {
	if p.Ch == nil {
		return 0
	}
	return int64(len(p.Ch))
}

func (p TimeChannel) Cap() int64 {
	if p.Ch == nil {
		return 0
	}
	return int64(cap(p.Ch))
}


func (p TimeChannel) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p TimeChannel) UUID() []byte {
	panic("A TimeChannel cannot have a UUID.")
	return nil
}

func (p TimeChannel) DBID() int64 {
	panic("A TimeChannel cannot have a DBID.")
	return 0
}

func (p TimeChannel) EnsureUUID() (theUUID []byte, err error) {
	panic("A TimeChannel cannot have a UUID.")
	return
}

func (p TimeChannel) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A TimeChannel cannot have a UUID.")
	return
}

func (p TimeChannel) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A TimeChannel cannot have a UUID.")
	return
}

func (p TimeChannel) UUIDstr() string {
	panic("A TimeChannel cannot have a UUID.")
	return ""
}

func (p TimeChannel) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A TimeChannel cannot have a UUID.")
	return
}

func (p TimeChannel) UUIDabbrev() string {
	panic("A TimeChannel cannot have a UUID.")
	return ""
}

func (p TimeChannel) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A TimeChannel cannot have a UUID.")
	return
}

func (p TimeChannel) RemoveUUID() {
	panic("A TimeChannel does not have a UUID.")
	return
}

func (p TimeChannel) Flags() int8 {
	panic("A TimeChannel has no Flags.")
	return 0
}

func (p TimeChannel) IsDirty() bool {
	return false
}
func (p TimeChannel) SetDirty() {
}
func (p TimeChannel) ClearDirty() {
}

func (p TimeChannel) IsIdReversed() bool {
	return false
}

func (p TimeChannel) SetIdReversed() {}

func (p TimeChannel) ClearIdReversed() {}

func (p TimeChannel) IsLoadNeeded() bool {
	return false
}

func (p TimeChannel) SetLoadNeeded()   {}
func (p TimeChannel) ClearLoadNeeded() {}

func (p TimeChannel) IsValid() bool { return true }
func (p TimeChannel) SetValid()     {}
func (p TimeChannel) ClearValid()   {}

func (p TimeChannel) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p TimeChannel) SetStoredLocally()     {}
func (p TimeChannel) ClearStoredLocally()   {}

func (p TimeChannel) IsProxy() bool { return false }

func (p TimeChannel) IsTransient() bool { return true }


func (p TimeChannel) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}






type Mutex sync.Mutex

// TODO

func (p Mutex) IsZero() bool {
	return sync.Mutex(p) == sync.Mutex{}  // unlocked mutex
}

func (p Mutex) Type() *RType {
	return MutexType
}

func (p Mutex) This() RObject {
	return p
}

/*
Hmmm. TODO
*/
func (p Mutex) IsUnit() bool {
	return true
}

/*
Hmmm. TODO Maybe a Mutex should be considered a collection???
*/
func (p Mutex) IsCollection() bool {
	return false
}

func (p Mutex) String() string {
   return fmt.Sprintf("%v",sync.Mutex(p))
}

func (p Mutex) HasUUID() bool {
	return false
}


/*
   TODO We have to figure out what to do with this.
*/
func (p Mutex) UUID() []byte {
	panic("A Mutex cannot have a UUID.")
	return nil
}

func (p Mutex) DBID() int64 {
	panic("A Mutex cannot have a DBID.")
	return 0
}

func (p Mutex) EnsureUUID() (theUUID []byte, err error) {
	panic("A Mutex cannot have a UUID.")
	return
}

func (p Mutex) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Mutex cannot have a UUID.")
	return
}

func (p Mutex) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Mutex cannot have a UUID.")
	return
}

func (p Mutex) UUIDstr() string {
	panic("A Mutex cannot have a UUID.")
	return ""
}

func (p Mutex) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Mutex cannot have a UUID.")
	return
}

func (p Mutex) UUIDabbrev() string {
	panic("A Mutex cannot have a UUID.")
	return ""
}

func (p Mutex) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Mutex cannot have a UUID.")
	return
}

func (p Mutex) RemoveUUID() {
	panic("A Mutex does not have a UUID.")
	return
}

func (p Mutex) Flags() int8 {
	panic("A Mutex has no Flags.")
	return 0
}

func (p Mutex) IsDirty() bool {
	return false
}
func (p Mutex) SetDirty() {
}
func (p Mutex) ClearDirty() {
}

func (p Mutex) IsIdReversed() bool {
	return false
}

func (p Mutex) SetIdReversed() {}

func (p Mutex) ClearIdReversed() {}

func (p Mutex) IsLoadNeeded() bool {
	return false
}

func (p Mutex) SetLoadNeeded()   {}
func (p Mutex) ClearLoadNeeded() {}

func (p Mutex) IsValid() bool { return true }
func (p Mutex) SetValid()     {}
func (p Mutex) ClearValid()   {}

func (p Mutex) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Mutex) SetStoredLocally()     {}
func (p Mutex) ClearStoredLocally()   {}

func (p Mutex) IsProxy() bool { return false }

func (p Mutex) IsTransient() bool { return true }


func (p Mutex) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (p *Mutex) Lock() {
	(*sync.Mutex)(p).Lock() 
}

func (p *Mutex) Unlock() {
	(*sync.Mutex)(p).Unlock() 
}





type RWMutex sync.RWMutex

// TODO

func (p RWMutex) IsZero() bool {
	return sync.RWMutex(p) == sync.RWMutex{}  // unlocked mutex
}

func (p RWMutex) Type() *RType {
	return RWMutexType
}

func (p RWMutex) This() RObject {
	return p
}

func (p RWMutex) IsUnit() bool {
	return true
}

/*
Hmmm. TODO Maybe a RWMutex should be considered a collection???
*/
func (p RWMutex) IsCollection() bool {
	return false
}

func (p RWMutex) String() string {
   return fmt.Sprintf("%v",sync.RWMutex(p))
}

func (p RWMutex) HasUUID() bool {
	return false
}


/*
   TODO We have to figure out what to do with this.
*/
func (p RWMutex) UUID() []byte {
	panic("A RWMutex cannot have a UUID.")
	return nil
}

func (p RWMutex) DBID() int64 {
	panic("A RWMutex cannot have a DBID.")
	return 0
}

func (p RWMutex) EnsureUUID() (theUUID []byte, err error) {
	panic("A RWMutex cannot have a UUID.")
	return
}

func (p RWMutex) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A RWMutex cannot have a UUID.")
	return
}

func (p RWMutex) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A RWMutex cannot have a UUID.")
	return
}

func (p RWMutex) UUIDstr() string {
	panic("A RWMutex cannot have a UUID.")
	return ""
}

func (p RWMutex) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A RWMutex cannot have a UUID.")
	return
}

func (p RWMutex) UUIDabbrev() string {
	panic("A RWMutex cannot have a UUID.")
	return ""
}

func (p RWMutex) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A RWMutex cannot have a UUID.")
	return
}

func (p RWMutex) RemoveUUID() {
	panic("A RWMutex does not have a UUID.")
	return
}

func (p RWMutex) Flags() int8 {
	panic("A RWMutex has no Flags.")
	return 0
}

func (p RWMutex) IsDirty() bool {
	return false
}
func (p RWMutex) SetDirty() {
}
func (p RWMutex) ClearDirty() {
}

func (p RWMutex) IsIdReversed() bool {
	return false
}

func (p RWMutex) SetIdReversed() {}

func (p RWMutex) ClearIdReversed() {}

func (p RWMutex) IsLoadNeeded() bool {
	return false
}

func (p RWMutex) SetLoadNeeded()   {}
func (p RWMutex) ClearLoadNeeded() {}

func (p RWMutex) IsValid() bool { return true }
func (p RWMutex) SetValid()     {}
func (p RWMutex) ClearValid()   {}

func (p RWMutex) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p RWMutex) SetStoredLocally()     {}
func (p RWMutex) ClearStoredLocally()   {}

func (p RWMutex) IsProxy() bool { return false }

func (p RWMutex) IsTransient() bool { return true }


func (p RWMutex) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (p *RWMutex) Lock() {
	(*sync.RWMutex)(p).Lock() 
}

func (p *RWMutex) Unlock() {
	(*sync.RWMutex)(p).Unlock() 
}

func (p *RWMutex) RLock() {
	(*sync.RWMutex)(p).Lock() 
}

func (p *RWMutex) RUnlock() {
	(*sync.RWMutex)(p).Unlock() 
}









type RTime Time

func (p RTime) IsZero() bool {
	return Time(p).IsZero()
}

func (p RTime) Type() *RType {
	return TimeType
}

func (p RTime) This() RObject {
	return p
}

func (p RTime) IsUnit() bool {
	return true
}

func (p RTime) IsCollection() bool {
	return false
}

func (p RTime) String() string {
	return Time(p).String() // TODO May want to change this: Formats as: "2006-01-02 15:04:05.999999999 -0700 MST"
}

func (p RTime) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p RTime) UUID() []byte {
	panic("An RTime cannot have a UUID.")
	return nil
}

func (p RTime) DBID() int64 {
	panic("An RTime cannot have a DBID.")
	return 0
}

func (p RTime) EnsureUUID() (theUUID []byte, err error) {
	panic("An RTime cannot have a UUID.")
	return
}

func (p RTime) UUIDuint64s() (id uint64, id2 uint64) {
	panic("An RTime cannot have a UUID.")
	return
}

func (p RTime) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("An RTime cannot have a UUID.")
	return
}

func (p RTime) UUIDstr() string {
	panic("An RTime cannot have a UUID.")
	return ""
}

func (p RTime) EnsureUUIDstr() (uuidstr string, err error) {
	panic("An RTime cannot have a UUID.")
	return
}

func (p RTime) UUIDabbrev() string {
	panic("An RTime cannot have a UUID.")
	return ""
}

func (p RTime) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("An RTime cannot have a UUID.")
	return
}

func (p RTime) RemoveUUID() {
	panic("An RTime does not have a UUID.")
	return
}

func (p RTime) Flags() int8 {
	panic("An RTime has no Flags.")
	return 0
}

func (p RTime) IsDirty() bool {
	return false
}
func (p RTime) SetDirty() {
}
func (p RTime) ClearDirty() {
}

func (p RTime) IsIdReversed() bool {
	return false
}

func (p RTime) SetIdReversed() {}

func (p RTime) ClearIdReversed() {}

func (p RTime) IsLoadNeeded() bool {
	return false
}

func (p RTime) SetLoadNeeded()   {}
func (p RTime) ClearLoadNeeded() {}

func (p RTime) IsValid() bool { return true }
func (p RTime) SetValid()     {}
func (p RTime) ClearValid()   {}

func (p RTime) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p RTime) SetStoredLocally()     {}
func (p RTime) ClearStoredLocally()   {}

func (p RTime) IsProxy() bool { return false }

func (p RTime) IsTransient() bool { return false }

func (p RTime) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}










type Int int64

func (p Int) IsZero() bool {
	return p == 0
}

func (p Int) Type() *RType {
	return IntType
}

func (p Int) This() RObject {
	return p
}

func (p Int) IsUnit() bool {
	return true
}

func (p Int) IsCollection() bool {
	return false
}

func (p Int) String() string {
	return strconv.Itoa(int(p))
}

func (p Int) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Int) UUID() []byte {
	panic("An Int cannot have a UUID.")
	return nil
}

func (p Int) DBID() int64 {
	panic("An Int cannot have a DBID.")
	return 0
}

func (p Int) EnsureUUID() (theUUID []byte, err error) {
	panic("An Int cannot have a UUID.")
	return
}

func (p Int) UUIDuint64s() (id uint64, id2 uint64) {
	panic("An Int cannot have a UUID.")
	return
}

func (p Int) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("An Int cannot have a UUID.")
	return
}

func (p Int) UUIDstr() string {
	panic("An Int cannot have a UUID.")
	return ""
}

func (p Int) EnsureUUIDstr() (uuidstr string, err error) {
	panic("An Int cannot have a UUID.")
	return
}

func (p Int) UUIDabbrev() string {
	panic("An Int cannot have a UUID.")
	return ""
}

func (p Int) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("An Int cannot have a UUID.")
	return
}

func (p Int) RemoveUUID() {
	panic("An Int does not have a UUID.")
	return
}

func (p Int) Flags() int8 {
	panic("An Int has no Flags.")
	return 0
}

func (p Int) IsDirty() bool {
	return false
}
func (p Int) SetDirty() {
}
func (p Int) ClearDirty() {
}

func (p Int) IsIdReversed() bool {
	return false
}

func (p Int) SetIdReversed() {}

func (p Int) ClearIdReversed() {}

func (p Int) IsLoadNeeded() bool {
	return false
}

func (p Int) SetLoadNeeded()   {}
func (p Int) ClearLoadNeeded() {}

func (p Int) IsValid() bool { return true }
func (p Int) SetValid()     {}
func (p Int) ClearValid()   {}

func (p Int) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Int) SetStoredLocally()     {}
func (p Int) ClearStoredLocally()   {}

func (p Int) IsProxy() bool { return false }

func (p Int) IsTransient() bool { return false }

func (p Int) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

type Int32 int32

func (p Int32) IsZero() bool {
	return p == 0
}

func (p Int32) Type() *RType {
	return Int32Type
}

func (p Int32) This() RObject {
	return p
}

func (p Int32) IsUnit() bool {
	return true
}

func (p Int32) IsCollection() bool {
	return false
}

func (p Int32) String() string {
	return strconv.Itoa(int(p))
}

func (p Int32) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Int32) UUID() []byte {
	panic("An Int32 cannot have a UUID.")
	return nil
}

func (p Int32) DBID() int64 {
	panic("An Int32 cannot have a DBID.")
	return 0
}

func (p Int32) EnsureUUID() (theUUID []byte, err error) {
	panic("An Int32 cannot have a UUID.")
	return
}

func (p Int32) UUIDuint64s() (id uint64, id2 uint64) {
	panic("An Int32 cannot have a UUID.")
	return
}

func (p Int32) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("An Int32 cannot have a UUID.")
	return
}

func (p Int32) UUIDstr() string {
	panic("An Int32 cannot have a UUID.")
	return ""
}

func (p Int32) EnsureUUIDstr() (uuidstr string, err error) {
	panic("An Int32 cannot have a UUID.")
	return
}

func (p Int32) UUIDabbrev() string {
	panic("An Int32 cannot have a UUID.")
	return ""
}

func (p Int32) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("An Int32 cannot have a UUID.")
	return
}

func (p Int32) RemoveUUID() {
	panic("An Int32 does not have a UUID.")
	return
}

func (p Int32) Flags() int8 {
	panic("An Int32 has no Flags.")
	return 0
}

func (p Int32) IsDirty() bool {
	return false
}
func (p Int32) SetDirty() {
}
func (p Int32) ClearDirty() {
}

func (p Int32) IsIdReversed() bool {
	return false
}

func (p Int32) SetIdReversed() {}

func (p Int32) ClearIdReversed() {}

func (p Int32) IsLoadNeeded() bool {
	return false
}

func (p Int32) SetLoadNeeded()   {}
func (p Int32) ClearLoadNeeded() {}

func (p Int32) IsValid() bool { return true }
func (p Int32) SetValid()     {}
func (p Int32) ClearValid()   {}

func (p Int32) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Int32) SetStoredLocally()     {}
func (p Int32) ClearStoredLocally()   {}

func (p Int32) IsProxy() bool { return false }

func (p Int32) IsTransient() bool { return false }

func (p Int32) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

type Uint uint64

func (p Uint) IsZero() bool {
	return p == 0
}

func (p Uint) Type() *RType {
	return UintType
}

func (p Uint) This() RObject {
	return p
}

func (p Uint) IsUnit() bool {
	return true
}

func (p Uint) IsCollection() bool {
	return false
}

func (p Uint) String() string {
	return strconv.Itoa(int(p))
}

func (p Uint) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Uint) UUID() []byte {
	panic("A Uint cannot have a UUID.")
	return nil
}

func (p Uint) DBID() int64 {
	panic("A Uint cannot have a DBID.")
	return 0
}

func (p Uint) EnsureUUID() (theUUID []byte, err error) {
	panic("A Uint cannot have a UUID.")
	return
}

func (p Uint) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Uint cannot have a UUID.")
	return
}

func (p Uint) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Uint cannot have a UUID.")
	return
}

func (p Uint) UUIDstr() string {
	panic("A Uint cannot have a UUID.")
	return ""
}

func (p Uint) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Uint cannot have a UUID.")
	return
}

func (p Uint) UUIDabbrev() string {
	panic("A Uint cannot have a UUID.")
	return ""
}

func (p Uint) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Uint cannot have a UUID.")
	return
}

func (p Uint) RemoveUUID() {
	panic("A Uint does not have a UUID.")
	return
}

func (p Uint) Flags() int8 {
	panic("A Uint has no Flags.")
	return 0
}

func (p Uint) IsDirty() bool {
	return false
}
func (p Uint) SetDirty() {
}
func (p Uint) ClearDirty() {
}

func (p Uint) IsIdReversed() bool {
	return false
}

func (p Uint) SetIdReversed() {}

func (p Uint) ClearIdReversed() {}

func (p Uint) IsLoadNeeded() bool {
	return false
}

func (p Uint) SetLoadNeeded()   {}
func (p Uint) ClearLoadNeeded() {}

func (p Uint) IsValid() bool { return true }
func (p Uint) SetValid()     {}
func (p Uint) ClearValid()   {}

func (p Uint) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Uint) SetStoredLocally()     {}
func (p Uint) ClearStoredLocally()   {}

func (p Uint) IsProxy() bool { return false }

func (p Uint) IsTransient() bool { return false }

func (p Uint) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

type Uint32 uint32

func (p Uint32) IsZero() bool {
	return p == 0
}

func (p Uint32) Type() *RType {
	return Uint32Type
}

func (p Uint32) This() RObject {
	return p
}

func (p Uint32) IsUnit() bool {
	return true
}

func (p Uint32) IsCollection() bool {
	return false
}

func (p Uint32) String() string {
	return strconv.Itoa(int(p))
}

func (p Uint32) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Uint32) UUID() []byte {
	panic("A Uint32 cannot have a UUID.")
	return nil
}

func (p Uint32) DBID() int64 {
	panic("A Uint32 cannot have a DBID.")
	return 0
}

func (p Uint32) EnsureUUID() (theUUID []byte, err error) {
	panic("A Uint32 cannot have a UUID.")
	return
}

func (p Uint32) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Uint32 cannot have a UUID.")
	return
}

func (p Uint32) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Uint32 cannot have a UUID.")
	return
}

func (p Uint32) UUIDstr() string {
	panic("A Uint32 cannot have a UUID.")
	return ""
}

func (p Uint32) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Uint32 cannot have a UUID.")
	return
}

func (p Uint32) UUIDabbrev() string {
	panic("A Uint32 cannot have a UUID.")
	return ""
}

func (p Uint32) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Uint32 cannot have a UUID.")
	return
}

func (p Uint32) RemoveUUID() {
	panic("A Uint32 does not have a UUID.")
	return
}

func (p Uint32) Flags() int8 {
	panic("A Uint32 has no Flags.")
	return 0
}

func (p Uint32) IsDirty() bool {
	return false
}
func (p Uint32) SetDirty() {
}
func (p Uint32) ClearDirty() {
}

func (p Uint32) IsIdReversed() bool {
	return false
}

func (p Uint32) SetIdReversed() {}

func (p Uint32) ClearIdReversed() {}

func (p Uint32) IsLoadNeeded() bool {
	return false
}

func (p Uint32) SetLoadNeeded()   {}
func (p Uint32) ClearLoadNeeded() {}

func (p Uint32) IsValid() bool { return true }
func (p Uint32) SetValid()     {}
func (p Uint32) ClearValid()   {}

func (p Uint32) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Uint32) SetStoredLocally()     {}
func (p Uint32) ClearStoredLocally()   {}

func (p Uint32) IsProxy() bool { return false }

func (p Uint32) IsTransient() bool { return false }

func (p Uint32) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

type Float float64

func (p Float) IsZero() bool {
	return p == 0
}

func (p Float) Type() *RType {
	return FloatType
}

func (p Float) This() RObject {
	return p
}

func (p Float) IsUnit() bool {
	return true
}

func (p Float) IsCollection() bool {
	return false
}

func (p Float) String() string {
	return strconv.FormatFloat(float64(p), 'G', -1, 64)
}

func (p Float) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Float) UUID() []byte {
	panic("A Float cannot have a UUID.")
	return nil
}

func (p Float) DBID() int64 {
	panic("A Float cannot have a DBID.")
	return 0
}

func (p Float) EnsureUUID() (theUUID []byte, err error) {
	panic("A Float cannot have a UUID.")
	return
}

func (p Float) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Float cannot have a UUID.")
	return
}

func (p Float) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Float cannot have a UUID.")
	return
}

func (p Float) UUIDstr() string {
	panic("A Float cannot have a UUID.")
	return ""
}

func (p Float) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Float cannot have a UUID.")
	return
}

func (p Float) UUIDabbrev() string {
	panic("A Float cannot have a UUID.")
	return ""
}

func (p Float) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Float cannot have a UUID.")
	return
}

func (p Float) RemoveUUID() {
	panic("A Float does not have a UUID.")
	return
}

func (p Float) Flags() int8 {
	panic("A Float has no Flags.")
	return 0
}

func (p Float) IsDirty() bool {
	return false
}
func (p Float) SetDirty() {
}
func (p Float) ClearDirty() {
}

func (p Float) IsIdReversed() bool {
	return false
}

func (p Float) SetIdReversed() {}

func (p Float) ClearIdReversed() {}

func (p Float) IsLoadNeeded() bool {
	return false
}

func (p Float) SetLoadNeeded()   {}
func (p Float) ClearLoadNeeded() {}

func (p Float) IsValid() bool { return true }
func (p Float) SetValid()     {}
func (p Float) ClearValid()   {}

func (p Float) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Float) SetStoredLocally()     {}
func (p Float) ClearStoredLocally()   {}

func (p Float) IsProxy() bool { return false }

func (p Float) IsTransient() bool { return false }

func (p Float) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

type Bool bool

func (p Bool) IsZero() bool {
	return !bool(p)
}

func (p Bool) Type() *RType {
	return BoolType
}

func (p Bool) This() RObject {
	return p
}

func (p Bool) IsUnit() bool {
	return true
}

func (p Bool) IsCollection() bool {
	return false
}

func (p Bool) String() string {
	return strconv.FormatBool(bool(p))
}

func (p Bool) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Bool) UUID() []byte {
	panic("A Bool cannot have a UUID.")
	return nil
}

func (p Bool) DBID() int64 {
	panic("A Bool cannot have a DBID.")
	return 0
}

func (p Bool) EnsureUUID() (theUUID []byte, err error) {
	panic("A Bool cannot have a UUID.")
	return
}

func (p Bool) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Bool cannot have a UUID.")
	return
}

func (p Bool) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Bool cannot have a UUID.")
	return
}

func (p Bool) UUIDstr() string {
	panic("A Bool cannot have a UUID.")
	return ""
}

func (p Bool) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Bool cannot have a UUID.")
	return
}

func (p Bool) UUIDabbrev() string {
	panic("A Bool cannot have a UUID.")
	return ""
}

func (p Bool) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Bool cannot have a UUID.")
	return
}

func (p Bool) RemoveUUID() {
	panic("A Bool does not have a UUID.")
	return
}

func (p Bool) Flags() int8 {
	panic("A Bool has no Flags.")
	return 0
}

func (p Bool) IsDirty() bool {
	return false
}
func (p Bool) SetDirty() {
}
func (p Bool) ClearDirty() {
}

func (p Bool) IsIdReversed() bool {
	return false
}

func (p Bool) SetIdReversed() {}

func (p Bool) ClearIdReversed() {}

func (p Bool) IsLoadNeeded() bool {
	return false
}

func (p Bool) SetLoadNeeded()   {}
func (p Bool) ClearLoadNeeded() {}

func (p Bool) IsValid() bool { return true }
func (p Bool) SetValid()     {}
func (p Bool) ClearValid()   {}

func (p Bool) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Bool) SetStoredLocally()     {}
func (p Bool) ClearStoredLocally()   {}

func (p Bool) IsProxy() bool { return false }

func (p Bool) IsTransient() bool { return false }

func (p Bool) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}





type Nil byte

func (p Nil) IsZero() bool {
	return true
}

func (p Nil) Type() *RType {
	return NothingType
}

func (p Nil) This() RObject {
	return p
}

func (p Nil) IsUnit() bool {
	return true
}

func (p Nil) IsCollection() bool {
	return false
}

func (p Nil) String() string {
	return "*nil*"
}

func (p Nil) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Nil) UUID() []byte {
	panic("A Nil cannot have a UUID.")
	return nil
}

func (p Nil) DBID() int64 {
	panic("A Nil cannot have a DBID.")
	return 0
}

func (p Nil) EnsureUUID() (theUUID []byte, err error) {
	panic("A Nil cannot have a UUID.")
	return
}

func (p Nil) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Nil cannot have a UUID.")
	return
}

func (p Nil) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Nil cannot have a UUID.")
	return
}

func (p Nil) UUIDstr() string {
	panic("A Nil cannot have a UUID.")
	return ""
}

func (p Nil) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Nil cannot have a UUID.")
	return
}

func (p Nil) UUIDabbrev() string {
	panic("A Nil cannot have a UUID.")
	return ""
}

func (p Nil) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Nil cannot have a UUID.")
	return
}

func (p Nil) RemoveUUID() {
	panic("A Nil does not have a UUID.")
	return
}

func (p Nil) Flags() int8 {
	panic("A Nil has no Flags.")
	return 0
}

func (p Nil) IsDirty() bool {
	return false
}
func (p Nil) SetDirty() {
}
func (p Nil) ClearDirty() {
}

func (p Nil) IsIdReversed() bool {
	return false
}

func (p Nil) SetIdReversed() {}

func (p Nil) ClearIdReversed() {}

func (p Nil) IsLoadNeeded() bool {
	return false
}

func (p Nil) SetLoadNeeded()   {}
func (p Nil) ClearLoadNeeded() {}

func (p Nil) IsValid() bool { return true }
func (p Nil) SetValid()     {}
func (p Nil) ClearValid()   {}

func (p Nil) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Nil) SetStoredLocally()     {}
func (p Nil) ClearStoredLocally()   {}

func (p Nil) IsProxy() bool { return false }

func (p Nil) IsTransient() bool { return false }

func (p Nil) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

const NIL Nil = 0



type Float32 float32
type Byte byte
type Bit byte
type CodePoint uint32  // Should we call it Rune ?? Probably not. CodePoint is unambiguous, if boring.



type Complex complex128

func (p Complex) IsZero() bool {
	return p == 0
}

func (p Complex) Type() *RType {
	return ComplexType
}

func (p Complex) This() RObject {
	return p
}

func (p Complex) IsUnit() bool {
	return true
}

func (p Complex) IsCollection() bool {
	return false
}

func (p Complex) String() string {
    return fmt.Sprintf("%G",complex128(p))	
}

func (p Complex) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Complex) UUID() []byte {
	panic("A Complex cannot have a UUID.")
	return nil
}

func (p Complex) DBID() int64 {
	panic("A Complex cannot have a DBID.")
	return 0
}

func (p Complex) EnsureUUID() (theUUID []byte, err error) {
	panic("A Complex cannot have a UUID.")
	return
}

func (p Complex) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Complex cannot have a UUID.")
	return
}

func (p Complex) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Complex cannot have a UUID.")
	return
}

func (p Complex) UUIDstr() string {
	panic("A Complex cannot have a UUID.")
	return ""
}

func (p Complex) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Complex cannot have a UUID.")
	return
}

func (p Complex) UUIDabbrev() string {
	panic("A Complex cannot have a UUID.")
	return ""
}

func (p Complex) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Complex cannot have a UUID.")
	return
}

func (p Complex) RemoveUUID() {
	panic("A Complex does not have a UUID.")
	return
}

func (p Complex) Flags() int8 {
	panic("A Complex has no Flags.")
	return 0
}

func (p Complex) IsDirty() bool {
	return false
}
func (p Complex) SetDirty() {
}
func (p Complex) ClearDirty() {
}

func (p Complex) IsIdReversed() bool {
	return false
}

func (p Complex) SetIdReversed() {}

func (p Complex) ClearIdReversed() {}

func (p Complex) IsLoadNeeded() bool {
	return false
}

func (p Complex) SetLoadNeeded()   {}
func (p Complex) ClearLoadNeeded() {}

func (p Complex) IsValid() bool { return true }
func (p Complex) SetValid()     {}
func (p Complex) ClearValid()   {}

func (p Complex) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Complex) SetStoredLocally()     {}
func (p Complex) ClearStoredLocally()   {}

func (p Complex) IsProxy() bool { return false }

func (p Complex) IsTransient() bool { return false }

func (p Complex) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}



type Complex32 complex64

func (p Complex32) IsZero() bool {
	return p == 0
}

func (p Complex32) Type() *RType {
	return Complex32Type
}

func (p Complex32) This() RObject {
	return p
}

func (p Complex32) IsUnit() bool {
	return true
}

func (p Complex32) IsCollection() bool {
	return false
}

func (p Complex32) String() string {
    return fmt.Sprintf("%G",complex64(p))	
}

func (p Complex32) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Complex32) UUID() []byte {
	panic("A Complex32 cannot have a UUID.")
	return nil
}

func (p Complex32) DBID() int64 {
	panic("A Complex32 cannot have a DBID.")
	return 0
}

func (p Complex32) EnsureUUID() (theUUID []byte, err error) {
	panic("A Complex32 cannot have a UUID.")
	return
}

func (p Complex32) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Complex32 cannot have a UUID.")
	return
}

func (p Complex32) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Complex32 cannot have a UUID.")
	return
}

func (p Complex32) UUIDstr() string {
	panic("A Complex32 cannot have a UUID.")
	return ""
}

func (p Complex32) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Complex32 cannot have a UUID.")
	return
}

func (p Complex32) UUIDabbrev() string {
	panic("A Complex32 cannot have a UUID.")
	return ""
}

func (p Complex32) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Complex32 cannot have a UUID.")
	return
}

func (p Complex32) RemoveUUID() {
	panic("A Complex32 does not have a UUID.")
	return
}

func (p Complex32) Flags() int8 {
	panic("A Complex32 has no Flags.")
	return 0
}

func (p Complex32) IsDirty() bool {
	return false
}
func (p Complex32) SetDirty() {
}
func (p Complex32) ClearDirty() {
}

func (p Complex32) IsIdReversed() bool {
	return false
}

func (p Complex32) SetIdReversed() {}

func (p Complex32) ClearIdReversed() {}

func (p Complex32) IsLoadNeeded() bool {
	return false
}

func (p Complex32) SetLoadNeeded()   {}
func (p Complex32) ClearLoadNeeded() {}

func (p Complex32) IsValid() bool { return true }
func (p Complex32) SetValid()     {}
func (p Complex32) ClearValid()   {}

func (p Complex32) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Complex32) SetStoredLocally()     {}
func (p Complex32) ClearStoredLocally()   {}

func (p Complex32) IsProxy() bool { return false }

func (p Complex32) IsTransient() bool { return false }

func (p Complex32) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}



























type String string

func (p String) IsZero() bool {
	return p == ""
}

func (p String) Type() *RType {
	return StringType
}

func (p String) This() RObject {
	return p
}

func (p String) IsUnit() bool {
	return true
}

func (p String) IsCollection() bool {
	return false
}

func (p String) String() string {
	return string(p)
}

func (p String) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p String) UUID() []byte {
	panic("A String cannot have a UUID.")
	return nil
}

func (p String) DBID() int64 {
	panic("A String cannot have a DBID.")
	return 0
}

func (p String) EnsureUUID() (theUUID []byte, err error) {
	panic("A String cannot have a UUID.")
	return
}

func (p String) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A String cannot have a UUID.")
	return
}

func (p String) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A String cannot have a UUID.")
	return
}

func (p String) UUIDstr() string {
	panic("A String cannot have a UUID.")
	return ""
}

func (p String) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A String cannot have a UUID.")
	return
}

func (p String) UUIDabbrev() string {
	panic("A String cannot have a UUID.")
	return ""
}

func (p String) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A String cannot have a UUID.")
	return
}

func (p String) RemoveUUID() {
	panic("A String does not have a UUID.")
	return
}

func (p String) Flags() int8 {
	panic("A String has no Flags.")
	return 0
}

func (p String) IsDirty() bool {
	return false
}
func (p String) SetDirty() {
}
func (p String) ClearDirty() {
}

func (p String) IsIdReversed() bool {
	return false
}

func (p String) SetIdReversed() {}

func (p String) ClearIdReversed() {}

func (p String) IsLoadNeeded() bool {
	return false
}

func (p String) SetLoadNeeded()   {}
func (p String) ClearLoadNeeded() {}

func (p String) IsValid() bool { return true }
func (p String) SetValid()     {}
func (p String) ClearValid()   {}

func (p String) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p String) SetStoredLocally()     {}
func (p String) ClearStoredLocally()   {}

func (p String) IsProxy() bool { return false }

func (p String) IsTransient() bool { return false }

func (p String) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

/*
The Proxy type of RObject is a special relish-system-internal type.
Proxy RObjects are used as the members in an RObject collection which has been fetched from the database,
as a way of facilitating lazy loading of the actual objects in the collection.
The int64 value of the Proxy is the dbid of the real robject.

We might also use Proxy objects as the value of attributes at some times in the persistence lifecycle. Not sure when.

Oh also, we might need UUProxy objects whose underlying type might be complex128 and which represent the 
uuid of a remote object.

*/

type Proxy int64

/*
TODO This will require fetching the real object
*/
func (p Proxy) IsZero() bool {
	return false
}

func (p Proxy) Type() *RType {
	return ProxyType
}

func (p Proxy) This() RObject {
	return p
}

func (p Proxy) IsUnit() bool {
	return true
}

func (p Proxy) IsCollection() bool {
	return false
}

func (p Proxy) String() string {
	return strconv.Itoa(int(p))
}

func (p Proxy) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Proxy) UUID() []byte {
	panic("A Proxy cannot have a UUID.")
	return nil
}

func (p Proxy) DBID() int64 {
	return int64(p)
}

func (p Proxy) EnsureUUID() (theUUID []byte, err error) {
	panic("A Proxy cannot have a UUID.")
	return
}

func (p Proxy) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Proxy cannot have a UUID.")
	return
}

func (p Proxy) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Proxy cannot have a UUID.")
	return
}

func (p Proxy) UUIDstr() string {
	panic("A Proxy cannot have a UUID.")
	return ""
}

func (p Proxy) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Proxy cannot have a UUID.")
	return
}

func (p Proxy) UUIDabbrev() string {
	panic("A Proxy cannot have a UUID.")
	return ""
}

func (p Proxy) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Proxy cannot have a UUID.")
	return
}

func (p Proxy) RemoveUUID() {
	panic("A Proxy does not have a UUID.")
	return
}

func (p Proxy) Flags() int8 {
	panic("A Proxy has no Flags.")
	return 0
}

func (p Proxy) IsDirty() bool {
	return false
}
func (p Proxy) SetDirty() {
}
func (p Proxy) ClearDirty() {
}

func (p Proxy) IsIdReversed() bool {
	return false
}

func (p Proxy) SetIdReversed() {}

func (p Proxy) ClearIdReversed() {}

func (p Proxy) IsLoadNeeded() bool {
	return false
}

func (p Proxy) SetLoadNeeded()   {}
func (p Proxy) ClearLoadNeeded() {}

func (p Proxy) IsValid() bool { return true }
func (p Proxy) SetValid()     {}
func (p Proxy) ClearValid()   {}

func (p Proxy) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Proxy) SetStoredLocally()     {}
func (p Proxy) ClearStoredLocally()   {}

func (p Proxy) IsProxy() bool { return true }

func (p Proxy) IsTransient() bool { return false }

func (p Proxy) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

/*
Example code:

import (
       "fmt"
       "strings"
)

type RObject interface {
   This() RObject
   Boogah() string
   Exclaim() string
}

type rstring string

func (s rstring) This() RObject { return s }
func (s rstring) Boogah() string { return string(s) } 
func (s rstring) Exclaim() string { return fmt.Sprintf("My length is %v",len(s)) }
func main() {

	var r rstring = "Bagooh!"
	c := strings.Contains(string(r),"goo")
	fmt.Println(r)
	fmt.Println(c)
	var s string = "I am a string."
	r2 := rstring(s)
	fmt.Println(r2)
	fmt.Println(r2.This())
	fmt.Println(r2.This().Exclaim())	
}

*/
