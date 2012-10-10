// Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.

package data

/*
   robject.go -  relish data objects - this source fie defines
                 generic object stuff + unitary (non-collection) objects 
*/

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
//	"os"
	"strconv"
	"io"
	"crypto/rand"
)

///////////////////////////////////////////////////////////////////////////
////////// DATA OBJECTS (Typed)
///////////////////////////////////////////////////////////////////////////

/*
Represents an instance of an object in the Relish data model.
Objects have a type
They can have attributes and participate in relations.
They may be a collection or a unit.
*/
type RObject interface {
	Type() *RType
	This() RObject // A reference to the most specific (outer wrapping) struct or the primitive 
	IsUnit() bool
	IsCollection() bool
	
	/*
	rivers1
	2502956568
	
	If this is a RCollection, return a slice or map of the elements, depending on the subtype of collection.
	If not a collection,an error is returned as the second return value.
    Can be used by go's template range action to iterate through collection in a template.	
	*/
	Iterable() (sliceOrMap interface{}, err error)

	String() string

	Flags() int8 // version of flags byte suitable for storing in the db RObject row

	/*
	   Whether this object has a uuid yet.
	*/
	HasUUID() bool

	/*
	   Return the UUID. Assumes the object has one. Returns a 16-byte slice.
	*/
	UUID() []byte

	/*
	   Create if needed and return the object's 16-byte UUID. Uses /dev/urandom
	   An error will be returned if object had no uuid and /dev/urandom cannot be read.
	*/
	EnsureUUID() (theUUID []byte, err error)

	/*
	   Return the object's uuid as two uint64s. Assumes object has a UUID.
	*/
	UUIDuint64s() (id uint64, id2 uint64)

	/*
	   Return the two uint64s representation of the UUID. Creates a UUID if object doesn't have one.
	*/
	EnsureUUIDuint64s() (id uint64, id2 uint64, err error)

	/*
	   Return the string representation of the UUID. Assumes object has a UUID.
	*/
	UUIDstr() string

	/*
	   Return the string representation of the UUID. Creates a UUID if object doesn't have one.
	*/
	EnsureUUIDstr() (uuidstr string, err error)

	/*
	   Return an abbreviated string representation of the UUID. Assumes object has a UUID.
	*/
	UUIDabbrev() string

	/*
	   Return an abbreviated string representation of the UUID. Creates a UUID if object doesn't have one.
	*/
	EnsureUUIDabbrev() (uuidstr string, err error)

	/*
	   Sets to nil the UUID value in the in-memory object instance. You should never call this. Only the
	   persistence subsystem should call this.
	*/
	RemoveUUID()

	/*
	   ID to be used as the object's record ID in the local SQLite database.
	*/
	DBID() int64

	/*
	   Whether the object is modified since last committed to db.
	*/
	IsDirty() bool
	SetDirty()
	ClearDirty()

	/*
	   Whether the object's uuid is reversed in the db object (i.e. bytes 8-15 of uuid are the db id).
	*/
	IsIdReversed() bool
	SetIdReversed()
	ClearIdReversed()

	IsLoadNeeded() bool
	SetLoadNeeded()
	ClearLoadNeeded()

	/*
	   The object meets its class invariant constraints and attribute and relation cardinality constraints.
	*/
	IsValid() bool
	SetValid()
	ClearValid()

	/*
	   The object is stored in the local sqlite database.
	*/
	IsStoredLocally() bool
	SetStoredLocally()
	ClearStoredLocally()

	/*
	   Whether this is a proxy storing only the dbid of the real object. If so, a
	   collection[i] = db.Fetch(collection[i]) must be used to get the real object into memory,
	   and fetch should use a runtime memcache to make sure we only have one copy of the object.
	*/
	IsProxy() bool

	/*
	   Is this object unstorable and un-streamable?	
	   Currently RMethods and RMultiMethods are transient.
	*/
	IsTransient() bool

    /*
    zero is defined more widely than just for numbers. This is similar to languages like
    Go and Python and Lisp.
    Empty lists, sets, and maps are zero.
    nil is zero.
    0 is zero
    0.0 is zero
    false is zero

    zero-ness is what is tested for in ifs, whiles, and by the and, or, not functions.
    */
	IsZero() bool
}

/*
   Abstract
*/
type robject struct {
	rtype *RType
	uuid  []byte // will be 16 bytes
	this  RObject
	flags byte

	// Do we need an idReversed bool attribute here?
}

// Could it be that dirty objects are just a list in the RuntimeEnv instead?

const FLAG_DIRTY = 0x1       // object's attributes or relations have been changed since last committed to db.
const FLAG_LOAD_NEEDED = 0x2 // object state must be reloaded from db because has uuid but never loaded 
// or because is dirty and transaction was aborted
const FLAG_VALID = 0x4 // Object has all valid attribute values and ok attribute and relationship cardinality
const FLAG_UNUSED_1 = 0x8
const FLAG_UNUSED_2 = 0x10
const FLAG_UNUSED_3 = 0x20
const FLAG_STORED_LOCALLY = 0x40
const FLAG_ID_REVERSED = 0x80 // The uuid is stored reversed in the db object (db id=bytes 8-15 of uuid).

