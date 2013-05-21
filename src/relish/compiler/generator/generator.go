// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// "code generator" for relish.
// Currently, tours the file ast (should be package ast) and creates types and methods.

package generator

import (
	"fmt"	
	"strings"
	"relish/compiler/ast"
	"relish/compiler/token"
	"relish/runtime/data"
	"relish/runtime/interp"
	"relish/rterr"	
	"relish"
	. "relish/defs"
	. "relish/dbg"
)



// The parser structure holds the parser's internal state.
type Generator struct {
	files map[*ast.File]string  // map from ast filenode to filename root without .rel or .rlc suffix
	Interp *interp.Interpreter
	th *interp.Thread
	packageName string // Full name of origin, artifact, and package
	packagePath string // Full name of origin, artifact, and package, ending in a /
	pkg *data.RPackage // the current package which this file is being generated into
}

func NewGenerator(files map[*ast.File]string) *Generator {
	interpreter := interp.NewInterpreter(data.RT)
	thread := interpreter.NewThread(nil)
	var packageName string
	var packagePath string
	for file := range files {
		packageName = file.Name.Name
		packagePath = packageName + "/"
		break
	}
	return &Generator{files,interpreter,thread,packageName,packagePath,nil}
}

/*
TODO HAVE TO CHANGE THIS TO GENERATE CODE FOR A WHOLE PACKAGE !!!!

Given a Relish intermediate-code file node tree, create objects in the runtime for data-types, methods, and constants defined in the file.
Assumes imported packages have already been loaded; thus the objects defined in files of the imported packages have been generated.
*/
func (g *Generator) GenerateCode() {	
   types := make(map[*data.RType]bool)	
   attributeOrderings := make(map[string]*data.AttributeSpec)
   g.generatePackage()
   g.generateTypes(types, attributeOrderings)
   g.generateMethods()		
   g.generateConstants()
   g.generateRelations(types, attributeOrderings) 
   g.configureAttributeSortings(attributeOrderings)
   g.ensureAttributeAndRelationTables(types)
   g.Interp.DeregisterThread(g.th)
   // g.pkg.ListMethods() // Debugging - temporary
}


/*
Checks to see if the package which this source code file says its in already
exists in the runtime. If not, creates it. Sets the pkg variable of the Generator
to refer to the package.
*/
func (g *Generator) generatePackage() {
	
    relish.EnsureDatabase() 
    // creates and/or creates a connection to the running artifact's database.
    // Amongst other things, initializes from db the maps between package names and shortnames in the runtime.
	
    g.pkg = data.RT.Packages[g.packageName]
    if g.pkg != nil {
    	rterr.Stopf("Package %s has already been loaded. Error to load it again.",g.packageName)
    }
	g.pkg = data.RT.CreatePackage(g.packageName, false)  // This should init MMMap from core/builtin package   
	
	g.updatePackageDependenciesAndMultiMethodMap()
	
	g.th.ExecutingPackage = g.pkg
}

/*
Go through the (already loaded) packages that the newly generated package is dependent on, and update the new package's
multimethod map to incorporate multimethods and methods from the dependency packages.
Now does this all at once for all files in the current package being loaded.
IMPORTANT NOTE: That means every file in the current package has access to method-implementations declared in packages which
are not explicitly imported into that particular code file of the current package, but are imported into another
file of the current package. Anyway, currently, the particular code file also has access (through dispatch) to method-implementations
declared even in indirect dependency packages of any of the packages explicitly imported into the current package.  
*/
func (g *Generator) updatePackageDependenciesAndMultiMethodMap() {
	for file := range g.files {
		imports := file.RelishImports  // package specifications
		for _,importedPackageSpec := range imports {		
			dependencyPackageName := importedPackageSpec.OriginAndArtifactName + "/pkg/" + importedPackageSpec.PackageName
			dependencyPackage, dependencyAlreadyProcessed := g.pkg.Dependencies[dependencyPackageName]
	        if ! dependencyAlreadyProcessed {
			   dependencyPackage = data.RT.Packages[dependencyPackageName]
			   g.pkg.Dependencies[dependencyPackageName] = dependencyPackage

			   g.updatePackageMultiMethodMap(dependencyPackage)
		    }

		}
    }
}

