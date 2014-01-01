// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// This file contains gob pickling and unpickling suppport for ASTs.

package ast

import (
	"encoding/gob"
	"os"
)

/*
   Serializes a relish abstract syntax tree into a file. This file can be thought of as an intermediate code file,
   since it contains the pre-parsed tree.
*/
func Pickle(fileNode *File, pickleFilePath string) (err error) {
	var file *os.File	
	file,err = os.Create(pickleFilePath) 
	defer file.Close()

	encoder := gob.NewEncoder(file) 	
 
    err = encoder.Encode(fileNode) 
   
    return 	
}

/*
   Make sure AST tree node types are registered with gob.
*/
func init() {
	registerAstNodeTypes()
}

/*
   gob type registration.
*/
func registerAstNodeTypes() {
	gob.Register(&RelationDecl{})
	gob.Register(&TypeDecl{})
	gob.Register(&MethodDeclaration{})
	gob.Register(&Closure{})	
	gob.Register(&ConstantDecl{})
	gob.Register(&InputArgDecl{})
	gob.Register(&ReturnArgDecl{})
	gob.Register(&AttributeDecl{})
	gob.Register(&CollectionTypeSpec{})
	gob.Register(&Comment{})
	gob.Register(&CommentGroup{})
	gob.Register(&Field{})
	gob.Register(&FieldList{})
	gob.Register(&BadExpr{})
	gob.Register(&Ident{})
	gob.Register(&Constant{})
	gob.Register(&Ellipsis{})
	gob.Register(&BasicLit{})
	gob.Register(&FuncLit{})
	gob.Register(&CompositeLit{})
	gob.Register(&ParenExpr{})
	gob.Register(&SelectorExpr{})
	gob.Register(&IndexExpr{})
	gob.Register(&SliceExpr{})
	gob.Register(&TypeAssertion{})
	gob.Register(&TypeAssertExpr{})
	gob.Register(&CallExpr{})
	gob.Register(&MethodCall{})
	gob.Register(&ListConstruction{})
	gob.Register(&SetConstruction{})
	gob.Register(&MapConstruction{})		
	gob.Register(&StarExpr{})
	gob.Register(&UnaryExpr{})
	gob.Register(&BinaryExpr{})
	gob.Register(&KeyValueExpr{})
	gob.Register(&ArrayType{})
	gob.Register(&StructType{})
	gob.Register(&FuncType{})
	gob.Register(&InterfaceType{})
	gob.Register(&MapType{})
	gob.Register(&ChanType{})
	gob.Register(&BadStmt{})
	gob.Register(&DeclStmt{})
	gob.Register(&EmptyStmt{})
	gob.Register(&LabeledStmt{})
	gob.Register(&ExprStmt{})
	gob.Register(&SendStmt{})
	gob.Register(&IncDecStmt{})
	gob.Register(&AssignmentStatement{})
	gob.Register(&GoStatement{})
	gob.Register(&DeferStatement{})
	gob.Register(&ReturnStatement{})
	gob.Register(&BreakStatement{})
	gob.Register(&ContinueStatement{})	
	gob.Register(&BranchStmt{})
	gob.Register(&BlockStatement{})
	gob.Register(&IfStatement{})
	gob.Register(&CaseClause{})
	gob.Register(&SwitchStmt{})
	gob.Register(&TypeSwitchStmt{})
	gob.Register(&CommClause{})
	gob.Register(&SelectStmt{})
	gob.Register(&WhileStatement{})
	gob.Register(&ForStatement{})
	gob.Register(&RangeStatement{})
	gob.Register(&ImportSpec{})
	gob.Register(&RelishImportSpec{})
	gob.Register(&ValueSpec{})
	gob.Register(&TypeSpec{})
	gob.Register(&AritySpec{})
	gob.Register(&BadDecl{})
	gob.Register(&GenDecl{})
	gob.Register(&FuncDecl{})
	gob.Register(&File{})
	gob.Register(&Package{})
}

/*
   Returns an ast.File abstract syntax tree, after reading it from a file and deserializing it.
*/
func Unpickle(pickleFilePath string) (fileNode *File, err error) {
	var file *os.File
	file, err = os.Open(pickleFilePath)
	if err != nil {
		return
	}
	defer file.Close()
	
	decoder := gob.NewDecoder(file) 

    err = decoder.Decode(&fileNode)

    return
}