/*
TODO We will have to see how to handle this. nil should be zero. But perhaps we should be able to 
define type-specific zero functions.
*/
func (o robject) IsZero() bool {
	return false
}

func (o robject) This() RObject {
	return o.this
}

func (o robject) Flags() int8 {
	return int8(o.flags)
}

func (o robject) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}



func (o robject) IsDirty() bool { return o.flags&FLAG_DIRTY != 0 }
func (o *robject) SetDirty()    { o.flags |= FLAG_DIRTY }
func (o *robject) ClearDirty()  { o.flags &^= FLAG_DIRTY }

func (o robject) IsLoadNeeded() bool { return o.flags&FLAG_LOAD_NEEDED != 0 }
func (o *robject) SetLoadNeeded()    { o.flags |= FLAG_LOAD_NEEDED }
func (o *robject) ClearLoadNeeded()  { o.flags &^= FLAG_LOAD_NEEDED }

func (o robject) IsValid() bool { return o.flags&FLAG_VALID != 0 }
func (o *robject) SetValid()    { o.flags |= FLAG_VALID }
func (o *robject) ClearValid()  { o.flags &^= FLAG_VALID }

func (o robject) IsIdReversed() bool { return o.flags&FLAG_ID_REVERSED != 0 }
func (o *robject) SetIdReversed()    { o.flags |= FLAG_ID_REVERSED }
func (o *robject) ClearIdReversed()  { o.flags &^= FLAG_ID_REVERSED }

func (o robject) IsStoredLocally() bool { return o.flags&FLAG_STORED_LOCALLY != 0 }
func (o *robject) SetStoredLocally()    { o.flags |= FLAG_STORED_LOCALLY }
func (o *robject) ClearStoredLocally()  { o.flags &^= FLAG_STORED_LOCALLY }

func (o *robject) IsProxy() bool { return false }

func (o *robject) IsTransient() bool { return false }

func (o *robject) String() string {
	var id string
	if o.HasUUID() {
		id = o.UUIDabbrev()
	} else {
		id = strconv.FormatUint(o.LocalID(), 10)
	}
	return fmt.Sprintf("%v:%v", o.Type(), id)
}

/*
   Returns the localId. Creates the local id if it does not yet exist.
   An object should not have both a local id and a UUID.
*/
func (o robject) LocalID() uint64 {
	localId, found := RT.objectIds[o.this]
	if !found {
		localId = RT.NextLocalID()
		RT.objectIds[o.this] = localId
	}
	return localId
}

/*
   Whether this object has a uuid yet.
*/
func (o robject) HasUUID() bool {
	return o.uuid != nil
}

/*
   Return the UUID. Assumes the object has one. Returns a 16-byte slice.
*/
func (o robject) UUID() []byte {
	return o.uuid
}


/*
   Create if needed and return the object's 16-byte UUID. Uses /dev/urandom
   An error will be returned if object had no uuid and /dev/urandom cannot be read.
*/
func (o *robject) EnsureUUID() (theUUID []byte, err error) {
	if o.uuid == nil {
//		var f *os.File
//		f, err = os.OpenFile("/dev/urandom", os.O_RDONLY, 0)
//		defer f.Close()
//		if err != nil {
//			return
//		}
		aUUID := make([]byte, 16)
//		var n int
		
		_,err = io.ReadFull(rand.Reader, aUUID)
		//n, err = f.Read(theUUID)
		if err != nil {
			return
		}
//		if n != 16 {
//			err = errors.New(fmt.Sprintf("When creating UUID could not read 16 random bytes from /dev/urandom. Read %v bytes.",n))
//			return
//		}
		o.uuid = aUUID
		delete(RT.objectIds, o.this) // Remove the local id from the runtime.
		return
	}
	theUUID = o.uuid
	return
}


/*
   Create if needed and return the object's 16-byte UUID. Uses /dev/urandom
   An error will be returned if object had no uuid and /dev/urandom cannot be read.

func (o *robject) EnsureUUID() (theUUID []byte, err error) {
	if o.uuid == nil {
		var f *os.File
		f, err = os.OpenFile("/dev/urandom", os.O_RDONLY, 0)
		defer f.Close()
		if err != nil {
			return
		}
		aUUID := make([]byte, 16)
		var n int
		n, err = f.Read(aUUID)
		if err != nil {
			return
		}
		if n != 16 {
			err = errors.New(fmt.Sprintf("When creating UUID could not read 16 random bytes from /dev/urandom. Read %v bytes.",n))
			return
		}
		o.uuid = aUUID
		delete(RT.objectIds, o.this) // Remove the local id from the runtime.
		return
	}
	theUUID = o.uuid
	return
}
*/