/*
Update the package's multimethod map to incorporate exported multimethods and methods from a dependency package.
Helper for updatePackageDependenciesAndMultiMethodMap.
*/
func (g *Generator) updatePackageMultiMethodMap(dependencyPackage *data.RPackage) {
//   if strings.HasSuffix(g.pkg.Name,"basic_tests") {
//      defer Un(Trace(ALWAYS_, "updatePackageMultiMethodMap of", g.pkg.Name,"from",dependencyPackage.Name))   
//   }	
   for multiMethodName,multiMethod := range dependencyPackage.MultiMethods {
   // Loop through all of the dependency package's multimethods.


	
      // Consider only those multi-methods which are actually owned directly by the dependency package

      if multiMethod.Pkg != dependencyPackage {
//	      if strings.HasSuffix(g.pkg.Name,"basic_tests") && multiMethod.Pkg != data.RT.InbuiltFunctionsPackage {
//		     fmt.Println("multimethod is not owned by dependency package!") 
//		     fmt.Println("it is owned by",multiMethod.Pkg) 		
//		  }
		
	      continue  // must be a multimethod owned by a dependency of the dependency - ignore
      }

      // And consider only the exported multimethods of the dependency package

      if ! multiMethod.IsExported {
	     continue
      }  


//      if multiMethodName == multiMethod.Name {
//         fmt.Println(multiMethodName)
//      } else {
//	     fmt.Println(multiMethodName,"->",multiMethod.Name)
//      }
	

	  if multiMethod.MaxArity == 0 {
			
	     // This multimethod from the imported package takes no arguments, so we need to find it
	     // it by its package-qualified name. In source code, a name like vehicles.startRace
	     // will be needed to invoke this method.
	     multiMethod = multiMethod.Clone(g.pkg)
	     multiMethod.IsExported = false  // This cloned multimethod is still not "owned" by current package,
	     // and it should not be re-exported, so mark it as non-exportable.
	     qualifiedMethodName := dependencyPackage.Name + "/" + multiMethodName
	     g.pkg.MultiMethods[qualifiedMethodName] = multiMethod 			
         continue
	  } 

  	  myMultiMethod := g.pkg.MultiMethods[multiMethodName]


   	  if myMultiMethod == nil {
       // If current package does not yet have this multimethod at all...	
	
   	   	   g.pkg.MultiMethods[multiMethodName] = multiMethod  // use the mm from dependency package

   	  } else if myMultiMethod != multiMethod {
	   // Current package has a different mm instance than the dependency package's mm
	      		
   	   	   // Merge them if possible, else complain!! throw runtime error!

           if myMultiMethod.NumReturnArgs != multiMethod.NumReturnArgs {
	          rterr.Stopf("Package %s defines %s to return %d values so can't be imported directly or indirectly into package %s which defines %s to return %d values.",dependencyPackage.Name,multiMethod.Name,multiMethod.NumReturnArgs,g.pkg.Name,myMultiMethod.Name,myMultiMethod.NumReturnArgs)
	
	          // TODO Shit! I'm inheriting methods indirectly!!! Not allowed !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	
	          // Will have to actually create new multimethods for EACH new package right up front, 
	          // putting in ONLY those methods which are defined in the DIRECT dependency packages.
           }

   	   	   // To merge them, I have to know if my multiMethod really belongs to me or if borrowing
   	   	   // it so I have to copy it.

           if myMultiMethod.Pkg != g.pkg {
	       // The mm in current package's list does not belong to current package
	       // So create a clone mm which is owned by the current package
	
	          myMultiMethod = myMultiMethod.Clone(g.pkg)
	          myMultiMethod.IsExported = false  // This merged multimethod is still not "owned" by current package,
	          // and it should not be re-exported, so mark it as non-exportable.
	
	          g.pkg.MultiMethods[multiMethodName] = myMultiMethod	
           }

           // Merge in methods from the dependency package's mm into current-package-owned mm of the same name.
           myMultiMethod.MergeInNewMethodsFrom(multiMethod)

   	  }
   	  // else I've (current package has) already got the multimethod - everything's cool
   }
}






/*
If the type name is unqualified (has no package path), prefixes it with the current package path.
A little bit inefficient.
*/
func (g *Generator) qualifyTypeName(typeName string) string {
   return g.Interp.QualifyTypeName(g.packagePath, typeName)
   /*
   if strings.LastIndex(typeName,"/") == -1 && ! BuiltinTypeName[typeName] {
       return g.packagePath + typeName  	
   }	
   return typeName
   */
}



