// Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.

package data

/*
   code-go - methods and expressions    (methods are data too)
*/

import (
	"fmt"
	"errors"
	"relish/compiler/ast"
)

/*
   Logical method - i.e. a family of methods that all have the same purpose
   and are selected depending on the types of the arguments.

   data:

   methods - for each parameter arity, all method implementations of that arity 
             of this multi-method.
   cachedMethods - for each typetuple representing actual-argument type tuples
   that the multi-method has been called on so far in the process execution,
   the most compatible most specific argument-signature method implementation
   for the type-tuple. Added to as more calls of the multi-method (on different
   types of arguments) are made.

NOTE!!!!!! METHODS should be a map by arity to
a map by pointer to first parameter type to a list of methods.

Don't use this for zero arity of course.

The zero arity method should be findable from a special type tuple with a zero-length list of types.

*/
type RMultiMethod struct {
	Name          string
	Methods       map[int][]*RMethod       // methods of each arity which implement this multimethod
	CachedMethods map[*RTypeTuple]*RMethod //
	NumReturnArgs int
	MaxArity int	
	Pkg            *RPackage  // the package that this multimethod is owned by
	                          // Note that a method may be referenced by multiple packages' multimethods
	                          // the method itself points to one of these (does not matter which - it just gets its name from the mm)	
}

/*
   Constructor of a multi-method. Sets its name and makes its maps.
*/
func newRMultiMethod(name string, numReturnArgs int, pkg *RPackage) *RMultiMethod {
	return &RMultiMethod{Name: name, Methods: make(map[int][]*RMethod), CachedMethods: make(map[*RTypeTuple]*RMethod), NumReturnArgs: numReturnArgs}
}

/*
   Clone this multimethod to make an initially identical one but owned by the argument package.
*/ 
func (p * RMultiMethod) Clone(pkg *RPackage) *RMultiMethod {
	mm := newRMultiMethod(p.Name,p.NumReturnArgs,pkg)
	
    for arity,methodList := range p.Methods {
	   var ms []*RMethod
	   mm.Methods[arity] = append(ms,methodList...)
    }	
	
	for tt,method := range p.CachedMethods {
		mm.CachedMethods[tt] = method
	}
	mm.MaxArity = p.MaxArity
	return mm
}

/*
For methods which are not found in p but are found in q, add them to p
*/
func (p * RMultiMethod) MergeInNewMethodsFrom(q *RMultiMethod) {
	for arity,pMethods := range p.Methods {
		qMethods, found := q.Methods[arity]
		if found {
			var methodsToAdd []*RMethod
			for _, mq := range qMethods {
				inP := false
				for _,mp := range pMethods {
					if mq == mp {
						inP = true
						break
					}
				} 
				if ! inP {
					methodsToAdd = append(methodsToAdd,mq)
				}
			}
			p.Methods[arity] = append(pMethods, methodsToAdd...)
		}
    }

    // Now clear the cached methods map of p, since we have new methods to consider.
    // The cache should not have anything in it anyway, since should be calling this
    // at initial code loading time before any relish execution takes place.

    if len(p.CachedMethods) > 0 {
	   p.CachedMethods = make(map[*RTypeTuple]*RMethod)
    }
}


// RObject interface methods

func (p *RMultiMethod) IsZero() bool {
	return false
}

func (p *RMultiMethod) Type() *RType {
	return MultiMethodType
}

func (p *RMultiMethod) This() RObject {
	return p
}

func (p *RMultiMethod) IsUnit() bool {
	return true
}

func (p *RMultiMethod) IsCollection() bool {
	return false
}

func (p *RMultiMethod) String() string {
	return p.Name
}

func (p *RMultiMethod) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p *RMultiMethod) UUID() []byte {
	panic("A multimethod cannot have a UUID.")
	return nil
}

func (p *RMultiMethod) DBID() int64 {
	panic("A multimethod cannot have a DBID.")
	return 0
}

func (p *RMultiMethod) EnsureUUID() (theUUID []byte, err error) {
	panic("A multimethod cannot have a UUID.")
	return
}

func (p *RMultiMethod) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A multimethod cannot have a UUID.")
	return
}

func (p *RMultiMethod) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A multimethod cannot have a UUID.")
	return
}

func (p *RMultiMethod) UUIDstr() string {
	panic("A multimethod cannot have a UUID.")
	return ""
}

func (p *RMultiMethod) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A multimethod cannot have a UUID.")
	return
}

func (p *RMultiMethod) UUIDabbrev() string {
	panic("A multimethod cannot have a UUID.")
	return ""
}

func (p *RMultiMethod) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A multimethod cannot have a UUID.")
	return
}

func (p *RMultiMethod) RemoveUUID() {
	panic("A multimethod does not have a UUID.")
	return
}

func (p *RMultiMethod) Flags() int8 {
	panic("A multimethod has no Flags.")
	return 0
}

func (p *RMultiMethod) IsDirty() bool {
	return false
}
func (p *RMultiMethod) SetDirty() {
}
func (p *RMultiMethod) ClearDirty() {
}

func (p *RMultiMethod) IsIdReversed() bool {
	return false
}

