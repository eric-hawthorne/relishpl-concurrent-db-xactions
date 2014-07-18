// Portions of the source code in this file 
// are Copyright 2009 The Go Authors. All rights reserved.
// Use of such source code is governed by a BSD-style
// license that can be found in the GO_LICENSE file.

// Modifications and additions which convert code to be part of a relish-language compiler 
// are Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of such source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// Package parser implements a parser for relish source files. Input may be
// provided in a variety of forms (see the various Parse* functions); the
// output is an abstract syntax tree (AST) representing the relish source. The
// parser is invoked through one of the Parse* functions.
//

// EGH It's now a relish parser
// Note: Error message generation from this parser should be improved in three ways:
// 1. If the correct construct is found but at wrong indentation, that fact should be stated.
// 2. If the correct indent level is present but an incorrect symbol, the
//    error should mention the expectation of whatever is legal there, even if the whole nested part is optional.
// 3. (Maybe) Parser should continue at next top level construct after error.


// This parser.go source code file is organized into the following sections:

// ----------------------------------------------------------------------------
// Parser Initialization
// ----------------------------------------------------------------------------
// Identifiers
// ----------------------------------------------------------------------------
// Common productions
// ----------------------------------------------------------------------------
// Blocks
// ----------------------------------------------------------------------------
// Expressions
// ----------------------------------------------------------------------------
// Statements
// ----------------------------------------------------------------------------
// Declarations
// ----------------------------------------------------------------------------
// Source files
// ----------------------------------------------------------------------------
// Parsing Grammar Support
// ----------------------------------------------------------------------------
// Scoping support  
// ----------------------------------------------------------------------------
// Parsing support





package parser

import (
	"fmt"
	"strings"
    "strconv"
	"relish/compiler/ast"
	"relish/compiler/scanner"
	"relish/compiler/token"
	. "relish/defs"
	"relish/dbg"
)

// The mode parameter to the Parse* functions is a set of flags (or 0).
// They control the amount of source code parsed and other optional
// parser functionality.
//
const (
	PackageClauseOnly uint = 1 << iota // parsing stops after package clause
	ImportsOnly                        // parsing stops after import declarations
	ParseComments                      // parse comments and add them to AST
	Trace                              // print a trace of parsed productions
	DeclarationErrors                  // report declaration errors
	SpuriousErrors                     // report all (not just the first) errors per line
)

// The parser structure holds the parser's internal state.
type parser struct {
	file *token.File
	scanner.ErrorVector
	scanner.Scanner

	// Tracing/debugging
	mode   uint // parsing mode
	trace  bool // == (mode & Trace != 0)
	indent uint // indentation used for tracing output


	// Comments
	comments    []*ast.CommentGroup
	leadComment *ast.CommentGroup // last lead comment
	lineComment *ast.CommentGroup // last line comment

	// Next token
	pos token.Pos   // token position
	tok token.Token // one token look-ahead
	lit string      // token literal

	// Ordinary identifier scopes
	pkgScope   *ast.Scope        // pkgScope.Outer == nil
	topScope   *ast.Scope        // top-most scope; may be pkgScope
	unresolved []*ast.Ident      // unresolved identifiers
	imports    []*ast.ImportSpec // list of imports
	
	packagePath string	// current package being parsed
	currentOriginAndArtifactName string
	
	// local aliases of imported packages
	importPackageAliases map[string] bool
	
	// map from local alias of imported package to full package name 
	importPackageAliasExpansions map[string] string
	
	// reserved words
	reservedWords map[string] bool	
	
	// Names of variables in the current method definition - anything else is a method name
	currentScopeVariables map[string] bool
	currentScopeVariableOffsets map[string] int
	currentScopeVariableOffset int
	currentScopeReturnArgOffset int	
	
	currentClosureMethodNum int
	
	// Names of variables in the enclosing method definition of a closure method decl 
	outerScopeVariables map[string] bool
	outerScopeVariableOffsets map[string] int
	outerScopeVariableOffset int
	outerScopeReturnArgOffset int	
	
	parsingClosure bool
	closureMethodName string  // package-unique name assigned to the current closure method declaration
	closureFreeVars []*ast.Ident
	closureFreeVarBindings []int  // list of enclosing-method var offsets of free vars in closure-method

    closureMethodDecls []*ast.MethodDeclaration

	// Label scope
	// (maintained by open/close LabelScope)
	labelScope  *ast.Scope     // label scope for current function
	targetStack [][]*ast.Ident // stack of unresolved labels

    // Probable cause of an error. Set speculatively. Used to create better error messages.
	probableErrorCause string
	probableCausePos token.Pos   // token position	
}



// ----------------------------------------------------------------------------
// Parser Initialization


// scannerMode returns the scanner mode bits given the parser's mode bits.
func scannerMode(mode uint) uint {
	var m uint = scanner.InsertSemis
	if mode&ParseComments != 0 {
		m |= scanner.ScanComments
	}
	return m
}

func (p *parser) init(fset *token.FileSet, filename string, src []byte, mode uint) {
	p.file = fset.AddFile(filename, fset.Base(), len(src))
	p.Scanner.Init(p.file, src, p, scannerMode(mode))

	p.mode = mode
	p.trace = mode&Trace != 0 // for convenience (p.trace is used frequently)

    p.ErrorVector.StopOnFirstError = true

	//p.next()


	// set up the pkgScope here (as opposed to in parseFile) because
	// there are other parser entry points (ParseExpr, etc.)
	p.openScope()
	p.pkgScope = p.topScope

	// for the same reason, set up a label scope
	p.openLabelScope()
	
    p.importPackageAliases = make(map[string] bool)	

    p.importPackageAliasExpansions = make(map[string] string)	

    p.reservedWords = map[string]bool{
       "if": true,
       "elif": true,
       "else": true,
       "while": true,
       "for": true,
       "in": true,
       "as": true,
       "of": true,       
       "continue": true,
       "break": true,
       "go": true,
       "true": true,
       "false": true,
       "nil": true,
       "func": true,
       "apply": true,
   }	
}


// ----------------------------------------------------------------------------
// Identifiers


/*
Parses a method name which is either of form fooBar or vehicles.fooBar.
If the former, creates an Ident whose Name is like "fooBar"
If the latter, creates an Ident whose Name is like "the/full/pkg/packg/name/fooBar"
Also accepts a (qualified or unqualified)type name.
If it sees a  type name, creates an Ident whose Name is like "the/full/pkg/packg/name/Car"
*/
func (p *parser) parseMethodName(allowImported bool, methodName **ast.Ident) bool {
   if p.trace {
      defer un(trace(p, "MethodName"))
   }   
   st := p.State()
   pos := p.Pos()

   var foundPackage bool
   var packageAlias *ast.Ident
   if allowImported {
	   foundPackage = p.parsePackageAlias(true,&packageAlias)
	   if foundPackage {
	      if ! p.Match1('.') {
		      return p.Fail(st)
	      }	
	   }
   }

   var kind token.Token
   var name string
   var foundMethodName,foundTypeName bool

   foundMethodName,name = p.ScanMethodName() 
   if  foundMethodName {
	   if p.importPackageAliases[name] {
	      return p.Fail(st)	
	   }
	   if name == "apply" {
	      kind = token.CLOSURE
	   } else if p.reservedWords[name] {
          return p.Fail(st)	
       } else {
	      kind = token.FUNC
       }
   }
   if ! foundMethodName {
      foundTypeName,name = p.ScanTypeName()    
      if ! foundTypeName {
	     return false
      }
      kind = token.TYPE
   }
 
// Are method names package specific or not??????

// Should only types here be packageAlias prefixed? How are package alias prefixes expanded here?  
/*
   if foundPackage {
      name = packageAlias.Name + "." + name	
   }
*/
   if foundPackage {
       name = p.importPackageAliasExpansions[packageAlias.Name] + "/" + name	
   } else if foundTypeName && ! BuiltinTypeName[name] {	  
       name = p.packagePath + name	     
   } else if strings.HasPrefix(name,"init") && len(name) > 5 && 'A' <= name[4] && name[4] <= 'Z' && !BuiltinTypeName[name[4:]] {
       name = p.packagePath + name	
   }


   *methodName = &ast.Ident{pos, name, nil, kind, -1}
   return true
}


/////
//
// onelineinput space onelinereturns
//             below(col) onelineOrIndentedReturn
// OR
// indentedinput below Indentedreturn
//
// TODO Have to account for case of no input arguments in which case
// space onelinereturns or
// below oneLineOrIndentedReturn
//////

func (p *parser) parseRelEndName(endName **ast.Ident) bool {
   if p.trace {
      defer un(trace(p, "RelEndName"))
   }   
   st := p.State()
   pos := p.Pos()

   found,name := p.ScanRelEndName() 
   if ! found {
      return false
   }
   if p.importPackageAliases[name] {
      return p.Fail(st)	
   }
   *endName = &ast.Ident{pos, name, nil, token.VAR, -1}
   return true
}

func (p *parser) parseVarName(varName **ast.Ident, mustBeDefined bool) bool {
   if p.trace {
      defer un(trace(p, "VarName"))
   }   
   st := p.State()
   pos := p.Pos()

   found,name := p.ScanVarName() 
   if ! found {
      return false
   }
   if p.importPackageAliases[name] {
      return p.Fail(st)	
   }

   if p.reservedWords[name] {
      return p.Fail(st)	
   }

   offset := -99

   foundFreeVar := false  

   if p.currentScopeVariables[name] { // set the Offset to the right local var or return arg
      offset = p.currentScopeVariableOffsets[name] 	
   } else if p.parsingClosure {
      if p.outerScopeVariables[name] {
	      foundFreeVar = true	
	      offset = p.outerScopeVariableOffsets[name] 	
	   	  dbg.Log(dbg.PARSE_,"------free var name %s --------",name)
     } else if mustBeDefined {
   	    dbg.Logln(dbg.PARSE_,"------while parsingClosure name not found as outer local var or outer return arg --------")
	    dbg.Logln(dbg.PARSE_,p.outerScopeVariables)
	    dbg.Logln(dbg.PARSE_,name)
	    dbg.Logln(dbg.PARSE_,"-----------------------------------------------------------------------------------------")	
        return p.Fail(st)	
     }
   } else if mustBeDefined {
   	    dbg.Logln(dbg.PARSE_,"------name not found as local var or return arg --------")
	    dbg.Logln(dbg.PARSE_,p.currentScopeVariables)
	    dbg.Logln(dbg.PARSE_,name)
	    dbg.Logln(dbg.PARSE_,"--------------------------------------------------------")
        return p.Fail(st)	
   }

   *varName = &ast.Ident{pos, name, nil, token.VAR, offset}
   
   if foundFreeVar {
      p.closureFreeVars = append(p.closureFreeVars, *varName)	
   }

   return true
}

func (p *parser) parseKeywordParameterName(paramName *string) bool {
   if p.trace {
      defer un(trace(p, "KeywordParameterName"))
   }   
   st := p.State()

   found,name := p.ScanVarName() 
   if ! found {
      return false
   }

   if p.reservedWords[name] {
      return p.Fail(st)	
   }

   *paramName = name

   return true
}


func (p *parser) parsePackageAlias(shouldExist bool, packageName **ast.Ident) bool {
   if p.trace {
      defer un(trace(p, "PackageAlias"))
   }   
   st := p.State()
   pos := p.Pos()

   found,name := p.ScanPackageAlias() 
   if ! found {
      return false
   }
   if shouldExist {
      if ! p.importPackageAliases[name] {
         return p.Fail(st)	
      }       	
   } else {
      if p.importPackageAliases[name] {
         return p.stop(fmt.Sprintf("Package identifier '%s' is being redefined. Already defined in this file",name))
      }	
   }

   *packageName = &ast.Ident{pos, name, nil, token.PACKAGE, -1}
   return true
}




/*
TODO: Note need to do the package thing for constants too.
*/
func (p *parser) parseTypeName(allowImported bool, typeName **ast.Ident) bool {
   if p.trace {
      defer un(trace(p, "TypeName"))
   }   
   st := p.State()
   pos := p.Pos()
   var foundPackage bool
   var packageAlias *ast.Ident
   if allowImported {
	   foundPackage = p.parsePackageAlias(true,&packageAlias)
	   if foundPackage {
	      if ! p.Match1('.') {
		      return p.Fail(st)
	      }	
	   }
   }

   found,name := p.ScanTypeName() 
   if ! found {
      return p.Fail(st)
   }

   if foundPackage {
       name = p.importPackageAliasExpansions[packageAlias.Name] + "/" + name	
   } else if ! BuiltinTypeName[name] {	  
       name = p.packagePath + name	     
   }

   *typeName = &ast.Ident{pos, name, nil, token.TYPE,-1}
   return true
}




func (p *parser) parseConstName(allowImported bool, constant **ast.Ident) bool {
   if p.trace {
      defer un(trace(p, "ConstName"))
   }   

   st := p.State()
   pos := p.Pos()
   var foundPackage bool
   var packageAlias *ast.Ident
   if allowImported {
	   foundPackage = p.parsePackageAlias(true,&packageAlias)
	   if foundPackage {
	      if ! p.Match1('.') {
		      return p.Fail(st)
	      }	
	   }
   }

   found,name := p.ScanConstName() 
   if ! found {
      return p.Fail(st)
   }

   if foundPackage {
      name = p.importPackageAliasExpansions[packageAlias.Name] + "/" + name	
   } else if allowImported {
	  // Should we expand (qualify) this at runtime instead, using the "current package"?
	  // upside is smaller executables
	  // downside is either string concetenation is needed or package should have hashtable of constants
	  // instead of runtime having hashtable. I like this solution actually, but it would mean
	  // we actually have to disassemble a composite constant name during generation. I like it though!!
	  // but it doesn't work. 
      name = p.packagePath + name	
   }
   *constant = &ast.Ident{pos, name, nil, token.CONST, -1}
   return true
}




// ----------------------------------------------------------------------------
// Common productions



/*
   A type specification.
   Can be used in several places:
   -In method input parameter and return value signatures
   -In the header portion of a type declaration
   -As an attribute type specification
   -As a relation-end type specification
   -As a type constructor
   -As the type constraint on a List, Set or Map construction literal.

   These different uses may place different limitations on the expression of the type specification.
   These limitations on type spec flexibility are specified as the boolean arguments to this method.

   For now, just parse a simple type name. or a collection type spec TODO handle parameterized types.
   Also, need to be able to handle T <: T1
   Re: parameterized type specs. Need to know when you can have a SomeType of T1 T2 vs a SomeType of Int Float
   canBeMaybe means the type spec could be like ?SomeType which means nil or a valid instance of the type.
   canBeParameterized means allow this to be a parameterized type
   canBeVariable means this is allowed to be a type variable
   canSpecifySuperTypes means this expression can include <: supTyp1 supTyp2 etc
   forceCollection means that this type is actually a collection even if no collection type expression occurs. 
   If forceCollection is true and there is no collection type expression, this type is a Set of the base type.
*/
func (p *parser) parseTypeSpec(canBeMultiLineOrMap bool,  // if false, single line and not map required.
	                           canBeMaybe bool,
	                           canBeParameterized bool, 
	                           canBeVariable bool, 
	                           canSpecifySuperTypes bool, 
	                           forceCollection bool,
	                           typeSpec **ast.TypeSpec) bool {
   
    st := p.State()
    nilAllowed := false
    nilElementsAllowed := false
    collectionTypeSpecFound := false
    if canBeMaybe {   
    	if p.Match1('?') {
           nilAllowed = true
    	}
    }

    var collectionTypeSpec *ast.CollectionTypeSpec

    // TODO   NOT HANDLING MAPS YET!!!!!!
    if p.parseCollectionTypeSpec(&collectionTypeSpec) {
       collectionTypeSpecFound = true
	   p.required(p.Space(),"a space then a type name")
	} else if forceCollection {
	   collectionTypeSpec = &ast.CollectionTypeSpec{token.SET,p.Pos(),p.Pos()+1,false,false,""}
    }

	var typeName *ast.Ident

// egh 2014	
//	if collectionTypeSpecFound {
//    	if p.Match1('?') {
//           nilElementsAllowed = true
//    	}		
//	}

	var params []*ast.TypeSpec = nil
	
	typeCol := p.Col()
	
	if collectionTypeSpecFound {
	   var elementTypeSpec *ast.TypeSpec		
       if p.parseTypeSpec(canBeMultiLineOrMap, true,true,false,false,false,&elementTypeSpec) {
		   params = append(params, elementTypeSpec)	       
       } else {
	      return p.Fail(st)       	
       }
	} else if ! p.parseTypeName(true, &typeName) {   // TODO This needs to be another full typespec here!!!!!!!!
	   return p.Fail(st)
	}
	
	// Have to look for TypeName=>TypeName or 
	// TypeName
	// =>
	// TypeName
	//
	// IDEA: Maybe the key type goes as a sub part of the collection type!!!!
	// Do this when changing the collection type from token.SET to token.MAP
	
	// Look for evidence that it is a Map type as opposed to a Set.
	
	if canBeMultiLineOrMap && collectionTypeSpec != nil && collectionTypeSpec.Kind == token.SET {
	    var valTypeSpec *ast.TypeSpec		
	    knownToBeMap := false
		if p.Match2(' ','>') {
	       p.required(p.Space(),"a space followed by the map value-type specification")
           p.required(p.parseTypeSpec(true,true,true,false,false,false,&valTypeSpec),"the map value-type specification")		
		   knownToBeMap = true

		} else if p.Below(typeCol) {
           p.required(p.Match1('>'),"> T : the map's value-type")
	
		   foundValType := false
		   if p.Space() {
			   if p.parseTypeSpec(true, true,true,false,false,false,&valTypeSpec) {
			      foundValType = true
			   } else {
			      p.required(p.Space() || (p.Ch()  == '\n'),"the map value-type specification")
			   }
		    }
			if ! foundValType {
			   p.required(p.Indent(typeCol) && p.parseTypeSpec(true, true,true,false,false,false,&valTypeSpec),
			              "the map value-type specification, after a space or indented on next line")                                        
			}
		    knownToBeMap = true			
		}
		
		if knownToBeMap {
		   collectionTypeSpec.Kind = token.MAP			
		   params = append(params, valTypeSpec)	
		}	
	}
	
	
    *typeSpec = &ast.TypeSpec{CollectionSpec: collectionTypeSpec, Name: typeName, Params: params, NilAllowed: nilAllowed, NilElementsAllowed: nilElementsAllowed}
    return true
}





/*
Note: Type parameters are not implemented in relish type checking yet.    
*/
func (p *parser) parseTypeParameters(col int, typeSpec *ast.TypeSpec) bool {
	if p.trace {
	   defer un(trace(p, "TypeParameters"))
	}	
		
	if ! p.Match(" of") {
		return false
	}
	
    p.required(
               p.parseOneLineTypeParameters(typeSpec) ||
               p.parseVerticalTypeParameters(col,typeSpec),
               "a type parameter e.g. T")
                 	
	return true
}

func (p *parser) parseOneLineTypeParameters(typeSpec *ast.TypeSpec) bool {
	if p.trace {
	   defer un(trace(p, "OneLineTypeParameters"))
	}	
	
    st := p.State()
	
	if ! p.Space() {
		return false
	}
	
	var paramTypeSpec *ast.TypeSpec
	
    if ! p.parseTypeParameter(&paramTypeSpec) {
       return p.Fail(st)	
    }
	
    // translate
    typeSpec.Params = append(typeSpec.Params, paramTypeSpec)

   st2 := p.State()
   for p.Space() {
       if p.parseTypeParameter(&paramTypeSpec) {
         // translate
         typeSpec.Params = append(typeSpec.Params, paramTypeSpec)	
	   } else {
           p.Fail(st2)
           break	
       }	
       st2 = p.State()
   }
	
	// TODO Create the ast nodes.
	
	return true
}

func (p *parser) parseTypeParameter(typeSpec **ast.TypeSpec) bool {
   return p.parseTypeSpec(false, false,true, true, true, false, typeSpec)	
}




func (p *parser) parseVerticalTypeParameters(col int, typeSpec *ast.TypeSpec) bool {
	if p.trace {
	   defer un(trace(p, "VerticalTypeParameters"))
	}	

    st := p.State()
	
	if ! p.Indent(col) {
	   return false
	}

	var paramTypeSpec *ast.TypeSpec
    if ! p.parseTypeParameter(&paramTypeSpec) {
       return p.Fail(st)	
    }	

    // translate
    typeSpec.Params = append(typeSpec.Params, paramTypeSpec)

   for p.Indent(col) {
       p.required(p.parseTypeParameter(&paramTypeSpec),"a type parameter e.g. T")	

       // translate
       typeSpec.Params = append(typeSpec.Params, paramTypeSpec)
   }	
		
   // TODO Create the ast nodes.
	
   return true
}





// ----------------------------------------------------------------------------
// Blocks



/*
A list of statements forming the body (implementation) of a method.
*/
func (p *parser) parseMethodBody(col int,methodDecl *ast.MethodDeclaration) bool {
    if p.trace {
       defer un(trace(p, "MethodBody"))
    }	

    // parse
    st := p.State()
    if ! p.BlanksAndIndent(col,true) {
       p.checkForIndentError()	
	   return false
    }

    blockPos := p.Pos()

    var stmts []ast.Stmt

    nReturnVals := methodDecl.NumAnonymousReturnVals()
    noResults := (methodDecl.NumReturnVals() == 0)

    if ! p.parseMethodBodyStatement(&stmts, nReturnVals, noResults) {
	   return p.Fail(st)
    }
    for p.BlanksAndIndent(col, true) {
        p.required(p.parseMethodBodyStatement(&stmts, nReturnVals, noResults),"a statement")
    }

    for p.BlanksThenIndentedLineComment(col) {}  // Gobble trailing indented line comments

    p.checkForIndentError()	

    methodDecl.Body = &ast.BlockStatement{blockPos,stmts}

    // TODO Need to check for return argument assignments, compatibility etc

    return true   
}






// ----------------------------------------------------------------------------
// Expressions




func (p *parser) parseExpression(x *ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "Expression"))
    }
    return p.parseIndentedExpression(x) || p.parseOneLineExpression(x, false, false) 
}