/*
Returns the fully qualified name of the type.
If the type spec is of a collection type, then ensures that the composite type exists in the runtime, 
and if not, creates it. 
TODO Handle Maps
TODO Handle nested collection types, like [] [] String

TypeSpec struct {
	Doc            *CommentGroup       // associated documentation; or nil
	Name           *Ident              // type name (or type variable name)
	Type           Expr                // *Ident, *ParenExpr, *SelectorExpr, *StarExpr, or any of the *XxxTypes
	Comment        *CommentGroup       // line comments; or nil
	Params         []*TypeSpec         // Type parameters (egh)
	SuperTypes     []*TypeSpec         // Only valid if this is a type variable (egh)
	CollectionSpec *CollectionTypeSpec // nil or a collection specification
	typ interface{}  // The actual *RType - associated at generation time.
}

type CollectionTypeSpec struct {
	Kind        token.Token
	LDelim      token.Pos
	RDelim      token.Pos
	IsSorting   bool
	IsAscending bool
	OrderFunc   string
}
*/
func (g *Generator) ensureTypeName(typeSpec *ast.TypeSpec, fileNameRoot string) string {
	
   typ, err := g.Interp.EnsureType(g.packagePath, typeSpec) 
   if err != nil {
	  rterr.Stopf("Error in file %s: %s", fileNameRoot +".rel", err.Error())	   	
   }
   return typ.Name

/*   
   var typ *data.RType
   var err error	
   baseTypeName := g.qualifyTypeName(typeSpec.Name.Name) 
   baseType, baseTypeFound := data.RT.Types[baseTypeName]

   if ! baseTypeFound {
      rterr.Stopf("Error in file %s: Type %s not found.", fileNameRoot +".rel", baseTypeName)
      // panic(fmt.Sprintf("Error in file %s: Type %s not found.", fileNameRoot +".rel", baseTypeName))		
   }
  	
   if typeSpec.CollectionSpec == nil {
	  typ = baseType
   } else {	
	    switch typeSpec.CollectionSpec.Kind {
	    case token.LIST:		
			typ, err = data.RT.GetListType(baseType)
			if err != nil {
			   rterr.Stopf("Error in file %s: %s", fileNameRoot +".rel", err.Error())				
			}
	    case token.SET:
			typ, err = data.RT.GetSetType(baseType)
			if err != nil {
			   rterr.Stopf("Error in file %s: %s", fileNameRoot +".rel", err.Error())				
			}			
	    case token.MAP:
	       panic("I don't handle Map types yet.")			
	    }
	}	
    return typ.Name
*/
}


/*
Processes the TypeDecls from a set of ast.File objects (which have been created by the parser.)
Generates the runtime environment's objects for datatypes and attributes, and also ensures that db tables exist for these.

Runtime *data.RType's are placed into the argument hashtable once created here.
*/
func (g *Generator) generateTypes(types map[*data.RType]bool, attributeOrderings map[string]*data.AttributeSpec) {

	allTypeDecls := make(map[string]*ast.TypeDecl)  // Map from full type name to *ast.TypeDecl

	typeDeclFile := make(map[string]string)  // map of type name to which file it is declared in - the filenameroot is the value in the map
	
	g.generateTypesWithoutAttributes(allTypeDecls, types, typeDeclFile)
	g.generateAttributes(allTypeDecls, types, typeDeclFile, attributeOrderings)
	g.ensureTypeTables(types)
}

/*
Processes the TypeDecls from a set of ast.File objects (which have been created by the parser.)
Generates the runtime environment's objects for datatypes and attributes, and also ensures that db tables exist for these.

Runtime *data.RType's are placed into the argument hashtable once created here.

Note. This function has to operate recursively, since a given supertype (parent type) might not yet exist as a generated
RType object, but might be in this same package so needs to be generated now anyway.

Note. This function does not create the attribute specification part of the runtime RType objects.

Attribute generation has to wait til all of the RTtypes for the package are generated and put in the RT.types map,
so attribute generation will be done as a separate pass.


	allTypeDecls := make(map[string]*ast.TypeDecl)

*/
func (g *Generator) generateTypesWithoutAttributes(allTypeDecls map[string]*ast.TypeDecl, types map[*data.RType]bool, typeDeclFile map[string]string) {
	
	typesBeingGenerated := make(map[string]bool)  // full names of types in the middle of being generated. Use to avoid inheritance loops.


	// Collect all type declarations from all files into a map by full type name
    //
	for file, fileRoot := range g.files { 
		for _,typeDeclaration := range file.TypeDecls {
		   // egh Nov 12, 12	
		   //typeName := g.packagePath + typeDeclaration.Spec.Name.Name		
		   typeName := typeDeclaration.Spec.Name.Name
		   if allTypeDecls[typeName] != nil {
			  slashPos := strings.LastIndex(typeName, "/")
			  typeLocalName := typeName[slashPos+1:]
		   	  rterr.Stopf("Type '%s' declaration in file %s.rel is 2nd declaration of this type found in package.",typeLocalName,fileRoot)
		   }
           allTypeDecls[typeName] = typeDeclaration
           typeDeclFile[typeName] = fileRoot
        }
    }

    // Generate the runtime RType objects, recursively.

    for typeName, typeDecl := range allTypeDecls {

	   _, found := data.RT.Types[typeName]
	   if found {
		  continue // the recursion has already generated this type.
	   }
       g.generateTypeWithoutAttributes(typeName, typeDecl, allTypeDecls, types, typesBeingGenerated, typeDeclFile)

    }
}