func (p *RMultiMethod) SetIdReversed() {}

func (p *RMultiMethod) ClearIdReversed() {}

func (p *RMultiMethod) IsLoadNeeded() bool {
	return false
}

func (p *RMultiMethod) SetLoadNeeded()   {}
func (p *RMultiMethod) ClearLoadNeeded() {}

func (p *RMultiMethod) IsValid() bool { return true }
func (p *RMultiMethod) SetValid()     {}
func (p *RMultiMethod) ClearValid()   {}

func (p *RMultiMethod) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p *RMultiMethod) SetStoredLocally()     {}
func (p *RMultiMethod) ClearStoredLocally()   {}

func (p *RMultiMethod) IsProxy() bool { return false }

func (o *RMultiMethod) IsTransient() bool { return true }

func (o *RMultiMethod) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

/*
   A method implementation that applies to a particular tuple of argument types.
*/
type RMethod struct {
	multiMethod    *RMultiMethod
	ParameterNames []string               // names of parameters
	Signature      *RTypeTuple            // types of parameters
	ReturnSignature *RTypeTuple           // types of return values
	ReturnArgsNamed bool                  // Whether the return arguments are named and have to be initialized on method call.
	Code           *ast.MethodDeclaration // abstract syntax tree
	NumReturnArgs  int
	NumLocalVars   int
	PrimitiveCode  func([]RObject) []RObject
	Pkg            *RPackage  // the package that this method is defined in
}

func (p *RMethod) IsZero() bool {
	return false
}

func (m RMethod) String() string {
	return fmt.Sprintf("%s %v %v", m.multiMethod.Name, m.ParameterNames, m.Signature)
}

func (p *RMethod) Type() *RType {
	return MethodType
}

func (p *RMethod) This() RObject {
	return p
}

func (p *RMethod) IsUnit() bool {
	return true
}

func (p *RMethod) IsCollection() bool {
	return false
}

func (p *RMethod) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p *RMethod) UUID() []byte {
	panic("A Method cannot have a UUID.")
	return nil
}

func (p *RMethod) DBID() int64 {
	panic("A Method cannot have a DBID.")
	return 0
}

func (p *RMethod) EnsureUUID() (theUUID []byte, err error) {
	panic("A Method cannot have a UUID.")
	return
}

func (p *RMethod) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Method cannot have a UUID.")
	return
}

func (p *RMethod) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Method cannot have a UUID.")
	return
}

func (p *RMethod) UUIDstr() string {
	panic("A Method cannot have a UUID.")
	return ""
}

func (p *RMethod) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Method cannot have a UUID.")
	return
}

func (p *RMethod) UUIDabbrev() string {
	panic("A Method cannot have a UUID.")
	return ""
}

func (p *RMethod) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Method cannot have a UUID.")
	return
}

func (p *RMethod) RemoveUUID() {
	panic("A Method does not have a UUID.")
	return
}

func (p *RMethod) Flags() int8 {
	panic("A Method has no Flags.")
	return 0
}

func (p *RMethod) IsDirty() bool {
	return false
}
func (p *RMethod) SetDirty() {
}
func (p *RMethod) ClearDirty() {
}

func (p *RMethod) IsIdReversed() bool {
	return false
}

func (p *RMethod) SetIdReversed() {}

func (p *RMethod) ClearIdReversed() {}

func (p *RMethod) IsLoadNeeded() bool {
	return false
}

func (p *RMethod) SetLoadNeeded()   {}
func (p *RMethod) ClearLoadNeeded() {}

func (p *RMethod) IsValid() bool { return true }
func (p *RMethod) SetValid()     {}
func (p *RMethod) ClearValid()   {}

func (p *RMethod) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p *RMethod) SetStoredLocally()     {}
func (p *RMethod) ClearStoredLocally()   {}

func (p *RMethod) IsProxy() bool { return false }

func (o *RMethod) IsTransient() bool { return true }

func (o *RMethod) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}


/*
   Create a new relish method.
   Methods are collected by name into multi-methods.
   A multi-method maintains a list of method implementations under each
   parameter arity. At runtime, multi-method dispatch selects for application
   the method (of the correct arity), whose parameter type signature is 
   compatible with (same or a generalization of) the types of the actual arguments
   and whose parameter type signature is closest in type to the actual argument types,
   and whose parameter type signature, being arg-type-compatible and closest to the
   arg types, consists of the most specific types.

   This function determines whether a multi-method of the given name already
   exists in the runtime environment. If so, it adds the method implementation
   with the given argument type tuple to the appropriate list of method implementations
   of the multimethod. If the multi-method does not yet exist, it is first
   created and added to the runtime environment.

   If packageName is the empty string, it defaults to "relish.pl2012/core/inbuilt"

   Returns an error if a method implementation with the name and parameter type signature
   already exists. (What scope?)
   TODO How do we handle incremental compilation that includes redefinitions of method
   implementations?
*/
func (rt *RuntimeEnv) CreateMethod(packageName string, methodName string, parameterNames []string, parameterTypes []string, 
	returnValTypes []string,
	                  returnValNamed bool, numLocalVars int, allowRedefinition bool) (*RMethod, error) {
		return rt.CreateMethodGeneral(packageName, methodName, parameterNames, parameterTypes, 
						                  nil,
						                  nil,
						                  "",
						                  "",
						                  "",
						                  "",
						                  returnValTypes,
						                  returnValNamed, 
						                  numLocalVars, 
						                  allowRedefinition)
}