func (p *parser) parseOneLineExpression(x *ast.Expr, isOneOfMultiple bool, isInsideBrackets bool) bool {
    if p.trace {
       defer un(trace(p, "OneLineExpression"))
    }

    st := p.State()

    if p.parseOneLineLiteral(x) {
	   return true
    }
/*
    var constant *ast.Ident
    if p.parseConstName(true,&constant) {
	   // translate
	   *x = constant	
	   return true
    }
    
   
    if p.parseOneLineVariableReference(false, false, true, x) {
	   return true
    }
*/    
    if p.parseOneLineVariableOrConstReference(false, false, false, true, x) {
      return true
    }    
    

    // if isInsideBrackets can't have a method call of any kind nor a constructor invocation of any kind. 

    if ! isInsideBrackets {
	
	   if isOneOfMultiple { // need a bracketed method call
	       var mcs *ast.MethodCall
           var lcs *ast.ListConstruction         
           var mpcs *ast.MapConstruction 	 
           var scs *ast.SetConstruction      
	       if p.Match1('(') {
	
	            p.required(p.parseOneLineMethodCall(&mcs,true) || p.parseOneLineListConstruction(&lcs,true) || p.parseOneLineMapOrSetConstruction(&mpcs,&scs,true), "a subroutine call or constructor invocation") 
		
		        p.required(p.Match1(')'),"closing bracket )")
		
		        // translate
	            if mcs != nil {
	              *x = mcs
	            } else if lcs != nil {
			          *x = lcs
	            } else if mpcs != nil {
			          *x = mpcs
	            } else {
		              *x = scs
	            }

		        return true
	       }
     } else { // not one of multiple
	       var mcs *ast.MethodCall
	       if p.parseOneLineMethodCall(&mcs,false) {
		      // translate
		      *x = mcs
		      return true
	       }
           var lcs *ast.ListConstruction
           if p.parseOneLineListConstruction(&lcs,false) {
              // translate
              *x = lcs
              return true
           }    
           var mpcs *ast.MapConstruction
           var scs *ast.SetConstruction
           if p.parseOneLineMapOrSetConstruction(&mpcs,&scs,false) {
              // translate
              if mpcs != nil {
                  *x = mpcs
			  } else {
				  *x = scs
			  }
              return true
           }
       }
    }

    // Generate a guess of the programmer's intention for a better error message.

	if p.Space() && p.parseOneLineExpression(x, isOneOfMultiple, isInsideBrackets)  {
	  	p.setProbableCause("Extra space before expression.")
	}
	return p.Fail(st)
}


/* after a space */
func (p *parser) parseSingleOneLineExpression(xs *[]ast.Expr, isInsideBrackets bool) bool {
    if p.trace {
       defer un(trace(p, "SingleOneLineExpression"))
    }

    isOneOfMultiple := false

    var x ast.Expr
    if ! p.parseOneLineExpression(&x, isOneOfMultiple, isInsideBrackets) {
	   return false
    }

    // translate
    *xs = nil
    *xs = append(*xs,x)
    return true
}


func (p *parser) parseIndentedExpression(x *ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "IndentedExpression"))
    }

    st := p.State()

    if p.parseMultilineLiteral(x) {
       return true	
    }

    if p.parseIndentedVariableOrConstReference(false, false, true, x) {  
	   return true
    }

    var clos *ast.Closure
    if p.parseClosure(&clos) {
	   *x = clos	
	   return true
    }


    // Need an apply expression - apply myClosure arg arg arg

    var mcs *ast.MethodCall
    if p.parseIndentedMethodCall(&mcs) {
	   // translate
	   *x = mcs	
	   return true
    }

    var lcs *ast.ListConstruction
    if p.parseIndentedListConstruction(&lcs) {
     // translate
     *x = lcs 
     return true
    }

    var mpcs *ast.MapConstruction
    var scs *ast.SetConstruction
    if p.parseIndentedMapOrSetConstruction(&mpcs,&scs) {
     //translate
     if mpcs != nil {
        *x = mpcs
     } else {
        *x = scs
     } 
     return true     
    }    


    // Generate a guess of the programmer's intention for a better error message.

	if p.Space() && p.parseIndentedExpression(x)  {
	  	p.setProbableCause("Extra space before expression.")
	}

    return p.Fail(st)
}

/* after a space 
   at least two expressions on same line
*/
func (p *parser) parseMultipleOneLineExpressions(xs *[]ast.Expr, isInsideBrackets bool) bool {
    if p.trace {
       defer un(trace(p, "MultipleOneLineExpressions"))
    }
    
    st := p.State()

    isOneOfMultiple := true

    var x ast.Expr
    if ! p.parseOneLineExpression(&x, isOneOfMultiple, isInsideBrackets) { 
	   return false
    }

    if ! p.Space() {
	   return p.Fail(st)	
    }

    var x2 ast.Expr
    if ! p.parseOneLineExpression(&x2, isOneOfMultiple, isInsideBrackets) { 
	   return p.Fail(st)
    }

    // translate
    *xs = nil
    *xs = append(*xs,x)

    // translate
    *xs = append(*xs,x2)

    st2 := p.State()
    for ; p.Space() ; {
	   if ! p.parseOneLineExpression(&x, isOneOfMultiple, isInsideBrackets) {
		   break	
	   }
	
	   // translate
       *xs = append(*xs,x)	
       
       st2 = p.State()	
    }
    p.Fail(st2)
    return true
}





/* 
*/
func (p *parser) parseOneLineMapEntryExpressions(keys *[]ast.Expr, vals *[]ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "OneLineMapEntryExpressions"))
    }
    
    st := p.State()

    isOneOfMultiple := false
    isInsideBrackets := false

    var key ast.Expr
    if ! p.parseOneLineExpression(&key, isOneOfMultiple, isInsideBrackets) { 
	   return false
    }

    if ! p.Match2('=','>') {
	    return p.Fail(st)
    }

    var val ast.Expr
    p.required(p.parseOneLineExpression(&val, isOneOfMultiple, isInsideBrackets),"a value expression to be put in the map") 

    // translate
    *keys = nil
    *keys = append(*keys,key)
    *vals = nil
    *vals = append(*vals,val)


    st2 := p.State()
    for ; p.Space() ; {
	   if ! p.parseOneLineExpression(&key, isOneOfMultiple, isInsideBrackets) {
		   break	
	   }
	
	   // translate
       *keys = append(*keys,key)	
	
	   p.required(p.Match2('=','>'), "=> and a value expression to be put in the map") 

	   var val ast.Expr
	   p.required(p.parseOneLineExpression(&val, isOneOfMultiple, isInsideBrackets),"a value expression to be put in the map")
	
	   // translate
       *vals = append(*vals,val)	
       
       st2 = p.State()	
    }
    p.Fail(st2)
    return true
}



/*
May be only one expression, but must be indented below the current line.
*/
func (p *parser) parseIndentedMapEntryExpressions(col int, keys *[]ast.Expr, vals *[]ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "IndentedMapEntryExpressions"))
    }

    st := p.State()

    isOneOfMultiple := false
    isInsideBrackets := false

    if ! p.Indent(col) {
     return false
    }

    var arrowCol int

    var key ast.Expr
    if ! p.parseOneLineExpression(&key, isOneOfMultiple, isInsideBrackets) { 
     return p.Fail(st)
    }
    if ! p.Space() {
      p.required( ! p.Match2('=','>'),"one or more spaces followed by =>")
      return p.Fail(st)
    }
    for p.Space() {  // continue til not a space
    }

    arrowCol = p.Col()

    if ! p.Match2('=','>') {
      return p.Fail(st)
    }

    var val ast.Expr
    p.required(p.Space() && p.parseExpression(&val),"a space then a value expression to be put in the map") 

    // translate
    *keys = nil
    *keys = append(*keys,key)
    *vals = nil
    *vals = append(*vals,val)

    for  p.Indent(col)  {
       p.required(p.parseOneLineExpression(&key, isOneOfMultiple, isInsideBrackets),"a key expression")
  
       // translate
       *keys = append(*keys,key)  
  
       for p.Space() {  // continue til not a space
       }  
       arrowFoundAtCol := p.Col()
       if arrowFoundAtCol != arrowCol {
          if p.Ch() == '=' {
             p.stop("All =>'s must be lined up vertically below each other")            
          } else {
             p.required(false, "=> and a value expression to be put in the map")             
          }

       }
       p.required(p.Match2('=','>'), "=> and a value expression to be put in the map") 

       var val ast.Expr
       p.required(p.Space() && p.parseExpression(&val),"a space then a value expression to be put in the map") 
  
       // translate
       *vals = append(*vals,val)  
    }
    return true    
}




/*
May be only one expression, but must be indented below the current line.
*/
func (p *parser) parseIndentedExpressions(col int, xs *[]ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "IndentedExpressions"))
    }

    st := p.State()    
    if ! p.Indent(col) {

    	// Look for indentation errors as probable causes of not finding this
	    if p.Indent(col-1) {
           if ! (p.Match1(']') || p.Match1('}')) {
              p.setProbableCause("Must indent with 3 spaces. Only indented 2 spaces.")
           }
	    } else if p.Indent(col+1) {
	       p.setProbableCause("Must indent with 3 spaces. Indented with 4 spaces.")  	        
	    } 

	   return p.Fail(st)
    }
    var x ast.Expr
    p.required(p.parseExpression(&x),"an expression")

    // translate
    *xs = nil
    *xs = append(*xs,x)

    for ; p.Indent(col) ; {
	   p.required(p.parseExpression(&x),"an expression")
	
       // translate
       *xs = append(*xs,x)	
    }

    st2 := p.State()

    // Look for indentation errors as probable causes of not finding this    
    if p.Indent(col-1) {
       if ! (p.Match1(']') || p.Match1('}')) {
          p.setProbableCause("Must indent with 3 spaces. Only indented 2 spaces.")
       }
       p.Fail(st2)        

    } else if p.Indent(col+1) {
       p.setProbableCause("Must indent with 3 spaces. Indented with 4 spaces.")  	
       p.Fail(st2)         
    } 

    return true
}



/*
May be only one expression or keyword param assignment, but must be indented below the current line.
If there is something indented but it is neither an expression nor keyword param assignment, is an error.
*/
func (p *parser) parseIndentedExpressionsOrKeywordParamAssignments(col int, xs *[]ast.Expr, firstAssignmentPos *int32, kws map[string]ast.Expr) bool {
	
    if p.trace {
       defer un(trace(p, "IndentedExpressionsOrKeywordParamAssignments"))
    }

    *firstAssignmentPos = -1

    var i int32 = 0
    var x ast.Expr
    var key string

    foundIndent := false
    foundKeywords := false
    foundVariadic := false

    // translate
    *xs = nil

    for p.Indent(col) {
	
	   foundIndent = true
	
       if p.parseKeywordParameterAssignmentStatement(&key, &x) {
	       if foundVariadic {
              p.stop("keyword parameter assignment statement cannot appear after variadic argument") 	 		
	       }
		   foundKeywords = true
	       if *firstAssignmentPos == -1 {	
	          *firstAssignmentPos = i
	       }
          
          // translate	
	      kws[key] = x
		
	   } else if p.parseExpression(&x) {
		   i++
		   if foundKeywords {
		      foundVariadic = true	
	       }
	
	       // translate
	       *xs = append(*xs,x)	
	   } else {
		
           p.stop("Expecting an expression or a keyword parameter assignment statement") 	          
       
       }
    }

    p.checkForIndentError()	

    return foundIndent
}


/*
Requires at least two expressions, one below the other.
*/
func (p *parser) parseVerticalExpressions(col int, xs *[]ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "VerticalExpressions"))
    }
    st := p.State()

    var x ast.Expr

    if ! p.parseExpression(&x) {
	   return false;
    }
    if ! p.Below(col) {
	   return p.Fail(st)
    }

    // translate
    *xs = nil
    *xs = append(*xs,x)

    p.required(p.parseExpression(&x),"an expression")

    // translate
    *xs = append(*xs,x)

    for ; p.Below(col) ; { 
	   p.required(p.parseExpression(&x),"an expression")
	
       // translate
       *xs = append(*xs,x)	
	}
	return true
}




/*
Requires at least two expressions, one below the other.

TODO NOTE THIS IS NOT RIGHT!!! FIX IT LIKE parseIndentedExpressionsOrKeywordParamAssignments so it always looks for a keyword
param assignment on each row before looking for an expression!!!!!!!!
*/
func (p *parser) parseVerticalExpressionsOrKeywordParamAssignments(col int, xs *[]ast.Expr, firstAssignmentPos *int32, kws map[string]ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "VerticalExpressionsOrKeywordParamAssignments"))
    }
    st := p.State()

    *firstAssignmentPos = -1

    var i int32 = 0
    var x ast.Expr
    var key string

    foundPositional := false
    foundKeywords := false
    foundTwoRows := false
    foundVariadic := false

    // translate
    *xs = nil

    if p.parseExpression(&x) {
	   i++
	   foundPositional = true
	
       // translate
       *xs = append(*xs,x)	

    } else if p.parseKeywordParameterAssignmentStatement(&key, &x) {
	   foundKeywords = true
       *firstAssignmentPos = i	
	
       // translate	
	   kws[key] = x	
	
    } else {
	   return false
    }
	
	
    st2 := p.State()

    for p.Below(col) {
	
	   if ! p.parseExpression(&x) {
		  p.Fail(st2)
	      break	
	   }
	   i++
       st2 = p.State()	
	   foundPositional = true
	   if foundKeywords {
	      foundVariadic = true	
	   }	
	   foundTwoRows = true
	   
       // translate
       *xs = append(*xs,x)	
    }


    if ! foundVariadic {  // There has been no positional arg after a keyword arg yet

	    for p.Below(col) {
	
	       if ! p.parseKeywordParameterAssignmentStatement(&key, &x) {
		       p.Fail(st2)
		       break
		   }
           st2 = p.State()		
		   foundKeywords = true
	       foundTwoRows = true		
	       if *firstAssignmentPos == -1 {	
	          *firstAssignmentPos = i	
           }
	       // translate	
		   kws[key] = x	
	    }
    }

    if ! foundTwoRows {
	   return p.Fail(st)
    }

    if ! foundVariadic {
	    for p.Below(col) {
	
		   p.required(p.parseExpression(&x),"an expression")
	
	       // translate
	       *xs = append(*xs,x)	
	    }
    }

    if ! foundPositional && ! foundKeywords {
        p.stop("Expecting an expression or parameter assignment statement") 	          
    }

    return true
}








/*
Requires at least one multi-line expression, possibly more, one below the other.
*/
func (p *parser) parseLenientVerticalExpressions(col int, xs *[]ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "LenientVerticalExpressions"))
    }
    st := p.State()

    var x ast.Expr

    multiLines := false

    if ! p.parseExpression(&x) {
	   return false;
    }
    st2 := p.State()
    if p.isLower(st,st2) {
	   multiLines = true
    }

    // translate
    *xs = nil
    *xs = append(*xs,x)

    for ; p.Below(col) ; { 
	   p.required(p.parseExpression(&x),"an expression")
	   multiLines = true	
	
       // translate
       *xs = append(*xs,x)	
	}
	
	if ! multiLines {
		return p.Fail(st)
	}
	
	return true
}



/*
Requires that the expression or multiple vertical aligned expressions take up at least two lines of the file vertically.

TODO NOTE THIS IS NOT RIGHT!!! FIX IT LIKE parseIndentedExpressionsOrKeywordParamAssignments so it always looks for a keyword
param assignment on each row before looking for an expression!!!!!!!!
*/
func (p *parser) parseLenientVerticalExpressionsOrKeywordParamAssignments(col int, xs *[]ast.Expr, firstAssignmentPos *int32, kws map[string]ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "LenientVerticalExpressionsOrKeywordParamAssignments"))
    }
    st := p.State()

    *firstAssignmentPos = -1

    var i int32 = 0
    var x ast.Expr
    var key string

    multiLines := false

    foundPositional := false
    foundKeywords := false
    foundTwoRows := false
    foundVariadic := false

    // translate
    *xs = nil

    if p.parseExpression(&x) {
     i++
     foundPositional = true
  
       // translate
       *xs = append(*xs,x)  

    } else if p.parseKeywordParameterAssignmentStatement(&key, &x) {
     foundKeywords = true
       *firstAssignmentPos = i  
  
       // translate 
     kws[key] = x 
  
    } else {
     return false
    }
  
  
    st2 := p.State()
    if p.isLower(st,st2) {
     multiLines = true
    }    

    for p.Below(col) {
     if ! p.parseExpression(&x) {
      p.Fail(st2)
        break 
     }
     i++
       st2 = p.State()  
     foundPositional = true
     if foundKeywords {
        foundVariadic = true  
     }  
     foundTwoRows = true
     
       // translate
       *xs = append(*xs,x)  
    }


    if ! foundVariadic {  // There has been no positional arg after a keyword arg yet

      for p.Below(col) {
  
         if ! p.parseKeywordParameterAssignmentStatement(&key, &x) {
           p.Fail(st2)
           break
       }
           st2 = p.State()    
       foundKeywords = true
         foundTwoRows = true    
         if *firstAssignmentPos == -1 { 
            *firstAssignmentPos = i 
           }
         // translate 
       kws[key] = x 
      }
    }

    if ! (foundTwoRows || multiLines) {
     return p.Fail(st)
    }

    if ! foundVariadic {
      for p.Below(col) {
  
       p.required(p.parseExpression(&x),"an expression")
  
         // translate
         *xs = append(*xs,x)  
      }
    }

    if ! foundPositional && ! foundKeywords {
        p.stop("Expecting an expression or parameter assignment statement")             
    }

    return true
}


/*
TODO add more complex still init-time evaluable expressions involving function calls on literals and other constants
*/
func (p *parser) parseConstExpr(x *ast.Expr) bool {
	if p.trace {
       defer un(trace(p, "ConstExpr"))
    }
    
    return p.parseLiteral(x)
}



func (p *parser) parseLiteral(x *ast.Expr) bool {
	if p.trace {
       defer un(trace(p, "Literal"))
    }
    return p.parseMultilineLiteral(x) || p.parseOneLineLiteral(x) 
}

func (p *parser) parseOneLineLiteral(x *ast.Expr) bool {
	if p.trace {
       defer un(trace(p, "OneLineLiteral"))
    }
    return p.parseNumberLiteral(x) || p.parseStringLiteral(x) || p.parseBooleanLiteral(x) || p.parseNilLiteral(x) || p.parseRawStringLiteral(x, false)
}

func (p *parser) parseNumberLiteral(x *ast.Expr) bool {
	if p.trace {
       defer un(trace(p, "NumberLiteral"))
    }
    st := p.State()
    pos := p.Pos()
    negated := false
    if p.Match1('-') {
	   negated = true
    }  
    found, tok, lit := p.ScanNumber()
    if ! found {
	   if negated {
	      return p.Fail(st)	
	   }
	   return false // Look for String literals or Boolean literals
    }
    if negated {
	   lit = "-" + lit
    }
    dbg.Log(dbg.PARSE_,"%s '%s'\n",tok,lit)

    *x = &ast.BasicLit{pos,tok,lit}
    return true
}

func (p *parser) parseBooleanLiteral(x *ast.Expr) bool {
	if p.trace {
       defer un(trace(p, "BooleanLiteral"))
    }
    pos := p.Pos()
    var lit string
    if p.MatchWord("true") {
	   lit = "true"
    } else if p.MatchWord("false") {
	   lit = "false"
    } else {
	   return false
    }

    tok := token.BOOL
    dbg.Log(dbg.PARSE_,"%s '%s'\n",tok,lit)

    *x = &ast.BasicLit{pos,tok,lit}
    return true
}

func (p *parser) parseNilLiteral(x *ast.Expr) bool {
	if p.trace {
       defer un(trace(p, "NilLiteral"))
    }
    pos := p.Pos()
    var lit string
    if p.MatchWord("nil") {
	      lit = "nil"
    } else {
	      return false
    }
    tok := token.NIL
    dbg.Log(dbg.PARSE_,"%s '%s'\n",tok,lit)

    *x = &ast.BasicLit{pos,tok,lit}
    return true
}


func (p *parser) parseStringLiteral(x *ast.Expr) bool {
	if p.trace {
       defer un(trace(p, "StringLiteral"))
    }
    pos := p.Pos()
    found, lit := p.ScanString()
    if ! found {
	   return false 
    }
    dbg.Log(dbg.PARSE_,"String literal \"%s\"\n",lit)

    var err error
    lit,err = strconv.Unquote(`"` + lit + `"`)
    if err != nil {
       p.stop(err.Error())
    }

    *x = &ast.BasicLit{pos,token.STRING,lit}
    return true
}


func (p *parser) parseMultilineLiteral(x *ast.Expr) bool {
	if p.trace {
       defer un(trace(p, "MultilineLiteral"))
    }
    return  p.parseMultilineStringLiteral(x) || p.parseRawStringLiteral(x, true)
}

func (p *parser) parseMultilineStringLiteral(x *ast.Expr) bool {
	if p.trace {
       defer un(trace(p, "MultilineStringLiteral"))
    }
    pos := p.Pos()
	if ! ( p.Match(`"""`) &&
	       p.required(p.BlankToEOL(),`nothing on line after """`) ) {
	    return false
	}
// Formerly had another """ at column 1 to start the multi-line string literal.	
//    p.required(p.Below(1) && p.Match(`"""`), `""" at beginning of line`)
//    p.required(p.BlankToEOL(),`nothing on line after """`)
    var ch rune
    ch = p.Ch()	
    if ch < 0 {
	   p.stop(`Multiline string not terminated. Expecting terminating """ at column 1`)
	   return false
    }
	p.Next()
	st2 := p.State()
	startOffset := st2.Offset
   	
    
    found,contentEndOffset := p.ConsumeTilMatchAtColumn(`"""`,1)
    if ! found {
	  p.stop(`Multiline string not terminated. Expecting terminating """ at column 1`)	
	  return false
	}
    p.required(p.BlankToEOL(),`nothing on line after """`) 

    lit := p.Substring(startOffset, contentEndOffset) 
      	
    dbg.Log(dbg.PARSE_,"String literal \"%s\"\n",lit)

    *x = &ast.BasicLit{pos,token.STRING,lit}
    return true
}


func (p *parser) parseRawStringLiteral(x *ast.Expr, multiLine bool) bool {
  if p.trace {
       defer un(trace(p, "RawStringLiteral"))
    }
    pos := p.Pos()
    st := p.State()
  if ! p.Match("```") {
      return false
  }
// Formerly had another """ at column 1 to start the multi-line string literal. 
//    p.required(p.Below(1) && p.Match(`"""`), `""" at beginning of line`)
//    p.required(p.BlankToEOL(),`nothing on line after """`)
    var ch rune
    ch = p.Ch() 
    if ch < 0 {
     p.stop("Raw string not terminated. Expecting terminating ```")
     return false
    } else if ch == '\n' {  // Special case for multi-line raw strings. Do not include the first \n
      p.Next()     
    } 
    st2 := p.State()
    
    startOffset := st2.Offset
    
    
    found,isMultiLine,contentEndOffset := p.ConsumeTilMatch("```")
    if ! found {
       p.stop("Raw string not terminated. Expecting terminating ```")  
       return false
    }
    if isMultiLine != multiLine {
       return p.Fail(st)
    }

    lit := p.Substring(startOffset, contentEndOffset) 
        
    dbg.Log(dbg.PARSE_,"String literal \"%s\"\n",lit)

    *x = &ast.BasicLit{pos,token.STRING,lit}
    return true
}