/*
   Generate a runtime RType object for the type declaration.
   If the parent types are not already generated and are in this package, recurse to generate the parent type before finishing generating this one.
*/
func (g *Generator) generateTypeWithoutAttributes(typeName string, typeDeclaration *ast.TypeDecl, allTypeDecls map[string]*ast.TypeDecl, types map[*data.RType]bool, typesBeingGenerated map[string]bool, typeDeclFile map[string]string) {
       
   _,found := typesBeingGenerated[typeName]
   if found {
	  rterr.Stopf("Type '%s' declaration in file %s.rel is involved in a type inheritance loop!",typeDeclaration.Spec.Name.Name,typeDeclFile[typeName])
   }
   typesBeingGenerated[typeName] = true

   typeSpec := typeDeclaration.Spec    	
   slashPos := strings.LastIndex(typeSpec.Name.Name,"/") 
   typeLocalName := typeSpec.Name.Name[slashPos+1:]
   typeShortName := g.pkg.ShortName + "/" + typeLocalName

   var parentTypeNames []string

   for _,parentTypeSpec := range typeSpec.SuperTypes {

      parentTypeName := g.qualifyTypeName(parentTypeSpec.Name.Name)
	  
	  _, parentFound := data.RT.Types[parentTypeName]
	  if ! parentFound {
	  	 parentTypeDecl, parentDeclaredInPackage := allTypeDecls[parentTypeName]
	  	 if parentDeclaredInPackage {
            g.generateTypeWithoutAttributes(parentTypeName, parentTypeDecl, allTypeDecls, types, typesBeingGenerated, typeDeclFile)    
         }  

         // If the parent type was not declared in this package but is not already declared in some other package, then the
         // parent type is missing in action. 
         // Ignore this problem for the moment here. Let the rt.CreateType method report this error.
	  }
	  parentTypeNames = append(parentTypeNames, parentTypeName)
   } 
	
   theNewType, err := data.RT.CreateType(typeName, typeShortName, parentTypeNames)
   if err != nil {
      // panic(err)
      sourceFilename := typeDeclFile[typeName] + ".rel"
	  rterr.Stopf("Error creating type %s (%s): %s", typeName, sourceFilename, err.Error())
   }	

   types[theNewType] = true	 // record that this is one of the new RTypes we generated !

   delete(typesBeingGenerated, typeName)  // remove from the inheritance loop detection map
}