/*
   Create a new relish method.
   Methods are collected by name into multi-methods.
   A multi-method maintains a list of method implementations under each
   parameter arity. At runtime, multi-method dispatch selects for application
   the method (of the correct arity), whose parameter type signature is 
   compatible with (same or a generalization of) the types of the actual arguments
   and whose parameter type signature is closest in type to the actual argument types,
   and whose parameter type signature, being arg-type-compatible and closest to the
   arg types, consists of the most specific types.

   This function determines whether a multi-method of the given name already
   exists in the runtime environment. If so, it adds the method implementation
   with the given argument type tuple to the appropriate list of method implementations
   of the multimethod. If the multi-method does not yet exist, it is first
   created and added to the runtime environment.

   Returns an error if a method implementation with the name and parameter type signature
   already exists. (What scope?)
   TODO How do we handle incremental compilation that includes redefinitions of method
   implementations?
*/
func (rt *RuntimeEnv) CreateMethodGeneral(packageName string, methodName string, parameterNames []string, parameterTypes []string, 
	                  // FROM HERE DOWN IS NEW
	                  keyWordParameterDefaults map[string] RObject,
	                  keyWordParameterTypes map[string]string,
	                  wildcardKeywordsParameterName string,
	                  wildcardKeywordsParameterType string,
	                  variadicParameterName string,
	                  variadicParameterType string,
	                  // FROM HERE UP IS NEW
	                  returnValTypes []string,
	                  returnArgsNamed bool,
	                  numLocalVars int, allowRedefinition bool) (*RMethod, error) {

	if packageName == "" { 
		packageName = "relish.pl2012/core/inbuilt"
	}

    // Set the package into which the method is to be added.
    // If it is the default inbuilt functions package and does not exist, create it.
	pkg := rt.Packages[packageName]
	if pkg == nil && packageName == "relish.pl2012/core/inbuilt" {
		pkg = rt.CreatePackage("relish.pl2012/core/inbuilt")
	}


	arity := len(parameterTypes)
    numReturnArgs := len(returnValTypes)

    // TODO - Have to find and/or create this in the pkg object !!!!!!!!!!!!!!!!!
    //
    //
    //
    //

	//multiMethod, found := rt.MultiMethods[methodName]
	multiMethod, found := pkg.MultiMethods[methodName]


	if !found {
		multiMethod = newRMultiMethod(methodName, numReturnArgs, pkg)
		pkg.MultiMethods[methodName] = multiMethod
	} else {
		if multiMethod.NumReturnArgs != numReturnArgs {
		   return nil, fmt.Errorf("Method '%v' is already defined to have %v return arguments and cannot have other than that.", methodName, multiMethod.NumReturnArgs)
        }
        if multiMethod.Pkg != pkg { // Make sure we have a multimethod that this package is allowed to modify
	
	                                // TODO ---- WARNING ----- We modify a multimethod when caching! is it ok to 
	                                // do caching on a dependency package's multimethod? probably not!! 
	                                // Needs more thought. Ok for a package's multimethod to cache typetuples
	                                // where the types are not even known within that package??? Maybe it's ok
	                                // because we only cache based on a lookup done with that multimethod in first place!!!
	
	        multiMethod = multiMethod.Clone(pkg)
	        pkg.MultiMethods[methodName] = multiMethod
        }
	}
	
	
	if arity > multiMethod.MaxArity {
		multiMethod.MaxArity = arity
	}

	typeTuple, err := rt.getTypeTupleFromTypes(parameterTypes)
	if err != nil {
		return nil, err
	}
	resultTypeTuple, err := rt.getTypeTupleFromTypes(returnValTypes)
	if err != nil {
		return nil, err
	}	
	method := &RMethod{
		multiMethod:    multiMethod,
		ParameterNames: parameterNames,
		Signature:      typeTuple,
		ReturnSignature: resultTypeTuple,
		ReturnArgsNamed: returnArgsNamed,		
		NumReturnArgs:  numReturnArgs,
		NumLocalVars:   numLocalVars,
		Pkg:            pkg,
	}

	methodsOfRightArity, multiMethodHasArity := multiMethod.Methods[arity]

	_, found = multiMethod.CachedMethods[typeTuple]
	if found {
		if !allowRedefinition {
			return nil, fmt.Errorf("Method '%v' is already defined for types %v.", methodName, typeTuple)
		}

		for i, oldMethod := range methodsOfRightArity {
			if oldMethod.Signature == typeTuple {
				methodsOfRightArity[i] = method
			}
		}

	} else { // new method

		if multiMethodHasArity {
			multiMethod.Methods[arity] = append(methodsOfRightArity, method)
		} else {
			multiMethod.Methods[arity] = []*RMethod{method}
		}
	}
	return method, nil
}
