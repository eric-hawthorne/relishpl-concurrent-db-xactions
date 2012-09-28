// Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.

package data

/*
   rcollection.go - relish collection objects
*/

import (
	"fmt"
	"sort"
)

const MAX_CARDINALITY = 999999999999999999 // Replace with highest int64?

///////////////////////////////////////////////////////////////////////////
////////// COLLECTIONS
///////////////////////////////////////////////////////////////////////////

/*
[] Widget             List

[<] Widget            Sorted list using natural order of Widgets (which must be defined)

[<attr] Widget        Sorted list using attribute/unary function of Widget

[<less] Widget        Sorted ist using binary comparison function over widgets

{} Widget             Set

{<} Widget            Sorted set using natural order of Widgets (which must be defined)

{<attr} Widget        Sorted set using attribute/unary function of Widget

{<less} Widget        Sorted set using binary comparison function over widgets (if "less" unary func defined it will be used instead)


{} String=>Widget     Map

{} name=>Widget       Map using name attribute of Widget (but does not update itself if widget.name is changed) 

{<} String=>Widget    Sorted map using natural order of Strings (which must be defined)

{<} name=>Widget      Sorted map using name attribute/unary function of Widget (but does not update itself if widget.name is changed) 

*/

type RCollection interface {
	RObject
	ElementType() *RType
	Length() int64
	Cap() int64
	MinCard() int64
	MaxCard() int64
	IsMap() bool
	IsSet() bool
	IsList() bool
	IsOrdered() bool // true if either the collection is maintained in a sorted order, or if the collection
	// at least holds each member in an index-accessible position and returns members
	// when iterated over in order of their sequential indexed position.

	IsSorting() bool // true if the collection has a defined sort order other index-position.
	IsCardOk() bool  // Is my current cardinality within my cardinality constraints?
	Owner() RObject  // if non-nil, this collection is the implementation of a multi-valued attribute.
	// Returns an iterator that yields the objects in the collection. A Map iterator returns the keys.
	Iter() <-chan RObject // Usage: for obj := range s.Iter()  or ch := s.Iter(); x, ok = <-ch

}

/*
A collection which can have a member added. It is added at the end (appended) if this is an non-sorting list.
It is added in the appropriate place in the order, if this is a sorting collection.
It is added in undetermined place if an unordered set.
*/
type AddableCollection interface {
	Add(obj RObject, context MethodEvaluationContext) (added bool, newLen int)

	/*
		This version of the add method does not sort. It assumes that it is being called with element objects
		already known to be simply insertable (at the end of if applicable) the collection.
		Used by the persistence service. Do not use for general use of the collection.
	*/
	AddSimple(obj RObject) (newLen int)
}

/*
A collection which can have a member removed.
*/
type RemovableCollection interface {
	/*
	   removedIndex will be -1 if not applicable or if removed is false
	*/
	Remove(obj RObject) (removed bool, removedIndex int)

	/*
		Removes all members of the in-memory aspect of the collection, setting its len to 0. 
		Does not affect the database-persisted aspect of the collection. 
		Used to refresh the collection from the db.
	*/
	ClearInMemory()
}

type OrderedCollection interface {
	Index(obj RObject, start int) int
}

/*
A collection which can return its go list implementation
*/
type List interface {
	RCollection
	AddableCollection
	RemovableCollection
	OrderedCollection
	At(i int) RObject
	Vector() *RVector
	AsSlice() []RObject
}

/*
A collection which can return its go map implementation
*/
type Set interface {
	AddableCollection
	RemovableCollection
	BoolMap() map[RObject]bool
}

type Map interface {
	RCollection
	Get(key RObject) (val RObject, found bool)
}

/*
func (c *container) Iter () <-chan item {
    ch := make(chan item);
    go func () {
        for i := 0; i < c.size; i++ {
            ch <- c.items[i]
        }
    } ();
    return ch
}

*/

/*
   Abstract 
*/
type rcollection struct {
	robject
	minCard     int64
	maxCard     int64
	elementType *RType
	owner       RObject

	sortWith *sortOp // Which attribute of a member, or which unary func of member or which less function to sort with. May be nil.
}