/*
Note. It is difficult to tell the difference between a variable name
and a method name, thus it is difficult to know whether
we are starting a prefix method call, or have a list of variables.

One clue is that we have to keep track of all local-var and parameter names when 
in a method body, and only those could possibly be (beginning) variable references.

Note that the "method" names we encounter in dot notation have no arguments, in the usual sense,
so they cannot be the start of a method call.

NEED TO PASS AN IDENT OUT OF THIS ESPECIAlLY IF A NEW VARIABLE
BECAUSE in parseLHSList we are not collecting the variables correctly yet!!!!!!!!!
*/
func (p *parser) parseOneLineVariableOrConstReference(mustBeVar bool, mustBeLocalVar bool, mustBeAssignable bool, mustBeDefined bool, expr *ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "OneLineVariableOrConstReference"))
    }
    
    var varOrConstName *ast.Ident
    var isConst bool
    
    if ! mustBeVar {
       isConst = p.parseConstName(true,&varOrConstName) 
    }
    if ! isConst {
       if ! p.parseVarName(&varOrConstName, mustBeDefined) {
	       return false
       }
    } 

    var x ast.Expr = varOrConstName

    for p.parseOneLineIndexOrSliceExpression(x,&x) {	
	    p.required(! mustBeLocalVar, "a simple local variable name, not an indexed expression")	
    }

    if p.Match("...") {
	
    }

    for p.Match1('.') {
	    p.required(! mustBeLocalVar, "a simple local variable name, not an object attribute selection expression")
	    p.required(p.parseOneLineVariableReference1(x,&x),"a variable or accessor-method name")
    }

	*expr = x

    // Check to make sure the reference is an assignable storage location if needed.
    return true
}


func (p *parser) parseOneLineVariableReference1(exprSoFar ast.Expr, expr *ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "OneLineVariableReference1"))
    }
    var varName *ast.Ident
    if ! p.parseVarName(&varName,false) {
	    return false
    }

    // translate
    var x ast.Expr = &ast.SelectorExpr{exprSoFar,varName}

    for p.parseOneLineIndexOrSliceExpression(x,&x) {	
    }
    for p.Match1('.') {
	    p.required(p.parseOneLineVariableReference1(x,&x),"a variable or accessor-method name")
    }

	*expr = x
	
    return true
}


func (p *parser) parseIndentedVariableOrConstReference(mustBeVar bool, mustBeAssignable bool,mustBeDefined bool, expr *ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "IndentedVariableOrConstReference"))
    }

    var varOrConstName *ast.Ident
    var isConst bool

    if ! mustBeVar {
       isConst = p.parseConstName(true,&varOrConstName) 
    }
    if ! isConst {
       if ! p.parseVarName(&varOrConstName, mustBeDefined) {
          return false
       }
    } 

    var x ast.Expr = varOrConstName


    for p.parseIndexOrSliceExpression(x,&x) {	
    }
    for p.Match1('.') {
	    p.required(p.parseIndentedVariableReference1(x,&x),"a variable or accessor-method name")
    }
    // Check to make sure the reference is an assignable storage location if needed.

	*expr = x
	
    return true
}


func (p *parser) parseIndentedVariableReference1(exprSoFar ast.Expr, expr *ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "IndentedVariableReference1"))
    }
    col := p.Col()
    var varName *ast.Ident
    if ! p.parseVarName(&varName,false) {
	    return false
    }

    // translate
    var x ast.Expr = &ast.SelectorExpr{exprSoFar,varName}

    for p.parseIndexOrSliceExpression(x,&x) {	
    }
    for p.Match1('.') {
	    p.optional(p.Indent(col))
	    p.required(p.parseIndentedVariableReference1(x,&x),"a variable or accessor-method name")
    }

	*expr = x
	
    return true
}

func (p *parser) parseIndexExpression(exprSoFar ast.Expr, x *ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "IndexExpression"))
    }
    return p.parseIndentedIndexExpression(exprSoFar,x) || p.parseOneLineIndexExpression(exprSoFar,x)
}

func (p *parser) parseIndexOrSliceExpression(exprSoFar ast.Expr, x *ast.Expr) bool {
	if p.trace {
	   defer un(trace(p, "IndexOrSliceExpression"))
	}	
	return p.parseIndentedIndexOrSliceExpression(exprSoFar, x) || p.parseOneLineIndexOrSliceExpression(exprSoFar, x)
}


func (p *parser) parseIndentedIndexExpression(exprSoFar ast.Expr, x *ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "IndentedIndexExpression"))
    }
    st := p.State()
    col := p.Col()
    lBracketPos := p.Pos()

    if ! p.Match1('[') {
       return false
    }

    askWhether := false

    assertExists := p.Match1('!')
    if ! assertExists {
       if p.Match1('?') {
          askWhether = true
       }
    }

    if ! p.Indent(col) {
	   return p.Fail(st)
	}
	if ! p.parseExpression(x) {
		return p.Fail(st)
	}
    p.required(p.Below(col),"] below [")

    rBracketPos := p.Pos()
    p.required(p.Match1(']'),"] below [") 

    *x = &ast.IndexExpr{exprSoFar,lBracketPos, *x, rBracketPos, assertExists, askWhether}

    return true
}

func (p *parser) parseOneLineIndexExpression(exprSoFar ast.Expr, x *ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "OneLineIndexExpression"))
    }
    lBracketPos := p.Pos()

    if ! p.Match1('[') {
	   return false
    }

    askWhether := false

    assertExists := p.Match1('!')
    if ! assertExists {
       if p.Match1('?') {
          askWhether = true
       }
    }
    if assertExists || askWhether {
        p.required(p.Space(),"a space, followed by the index expression")
    }

    p.required(p.parseOneLineExpression(x,false,false),"an expression")
    rBracketPos := p.Pos()
    p.required(p.Match1(']'),"']'")

    *x = &ast.IndexExpr{exprSoFar,lBracketPos, *x, rBracketPos, assertExists, askWhether}

    return true
}

func (p *parser) parseIndentedIndexOrSliceExpression(exprSoFar ast.Expr, x *ast.Expr) bool {
	return p.parseIndentedSliceExpression(exprSoFar, x) || p.parseIndentedIndexExpression(exprSoFar, x)
}

func (p *parser) parseIndentedSliceExpression(exprSoFar ast.Expr, x *ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "IndentedSliceExpression"))
    }
    st := p.State()
    col := p.Col()
    lBracketPos := p.Pos()

    if ! p.Match1('[') {
       return false
    }
    if ! p.Indent(col) {
	   return p.Fail(st)
	}
	
	var low, high ast.Expr
		
	if ! p.Match1(':') {
	  if ! p.parseExpression(&low) {
	     return p.Fail(st)
	  }
	  if ! p.Indent(col) {
	     return p.Fail(st)			
	  } 
	  if ! p.Match1(':') {
	     return p.Fail(st)	
	  }
	}	
	
    if p.Indent(col) {	
	   p.required(p.parseExpression(&high), "an expression, indented from the brackets")
	}
	
    p.required(p.Below(col),"] below [")

    rBracketPos := p.Pos()
    p.required(p.Match1(']'),"] below [") 

    *x = &ast.SliceExpr{exprSoFar,lBracketPos, low, high, rBracketPos}

    return true
}


func (p *parser) parseOneLineIndexOrSliceExpression(exprSoFar ast.Expr, x *ast.Expr) bool {
	return p.parseOneLineSliceExpression(exprSoFar, x) || p.parseOneLineIndexExpression(exprSoFar, x)
}


func (p *parser) parseOneLineSliceExpression(exprSoFar ast.Expr, x *ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "OneLineSliceExpression"))
    }
    st := p.State()
    lBracketPos := p.Pos()

    if ! p.Match1('[') {
	   return false
    }

	var low, high ast.Expr
    
    if ! p.Match1(':') {
      if ! p.parseOneLineExpression(&low,false,false) {
	     return p.Fail(st)
      }
      if ! p.Match1(':') {
	     return p.Fail(st)	
      }
    }
    p.optional(p.parseOneLineExpression(&high,false,false))
    rBracketPos := p.Pos()
    p.required(p.Match1(']'),"']'")

    *x = &ast.SliceExpr{exprSoFar,lBracketPos, low, high, rBracketPos}

    return true
}





func (p *parser) parseMethodCall(stmt **ast.MethodCall) bool {
    if p.trace {
       defer un(trace(p, "MethodCall"))
    }
    return p.parseIndentedMethodCall(stmt) || p.parseOneLineMethodCall(stmt, false)
}

/*
foo a1 (bar a2 a3) 
foo a1 (bar a2 a3) list1...      <- TODO
bar foo (baz 1 3) "froboz"   <- all remaining args pertain to call of method foo
*/
func (p *parser) parseOneLineMethodCall(stmt **ast.MethodCall, isInsideBrackets bool) bool {
    if p.trace {
       defer un(trace(p, "OneLineMethodCall"))
    }
    var methodName *ast.Ident

    if ! p.parseMethodName(true,&methodName) {
	   return false
    }

    var xs []ast.Expr

    st2 := p.State()

    if p.Space() { // May be arguments
       if ! (p.parseMultipleOneLineExpressions(&xs, isInsideBrackets) || p.parseSingleOneLineExpression(&xs, isInsideBrackets)) {



	      p.Fail(st2)
	   }
    }
	
	// translate
	*stmt = &ast.MethodCall{Fun:methodName,Args:xs}
	
    return true
}

/*
	// EGH A MethodCall node represents an expression followed by an argument list.
	MethodCall struct {
		Fun      Expr      // function expression
		Args     []Expr    // function arguments; or nil
		Ellipsis token.Pos // position of "...", if any
	}
*/	

/*
   foo
      9
      "The end."

   foo 9
       "The end"

   but not...

   foo bar 9
           "The end"

   foo 
      a1 
      bar a2 a3
      key1 = val1
      key2 = val2
      variadic1
      variadic2
 
   foo
      a1
      a2
      list1...


*/
func (p *parser) parseIndentedMethodCall(stmt **ast.MethodCall) bool {
    if p.trace {
       defer un(trace(p, "IndentedMethodCall"))
    }
    st := p.State()
    var methodName *ast.Ident

    if ! p.parseMethodName(true,&methodName) {
	   return false
    }

    var xs []ast.Expr
    var kws map[string]ast.Expr = make(map[string]ast.Expr)
    var firstAssignmentPos int32 = -1 // position in arg list of first assignment statement. = #positionalArgs

    if ! p.parseIndentedExpressionsOrKeywordParamAssignments(st.RuneColumn, &xs, &firstAssignmentPos, kws ) {
        if ! p.Space() {
	       return p.Fail(st)
	    }
	    if ! p.parseLenientVerticalExpressionsOrKeywordParamAssignments(p.Col(), &xs, &firstAssignmentPos, kws) {
	       return p.Fail(st)	
	    }
    }

// parseVerticalExpressionsOrKeywordParamAssignments

    // TODO Need to type check the keyword param assignments, but may have to wait til call evaluation time?
    // Shouldn't have to, because all methods of a given name and # of positional parameters must
    // have the same set of keyword paramters. <== IMPORTANT RULE
    // Except what about the special keywords arg declaration.

	*stmt = &ast.MethodCall{Fun:methodName,Args:xs, KeywordArgs:kws, NumPositionalArgs: firstAssignmentPos }
	
    return true
}





func (p *parser) parseListConstruction(stmt **ast.ListConstruction) bool {
    if p.trace {
       defer un(trace(p, "ListConstruction"))
    }
    return p.parseIndentedListConstruction(stmt) || p.parseOneLineListConstruction(stmt, false)
}

/*
Note, somehow, these have to be prevented from being standalone statements. Can only appear in expression context.

[]Car
["First" "Second" "Third"]String
[]Car "year > 2010 order by year"
[1 2 45 6]

  TypeSpec struct {
    Doc            *CommentGroup       // associated documentation; or nil
    Name           *Ident              // type name (or type variable name)
    Type           Expr                // *Ident, *ParenExpr, *SelectorExpr, *StarExpr, or any of the *XxxTypes
    Comment        *CommentGroup       // line comments; or nil
    Params         []*TypeSpec         // Type parameters (egh)
    SuperTypes     []*TypeSpec         // Only valid if this is a type variable (egh)
    CollectionSpec *CollectionTypeSpec // nil or a collection specification
  }

*/
func (p *parser) parseOneLineListConstruction(stmt **ast.ListConstruction, isInsideBrackets bool) bool {
    if p.trace {
       defer un(trace(p, "OneLineListConstruction"))
    }

//    pos := p.Pos()
//    var end token.Pos

    var typeSpec *ast.TypeSpec

//    var collectionTypeSpec *ast.CollectionTypeSpec

    var elementExprs []ast.Expr    

    emptyList := false
    hasType := false
    
    if p.Match2('[',']') {
//        end = p.Pos()
        emptyList = true
    } else if p.Match1('[') {

        // Get the list of expressions inside the list literal square-brackets

        p.required(p.parseMultipleOneLineExpressions(&elementExprs, false) || p.parseSingleOneLineExpression(&elementExprs, false),
                   "zero or more list-element expressions on the same line as [, followed by closing square-bracket ]")

        p.required(p.Match1(']'), "a closing square-bracket ], or a space followed by another list-element expression")   
//        end = p.Pos()            
    } else { 
	     return false // Did not match a list-specifying square-bracket
    }

    if p.parseTypeSpec(false, true,true,false,false,false,&typeSpec) {
       hasType = true 
    }

    if ! hasType { 
       if emptyList { // Oops - no way to infer the element-type constraint
          p.stop("An empty list literal must specify its element type. e.g. []Widget")
          return false // superfluous when parser is in single-error mode as now
       }

       // Try to infer the element-type constraint from the statically known types of the arguments.

       // TODO TODO TODO !!!!!!!!!
       // Once we are doing static type inference and type checking of variables and expressions
       // we will be able to do this properly when element types are not all the same.  
       if ! p.crudeTypeInfer(elementExprs, &typeSpec) {
          p.stop("Cannot infer element-type constraint of list. Specify element type. e.g. []Widget")
       } 
    }
/*
    isSorting := false // TODO allow sorting-list specifications in list constructions!!!!
    isAscending := false 
    orderFunc := "" 
    collectionTypeSpec = &ast.CollectionTypeSpec{token.LIST,pos,end,isSorting,isAscending,orderFunc}

    typeSpec.CollectionSpec = collectionTypeSpec
*/
    

    var queryStringExprs []ast.Expr  // there is only allowed to be one query string
    var queryStringExpr ast.Expr

    if emptyList {
	    st2 := p.State()

	    if p.Space() { // May be arguments
	       if p.parseSingleOneLineExpression(&queryStringExprs, isInsideBrackets) {
	          queryStringExpr = queryStringExprs[0]
	       } else {
		        p.Fail(st2)
		     }
	    }
	}
	
	// translate
	*stmt = &ast.ListConstruction{Type:typeSpec, Elements: elementExprs, Query:queryStringExpr}
	
    return true
}




/*

[
   "First" 
   "Second" 
    "Third"
]String

   []Car """
year > 2010 
order by year
"""

[]Car 
   "year > 2010"

   []Car
      """
year > 2010 
order by year
"""

[
   1
   2
   45
   6
]

"one or more expressions - the elements of the list - or the closing bracket ]"
*/
func (p *parser) parseIndentedListConstruction(stmt **ast.ListConstruction) bool {
    if p.trace {
       defer un(trace(p, "IndentedListConstruction"))
    }
    st := p.State()
//    pos := p.Pos()
//    var end token.Pos

    var typeSpec *ast.TypeSpec

//    var collectionTypeSpec *ast.CollectionTypeSpec

    var rangeStmt *ast.RangeStatement

    var elementExprs []ast.Expr    

    emptyList := false
    hasType := false

    if p.Match2('[',']') {
//        end = p.Pos()
        emptyList = true
    } else if p.Match1('[') {
        if p.parseIndentedForGenerator(st.RuneColumn, &rangeStmt) {
	        p.required(p.Below(st.RuneColumn) && p.Match1(']'), "closing square-bracket ] aligned exactly below opening [")  	
		} else {
	        // Get the list of expressions inside the list literal square-brackets

	        if p.parseIndentedExpressions(st.RuneColumn, &elementExprs) {
	            p.required(p.Below(st.RuneColumn) && p.Match1(']'), "closing square-bracket ] aligned exactly below opening [")   
	        } else {
	           p.required(p.parseMultipleOneLineExpressions(&elementExprs, false) || p.parseSingleOneLineExpression(&elementExprs, false),
	                   "zero or more list-element expressions, followed by closing square-bracket ]")  
	           p.required(p.Match1(']'), "a closing square-bracket ], or a space followed by another list-element expression")              
	        }
	    }
//        end = p.Pos()            
    } else { 
       return false // Did not match a list-specifying square-bracket
    }

    if p.parseTypeSpec(true, true,true,false,false,false,&typeSpec) {
       hasType = true 
    }

    if ! hasType { 
       if emptyList { // Oops - no way to infer the element-type constraint
          p.stop("An empty list literal must specify its element type. e.g. []Widget")
          return false // superfluous when parser is in single-error mode as now
       }

       // Try to infer the element-type constraint from the statically known types of the arguments.

       // TODO TODO TODO !!!!!!!!!
       // Once we are doing static type inference and type checking of variables and expressions
       // we will be able to do this properly when element types are not all the same.  
       if ! p.crudeTypeInfer(elementExprs, &typeSpec) {
          p.stop("Cannot infer element-type constraint of list. Specify element type. e.g. []Widget")
       } 
    }
/*
    isSorting := false // TODO allow sorting-list specifications in empty list constructions!!!!
    isAscending := false 
    orderFunc := "" 
    collectionTypeSpec = &ast.CollectionTypeSpec{token.LIST,pos,end,isSorting,isAscending,orderFunc}

    typeSpec.CollectionSpec = collectionTypeSpec
*/



    var queryStringExprs []ast.Expr  // there is only allowed to be one query string
    var queryStringExpr ast.Expr

    if emptyList {
	    st2 := p.State()

	    if p.parseIndentedExpressions(st.RuneColumn, &queryStringExprs ) {
	       if len(queryStringExprs) == 1 {
	           queryStringExpr = queryStringExprs[0]
	       } else {
	          p.stop("Only one argument is allowed after a list constructor - a String containing SQL-formatted selection criteria")
	       }
	    } else if ! p.isLower(st,st2) {
	        return p.Fail(st)  // Sorry, after all that, it is not a multiLine list construction
	    }    
    }

    // translate
    *stmt = &ast.ListConstruction{Type:typeSpec,  Generator:rangeStmt, Elements: elementExprs, Query:queryStringExpr}  
	
    return true
}

func (p *parser) parseIndentedForGenerator(col int, stmt **ast.RangeStatement) bool {
   st := p.State()	
   if ! p.Indent(col) {
      return false	
   }
   if ! p.parseForRangeStatement(stmt, 1, true) {
	  return p.Fail(st)
   }
   return true
}


func (p *parser) parseMapOrSetConstruction(mapStmt **ast.MapConstruction, setStmt **ast.SetConstruction) bool {
    if p.trace {
       defer un(trace(p, "MapOrSetConstruction"))
    }
    return p.parseIndentedMapOrSetConstruction(mapStmt, setStmt) || p.parseOneLineMapOrSetConstruction(mapStmt, setStmt, false)
}