/*
Processes the TypeDecls list of a ast.File object (which has been created by the parser.)
Generates the runtime environment's objects for attributes of datatypes.
Assumes the RType objects have already been created in the runtime for each datatype by a previous pass over the intermediate-code files.
*/
func (g *Generator) generateAttributes(allTypeDecls map[string]*ast.TypeDecl, types map[*data.RType]bool, typeDeclFile map[string]string, orderings map[string]*data.AttributeSpec) {
    for theNewType := range types {
       typeName := theNewType.Name
       typeDeclaration := allTypeDecls[typeName]
       sourceFilename := typeDeclFile[typeName] + ".rel"

	   for _,attrDecl := range typeDeclaration.Attributes {
		  var minCard int32 = 1
		  var maxCard int32 = 1
		  
          attributeName := attrDecl.Name.Name
          multiValuedAttribute := (attrDecl.Arity != nil)
          if multiValuedAttribute {
	         minCard = int32(attrDecl.Arity.MinCard)
	         maxCard = int32(attrDecl.Arity.MaxCard)  // -1 means N
          }
          

          var collectionType string 

          var orderFuncOrAttrName string = ""
          var isAscending bool 

          if attrDecl.Type.CollectionSpec == nil {
	          if attrDecl.Type.NilAllowed {
		         minCard = 0
		      }
	      } else {
              switch attrDecl.Type.CollectionSpec.Kind {
	             case token.SET:
			        if attrDecl.Type.CollectionSpec.IsSorting {
			           collectionType = "sortedset"
                       orderFuncOrAttrName = attrDecl.Type.CollectionSpec.OrderFunc	
                       isAscending = attrDecl.Type.CollectionSpec.IsAscending		
		            } else {
			           collectionType = "set"			
		            }		
		         case token.LIST:		
			        if attrDecl.Type.CollectionSpec.IsSorting {			
			           collectionType = "sortedlist"
                       orderFuncOrAttrName = attrDecl.Type.CollectionSpec.OrderFunc	
                       isAscending = attrDecl.Type.CollectionSpec.IsAscending			
		            } else {
			           collectionType = "list"			
		            }
			     case token.MAP:
			        if attrDecl.Type.CollectionSpec.IsSorting {
			           collectionType = "sortedmap"
                       orderFuncOrAttrName = attrDecl.Type.CollectionSpec.OrderFunc	
                       isAscending = attrDecl.Type.CollectionSpec.IsAscending			
		            } else {
			           collectionType = "map"			
		            }				
	           }	
          } 


/*	type CollectionTypeSpec struct {
	   Kind token.Token
	   LDelim token.Pos
	   RDelim token.Pos
	   IsSorting bool
	   IsAscending bool
	   OrderFunc string
	}
*/	
 
          attributeTypeName := g.qualifyTypeName(attrDecl.Type.Name.Name)
                            // g.ensureTypeName(attrDecl.Type, sourceFilename)  ??????

          
/*"vector"
RelEnd
   ...
   CollectionType string // "list", "sortedlist","set", "sortedset", "map", "stringmap", "sortedmap","sortedstringmap",""
   OrderAttrName string   // which primitive attribute of other is it ordered by when retrieving? "" if none

   OrderMethod *RMultiMethod
*/

         

	      _,err := data.RT.CreateAttribute(typeName,
									 	 attributeTypeName,
										 attributeName,
										 minCard,
										 maxCard,   // Is the -1 meaning N respected in here???? TODO
										 collectionType,
				                         orderFuncOrAttrName,
				                         isAscending,
										 false,
										 false,
										 false,
										 orderings)
		   if err != nil {
		      rterr.Stopf("Error creating attribute %s.%s (%s): %s", typeName, attributeName, sourceFilename, err.Error())
		   }
        }
    }
}

// g.Interp.Dispatcher()

/*
Ensure the persistence data model is created for the type.
*/
func (g *Generator) ensureTypeTables(types map[*data.RType]bool) {

    for theNewType := range types {
		err := data.RT.DB().EnsureTypeTable(theNewType) 
		if err != nil {
		      panic(err)
		}
    }
}



/*
Give the sorted multi-valued attributes their sorting attribute or sorting method.
This had to be delayed until all attributes and methods in the package were
generated.
*/
func (g *Generator) configureAttributeSortings(orderings map[string]*data.AttributeSpec) {
	for key, attr := range orderings {
       		
       orderFuncOrAttrName := key[strings.LastIndex(key,KEY_PART_SEPARATOR)+len(KEY_PART_SEPARATOR):]		
       g.configureAttributeSorting(orderFuncOrAttrName, attr)        		
	}
}

/*
Give the attribute its sorting attribute or sorting method.
This had to be delayed until all attributes and methods in the package were
generated.
*/
func (g *Generator) configureAttributeSorting(orderFuncOrAttrName string, attr *data.AttributeSpec) {

	var orderAttr *data.AttributeSpec
	var attrFound bool
	var orderMethod *data.RMultiMethod

	// TODO Find out which of attribute or method it is. How????
	// Should be along the lines of: 
	// If there is a method of that name defined which takes only typ2 as arg signature, then use that function,
	// Otherwise could check if there is an attribute of that name on typ2 or its supertypes, and if not,
	// throw a type compatibility panic.

	var orderMethodArity int32 = 0

	orderMethod, methodFound := g.pkg.MultiMethods[orderFuncOrAttrName]

	if methodFound {
		
		dispatcher := g.Interp.Dispatcher()

		// Find out if there is an arity 1 method
		unaryMethod, _ := dispatcher.GetMethodForTypes(orderMethod, attr.Part.Type)
		// Find out if there is an arity 2 method
		binaryMethod, _ := dispatcher.GetMethodForTypes(orderMethod, attr.Part.Type, attr.Part.Type)

		if unaryMethod != nil && binaryMethod != nil {
			rterr.Stopf("Can't order collection because ambiguous unary and binary method '%s' for type %s.", orderFuncOrAttrName, attr.Part.Type.Name)
		}

		if unaryMethod != nil {
			orderMethodArity = 1
		} else if binaryMethod != nil {
			orderMethodArity = 2
		} else {
			rterr.Stopf("Can't order collection. No unary or binary method '%s' found for type %s.", orderFuncOrAttrName, attr.Part.Type.Name)
		}
	} else {
		orderAttr, attrFound = attr.Part.Type.GetAttribute(orderFuncOrAttrName)
		if !attrFound {
			rterr.Stopf("Can't order collection. No method or attribute '%s' found in type %s or supertypes.", orderFuncOrAttrName, attr.Part.Type.Name)
		}
	}

	attr.Part.OrderAttr = orderAttr
	attr.Part.OrderMethod = orderMethod
	attr.Part.OrderMethodArity = orderMethodArity  // 1 or 2
}