/*
Only one of the attr or unaryFunction will be non-nil.
If attr or unaryFunction is non-nil, then lessFunction must be the "lt" multiMethod.

collection.sortWith.lessFunction,_ := RT.MultiMethods["lt"]

If attr and unaryFunction are nil, lessFunction may be any binary boolean function which has a method whose
parameter signature is compatible with a pair of values of the elementType of the collection. lessFunction MAY
be the "lt" function in this case but need not be. The function is treated as a "less-than" predicate.

*/
type sortOp struct {
	attr          *AttributeSpec
	unaryFunction *RMultiMethod
	lessFunction  *RMultiMethod
	descending    bool
}

func (o *rcollection) String() string {
	return (&(o.robject)).String()
}

func (c rcollection) IsUnit() bool {
	return false
}

func (c rcollection) IsCollection() bool {
	return true
}

func (c rcollection) MinCard() int64 {
	return c.minCard
}

func (c rcollection) MaxCard() int64 {
	return c.maxCard
}

func (c rcollection) ElementType() *RType {
	return c.elementType
}

/*
   If not nil, it means this collection is the implemnentation of a multiple-valued attribute.
*/
func (c rcollection) Owner() RObject {
	return c.owner
}

/*
A set of relish objects constrained to be of some type.
Implements RCollection
Object address defines element equality. May want to fix that!!! It may not even be true.
*/
type rset struct {
	rcollection
	m map[RObject]bool // use this as set 
}

func (s *rset) BoolMap() map[RObject]bool {
	return s.m
}

func (s *rset) Add(obj RObject, context MethodEvaluationContext) (added bool, newLen int) {
	if s.m == nil {
		s.m = make(map[RObject]bool)
	}
	added = !s.m[obj]
	s.m[obj] = true
	newLen = len(s.m)
	return
}

func (s *rset) AddSimple(obj RObject) (newLen int) {
	if s.m == nil {
		s.m = make(map[RObject]bool)
	}
	s.m[obj] = true
	newLen = len(s.m)
	return
}

func (s *rset) Remove(obj RObject) (removed bool, removedIndex int) {
	removed = s.m[obj]
	delete(s.m, obj) // delete(s.m,obj)
	removedIndex = -1
	return
}

func (s *rset) ClearInMemory() {
	s.m = nil
}

/*
Weird behaviour: Only if the iteration is allowed to complete (i.e. to exhaust the map) will proxies
in the map be replaced by real objects.

TODO MAPS AND PROXIES ARE NOT HAPPY TOGETHER YET!!!!
IT WOULD ONLY WORK IF THE ROBJECT IDENTITY IN THE MAP IS BASED ON THE DBID instead of object address.
TODO
*/
func (c *rset) Iter() <-chan RObject {
	ch := make(chan RObject)
	go func() {
		var fromPersistence map[RObject]RObject
		for robj, _ := range c.m {
			if robj.IsProxy() {
				var err error
				proxy := robj.(Proxy)
				robj, err = RT.DB().Fetch(int64(proxy), 0)
				if err != nil {
					panic(fmt.Sprintf("Error fetching set element: %s", err))
				}

				if fromPersistence == nil {
					fromPersistence = make(map[RObject]RObject)
				}
				fromPersistence[proxy] = robj

			}
			ch <- robj
		}

		// Replace proxies in the set with real objects.

		// TODO Need to mutex lock the map here to guarantee the len of the map is always correct.
		for proxy, robj := range fromPersistence {
			c.m[robj] = true
			delete(c.m, proxy) // delete(c,m)				
		}

		close(ch)
	}()
	return ch
}

/*
Creates a fresh new slice.
*/
func (c *rset) AsSlice() []RObject {
    s := make([]RObject,0,c.Length())     
	for obj := range c.Iter() {
        s = append(s,obj)
	}
    return s
}

/*
*/
func (s *rset) Iterable() (interface{},error) {
	return s.AsSlice(),nil
}


/*
   Constructor
*/
func (rt *RuntimeEnv) Newrset(elementType *RType, minCardinality, maxCardinality int64, owner RObject) (coll RCollection, err error) {
	typ, err := rt.getSetType(elementType)
	if err != nil {
		return nil, err
	}
	if maxCardinality == -1 {
		maxCardinality = MAX_CARDINALITY
	}
	coll = &rset{rcollection{robject{rtype: typ}, minCardinality, maxCardinality, elementType, owner, nil}, nil}
	return
}

func (s rset) Length() int64   { return int64(len(s.m)) }
func (s rset) Cap() int64      { return int64(len(s.m)) * 2 }
func (s rset) IsMap() bool     { return false }
func (s rset) IsSet() bool     { return true }
func (s rset) IsList() bool    { return false }
func (s rset) IsOrdered() bool { return false } // This may change!! Actually, need an rsortedset type
func (s rset) IsSorting() bool { return false } // This may change!! Depends on presence of an ordering function in rsortedset
func (s rset) IsCardOk() bool  { return s.Length() >= s.MinCard() && s.Length() <= s.MaxCard() }