/*
Note, somehow, these have to be prevented from being standalone statements. Can only appear in expression context.

[]Car
["First" "Second" "Third"]String
[]Car "year > 2010 order by year"
[1 2 45 6]

  TypeSpec struct {
    Doc            *CommentGroup       // associated documentation; or nil
    Name           *Ident              // type name (or type variable name)
    Type           Expr                // *Ident, *ParenExpr, *SelectorExpr, *StarExpr, or any of the *XxxTypes
    Comment        *CommentGroup       // line comments; or nil
    Params         []*TypeSpec         // Type parameters (egh)
    SuperTypes     []*TypeSpec         // Only valid if this is a type variable (egh)
    CollectionSpec *CollectionTypeSpec // nil or a collection specification
  }

*/
func (p *parser) parseOneLineMapOrSetConstruction(mapStmt **ast.MapConstruction, setStmt **ast.SetConstruction, isInsideBrackets bool) bool {
    if p.trace {
       defer un(trace(p, "OneLineMapOrSetConstruction"))
    }

//    pos := p.Pos()
//    var end token.Pos

    var typeSpec *ast.TypeSpec
    var valTypeSpec *ast.TypeSpec

//    var collectionTypeSpec *ast.CollectionTypeSpec

    var keyExprs []ast.Expr
    var elementExprs []ast.Expr    

    emptyMapOrSet := false
    hasType := false
    knownToBeMap := false
    knownToBeSet := false
    mapEntriesFound := false
    setEntriesFound := false
    
    if p.Match2('{','}') {
//        end = p.Pos()
        emptyMapOrSet = true
    } else if p.Match1('{') {

        // Get the list of expressions inside the map or set literal squiggly-brackets
        // Note that one of these must exist, because it is not {}

        mapEntriesFound = p.parseOneLineMapEntryExpressions(&keyExprs, &elementExprs) 

        if ! mapEntriesFound { 
			setEntriesFound = (p.parseMultipleOneLineExpressions(&elementExprs, false) || 
			                   p.parseSingleOneLineExpression(&elementExprs, false))
		}
		
		p.required(mapEntriesFound || setEntriesFound, "zero or more set-element or map-entry expressions on the same line as {, followed by closing squiggly-bracket }")

        knownToBeMap = mapEntriesFound
        knownToBeSet = setEntriesFound

/*
        p.required(p.parseMultipleOneLineMapEntryExpressions(&elementExprs, false) || p.parseSingleOneLineMapEntryExpression(&elementExprs, false),
                   "zero or more list-element expressions on the same line as {, followed by closing squiggly-bracket }")

		p.required(p.parseMultipleOneLineMapEntryExpressions(&elementExprs, false) || p.parseSingleOneLineMapEntryExpression(&elementExprs, false),
		           "zero or more list-element expressions on the same line as {, followed by closing squiggly-bracket }")
*/

        p.required(p.Match1('}'), "a closing squiggly-bracket }, or a space followed by another mapping or element expression")   
//        end = p.Pos()            
    } else { 
	     return false // Did not match a map-or-set-specifying squiggly-bracket
    }

//    col := p.Col()
    if p.parseTypeSpec(false, false,true,false,false,false,&typeSpec) {
       hasType = true 
    }

    if hasType {
	
       if knownToBeMap {	
	       p.required(p.Match2(' ','>')," > T : the map's value-type") 
       } else if knownToBeSet {
	      if p.Match2(' ','>') {
		     p.stop("A map literal must contain map entries {a=>b c=>d} not simple elements {a b}")
	      }
       } else { // could be either map or set
	      if p.Match2(' ','>') {	  
	         knownToBeMap = true
	      } else {
		     knownToBeSet = true
	      }
	   }
	   
	   if knownToBeMap {
	      p.required(p.Space(),"a space followed by the map value-type specification")
          p.required(p.parseTypeSpec(false, true,true,false,false,false,&valTypeSpec),"the map value-type specification")
	   }
    }


    if ! hasType { 
       if emptyMapOrSet { // Oops - no way to infer the element-type constraint
          p.stop("An empty map/set literal must specify its entry/element type. e.g. {}Widget or {}String > Widget")
          return false // superfluous when parser is in single-error mode as now
       }

       // Try to infer the element-type constraint from the statically known types of the arguments.

       // TODO TODO TODO !!!!!!!!!
       // Once we are doing static type inference and type checking of variables and expressions
       // we will be able to do this properly when element types are not all the same.  

       if knownToBeMap {
	       if ! p.crudeTypeInfer(keyExprs, &typeSpec) {
	          p.stop("Cannot infer key type constraint of map. Specify types for map. e.g. {}String > Widget")
	       }	 
	       if ! p.crudeTypeInfer(elementExprs, &valTypeSpec) {
	          p.stop("Cannot infer value type constraint of map. Specify types for map. e.g. {}String > Widget")
	       }	
	   } else { // It's a Set
	       if ! p.crudeTypeInfer(elementExprs, &typeSpec) {
	          p.stop("Cannot infer element-type constraint of set. Specify element type. e.g. {}Widget")
	       } 
       }
    }
/*
    isSorting := false // TODO allow sorting specifications in empty set/map constructions!!!!
    isAscending := false 
    orderFunc := "" 

    if knownToBeMap {
       collectionTypeSpec = &ast.CollectionTypeSpec{token.MAP,pos,end,isSorting,isAscending,orderFunc}
    } else {
       collectionTypeSpec = &ast.CollectionTypeSpec{token.SET,pos,end,isSorting,isAscending,orderFunc}	
    }
    typeSpec.CollectionSpec = collectionTypeSpec
*/
    if knownToBeMap {
	
		// translate
		*mapStmt = &ast.MapConstruction{Type:typeSpec, ValType: valTypeSpec, Keys: keyExprs, Elements: elementExprs}	
	
   } else {
    
	    var queryStringExprs []ast.Expr  // there is only allowed to be one query string
	    var queryStringExpr ast.Expr
        
        if emptyMapOrSet {
		    st2 := p.State()

		    if p.Space() { // May be arguments
		       if p.parseSingleOneLineExpression(&queryStringExprs, isInsideBrackets) {
		          queryStringExpr = queryStringExprs[0]
		       } else {
			        p.Fail(st2)
			     }
		    }
        }
	
	    // translate
	    *setStmt = &ast.SetConstruction{Type:typeSpec, Elements: elementExprs, Query:queryStringExpr}
	}
    return true
}


func (p *parser) parseIndentedMapOrSetConstruction(mapStmt **ast.MapConstruction, setStmt **ast.SetConstruction) bool {
    if p.trace {
       defer un(trace(p, "IndentedMapOrSetConstruction"))
    }

    st := p.State()
//  	pos := p.Pos()
//  	var end token.Pos

  	var typeSpec *ast.TypeSpec
  	var valTypeSpec *ast.TypeSpec

//  	var collectionTypeSpec *ast.CollectionTypeSpec

    var rangeStmt *ast.RangeStatement

  	var keyExprs []ast.Expr
  	var elementExprs []ast.Expr    

  	emptyMapOrSet := false
  	hasType := false
  	knownToBeMap := false
  	knownToBeSet := false
  	mapEntriesFound := false
  	setEntriesFound := false

  	if p.Match2('{','}') {
//  	    end = p.Pos()
  	    emptyMapOrSet = true
  	} else if p.Match1('{') {
        if p.parseIndentedForGenerator(st.RuneColumn, &rangeStmt) {		
	        p.required(p.Below(st.RuneColumn) && p.Match1('}'), "closing squiggly-bracket } aligned exactly below opening {")  
	    } else {	
	        // Get the list of expressions inside the map or set literal squiggly-brackets
	        // Note that one of these must exist, because it is not {}

	        mapEntriesFound = p.parseIndentedMapEntryExpressions(st.RuneColumn, &keyExprs, &elementExprs) 

	        if ! mapEntriesFound { 
	  			   setEntriesFound = p.parseIndentedExpressions(st.RuneColumn, &elementExprs)
	  		  }

	        if mapEntriesFound || setEntriesFound {
	           p.required(p.Below(st.RuneColumn) && p.Match1('}'), "closing squiggly-bracket } aligned exactly below opening {") 
	        }
  		
	        knownToBeMap = mapEntriesFound
	        knownToBeSet = setEntriesFound

	  	    if ! (mapEntriesFound || setEntriesFound) {
  	    
	  	      mapEntriesFound = p.parseOneLineMapEntryExpressions(&keyExprs, &elementExprs) 

	  		    if ! mapEntriesFound { 
	  				   setEntriesFound = (p.parseMultipleOneLineExpressions(&elementExprs, false) || 
	  				                      p.parseSingleOneLineExpression(&elementExprs, false))
	  				}
	  			  if mapEntriesFound || setEntriesFound {
	  		          p.required(p.Match1('}'), "a closing squiggly-bracket }, or a space followed by another mapping or element expression")  
	  				}
	  	    }

	  		  p.required(mapEntriesFound || setEntriesFound, "zero or more set-element or map-entry expressions, followed by closing squiggly-bracket }")

	  	    knownToBeMap = mapEntriesFound
	  	    knownToBeSet = setEntriesFound
         }
//         end = p.Pos()            
      } else { 
         return false // Did not match a map-or-set-specifying squiggly-bracket
      }

    typeCol := p.Col()
  	if p.parseTypeSpec(false,false,true,false,false,false,&typeSpec) { // TODO first arg should be true
  	   hasType = true                                                 // replacing some of the below
  	}

  	if hasType {
       foundTypeRelation := false
       horizontalTypeRelationLayout := false
  	   if knownToBeMap {	
           foundTypeRelation = p.Match2(' ','>')
           if foundTypeRelation {
               horizontalTypeRelationLayout = true
           } else {  // look for vertical
               if p.Below(typeCol) {
                  p.required(p.Match1('>'),"> T : the map's value-type")
                  foundTypeRelation = true
               }
           }
  	       p.required(foundTypeRelation," > T : the map's value-type") 
  	   } else if knownToBeSet {
  	      if p.Match2(' ','>') {
  		     p.stop("A map literal must contain map entries {a=>b c=>d} not simple elements {a b}")
  	      }
  	   } else { // could be either map or set
  	      if p.Match2(' ','>') {	  
  	         knownToBeMap = true
             foundTypeRelation = true
             horizontalTypeRelationLayout = true
          } else if p.Below(typeCol) {
             p.required(p.Match1('>'),"> T : the map's value-type")
             knownToBeMap = true           
             foundTypeRelation = true        
          } else {
  		     knownToBeSet = true
  	      }
  	   }
     
  	   if knownToBeMap {
          if horizontalTypeRelationLayout {
  	         p.required(p.Space(),"a space followed by the map value-type specification")
             p.required(p.parseTypeSpec(true, true,true,false,false,false,&valTypeSpec),"the map value-type specification")           
          } else { // vertical layout 
             foundValType := false
             if p.Space() {
                if p.parseTypeSpec(true, true,true,false,false,false,&valTypeSpec) {
                   foundValType = true
                } else {
                   p.required(p.Space() || (p.Ch()  == '\n'),"the map value-type specification")
                }
             }
             if ! foundValType {
                p.required(p.Indent(typeCol) && p.parseTypeSpec(true, true,true,false,false,false,&valTypeSpec),
                           "the map value-type specification, after a space or indented on next line")                                        
             } 
          }  
  	   }
  	}

  	if ! hasType { 
  	   if emptyMapOrSet { // Oops - no way to infer the element-type constraint
  	      p.stop("An empty map/set literal must specify its entry/element type. e.g. {}Widget or {}String > Widget")
  	      return false // superfluous when parser is in single-error mode as now
  	   }

  	   // Try to infer the element-type constraint from the statically known types of the arguments.

  	   // TODO TODO TODO !!!!!!!!!
  	   // Once we are doing static type inference and type checking of variables and expressions
  	   // we will be able to do this properly when element types are not all the same.  

  	   if knownToBeMap {
  	       if ! p.crudeTypeInfer(keyExprs, &typeSpec) {
  	          p.stop("Cannot infer key type constraint of map. Specify types for map. e.g. {}String > Widget")
  	       }	 
  	       if ! p.crudeTypeInfer(elementExprs, &valTypeSpec) {
  	          p.stop("Cannot infer value type constraint of map. Specify types for map. e.g. {}String > Widget")
  	       }	
  	   } else { // It's a Set
  	       if ! p.crudeTypeInfer(elementExprs, &typeSpec) {
  	          p.stop("Cannot infer element-type constraint of set. Specify element type. e.g. {}Widget")
  	       } 
  	   }
  	}
/*
    isSorting := false // TODO allow sorting specifications in empty set/map constructions!!!!
    isAscending := false 
    orderFunc := "" 

    if knownToBeMap {
       collectionTypeSpec = &ast.CollectionTypeSpec{token.MAP,pos,end,isSorting,isAscending,orderFunc}
    } else {
       collectionTypeSpec = &ast.CollectionTypeSpec{token.SET,pos,end,isSorting,isAscending,orderFunc}  
    }
    typeSpec.CollectionSpec = collectionTypeSpec
*/
    if knownToBeMap {
  
      // translate
      *mapStmt = &ast.MapConstruction{Type:typeSpec, ValType: valTypeSpec,  Generator: rangeStmt, Keys: keyExprs, Elements: elementExprs}  
  
    } else {
    
       var queryStringExprs []ast.Expr  // there is only allowed to be one query string
       var queryStringExpr ast.Expr

       if emptyMapOrSet {
	       st2 := p.State()

	      if p.parseIndentedExpressions(st.RuneColumn, &queryStringExprs ) {
	         if len(queryStringExprs) == 1 {
	             queryStringExpr = queryStringExprs[0]
	         } else {
	            p.stop("Only one argument is allowed after a set constructor - a String containing SQL-formatted selection criteria")
	         }
	      } else if ! p.isLower(st,st2) {
	          return p.Fail(st)  // Sorry, after all that, it is not a multiLine set construction
	      }   
       } 
  
      // translate
      *setStmt = &ast.SetConstruction{Type:typeSpec, Generator: rangeStmt, Elements: elementExprs, Query:queryStringExpr}
    }
    return true
}


/*
Parse a closure expression.
If succeeds, sets the closure var to point to the Closure ast node, and also
has added the closure-method declaration to the file's method declarations.
*/
func (p *parser) parseClosure(closure **ast.Closure) bool {
	if p.trace {
	   defer un(trace(p, "Closure"))
	}	
	
	alreadyInClosureDeclaration := false
	if p.parsingClosure {
	   alreadyInClosureDeclaration = true
	}
	if ! alreadyInClosureDeclaration {
       p.pushVariableScope()
       defer p.popVariableScope()
    }
	
   if p.parseMethodDeclaration(&(p.closureMethodDecls)) {
	 
	  closureMethodDecl := p.closureMethodDecls[len(p.closureMethodDecls)-1]
	
	  // generate 
	
	  // fmt.Println("In parseClosure: p.closureFreeVarBindings", p.closureFreeVarBindings)
	  *closure = &ast.Closure{closureMethodDecl.Name.NamePos, p.closureMethodName, p.closureFreeVarBindings}
	  // fmt.Println("closure.Bindings",(*closure).Bindings)
	
      if alreadyInClosureDeclaration {
	     p.stop("A closure cannot be nested in another closure declaration")
      }	
	
	  return true
   }
   return false
}



/*
	// egh A TypeAssertion node represents an expression preceded by a
	// type assertion.
	//
	TypeAssertion struct {
		Type TypeSpec // asserted type		
		X    Expr // expression
	}
	
	Bag of 
	   List of Tricks
	:
	   summon "My bag"

func (p *parser) parseTypeAssertion(assertion **ast.TypeAssertion) bool {
	typeSpec *ast.TypeSpec
	
	st := p.State()
	if ! p.parseTypeSpec
}
*/




// ----------------------------------------------------------------------------
// ---- Type Inference

//
// TODO   MAJOR !!  Add a lot more of this, such as giving type constraints to local variables based on 
// the statically inferable types of the values assigned to them within the method body
//


/*
If it can, uses the preliminary compile-time type information about the element expressions to 
infer the typespec which is to be the element-type constraint of a collection. Sets the typeSpec argument to
point to the element type constraint type spec that it creates.

Currently is very crude, and can only work on elementExprs which are some kinds of RelishPrimitive literals e.g. 
String and various Numeric literals, maybe also Bool <- TODO not yet!!

If it cannot decide the typespec based on this crude assessment, it returns false and does not create a TypeSpec.
*/
func (p *parser) crudeTypeInfer(elementExprs []ast.Expr, typeSpec **ast.TypeSpec) bool {
  
  var primitiveTypes map[token.Token]bool 
  primitiveTypes = make(map[token.Token]bool)
  
  var pos token.Pos
  pos = token.NoPos
  for _,elementExpr := range elementExprs {
     switch elementExpr.(type) {
      case *ast.BasicLit :
       lit := elementExpr.(*ast.BasicLit)
         primitiveTypes[lit.Kind] = true     
         if pos == token.NoPos {
          pos = lit.Pos()
         }        
      default : 
         return false // Not a RelishPrimitive literal. Can't handle here.
     }
  }

    // TODO Not handling TRUE FALSE (which should be true false) here
    // There needs to be a BasicLit which is of tok.BOOL
    var typeName *ast.Ident
    var typName string
    if len(primitiveTypes) == 1 {
        for tok,_ := range primitiveTypes {
         typName = strings.Title(strings.ToLower(tok.String()))
           typeName = &ast.Ident{pos, typName, nil, token.TYPE,-1}  
        }
    } else if len(primitiveTypes) == 2 { // See if types are all Numeric, else Choose RelishPrimitive
     if primitiveTypes[token.INT] && primitiveTypes[token.FLOAT] {
       typName = "Numeric"
     } else {
        typName = "RelishPrimitive" 
     }
    } else { // has to be RelishPrimitive
    typName = "RelishPrimitive"
    }
    
    typeName = &ast.Ident{pos, typName, nil, token.TYPE,-1}
  
  *typeSpec = &ast.TypeSpec{Name: typeName}
  
    return true
}





// ----------------------------------------------------------------------------
// Statements





func (p *parser) parseMethodBodyStatement(stmts *[]ast.Stmt, nReturnVals int, noResults bool) bool {
    if p.trace {
       defer un(trace(p, "MethodBodyStatement"))
    }

    var s ast.Stmt
    // parse
    if p.parseControlStatement(&s, nReturnVals, false) {
	   // translate
	   *stmts = append(*stmts,s)
	   return true
    }
    
    var rs *ast.ReturnStatement
    // parse
    if p.parseReturnStatement(&rs, nReturnVals, false) {
	
	   if nReturnVals == 0 {
		  if noResults {
              p.stop("=> statement not allowed, except in 'if' 'while' or 'for', because this method returns no results")		
          } else {
	          p.stop("=> statement not allowed, except in 'if' 'while' or 'for', because this method has named return arguments")		
          }
       }

	   // translate
	   *stmts = append(*stmts,rs)
	   return true	
    }

    var as *ast.AssignmentStatement
    // parse
    if p.parseAssignmentStatement(&as) {	
	   // translate
	   *stmts = append(*stmts,as)	
	   return true	
    }

    var mcs *ast.MethodCall
    // parse
    if p.parseMethodCall(&mcs) {
	   // translate
	   *stmts = append(*stmts,mcs)	
	   return true	
    }

    return false
}

/*
Add a break and continue statement and also a return statement modifier "."
*/
func (p *parser) parseIfClauseStatement(stmts *[]ast.Stmt, nReturnVals int, isInsideLoop bool) bool {
    if p.trace {
       defer un(trace(p, "IfClauseStatement"))
    }

    var s ast.Stmt
    // parse
    if p.parseControlStatement(&s, nReturnVals, isInsideLoop) {
	   // translate
	   *stmts = append(*stmts,s)
	   return true
    }
    
    var rs *ast.ReturnStatement
    // parse
    if p.parseReturnStatement(&rs, nReturnVals, false) {
	   // translate
	   *stmts = append(*stmts,rs)
	   return true	
    }

    var as *ast.AssignmentStatement
    // parse
    if p.parseAssignmentStatement(&as) {	
	   // translate
	   *stmts = append(*stmts,as)	
	   return true	
    }

    var mcs *ast.MethodCall
    // parse
    if p.parseMethodCall(&mcs) {
	   // translate
	   *stmts = append(*stmts,mcs)	
	   return true	
    }

    var bks *ast.BreakStatement
    if p.parseBreakStatement(&bks) {
    	if ! isInsideLoop {
	       p.stop("break statement can only occur inside a loop")	
	   }
	   // translate
	   *stmts = append(*stmts,bks)	
	   return true	    	
    }

    var cs *ast.ContinueStatement
    if p.parseContinueStatement(&cs) {
    	if ! isInsideLoop {
	       p.stop("continue statement can only occur inside a loop")	
	   }
	   // translate
	   *stmts = append(*stmts,cs)	
	   return true	    	
    }    


    return false
}

func (p *parser) parseBreakStatement(stmt **ast.BreakStatement) bool {
    if p.trace {
       defer un(trace(p, "BreakStatement"))
    }	
    pos := p.Pos()	
	if ! p.MatchWord("break") {
		return false
	}
    // translate
    *stmt = &ast.BreakStatement{pos}	

    return true
}


func (p *parser) parseContinueStatement(stmt **ast.ContinueStatement) bool {
    if p.trace {
       defer un(trace(p, "ContinueStatement"))
    }	
    pos := p.Pos()	
	if ! p.MatchWord("continue") {
		return false
	}

    // translate
    *stmt = &ast.ContinueStatement{pos}

    return true
}


/*
How is this different from method body?
*/
func (p *parser) parseLoopBodyStatement(stmts *[]ast.Stmt, nReturnVals int) bool {
    if p.trace {
       defer un(trace(p, "LoopBodyStatement"))
    }

    var s ast.Stmt
    // parse
    if p.parseControlStatement(&s, nReturnVals, true) {
	   // translate
	   *stmts = append(*stmts,s)
	   return true
    }
    
    var rs *ast.ReturnStatement
    // parse
    if p.parseReturnStatement(&rs, nReturnVals, false) {
	   // translate
	   *stmts = append(*stmts,rs)
	   return true	
    }

    var as *ast.AssignmentStatement
    // parse
    if p.parseAssignmentStatement(&as) {	
	   // translate
	   *stmts = append(*stmts,as)	
	   return true	
    }

    var mcs *ast.MethodCall
    // parse
    if p.parseMethodCall(&mcs) {
	   // translate
	   *stmts = append(*stmts,mcs)	
	   return true	
    }

    return false
}


func (p *parser) parseControlStatement(stmt *ast.Stmt, nReturnVals int, isInsideLoop bool) bool {
    if p.trace {
       defer un(trace(p, "ControlStatement"))
    }
    var ifStmt *ast.IfStatement
    if p.parseIfStatement(&ifStmt, nReturnVals, false, isInsideLoop) {
	   *stmt = ifStmt
	   return true
    } 
    var whileStmt *ast.WhileStatement
    if p.parseWhileStatement(&whileStmt, nReturnVals, isInsideLoop) {
	   *stmt = whileStmt
	   return true
    }  
    var rangeStmt *ast.RangeStatement
    if p.parseForRangeStatement(&rangeStmt, nReturnVals, false) {
	   *stmt = rangeStmt
	   return true
    }   

    var forStmt *ast.ForStatement
    if p.parseForStatement(&forStmt, nReturnVals) {
	   *stmt = forStmt
	   return true
    }

    var goStmt *ast.GoStatement
    if p.parseGoStatement(&goStmt) {
	   *stmt = goStmt
	   return true
    }
    var deferStmt *ast.DeferStatement
    if p.parseDeferStatement(&deferStmt) {
	   *stmt = deferStmt
	   return true
    }
/*
    var forStmt *ast.ForStatement
    if p.parseForStatement(&forStmt, nReturnVals) {
	   *stmt = forStmt
	   return true
    }
*/
    return false
}