func (g *Generator) ensureAttributeAndRelationTables(types map[*data.RType]bool) {
	for typ := range types {
		// ensure the persistence data model is created for  the type's attributes and relations

		err := data.RT.DB().EnsureAttributeAndRelationTables(typ) 
		if err != nil {
		      panic(err)
		}		
	}
}

/*
TODO Need to add the constraint that the method is public here !!!!
*/
func (g *Generator) isWebDialogHandlerMethod(fileNameRoot string) bool {
	return strings.HasSuffix(fileNameRoot,"dialog") && strings.Contains(g.packagePath,"/pkg/web/")
}


func (g *Generator) generateMethods() {
	
	for file, fileNameRoot := range g.files {	
		for _,methodDeclaration := range file.MethodDecls {

		   methodName := methodDeclaration.Name.Name
		   
		   // main functions and init<TypeName> functions and web dialog handler functions are explicitly package name qualified.
		   //
	  	   // TODO   OOPS Which package are we executing when looking for web handler methods?????
	
	       // TODO TODO Do we really need to prepend package name to init... method names?
	       // This implies we cannot add constructors of type package1.Type1 in package2
	
		   if ((methodName == "main") || 
/* Now done at parser.go: 1877
		       (strings.HasPrefix(methodName,"init") && len(methodName) > 5 && 'A' <= methodName[4] && methodName[4] <= 'Z' && !BuiltinTypeName[methodName[4:]]) ||
		       */ 
		       (g.isWebDialogHandlerMethod(fileNameRoot))) {
		      methodName = g.packagePath + methodName	
		   }
	
		   var parameterNames []string
		   var nilArgAllowed []bool
		   var parameterTypes []string

		
           var wildcardKeywordsParameterName string
           var wildcardKeywordsParameterType string		
		   var variadicParameterName string
		   var variadicParameterType string	
		
		   var returnValTypes []string
		   var returnArgsAreNamed bool		   
		   for _,inputArgDecl := range methodDeclaration.Type.Params {
			
			  nilArgAllowed = append(nilArgAllowed, inputArgDecl.Type.NilAllowed)
			
			  if inputArgDecl.IsVariadic {
                 if inputArgDecl.Type.CollectionSpec.Kind == token.LIST { 
			         variadicParameterName = inputArgDecl.Name.Name
			         variadicParameterType = g.ensureTypeName(inputArgDecl.Type, fileNameRoot)
//			         variadicParameterType = g.qualifyTypeName(inputArgDecl.Type.Name.Name)  // g.ensureTypeName(inputArgDecl.Type, fileNameRoot) ???
			
			
	              } else { // inputArgDecl.Type..CollectionSpec.Kind == token.MAP				
				     wildcardKeywordsParameterName = inputArgDecl.Name.Name
				     wildcardKeywordsParameterType = g.ensureTypeName(inputArgDecl.Type, fileNameRoot)
//				     wildcardKeywordsParameterType = g.qualifyTypeName(inputArgDecl.Type.Name.Name)  // g.ensureTypeName(inputArgDecl.Type, fileNameRoot)

			      }
			  } else {
			     parameterNames = append(parameterNames, inputArgDecl.Name.Name)
			     parameterTypes = append(parameterTypes, g.ensureTypeName(inputArgDecl.Type, fileNameRoot))
		      }
		   }	   
	
		   for _,returnArgDecl := range methodDeclaration.Type.Results {
               if returnArgDecl.Name != nil {
               	  returnArgsAreNamed = true
               }
			   returnValTypes = append(returnValTypes, g.ensureTypeName(returnArgDecl.Type, fileNameRoot))               
		   }

           var rMethod *data.RMethod
           var err error


           if methodDeclaration.IsClosureMethod {
               rMethod, err = data.RT.CreateClosureMethod(g.packageName,
			   	                                          file,
			   	                                          methodName,
			   	                                          parameterNames, 
				   	                                      nilArgAllowed,		   	                                          
				                                          parameterTypes, 
				                                          returnValTypes,
				                                          returnArgsAreNamed,
				                                          methodDeclaration.NumLocalVars,
				                                          methodDeclaration.NumFreeVars )				
	       } else {
		       allowRedefinition := false	

			   // FuncType.	Params  []*InputArgDecl // input parameter declarations. Can be empty list.
		   	   // 	        Results []*ReturnArgDecl // (outgoing) result declarations; Can be empty list.	
	
			   rMethod, err = data.RT.CreateMethodV(g.packageName,
			   	                                   file,
			   	                                   methodName,
			   	                                   parameterNames, 
			   	                                   nilArgAllowed,
				                                   parameterTypes, 
												   wildcardKeywordsParameterName,
												   wildcardKeywordsParameterType,				
												   variadicParameterName,
												   variadicParameterType,				
				                                   returnValTypes,
				                                   returnArgsAreNamed,
				                                   methodDeclaration.NumLocalVars,
				                                   allowRedefinition  )
		   }
		
		   if err != nil {
		       // panic(err)
			   rterr.Stopf("Error creating method %s (%s): %s", methodName, fileNameRoot +".rel", err.Error())	
		   }
	
	       rMethod.Code = methodDeclaration // abstract syntax tree	

		   Logln(GENERATE__, rMethod)		
	    }	
    }
}