/*
A sorted set of relish objects constrained to be of some type.
Implements RCollection
Object address defines element equality. May want to fix that!!! It may not even be true.
*/
type rsortedset struct {
	rcollection
	m map[RObject]bool // use this as set 
	v *RVector
}

func (s *rsortedset) BoolMap() map[RObject]bool {
	return s.m
}

func (s *rsortedset) Add(obj RObject, context MethodEvaluationContext) (added bool, newLen int) {
	if s.m == nil {
		s.m = make(map[RObject]bool)
		s.v = new(RVector)
	}
	_, found := s.m[obj]
	if !found {
		added = true
		s.m[obj] = true
		s.v.Push(obj)

		RT.SetEvalContext(s, context)
		defer RT.UnsetEvalContext(s)
		sort.Sort(s)
	}
	newLen = len(s.m)
	return
}

func (s *rsortedset) AddSimple(obj RObject) (newLen int) {
	if s.m == nil {
		s.m = make(map[RObject]bool)
		s.v = new(RVector)
	}

	s.m[obj] = true
	s.v.Push(obj)
	newLen = len(s.m)
	return
}

func (s *rsortedset) At(i int) RObject {
	obj := s.v.At(i).(RObject)
	if obj.IsProxy() {
		var err error
		proxy := obj.(Proxy)
		obj, err = RT.DB().Fetch(int64(proxy), 0)
		if err != nil {
			panic(fmt.Sprintf("Error fetching sorted-set element [%v]: %s", i, err))
		}
	}
	return obj
}

/*
{<} Widget            Sorted set using natural order of Widgets (which must be defined)

{<attr} Widget        Sorted set using attribute/unary function of Widget

{<less} Widget        Sorted set using binary comparison function over widgets (if "less" unary func defined it will be used instead)


type Interface interface {
    // Len is the number of elements in the collection.
    Len() int
    // Less returns whether the element with index i should sort
    // before the element with index j.
    Less(i, j int) bool
    // Swap swaps the elements with indexes i and j.
    Swap(i, j int)
}
*/

func (s *rsortedset) Len() int {
	return len(s.m)
}

func (s *rsortedset) Less(i, j int) bool {
	// TODO !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!	

	return false

}

/*
type sortOp {
	attr *AttributeSpec
	unaryFunction *RMultiMethod
	lessFunction *RMultiMethod
	descending bool
}
*/

/*
Not valid to call on indexes >= the length of the collection.
*/
func (s *rsortedset) Swap(i, j int) {
	s.v.Swap(i, j)
}

/*
Returns the index of the first-found occurrence of the argument object with the search beginning at the start index.
TODO Should make this more efficient by doing a binary search. !!!!!!!

*/
func (s *rsortedset) Index(obj RObject, start int) int {
	if s.m != nil {
		ln := len(*(s.v))
		for i := start; i < ln; i++ {
			if obj == s.v.At(i) {
				return i
			}
		}
	}
	return -1
}

func (s *rsortedset) Remove(obj RObject) (removed bool, removedIndex int) {
	if s.v == nil {
		removedIndex = -1
	} else {
		delete(s.m, obj) // delete (s.m,obj)	
		removedIndex = s.Index(obj, 0)
		if removedIndex >= 0 {
			s.v.Delete(removedIndex)
			removed = true
		}
	}
	return
}

func (s *rsortedset) ClearInMemory() {
	s.m = nil
	if s.v != nil {
		s.v = s.v.Resize(0, s.v.Cap())
	}
}

/*
 */
func (c *rsortedset) Iter() <-chan RObject {
	ch := make(chan RObject)
	go func() {
		if c.v != nil {
			for _, obj := range *(c.v) {
				robj := obj.(RObject)
				if robj.IsProxy() {
					var err error
					proxy := robj.(Proxy)
					robj, err = RT.DB().Fetch(int64(proxy), 0)
					if err != nil {
						panic(fmt.Sprintf("Error fetching list element: %s", err))
					}
				}
				ch <- robj
			}
		}
		close(ch)
	}()
	return ch
}

/*
Do not modify the returned slice
*/
func (s *rsortedset) AsSlice() []RObject {
	return []RObject(*(s.v))
}