/*
if expr
   block
[elif expr
   block}...
[else
   block]

  If isGeneratorExpr is true, the blocks can only contain "yield" statements i.e. ReturnStatements with IsYield==true
*/
func (p *parser) parseIfStatement(ifStmt **ast.IfStatement, nReturnVals int, isGeneratorExpr bool, isInsideLoop bool) bool {
    if p.trace {
       defer un(trace(p, "IfStatement"))
    }

    pos := p.Pos()
   
    // parse
    col := p.Col()
    if ! p.Match("if ") {
 	   return false
    }

    // translate
    var x ast.Expr
	var stmtList []ast.Stmt

    // parse
    p.required(p.parseExpression(&x),"an expression") 

    st2 := p.State()
    if ! p.BlanksAndIndent(col, true) {  // TODO Should limit this to one allowed blank line in each gap
       p.checkForIndentError()
       p.required(false, "a statement, indented from the 'if'")
    }     

    blockPos := p.Pos()

    if isGeneratorExpr {
       var rs *ast.ReturnStatement  
       var ifStatmt *ast.IfStatement

       // parse
       if p.parseIfStatement(&ifStatmt, nReturnVals, true, false) {
          stmtList = []ast.Stmt{ifStatmt}
       } else {
          p.required(p.parseReturnStatement(&rs, nReturnVals, true), "an 'if' or an expression or expressions that this generator will yield")
          stmtList = []ast.Stmt{rs}
       }
    } else {

        // parse
        p.required(p.parseIfClauseStatement(&stmtList, nReturnVals, isInsideLoop),"a statement")
        for ; p.BlanksAndIndent(col,true) ; {
           p.required(p.parseIfClauseStatement(&stmtList, nReturnVals, isInsideLoop),"a statement")	
        }

        for p.BlanksThenIndentedLineComment(col) {}  // Gobble trailing indented line comments        
        
        p.checkForIndentError()
    } 

    // translate
    body := &ast.BlockStatement{blockPos,stmtList}
    if0 := &ast.IfStatement{pos, x, body, nil}
    lastIf := if0
    *ifStmt = if0

    // parse
    st2 = p.State()
    for p.BlanksAndBelow(col, true)  {

       pos = p.Pos()

       if p.Match("elif ") {
	
	      p.required(p.parseExpression(&x),"an expression") 
	      if ! p.BlanksAndIndent(col, true) {
	      	 p.checkForIndentError()
	         p.required(false, "a statement, indented from the 'elif'")
	      }
	
	      // translate
 	      var stmtList2 []ast.Stmt
        	blockPos = p.Pos()	
	     
	        if isGeneratorExpr {
	           var rs *ast.ReturnStatement  
	           var ifStatmt *ast.IfStatement

	           // parse
	           if p.parseIfStatement(&ifStatmt, nReturnVals, true, false) {
	              stmtList2 = []ast.Stmt{ifStatmt}
	           } else {
	              p.required(p.parseReturnStatement(&rs, nReturnVals, true), "an 'if' or an expression or expressions that this generator will yield")
	              stmtList2 = []ast.Stmt{rs}
	           }
	        } else {

	            // parse
	            p.required(p.parseIfClauseStatement(&stmtList2, nReturnVals, isInsideLoop),"a statement")
	            for ; p.BlanksAndIndent(col,true) ; {
	               p.required(p.parseIfClauseStatement(&stmtList2, nReturnVals, isInsideLoop),"a statement") 
	            }
        
                for p.BlanksThenIndentedLineComment(col) {}  // Gobble trailing indented line comments        
        
                p.checkForIndentError()


	        } 
	      st2 = p.State()
	
	      // translate
	      body = &ast.BlockStatement{blockPos,stmtList2}
	      if1 := &ast.IfStatement{pos, x, body, nil}
	      lastIf.Else = if1
	      lastIf = if1	
	
	   // parse
       } else if p.Match("else") {

	      if ! p.BlanksAndIndent(col, true) {
	      	 p.checkForIndentError()
	         p.required(false, "a statement, indented from the 'else'")
	      }	      

          // translate
          var stmtList3 []ast.Stmt
          blockPos = p.Pos()
	
          if isGeneratorExpr {
             var rs *ast.ReturnStatement  
             var ifStatmt *ast.IfStatement

             // parse
             if p.parseIfStatement(&ifStatmt, nReturnVals, true, false) {
                stmtList3 = []ast.Stmt{ifStatmt}
             } else {
                p.required(p.parseReturnStatement(&rs, nReturnVals, true), "an 'if' or an expression or expressions that this generator will yield")
                stmtList3 = []ast.Stmt{rs}
             }

          } else {

              // parse
              p.required(p.parseIfClauseStatement(&stmtList3, nReturnVals, isInsideLoop),"a statement")
              for ; p.BlanksAndIndent(col,true) ; {
                 p.required(p.parseIfClauseStatement(&stmtList3, nReturnVals, isInsideLoop),"a statement") 
              }

              for p.BlanksThenIndentedLineComment(col) {}  // Gobble trailing indented line comments        
        
              p.checkForIndentError()
          }  
	
	        // translate
	        body = &ast.BlockStatement{blockPos,stmtList3}
	        lastIf.Else = body	
	
	        break       	  
       } else {
	       p.Fail(st2)
           break
       }
    }
    return true
}

/*
while expr
   block
[elif expr
   block}...
[else
   block]

WhileStatement struct {
	While token.Pos // position of "while" keyword		
	Cond Expr      // condition
	Body *BlockStatement
	Else Stmt // else branch; or nil		
}



*/
func (p *parser) parseWhileStatement(whileStmt **ast.WhileStatement, nReturnVals int, isInsideLoop bool) bool {
    if p.trace {
       defer un(trace(p, "WhileStatement"))
    }

    pos := p.Pos()

    // parse
    col := p.Col()
    if ! p.Match("while ") {
	   return false
    }

    // translate
    var x ast.Expr
	var stmtList []ast.Stmt
    // parse
    p.required(p.parseExpression(&x),"an expression") 

    if ! p.BlanksAndIndent(col, true) {  // TODO Should limit this to one allowed blank line in each gap
       p.checkForIndentError()	       
       p.required(false, "a statement, indented from the 'while'")
    } 


    blockPos := p.Pos()

    p.required(p.parseLoopBodyStatement(&stmtList, nReturnVals),"a statement")

    for p.BlanksAndIndent(col, true) {
       p.required(p.parseLoopBodyStatement(&stmtList, nReturnVals),"a statement")	
    }

    for p.BlanksThenIndentedLineComment(col) {}  // Gobble trailing indented line comments

    p.checkForIndentError()

    // translate
    body := &ast.BlockStatement{blockPos,stmtList}
    while0 := &ast.WhileStatement{pos, x, body, nil}
    var lastIf *ast.IfStatement
    lastWhile := while0
    *whileStmt = while0

    // parse
    st2 := p.State()
    for p.BlanksAndBelow(col, true) {

       pos = p.Pos()

       if p.Match("elif ") {
	
	      p.required(p.parseExpression(&x),"an expression") 
	      
	      if ! p.BlanksAndIndent(col, true) {
	      	 p.checkForIndentError()
	         p.required(false, "a statement, indented from the 'elif'")
	      }	


	      // translate
 	      var stmtList2 []ast.Stmt
          blockPos = p.Pos()	
	     
	      // parse
	      p.required(p.parseIfClauseStatement(&stmtList2, nReturnVals, isInsideLoop),"a statement")
	      for ; p.BlanksAndIndent(col, true) ; {
	         p.required(p.parseIfClauseStatement(&stmtList2, nReturnVals, isInsideLoop),"a statement")
	      }	

          for p.BlanksThenIndentedLineComment(col) {}  // Gobble trailing indented line comments        
        
          p.checkForIndentError()

	      st2 = p.State()
	
	      // translate
	      body = &ast.BlockStatement{blockPos,stmtList2}
	      if1 := &ast.IfStatement{pos, x, body, nil}
	      if lastIf == nil {
		     lastWhile.Else = if1
	      } else {
	         lastIf.Else = if1
          }
	      lastIf = if1	
	
	   // parse
       } else if p.Match("else") {

	      if ! p.BlanksAndIndent(col, true) {
	      	 p.checkForIndentError()
	         p.required(false, "a statement, indented from the 'else'")
	      }		      
	
          // translate
          var stmtList3 []ast.Stmt
          blockPos = p.Pos()
	
	      // parse
	      p.required(p.parseIfClauseStatement(&stmtList3, nReturnVals, isInsideLoop),"a statement")	
	      for ; p.BlanksAndIndent(col,true) ; {
		 	p.required(p.parseIfClauseStatement(&stmtList3, nReturnVals, isInsideLoop),"a statement")
	      }  

          for p.BlanksThenIndentedLineComment(col) {}  // Gobble trailing indented line comments        
        
          p.checkForIndentError()

	
	      // translate
	      body = &ast.BlockStatement{blockPos,stmtList3}	
	      if lastIf == nil {
		     lastWhile.Else = body
	      } else {
	         lastIf.Else = body
          }	
	
	      break       	  
       } else {
	      p.Fail(st2)
	      break
       }
    }
    return true
}	
	



/*
for i = 0 
    less i n
    i = plus i 1
   statement
   statement

for i j =
       min a b
       min c d
    and less i m
        less j n
    i j = 
       plus i 1
       calcJ i
   statement
   statement

for i = 0   less i n   i = plus i 1
 



// EGH A ForStatement represents a for statement.
ForStatement struct {
	For  token.Pos // position of "for" keyword
	Init Stmt      // initialization statement
	Cond Expr      // condition
	Post Stmt    // post iteration statement
	Body *BlockStatement
}


*/

func (p *parser) parseForStatement(forStmt **ast.ForStatement, nReturnVals int) bool {
    if p.trace {
       defer un(trace(p, "ForStatement"))
    }

    pos := p.Pos()

    // parse
    col := p.Col()
    if ! p.Match("for ") {
	   return false
    }
    exprCol := p.Col()

    var init *ast.AssignmentStatement
    var cond ast.Expr
    var post *ast.AssignmentStatement

	var stmtList []ast.Stmt

    // parse
    p.required(p.parseAssignmentStatement(&init), "an assignment statement or 'var(s) in collections' expression")    	
	
    if p.Below(exprCol) { // vertical expressions
	   p.required(p.parseExpression(&cond), "a condition expression")
	   p.required(p.Below(exprCol),"a post-iteration assignment statement, directly below the condition expression")
	   p.required(p.parseAssignmentStatement(&post), "a post-iteration assignment statement")
	
       if len(init.Lhs) != len(post.Lhs) {
          p.stop("Number of post-iteration assigned variables must be same as number of initialized variables")
       }	
	
       for p.BlanksThenIndentedLineComment(exprCol) {}  // Gobble trailing indented line comments	
	
    } else {
	   p.required(p.TripleSpace(), "a condition expression, separated by 3 spaces from assignment statement, or directly below assignment statement")
	 
	   p.required(p.parseOneLineExpression(&cond, false, false), "a condition expression, separated by 3 spaces from assignment statement, or directly below assignment statement")
	
       if len(init.Lhs) > 1 {
	       p.stop("The single-line, 3-space separated form of 'for i' can only initialize (assign to) one index variable")
	   }

	   p.required(p.TripleSpace(), "a post-iteration assignment statement, separated by 3 spaces from condition")
	
	   p.required(p.parseAssignmentStatement(&post), "a post-iteration assignment statement, separated by 3 spaces from condition")

       if len(post.Lhs) > 1 {
          p.stop("The single-line, 3-space separated form of 'for i' can only assign to a single index variable")
       }
    } 


   
    // loop body statements

	if ! p.BlanksAndIndent(col, true) {  // TODO Should limit this to one allowed blank line in each gap    
	   p.checkForIndentError()	       
	   p.required(false, "a statement, indented from the 'for'")
	} 

	blockPos := p.Pos()

	// parse
    p.required(p.parseLoopBodyStatement(&stmtList, nReturnVals),"a statement")
    for p.BlanksAndIndent(col,true) {
       p.required(p.parseLoopBodyStatement(&stmtList, nReturnVals),"a statement")	
    }

    for p.BlanksThenIndentedLineComment(col) {}  // Gobble trailing indented line comments

    p.checkForIndentError()	
	

	// translate
	body := &ast.BlockStatement{blockPos,stmtList}	
	
	*forStmt = &ast.ForStatement{pos, init, cond, post, body}

	return true
}



/*
for i val in someList
   statement
   statement
*/
func (p *parser) parseForRangeStatement(rangeStmt **ast.RangeStatement, nReturnVals int,  isGeneratorExpr bool) bool {
    if p.trace {
       defer un(trace(p, "ForRangeStatement"))
    }

    st := p.State()
    pos := p.Pos()

    // parse
    col := p.Col()
    if ! p.Match("for ") {
	   return false
    }

    // translate
    // var x ast.Expr

    var keyAndValueVariables []ast.Expr
    var collectionExprs []ast.Expr

	var stmtList []ast.Stmt



    // parse

    // look for an lhs like an assignment statement
    // Currently disallowing anything except local variable names in the lhs expression list

    if ! p.parseLHSList(true, &keyAndValueVariables) {
       return p.Fail(st)
    }

    if p.Match(" in ") {
        isInsideBrackets := false	

        p.required(p.parseLenientVerticalExpressions(p.Col(),&collectionExprs) || 
                   p.parseMultipleOneLineExpressions(&collectionExprs,isInsideBrackets) || 
                   p.parseSingleOneLineExpression(&collectionExprs,isInsideBrackets), 
                   "an expression evaluating to a collection")
    } else if p.Indent(col) {
	    col2 := p.Col()
	    if ! p.Match2('i','n') {
	       return p.Fail(st)
	    }
        if ! p.parseIndentedExpressions(col2, &collectionExprs) {	
		    return p.Fail(st)
	    }
    } else {
	    return p.Fail(st)
    }

    // declare any undeclared variables
    for _,expr := range keyAndValueVariables {
	   switch expr.(type) {
	      case *ast.Ident:
		     variable := expr.(*ast.Ident)
             p.ensureCurrentScopeVariable(variable, false)
	      default: 
		    dbg.Logln(dbg.PARSE_,"------------IS NOT A VARIABLE ---------------")
		    dbg.Logln(dbg.PARSE_,expr)
       }
    }   

    // TODO for the above: Check that each expression's type is a collection

    if ! p.BlanksAndIndent(col, true) {  // TODO Should limit this to one allowed blank line in each gap    
       p.checkForIndentError()	       
       p.required(false, "a statement, indented from the 'for'")
    } 

    
    blockPos := p.Pos()


    if isGeneratorExpr {
	
       var rs *ast.ReturnStatement	
       var ifStmt *ast.IfStatement

       // parse
       if p.parseIfStatement(&ifStmt, nReturnVals, true, false) {
          stmtList = []ast.Stmt{ifStmt}
       } else {
	        p.required(p.parseReturnStatement(&rs, nReturnVals, true), "an 'if' or an expression or expressions that this generator will yield")
          stmtList = []ast.Stmt{rs}
       }
    } else {

       // parse
	     p.required(p.parseLoopBodyStatement(&stmtList, nReturnVals),"a statement")
	     for ; p.BlanksAndIndent(col,true) ; {
	        p.required(p.parseLoopBodyStatement(&stmtList, nReturnVals),"a statement")	
	     }

         for p.BlanksThenIndentedLineComment(col) {}  // Gobble trailing indented line comments

         p.checkForIndentError()	
    }

    // translate
    body := &ast.BlockStatement{blockPos,stmtList}
    *rangeStmt = &ast.RangeStatement{pos, keyAndValueVariables, collectionExprs, body}

    return true
}
/*
	For        token.Pos   // position of "for" keyword
	Key        Expr
	Value []Expr           // One or both of Key or Value may be nil
	                       // This is not correct!!! we need to handle multiple values
	X          []Expr        // value to range over NOT CORRECT!!! Need to handle multiple expressions.
	Body       *BlockStatement
*/




/*
go methodCall

GoStatement struct {
	Go   token.Pos // position of "go" keyword
	Call *MethodCall
}

*/
func (p *parser) parseGoStatement(goStmt **ast.GoStatement) bool {
    if p.trace {
       defer un(trace(p, "GoStatement"))
    }

    pos := p.Pos()

    // parse
    if ! p.Match("go ") {
	   return false
    }

    var mcs *ast.MethodCall
    // parse
    p.required(p.parseMethodCall(&mcs),"a method call")

    // translate
    *goStmt = &ast.GoStatement{pos, mcs}

    return true
}


/*
defer methodCall

DeferStatement struct {
	Defer   token.Pos // position of "go" keyword
	Call *MethodCall
}

*/
func (p *parser) parseDeferStatement(deferStmt **ast.DeferStatement) bool {
    if p.trace {
       defer un(trace(p, "DeferStatement"))
    }

    pos := p.Pos()

    // parse
    if ! p.Match("defer ") {
	   return false
    }

    var mcs *ast.MethodCall
    // parse
    p.required(p.parseMethodCall(&mcs),"a method call")

    // translate
    *deferStmt = &ast.DeferStatement{pos, mcs}

    return true
}




/*
   => foo 9
      b

ReturnStatement struct {
	Return  token.Pos // position of "=>" keyword
	Results []Expr    // result expressions; or nil
}

Note: If nReturnVals == 0, the return statement is not allowed
to have arguments after it (because the method declaration declared named return arguments or no return args.)

*/
func (p *parser) parseReturnStatement(stmt **ast.ReturnStatement, nReturnVals int, isYield bool) bool {
    if p.trace {
       defer un(trace(p, "ReturnStatement"))
    }

    pos := p.Pos()

    var hasSpace bool
    if isYield { // There is no => arrow. Maybe eventually use a -> arrow???????? or <- 
	   hasSpace = true
    } else {
       if ! p.Match2('=','>') {
	      return false
       }
       hasSpace = p.Space()
    }
    if ! hasSpace {
	    if nReturnVals > 0 {
	       p.required(false,"a single space after '=>' followed by values to return from method")
	    }
    }
    isInsideBrackets := false

    var xs []ast.Expr

    var foundResultExprs bool
    if hasSpace {
       foundResultExprs = (p.parseLenientVerticalExpressions(p.Col(),&xs) || 
                           p.parseMultipleOneLineExpressions(&xs,isInsideBrackets) || 
                           p.parseSingleOneLineExpression(&xs,isInsideBrackets))
    }
    if foundResultExprs {
	   if nReturnVals == 0 {
	       p.stop("=> val... statement cannot appear here, because this method declares no anonymous result values")	
  	   } 
    } else {
	    if nReturnVals > 0 {
		   p.stop("=> must be followed by a literal value or constant or variable/attribute reference or method call")
	   }
    }

    // translate
    *stmt = &ast.ReturnStatement{pos,xs, isYield}

    return true
}


/*
AssignmentStatement struct {
	Lhs    []Expr
	TokPos token.Pos   // position of Tok
	Tok    token.Token // assignment token, DEFINE
	Rhs    []Expr
}
*/
func (p *parser) parseAssignmentStatement(stmt **ast.AssignmentStatement) bool {
    if p.trace {
       defer un(trace(p, "AssignmentStatement"))
    }
    st := p.State()

    var lhsList []ast.Expr

    if ! p.parseLHSList(false, &lhsList) {
	   return false
    }

    operatorPos := p.Pos() + 1
 
    var operator token.Token 
	
    if p.Match2(' ','=') {
		operator = token.ASSIGN	
	} else if p.Match(" +=") {
		operator = token.ADD_ASSIGN
	} else if p.Match(" -=") {
		operator = token.SUB_ASSIGN
	} else if p.Match(" <-") {
		operator = token.ARROW	
		if len(lhsList) > 1 {
 	       p.stop("Can only send data to one channel with a '<-' operator.")				
		}
    } else {
	    return p.Fail(st)
    }

    for _,expr := range lhsList {
	   switch expr.(type) {
	      case *ast.Ident:
		     variable := expr.(*ast.Ident)
             p.ensureCurrentScopeVariable(variable, false)
	      default: 
		    dbg.Logln(dbg.PARSE_,"------------IS NOT A VARIABLE ---------------")
		    dbg.Logln(dbg.PARSE_,expr)
       }
    }   

    isInsideBrackets := false

    var xs []ast.Expr

    if ! p.parseIndentedExpressions(st.RuneColumn, &xs) {
	   p.required(p.Space(),"a space followed by an expression")
	   if ! p.parseLenientVerticalExpressions(p.Col(), &xs) {    // must have at least one second-line element
		   p.required(p.parseMultipleOneLineExpressions(&xs,isInsideBrackets) ||
		              p.parseSingleOneLineExpression(&xs,isInsideBrackets), "an expression")
	   }
    }

    // translate
    *stmt = &ast.AssignmentStatement{Lhs:lhsList,TokPos:operatorPos,Tok:operator,Rhs:xs}

    // TODO Need to match number of expression return values with number of lhs's and type check.

    return true
}



/*
Parse a list of left-hand-side (of an assignment statement) expressions,
which can include local variables, method parameters, constants, or indexed expressions.
*/
func (p *parser) parseLHSList(mustBeLocalVar bool, lhsList *[]ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "LHSList"))
    }

    var x ast.Expr

    if ! p.parseOneLineVariableOrConstReference(true, mustBeLocalVar, true, false, &x) {
	    return false
    }

    // translate
	*lhsList = append(*lhsList,x)
	
    st2 := p.State()
    for p.Space() && p.parseOneLineVariableOrConstReference(true, mustBeLocalVar, true, false, &x) {
	
	   // translate
	   *lhsList = append(*lhsList,x)
	
	   st2 = p.State()
    } 
    p.Fail(st2)
    return true
}



/*
Parses a keyword parameter assignment statement (in a method call)
e.g.
x = expr

*/
func (p *parser) parseKeywordParameterAssignmentStatement(paramName *string, rhs *ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "KeywordParameterAssignmentStatement"))
    }
    st := p.State()

    if ! p.parseKeywordParameterName(paramName) {
	   return false
    }
	
    if ! p.Match2(' ','=') {
	    return p.Fail(st)
    }

    isInsideBrackets := false

    var xs []ast.Expr

    if ! p.parseIndentedExpressions(st.RuneColumn, &xs) {
	   p.required(p.Space(),"a space followed by an expression")
	   if ! p.parseLenientVerticalExpressions(p.Col(), &xs) {    // must have at least one second-line element
		   p.required(p.parseSingleOneLineExpression(&xs,isInsideBrackets), "an expression")
	   }
    }
    if len(xs) > 1 {
 	   p.stop(fmt.Sprintf("Cannot assign multiple expressions to the single keyword parameter %s.",*paramName))	
    }

    // translate
    *rhs = xs[0]


    return true
}




// ----------------------------------------------------------------------------
// Declarations








/*
   A group of constants with no blank lines between them.
*/
func (p *parser) parseConstantDeclarationBlock(constDecls *[]*ast.ConstantDecl) bool {
	if p.trace {
       defer un(trace(p, "ConstantDeclarationBlock"))
    }
    if ! p.parseConstantDeclaration(constDecls) {
	   return false
	}
	for {
	   st2 := p.State()
	   if ! p.Below(1) {
	       break	
	   }
	   if ! p.parseConstantDeclaration(constDecls) {
		   p.Fail(st2)
		   break
	   }
    }
	return true
}