/*
If the const name is unqualified (has no package path), prefixes it with the current package path.
A little bit inefficient.
*/
func (g *Generator) qualifyConstName(constName string) string {
   if strings.LastIndex(constName,"/") == -1 {
       return g.packagePath + constName  	
   }	
   return constName
}


func (g *Generator) generateConstants () {
	for file, fileNameRoot := range g.files {	
		for _,constDeclaration := range file.ConstantDecls {
		
		   constantName := g.qualifyConstName(constDeclaration.Name.Name)
		   g.Interp.EvalExpr(g.th,constDeclaration.Value)
		   obj := g.th.Pop()
		   err := data.RT.CreateConstant(constantName,obj)
		   if err != nil {
		       // panic(err)
			   rterr.Stopf("Error in file %s: %s", fileNameRoot +".rel", err.Error())	
		   }	
	    }	
    }
}

func (g *Generator) TestWalk() {
	for file, fileNameRoot := range g.files {	
		fmt.Println("======================",fileNameRoot+".rel","========================")	
	   p := &NodePrinter{}
	   ast.Walk(p, file)
    }
}



/*
Processes the RelationDecls list of a ast.File object (which has been created by the parser.)
Generates the runtime environment's objects for relations between datatypes, and also ensures 
that db tables exist for these.
*/
func (g *Generator) generateRelations(types map[*data.RType]bool, orderings map[string]*data.AttributeSpec) {

	for file, fileNameRoot := range g.files {
		for _,relationDeclaration := range file.RelationDecls {
		
	       end1 := relationDeclaration.End1
	       end2 := relationDeclaration.End2 


		   var minCard1 int32 = 1
		   var maxCard1 int32 = 1
	  
	       attributeName1 := end1.Name.Name
	       multiValuedAttribute1 := (end1.Arity != nil)
	       if multiValuedAttribute1 {
	          minCard1 = int32(end1.Arity.MinCard)
	          maxCard1 = int32(end1.Arity.MaxCard)  // -1 means N
	       }
        
	        var collectionType1 string 

	        var orderFuncOrAttrName1 string = ""
	        var isAscending1 bool 

	        if end1.Type.CollectionSpec != nil {
	            switch end1.Type.CollectionSpec.Kind {
	            case token.SET:
		        if end1.Type.CollectionSpec.IsSorting {
		           collectionType1 = "sortedset"
	                     orderFuncOrAttrName1 = end1.Type.CollectionSpec.OrderFunc	
	                     isAscending1 = end1.Type.CollectionSpec.IsAscending		
	            } else {
		           collectionType1 = "set"			
	            }		
	         case token.LIST:
		        if end1.Type.CollectionSpec.IsSorting {
		           collectionType1 = "sortedlist"
	                     orderFuncOrAttrName1 = end1.Type.CollectionSpec.OrderFunc	
	                     isAscending1 = end1.Type.CollectionSpec.IsAscending			
	            } else {
		           collectionType1 = "list"			
	            }
		     case token.MAP:
		        if end1.Type.CollectionSpec.IsSorting {
		           collectionType1 = "sortedmap"
	                     orderFuncOrAttrName1 = end1.Type.CollectionSpec.OrderFunc	
	                     isAscending1 = end1.Type.CollectionSpec.IsAscending			
	            } else {
		           collectionType1 = "map"			
	            }				
	          }	
	        }

		   var minCard2 int32 = 1
		   var maxCard2 int32 = 1
	  
	       attributeName2 := end2.Name.Name
	       multiValuedAttribute2 := (end2.Arity != nil)
	       if multiValuedAttribute2 {
	          minCard2 = int32(end2.Arity.MinCard)
	          maxCard2 = int32(end2.Arity.MaxCard)  // -1 means N
	       }
        
	        var collectionType2 string 

	        var orderFuncOrAttrName2 string = ""
	        var isAscending2 bool 

	        if end2.Type.CollectionSpec != nil {
	            switch end2.Type.CollectionSpec.Kind {
	            case token.SET:
		        if end2.Type.CollectionSpec.IsSorting {
		           collectionType2 = "sortedset"
	                     orderFuncOrAttrName2 = end2.Type.CollectionSpec.OrderFunc	
	                     isAscending2 = end2.Type.CollectionSpec.IsAscending		
	            } else {
		           collectionType2 = "set"			
	            }		
	         case token.LIST:
		        if end2.Type.CollectionSpec.IsSorting {
		           collectionType2 = "sortedlist"
	                     orderFuncOrAttrName2 = end2.Type.CollectionSpec.OrderFunc	
	                     isAscending2 = end2.Type.CollectionSpec.IsAscending			
	            } else {
		           collectionType2 = "list"			
	            }
		     case token.MAP:
		        if end2.Type.CollectionSpec.IsSorting {
		           collectionType2 = "sortedmap"
	                     orderFuncOrAttrName2 = end2.Type.CollectionSpec.OrderFunc	
	                     isAscending2 = end2.Type.CollectionSpec.IsAscending			
	            } else {
		           collectionType2 = "map"			
	            }				
	          }	
	        }

	        typeName1 := g.qualifyTypeName(end1.Type.Name.Name)
	        typeName2 := g.qualifyTypeName(end2.Type.Name.Name)

	        type1, type2, err := data.RT.CreateRelation(typeName1,
		                                    attributeName1,
		                                    minCard1,
											maxCard1,
											collectionType1,
											orderFuncOrAttrName1,
											isAscending1,	
											typeName2,
											attributeName2,
											minCard2,
											maxCard2,
											collectionType2,
											orderFuncOrAttrName2,
											isAscending2,	
											false,
											orderings) 



	       if err != nil {
	           // panic(err)
			   rterr.Stopf("Error creating relation %s %s -- %s %s (%s): %s", typeName1, attributeName1, attributeName2, typeName2, fileNameRoot +".rel", err.Error())	
	       }

	       types[type1] = true	
	       types[type2] = true	
	   }
	}
}