/*
Do not modify the returned slice
*/
func (s *rsortedset) Iterable() (interface{},error) {
	return s.AsSlice(),nil
}


/*
   Constructor
*/
func (rt *RuntimeEnv) Newrsortedset(elementType *RType, minCardinality, maxCardinality int64, owner RObject, sortWith *sortOp) (coll RCollection, err error) {
	typ, err := rt.getSetType(elementType)
	if err != nil {
		return nil, err
	}
	if maxCardinality == -1 {
		maxCardinality = MAX_CARDINALITY
	}
	coll = &rsortedset{rcollection{robject{rtype: typ}, minCardinality, maxCardinality, elementType, owner, sortWith}, nil, nil}
	return
}

func (s rsortedset) Length() int64   { return int64(len(s.m)) }
func (s rsortedset) Cap() int64      { return int64(len(s.m)) * 2 }
func (s rsortedset) IsMap() bool     { return false }
func (s rsortedset) IsSet() bool     { return true }
func (s rsortedset) IsList() bool    { return false }
func (s rsortedset) IsOrdered() bool { return true } // This may change!! Actually, need an rsortedset type
func (s rsortedset) IsSorting() bool { return true } // This may change!! Depends on presence of an ordering function in rsortedset
func (s rsortedset) IsCardOk() bool  { return s.Length() >= s.MinCard() && s.Length() <= s.MaxCard() }

/*
A list of relish objects constrained to be of some type.
Implements RCollection
*/
type rlist struct {
	rcollection
	v *RVector
}

func (c *rlist) Iter() <-chan RObject {
	ch := make(chan RObject)
	go func() {
		if c.v != nil {
			for _, obj := range *(c.v) {
				robj := obj.(RObject)
				if robj.IsProxy() {
					var err error
					proxy := robj.(Proxy)
					robj, err = RT.DB().Fetch(int64(proxy), 0)
					if err != nil {
						panic(fmt.Sprintf("Error fetching list element: %s", err))
					}
				}
				ch <- robj
			}
		}
		close(ch)
	}()
	return ch
}

/*
Return the underlying collection.
*/
func (s *rlist) Vector() *RVector {
	return s.v
}

func (s *rlist) AsSlice() []RObject {
	return []RObject(*(s.v))
}

func (s *rlist) Iterable() (interface{},error) {
	return s.AsSlice(),nil
}

// RT.SetEvalContext(obj, context)
// defer RT.UnsetEvalContext(obj)
// context := RT.GetEvalContext(obj)	

func (s *rlist) Add(obj RObject, context MethodEvaluationContext) (added bool, newLen int) {
	if s.v == nil {
		s.v = new(RVector)
	}
	s.v.Push(obj)

	if s.IsSorting() {
		RT.SetEvalContext(s, context)
		defer RT.UnsetEvalContext(s)
		sort.Sort(s)
	}
	added = true
	newLen = s.v.Len()
	return
}

func (s *rlist) AddSimple(obj RObject) (newLen int) {
	if s.v == nil {
		s.v = new(RVector)
	}
	s.v.Push(obj)
	newLen = s.v.Len()
	return
}

func (s *rlist) At(i int) RObject {
	obj := s.v.At(i).(RObject)
	if obj.IsProxy() {
		var err error
		proxy := obj.(Proxy)
		obj, err = RT.DB().Fetch(int64(proxy), 0)
		if err != nil {
			panic(fmt.Sprintf("Error fetching list element [%v]: %s", i, err))
		}
	}
	return obj
}

func (s *rlist) Len() int {
	return s.v.Len()
}

