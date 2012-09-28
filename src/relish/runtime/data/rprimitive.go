// Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.

package data

import ( 
	     "strconv"
         "errors"
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
var ProxyType *RType

var CallableType *RType
var MultiMethodType *RType
var MethodType *RType

var AnyType *RType
var NonPrimitiveType *RType
var StructType *RType
var CollectionType *RType

var InChannelType *RType
var OutChannelType *RType
var ChannelType *RType

func (rt *RuntimeEnv) createPrimitiveTypes() {

	PrimitiveType, _ = rt.CreateType("RelishPrimitive", []string{})
	NumericType, _ = rt.CreateType("Numeric", []string{"RelishPrimitive"})
	IntegerType, _ = rt.CreateType("Integer", []string{"Numeric"})
	IntType, _ = rt.CreateType("Int", []string{"Integer"})
	Int32Type, _ = rt.CreateType("Int32", []string{"Integer"})
	Int16Type, _ = rt.CreateType("Int16", []string{"Integer"})
	Int8Type, _ = rt.CreateType("Int8", []string{"Integer"})
	UintType, _ = rt.CreateType("Uint", []string{"Integer"})
	Uint32Type, _ = rt.CreateType("Uint32", []string{"Integer"})
	Uint16Type, _ = rt.CreateType("Uint16", []string{"Integer"})
	ByteType, _ = rt.CreateType("Byte", []string{"Integer"})
	BitType, _ = rt.CreateType("Bit", []string{"Integer"})
	BoolType, _ = rt.CreateType("Bool", []string{"RelishPrimitive"})
	TextType, _ = rt.CreateType("Text", []string{"RelishPrimitive"})
	CodePointType, _ = rt.CreateType("CodePoint", []string{"Text"})
	RealType, _ = rt.CreateType("Real", []string{"Numeric"})
	FloatType, _ = rt.CreateType("Float", []string{"Real"})
	Float32Type, _ = rt.CreateType("Float32", []string{"Real"})
	ComplexNumberType, _ = rt.CreateType("ComplexNumber", []string{"Numeric"})
	ComplexType, _ = rt.CreateType("Complex", []string{"ComplexNumber"})
	Complex32Type, _ = rt.CreateType("Complex32", []string{"ComplexNumber"})
	StringType, _ = rt.CreateType("String", []string{"Text"})
	ProxyType, _ = rt.CreateType("Proxy", []string{})

	CallableType, _ = rt.CreateType("Callable", []string{"Text"})
	MultiMethodType, _ = rt.CreateType("MultiMethod", []string{"Callable"})
	MethodType, _ = rt.CreateType("Method", []string{"Callable"})
	// Do I need a "Closure" type???

	AnyType, _ = rt.CreateType("Any", []string{})
	NonPrimitiveType, _ = rt.CreateType("NonPrimitive", []string{})
	StructType, _ = rt.CreateType("Struct", []string{})
	CollectionType, _ = rt.CreateType("Collection", []string{})
	
	// Do a need Iterable and InChannel <: Iterable, Collection <: Iterable
	
	InChannelType, _ = rt.CreateType("InChannel", []string{})
	OutChannelType, _ = rt.CreateType("OutChannel", []string{})
	ChannelType, _ = rt.CreateType("Channel", []string{"InChannel","OutChannel"})		
}


type Channel chan RObject

// TODO

func (p Channel) IsZero() bool {
	return p == nil
}

func (p Channel) Type() *RType {
	return ChannelType
}

func (p Channel) This() RObject {
	return p
}

func (p Channel) IsUnit() bool {
	return true
}

func (p Channel) IsCollection() bool {
	return false
}

func (p Channel) String() string {
	return "Channel"
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

type Float32 float32
type Byte byte
type Bit byte
type CodePoint uint32
type Complex complex128
type Complex32 complex64

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