/*
Note. Constants must be deeply immutable. Do complex object values assigned to a constant become frozen?
*/
func (p *parser) parseConstantDeclaration(constDecls *[]*ast.ConstantDecl) bool {
	if p.trace {
       defer un(trace(p, "ConstantDeclaration"))
    }

    st := p.State()

    var constant *ast.Ident
    var expr ast.Expr

    if ! p.parseConstName(false,&constant) {
	   return false
    }

    if ! p.required(p.Match2(' ','='),"= after one space") {
    	return p.Fail(st)
    }

//    if p.Space() {
//	    p.required(p.parseConstExpr(&expr) || (p.Indent(st.RuneColumn) && p.parseConstExpr(&expr)),"a constant expression (a literal, or a function of literals and constants) on the same line or indented below.")
//   } else {
//	    p.required(p.Indent(st.RuneColumn) && p.parseConstExpr(&expr),"a constant expression (a literal, or a function of literals and constants) on the same line or indented below.")
//    }

	if p.Space() {
	    p.required(p.parseExpression(&expr) || (p.Indent(st.RuneColumn) && p.parseExpression(&expr)),"a constant expression (a literal, or a function of literals and constants) on the same line or indented below.")
	} else {
	    p.required(p.Indent(st.RuneColumn) && p.parseExpression(&expr),"a constant expression (a literal, or a function of literals and constants) on the same line or indented below.")
	}

    constDecl := &ast.ConstantDecl{Name:constant,Value:expr}

    *constDecls = append(*constDecls, constDecl)
	return true
}




/*
Polygon <: Shape2D ClosedCurve  // Only if all types are simple single word (non-parameterized)
"""
"""

Polygon 
<:
   Shape2D
   ClosedCurve
"""
"""

Tree of T 
"""
"""

RedBlackTree of T
<: 
   Tree of T




Tree of 
   T <: HasEquality


RedBlackTree of
   T <: HasEquality
<: 
   Tree of
      T <: HasEquality


WeightedPair of T1 T2
<:
   Pair of T1 T2

WeightedPair of 
   T1 
   T2
<:
   Pair of 
      T1 
      T2

a.b (in a setter/getter implementation)  refers to the b getter method, whereas
b refers to the raw attribute.

*/
func (p *parser) parseTypeDeclaration(typeDecls *[]*ast.TypeDecl) bool {
    if p.trace {
       defer un(trace(p, "TypeDeclaration"))
    }		
    var typeDecl *ast.TypeDecl
    col := p.Col()
    if p.parseTypeHeader(&typeDecl) &&
       p.optional(p.parseTypeBody(col,typeDecl)) {
		
	    *typeDecls = append(*typeDecls,typeDecl)
	    return true
    } 
    return false	
	
}

/*
Parse the type name, type parameters, and supertypes part of a type declaration, but not 
the body that contains attributes, getters, and setters. 
*/
func (p *parser) parseTypeHeader(typeDecl **ast.TypeDecl) bool {
    if p.trace {
       defer un(trace(p, "TypeHeader"))
    }
    st := p.State()
    col := st.RuneColumn

    var typeName *ast.Ident

    if ! p.parseTypeName(false,&typeName) {
	   return false
    }

    typeSpec := &ast.TypeSpec{Name: typeName}

	//    p.clearVariableScope() // start collecting arg and local variable names - do we need the analogous in a type decl?


	*typeDecl = &ast.TypeDecl{Spec: typeSpec}

    return (p.optional(
	                     p.parseOneLineSupertypes(typeSpec) ||
                         (  p.optional(p.parseTypeParameters(col, typeSpec)) &&                  
	                        p.optional(p.parseVerticalSupertypes(col,typeSpec)) )) &&
            p.required(p.parseTypeComment(col,*typeDecl),fmt.Sprintf("type header comment beginning with \"\"\" at column %v",col)) ) || 	
	       p.Fail(st)
}



/*
A list of supertypes all on one line, in a type declaration.    
*/
func (p *parser) parseOneLineSupertypes(typeSpec *ast.TypeSpec) bool {
	if p.trace {
	   defer un(trace(p, "OneLineSupertypes"))
	}	
	
	if ! p.Match(" <:") {
		return false
	}
	
	p.required(p.Space(),"a space then a type name after <:")
	
    var superTypeSpec *ast.TypeSpec

    p.required( p.parseTypeSpec(false, false,false,false,false,false,&superTypeSpec), "a type name." )

    // translate
    typeSpec.SuperTypes = append(typeSpec.SuperTypes, superTypeSpec)

    st2 := p.State()

    for p.Space() {
	   if p.parseTypeSpec(false, false,false,false,false,false,&superTypeSpec) {
		
		   // translate
           typeSpec.SuperTypes = append(typeSpec.SuperTypes, superTypeSpec)		
	    } else {
	      	p.Fail(st2)
	        break
	   }	
       st2 = p.State()	
    } 	

	
	// TODO Create the ast nodes. !!!!!!!
	
	return true
}




/*
<:
   SuperType1 of 
      T
      T2
   SuperType2
*/
func (p *parser) parseVerticalSupertypes(col int, typeSpec *ast.TypeSpec) bool {
	if p.trace {
	   defer un(trace(p, "VerticalSupertypes"))
	}	
	
    st := p.State()

    if ! p.Below(col) {
	   return false
    }
	
	if ! p.Match("<:") {
		return p.Fail(st)
	}
	
    p.required(p.Indent(col), "a type name, below <: and indented")


    var superTypeSpec *ast.TypeSpec

    p.required(p.parseTypeSpec(true, false,true, true, false, false, &superTypeSpec), "a type specification" )	

    // translate
    typeSpec.SuperTypes = append(typeSpec.SuperTypes, superTypeSpec)
	
    for p.Indent(col) {
       p.required(p.parseTypeSpec(true, false,true, true, false, false, &superTypeSpec),"a type specification")	

       // translate
       typeSpec.SuperTypes = append(typeSpec.SuperTypes, superTypeSpec)	
    }
	
	// TODO Create the ast nodes.
	
	return true
}


/*
Parses the mandatory """ comment at the top of a type declaration

TODO create the ast node for the comment. !!!!!!!!!
*/	
func (p *parser) parseTypeComment(col int, typeDecl *ast.TypeDecl) bool {
   if p.trace {
      defer un(trace(p, "TypeComment"))
   }	
   st := p.State()

   if ! p.Below(col) {
      return false
   }
   
   // db.Log(dbg.PARSE_,"Got here ch=%s\n",string(p.Ch()))	

   if ! ( p.Match(`"""`) &&
          p.required(p.BlankToEOL(),`nothing on line after """`) ) {
       return p.Fail(st)
   }
   st2 := p.State()
 
   if ! p.required(p.BlanksAndBelow(2,false),"comment content - Must begin at column 2 of file") {
       return p.Fail(st)    	
   }

   found,contentEndOffset := p.ConsumeTilMatchAtColumn(`"""`,1)
   if ! found {
	  dbg.Logln(dbg.PARSE_,`Did not consume till """.`)	
      return p.Fail(st)	
   }
   if ! p.required(p.BlankToEOL(),`nothing on line after """`) {
      return p.Fail(st)	
   }

   commentContent := p.Substring(st2.Offset,contentEndOffset) 	
   // TODO
   // Check the content to make sure none of it is in the first column.
   // Also, produce the actual content string, with first column removed.
   dbg.Logln(dbg.PARSE_,"Comment Content:")
   dbg.Logln(dbg.PARSE_,commentContent)
   dbg.Logln(dbg.PARSE_,"END Comment Content")
   return true
}
		
	


/*
Car

   < wheels  []Wheel
   driver    Person
   passenger []Person
   < trunk   Cavity


Car

 < wheels    []Wheel
   driver    Person
   passenger []Person
 < trunk     Cavity


Car

   wheels    []Wheel 
   """
    < This is the wheels.
   """

   driver    Person
   passenger []Person
   trunk     Cavity


   Don't handle getters and setters yet.

   New default idea

 < attr1 String default "yeah!"  // This is a read-only attribute at this level

   New setter syntax idea

 > attr2 String = s default "yeah!"  // write only attribute at this level. Defaults to "yeah" until written to
   """
   """
      if checkThis s
         s = fixup minus s
      attr2 = s

   TODO Have to handle __clan__ __clan_or_kin__ __kin__ __private__ sections within the type declaration.
*/
func (p *parser) parseTypeBody(col int, typeDecl *ast.TypeDecl) bool {
    if p.trace {
       defer un(trace(p, "TypeBody"))
    }	

    // parse
    st := p.State()

    var attrs []*ast.AttributeDecl

    if p.BlanksAndIndent(col,true) {
       if ! p.parseReadWriteAttributeDecl(&attrs,"public") {
  	      return p.Fail(st)
       }
    } else if p.BlanksAndMiniIndent(col, true) {
       if ! (p.parseReadOnlyAttributeDecl(&attrs,"public") || p.parseWriteOnlyAttributeDecl(&attrs,"public")) {	
	      return p.Fail(st)
       }	
    } else {
	    return false
    }

    for {
	    if p.BlanksAndIndent(col, true) {
	       p.required(p.parseReadWriteAttributeDecl(&attrs,"public"),"an attribute declaration")
	    } else if p.BlanksAndMiniIndent(col, true) {
	       p.required(p.parseReadOnlyAttributeDecl(&attrs,"public") || p.parseWriteOnlyAttributeDecl(&attrs,"public"),"an attribute declaration starting with < or >") 	
	    } else {
		    break
	    }	
    }

    typeDecl.Attributes = attrs

    // TODO Need to check for repeated attributes, ambiguous path overriding attr name problems etc

    return true   
}

/*
   attrName SomeType
*/
func (p *parser) parseReadWriteAttributeDecl(attrs *[]*ast.AttributeDecl, visibilityLevel string) bool {
	return p.parseAttributeDecl(attrs,true,true,visibilityLevel)
}

func (p *parser) parseReadOnlyAttributeDecl(attrs *[]*ast.AttributeDecl, visibilityLevel string) bool {
    st := p.State()	
    if ! p.Match2('<',' ') {
	   return false
    }	
	return p.parseAttributeDecl(attrs,true,false,visibilityLevel)	|| p.Fail(st)
}

func (p *parser) parseWriteOnlyAttributeDecl(attrs *[]*ast.AttributeDecl,visibilityLevel string) bool {
    st := p.State()	
	if ! p.Match2('>',' ') {
		return false
	}
	return p.parseAttributeDecl(attrs,false,true,visibilityLevel)	|| p.Fail(st)
}


func (p *parser) parseAttributeDecl(attrs *[]*ast.AttributeDecl, read, write bool, visibilityLevel string) bool {
	
    var attrName *ast.Ident

    st := p.State()

    if ! p.parseVarName(&attrName,false) { // Do we eventually make this optional (stuttering avoidance default? Maybe not.)
	   return false
    }

    if ! p.Space() {
        return p.Fail(st)
    }

    var aritySpec *ast.AritySpec

    // Later we must check for 1 as maxArity which is not allowed in this context.
    forceCollection := p.parseAritySpec(&aritySpec)

   
    // Note: The {} or {<} or [] or [<] or {<width} etc are part of a type spec. 

    var typeSpec *ast.TypeSpec
    if ! p.parseTypeSpec(true, true,true,true,false,forceCollection,&typeSpec) {
	    return p.Fail(st)
    }

	attr := &ast.AttributeDecl{Name:attrName, Arity:aritySpec, Type:typeSpec}
	
	// Here, we have to check whether the name occurs already in this type's attribute scope.
	// If it does, there are legal and illegal ways that can happen.
	// Illegal is two of same name at same visibility level, unless one is the getter method definition
	// and one is the setter definition.
	//
	//   attr1 Int
	//   """
	//    In a getter declaration, the return the value of the attribute statement is always implicit.
	//    and always at the end. But you could use a zero-arg => to return before bottom of code.
	//   """
	//      if isZero attr1
	//         attr1 = 1  
	//       
	//   attr1 Int = i           
	//   """ 
	//    Question: Do we say the attr1 = i is implicit at end of code except you can use a defer to get 
	//    post processing done if you want? May be. I don't think the attr1 = i should be implicit at all.
	//    We need to make it really clear when it happens wrt other code in the setter.
	//    Does it have to be attr1 = i or can it be attr1 = anything at all? Anything.
	//   """
	//      if odd i
	//          panic "Can only assign an even number to attr1."
	//      attr1 = i
	//   
	// Also illegal is to to have two declarations of the attribute which are of different type.
	// Legal is to have a read at a more public level then a write at a less public level, or
	// vice versa.
	// (Note: cannot declare the missing operation at private level if it has no getter/setter for the
    //  private operation, because that private operation is implicit.)
    // 
    // Illegal to redeclare the same operation type or access type at two different visibility levels.
    //
    // Illegal to define a getter like this
    // > attr1 Int  
	//     if ...
	// 
	// because we said it was write-only at that level.
	//
	// Analogously, illegal to define < attr1 Int = i
	//
	// In the lower visibility level, when redeclaring the other allowed operation, do we need to include
	// the < or > ? It may be clearer, even though it is redundant. It tells us maybe we should go look for
	// where the other operation is declared, at another visibility level.
	//
	// TODO
	
	*attrs = append(*attrs,attr)
	
	
    // p.ensureCurrentScopeVariable(argName, false)	// should change this to refer to the input arg ast node. 
	
	return true
}

/*
   Temporary implementation - need to handle indented type spec

func (p *parser) parseAttributeDecl(attrs *[]*ast.AttributeDecl, read, write bool, visibilityLevel string) bool {
	
   return p.parseOneLineAttributeDecl(attrs,read,write,visibilityLevel)
}
*/









/*
Note: A method with no input arguments must have an implementation.
*/
func (p *parser) parseMethodDeclaration(methodDecls *[]*ast.MethodDeclaration) bool {
    if p.trace {
       defer un(trace(p, "MethodDeclaration"))
    }	

    var methodDecl *ast.MethodDeclaration 
    col := p.Col()
    if p.parseMethodHeader(&methodDecl) {
       foundMethodBody := p.parseMethodBody(col,methodDecl)
	
	    // if no input params, must have a method body - abstract method w no params does not make sense
        if len(methodDecl.Type.Params) == 0 {
	       p.required(foundMethodBody, "method body statements")	
        }
	
 	    if p.parsingClosure {
	       // A closure declaration must have some method body statements
	       p.required(foundMethodBody, "method body statements")

           p.fixUpFreeVarOffsets(p.currentScopeVariableOffset)	
           methodDecl.NumFreeVars = len(p.closureFreeVarBindings)
           methodDecl.IsClosureMethod = true
        }
	    methodDecl.NumLocalVars = p.currentScopeVariableOffset - 3 - len(methodDecl.Type.Params)
		
	    *methodDecls = append(*methodDecls,methodDecl)
	

	    return true
    } 
    return false
}


/*

foo a Int b Int > Int er.Error
"""
"""

foo a Int b Int 
> Int er.Error
"""
"""

foo a Int b Int 
> 
   Int 
   er.Error
"""
"""

foo 
   a Int 
   b Int 
> 
   Int 
   er.Error
"""
"""

argTest
   a Float
   b String
   bar Int = 0
   ...v String
"""
This declaration demonstrates keyword parameters (ie those with default values)
and a variadic argument at the end of the argument list.
"""



Note: A method with no arguments must have an implementation.

*/
func (p *parser) parseMethodHeader(methodDecl **ast.MethodDeclaration) bool {
    if p.trace {
       defer un(trace(p, "MethodHeader"))
    }
    st := p.State()
    col := st.RuneColumn
    pos := p.Pos()

    var methodName *ast.Ident

    if p.parsingClosure {
	   if ! p.MatchWord("func") {
	      return false	
	   } else {
		  // Generate a unique-in-package name for the method.
		  p.closureMethodName = fmt.Sprintf("Func__%s__%d", p.file.Name(), p.currentClosureMethodNum)
		  p.currentClosureMethodNum++
          methodName = &ast.Ident{pos, p.closureMethodName, nil, token.FUNC, -1}	
	   }
    } else if ! p.parseMethodName(false,&methodName) {
	   return false
    }

    p.clearVariableScope() // start collecting arg and local variable names

    *methodDecl = &ast.MethodDeclaration{Name:methodName}

    return (p.optional(p.parseOneLineArgSignature(col,*methodDecl) ||
                       p.parseIndentedArgSignature(col,*methodDecl) ||
                       p.parseReturnArgsOnlySignature(col,*methodDecl)  )  &&
            p.required(p.parseMethodComment(col,*methodDecl),fmt.Sprintf("method comment beginning with \"\"\" at column %v",col)) ) || 
           p.Fail(st)
}


/*
    Arg signature that starts with single-line input args and may
    have 1. return args on same line 2. return args all on next line or
    3. return args indented starting on next line.
*/
func (p *parser) parseOneLineArgSignature(col int, methodDecl *ast.MethodDeclaration) bool {
	if p.trace {
	   defer un(trace(p, "OneLineArgSignature"))
	}	
	
	funcType := &ast.FuncType{Func:methodDecl.Name.Pos(),AfterName:methodDecl.Name.End()}
	
/*
// A FuncType node represents a function type.
FuncType struct {
	Func    token.Pos  // position of subroutine name or "lambda" keyword
	Params  []*InputArgDecl // input parameter declarations. Can be empty list.
	Results []*ReturnArgDecl // (outgoing) result declarations; Can be empty list.
}
*/
	
	if ! p.parseOneLineInputArgSignature(funcType) {
		return false
	} 
    st2 := p.State()		
    foundOneLineReturnArgSignature := false
	if p.Space() {
		if p.parseOneLineReturnArgSignature(funcType) {
			foundOneLineReturnArgSignature = true
		} else {
			p.Fail(st2)
		}
	}
	if ! foundOneLineReturnArgSignature {
		if p.Below(col) {
			if ! (p.parseOneLineReturnArgSignature(funcType) || p.parseIndentedReturnArgSignature(col,funcType)) {
				p.Fail(st2)
			}
		}
	}
	
	methodDecl.Type = funcType
	
	return true
}




/*
    Arg signature that starts with indented input args and may
    have return args indented starting on next line.
*/
func (p *parser) parseIndentedArgSignature(col int, methodDecl *ast.MethodDeclaration) bool {
	if p.trace {
	   defer un(trace(p, "IndentedArgSignature"))
	}	

	funcType := &ast.FuncType{Func:methodDecl.Name.Pos(),AfterName:methodDecl.Name.End()}
	
	if ! p.parseIndentedInputArgSignature(col,funcType) {
		return false
	} 
    st2 := p.State()		
	if p.Below(col) {
		if ! p.parseIndentedReturnArgSignature(col,funcType) {
			p.Fail(st2)
		}
	}
	
	methodDecl.Type = funcType	
	return true	
}

/*
    Arg signature that has no input args and has
    1. return args on same line 2. return args all on next line or
    3. return args indented starting on next line.
*/
func (p *parser) parseReturnArgsOnlySignature(col int, methodDecl *ast.MethodDeclaration) bool {
	if p.trace {
	   defer un(trace(p, "ReturnArgsOnlySignature"))
	}	
	
    st := p.State()		
    foundReturnArgSignature := false

	funcType := &ast.FuncType{Func:methodDecl.Name.Pos(),AfterName:methodDecl.Name.End()}
	
	if p.Space() {
		if p.parseOneLineReturnArgSignature(funcType) {
			foundReturnArgSignature = true
		} else {
			if ! (p.Space() || p.Match1('\n')) {
         p.Fail(st)
      }
		}
	}
	if ! foundReturnArgSignature {
		if p.Below(col) {
			if p.parseOneLineReturnArgSignature(funcType) || p.parseIndentedReturnArgSignature(col,funcType) {
				foundReturnArgSignature = true
			} else {
				p.Fail(st)
			}
		}
	}
	
	methodDecl.Type = funcType	
		
	return foundReturnArgSignature
}


/*
   Adds the argument declarations to the function type ast node.
*/
func (p *parser) parseOneLineInputArgSignature(funcType *ast.FuncType) bool {
    if p.trace {
       defer un(trace(p, "OneLineInputArgSignature"))
    }	
    st := p.State()
    if ! p.Space()  {
	   return false
    }

    isKeywordDefaultedParam := false
    isVariadicKeywordsParam := false
    isVariadicListParam := false
    foundVariadicKeywordsParam := false
    foundVariadicListParam := false

    var inputArgDecls []*ast.InputArgDecl

    if ! p.parseOneLineInputArgDecl(&inputArgDecls,&isKeywordDefaultedParam, &isVariadicKeywordsParam, &isVariadicListParam) {
	   return p.Fail(st)
    }

    foundVariadicKeywordsParam = foundVariadicKeywordsParam || isVariadicKeywordsParam
    foundVariadicListParam = foundVariadicListParam || isVariadicListParam

    st2 := p.State()
    for ; p.Space(); {
	
	   if ! p.parseOneLineInputArgDecl(&inputArgDecls,&isKeywordDefaultedParam, &isVariadicKeywordsParam, &isVariadicListParam) {
	       	p.Fail(st2)
            funcType.Params = inputArgDecls	
	        return true
	   }
	
       if foundVariadicListParam {
          p.stop("Cannot declare another input parameter after the variadic list parameter declaration")	
       }	

	   st2 = p.State()
	
	   if isVariadicListParam { // This has to be the last input parameter declaration.
	      continue	
	   }

	   if foundVariadicKeywordsParam {
	      p.stop("Only the variadic list parameter can be declared after the variadic keywords map parameter")	
	   }


	   foundVariadicKeywordsParam = foundVariadicKeywordsParam || isVariadicKeywordsParam
       foundVariadicListParam = foundVariadicListParam || isVariadicListParam		
    }

    funcType.Params = inputArgDecls

    return true
}

/*
Allowed is:
[positionalParam]...
[defaultedKeywordParam]...
[variadicKeywordsParam]
[variadicListParam]                          

a Int
b String
c String = "Foo"
d Int = 0
...e {} String=>Int
...f [] Object
 
*/
func (p *parser) parseIndentedInputArgSignature(col int,funcType *ast.FuncType) bool {
    if p.trace {
       defer un(trace(p, "IndentedInputArgSignature"))
    }
    st := p.State()
    if ! p.Indent(col)  {
	   return false
    }

    isKeywordDefaultedParam := false
    isVariadicKeywordsParam := false
    isVariadicListParam := false
    foundKeywordDefaultedParam := false
    foundVariadicKeywordsParam := false

    var inputArgDecls []*ast.InputArgDecl

    if ! p.parseInputArgDecl(&inputArgDecls, &isKeywordDefaultedParam, &isVariadicKeywordsParam, &isVariadicListParam) {
	   return p.Fail(st)
    }

    foundKeywordDefaultedParam = foundKeywordDefaultedParam || isKeywordDefaultedParam
    foundVariadicKeywordsParam = foundVariadicKeywordsParam || isVariadicKeywordsParam

    st2 := p.State()
    for p.Indent(col) {
	   
	   if isVariadicListParam {
	      p.stop("Cannot declare any more input parameters after declaration of variadic list parameter")	
	   }
	 
	   if ! p.parseInputArgDecl(&inputArgDecls,&isKeywordDefaultedParam, &isVariadicKeywordsParam, &isVariadicListParam) {
	       	p.Fail(st2)
            funcType.Params = inputArgDecls	
	        return true
	   }
	
	   if isVariadicListParam { // This has to be the last input parameter declaration.
	      continue	
	   }
	
	   if foundVariadicKeywordsParam {
	      p.stop("Only the variadic list parameter can be declared after the variadic keywords map parameter")	
	   }
	
	   if foundKeywordDefaultedParam && ! (isKeywordDefaultedParam || isVariadicKeywordsParam) {
	      p.stop("Cannot declare an ordinary non-defaulted, non-variadic parameter after a defaulted parameter declaration")	
	   }
	
       foundKeywordDefaultedParam = foundKeywordDefaultedParam || isKeywordDefaultedParam
       foundVariadicKeywordsParam = foundVariadicKeywordsParam || isVariadicKeywordsParam	
	   
	   st2 = p.State()
    }

    funcType.Params = inputArgDecls

    return true
}