func (s *rlist) Less(i, j int) bool {
	// TODO !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!	

	if s.sortWith == nil { // Not a sorted list. So sorting is an (expensive) no-op.
		return i < j
	}

	var evalContext MethodEvaluationContext
	//var isLess RObject

	if s.sortWith.attr != nil {

		// Get attr value of both list members

		obj1 := s.At(i)
		val1, found := RT.AttrVal(obj1, s.sortWith.attr)
		if !found {
			panic(fmt.Sprintf("Object %v has no value for attribute %s", obj1, s.sortWith.attr.Part.Name))
		}

		obj2 := s.At(j)
		val2, found := RT.AttrVal(obj2, s.sortWith.attr)
		if !found {
			panic(fmt.Sprintf("Object %v has no value for attribute %s", obj2, s.sortWith.attr.Part.Name))
		}

		// Use the "less" multimethod to compare them.

		evalContext = RT.GetEvalContext(s)

		// Assumes that the sortWith has been given the "less" multimethod. TODO!
		// 
		isLess := evalContext.EvalMultiMethodCall(s.sortWith.lessFunction, []RObject{val1, val2})

		if s.sortWith.descending {
			return isLess.IsZero()
		}
		return !isLess.IsZero()

	} else if s.sortWith.unaryFunction != nil {

		// Evaluate the unary function separately on both list members

		evalContext = RT.GetEvalContext(s)

		obj1 := s.At(i)
		val1 := evalContext.EvalMultiMethodCall(s.sortWith.unaryFunction, []RObject{obj1})

		obj2 := s.At(j)
		val2 := evalContext.EvalMultiMethodCall(s.sortWith.unaryFunction, []RObject{obj2})

		// Use the "less" multimethod to compare them.

		evalContext := RT.GetEvalContext(s)

		// Assumes that the sortWith has been given the "less" multimethod. TODO!
		//
		isLess := evalContext.EvalMultiMethodCall(s.sortWith.lessFunction, []RObject{val1, val2})

		if s.sortWith.descending {
			return isLess.IsZero()
		}
		return !isLess.IsZero()

		// Use the inbuilt "less" multimethod to compare the function return values.

	}
	// else ... lessFunction

	// Apply the multi-method to the two list members. It may be just the "less" multimethod.

	// Get attr value of both list members

	obj1 := s.At(i)

	obj2 := s.At(j)

	// Use the multimethod to compare them.

	evalContext = RT.GetEvalContext(s)

	isLess := evalContext.EvalMultiMethodCall(s.sortWith.lessFunction, []RObject{obj1, obj2})

	if s.sortWith.descending {
		return isLess.IsZero()
	}
	return !isLess.IsZero()

}

/*
type sortOp {
	attr *AttributeSpec
	unaryFunction *RMultiMethod
	lessFunction *RMultiMethod
	descending bool
}
*/

/*
Not valid to call on indexes >= the length of the collection.
*/
func (s *rlist) Swap(i, j int) {
	s.v.Swap(i, j)
}

/*
Returns the index of the first-found occurrence of the argument object with the search beginning at the start index.
*/
func (s *rlist) Index(obj RObject, start int) int {
	if s.v != nil {
		ln := len(*(s.v))
		for i := start; i < ln; i++ {
			if obj == s.v.At(i) {
				return i
			}
		}
	}
	return -1
}

func (s *rlist) Remove(obj RObject) (removed bool, removedIndex int) {
	if s.v == nil {
		removedIndex = -1
	} else {
		removedIndex = s.Index(obj, 0)
		if removedIndex >= 0 {
			s.v.Delete(removedIndex)
			removed = true
		}
	}
	return
}

func (s *rlist) ClearInMemory() {
	if s.v != nil {
		s.v = s.v.Resize(0, s.v.Cap())
	}
}

/*
   Constructor
*/
func (rt *RuntimeEnv) Newrlist(elementType *RType, minCardinality, maxCardinality int64, owner RObject, sortWith *sortOp) (coll RCollection, err error) {
	typ, err := rt.getListType(elementType)
	if err != nil {
		return nil, err
	}
	if maxCardinality == -1 {
		maxCardinality = MAX_CARDINALITY
	}
	coll = &rlist{rcollection{robject{rtype: typ}, minCardinality, maxCardinality, elementType, owner, sortWith}, nil}
	return
}

func (s rlist) Length() int64   { return int64(s.v.Len()) }
func (s rlist) Cap() int64      { return int64(s.v.Cap()) }
func (s rlist) IsMap() bool     { return false }
func (s rlist) IsSet() bool     { return false }
func (s rlist) IsList() bool    { return true }
func (s rlist) IsOrdered() bool { return true } // This may change!! Depends on presence of an ordering function
func (s rlist) IsSorting() bool { return s.sortWith != nil }
func (s rlist) IsCardOk() bool  { return s.Length() >= s.MinCard() && s.Length() <= s.MaxCard() }

type rstringmap struct {
	rcollection
	m map[string]RObject
}

func (c *rstringmap) Iter() <-chan RObject {
	ch := make(chan RObject)
	go func() {
		for key, _ := range c.m {
			ch <- String(key)
		}
		close(ch)
	}()
	return ch
}

func (s *rstringmap) Iterable() (interface{},error) {
	return s.m,nil
}