/*
   Return the object's uuid as two uint64s. Assumes object has a UUID.
*/
func (o robject) UUIDuint64s() (id uint64, id2 uint64) {
    if o.uuid == nil {
	    panic(fmt.Sprintf("Object %s uuid is nil !",&o))
    }
	buf := bytes.NewBuffer(o.UUID())
	err := binary.Read(buf, binary.BigEndian, &id)
	if err != nil {
		panic(err)
	}
	err = binary.Read(buf, binary.BigEndian, &id2)
	if err != nil {
		panic(err)
	}
	return
}

/*
   Return the two uint64s representation of the UUID. Creates a UUID if object doesn't have one.
*/
func (o *robject) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	_, err = o.EnsureUUID()
	if err != nil {
		return
	}
	id, id2 = o.UUIDuint64s()
	return
}

/*
   Return the string representation of the UUID. Assumes object has a UUID.
*/
func (o robject) UUIDstr() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", o.uuid[0:4], o.uuid[4:6], o.uuid[6:8], o.uuid[8:10], o.uuid[10:])
}

/*
   Return an abbreviation of the string representation of the UUID. Assumes object has a UUID.
*/
func (o robject) UUIDabbrev() string {
	return fmt.Sprintf("%x..%x", o.uuid[0:4], o.uuid[14:])
}

/*
   Return the string representation of the UUID. Creates a UUID if object doesn't have one.
*/
func (o *robject) EnsureUUIDstr() (uuidstr string, err error) {
	_, err = o.EnsureUUID()
	if err != nil {
		return
	}
	uuidstr = o.UUIDstr()
	return
}

/*
   Return the string representation of the UUID. Creates a UUID if object doesn't have one.
*/
func (o *robject) EnsureUUIDabbrev() (uuidstr string, err error) {
	_, err = o.EnsureUUID()
	if err != nil {
		return
	}
	uuidstr = o.UUIDabbrev()
	return
}

func (o *robject) RemoveUUID() {
	o.uuid = nil
}

func (o robject) DBID() int64 {
	id, id2 := o.UUIDuint64s()
	if o.IsIdReversed() {
		return int64(id2)
	}
	return int64(id)
}

func DBID(id1 int64, id2 int64, flags int) int64 {
	if flags&FLAG_ID_REVERSED != 0 {
		return id2
	}
	return id1
}

/*
   Used when summoning an object from the database. The object has already been
   created in memory. Now its contents have to be restored from the DB.
   This method accepts the object's id,id2,and flags from the database and
   restores this object's UUID and flags.
*/
func (o *robject) RestoreIdsAndFlags(id, id2 int64, flags int) {
	theUUID := make([]byte, 0, 16)
	o.flags = byte(flags)

	if o.IsIdReversed() {
		id, id2 = id2, id
	}
	uid := uint64(id)
	uid2 := uint64(id2)

	// write id, then id2 into the byte array.
	buf := bytes.NewBuffer(theUUID)

	err := binary.Write(buf, binary.BigEndian, uid)
	if err != nil {
		panic(err)
	}
	err = binary.Write(buf, binary.BigEndian, uid2)
	if err != nil {
		panic(err)
	}

	o.uuid = buf.Bytes()
	delete(RT.objectIds, o.this) // Remove the local id from the runtime. OOPS! Probably should never be a local id in this case!!!
}

/*
   Used when summoning an object from the database. The object has already been
   created in memory. Now its contents have to be restored from the DB.
   This method accepts the object's id,id2,and flags from the database and
   restores this object's UUID and flags.
*/
type Persistable interface {
	RestoreIdsAndFlags(id, id2 int64, flags int)
}

func (o robject) Type() *RType { return o.rtype }

/*
A unitary object. i.e. Not a collection.
*/
type runit struct {
	robject
}

func (o *runit) String() string {
	return (&(o.robject)).String()
}

func (u runit) IsUnit() bool {
	return true
}

func (u runit) IsCollection() bool {
	return false
}

/*
Create a new unitary object.
*/

func (rt *RuntimeEnv) NewObject(typeName string) (RObject, error) {
	
	// TODO Need to handle parameterized types here
	typ, found := rt.Types[typeName]
	if !found {
		return nil, fmt.Errorf("Type '%s' not found.", typeName)
	}
	
	// Check for special primitive types
	//
	if typ == ChannelType { // This will not be ChannelType, it will be a derived type <: Channel of T
		
		// TODO Need to handle parameterized types properly here.
		
		// For example, have to re-point the channel's type to a Channel of T type
		
		// TODO Need to handle capacity of channel argument here.
		
		channel := make(Channel)
		return channel, nil
	}
	
	// It's not a primitive type
	
	unit := &runit{robject{rtype: typ}}
	unit.robject.this = unit
	return unit, nil
}
