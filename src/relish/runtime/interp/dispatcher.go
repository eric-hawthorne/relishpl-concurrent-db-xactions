// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the operation of the intermediate-code interpreter
// in the relish language.

package interp

/*
   dispatcher.go -  multi-method dispatcher
*/


import (
   . "relish/runtime/data"
   . "relish/dbg"
)



type dispatcher struct {
   // typeTupleTree *TypeTupleTreeNode  // obsolete
   typeTupleTrees []*TypeTupleTreeNode  // new   
   emptyTypeTuple *RTypeTuple // a cached type tuple, to speed dispatch in special cases.

}

func newDispatcher(rt *RuntimeEnv) (d *dispatcher) {
   emptyTypeTuple := rt.TypeTupleTrees[0].GetTypeTuple([]RObject{})
   d =  &dispatcher{rt.TypeTupleTrees, emptyTypeTuple}
   
   return
}

/*
   The main dispatch function.
   First looks up a cache (map) of method implementations keyed by typetuples.
   a method will be found in this cache if the type tuple of the arguments
   has had the multimethod called on it before.
   If there is a cache miss, uses a multi-argument dynamic dispatch algorithm
   to find the best-matching method implementation (then caches the find under
   the type-tuple of the arguments for next time.)
   Returns the best method implementation for the types of the argument objects,
   or nil if the multimethod has no method signature which is compatible with
   the types of the argument objects.
   Also returns the type-tuple of the argument objects, which can be used to
   report the lack of a compatible method.
*/
func (d *dispatcher) GetMethod(mm *RMultiMethod, args []RObject) (*RMethod,*RTypeTuple) {
   typeTuple := d.typeTupleTrees[len(args)].GetTypeTuple(args)
   method,found := mm.CachedMethods[typeTuple]
   if !found {
      method = d.dynamicDispatch(mm,typeTuple)
      if method != nil {
         mm.CachedMethods[typeTuple] = method
      }
   }
   return method,typeTuple
}

	

/*
Special, degenerate case when we know there is only a single method implementation associated with the multimethod.
Returns the method, or nil if there is none. May return a method of any arity.
Note that this process may result in a multimethod whose method of non-zero arity is in its CachedMethods map as having zero arity. No real
big deal as long as this method is only ever called on multimethods that actually only have a single method implementation.
*/
func (d *dispatcher) GetSingletonMethod(mm *RMultiMethod) *RMethod {
   method,found := mm.CachedMethods[d.emptyTypeTuple]
   if !found {
      for _,methods := range mm.Methods {
         method = methods[0]
         break
      }      
      if method != nil {
         mm.CachedMethods[d.emptyTypeTuple] = method
      }
   }
   return method   
}






/*
   Same as GetMethod but for types instead of object instances.
*/
func (d *dispatcher) GetMethodForTypes(mm *RMultiMethod, types ...*RType) (*RMethod,*RTypeTuple) {
   typeTuple := d.typeTupleTrees[len(types)].GetTypeTupleFromTypes(types)
   method,found := mm.CachedMethods[typeTuple]
   if !found {
      method = d.dynamicDispatch(mm,typeTuple)
      if method != nil {
         mm.CachedMethods[typeTuple] = method
      }
   }
   return method,typeTuple
}

/*
   Find the method implementation of the multimethod whose parameter-type
   signature is more general than but minimal Euclidean distance
   (in multi-dimensional type specialization space) from the type-tuple
   of the actual argument objects.
   Specificity of the method that is chosen is determined in two ways.
   First, the method which is minimally different in types from the argument types
   is found. If more than one method is equally close in type signature to the
   argument types (measuring Euclidean distance down the specialization paths),
   then the tie is broken by selecting the method whose signature is most specific
   in types compared to the top types in the ontology known by the process.
   If there is still a tie, the method which was encountered first (in the multimethod's
   list of methods of a particular arity) is chosen. This is somewhat arbitrary.

   TODO Should I search downward in specialization chains from the type-tuple signature of     
   each correct-arity method of the multimethod to find the argument type tuple,
   or up the supertype path from each type in the argument type-tuple?

   TODO This can no doubt be optimized. Try upwards first.  

   Returns the most specific type-compatible method or nil if none is found.


*/
func (d *dispatcher) dynamicDispatch(mm *RMultiMethod, argTypeTuple *RTypeTuple) *RMethod {
   candidateMethods,found := mm.Methods[len(argTypeTuple.Types)]
   if ! found {

      
      Log(INTERP2_, "No '%s' method has arity %v.\n",mm.Name,len(argTypeTuple.Types))      
      return nil
   }
   
   var minSpecializationDistance float64 = 99999
   var maxSupertypeSpecificity float64 = 99999
   var closestCandidateMethod *RMethod = nil
   for _,candidateMethod := range candidateMethods {
      // DEBUG fmt.Printf("Checking for match with %v.\n",candidateMethod)
      specializationDistance,supertypeSpecificity,incompatible := argTypeTuple.SpecializationDistanceFrom(candidateMethod.Signature)
      // DEBUG fmt.Printf("specializationDistance=%v, supertypeSpecificity=%v, incompatible=%v\n",specializationDistance,supertypeSpecificity,incompatible)
      if incompatible {
         continue
      }
      if specializationDistance < minSpecializationDistance {
         closestCandidateMethod = candidateMethod
         minSpecializationDistance = specializationDistance
         maxSupertypeSpecificity = supertypeSpecificity
      } else if specializationDistance == minSpecializationDistance {
         if supertypeSpecificity > maxSupertypeSpecificity {
             closestCandidateMethod = candidateMethod
             maxSupertypeSpecificity = supertypeSpecificity
         }
      }
   }
   return closestCandidateMethod
}