func (s rstringmap) Length() int64   { return int64(len(s.m)) }
func (s rstringmap) Cap() int64      { return int64(len(s.m)) * 2 }
func (s rstringmap) IsMap() bool     { return true }
func (s rstringmap) IsSet() bool     { return false }
func (s rstringmap) IsList() bool    { return false }
func (s rstringmap) IsOrdered() bool { return false } // This may change!! Need an rorderedstringmap
func (s rstringmap) IsSorting() bool { return false } // This may change!! Depends on presence of an ordering function in rorderedstringmap
func (s rstringmap) IsCardOk() bool  { return s.Length() >= s.MinCard() && s.Length() <= s.MaxCard() }

func (s *rstringmap) Get(key RObject) (val RObject, found bool) {
	k := string(key.(String))
	val, found = s.m[k]
	return
}

type ruint64map struct {
	rcollection
	m map[uint64]RObject
}

func (c *ruint64map) Iter() <-chan RObject {
	ch := make(chan RObject)
	go func() {
		for key, _ := range c.m {
			ch <- Uint(key)
		}
		close(ch)
	}()
	return ch
}

func (s *ruint64map) Iterable() (interface{},error) {
	return s.m,nil
}

func (s ruint64map) Length() int64   { return int64(len(s.m)) }
func (s ruint64map) Cap() int64      { return int64(len(s.m)) * 2 }
func (s ruint64map) IsMap() bool     { return true }
func (s ruint64map) IsSet() bool     { return false }
func (s ruint64map) IsList() bool    { return false }
func (s ruint64map) IsOrdered() bool { return false } // This may change!! Need an rordereduintmap
func (s ruint64map) IsSorting() bool { return false } // This may change!! Depends on presence of an ordering function in rorderedstringmap
func (s ruint64map) IsCardOk() bool  { return s.Length() >= s.MinCard() && s.Length() <= s.MaxCard() }

func (s *ruint64map) Get(key RObject) (val RObject, found bool) {
	k := uint64(key.(Uint))
	val, found = s.m[k]
	return
}

/*
Can I just use a ruint64map and cast the int64 to uint64? probably
*/
type rint64map struct {
	rcollection
	m map[int64]RObject
}

func (c *rint64map) Iter() <-chan RObject {
	ch := make(chan RObject)
	go func() {
		for key, _ := range c.m {
			ch <- Int(key)
		}
		close(ch)
	}()
	return ch
}

func (s *rint64map) Iterable() (interface{},error) {
	return s.m,nil
}

func (s rint64map) Length() int64   { return int64(len(s.m)) }
func (s rint64map) Cap() int64      { return int64(len(s.m)) * 2 }
func (s rint64map) IsMap() bool     { return true }
func (s rint64map) IsSet() bool     { return false }
func (s rint64map) IsList() bool    { return false }
func (s rint64map) IsOrdered() bool { return false } // This may change!! Depends on presence of an ordering function
func (s rint64map) IsSorting() bool { return false } // This may change!! Depends on presence of an ordering function in rorderedstringmap
func (s rint64map) IsCardOk() bool  { return s.Length() >= s.MinCard() && s.Length() <= s.MaxCard() }

func (s *rint64map) Get(key RObject) (val RObject, found bool) {
	k := int64(key.(Int))
	val, found = s.m[k]
	return
}

type rpointermap struct {
	rcollection
	m map[*runit]RObject
}

func (c *rpointermap) Iter() <-chan RObject {
	ch := make(chan RObject)
	go func() {
		for key, _ := range c.m {
			ch <- key.This()
		}
		close(ch)
	}()
	return ch
}

func (s *rpointermap) Iterable() (interface{},error) {
	return s.m,nil
}

func (s rpointermap) Length() int64   { return int64(len(s.m)) }
func (s rpointermap) Cap() int64      { return int64(len(s.m)) * 2 }
func (s rpointermap) IsMap() bool     { return true }
func (s rpointermap) IsSet() bool     { return false }
func (s rpointermap) IsList() bool    { return false }
func (s rpointermap) IsOrdered() bool { return false } // This may change!! Depends on presence of an ordering function
func (s rpointermap) IsSorting() bool { return false } // This may change!! Depends on presence of an ordering function in rorderedstringmap
func (s rpointermap) IsCardOk() bool  { return s.Length() >= s.MinCard() && s.Length() <= s.MaxCard() }

func (s *rpointermap) Get(key RObject) (val RObject, found bool) {
	unit := key.(*runit)
	val, found = s.m[unit]
	return
}