func (p *parser) parseOneLineInputArgDecl(inputArgs *[]*ast.InputArgDecl, isKeywordDefaultedParam *bool, isVariadicKeywordsParam *bool, isVariadicListParam *bool) bool {
	
    var argName *ast.Ident

    *isKeywordDefaultedParam = false  // These are not allowed to be expressed in a one line input args declaration.
    *isVariadicKeywordsParam = false 
    *isVariadicListParam = false
    isVariadic := false

    st := p.State()

    if p.Match("...") {
       isVariadic = true
    }

    if ! p.parseVarName(&argName,false) { // Do we eventually make this optional (stuttering avoidance default? Maybe not.)
	   if isVariadic {
	      p.stop("Expecting a parameter name")	
	   } else {
	      return false
       }
    }

    if ! p.Space() {
        return p.Fail(st)
    }

    var typeSpec *ast.TypeSpec
    if ! p.parseTypeSpec(false, true,true,true,false,false,&typeSpec) {
	    return p.Fail(st)
    }

	inputArg := &ast.InputArgDecl{argName,typeSpec,nil, isVariadic}
	
	*inputArgs = append(*inputArgs,inputArg)
	
    p.ensureCurrentScopeVariable(argName, false)	// should change this to refer to the input arg ast node. 
	
	return true
}

/*
   Temporary implementation - need to handle indented type spec

   isKeywordDefaultedParam means

   a Int = 2
   
   isVariadicKeywordsParam means

   ...a {} String=>Int

   isVariadicListParam means

   ...a [] Int

*/
func (p *parser) parseInputArgDecl(inputArgs *[]*ast.InputArgDecl, isKeywordDefaultedParam *bool, isVariadicKeywordsParam *bool, isVariadicListParam *bool) bool {
	
   // OBSOLETE return p.parseOneLineInputArgDecl(inputArgs)

   // THIS IS THE ONE LINE VERSION

    var argName *ast.Ident

    *isKeywordDefaultedParam = false  
    *isVariadicKeywordsParam = false 
    *isVariadicListParam = false 
    isVariadic := false

    st := p.State()

    if p.Match("...") {
	   isVariadic = true
    }

    if ! p.parseVarName(&argName,false) { // Do we eventually make this optional (stuttering avoidance default? Maybe not.)
	   if isVariadic {
	      p.stop("Expecting a parameter name")	
	   } else {
	      return false
       }
    }

    if ! p.Space() {
        return p.Fail(st)
    }



    var typeSpec *ast.TypeSpec
    if ! p.parseTypeSpec(true,true,true,true,false,false,&typeSpec) {
	    return p.Fail(st)
    }

    if isVariadic {
       if typeSpec.CollectionSpec != nil {
           if typeSpec.CollectionSpec.Kind == token.LIST {
	          *isVariadicListParam = true	 
	          dbg.Logln(dbg.PARSE_,"SETTING *isVariadicListParam to true")         
	       } else if typeSpec.CollectionSpec.Kind == token.MAP {
	          *isVariadicKeywordsParam = true		
      	   } else {
	          p.stop("A variadic parameter must be a list or a map")
           }
	   } else {
	       p.stop("A variadic parameter must be a list or a map")		
	   }
    }

    var defaultValue ast.Expr = nil

    // Parse the default assignment expression if any

    if p.Match(" = ") {
        p.required(p.parseExpression(&defaultValue),"an expression")   
        *isKeywordDefaultedParam = true
    }

	inputArg := &ast.InputArgDecl{argName,typeSpec, defaultValue, isVariadic}
	inputArg.IsVariadic = isVariadic
	
	*inputArgs = append(*inputArgs,inputArg)
	
    p.ensureCurrentScopeVariable(argName, false)	// should change this to refer to the input arg ast node. 
	
	return true

}


func (p *parser) parseOneLineReturnArgSignature(funcType *ast.FuncType) bool {
    if p.trace {
       defer un(trace(p, "OneLineReturnArgSignature"))
    }	

    st := p.State()

    if ! p.Match1('>') {
       return false
    }

    if ! p.Space()  {
	   return p.Fail(st)
    }

   var returnArgDecls []*ast.ReturnArgDecl

   if ! p.parseOneLineReturnArgDecl(&returnArgDecls) {
	   return p.Fail(st)
   }

   st2 := p.State()
   for ; p.Space(); {
	   if ! p.parseOneLineReturnArgDecl(&returnArgDecls) {
	       	p.Fail(st2)
          break
	   }
	   st2 = p.State()
   }

   // Loop through the return arg declarations.
   // Check that either they all have variable names or none do.
   // If they have variable names, assign stack-offsets to the idents.
   hasOneReturnArgName := false
   hasAllReturnArgNames := true
   for i := len(returnArgDecls)-1; i >= 0; i-- {
       returnArgDecl := returnArgDecls[i]
       if returnArgDecl.Name == nil {
           hasAllReturnArgNames = false
       } else { 
           hasOneReturnArgName = true
           p.ensureCurrentScopeVariable(returnArgDecl.Name, true)              
       }
   }
   if hasOneReturnArgName && ! hasAllReturnArgNames {
        p.stop("If one return value declaration has a variable name, all must be named")
        return false
   }

   funcType.Results = returnArgDecls  
   return true

}

func (p *parser) parseIndentedReturnArgSignature(col int, funcType *ast.FuncType) bool {
    if p.trace {
       defer un(trace(p, "IndentedReturnArgSignature"))
    }	

    st := p.State()

    if ! p.Match1('>') {
       return false
    }

    if ! p.Indent(col) {
        return p.Fail(st)
    }

    var returnArgDecls []*ast.ReturnArgDecl

    if ! p.parseReturnArgDecl(&returnArgDecls) {
       return p.Fail(st)
    }

    st2 := p.State()
    for p.Indent(col) {

       if ! p.parseReturnArgDecl(&returnArgDecls) {
            p.Fail(st2)
            break
       }     
       st2 = p.State()
    }

    // Loop through the return arg declarations.
    // Check that either they all have variable names or none do.
    // If they have variable names, assign stack-offsets to the idents.
    hasOneReturnArgName := false
    hasAllReturnArgNames := true
    for i := len(returnArgDecls)-1; i >= 0; i-- {
       returnArgDecl := returnArgDecls[i]
       if returnArgDecl.Name == nil {
           hasAllReturnArgNames = false
       } else { 
           hasOneReturnArgName = true
           p.ensureCurrentScopeVariable(returnArgDecl.Name, true)              
       }
    }
    if hasOneReturnArgName && ! hasAllReturnArgNames {
        p.stop("If one return value declaration has a variable name, all must be named")
        return false
    }

    funcType.Results = returnArgDecls  
    return true
}


func (p *parser) parseOneLineReturnArgDecl(returnArgs *[]*ast.ReturnArgDecl) bool {
	
    var argName *ast.Ident

    st := p.State()

    if p.parseVarName(&argName,false) { 
       if ! p.Space() {
           return p.Fail(st)
       }
   }

    var typeSpec *ast.TypeSpec
    // Should I allow maybe? Yes probably.
    if ! p.parseTypeSpec(false, false,true, true, false, false, &typeSpec) {
	    return p.Fail(st)
    }

	returnArg := &ast.ReturnArgDecl{argName,typeSpec}
	
	*returnArgs = append(*returnArgs,returnArg)
	
// Now doing this in the function that parses the whole return arg signature.  
//    if argName != nil {
//       p.ensureCurrentScopeVariable(argName, true)	 	
//	  }
	return true
}

/*
   Temporary implementation - need to handle indented type spec
*/
func (p *parser) parseReturnArgDecl(returnArgs *[]*ast.ReturnArgDecl) bool {
	
   return p.parseOneLineReturnArgDecl(returnArgs)
}





/*
Parses the mandatory """ comment at the top of a method declaration.

TODO create the ast node for the comment. !!!!!!!!!
*/	
func (p *parser) parseMethodComment(col int, methodDecl *ast.MethodDeclaration) bool {
   if p.trace {
      defer un(trace(p, "MethodComment"))
   }	
   st := p.State()

   if ! p.Below(col) {
      return false
   }
   
   // dbg.Log(dbg.PARSE_,"Got here ch=%s\n",string(p.Ch()))	

   if ! ( p.Match(`"""`) &&
          p.required(p.BlankToEOL(),`nothing on line after """`) ) {
       return p.Fail(st)
   }
   st2 := p.State()
 
   if ! p.required(p.BlanksAndBelow(col+1, false),fmt.Sprintf("comment content - Must begin at column %d of file",col+1)) {
       return p.Fail(st)    	
   }

   found,contentEndOffset := p.ConsumeTilMatchAtColumn(`"""`,col)
   if ! found {
	  dbg.Logln(dbg.PARSE_,`Did not consume till """.`)	
      return p.Fail(st)	
   }
   if ! p.required(p.BlankToEOL(),`nothing on line after """`) {
      return p.Fail(st)	
   }

   commentContent := p.Substring(st2.Offset,contentEndOffset) 	
   // TODO   
   // Check the content to make sure none of it is in the first column.
   // Also, produce the actual content string, with first column removed.
   dbg.Logln(dbg.PARSE_,"Comment Content:")
   dbg.Logln(dbg.PARSE_,commentContent)
   dbg.Logln(dbg.PARSE_,"END Comment Content")
   return true
}



/*
Relations are declarations in program text of a relationship between two data types.
They are expressed in a textual version of the two boxes joined by a relation line that
you would find in an Entity-Relation Diagram (ERD). e.g.

KeyFob 0 1 -- N Key
"""
 Can only tell this is a relation when see the digit or collection type token after the type.
 Looks just like a type declaration otherwise.
"""

House {} 0 2 neighbours -- aNeighbourOf 0 2 {} House

Does at least one of the types have to be declared in the same package? Or even both? Seems too limiting.
This question will need a lot of thought.
Perhaps you cannot demand "at least 1" when relating to something from another package.
Does the rel end name then need to be qualified by package name if used by another package?
*/	
func (p *parser) parseRelationDeclaration(relDecls *[]*ast.RelationDecl) bool {
    if p.trace {
       defer un(trace(p, "RelationDeclaration"))
    }	
	st := p.State()	
	pos := p.Pos()
	var rel1TypeName, rel2TypeName, rel1EndName, rel2EndName *ast.Ident
	var arity1Spec, arity2Spec *ast.AritySpec
	var collection1Spec,collection2Spec *ast.CollectionTypeSpec
	
	if ! (p.parseTypeName(true, &rel1TypeName) && p.Space() &&
	p.optional(p.parseCollectionTypeSpec(&collection1Spec) && p.Space()) &&
	p.parseAritySpec(&arity1Spec) &&
	p.optional(p.parseRelEndName(&rel1EndName) && p.Space()) &&
	p.Match("-- ") &&
	p.optional(p.parseRelEndName(&rel2EndName) && p.Space()) && 	
	p.parseAritySpec(&arity2Spec) &&	
	p.optional(p.parseCollectionTypeSpec(&collection2Spec) && p.Space()) &&
	p.parseTypeName(true, &rel2TypeName) ) {
	   return p.Fail(st)		
	}
    p.optional(p.parseMethodComment(1,nil))
	
	
  // Generate relation-end "attribute" names from (properly pluralized as needed) end-type names, if
  // end attribute names have been omitted from the relation declaration. 
  
	if rel1EndName == nil {
	    slashPos := strings.LastIndex(rel1TypeName.Name,"/") 
	    typ1Name := rel1TypeName.Name[slashPos+1:]
    	end1Name := strings.ToLower(typ1Name[0:1]) + typ1Name[1:]
     if arity1Spec.MaxCard != 1 {
        if strings.HasSuffix(end1Name,"s") {
           end1Name += "es"
        } else if endsInYafterConsonant(end1Name) {
	       end1Name = end1Name[:len(end1Name)-1] + "ies"
	    } else {
           end1Name += "s"
        }
     }
		 rel1EndName = &ast.Ident{pos, end1Name, nil, token.VAR, -99}
	}

  if rel2EndName == nil {
		slashPos := strings.LastIndex(rel2TypeName.Name,"/") 
	    typ2Name := rel2TypeName.Name[slashPos+1:]
		end2Name := strings.ToLower(typ2Name[0:1]) + typ2Name[1:]
     if arity2Spec.MaxCard != 1 {
        if strings.HasSuffix(end2Name,"s") {
           end2Name += "es"
        } else if endsInYafterConsonant(end2Name) {
	       end2Name = end2Name[:len(end2Name)-1] + "ies"
        } else {
           end2Name += "s"
        }
     }
     rel2EndName = &ast.Ident{pos, end2Name, nil, token.VAR, -99}
  }


  // Here we have to create a proper type specification for each end, considering
  // whether we have a collectionTypeSpec already, and also considering the arity.MaxCard of each end.
  // If arity.MaxCard == 1, we should I suppose not create a collection type spec for that end.

  if arity1Spec.MaxCard == 1 { 
     if collection1Spec != nil {
        // Check for explicit [] or {} collection spec for an end that
        // has been declared max card 1, and forbid these.      
        p.stop("Cannot specify a relation-end collection type when max cardinality = 1")
     }
  } else { // max-arity of end1 is more than 1  
     if collection1Spec == nil {  // create a Set CollectionTypeSpec as a default
        collection1Spec = &ast.CollectionTypeSpec{token.SET,pos,pos+1,false,false,""}
     }
  }    

  // Create the end-type spec. Note that the CollectionSpec may be nil
  type1Spec := &ast.TypeSpec{CollectionSpec: collection1Spec, Name: rel1TypeName}
  


  if arity2Spec.MaxCard == 1 { 
     if collection2Spec != nil {
        // Check for explicit [] or {} collection spec for an end that
        // has been declared max card 1, and forbid these.      
        p.stop("Cannot specify a relation-end collection type when max cardinality = 1")
     }
  } else { // max-arity of end2 is more than 1  
     if collection2Spec == nil {  // create a Set CollectionTypeSpec as a default
        collection2Spec = &ast.CollectionTypeSpec{token.SET,pos,pos+1,false,false,""}
     }
  }

  // Create the end-type spec. Note that the CollectionSpec may be nil
  type2Spec := &ast.TypeSpec{CollectionSpec: collection2Spec, Name: rel2TypeName}




  // But hold on, wouldn't it be simpler to always create a set if there's a missing collection spec,
  // even if the MaxCard is 1 ? Then we don't have to special-case card=1 relations.
  //
  // To decide this, we have to review how relations were to be represented in the DB.
  //



    end1Decl := &ast.AttributeDecl{Name:rel1EndName, Arity:arity1Spec, Type:type1Spec}

    end2Decl := &ast.AttributeDecl{Name:rel2EndName, Arity:arity2Spec, Type:type2Spec}

    relDecl := &ast.RelationDecl{end1Decl, end2Decl}

    *relDecls = append(*relDecls,relDecl)
    return true

}	

/*
Helper used in pluralization of automatically inferred relation-end attribute names.
*/
func endsInYafterConsonant(s string) bool {
	if ! strings.HasSuffix(s,"y") {
		return false
	}
	n := len(s) - 2
	if n < 0 {
		return false
	}
	c := s[n]
	if c == 'a' || c == 'e' || c == 'o' || c == 'u' {
		return false
	}
	return true
}


/*
   List empty literal []
   Set empty literal {}
   Map empty literal  {=>}
   foo {} String=>ClothingItem    
   foo {>size} ClothingItem 

/*
Simple version first. {} or [] 
Need to handle [>] [<] {>} {<} (uses natural order of type if defined - error otherwise)
[<field] [>field] [<func] [>func] 
*/
func (p *parser) parseCollectionTypeSpec(collectionTypeSpec **ast.CollectionTypeSpec) bool {
    if p.trace {
       defer un(trace(p, "CollectionTypeSpec"))
    }	

    return p.parseListTypeSpec(collectionTypeSpec) || p.parseSetTypeSpec(collectionTypeSpec)	
}

/*
Parse the short-hand notation for List of SomeType  e.g. List of String, or List of Widget.

[] String  // meaning a list of strings
[<] String  // meaning a sorting list of strings, sorted by the natural ordering of String.

[<weight] Widget  // meaning a sorting list of Widgets, ordered by the (natural ordering of the) 
                  // weight attribute of the widget.
*/
func (p *parser) parseListTypeSpec(collectionTypeSpec **ast.CollectionTypeSpec) bool {
    if p.trace {
       defer un(trace(p, "ListTypeSpec"))
    }	

    st := p.State()
    pos := p.Pos()

    if ! p.Match1('[') {
       return false
    }

	// TODO Have to do < and <attr and <foo
	var isSorting, isAscending bool
	var orderFunc string
	
	if p.Match("<") {
		isAscending = true
		isSorting = true
	} else if p.Match(">") {
        isSorting = true		
	}
    if isSorting {
	   _,orderFunc = p.ScanVarName() 
    }

    if ! p.Match1(']') {
       return p.Fail(st)
    }	
    end := p.Pos()

    *collectionTypeSpec = &ast.CollectionTypeSpec{token.LIST,pos,end,isSorting,isAscending,orderFunc}

    return true
}

/*
Parse the short-hand notation for Set of SomeType  e.g. Set of String, or Set of Widget.

{} String  // meaning a set of strings
{<} String  // meaning a sorting set of strings, sorted by the natural ordering of String.

{<weight} Widget  // meaning a sorting set of Widgets, ordered by the (natural ordering of the) 
                  // weight attribute of the widget.
*/
func (p *parser) parseSetTypeSpec(collectionTypeSpec **ast.CollectionTypeSpec) bool {
    if p.trace {
       defer un(trace(p, "SetTypeSpec"))
    }	

    st := p.State()
    pos := p.Pos()

    if ! p.Match1('{') {
       return false
    }

	// TODO Have to do < and <attr and <foo
	var isSorting, isAscending bool
	var orderFunc string
	
	if p.Match("<") {
		isAscending = true
		isSorting = true
	} else if p.Match(">") {
        isSorting = true		
	}
    if isSorting {
       _,orderFunc = p.ScanVarName() 
    }
	
    if ! p.Match1('}') {
       return p.Fail(st)
    }
    end := p.Pos()

    *collectionTypeSpec = &ast.CollectionTypeSpec{token.SET,pos,end,isSorting,isAscending,orderFunc}

    return true		
}

/*
At each end of the relation, the cardinality limits are specified.
The number of instances of the type at that end that must be associated with each instance
at the other end. Parse a relation-end cardinality limits specification, also known
as an arity specification.

1 means 1 1 
N  means 1 N 
Gobbles the single space after the second number.

minCard int64
maxCard int64  // -1 means N
pos int // start position in file of the first integer in the spec
end int // just after th

*/
func (p *parser) parseAritySpec(aritySpec **ast.AritySpec) bool {
    if p.trace {
       defer un(trace(p, "AritySpec"))
    }	
    st := p.State()	
    pos := p.Pos()
    var end token.Pos
    var ch rune

    var minCard int64
    var maxCard int64


	if p.Match2('N',' ') {
		minCard = 1
		maxCard = -1
	    end = p.Pos() - 1		
		goto Translate
	}
		
    ch = p.Ch()

    if ! scanner.IsAsciiDigit(ch)  {
	   return false
	}
	
	minCard = int64(ch - '0')
	
	p.Next()
	for {
	   ch = p.Ch()	
       if ! scanner.IsAsciiDigit(ch)  {
	      break
	   }	
	   minCard = minCard * 10 + int64(ch - '0')
	   p.Next()	
	}

	end = p.Pos()
		
    if ! p.Space() {
	   return p.Fail(st)
    }

  	if p.Match2('N',' ') {
		maxCard = -1	
  	    goto Translate	  
    }

    ch = p.Ch()

    if ! scanner.IsAsciiDigit(ch)  {
	   maxCard = minCard
	   dbg.Log(dbg.PARSE_,"succeeded on not a digit. '%s'\n",string(p.Ch()))	
	   goto Translate
    }	

	maxCard = int64(ch - '0')
	
	p.Next()
	for {
	   ch = p.Ch()	
       if ! scanner.IsAsciiDigit(ch)  {
	      break
	   }	
	   maxCard = maxCard * 10 + int64(ch - '0')
	   p.Next()	
	}
	
	end = p.Pos()
	
    if ! p.Space() {
	   dbg.Log(dbg.PARSE_,"not a space after second digits. '%s'\n",string(p.Ch()))
	   return p.Fail(st)
    }	   	

  Translate:

    if maxCard == 0 {
       p.stop("Maximum cardinality of a multi-valued attribute or relation-end cannot be zero")
    }
    *aritySpec = &ast.AritySpec{minCard,maxCard,pos,end}
    
    dbg.Logln(dbg.PARSE_,"successful arity spec parse.")
    return true
}




// ----------------------------------------------------------------------------
// Source files



