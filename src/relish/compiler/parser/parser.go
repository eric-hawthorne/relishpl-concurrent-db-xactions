// Substantial portions of the source code in this file 
// are Copyright 2009 The Go Authors. All rights reserved.
// Use of such source code is governed by a BSD-style
// license that can be found in the GO_LICENSE file.

// Modifications and additions which convert code to be part of a relish-language compiler 
// are Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
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
// 3. Parser should continue at next top level construct after error.



//
package parser

import (
	"fmt"
	"strings"
	"relish/compiler/ast"
	"relish/compiler/scanner"
	"relish/compiler/token"
	. "relish/defs"
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

	// Non-syntactic parser control
	exprLev int // < 0: in control clause, >= 0: in expression

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
	
	

	// Label scope
	// (maintained by open/close LabelScope)
	labelScope  *ast.Scope     // label scope for current function
	targetStack [][]*ast.Ident // stack of unresolved labels
}

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
       "continue": true,
       "break": true,
       "go": true,
   }	
}

func (p *parser) clearVariableScope() {
    p.currentScopeVariables = make(map[string] bool)
    p.currentScopeVariableOffsets = make(map[string] int)	
    p.currentScopeVariableOffset = 3 // room for base pointer + pushed method ref + code offset in current method
    p.currentScopeReturnArgOffset = -1 // below the base pointer
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
	fmt.Printf("ensureCSVar new %v %v %v knownRetArg %v\n",newVar,ident.Name,ident.Offset,knownToBeReturnArg)
}

// ----------------------------------------------------------------------------
// Relish


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
/*
func (p *parser) parseFile() bool {
   parseFileHeader();
   while(parseRelationDeclaration() 
         ||
         parseTypeDeclaration()
         ||
         parseConstantDeclaration()

   // if not at EOF complain, noting whether something may be inappropriately indented.
	
		required(parseRelationDeclaration(), "a class name e.g. SomeType or pkg1/SomeType");	
		while(parseRelationDeclaration());	
		while(whiteSpaceThenEOL());
		required((c=ch()) == EOF,"End of file.");
	return true;
}


_, err := ParseFile(fset, filename, nil, DeclarationErrors)

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


    p.required(p.BlanksAndBelow(1),"import, type, method, relation, or constant declaration at column 1 of file")

    var importSpecs []*ast.RelishImportSpec

    p.optional(p.parseImports(&importSpecs) && 
               p.required(p.BlankLine() && 
                          p.BlanksAndBelow(1),
                         "type, method, relation, or constant declaration, or a line comment, at column 1 of file, after a blank line"))

    

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
       if ! p.BlanksAndBelow(1) {
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
    p.required(p.BlankOrCommentsToEOF(),"end of file, or type, method, relation, or constant declaration, or a line comment, at column 1 after a blank line")

    return &ast.File{
	   // Doc        *CommentGroup   // associated documentation; or nil (for relish should be a single comment)	
	   Top: pos,      // position of first character of file
   	   Name: &ast.Ident{Name: packageName, Kind:token.PACKAGE},    // package name
	   ConstantDecls: constDecls,          // top-level declarations; or nil	
   	   TypeDecls: typeDecls,         // top-level declarations; or nil	
	   RelationDecls: relDecls,          // top-level declarations; or nil
 	   MethodDecls: methodDecls,    // top-level declarations; or nil
	   // Scope      *Scope          // package scope (this file only)
	   // Imports    []*ImportSpec   // imports in this file
	   RelishImports: importSpecs,
	   // Unresolved []*Ident        // unresolved identifiers in this file
	   // Comments   []*CommentGroup // list of all comments in the source file
    }

    // return &ast.File{doc, pos, ident, decls, p.pkgScope, p.imports, p.unresolved[0:i], p.comments}
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


/*ELIDE*
*ELIDE*/

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
   
   // fmt.Printf("Got here ch=%s\n",string(p.Ch()))	

   if ! ( p.Match(`"""`) &&
          p.required(p.BlankToEOL(),`nothing on line after """`) ) {
       return p.Fail(st)
   }
   st2 := p.State()
 
   if ! p.required(p.BlanksAndBelow(2),"comment content - Must begin at column 2 of file") {
       return p.Fail(st)    	
   }

   found,contentEndOffset := p.ConsumeTilMatchAtColumn(`"""`,1)
   if ! found {
	  fmt.Println(`Did not consume till """.`)	
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
   fmt.Println("Comment Content:")
   fmt.Println(commentContent)
   fmt.Println("END Comment Content")
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
	  originAndArtifactName = p.currentOriginAndArtifactName
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
	    fmt.Println("OneLinePackagePath succeeded")
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
	       fmt.Println("TwoLinePackagePath succeeded")	
	       return true	
	    }
	  }
   }
   return p.Fail(st)
}


func (p *parser) parseLocalPackagePath(packagePath *string) bool {
   if p.trace {
      defer un(trace(p, "LocalPackagePath"))
   }	
   st := p.State()
   if p.ScanPackageName() {
	    st2 := p.State()
	    *packagePath = p.Substring(st.Offset,st2.Offset)
	    fmt.Println("LocalPackagePath succeeded")	
	    return true
   } 
   return p.Fail(st)
}

/*
   Parses a """ comment at the top of a method declaration.

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
   
   fmt.Printf("Got here ch=%s\n",string(p.Ch()))	

   if ! ( p.Match(`"""`) &&
          p.required(p.BlankToEOL(),`nothing on line after """`) ) {
       return p.Fail(st)
   }
   st2 := p.State()
 
   if ! p.required(p.BlanksAndBelow(2),"comment content - Must begin at column 2 of file") {
       return p.Fail(st)    	
   }

   found,contentEndOffset := p.ConsumeTilMatchAtColumn(`"""`,1)
   if ! found {
	  fmt.Println(`Did not consume till """.`)	
      return p.Fail(st)	
   }
   if ! p.required(p.BlankToEOL(),`nothing on line after """`) {
      return p.Fail(st)	
   }
// TODO
   commentContent := p.Substring(st2.Offset,contentEndOffset) 	
// Check the content to make sure none of it is in the first column.
// Also, produce the actual content string, with first column removed.
   fmt.Println("Comment Content:")
   fmt.Println(commentContent)
   fmt.Println("END Comment Content")
   return true
}
	
/*
   Parses a """ comment at the top of a type declaration

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
   
   fmt.Printf("Got here ch=%s\n",string(p.Ch()))	

   if ! ( p.Match(`"""`) &&
          p.required(p.BlankToEOL(),`nothing on line after """`) ) {
       return p.Fail(st)
   }
   st2 := p.State()
 
   if ! p.required(p.BlanksAndBelow(2),"comment content - Must begin at column 2 of file") {
       return p.Fail(st)    	
   }

   found,contentEndOffset := p.ConsumeTilMatchAtColumn(`"""`,1)
   if ! found {
	  fmt.Println(`Did not consume till """.`)	
      return p.Fail(st)	
   }
   if ! p.required(p.BlankToEOL(),`nothing on line after """`) {
      return p.Fail(st)	
   }
// TODO
   commentContent := p.Substring(st2.Offset,contentEndOffset) 	
// Check the content to make sure none of it is in the first column.
// Also, produce the actual content string, with first column removed.
   fmt.Println("Comment Content:")
   fmt.Println(commentContent)
   fmt.Println("END Comment Content")
   return true
}
		
	
/*
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
	
	var rel1TypeName, rel2TypeName, rel1EndName, rel2EndName *ast.Ident
	var arity1Spec, arity2Spec *ast.AritySpec
	var collection1Spec,collection2Spec *ast.CollectionTypeSpec
	
	return (p.parseTypeName(true, &rel1TypeName) && p.Space() &&
	p.optional(p.parseCollectionTypeSpec(&collection1Spec) && p.Space()) &&
	p.parseAritySpec(&arity1Spec) &&
	p.optional(p.parseRelEndName(&rel1EndName) && p.Space()) &&
	p.Match("-- ") &&
	p.optional(p.parseRelEndName(&rel2EndName) && p.Space()) && 	
	p.parseAritySpec(&arity2Spec) &&	
	p.optional(p.parseCollectionTypeSpec(&collection2Spec) && p.Space()) &&
	p.parseTypeName(true, &rel2TypeName) ) ||
	p.Fail(st)
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
1 means 1 1 
N  means 0 N 
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
		minCard = 0
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
	   fmt.Printf("succeeded on not a digit. '%s'\n",string(p.Ch()))	
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
	   fmt.Printf("not a space after second digits. '%s'\n",string(p.Ch()))
	   return p.Fail(st)
    }	   	

  Translate:
    *aritySpec = &ast.AritySpec{minCard,maxCard,pos,end}
    
    fmt.Println("successful arity spec parse.")
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
Parse the type name, type paramters, and supertypes part of a type declaration, but not 
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

    p.required( p.parseTypeSpec(false,false,false,false,&superTypeSpec), "a type name." )

    // translate
    typeSpec.SuperTypes = append(typeSpec.SuperTypes, superTypeSpec)

    st2 := p.State()

    for p.Space() {
	   if p.parseTypeSpec(false,false,false,false,&superTypeSpec) {
		
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
   return p.parseTypeSpec(true, true, true, false, typeSpec)	
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

    p.required(p.parseTypeSpec(true, true, false, false, &superTypeSpec), "a type specification" )	

    // translate
    typeSpec.SuperTypes = append(typeSpec.SuperTypes, superTypeSpec)
	
    for p.Indent(col) {
       p.required(p.parseTypeSpec(true, true, false, false, &superTypeSpec),"a type specification")	

       // translate
       typeSpec.SuperTypes = append(typeSpec.SuperTypes, superTypeSpec)	
    }
	
	// TODO Create the ast nodes.
	
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

    if p.BlanksAndIndent(col) {
       if ! p.parseReadWriteAttributeDecl(&attrs,"public") {
  	      return p.Fail(st)
       }
    } else if p.BlanksAndMiniIndent(col) {
       if ! (p.parseReadOnlyAttributeDecl(&attrs,"public") || p.parseWriteOnlyAttributeDecl(&attrs,"public")) {	
	      return p.Fail(st)
       }	
    } else {
	    return false
    }

    for {
	    if p.BlanksAndIndent(col) {
	       p.required(p.parseReadWriteAttributeDecl(&attrs,"public"),"an attribute declaration")
	    } else if p.BlanksAndMiniIndent(col) {
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


func (p *parser) parseOneLineAttributeDecl(attrs *[]*ast.AttributeDecl, read, write bool, visibilityLevel string) bool {
	
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
    if ! p.parseTypeSpec(true,true,false,forceCollection,&typeSpec) {
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
*/
func (p *parser) parseAttributeDecl(attrs *[]*ast.AttributeDecl, read, write bool, visibilityLevel string) bool {
	
   return p.parseOneLineAttributeDecl(attrs,read,write,visibilityLevel)
}








// TODO NEXT !! Need to have MethodDeclaration and ConstantDeclarationBlock return ast nodes.