/*	type CollectionTypeSpec struct {
	   Kind token.Token
	   LDelim token.Pos
	   RDelim token.Pos
	   IsSorting bool
	   IsAscending bool
	   OrderFunc string
	}
*/	
 /*
          attributeTypeName := g.qualifyTypeName(attrDecl.Type.Name.Name)

          
/*"vector"
RelEnd
   ...
   CollectionType string // "list", "sortedlist","set", "sortedset", "map", "stringmap", "sortedmap","sortedstringmap",""
   OrderAttrName string   // which primitive attribute of other is it ordered by when retrieving? "" if none

   OrderMethod *RMultiMethod
*/
/*

	      _,err = data.RT.CreateAttribute(typeName,
									 	 attributeTypeName,
										 attributeName,
										 minCard,
										 maxCard,   // Is the -1 meaning N respected in here???? TODO
										 collectionType,
				                         orderFuncOrAttrName,
				                         isAscending,
										 false,
										 g.Interp.Dispatcher())
		   if err != nil {
		      panic(err)
		   }

        }

        // Now ensure the persistence data model is created for the type.


		err = data.RT.DB().EnsureTypeTable(theNewType) 
		if err != nil {
		      panic(err)
		}
		
		// ... and for the type's attributes and relations

		err = data.RT.DB().EnsureAttributeAndRelationTables(theNewType) 
		if err != nil {
		      panic(err)
		}
    }

*/    











type NodePrinter struct {
}

func (p *NodePrinter) Visit(node ast.Node) (w ast.Visitor) {
	if node != nil {
	   fmt.Println("***",node)
    }
	return p
}