/*
	
origin:
artifact:
package:

""" 
 lock.rel

 This file defines everything related to locks and keys and all the other things.
 This is limited to tumbler locks.
"""

KeyFob

Key

KeyFob 0 1 -- N Key
"""
 Can only tell this is a relation when see the digit or collection type token after the type.
 Looks just like a type declaration otherwise.
"""

House {} 0 2 neighbours -- aNeighbourOf 0 2 {} House
"""
A symmetric relationship like this how do we handle it?
"""

PI = 3.14159265357989

PI_SQUARED = 9.43

someMethod a Int b String > c Int e er.Error


*/
func (p *parser) parseFile() *ast.File {
	if p.trace {
		defer un(trace(p, "File"))
	}
	
	pos := p.Pos()
	
	var packageName string
	
    p.parseFileHeader(&packageName);

    pkgPos := strings.Index(packageName,"/pkg/")
    p.currentOriginAndArtifactName = packageName[:pkgPos]
    
    p.packagePath = packageName + "/"

    p.required(p.BlankLine(),"a blank line")


    p.required(p.BlanksAndBelow(1, false),"import, type, method, relation, or constant declaration at column 1 of file")

    var importSpecs []*ast.RelishImportSpec

    p.optional(p.parseImports(&importSpecs) && 
               p.required(p.BlankLine() && 
                          p.BlankLine() && 
                          p.BlanksAndBelow(1,false),
                         "type, method, relation, or constant declaration, or a line comment, at column 1 of file, after a gap of at least two blank lines"))

    

    var relDecls []*ast.RelationDecl
    var typeDecls []*ast.TypeDecl
    var methodDecls []*ast.MethodDeclaration
    var constDecls []*ast.ConstantDecl
    

    p.required (p.parseRelationDeclaration(&relDecls) || 
                p.parseTypeDeclaration(&typeDecls) || 
                p.parseMethodDeclaration(&methodDecls) || 
                p.parseConstantDeclarationBlock(&constDecls) ||
                p.LineComments(), "type, method, relation, or constant declaration, or a line comment") 	
    
    for {
       if ! p.BlankLine() {
          break
       }
       if ! p.BlankLine() {
          st2 := p.State()
          if p.Below(1) && p.LineComments() { 
             // fmt.Println("continue 2")                 
             continue 
          } else {
             // fmt.Printf(">%v<\n",p.Ch())                
             // fmt.Println("break 2")    
             // p.error(p.Pos(),"break 2")  
             p.Fail(st2)          
             break
          }
       }
       if ! p.BlanksAndBelow(1,false) {
	      break
       }
       if ! (p.parseRelationDeclaration(&relDecls) || 
             p.parseTypeDeclaration(&typeDecls) || 
             p.parseMethodDeclaration(&methodDecls) || 
             p.parseConstantDeclarationBlock(&constDecls) || 
             p.LineComments()) {
	      break
       }
    }
    p.required(p.BlankOrCommentsToEOF(),"end of file, or type, method, relation, or constant declaration, or a line comment, at column 1 after a gap of at least two blank lines")

    methodDecls = append(methodDecls, p.closureMethodDecls...)

    astFileNode := &ast.File{
	   // Doc        *CommentGroup   // associated documentation; or nil (for relish should be a single comment)	
	   Top: pos,      // position of first character of file
   	   Name: &ast.Ident{Name: packageName, Kind:token.PACKAGE},    // package name
	   ConstantDecls: constDecls,          // top-level declarations; or nil	
   	   TypeDecls: typeDecls,         // top-level declarations; or nil	
	   RelationDecls: relDecls,          // top-level declarations; or nil
 	   MethodDecls: methodDecls,    // top-level declarations; or nil
	   RelishImports: importSpecs,
	   // Comments   []*CommentGroup // list of all comments in the source file
    }

    astFileNode.StoreSourceFilePositionInfo(p.file)

    return astFileNode
}	

/*	
origin   skunkworks.everybitcounts.net2006
artifact building_energy_model/detached_house_sim
package  loads/electric_hot_water_heater	
*/

func (p *parser) parseFileHeader(fullPackageName *string) bool {
    if p.trace {
	   defer un(trace(p, "FileHeader"))
    }	
    var origin string
    var artifact string
    var packagePath string
	i := 1
	p.required(p.parseArtifactOriginDeclaration(&origin),"origin   <software_artifact_originator_domain_name><year_org_acquired_domain>") 
	p.required(p.BlankToEOL(),"nothing on line after origin declaration") 
	p.required(p.Below(i),"artifact <name_of_software_artifact>  (vertically below origin)") 
	p.required(p.parseArtifactDeclaration(&artifact),"artifact <name_of_software_artifact>  (vertically below origin)") 
	p.required(p.BlankToEOL(),"nothing on line after artifact declaration") 
    p.required(p.Below(i),"package  some/package/pathname  (vertically below artifact)") 
	p.required(p.parsePackageDeclaration(&packagePath),"package  some/package/pathname  (vertically below artifact)") 
	p.required(p.BlankToEOL(),"nothing on line after package declaration") 
	p.optional(p.parseFileComment())
	
	*fullPackageName = origin + "/" + artifact + "/pkg/" + packagePath
	return true
}
	

/*	
origin   skunkworks.everybitcounts.net2006	
*/
func (p *parser) parseArtifactOriginDeclaration(origin *string) bool {
   if p.trace {
      defer un(trace(p, "ArtifactOriginDeclaration"))
   }	
   st := p.State()
   if (p.Match("origin   ") &&
           p.required(p.ScanDomainName(), "the domain name of the originator of the software artifact") && 
           p.required(p.ScanYear(),"the year the code-originating organization first acquired the domain")   ) {
       st2 := p.State()
       *origin = p.Substring(st.Offset + 9,st2.Offset) 
       return true
   }
   return p.Fail(st)


}

/*	
artifact building_energy_model/detached_house_sim	
*/
func (p *parser) parseArtifactDeclaration(artifact *string) bool {
   if p.trace {
      defer un(trace(p, "ArtifactDeclaration")) 
   }	
   st := p.State()
   if (p.Match("artifact ") &&
            p.required(p.ScanArtifactName(), "the name of the software artifact")   ) {
	   st2 := p.State()
	   *artifact = p.Substring(st.Offset + 9,st2.Offset) 
	   return true	
   }
   return p.Fail(st)
}

/*	
package  loads/electric_hot_water_heater	
*/
func (p *parser) parsePackageDeclaration(packagePath *string) bool {
   if p.trace {
      defer un(trace(p, "PackageDeclaration"))
   }	
   st := p.State()
   if (p.Match("package  ") &&
          p.required(p.ScanPackageName(), "the name of the software package")   ) {
	   st2 := p.State()	
	   *packagePath = p.Substring(st.Offset + 9,st2.Offset) 
	   return true      	
   }

   return p.Fail(st)
}


// TODO: Want to support
// /*
// */
// style comments, but with the following differences from usual:
// 1. They can only be present at column 1 of the source file.
// 2. They nest. A comment is not finished til the nesting is unwound.
// They are only supposed to be used for temporary eliding of code for testing.


/*
   Parses a """ comment at the top of the file.
*/	
func (p *parser) parseFileComment() bool {
   if p.trace {
      defer un(trace(p, "FileComment"))
   }	
   st := p.State()

   if ! p.BlankLine() {
      return false	
   }
   if ! p.Below(1) {
      return p.Fail(st)
   }
   
   // dbg.Log(dbg.PARSE_,"Got here ch=%s\n",string(p.Ch()))	

   if ! ( p.Match(`"""`) &&
          p.required(p.BlankToEOL(),`nothing on line after """`) ) {
       return p.Fail(st)
   }
   st2 := p.State()
 
   if ! p.required(p.BlanksAndBelow(2,false),"comment content - Must begin at column 2 of file") {
       return p.Fail(st)    	
   }

   found,contentEndOffset := p.ConsumeTilMatchAtColumn(`"""`,1)
   if ! found {
	  dbg.Logln(dbg.PARSE_,`Did not consume till """.`)	
      return p.Fail(st)	
   }
   if ! p.required(p.BlankToEOL(),`nothing on line after """`) {
      return p.Fail(st)	
   }
//   if ! p.required(p.BlankLine(),"a blank line for spacing") {
//      return p.Fail(st)	
//   }


// TODO
   commentContent := p.Substring(st2.Offset,contentEndOffset) 	
// Check the content to make sure none of it is in the first column.
// Also, produce the actual content string, with first column removed.
   dbg.Logln(dbg.PARSE_,"Comment Content:")
   dbg.Logln(dbg.PARSE_,commentContent)
   dbg.Logln(dbg.PARSE_,"END Comment Content")
   return true
}



func (p *parser) parseImports(importSpecs *[]*ast.RelishImportSpec) bool {
   if p.trace {
      defer un(trace(p, "Imports"))
   }
   col := p.Col()
   if ! p.Match("import") {
	   return false
   }
   p.required(p.Indent(col),"a package path, indented below 'import'")
   p.required(p.parsePackageImport(importSpecs),"a valid package path")
   for ; p.Indent(col); {
      p.required(p.parsePackageImport(importSpecs),"a valid package path")	
   }
   return true
}

func (p *parser) parsePackageImport(importSpecs *[]*ast.RelishImportSpec) bool {
   if p.trace {
      defer un(trace(p, "PackageImport"))
   }

   var packagePath string

   col := p.Col()
   var multiline bool
   if ! p.parsePackagePath(&packagePath, &multiline) {
      return false
   }
   st2 := p.State()
   foundAlias := false

   var packageAlias *ast.Ident

   if p.Space() {
	  if p.Match("as ") && p.parsePackageAlias(false,&packageAlias) {
	     foundAlias = true	
	  } else {
         p.Fail(st2)	
      }
   }
   if ! foundAlias {
	   if multiline { 
		  col += 3
	   }
	   if p.Indent(col) {
		  if p.Match("as ") && p.parsePackageAlias(false,&packageAlias) {
		     foundAlias = true	
		  } else {
	         p.Fail(st2)	
	      }
	   }	
   }

   alias := ""
   if foundAlias {
	   if p.importPackageAliases[packageAlias.Name] {
	       p.stop(fmt.Sprintf("Package alias %s has already been defined.",packageAlias.Name))	
	   }
	   alias = packageAlias.Name
       p.importPackageAliases[packageAlias.Name] = true
   } else {
      slashPos := strings.LastIndex(packagePath, "/") 
      packageLocalName := packagePath[slashPos+1:] 	
        
      if p.importPackageAliases[packageLocalName] {
         p.stop(fmt.Sprintf("Package alias %s has already been defined.",packageLocalName))	
      }
	  alias = packageLocalName
      p.importPackageAliases[packageLocalName] = true
   }

   pkgPos := strings.Index(packagePath,"/pkg/")
   var originAndArtifactName string 
   var packageName string
   if pkgPos >= 0 {
	  originAndArtifactName = packagePath[:pkgPos]
	  packageName = packagePath[pkgPos+5:]
   } else {
      libArtifact,libArtifactFound := StandardLibPackageArtifact[packagePath]	
      if libArtifactFound {
	     originAndArtifactName = libArtifact
//    } else if StandardLibPackagePath[packagePath] {
//	      originAndArtifactName = "relish"	
	  } else {
	      originAndArtifactName = p.currentOriginAndArtifactName
      }
	  packageName = packagePath
   }

   fullPackageName := originAndArtifactName + "/pkg/" + packageName
   p.importPackageAliasExpansions[alias] = fullPackageName   

   importSpec := &ast.RelishImportSpec{alias, originAndArtifactName, packageName}   
   *importSpecs = append(*importSpecs,importSpec)

   return true
}



func (p *parser) parsePackagePath(packagePath *string, multiline *bool) bool {
   if p.trace {
      defer un(trace(p, "PackagePath"))
   }	
   return (p.parseOneLinePackagePath(packagePath) ||	
           p.parseTwoLinePackagePath(packagePath, multiline) ||
           p.parseLocalPackagePath(packagePath))
}


func (p *parser) parseOneLinePackagePath(packagePath *string) bool {	
   if p.trace {
      defer un(trace(p, "OneLinePackagePath"))
   }	
   st := p.State()
   if (p.ScanDomainName() &&
           p.ScanYear() && p.Match1('/') &&
           p.ScanArtifactName() && p.Match("/pkg/") && 
           p.ScanPackageName()     ) {
	    st2 := p.State()
	    *packagePath = p.Substring(st.Offset,st2.Offset)
	    dbg.Logln(dbg.PARSE_,"OneLinePackagePath succeeded")
	    return true
   }
   return p.Fail(st)
}

func (p *parser) parseTwoLinePackagePath(packagePath *string, multiline *bool) bool {
   if p.trace {
      defer un(trace(p, "TwoLinePackagePath"))
   }	
   st := p.State()
   col := p.Col()
   if (p.ScanDomainName() &&
       p.ScanYear() && p.Match1('/') &&
       p.ScanArtifactName() ) {
	  st2 := p.State()	
	  originAndArtifactName := p.Substring(st.Offset,st2.Offset)   
	  if p.Indent(col) && p.Match("/pkg/") {
		st3 := p.State()
        if p.ScanPackageName() {
	       st4 := p.State()
	       packageName :=  p.Substring(st3.Offset,st4.Offset) 
	       *packagePath = originAndArtifactName + "/pkg/" + packageName
	       *multiline = true
	       dbg.Logln(dbg.PARSE_,"TwoLinePackagePath succeeded")	
	       return true	
	    }
	  }
   }
   return p.Fail(st)
}

/*
Parse the package-path part of an origin,artifact, and package name specification.
e.g. geom_objects/model_2d
*/
func (p *parser) parseLocalPackagePath(packagePath *string) bool {
   if p.trace {
      defer un(trace(p, "LocalPackagePath"))
   }	
   st := p.State()
   if p.ScanPackageName() {
	    st2 := p.State()
	    *packagePath = p.Substring(st.Offset,st2.Offset)
	    dbg.Logln(dbg.PARSE_,"LocalPackagePath succeeded")	
	    return true
   } 
   return p.Fail(st)
}


	
// ----------------------------------------------------------------------------
// Parsing Grammar Support
	
/*
   e.g. required(parseVarName,"a variable name")

   Raises an error saying it is expecting the expected element type.

   elementFound should be the result of a parse<Element> function.
   whatIsExpected is the description of the thing that was expected.
*/	
func (p *parser) required(elementFound bool, whatIsExpected string) bool {
	if elementFound {
		return true
	}
	fs := p.FailedOnString()
	if len(fs) > 0 {
      //panic(fmt.Sprintf("Expecting %s.\nFound: %s", whatIsExpected, fs))    

		if p.haveProbableCause(p.FailedPos()) {
           p.error(p.FailedPos(),p.getProbableCause())
		} else {
		   // fmt.Println("Have a FailedPos and FailedOnString!!!")	
    	   p.error(p.FailedPos(),fmt.Sprintf("Expecting %s.\nFound: %s", whatIsExpected, fs))
        }
    } else {
	    var found string
        switch c := p.Ch(); c {
	       case -1:
		       found = "end of file"
		   case '\n':
		       found = "end of line"
	       default:
		       found = string(c)
        }
      //panic(fmt.Sprintf("Expecting %s.\nFound: %s", whatIsExpected, found))        
		if p.haveProbableCause(p.Pos()) {
           p.error(p.Pos(),p.getProbableCause())
		} else {        
    	   p.error(p.Pos(),fmt.Sprintf("Expecting %s.\nFound: %s", whatIsExpected, found))
	    }
    	//p.error(p.Pos(),fmt.Sprintf("%s.", whatIsExpected))
    }
	return false // Will never get here if in single error mode.
}

func (p *parser) optional(elementFound bool) bool {
	return elementFound || true
}



/*
true if belowState is on a lower file line than aboveState
*/
func (p *parser) isLower(aboveState scanner.ScanningState,belowState scanner.ScanningState) bool {
	return belowState.LineOffset > aboveState.LineOffset
}

		
func (p *parser) stop(errorMessage string) bool {
	fs := p.FailedOnString()
	if len(fs) > 0 {
    	p.error(p.FailedPos(),fmt.Sprintf("%s.\nFound: %s", errorMessage, fs))
    } else {
	    var found string
        switch c := p.Ch(); c {
	       case -1:
		       found = "end of file"
		   case '\n':
		       found = "end of line"
	       default:
		       found = string(c)
        }
    	p.error(p.Pos(),fmt.Sprintf("%s.\nFound: %s", errorMessage, found))
	    
    	//p.error(p.Pos(),fmt.Sprintf("%s.", errorMessage))	
    }
	return false // Will never get here if in single error mode.
}		
	
// ----------------------------------------------------------------------------
// Scoping support  




/*
Initialize a new variable scope.
*/
func (p *parser) clearVariableScope() {
    p.currentScopeVariables = make(map[string] bool)
    p.currentScopeVariableOffsets = make(map[string] int)	
    p.currentScopeVariableOffset = 3 // room for base pointer + pushed method ref + code offset in current method
    p.currentScopeReturnArgOffset = -1 // below the base pointer
}

/*
...About to parse a nested closure-method declaration, save variable scope of the enclosing method
   declaration.
*/
func (p *parser) pushVariableScope() {
    p.outerScopeVariables = p.currentScopeVariables 
    p.outerScopeVariableOffsets = p.currentScopeVariableOffsets	
    p.outerScopeVariableOffset = p.currentScopeVariableOffset
    p.outerScopeReturnArgOffset = p.currentScopeReturnArgOffset
    p.parsingClosure = true
}

/*
...After parsing a nested closure-method declaration, restore variable scope of the enclosing method
   declaration.
*/
func (p *parser) popVariableScope() {
    p.currentScopeVariables = p.outerScopeVariables 
    p.currentScopeVariableOffsets = p.outerScopeVariableOffsets	
    p.currentScopeVariableOffset = p.outerScopeVariableOffset 
    p.currentScopeReturnArgOffset = p.outerScopeReturnArgOffset 
    p.parsingClosure = false
    p.closureFreeVars = nil
}

/*
Magic for closure free-variable binding specifications...
Repairs the offsets of the idents that represent free vars inside the closure-method.
Sets the p.closureFreeVarBindings list 
- the enclosing-method var offsets of the free vars in the closure-method declaration.
- These should be used as the Closure.Bindings

TODO NEED TO CALL THIS AT END OF parseMethodDeclaration - before the popVariableScope() call happens!
*/
func (p *parser) fixUpFreeVarOffsets(startingFreeVarOffset int) {
	freeVarOuterMethodOffsetToNewOffset := make(map[int]int)
	p.closureFreeVarBindings = nil 
	currentFreeVarOffset := startingFreeVarOffset
	for _,freeVar := range p.closureFreeVars {
		
		// fmt.Println("currentFreeVarOffset=",currentFreeVarOffset)
		
		newOffset, offsetEncountered := freeVarOuterMethodOffsetToNewOffset[freeVar.Offset] 
		if offsetEncountered {
			freeVar.Offset = newOffset
		} else {
			p.closureFreeVarBindings = append(p.closureFreeVarBindings, freeVar.Offset)
			// fmt.Println("p.closureFreeVarBindings=",p.closureFreeVarBindings)
			freeVarOuterMethodOffsetToNewOffset[freeVar.Offset] = currentFreeVarOffset
			freeVar.Offset = currentFreeVarOffset
			currentFreeVarOffset++
		}
	}
}


func (p *parser) ensureCurrentScopeVariable(ident *ast.Ident, knownToBeReturnArg bool) {
	newVar := false
	if ! p.currentScopeVariables[ident.Name] {
   	   p.currentScopeVariables[ident.Name] = true  // should change this to refer to the input arg or whatever ast node.
       newVar = true

       if knownToBeReturnArg {
	   	   p.currentScopeVariableOffsets[ident.Name] = p.currentScopeReturnArgOffset  
	       ident.Offset = p.currentScopeReturnArgOffset	
	 	   p.currentScopeReturnArgOffset--	
	
       } else {
	   	   p.currentScopeVariableOffsets[ident.Name] = p.currentScopeVariableOffset  
	       ident.Offset = p.currentScopeVariableOffset	
	 	   p.currentScopeVariableOffset++
       }
    } 
	dbg.Log(dbg.PARSE_,"ensureCSVar new %v %v %v knownRetArg %v\n",newVar,ident.Name,ident.Offset,knownToBeReturnArg)
}



// The following methods are Go parsing relics that are not used by relish much if at all yet
// because relish scoping rules are deliberately dead simple. Are they needed?

func (p *parser) openScope() {
	p.topScope = ast.NewScope(p.topScope)
}

func (p *parser) closeScope() {
	p.topScope = p.topScope.Outer
}

func (p *parser) openLabelScope() {
	p.labelScope = ast.NewScope(p.labelScope)
	p.targetStack = append(p.targetStack, nil)
}

func (p *parser) closeLabelScope() {
	// resolve labels
	n := len(p.targetStack) - 1
	scope := p.labelScope
	for _, ident := range p.targetStack[n] {
		ident.Obj = scope.Lookup(ident.Name)
		if ident.Obj == nil && p.mode&DeclarationErrors != 0 {
			p.error(ident.Pos(), fmt.Sprintf("label %s undefined", ident.Name))
		}
	}
	// pop label scope
	p.targetStack = p.targetStack[0:n]
	p.labelScope = p.labelScope.Outer
}


// ----------------------------------------------------------------------------
// Parsing support

func (p *parser) printTrace(a ...interface{}) {
	const dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . " +
		". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . "
	const n = uint(len(dots))
	pos := p.file.Position(p.Pos())
	fmt.Printf("%5d:%3d: ", pos.Line, pos.Column)
	i := 2 * p.indent
	for ; i > n; i -= n {
		fmt.Print(dots)
	}
	fmt.Print(dots[0:i])
	fmt.Println(a...)
}

func trace(p *parser, msg string) *parser {
	p.printTrace(msg, "(")
	p.indent++
	return p
}

// Usage pattern: defer un(trace(p, "..."));
func un(p *parser) {
	p.indent--
	p.printTrace(")")
}


func (p *parser) error(pos token.Pos, msg string) {
	p.Error(p.file.Position(pos), msg)
}


/*
If an indentation error has been detected, set the probableErrorCause to an indentation error message.
*/
func (p *parser) checkForIndentError() {
	if p.IndentWobble != 0 {
		if p.IndentWobble == 1 {
           p.probableErrorCause = "Must indent with 3 spaces. Indented with 4 spaces."
		} else { // -1
           p.probableErrorCause = "Must indent with 3 spaces. Only indented 2 spaces."
		}
		p.probableCausePos = p.IndentWobblePos
	}
}

func (p *parser) setProbableCause(reason string) {
	p.probableErrorCause = reason
	p.probableCausePos = p.Pos()
}

func (p *parser) clearProbableCause() {
	p.probableErrorCause = ""
}

/*
There has been a probable cause (of a syntax error) recorded, and it is for a problem occurring near where the actual
parsing failure was finally detected.
*/
func (p *parser) haveProbableCause(failurePos token.Pos) bool {
	// fmt.Println("failurePos",failurePos,"p.probableCausePos",p.probableCausePos)
	return p.probableErrorCause != "" && failurePos <= p.probableCausePos + 3
}

func (p *parser) getProbableCause() string {
	return p.probableErrorCause
}