/*
Note: A method with no input arguments must have an implementation.
*/
func (p *parser) parseMethodDeclaration(methodDecls *[]*ast.MethodDeclaration) bool {
    if p.trace {
       defer un(trace(p, "MethodDeclaration"))
    }	
    var methodDecl *ast.MethodDeclaration 
    col := p.Col()
    if p.parseMethodHeader(&methodDecl) &&
       p.optional(p.parseMethodBody(col,methodDecl)) {
	
	    // TODO check if we have a method body if no input args.
	
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

    var methodName *ast.Ident

    if ! p.parseMethodName(false,&methodName) {
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
	   kind = token.FUNC
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

   if p.currentScopeVariables[name] { // set the Offset to the right local var or return arg
      offset = p.currentScopeVariableOffsets[name]	
   } else if mustBeDefined {
   	    fmt.Println("------name not found as local var or return arg --------")
	    fmt.Println(p.currentScopeVariables)
	    fmt.Println(name)
	    fmt.Println("--------------------------------------------------------")
       return p.Fail(st)	
   }

   *varName = &ast.Ident{pos, name, nil, token.VAR, offset}

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
			p.Fail(st)
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
    if ! p.parseTypeSpec(true,true,false,false,&typeSpec) {
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
    if ! p.parseTypeSpec(true,true,false,false,&typeSpec) {
	    return p.Fail(st)
    }

    if isVariadic {
       if typeSpec.CollectionSpec != nil {
           if typeSpec.CollectionSpec.Kind == token.LIST {
	          *isVariadicListParam = true	 
	          fmt.Println("SETTING *isVariadicListParam to true")         
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


/*
   For now, just parse a simple type name. or a collection type spec TODO handle parameterized types.
   Also, need to be able to handle T <: T1
   Re: parameterized type specs. Need to know when you can have a SomeType of T1 T2 vs a SomeType of Int Float

   canBeParameterized means allow this to be a parameterized type
   canBeVariable means this is allowed to be a type variable
   canSpecifySuperTypes means this expression can include <: supTyp1 supTyp2 etc
   forceCollection means that this type is actually a collection even if no collection type expression occurs. 
   If forceCollection is true and there is no collection type expression, this type is a Set of the base type.
*/
func (p *parser) parseTypeSpec(canBeParameterized bool, 
	                           canBeVariable bool, 
	                           canSpecifySuperTypes bool, 
	                           forceCollection bool,
	                           typeSpec **ast.TypeSpec) bool {
   

    var collectionTypeSpec *ast.CollectionTypeSpec

    // TODO   NOT HANDLING MAPS YET!!!!!!
    if p.parseCollectionTypeSpec(&collectionTypeSpec) {
	   p.required(p.Space(),"a space then a type name")
	} else if forceCollection {
	   collectionTypeSpec = &ast.CollectionTypeSpec{token.SET,p.Pos(),p.Pos()+1,false,false,""}
    }

	var typeName *ast.Ident
	
	if ! p.parseTypeName(true, &typeName) { 
	   return false
	}
	
	// Have to look for TypeName=>TypeName or 
	// TypeName
	// =>
	// TypeName
	//
	// IDEA: Maybe the key type goes as a sub part of the collection type!!!!
	// Do this when changing the collection type from token.SET to token.MAP
	
    *typeSpec = &ast.TypeSpec{CollectionSpec: collectionTypeSpec, Name: typeName}
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
   }
   *typeName = &ast.Ident{pos, name, nil, token.TYPE,-1}
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
	        funcType.Results = returnArgDecls
	        return true
	   }
	   st2 = p.State()
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
            funcType.Results = returnArgDecls              
            return true
       }     
       st2 = p.State()
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
    if ! p.parseTypeSpec(true, true, false, false, &typeSpec) {
	    return p.Fail(st)
    }

	returnArg := &ast.ReturnArgDecl{argName,typeSpec}
	
	*returnArgs = append(*returnArgs,returnArg)
	
    if argName != nil {
       p.ensureCurrentScopeVariable(argName, true)	 	
	}
	return true
}

/*
   Temporary implementation - need to handle indented type spec
*/
func (p *parser) parseReturnArgDecl(returnArgs *[]*ast.ReturnArgDecl) bool {
	
   return p.parseOneLineReturnArgDecl(returnArgs)
}





//foo a Int b Int > Int er.Error
func (p *parser) parseMethodBody(col int,methodDecl *ast.MethodDeclaration) bool {
    if p.trace {
       defer un(trace(p, "MethodBody"))
    }	

    // parse
    st := p.State()
    if ! p.Indent(col) {
	   return false
    }

    blockPos := p.Pos()

    var stmts []ast.Stmt

    col2 := p.Col()
    if ! p.parseMethodBodyStatement(&stmts) {
	   return p.Fail(st)
    }
    for p.BlanksAndBelow(col2) {
        p.required(p.parseMethodBodyStatement(&stmts),"a statement")
    }

    methodDecl.Body = &ast.BlockStatement{blockPos,stmts}

    // TODO Need to check for return argument assignments, compatibility etc

    return true   
}


func (p *parser) parseMethodBodyStatement(stmts *[]ast.Stmt) bool {
    if p.trace {
       defer un(trace(p, "MethodBodyStatement"))
    }

    var s ast.Stmt
    // parse
    if p.parseControlStatement(&s) {
	   // translate
	   *stmts = append(*stmts,s)
	   return true
    }
    
    var rs *ast.ReturnStatement
    // parse
    if p.parseReturnStatement(&rs) {
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
func (p *parser) parseIfClauseStatement(stmts *[]ast.Stmt) bool {
    if p.trace {
       defer un(trace(p, "IfClauseStatement"))
    }

    var s ast.Stmt
    // parse
    if p.parseControlStatement(&s) {
	   // translate
	   *stmts = append(*stmts,s)
	   return true
    }
    
    var rs *ast.ReturnStatement
    // parse
    if p.parseReturnStatement(&rs) {
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
How is this different from method body?
*/
func (p *parser) parseLoopBodyStatement(stmts *[]ast.Stmt) bool {
    if p.trace {
       defer un(trace(p, "LoopBodyStatement"))
    }

    var s ast.Stmt
    // parse
    if p.parseControlStatement(&s) {
	   // translate
	   *stmts = append(*stmts,s)
	   return true
    }
    
    var rs *ast.ReturnStatement
    // parse
    if p.parseReturnStatement(&rs) {
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


func (p *parser) parseControlStatement(stmt *ast.Stmt) bool {
    if p.trace {
       defer un(trace(p, "ControlStatement"))
    }
    var ifStmt *ast.IfStatement
    if p.parseIfStatement(&ifStmt) {
	   *stmt = ifStmt
	   return true
    } 
    var whileStmt *ast.WhileStatement
    if p.parseWhileStatement(&whileStmt) {
	   *stmt = whileStmt
	   return true
    }  
    var rangeStmt *ast.RangeStatement
    if p.parseForRangeStatement(&rangeStmt) {
	   *stmt = rangeStmt
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
    if p.parseForStatement(&forStmt) {
	   *stmt = forStmt
	   return true
    }
*/
    return false
}

/*
   => foo 9
      b

ReturnStatement struct {
	Return  token.Pos // position of "=>" keyword
	Results []Expr    // result expressions; or nil
}
*/
func (p *parser) parseReturnStatement(stmt **ast.ReturnStatement) bool {
    if p.trace {
       defer un(trace(p, "ReturnStatement"))
    }

    pos := p.Pos()

    if ! p.Match2('=','>') {
	   return false
    }
    p.required(p.Space(),"a single space after '=>'")
  
    isInsideBrackets := false

    var xs []ast.Expr

    p.required(p.parseLenientVerticalExpressions(p.Col(),&xs) || 
               p.parseMultipleOneLineExpressions(&xs,isInsideBrackets) || 
               p.parseSingleOneLineExpression(&xs,isInsideBrackets), 
               "a literal value or constant or variable reference or method call")

    // translate
    *stmt = &ast.ReturnStatement{pos,xs}

    return true
}




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

    if p.parseOneLineLiteral(x) {
	   return true
    }

    var constant *ast.Ident
    if p.parseConstName(true,&constant) {
	   // translate
	   *x = constant	
	   return true
    }
    
   
    if p.parseOneLineVariableReference(false, false, true, x) {
	   return true
    }

    // if isInsideBrackets can't have a method call of any kind nor a constructor invocation of any kind. 

    if ! isInsideBrackets {
	
	   if isOneOfMultiple { // need a bracketed method call
	       var mcs *ast.MethodCall
         var lcs *ast.ListConstruction         
	       
	       if p.Match1('(') {
	
	          p.required(p.parseOneLineMethodCall(&mcs,true) || p.parseOneLineListConstruction(&lcs,true),"a subroutine call or constructor invocation") 
		
		        p.required(p.Match1(')'),")")
		
		        // translate
            if mcs == nil {
              *x = lcs
            } else {
		          *x = mcs
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
       }
    }

    return false
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

    if p.parseIndentedLiteral(x) {
       return true	
    }

    if p.parseIndentedVariableReference(false, true, x) {  
	   return true
    }

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

    return false
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
May be only one expression, but must be indented below the current line.
*/
func (p *parser) parseIndentedExpressions(col int, xs *[]ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "IndentedExpressions"))
    }
    if ! p.Indent(col) {
	   return false
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
    return foundIndent	
}



/*
OBSOLETE - SEE ABOVE VERSION
May be only one expression, but must be indented below the current line.
*/
func (p *parser) parseIndentedExpressionsOrKeywordParamAssignments2(col int, xs *[]ast.Expr, firstAssignmentPos *int32, kws map[string]ast.Expr) bool {
	
	
	
	
    if p.trace {
       defer un(trace(p, "IndentedExpressionsOrKeywordParamAssignments"))
    }
    if ! p.Indent(col) {
	   return false
    }

    *firstAssignmentPos = -1

    var i int32 = 0
    var x ast.Expr

    foundPositional := false
    foundKeywords := false

    // translate
    *xs = nil

    if p.parseExpression(&x) {
	   i++
	   foundPositional = true
	
       // translate
       *xs = append(*xs,x)	
    }
  
// old  
//    p.required(p.parseExpression(&x),"an expression")

    st := p.State()

    for p.Indent(col) {
	
	   if ! p.parseExpression(&x) {
          p.Fail(st)		
	      break	
	   }
	   i++
       st = p.State()	
	
       // translate
       *xs = append(*xs,x)	
    }

	p.required(false,"after expressions")

    var key string


    for p.Indent(col) {
	
       if ! p.parseKeywordParameterAssignmentStatement(&key, &x) {
           p.Fail(st)	
	       break
	   }
	   foundKeywords = true
       *firstAssignmentPos = i	
       st = p.State()	
	
       // translate	
	   kws[key] = x	
    }


    for p.Indent(col) {
	
	   p.required(p.parseExpression(&x),"an expression")
	
       // translate
       *xs = append(*xs,x)	
    }

    if ! foundPositional && ! foundKeywords {
        p.stop("Expecting an expression or parameter assignment statement") 	          
    }

    return true
}




// func (p *parser) parseKeywordParameterAssignmentStatement(paramName *string, rhs *ast.Expr) bool {




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
true if belowState is on a lower file line than aboveState
*/
func (p *parser) isLower(aboveState scanner.ScanningState,belowState scanner.ScanningState) bool {
	return belowState.LineOffset > aboveState.LineOffset
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
	    if ! p.parseVerticalExpressionsOrKeywordParamAssignments(p.Col(), &xs, &firstAssignmentPos, kws) {
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

    pos := p.Pos()
    var end token.Pos

    var typeSpec *ast.TypeSpec

    var collectionTypeSpec *ast.CollectionTypeSpec

    var elementExprs []ast.Expr    

    emptyList := false
    hasType := false
    
    if p.Match2('[',']') {
        end = p.Pos()
        emptyList = true
    } else if p.Match1('[') {

        // Get the list of expressions inside the list literal square-brackets

        p.required(p.parseMultipleOneLineExpressions(&elementExprs, false) || p.parseSingleOneLineExpression(&elementExprs, false),
                   "zero or more list-element expressions on the same line as [, followed by closing square-bracket ]")

        p.required(p.Match1(']'), "closing square-bracket ]")   
        end = p.Pos()            
    } else { 
	     return false // Did not match a list-specifying square-bracket
    }

    if p.parseTypeSpec(true,false,false,false,&typeSpec) {
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

    isSorting := false // TODO allow sorting-list specifications in list constructions!!!!
    isAscending := false 
    orderFunc := "" 
    collectionTypeSpec = &ast.CollectionTypeSpec{token.LIST,pos,end,isSorting,isAscending,orderFunc}

    typeSpec.CollectionSpec = collectionTypeSpec

    

    var queryStringExprs []ast.Expr  // there is only allowed to be one query string
    var queryStringExpr ast.Expr

    st2 := p.State()

    if p.Space() { // May be arguments
       if p.parseSingleOneLineExpression(&queryStringExprs, isInsideBrackets) {
          queryStringExpr = queryStringExprs[0]
       } else {
	        p.Fail(st2)
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
    pos := p.Pos()
    var end token.Pos

    var typeSpec *ast.TypeSpec

    var collectionTypeSpec *ast.CollectionTypeSpec

    var elementExprs []ast.Expr    

    emptyList := false
    hasType := false
    


    if p.Match2('[',']') {
        end = p.Pos()
        emptyList = true
    } else if p.Match1('[') {

        // Get the list of expressions inside the list literal square-brackets

        if p.parseIndentedExpressions(st.RuneColumn, &elementExprs) {
            p.required(p.Below(st.RuneColumn) && p.Match1(']'), "closing square-bracket ] aligned exactly below opening [")   
        } else {
           p.required(p.parseMultipleOneLineExpressions(&elementExprs, false) || p.parseSingleOneLineExpression(&elementExprs, false),
                   "zero or more list-element expressions, followed by closing square-bracket ]")  
           p.required(p.Match1(']'), "closing square-bracket ]")              
        }
        end = p.Pos()            
    } else { 
       return false // Did not match a list-specifying square-bracket
    }

    if p.parseTypeSpec(true,false,false,false,&typeSpec) {
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

    isSorting := false // TODO allow sorting-list specifications in list constructions!!!!
    isAscending := false 
    orderFunc := "" 
    collectionTypeSpec = &ast.CollectionTypeSpec{token.LIST,pos,end,isSorting,isAscending,orderFunc}

    typeSpec.CollectionSpec = collectionTypeSpec




    var queryStringExprs []ast.Expr  // there is only allowed to be one query string
    var queryStringExpr ast.Expr

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

    // translate
    *stmt = &ast.ListConstruction{Type:typeSpec, Elements: elementExprs, Query:queryStringExpr}  
	
    return true
}


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
         typName = strings.Title(tok.String())
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
    } else {
	    return p.Fail(st)
    }

    for _,expr := range lhsList {
	   switch expr.(type) {
	      case *ast.Ident:
		     variable := expr.(*ast.Ident)
             p.ensureCurrentScopeVariable(variable, false)
	      default: 
		    fmt.Println("------------IS NOT A VARIABLE ---------------")
		    fmt.Println(expr)
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












// TODO TODO TODO 
func (p *parser) parseLHSList(mustBeLocalVar bool, lhsList *[]ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "LHSList"))
    }

    var x ast.Expr

    if ! p.parseOneLineVariableReference(mustBeLocalVar, true, false, &x) {
	    return false
    }

    // translate
	*lhsList = append(*lhsList,x)
	
    st2 := p.State()
    for p.Space() && p.parseOneLineVariableReference(mustBeLocalVar, true, false, &x) {
	
	   // translate
	   *lhsList = append(*lhsList,x)
	
	   st2 = p.State()
    } 
    p.Fail(st2)
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
func (p *parser) parseOneLineVariableReference(mustBeLocalVar bool, mustBeAssignable bool, mustBeDefined bool, expr *ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "OneLineVariableReference"))
    }
    var varName *ast.Ident
    if ! p.parseVarName(&varName, mustBeDefined) {
	    return false
    }

    var x ast.Expr = varName

    for p.parseOneLineIndexExpression(x,&x) {	
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

    for p.parseOneLineIndexExpression(x,&x) {	
    }
    for p.Match1('.') {
	    p.required(p.parseOneLineVariableReference1(x,&x),"a variable or accessor-method name")
    }

	*expr = x
	
    return true
}


func (p *parser) parseIndentedVariableReference(mustBeAssignable bool,mustBeDefined bool, expr *ast.Expr) bool {
    if p.trace {
       defer un(trace(p, "IndentedVariableReference"))
    }
    var varName *ast.Ident
    if ! p.parseVarName(&varName, mustBeDefined) {
	    return false
    }

    var x ast.Expr = varName

    for p.parseIndexExpression(x,&x) {	
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

    for p.parseIndexExpression(x,&x) {	
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
    if ! p.Indent(col) {
	   return p.Fail(st)
	}
	if ! p.parseExpression(x) {
		return p.Fail(st)
	}
    p.required(p.Below(col),"] below [")

    rBracketPos := p.Pos()
    p.required(p.Match1(']'),"] below [") 

    *x = &ast.IndexExpr{exprSoFar,lBracketPos, *x, rBracketPos}

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
    p.required(p.parseOneLineExpression(x,false,false),"an expression")
    rBracketPos := p.Pos()
    p.required(p.Match1(']'),"']'")

    *x = &ast.IndexExpr{exprSoFar,lBracketPos, *x, rBracketPos}

    return true
}


/*
if expr
   block
[elif expr
   block}...
[else
   block]
*/
func (p *parser) parseIfStatement(ifStmt **ast.IfStatement) bool {
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
    p.required(p.Indent(col),"a statement, indented from the 'if'")

    blockPos := p.Pos()

    p.required(p.parseIfClauseStatement(&stmtList),"a statement")
    for ; p.Indent(col) ; {
       p.required(p.parseIfClauseStatement(&stmtList),"a statement")	
    }

    // translate
    body := &ast.BlockStatement{blockPos,stmtList}
    if0 := &ast.IfStatement{pos, x, body, nil}
    lastIf := if0
    *ifStmt = if0

    // parse
    st2 := p.State()
    for ; p.Below(col) ; {

       pos = p.Pos()

       if p.Match("elif ") {
	
	      p.required(p.parseExpression(&x),"an expression") 
	      p.required(p.Indent(col),"a statement, indented from the 'elif'")
	
	      // translate
 	      var stmtList2 []ast.Stmt
          blockPos = p.Pos()	
	     
	      // parse
	      p.required(p.parseIfClauseStatement(&stmtList2),"a statement")
	      for ; p.Indent(col) ; {
	         p.required(p.parseIfClauseStatement(&stmtList2),"a statement")
	      }	
	      st2 = p.State()
	
	      // translate
	      body = &ast.BlockStatement{blockPos,stmtList2}
	      if1 := &ast.IfStatement{pos, x, body, nil}
	      lastIf.Else = if1
	      lastIf = if1	
	
	   // parse
       } else if p.Match("else") {
	      p.required(p.Indent(col),"a statement, indented from the 'else'")

          // translate
          var stmtList3 []ast.Stmt
          blockPos = p.Pos()
	
	      // parse
	      p.required(p.parseIfClauseStatement(&stmtList3),"a statement")	
	      for ; p.Indent(col) ; {
		 	     p.required(p.parseIfClauseStatement(&stmtList3),"a statement")
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
func (p *parser) parseWhileStatement(whileStmt **ast.WhileStatement) bool {
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
    p.required(p.Indent(col),"a statement, indented from the 'while'")

    blockPos := p.Pos()

    p.required(p.parseLoopBodyStatement(&stmtList),"a statement")
    for ; p.Indent(col) ; {
       p.required(p.parseLoopBodyStatement(&stmtList),"a statement")	
    }

    // translate
    body := &ast.BlockStatement{blockPos,stmtList}
    while0 := &ast.WhileStatement{pos, x, body, nil}
    var lastIf *ast.IfStatement
    lastWhile := while0
    *whileStmt = while0

    // parse
    st2 := p.State()
    for ; p.Below(col) ; {

       pos = p.Pos()

       if p.Match("elif ") {
	
	      p.required(p.parseExpression(&x),"an expression") 
	      p.required(p.Indent(col),"a statement, indented from the 'elif'")
	
	      // translate
 	      var stmtList2 []ast.Stmt
          blockPos = p.Pos()	
	     
	      // parse
	      p.required(p.parseIfClauseStatement(&stmtList2),"a statement")
	      for ; p.Indent(col) ; {
	         p.required(p.parseIfClauseStatement(&stmtList2),"a statement")
	      }	
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
	      p.required(p.Indent(col),"a statement, indented from the 'else'")
	
          // translate
          var stmtList3 []ast.Stmt
          blockPos = p.Pos()
	
	      // parse
	      p.required(p.parseIfClauseStatement(&stmtList3),"a statement")	
	      for ; p.Indent(col) ; {
		 	p.required(p.parseIfClauseStatement(&stmtList3),"a statement")
	      }  
	
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
       }
    }
    return true
}	
	

		
	
/*
    p.required(p.parseExpression(&x),"an expression") 
    p.required(p.Indent(col),"a statement, indented from the 'while'")
    p.required(p.parseLoopBodyStatement(),"a statement")
    for ; p.Indent(col) ; {
       p.required(p.parseLoopBodyStatement(),"a statement")	
   }
   st2 := p.State()
   for ; p.Below(col) ; {
      if p.Match("elif ") {
		 p.required(p.parseExpression(),"an expression") 
	     p.required(p.Indent(col),"a statement, indented from the 'elif'")
	     p.required(p.parseIfClauseStatement(),"a statement")
	     for ; p.Indent(col) ; {
	        p.required(p.parseIfClauseStatement(),"a statement")
	     }	
	     st2 = p.State()
      } else if p.Match("else") {
	     p.required(p.Indent(col),"a statement, indented from the 'else'")
	     p.required(p.parseIfClauseStatement(),"a statement")	
	     for ; p.Indent(col) ; {
			p.required(p.parseIfClauseStatement(),"a statement")
	     }  
	     break       	  
      } else {
	     p.Fail(st2)
      }
   }
   return true
}
*/


func (p *parser) parseForStatement(forStmt **ast.ForStatement) bool {
    if p.trace {
       defer un(trace(p, "ForStatement"))
    }

    return false
}
/*
    st := p.State()
    pos := p.Pos()

    // parse
    col := p.Col()
    if ! p.Match("for ") {
	   return false
    }

    // translate
    // var x ast.Expr

    var keyExpr ast.Expr
    var valueExprs []ast.Expr
    var collectionExprs []ast.Expr


	var stmtList []ast.Stmt



    // parse
    p.required(p.parseExpression(&x),"an expression") 
    p.required(p.Indent(col),"a statement, indented from the 'for'")

    blockPos := p.Pos()

    p.required(p.parseLoopBodyStatement(&stmtList),"a statement")
    for ; p.Indent(col) ; {
       p.required(p.parseLoopBodyStatement(&stmtList),"a statement")	
    }

    // translate
    body := &ast.BlockStatement{blockPos,stmtList}
    *forStmt := &ast.ForStatement{pos, keyLhs, valueLhs, collectionExprs, body}

	For        token.Pos   // position of "for" keyword
	Key        Expr
	Value []Expr           // One or both of Key or Value may be nil
	                       // This is not correct!!! we need to handle multiple values
	X          []Expr        // value to range over NOT CORRECT!!! Need to handle multiple expressions.
	Body       *BlockStatement
*/



/*
for i val in someList
   statement
   statement
*/
func (p *parser) parseForRangeStatement(rangeStmt **ast.RangeStatement) bool {
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
    }

    // declare any undeclared variables
    for _,expr := range keyAndValueVariables {
	   switch expr.(type) {
	      case *ast.Ident:
		     variable := expr.(*ast.Ident)
             p.ensureCurrentScopeVariable(variable, false)
	      default: 
		    fmt.Println("------------IS NOT A VARIABLE ---------------")
		    fmt.Println(expr)
       }
    }   


    // TODO Check that each expression's type is a collection

    p.required(p.Indent(col),"a statement, indented from the 'for'")
    
    blockPos := p.Pos()

    p.required(p.parseLoopBodyStatement(&stmtList),"a statement")
    for ; p.Indent(col) ; {
       p.required(p.parseLoopBodyStatement(&stmtList),"a statement")	
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

   
    if ! p.required(p.parseConstExpr(&expr),"a constant expression (a literal, or a function of literals and constants)") {
    	return p.Fail(st)	
    }

    constDecl := &ast.ConstantDecl{Name:constant,Value:expr}

    *constDecls = append(*constDecls, constDecl)
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
    return (p.Space() && p.parseOneLineLiteral(x)) || p.parseIndentedLiteral(x) 
}

func (p *parser) parseOneLineLiteral(x *ast.Expr) bool {
	if p.trace {
       defer un(trace(p, "OneLineLiteral"))
    }
    return p.parseNumberLiteral(x) || p.parseStringLiteral(x)
}

func (p *parser) parseNumberLiteral(x *ast.Expr) bool {
	if p.trace {
       defer un(trace(p, "NumberLiteral"))
    }
    pos := p.Pos()
    found, tok, lit := p.ScanNumber()
    if ! found {
	   return false // Look for String literals or Boolean literals
    }
    fmt.Printf("%s '%s'\n",tok,lit)

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
    fmt.Printf("String literal \"%s\"\n",lit)

    *x = &ast.BasicLit{pos,token.STRING,lit}
    return true
}


func (p *parser) parseIndentedLiteral(x *ast.Expr) bool {
	if p.trace {
       defer un(trace(p, "IndentedLiteral"))
    }
    return false // TODO
}


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
    	p.error(p.FailedPos(),fmt.Sprintf("Expecting %s.\nFound: %s", whatIsExpected, fs))
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
    	p.error(p.Pos(),fmt.Sprintf("Expecting %s.\nFound: %s", whatIsExpected, found))
	    
    	//p.error(p.Pos(),fmt.Sprintf("%s.", whatIsExpected))
    }
	return false // Will never get here if in single error mode.
}

func (p *parser) optional(elementFound bool) bool {
	return elementFound || true
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

func (p *parser) declare(decl, data interface{}, scope *ast.Scope, kind ast.ObjKind, idents ...*ast.Ident) {
	for _, ident := range idents {
		assert(ident.Obj == nil, "identifier already declared or resolved")
		obj := ast.NewObj(kind, ident.Name)
		// remember the corresponding declaration for redeclaration
		// errors and global variable resolution/typechecking phase
		obj.Decl = decl
		obj.Data = data
		ident.Obj = obj
		if ident.Name != "_" {
			if alt := scope.Insert(obj); alt != nil && p.mode&DeclarationErrors != 0 {
				prevDecl := ""
				if pos := alt.Pos(); pos.IsValid() {
					prevDecl = fmt.Sprintf("\n\tprevious declaration at %s", p.file.Position(pos))
				}
				p.error(ident.Pos(), fmt.Sprintf("%s redeclared in this block%s", ident.Name, prevDecl))
			}
		}
	}
}

func (p *parser) shortVarDecl(idents []*ast.Ident) {
	// Go spec: A short variable declaration may redeclare variables
	// provided they were originally declared in the same block with
	// the same type, and at least one of the non-blank variables is new.
	n := 0 // number of new variables
	for _, ident := range idents {
		assert(ident.Obj == nil, "identifier already declared or resolved")
		obj := ast.NewObj(ast.Var, ident.Name)
		// short var declarations cannot have redeclaration errors
		// and are not global => no need to remember the respective
		// declaration
		ident.Obj = obj
		if ident.Name != "_" {
			if alt := p.topScope.Insert(obj); alt != nil {
				ident.Obj = alt // redeclaration
			} else {
				n++ // new declaration
			}
		}
	}
	if n == 0 && p.mode&DeclarationErrors != 0 {
		p.error(idents[0].Pos(), "no new variables on left side of :=")
	}
}

// The unresolved object is a sentinel to mark identifiers that have been added
// to the list of unresolved identifiers. The sentinel is only used for verifying
// internal consistency.
var unresolved = new(ast.Object)

func (p *parser) resolve(x ast.Expr) {
	// nothing to do if x is not an identifier or the blank identifier
	ident, _ := x.(*ast.Ident)
	if ident == nil {
		return
	}
	assert(ident.Obj == nil, "identifier already declared or resolved")
	if ident.Name == "_" {
		return
	}
	// try to resolve the identifier
	for s := p.topScope; s != nil; s = s.Outer {
		if obj := s.Lookup(ident.Name); obj != nil {
			ident.Obj = obj
			return
		}
	}
	// all local scopes are known, so any unresolved identifier
	// must be found either in the file scope, package scope
	// (perhaps in another file), or universe scope --- collect
	// them so that they can be resolved later
	ident.Obj = unresolved
	p.unresolved = append(p.unresolved, ident)
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

// Advance to the next token.
func (p *parser) next0() {
	// Because of one-token look-ahead, print the previous token
	// when tracing as it provides a more readable output. The
	// very first token (!p.pos.IsValid()) is not initialized
	// (it is token.ILLEGAL), so don't print it .
	if p.trace && p.pos.IsValid() {
		s := p.tok.String()
		switch {
		case p.tok.IsLiteral():
			p.printTrace(s, p.lit)
		case p.tok.IsOperator(), p.tok.IsKeyword():
			p.printTrace("\"" + s + "\"")
		default:
			p.printTrace(s)
		}
	}

	p.pos, p.tok, p.lit = p.Scan()
}

// Consume a comment and return it and the line on which it ends.
func (p *parser) consumeComment() (comment *ast.Comment, endline int) {
	// /*-style comments may end on a different line than where they start.
	// Scan the comment for '\n' chars and adjust endline accordingly.
	endline = p.file.Line(p.pos)
	if p.lit[1] == '*' {
		// don't use range here - no need to decode Unicode code points
		for i := 0; i < len(p.lit); i++ {
			if p.lit[i] == '\n' {
				endline++
			}
		}
	}

	comment = &ast.Comment{p.pos, p.lit}
	p.next0()

	return
}

// Consume a group of adjacent comments, add it to the parser's
// comments list, and return it together with the line at which
// the last comment in the group ends. An empty line or non-comment
// token terminates a comment group.
//
func (p *parser) consumeCommentGroup() (comments *ast.CommentGroup, endline int) {
	var list []*ast.Comment
	endline = p.file.Line(p.pos)
	for p.tok == token.COMMENT && endline+1 >= p.file.Line(p.pos) {
		var comment *ast.Comment
		comment, endline = p.consumeComment()
		list = append(list, comment)
	}

	// add comment group to the comments list
	comments = &ast.CommentGroup{list}
	p.comments = append(p.comments, comments)

	return
}

// Advance to the next non-comment token. In the process, collect
// any comment groups encountered, and remember the last lead and
// and line comments.
//
// A lead comment is a comment group that starts and ends in a
// line without any other tokens and that is followed by a non-comment
// token on the line immediately after the comment group.
//
// A line comment is a comment group that follows a non-comment
// token on the same line, and that has no tokens after it on the line
// where it ends.
//
// Lead and line comments may be considered documentation that is
// stored in the AST.
//
func (p *parser) next() {
	p.leadComment = nil
	p.lineComment = nil
	line := p.file.Line(p.pos) // current line
	p.next0()

	if p.tok == token.COMMENT {
		var comment *ast.CommentGroup
		var endline int

		if p.file.Line(p.pos) == line {
			// The comment is on same line as the previous token; it
			// cannot be a lead comment but may be a line comment.
			comment, endline = p.consumeCommentGroup()
			if p.file.Line(p.pos) != endline {
				// The next token is on a different line, thus
				// the last comment group is a line comment.
				p.lineComment = comment
			}
		}

		// consume successor comments, if any
		endline = -1
		for p.tok == token.COMMENT {
			comment, endline = p.consumeCommentGroup()
		}

		if endline+1 == p.file.Line(p.pos) {
			// The next token is following on the line immediately after the
			// comment group, thus the last comment group is a lead comment.
			p.leadComment = comment
		}
	}
}

func (p *parser) error(pos token.Pos, msg string) {
	p.Error(p.file.Position(pos), msg)
}

func (p *parser) errorExpected(pos token.Pos, msg string) {
	msg = "expected " + msg
	if pos == p.pos {
		// the error happened at the current position;
		// make the error message more specific
		if p.tok == token.SEMICOLON && p.lit[0] == '\n' {
			msg += ", found newline"
		} else {
			msg += ", found '" + p.tok.String() + "'"
			if p.tok.IsLiteral() {
				msg += " " + p.lit
			}
		}
	}
	p.error(pos, msg)
}

func (p *parser) expect(tok token.Token) token.Pos {
	pos := p.pos
	if p.tok != tok {
		p.errorExpected(pos, "'"+tok.String()+"'")
	}
	p.next() // make progress
	return pos
}

func (p *parser) expectSemi() {
	if p.tok != token.RPAREN && p.tok != token.RBRACE {
		p.expect(token.SEMICOLON)
	}
}

func assert(cond bool, msg string) {
	if !cond {
		panic("relish/compiler/parser internal error: " + msg)
	}
}

// ----------------------------------------------------------------------------
// Identifiers

func (p *parser) parseIdent() *ast.Ident {
	pos := p.pos
	name := "_"
	if p.tok == token.IDENT {
		name = p.lit
		p.next()
	} else {
		p.expect(token.IDENT) // use expect() error handling
	}
	return &ast.Ident{pos, name, nil,token.IDENT, -1}
}

func (p *parser) parseIdentList() (list []*ast.Ident) {
	if p.trace {
		defer un(trace(p, "IdentList"))
	}

	list = append(list, p.parseIdent())
	for p.tok == token.COMMA {
		p.next()
		list = append(list, p.parseIdent())
	}

	return
}

// ----------------------------------------------------------------------------
// Common productions

// If lhs is set, result list elements which are identifiers are not resolved.
func (p *parser) parseExprList(lhs bool) (list []ast.Expr) {
	if p.trace {
		defer un(trace(p, "ExpressionList"))
	}

	list = append(list, p.checkExpr(p.parseExpr(lhs)))
	for p.tok == token.COMMA {
		p.next()
		list = append(list, p.checkExpr(p.parseExpr(lhs)))
	}

	return
}

func (p *parser) parseLhsList() []ast.Expr {
	list := p.parseExprList(true)
	switch p.tok {
	case token.DEFINE:
		// lhs of a short variable declaration
		p.shortVarDecl(p.makeIdentList(list))
	case token.COLON:
		// lhs of a label declaration or a communication clause of a select
		// statement (parseLhsList is not called when parsing the case clause
		// of a switch statement):
		// - labels are declared by the caller of parseLhsList
		// - for communication clauses, if there is a stand-alone identifier
		//   followed by a colon, we have a syntax error; there is no need
		//   to resolve the identifier in that case
	default:
		// identifiers must be declared elsewhere
		for _, x := range list {
			p.resolve(x)
		}
	}
	return list
}

func (p *parser) parseRhsList() []ast.Expr {
	return p.parseExprList(false)
}

// ----------------------------------------------------------------------------
// Types

func (p *parser) parseType() ast.Expr {
	if p.trace {
		defer un(trace(p, "Type"))
	}

	typ :=  p.tryType()

	if typ == nil {
		pos := p.pos
		p.errorExpected(pos, "type")
		p.next() // make progress
		return &ast.BadExpr{pos, p.pos}
	}

	return typ
}


// If the result is an identifier, it is not resolved.
func (p *parser) parseTypeNameOld() ast.Expr {
	if p.trace {
		defer un(trace(p, "TypeName"))
	}

	ident := p.parseIdent()
	// don't resolve ident yet - it may be a parameter or field name

	if p.tok == token.PERIOD {
		// ident is a package name
		p.next()
		p.resolve(ident)
		sel := p.parseIdent()
		return &ast.SelectorExpr{ident, sel}
	}

	return ident
}

func (p *parser) parseArrayType(ellipsisOk bool) ast.Expr {
	if p.trace {
		defer un(trace(p, "ArrayType"))
	}

	lbrack := p.expect(token.LBRACK)
	var len ast.Expr
	if ellipsisOk && p.tok == token.ELLIPSIS {
		len = &ast.Ellipsis{p.pos, nil}
		p.next()
	} else if p.tok != token.RBRACK {
		len = p.parseRhs()
	}
	p.expect(token.RBRACK)
	elt := p.parseType()

	return &ast.ArrayType{lbrack, len, elt}
}

func (p *parser) makeIdentList(list []ast.Expr) []*ast.Ident {
	idents := make([]*ast.Ident, len(list))
	for i, x := range list {
		ident, isIdent := x.(*ast.Ident)
		if !isIdent {
			pos := x.(ast.Expr).Pos()
			p.errorExpected(pos, "identifier")
			ident = &ast.Ident{pos, "_", nil,token.VAR, -1}
		}
		idents[i] = ident
	}
	return idents
}

func (p *parser) parseFieldDecl(scope *ast.Scope) *ast.Field {
	if p.trace {
		defer un(trace(p, "FieldDecl"))
	}

	doc := p.leadComment

	// fields
	list, typ := p.parseVarList(false)

	// optional tag
	var tag *ast.BasicLit
	if p.tok == token.STRING {
		tag = &ast.BasicLit{p.pos, p.tok, p.lit}
		p.next()
	}

	// analyze case
	var idents []*ast.Ident
	if typ != nil {
		// IdentifierList Type
		idents = p.makeIdentList(list)
	} else {
		// ["*"] TypeName (AnonymousField)
		typ = list[0] // we always have at least one element
		p.resolve(typ)
		if n := len(list); n > 1 || !isTypeName(deref(typ)) {
			pos := typ.Pos()
			p.errorExpected(pos, "anonymous field")
			typ = &ast.BadExpr{pos, list[n-1].End()}
		}
	}

	p.expectSemi() // call before accessing p.linecomment

	field := &ast.Field{doc, idents, typ, tag, p.lineComment}
	p.declare(field, nil, scope, ast.Var, idents...)

	return field
}

func (p *parser) parseStructType() *ast.StructType {
	if p.trace {
		defer un(trace(p, "StructType"))
	}

	pos := p.expect(token.STRUCT)
	lbrace := p.expect(token.LBRACE)
	scope := ast.NewScope(nil) // struct scope
	var list []*ast.Field
	for p.tok == token.IDENT || p.tok == token.MUL || p.tok == token.LPAREN {
		// a field declaration cannot start with a '(' but we accept
		// it here for more robust parsing and better error messages
		// (parseFieldDecl will check and complain if necessary)
		list = append(list, p.parseFieldDecl(scope))
	}
	rbrace := p.expect(token.RBRACE)

	// TODO(gri): store struct scope in AST
	return &ast.StructType{pos, &ast.FieldList{lbrace, list, rbrace}, false}
}

func (p *parser) parsePointerType() *ast.StarExpr {
	if p.trace {
		defer un(trace(p, "PointerType"))
	}

	star := p.expect(token.MUL)
	base := p.parseType()

	return &ast.StarExpr{star, base}
}


func (p *parser) tryVarType(isParam bool) ast.Expr {
	if isParam && p.tok == token.ELLIPSIS {
		pos := p.pos
		p.next()
		typ := p.tryIdentOrType(isParam) // don't use parseType so we can provide better error message
		if typ == nil {
			p.error(pos, "'...' parameter is missing type")
			typ = &ast.BadExpr{pos, p.pos}
		}
		if p.tok != token.RPAREN {
			p.error(pos, "can use '...' with last parameter type only")
		}
		return &ast.Ellipsis{pos, typ}
	}
	return p.tryIdentOrType(false)
}

func (p *parser) parseVarType(isParam bool) ast.Expr {
	typ := p.tryVarType(isParam)
	if typ == nil {
		pos := p.pos
		p.errorExpected(pos, "type")
		p.next() // make progress
		typ = &ast.BadExpr{pos, p.pos}
	}
	return typ
}

func (p *parser) parseVarList(isParam bool) (list []ast.Expr, typ ast.Expr) {
	if p.trace {
		defer un(trace(p, "VarList"))
	}

	// a list of identifiers looks like a list of type names
	for {
		// parseVarType accepts any type (including parenthesized ones)
		// even though the syntax does not permit them here: we
		// accept them all for more robust parsing and complain
		// afterwards
		list = append(list, p.parseVarType(isParam))
		if p.tok != token.COMMA {
			break
		}
		p.next()
	}

	// if we had a list of identifiers, it must be followed by a type
	typ = p.tryVarType(isParam)
	if typ != nil {
		p.resolve(typ)
	}

	return
}


func (p *parser) parseParameterList(scope *ast.Scope, ellipsisOk bool) (params []*ast.Field) {
	if p.trace {
		defer un(trace(p, "ParameterList"))
	}

	list, typ := p.parseVarList(ellipsisOk)
	if typ != nil {
		// IdentifierList Type
		idents := p.makeIdentList(list)
		field := &ast.Field{nil, idents, typ, nil, nil}
		params = append(params, field)
		// Go spec: The scope of an identifier denoting a function
		// parameter or result variable is the function body.
		p.declare(field, nil, scope, ast.Var, idents...)
		if p.tok == token.COMMA {
			p.next()
		}

		for p.tok != token.RPAREN && p.tok != token.EOF {
			idents := p.parseIdentList()
			typ := p.parseVarType(ellipsisOk)
			field := &ast.Field{nil, idents, typ, nil, nil}
			params = append(params, field)
			// Go spec: The scope of an identifier denoting a function
			// parameter or result variable is the function body.
			p.declare(field, nil, scope, ast.Var, idents...)
			if p.tok != token.COMMA {
				break
			}
			p.next()
		}

	} else {
		// Type { "," Type } (anonymous parameters)
		params = make([]*ast.Field, len(list))
		for i, x := range list {
			p.resolve(x)
			params[i] = &ast.Field{Type: x}
		}
	}

	return
}

func (p *parser) parseParameters(scope *ast.Scope, ellipsisOk bool) *ast.FieldList {
	if p.trace {
		defer un(trace(p, "Parameters"))
	}

	var params []*ast.Field
	lparen := p.expect(token.LPAREN)
	if p.tok != token.RPAREN {
		params = p.parseParameterList(scope, ellipsisOk)
	}
	rparen := p.expect(token.RPAREN)

	return &ast.FieldList{lparen, params, rparen}
}

func (p *parser) parseResult(scope *ast.Scope) *ast.FieldList {
	if p.trace {
		defer un(trace(p, "Result"))
	}

	if p.tok == token.LPAREN {
		return p.parseParameters(scope, false)
	}

	typ := p.tryType()
	if typ != nil {
		list := make([]*ast.Field, 1)
		list[0] = &ast.Field{Type: typ}
		return &ast.FieldList{List: list}
	}

	return nil
}

func (p *parser) parseSignature(scope *ast.Scope) (params, results *ast.FieldList) {
	if p.trace {
		defer un(trace(p, "Signature"))
	}

	params = p.parseParameters(scope, true)
	results = p.parseResult(scope)

	return
}

/*
func (p *parser) parseFuncType() (*ast.FuncType, *ast.Scope) {
	if p.trace {
		defer un(trace(p, "FuncType"))
	}

	pos := p.expect(token.FUNC)
	scope := ast.NewScope(p.topScope) // function scope
	params, results := p.parseSignature(scope)

	return &ast.FuncType{pos, params, results}, scope
}
*/

/*
func (p *parser) parseMethodSpec(scope *ast.Scope) *ast.Field {
	if p.trace {
		defer un(trace(p, "MethodSpec"))
	}

	doc := p.leadComment
	var idents []*ast.Ident
	var typ ast.Expr
	x := p.parseTypeNameOld()
	if ident, isIdent := x.(*ast.Ident); isIdent && p.tok == token.LPAREN {
		// method
		idents = []*ast.Ident{ident}
		scope := ast.NewScope(nil) // method scope
		params, results := p.parseSignature(scope)
		typ = &ast.FuncType{token.NoPos, params, results}
	} else {
		// embedded interface
		typ = x
		p.resolve(typ)
	}
	p.expectSemi() // call before accessing p.linecomment

	spec := &ast.Field{doc, idents, typ, nil, p.lineComment}
	p.declare(spec, nil, scope, ast.Fun, idents...)

	return spec
}
*/

/*
func (p *parser) parseInterfaceType() *ast.InterfaceType {
	if p.trace {
		defer un(trace(p, "InterfaceType"))
	}

	pos := p.expect(token.INTERFACE)
	lbrace := p.expect(token.LBRACE)
	scope := ast.NewScope(nil) // interface scope
	var list []*ast.Field
	for p.tok == token.IDENT {
		list = append(list, p.parseMethodSpec(scope))
	}
	rbrace := p.expect(token.RBRACE)

	// TODO(gri): store interface scope in AST
	return &ast.InterfaceType{pos, &ast.FieldList{lbrace, list, rbrace}, false}
}
*/

func (p *parser) parseMapType() *ast.MapType {
	if p.trace {
		defer un(trace(p, "MapType"))
	}

	pos := p.expect(token.MAP)
	p.expect(token.LBRACK)
	key := p.parseType()
	p.expect(token.RBRACK)
	value := p.parseType()

	return &ast.MapType{pos, key, value}
}

func (p *parser) parseChanType() *ast.ChanType {
	if p.trace {
		defer un(trace(p, "ChanType"))
	}

	pos := p.pos
	dir := ast.SEND | ast.RECV
	if p.tok == token.CHAN {
		p.next()
		if p.tok == token.ARROW {
			p.next()
			dir = ast.SEND
		}
	} else {
		p.expect(token.ARROW)
		p.expect(token.CHAN)
		dir = ast.RECV
	}
	value := p.parseType()

	return &ast.ChanType{pos, dir, value}
}

// If the result is an identifier, it is not resolved.
func (p *parser) tryIdentOrType(ellipsisOk bool) ast.Expr {
	switch p.tok {
	case token.IDENT:
		return p.parseTypeNameOld()
	case token.LBRACK:
		return p.parseArrayType(ellipsisOk)
	case token.STRUCT:
		return p.parseStructType()
	case token.MUL:
		return p.parsePointerType()
//	case token.FUNC:
//		typ, _ := p.parseFuncType()
//		return typ
//	case token.INTERFACE:
//		return p.parseInterfaceType()
	case token.MAP:
		return p.parseMapType()
	case token.CHAN, token.ARROW:
		return p.parseChanType()
	case token.LPAREN:
		lparen := p.pos
		p.next()
		typ := p.parseType()
		rparen := p.expect(token.RPAREN)
		return &ast.ParenExpr{lparen, typ, rparen}
	}

	// no type found
	return nil
}

func (p *parser) tryType() ast.Expr {
	typ := p.tryIdentOrType(false)
	if typ != nil {
		p.resolve(typ)
	}
	return typ
}


// ----------------------------------------------------------------------------
// Blocks

func (p *parser) parseStmtList() (list []ast.Stmt) {
	if p.trace {
		defer un(trace(p, "StatementList"))
	}

	for p.tok != token.CASE && p.tok != token.DEFAULT && p.tok != token.RBRACE && p.tok != token.EOF {
		list = append(list, p.parseStmt())
	}

	return
}

func (p *parser) parseBody(scope *ast.Scope) *ast.BlockStmt {
	if p.trace {
		defer un(trace(p, "Body"))
	}

	lbrace := p.expect(token.LBRACE)
	p.topScope = scope // open function scope
	p.openLabelScope()
	list := p.parseStmtList()
	p.closeLabelScope()
	p.closeScope()
	rbrace := p.expect(token.RBRACE)

	return &ast.BlockStmt{lbrace, list, rbrace}
}

func (p *parser) parseBlockStmt() *ast.BlockStmt {
	if p.trace {
		defer un(trace(p, "BlockStmt"))
	}

	lbrace := p.expect(token.LBRACE)
	p.openScope()
	list := p.parseStmtList()
	p.closeScope()
	rbrace := p.expect(token.RBRACE)

	return &ast.BlockStmt{lbrace, list, rbrace}
}

// ----------------------------------------------------------------------------
// Expressions

/*
func (p *parser) parseFuncTypeOrLit() ast.Expr {
	if p.trace {
		defer un(trace(p, "FuncTypeOrLit"))
	}

	typ, scope := p.parseFuncType()
	if p.tok != token.LBRACE {
		// function type only
		return typ
	}

	p.exprLev++
	body := p.parseBody(scope)
	p.exprLev--

	return &ast.FuncLit{typ, body}
}
*/

// parseOperand may return an expression or a raw type (incl. array
// types of the form [...]T. Callers must verify the result.
// If lhs is set and the result is an identifier, it is not resolved.
//
func (p *parser) parseOperand(lhs bool) ast.Expr {
	if p.trace {
		defer un(trace(p, "Operand"))
	}

	switch p.tok {
	case token.IDENT:
		x := p.parseIdent()
		if !lhs {
			p.resolve(x)
		}
		return x

	case token.INT, token.FLOAT, token.IMAG, token.CHAR, token.STRING:
		x := &ast.BasicLit{p.pos, p.tok, p.lit}
		p.next()
		return x

	case token.LPAREN:
		lparen := p.pos
		p.next()
		p.exprLev++
		x := p.parseRhsOrType() // types may be parenthesized: (some type)
		p.exprLev--
		rparen := p.expect(token.RPAREN)
		return &ast.ParenExpr{lparen, x, rparen}

//	case token.FUNC:
//		return p.parseFuncTypeOrLit()

	default:
		if typ := p.tryIdentOrType(true); typ != nil {
			// could be type for composite literal or conversion
			_, isIdent := typ.(*ast.Ident)
			assert(!isIdent, "type cannot be identifier")
			return typ
		}
	}

	pos := p.pos
	p.errorExpected(pos, "operand")
	p.next() // make progress
	return &ast.BadExpr{pos, p.pos}
}

func (p *parser) parseSelector(x ast.Expr) ast.Expr {
	if p.trace {
		defer un(trace(p, "Selector"))
	}

	sel := p.parseIdent()

	return &ast.SelectorExpr{x, sel}
}

func (p *parser) parseTypeAssertion(x ast.Expr) ast.Expr {
	if p.trace {
		defer un(trace(p, "TypeAssertion"))
	}

	p.expect(token.LPAREN)
	var typ ast.Expr
	if p.tok == token.TYPE {
		// type switch: typ == nil
		p.next()
	} else {
		typ = p.parseType()
	}
	p.expect(token.RPAREN)

	return &ast.TypeAssertExpr{x, typ}
}

func (p *parser) parseIndexOrSlice(x ast.Expr) ast.Expr {
	if p.trace {
		defer un(trace(p, "IndexOrSlice"))
	}

	lbrack := p.expect(token.LBRACK)
	p.exprLev++
	var low, high ast.Expr
	isSlice := false
	if p.tok != token.COLON {
		low = p.parseRhs()
	}
	if p.tok == token.COLON {
		isSlice = true
		p.next()
		if p.tok != token.RBRACK {
			high = p.parseRhs()
		}
	}
	p.exprLev--
	rbrack := p.expect(token.RBRACK)

	if isSlice {
		return &ast.SliceExpr{x, lbrack, low, high, rbrack}
	}
	return &ast.IndexExpr{x, lbrack, low, rbrack}
}

func (p *parser) parseCallOrConversion(fun ast.Expr) *ast.CallExpr {
	if p.trace {
		defer un(trace(p, "CallOrConversion"))
	}

	lparen := p.expect(token.LPAREN)
	p.exprLev++
	var list []ast.Expr
	var ellipsis token.Pos
	for p.tok != token.RPAREN && p.tok != token.EOF && !ellipsis.IsValid() {
		list = append(list, p.parseRhsOrType()) // builtins may expect a type: make(some type, ...)
		if p.tok == token.ELLIPSIS {
			ellipsis = p.pos
			p.next()
		}
		if p.tok != token.COMMA {
			break
		}
		p.next()
	}
	p.exprLev--
	rparen := p.expect(token.RPAREN)

	return &ast.CallExpr{fun, lparen, list, ellipsis, rparen}
}

func (p *parser) parseElement(keyOk bool) ast.Expr {
	if p.trace {
		defer un(trace(p, "Element"))
	}

	if p.tok == token.LBRACE {
		return p.parseLiteralValue(nil)
	}

	x := p.checkExpr(p.parseExpr(keyOk)) // don't resolve if map key
	if keyOk {
		if p.tok == token.COLON {
			colon := p.pos
			p.next()
			return &ast.KeyValueExpr{x, colon, p.parseElement(false)}
		}
		p.resolve(x) // not a map key
	}

	return x
}

func (p *parser) parseElementList() (list []ast.Expr) {
	if p.trace {
		defer un(trace(p, "ElementList"))
	}

	for p.tok != token.RBRACE && p.tok != token.EOF {
		list = append(list, p.parseElement(true))
		if p.tok != token.COMMA {
			break
		}
		p.next()
	}

	return
}

func (p *parser) parseLiteralValue(typ ast.Expr) ast.Expr {
	if p.trace {
		defer un(trace(p, "LiteralValue"))
	}

	lbrace := p.expect(token.LBRACE)
	var elts []ast.Expr
	p.exprLev++
	if p.tok != token.RBRACE {
		elts = p.parseElementList()
	}
	p.exprLev--
	rbrace := p.expect(token.RBRACE)
	return &ast.CompositeLit{typ, lbrace, elts, rbrace}
}

// checkExpr checks that x is an expression (and not a type).
func (p *parser) checkExpr(x ast.Expr) ast.Expr {
	switch unparen(x).(type) {
	case *ast.BadExpr:
	case *ast.Ident:
	case *ast.BasicLit:
	case *ast.FuncLit:
	case *ast.CompositeLit:
	case *ast.ParenExpr:
		panic("unreachable")
	case *ast.SelectorExpr:
	case *ast.IndexExpr:
	case *ast.SliceExpr:
	case *ast.TypeAssertExpr:
		// If t.Type == nil we have a type assertion of the form
		// y.(type), which is only allowed in type switch expressions.
		// It's hard to exclude those but for the case where we are in
		// a type switch. Instead be lenient and test this in the type
		// checker.
	case *ast.CallExpr:
	case *ast.StarExpr:
	case *ast.UnaryExpr:
	case *ast.BinaryExpr:
	default:
		// all other nodes are not proper expressions
		p.errorExpected(x.Pos(), "expression")
		x = &ast.BadExpr{x.Pos(), x.End()}
	}
	return x
}

// isTypeName returns true iff x is a (qualified) TypeName.
func isTypeName(x ast.Expr) bool {
	switch t := x.(type) {
	case *ast.BadExpr:
	case *ast.Ident:
	case *ast.SelectorExpr:
		_, isIdent := t.X.(*ast.Ident)
		return isIdent
	default:
		return false // all other nodes are not type names
	}
	return true
}

// isLiteralType returns true iff x is a legal composite literal type.
func isLiteralType(x ast.Expr) bool {
	switch t := x.(type) {
	case *ast.BadExpr:
	case *ast.Ident:
	case *ast.SelectorExpr:
		_, isIdent := t.X.(*ast.Ident)
		return isIdent
	case *ast.ArrayType:
	case *ast.StructType:
	case *ast.MapType:
	default:
		return false // all other nodes are not legal composite literal types
	}
	return true
}

// If x is of the form *T, deref returns T, otherwise it returns x.
func deref(x ast.Expr) ast.Expr {
	if p, isPtr := x.(*ast.StarExpr); isPtr {
		x = p.X
	}
	return x
}

// If x is of the form (T), unparen returns unparen(T), otherwise it returns x.
func unparen(x ast.Expr) ast.Expr {
	if p, isParen := x.(*ast.ParenExpr); isParen {
		x = unparen(p.X)
	}
	return x
}

// checkExprOrType checks that x is an expression or a type
// (and not a raw type such as [...]T).
//
func (p *parser) checkExprOrType(x ast.Expr) ast.Expr {
	switch t := unparen(x).(type) {
	case *ast.ParenExpr:
		panic("unreachable")
	case *ast.UnaryExpr:
	case *ast.ArrayType:
		if len, isEllipsis := t.Len.(*ast.Ellipsis); isEllipsis {
			p.error(len.Pos(), "expected array length, found '...'")
			x = &ast.BadExpr{x.Pos(), x.End()}
		}
	}

	// all other nodes are expressions or types
	return x
}

// If lhs is set and the result is an identifier, it is not resolved.
func (p *parser) parsePrimaryExpr(lhs bool) ast.Expr {
	if p.trace {
		defer un(trace(p, "PrimaryExpr"))
	}

	x := p.parseOperand(lhs)
L:
	for {
		switch p.tok {
		case token.PERIOD:
			p.next()
			
			if lhs {
				p.resolve(x)
			}
			switch p.tok {
			case token.IDENT:
				x = p.parseSelector(p.checkExpr(x))
			case token.LPAREN:
				x = p.parseTypeAssertion(p.checkExpr(x))
			default:
				pos := p.pos
				p.next() // make progress
				p.errorExpected(pos, "selector or type assertion")
				x = &ast.BadExpr{pos, p.pos}
			}
		case token.LBRACK:
			if lhs {
				p.resolve(x)
			}
			x = p.parseIndexOrSlice(p.checkExpr(x))
		case token.LPAREN:
			if lhs {
				p.resolve(x)
			}
			x = p.parseCallOrConversion(p.checkExprOrType(x))
		case token.LBRACE:
			if isLiteralType(x) && (p.exprLev >= 0 || !isTypeName(x)) {
				if lhs {
					p.resolve(x)
				}
				x = p.parseLiteralValue(x)
			} else {
				break L
			}
		default:
			break L
		}
		lhs = false // no need to try to resolve again
	}

	return x
}

// If lhs is set and the result is an identifier, it is not resolved.
func (p *parser) parseUnaryExpr(lhs bool) ast.Expr {
	if p.trace {
		defer un(trace(p, "UnaryExpr"))
	}

	switch p.tok {
	case token.ADD, token.SUB, token.NOT, token.XOR, token.AND:
		pos, op := p.pos, p.tok
		p.next()
		x := p.parseUnaryExpr(false)
		return &ast.UnaryExpr{pos, op, p.checkExpr(x)}

	case token.ARROW:
		// channel type or receive expression
		pos := p.pos
		p.next()
		if p.tok == token.CHAN {
			p.next()
			value := p.parseType()
			return &ast.ChanType{pos, ast.RECV, value}
		}

		x := p.parseUnaryExpr(false)
		return &ast.UnaryExpr{pos, token.ARROW, p.checkExpr(x)}

	case token.MUL:
		// pointer type or unary "*" expression
		pos := p.pos
		p.next()
		x := p.parseUnaryExpr(false)
		return &ast.StarExpr{pos, p.checkExprOrType(x)}
	}

	return p.parsePrimaryExpr(lhs)
}

// If lhs is set and the result is an identifier, it is not resolved.
func (p *parser) parseBinaryExpr(lhs bool, prec1 int) ast.Expr {
	if p.trace {
		defer un(trace(p, "BinaryExpr"))
	}

	x := p.parseUnaryExpr(lhs)
	for prec := p.tok.Precedence(); prec >= prec1; prec-- {
		for p.tok.Precedence() == prec {
			pos, op := p.pos, p.tok
			p.next()
			if lhs {
				p.resolve(x)
				lhs = false
			}
			y := p.parseBinaryExpr(false, prec+1)
			x = &ast.BinaryExpr{p.checkExpr(x), pos, op, p.checkExpr(y)}
		}
	}

	return x
}

// If lhs is set and the result is an identifier, it is not resolved.
// The result may be a type or even a raw type ([...]int). Callers must
// check the result (using checkExpr or checkExprOrType), depending on
// context.
func (p *parser) parseExpr(lhs bool) ast.Expr {
	if p.trace {
		defer un(trace(p, "Expression"))
	}

	return p.parseBinaryExpr(lhs, token.LowestPrec+1)
}

func (p *parser) parseRhs() ast.Expr {
	return p.checkExpr(p.parseExpr(false))
}

func (p *parser) parseRhsOrType() ast.Expr {
	return p.checkExprOrType(p.parseExpr(false))
}

// ----------------------------------------------------------------------------
// Statements

// Parsing modes for parseSimpleStmt.
const (
	basic = iota
	labelOk
	rangeOk
)

// parseSimpleStmt returns true as 2nd result if it parsed the assignment
// of a range clause (with mode == rangeOk). The returned statement is an
// assignment with a right-hand side that is a single unary expression of
// the form "range x". No guarantees are given for the left-hand side.
func (p *parser) parseSimpleStmt(mode int) (ast.Stmt, bool) {
	if p.trace {
		defer un(trace(p, "SimpleStmt"))
	}

	x := p.parseLhsList()

	switch p.tok {
	case
		token.DEFINE, token.ASSIGN, token.ADD_ASSIGN,
		token.SUB_ASSIGN, token.MUL_ASSIGN, token.QUO_ASSIGN,
		token.REM_ASSIGN, token.AND_ASSIGN, token.OR_ASSIGN,
		token.XOR_ASSIGN, token.SHL_ASSIGN, token.SHR_ASSIGN, token.AND_NOT_ASSIGN:
		// assignment statement, possibly part of a range clause
		pos, tok := p.pos, p.tok
		p.next()
		var y []ast.Expr
		isRange := false
		if mode == rangeOk && p.tok == token.RANGE && (tok == token.DEFINE || tok == token.ASSIGN) {
			pos := p.pos
			p.next()
			y = []ast.Expr{&ast.UnaryExpr{pos, token.RANGE, p.parseRhs()}}
			isRange = true
		} else {
			y = p.parseRhsList()
		}
		return &ast.AssignStmt{x, pos, tok, y}, isRange
	}

	if len(x) > 1 {
		p.errorExpected(x[0].Pos(), "1 expression")
		// continue with first expression
	}

	switch p.tok {
	case token.COLON:
		// labeled statement
		colon := p.pos
		p.next()
		if label, isIdent := x[0].(*ast.Ident); mode == labelOk && isIdent {
			// Go spec: The scope of a label is the body of the function
			// in which it is declared and excludes the body of any nested
			// function.
			stmt := &ast.LabeledStmt{label, colon, p.parseStmt()}
			p.declare(stmt, nil, p.labelScope, ast.Lbl, label)
			return stmt, false
		}
		// The label declaration typically starts at x[0].Pos(), but the label
		// declaration may be erroneous due to a token after that position (and
		// before the ':'). If SpuriousErrors is not set, the (only) error re-
		// ported for the line is the illegal label error instead of the token
		// before the ':' that caused the problem. Thus, use the (latest) colon
		// position for error reporting.
		p.error(colon, "illegal label declaration")
		return &ast.BadStmt{x[0].Pos(), colon + 1}, false

	case token.ARROW:
		// send statement
		arrow := p.pos
		p.next() // consume "<-"
		y := p.parseRhs()
		return &ast.SendStmt{x[0], arrow, y}, false

	case token.INC, token.DEC:
		// increment or decrement
		s := &ast.IncDecStmt{x[0], p.pos, p.tok}
		p.next() // consume "++" or "--"
		return s, false
	}

	// expression
	return &ast.ExprStmt{x[0]}, false
}

func (p *parser) parseCallExpr() *ast.CallExpr {
	x := p.parseRhsOrType() // could be a conversion: (some type)(x)
	if call, isCall := x.(*ast.CallExpr); isCall {
		return call
	}
	p.errorExpected(x.Pos(), "function/method call")
	return nil
}

func (p *parser) parseGoStmt() ast.Stmt {
	if p.trace {
		defer un(trace(p, "GoStmt"))
	}

	pos := p.expect(token.GO)
	call := p.parseCallExpr()
	p.expectSemi()
	if call == nil {
		return &ast.BadStmt{pos, pos + 2} // len("go")
	}

	return &ast.GoStmt{pos, call}
}

func (p *parser) parseDeferStmt() ast.Stmt {
	if p.trace {
		defer un(trace(p, "DeferStmt"))
	}

	pos := p.expect(token.DEFER)
	call := p.parseCallExpr()
	p.expectSemi()
	if call == nil {
		return &ast.BadStmt{pos, pos + 5} // len("defer")
	}

	return &ast.DeferStmt{pos, call}
}

func (p *parser) parseReturnStmt() *ast.ReturnStmt {
	if p.trace {
		defer un(trace(p, "ReturnStmt"))
	}

	pos := p.pos
	p.expect(token.RETURN)
	var x []ast.Expr
	if p.tok != token.SEMICOLON && p.tok != token.RBRACE {
		x = p.parseRhsList()
	}
	p.expectSemi()

	return &ast.ReturnStmt{pos, x}
}

func (p *parser) parseBranchStmt(tok token.Token) *ast.BranchStmt {
	if p.trace {
		defer un(trace(p, "BranchStmt"))
	}

	pos := p.expect(tok)
	var label *ast.Ident
	if tok != token.FALLTHROUGH && p.tok == token.IDENT {
		label = p.parseIdent()
		// add to list of unresolved targets
		n := len(p.targetStack) - 1
		p.targetStack[n] = append(p.targetStack[n], label)
	}
	p.expectSemi()

	return &ast.BranchStmt{pos, tok, label}
}

func (p *parser) makeExpr(s ast.Stmt) ast.Expr {
	if s == nil {
		return nil
	}
	if es, isExpr := s.(*ast.ExprStmt); isExpr {
		return p.checkExpr(es.X)
	}
	p.error(s.Pos(), "expected condition, found simple statement")
	return &ast.BadExpr{s.Pos(), s.End()}
}

func (p *parser) parseIfStmt() *ast.IfStmt {
	if p.trace {
		defer un(trace(p, "IfStmt"))
	}

	pos := p.expect(token.IF)
	p.openScope()
	defer p.closeScope()

	var s ast.Stmt
	var x ast.Expr
	{
		prevLev := p.exprLev
		p.exprLev = -1
		if p.tok == token.SEMICOLON {
			p.next()
			x = p.parseRhs()
		} else {
			s, _ = p.parseSimpleStmt(basic)
			if p.tok == token.SEMICOLON {
				p.next()
				x = p.parseRhs()
			} else {
				x = p.makeExpr(s)
				s = nil
			}
		}
		p.exprLev = prevLev
	}

	body := p.parseBlockStmt()
	var else_ ast.Stmt
	if p.tok == token.ELSE {
		p.next()
		else_ = p.parseStmt()
	} else {
		p.expectSemi()
	}

	return &ast.IfStmt{pos, s, x, body, else_}
}

func (p *parser) parseTypeList() (list []ast.Expr) {
	if p.trace {
		defer un(trace(p, "TypeList"))
	}

	list = append(list, p.parseType())
	for p.tok == token.COMMA {
		p.next()
		list = append(list, p.parseType())
	}

	return
}

func (p *parser) parseCaseClause(exprSwitch bool) *ast.CaseClause {
	if p.trace {
		defer un(trace(p, "CaseClause"))
	}

	pos := p.pos
	var list []ast.Expr
	if p.tok == token.CASE {
		p.next()
		if exprSwitch {
			list = p.parseRhsList()
		} else {
			list = p.parseTypeList()
		}
	} else {
		p.expect(token.DEFAULT)
	}

	colon := p.expect(token.COLON)
	p.openScope()
	body := p.parseStmtList()
	p.closeScope()

	return &ast.CaseClause{pos, list, colon, body}
}

func isExprSwitch(s ast.Stmt) bool {
	if s == nil {
		return true
	}
	if e, ok := s.(*ast.ExprStmt); ok {
		if a, ok := e.X.(*ast.TypeAssertExpr); ok {
			return a.Type != nil // regular type assertion
		}
		return true
	}
	return false
}

func (p *parser) parseSwitchStmt() ast.Stmt {
	if p.trace {
		defer un(trace(p, "SwitchStmt"))
	}

	pos := p.expect(token.SWITCH)
	p.openScope()
	defer p.closeScope()

	var s1, s2 ast.Stmt
	if p.tok != token.LBRACE {
		prevLev := p.exprLev
		p.exprLev = -1
		if p.tok != token.SEMICOLON {
			s2, _ = p.parseSimpleStmt(basic)
		}
		if p.tok == token.SEMICOLON {
			p.next()
			s1 = s2
			s2 = nil
			if p.tok != token.LBRACE {
				s2, _ = p.parseSimpleStmt(basic)
			}
		}
		p.exprLev = prevLev
	}

	exprSwitch := isExprSwitch(s2)
	lbrace := p.expect(token.LBRACE)
	var list []ast.Stmt
	for p.tok == token.CASE || p.tok == token.DEFAULT {
		list = append(list, p.parseCaseClause(exprSwitch))
	}
	rbrace := p.expect(token.RBRACE)
	p.expectSemi()
	body := &ast.BlockStmt{lbrace, list, rbrace}

	if exprSwitch {
		return &ast.SwitchStmt{pos, s1, p.makeExpr(s2), body}
	}
	// type switch
	// TODO(gri): do all the checks!
	return &ast.TypeSwitchStmt{pos, s1, s2, body}
}

func (p *parser) parseCommClause() *ast.CommClause {
	if p.trace {
		defer un(trace(p, "CommClause"))
	}

	p.openScope()
	pos := p.pos
	var comm ast.Stmt
	if p.tok == token.CASE {
		p.next()
		lhs := p.parseLhsList()
		if p.tok == token.ARROW {
			// SendStmt
			if len(lhs) > 1 {
				p.errorExpected(lhs[0].Pos(), "1 expression")
				// continue with first expression
			}
			arrow := p.pos
			p.next()
			rhs := p.parseRhs()
			comm = &ast.SendStmt{lhs[0], arrow, rhs}
		} else {
			// RecvStmt
			pos := p.pos
			tok := p.tok
			var rhs ast.Expr
			if tok == token.ASSIGN || tok == token.DEFINE {
				// RecvStmt with assignment
				if len(lhs) > 2 {
					p.errorExpected(lhs[0].Pos(), "1 or 2 expressions")
					// continue with first two expressions
					lhs = lhs[0:2]
				}
				p.next()
				rhs = p.parseRhs()
			} else {
				// rhs must be single receive operation
				if len(lhs) > 1 {
					p.errorExpected(lhs[0].Pos(), "1 expression")
					// continue with first expression
				}
				rhs = lhs[0]
				lhs = nil // there is no lhs
			}
			if lhs != nil {
				comm = &ast.AssignStmt{lhs, pos, tok, []ast.Expr{rhs}}
			} else {
				comm = &ast.ExprStmt{rhs}
			}
		}
	} else {
		p.expect(token.DEFAULT)
	}

	colon := p.expect(token.COLON)
	body := p.parseStmtList()
	p.closeScope()

	return &ast.CommClause{pos, comm, colon, body}
}

func (p *parser) parseSelectStmt() *ast.SelectStmt {
	if p.trace {
		defer un(trace(p, "SelectStmt"))
	}

	pos := p.expect(token.SELECT)
	lbrace := p.expect(token.LBRACE)
	var list []ast.Stmt
	for p.tok == token.CASE || p.tok == token.DEFAULT {
		list = append(list, p.parseCommClause())
	}
	rbrace := p.expect(token.RBRACE)
	p.expectSemi()
	body := &ast.BlockStmt{lbrace, list, rbrace}

	return &ast.SelectStmt{pos, body}
}

func (p *parser) parseForStmt() ast.Stmt {
	if p.trace {
		defer un(trace(p, "ForStmt"))
	}

	pos := p.expect(token.FOR)
	p.openScope()
	defer p.closeScope()

	var s1, s2, s3 ast.Stmt
	var isRange bool
	if p.tok != token.LBRACE {
		prevLev := p.exprLev
		p.exprLev = -1
		if p.tok != token.SEMICOLON {
			s2, isRange = p.parseSimpleStmt(rangeOk)
		}
		if !isRange && p.tok == token.SEMICOLON {
			p.next()
			s1 = s2
			s2 = nil
			if p.tok != token.SEMICOLON {
				s2, _ = p.parseSimpleStmt(basic)
			}
			p.expectSemi()
			if p.tok != token.LBRACE {
				s3, _ = p.parseSimpleStmt(basic)
			}
		}
		p.exprLev = prevLev
	}

	body := p.parseBlockStmt()
	p.expectSemi()

	if isRange {
		as := s2.(*ast.AssignStmt)
		// check lhs
		var key, value ast.Expr
		switch len(as.Lhs) {
		case 2:
			key, value = as.Lhs[0], as.Lhs[1]
		case 1:
			key = as.Lhs[0]
		default:
			p.errorExpected(as.Lhs[0].Pos(), "1 or 2 expressions")
			return &ast.BadStmt{pos, body.End()}
		}
		// parseSimpleStmt returned a right-hand side that
		// is a single unary expression of the form "range x"
		x := as.Rhs[0].(*ast.UnaryExpr).X
		return &ast.RangeStmt{pos, key, value, as.TokPos, as.Tok, x, body}
	}

	// regular for statement
	return &ast.ForStmt{pos, s1, p.makeExpr(s2), s3, body}
}

func (p *parser) parseStmt() (s ast.Stmt) {
	if p.trace {
		defer un(trace(p, "Statement"))
	}

	switch p.tok {
	case token.CONST, token.TYPE, token.VAR:
		s = &ast.DeclStmt{p.parseDecl()}
	case
		// tokens that may start a top-level expression
		token.IDENT, token.INT, token.FLOAT, token.CHAR, token.STRING, token.FUNC, token.LPAREN, // operand
		token.LBRACK, token.STRUCT, // composite type
		token.MUL, token.AND, token.ARROW, token.ADD, token.SUB, token.XOR: // unary operators
		s, _ = p.parseSimpleStmt(labelOk)
		// because of the required look-ahead, labeled statements are
		// parsed by parseSimpleStmt - don't expect a semicolon after
		// them
		if _, isLabeledStmt := s.(*ast.LabeledStmt); !isLabeledStmt {
			p.expectSemi()
		}
	case token.GO:
		s = p.parseGoStmt()
	case token.DEFER:
		s = p.parseDeferStmt()
	case token.RETURN:
		s = p.parseReturnStmt()
	case token.BREAK, token.CONTINUE, token.GOTO, token.FALLTHROUGH:
		s = p.parseBranchStmt(p.tok)
	case token.LBRACE:
		s = p.parseBlockStmt()
		p.expectSemi()
	case token.IF:
		s = p.parseIfStmt()
	case token.SWITCH:
		s = p.parseSwitchStmt()
	case token.SELECT:
		s = p.parseSelectStmt()
	case token.FOR:
		s = p.parseForStmt()
	case token.SEMICOLON:
		s = &ast.EmptyStmt{p.pos}
		p.next()
	case token.RBRACE:
		// a semicolon may be omitted before a closing "}"
		s = &ast.EmptyStmt{p.pos}
	default:
		// no statement found
		pos := p.pos
		p.errorExpected(pos, "statement")
		p.next() // make progress
		s = &ast.BadStmt{pos, p.pos}
	}

	return
}

// ----------------------------------------------------------------------------
// Declarations

type parseSpecFunction func(p *parser, doc *ast.CommentGroup, iota int) ast.Spec

func parseImportSpec(p *parser, doc *ast.CommentGroup, _ int) ast.Spec {
	if p.trace {
		defer un(trace(p, "ImportSpec"))
	}

	var ident *ast.Ident
	switch p.tok {
	case token.PERIOD:
		ident = &ast.Ident{p.pos, ".", nil,token.VAR, -1}
		p.next()
	case token.IDENT:
		ident = p.parseIdent()
	}

	var path *ast.BasicLit
	if p.tok == token.STRING {
		path = &ast.BasicLit{p.pos, p.tok, p.lit}
		p.next()
	} else {
		p.expect(token.STRING) // use expect() error handling
	}
	p.expectSemi() // call before accessing p.linecomment

	// collect imports
	spec := &ast.ImportSpec{doc, ident, path, p.lineComment}
	p.imports = append(p.imports, spec)

	return spec
}

func parseConstSpec(p *parser, doc *ast.CommentGroup, iota int) ast.Spec {
	if p.trace {
		defer un(trace(p, "ConstSpec"))
	}

	idents := p.parseIdentList()
	typ := p.tryType()
	var values []ast.Expr
	if typ != nil || p.tok == token.ASSIGN || iota == 0 {
		p.expect(token.ASSIGN)
		values = p.parseRhsList()
	}
	p.expectSemi() // call before accessing p.linecomment

	// Go spec: The scope of a constant or variable identifier declared inside
	// a function begins at the end of the ConstSpec or VarSpec and ends at
	// the end of the innermost containing block.
	// (Global identifiers are resolved in a separate phase after parsing.)
	spec := &ast.ValueSpec{doc, idents, typ, values, p.lineComment}
	p.declare(spec, iota, p.topScope, ast.Con, idents...)

	return spec
}

func parseTypeSpec(p *parser, doc *ast.CommentGroup, _ int) ast.Spec {
	if p.trace {
		defer un(trace(p, "TypeSpec"))
	}

	ident := p.parseIdent()

	// Go spec: The scope of a type identifier declared inside a function begins
	// at the identifier in the TypeSpec and ends at the end of the innermost
	// containing block.
	// (Global identifiers are resolved in a separate phase after parsing.)
	spec := &ast.TypeSpec{Doc:doc, Name:ident, Type:nil, Comment:nil}
	p.declare(spec, nil, p.topScope, ast.Typ, ident)

	spec.Type = p.parseType()
	p.expectSemi() // call before accessing p.linecomment
	spec.Comment = p.lineComment

	return spec
}

func parseVarSpec(p *parser, doc *ast.CommentGroup, _ int) ast.Spec {
	if p.trace {
		defer un(trace(p, "VarSpec"))
	}

	idents := p.parseIdentList()
	typ := p.tryType()
	var values []ast.Expr
	if typ == nil || p.tok == token.ASSIGN {
		p.expect(token.ASSIGN)
		values = p.parseRhsList()
	}
	p.expectSemi() // call before accessing p.linecomment

	// Go spec: The scope of a constant or variable identifier declared inside
	// a function begins at the end of the ConstSpec or VarSpec and ends at
	// the end of the innermost containing block.
	// (Global identifiers are resolved in a separate phase after parsing.)
	spec := &ast.ValueSpec{doc, idents, typ, values, p.lineComment}
	p.declare(spec, nil, p.topScope, ast.Var, idents...)

	return spec
}

func (p *parser) parseGenDecl(keyword token.Token, f parseSpecFunction) *ast.GenDecl {
	if p.trace {
		defer un(trace(p, "GenDecl("+keyword.String()+")"))
	}

	doc := p.leadComment
	pos := p.expect(keyword)
	var lparen, rparen token.Pos
	var list []ast.Spec
	if p.tok == token.LPAREN {
		lparen = p.pos
		p.next()
		for iota := 0; p.tok != token.RPAREN && p.tok != token.EOF; iota++ {
			list = append(list, f(p, p.leadComment, iota))
		}
		rparen = p.expect(token.RPAREN)
		p.expectSemi()
	} else {
		list = append(list, f(p, nil, 0))
	}

	return &ast.GenDecl{doc, pos, keyword, lparen, list, rparen}
}

func (p *parser) parseReceiver(scope *ast.Scope) *ast.FieldList {
	if p.trace {
		defer un(trace(p, "Receiver"))
	}

	pos := p.pos
	par := p.parseParameters(scope, false)

	// must have exactly one receiver
	if par.NumFields() != 1 {
		p.errorExpected(pos, "exactly one receiver")
		// TODO determine a better range for BadExpr below
		par.List = []*ast.Field{&ast.Field{Type: &ast.BadExpr{pos, pos}}}
		return par
	}

	// recv type must be of the form ["*"] identifier
	recv := par.List[0]
	base := deref(recv.Type)
	if _, isIdent := base.(*ast.Ident); !isIdent {
		p.errorExpected(base.Pos(), "(unqualified) identifier")
		par.List = []*ast.Field{&ast.Field{Type: &ast.BadExpr{recv.Pos(), recv.End()}}}
	}

	return par
}
/*

func (p *parser) parseFuncDecl() *ast.FuncDecl {
	if p.trace {
		defer un(trace(p, "FunctionDecl"))
	}

	doc := p.leadComment
	pos := p.expect(token.FUNC)
	scope := ast.NewScope(p.topScope) // function scope

	var recv *ast.FieldList
	if p.tok == token.LPAREN {
		recv = p.parseReceiver(scope)
	}

	ident := p.parseIdent()

	params, results := p.parseSignature(scope)

	var body *ast.BlockStmt
	if p.tok == token.LBRACE {
		body = p.parseBody(scope)
	}
	p.expectSemi()

	decl := &ast.FuncDecl{doc, recv, ident, &ast.FuncType{pos, params, results}, body}
	if recv == nil {
		// Go spec: The scope of an identifier denoting a constant, type,
		// variable, or function (but not method) declared at top level
		// (outside any function) is the package block.
		//
		// init() functions cannot be referred to and there may
		// be more than one - don't put them in the pkgScope
		if ident.Name != "init" {
			p.declare(decl, nil, p.pkgScope, ast.Fun, ident)
		}
	}

	return decl
}
*/

func (p *parser) parseDecl() ast.Decl {
	if p.trace {
		defer un(trace(p, "Declaration"))
	}

	var f parseSpecFunction
	switch p.tok {
	case token.CONST:
		f = parseConstSpec

	case token.TYPE:
		f = parseTypeSpec

	case token.VAR:
		f = parseVarSpec

//	case token.FUNC:
//		return p.parseFuncDecl()

	default:
		pos := p.pos
		p.errorExpected(pos, "declaration")
		p.next() // make progress
		decl := &ast.BadDecl{pos, p.pos}
		return decl
	}

	return p.parseGenDecl(p.tok, f)
}

func (p *parser) parseDeclList() (list []ast.Decl) {
	if p.trace {
		defer un(trace(p, "DeclList"))
	}

	for p.tok != token.EOF {
		list = append(list, p.parseDecl())
	}

	return
}

// ----------------------------------------------------------------------------
// Source files

/*
func (p *parser) parseFile() *ast.File {
	if p.trace {
		defer un(trace(p, "File"))
	}

	// package clause
	doc := p.leadComment
	pos := p.expect(token.PACKAGE)
	// Go spec: The package clause is not a declaration;
	// the package name does not appear in any scope.
	ident := p.parseIdent()
	if ident.Name == "_" {
		p.error(p.pos, "invalid package name _")
	}
	p.expectSemi()

	var decls []ast.Decl

	// Don't bother parsing the rest if we had errors already.
	// Likely not a Go source file at all.

	if p.ErrorVector.ErrorCount() == 0 && p.mode&PackageClauseOnly == 0 {
		// import decls
		for p.tok == token.IMPORT {
			decls = append(decls, p.parseGenDecl(token.IMPORT, parseImportSpec))
		}

		if p.mode&ImportsOnly == 0 {
			// rest of package body
			for p.tok != token.EOF {
				decls = append(decls, p.parseDecl())
			}
		}
	}

	assert(p.topScope == p.pkgScope, "imbalanced scopes")

	// resolve global identifiers within the same file
	i := 0
	for _, ident := range p.unresolved {
		// i <= index for current ident
		assert(ident.Obj == unresolved, "object already resolved")
		ident.Obj = p.pkgScope.Lookup(ident.Name) // also removes unresolved sentinel
		if ident.Obj == nil {
			p.unresolved[i] = ident
			i++
		}
	}

	return &ast.File{doc, pos, ident, decls, p.pkgScope, p.imports, p.unresolved[0:i], p.comments}
}
*/
