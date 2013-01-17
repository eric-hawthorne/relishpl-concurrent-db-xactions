// Substantial portions of the source code in this file 
// are Copyright 2009 The Go Authors. All rights reserved.
// Use of such source code is governed by a BSD-style
// license that can be found in the GO_LICENSE file.

// Modifications and additions which convert code to be part of a relish-language compiler 
// are Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of such source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// Package ast declares the types used to represent syntax trees for relish
// packages.
//
package ast

import (
	"relish/compiler/token"
	"unicode"
	"unicode/utf8"
)

// ----------------------------------------------------------------------------
// Interfaces
//
// There are 3 main classes of nodes: Expressions and type nodes,
// statement nodes, and declaration nodes. The node names usually
// match the corresponding Go spec production names to which they
// correspond. The node fields correspond to the individual parts
// of the respective productions.
//
// All nodes contain position information marking the beginning of
// the corresponding source text segment; it is accessible via the
// Pos accessor method. Nodes may contain additional position info
// for language constructs where comments may be found between parts
// of the construct (typically any larger, parenthesized subpart).
// That position information is needed to properly position comments
// when printing the construct.

// All node types implement the Node interface.
type Node interface {
	Pos() token.Pos // position of first character belonging to the node
	End() token.Pos // position of first character immediately after the node
}

// All expression nodes implement the Expr interface.
type Expr interface {
	Node
	exprNode()
}

// All statement nodes implement the Stmt interface.
type Stmt interface {
	Node
	stmtNode()
}

// All declaration nodes implement the Decl interface.
type Decl interface {
	Node
	declNode()
}

// egh
type RelationDecl struct {
	End1 *AttributeDecl
	End2 *AttributeDecl
}

/*
type AttributeDecl struct {
	Name  *Ident     // attribute name	
	Arity *AritySpec // Can be nil - if so, means the type is either not a collection or that the collection is not owned by object
	// that has the attribute.
	Type                                                                                                  *TypeSpec
	PublicReadable, PackageReadable, SubtypeReadable, PublicWriteable, PackageWriteable, SubtypeWriteable bool
	Reassignable, CollectionMutable, Mutable, DeeplyMutable                                               bool
}
*/

type TypeDecl struct {
	Spec *TypeSpec
	// Need some type body stuff here. attributes, getters, setters etc.
	Attributes []*AttributeDecl
}

func (c *RelationDecl) Pos() token.Pos { return 0 }
func (c *RelationDecl) End() token.Pos { return 0 }

func (c *TypeDecl) Pos() token.Pos { return 0 }
func (c *TypeDecl) End() token.Pos { return 0 }

func (c *CollectionTypeSpec) Pos() token.Pos { return c.LDelim }
func (c *CollectionTypeSpec) End() token.Pos { return c.RDelim }

func (c *CollectionTypeSpec) specNode() {}

// EGH A MethodDeclaration node represents a methoddeclaration, or more precisely either an 
// abstract multi-method subroutine declaration or a method declaration.
//
type MethodDeclaration struct {
	Doc          *CommentGroup // associated documentation; or nil
	IsGetter     bool
	IsSetter     bool
	Name         *Ident          // function/method name
	Type         *FuncType       // parameters and results, position of function declaration
	Body         *BlockStatement // function body; or nil (abstract subroutine interface declaration)
	NumLocalVars int
	NumFreeVars int  // may be > 0 if this is a closure method
	IsClosureMethod bool
}



// EGH A Closure node represents a closure expression which when evaluated at runtime will return
// an RClosure object with a reference to an RMethod and a bindings map.
type Closure struct {
	FuncPos token.Pos   // position of the func keyword
	MethodName string
	Bindings []int  // a list of varable/param stack-offsets
	                // in the enclosing method's stack frame
	                // These are the enclosing vars whose values must be captured
	                // each time the closure is encountered in code execution.
}

func (c *Closure) exprNode() {}

func (c *Closure) Pos() token.Pos { return c.FuncPos }
func (c *Closure) End() token.Pos { return c.FuncPos + 4 }


func (d *MethodDeclaration) ReturnArgsAreNamed() bool {
   rslts := d.Type.Results
   return len(rslts) > 0 && rslts[0].Name != nil
}

func (d *MethodDeclaration) NumAnonymousReturnVals() int {
   rslts := d.Type.Results
   n := len(rslts)
   if n == 0 {
      return 0	
   } else if rslts[0].Name != nil {
      return 0	
   }
   return n
}

func (d *MethodDeclaration) NumReturnVals() int {
   rslts := d.Type.Results
   return len(rslts)
}

type ConstantDecl struct {
	Name  *Ident // constant name	
	Value Expr
}

// EGH?
type InputArgDecl struct {
	Name *Ident // input parameter name	
	Type *TypeSpec
	Default Expr
    IsVariadic bool	
}

// EGH?
type ReturnArgDecl struct {
	Name *Ident // output argument name - may be nil	
	Type *TypeSpec
}

// EGH?
type AttributeDecl struct {
	Name  *Ident     // attribute name	
	Arity *AritySpec // Can be nil - if so, means the type is either not a collection or that the collection is not owned by object
	// that has the attribute.
	Type                                                                                                  *TypeSpec
	PublicReadable, PackageReadable, SubtypeReadable, PublicWriteable, PackageWriteable, SubtypeWriteable bool
	Reassignable, CollectionMutable, Mutable, DeeplyMutable                                               bool
}

func (a *AttributeDecl) IsReassignable() bool {
	return a.Reassignable
}

/*
If this is a multivalued attribute, it means the collection is immutable.
Or is this for a collection-valued attribute.
*/
func (a *AttributeDecl) IsCollectionMutable() bool {
	return a.CollectionMutable
}

/*
Is the attibute-value object itself mutable (or the objects in the collection if this is multivalued attribute.)
*/
func (a *AttributeDecl) IsMutable() bool {
	return a.CollectionMutable
}

/*
Whether the tree of objects referred to by the attribute-value object is mutable.
*/
func (a *AttributeDecl) IsDeeplyMutable() bool {
	return a.DeeplyMutable
}

func (a *AttributeDecl) IsPublicReadable() bool {
	return a.PublicReadable
}

func (a *AttributeDecl) IsPackageReadable() bool {
	return a.PackageReadable
}

func (a *AttributeDecl) IsSubtypeReadable() bool {
	return a.SubtypeReadable
}

func (a *AttributeDecl) IsPublicWriteable() bool {
	return a.PublicWriteable
}

func (a *AttributeDecl) IsPackageWriteable() bool {
	return a.PackageWriteable
}

func (a *AttributeDecl) IsSubtypeWriteable() bool {
	return a.SubtypeWriteable
}

type CollectionTypeSpec struct {
	Kind        token.Token
	LDelim      token.Pos
	RDelim      token.Pos
	IsSorting   bool
	IsAscending bool
	OrderFunc   string
}

// ----------------------------------------------------------------------------
// Comments

// A Comment node represents a single //-style or /*-style comment.
type Comment struct {
	Slash token.Pos // position of "/" starting the comment
	Text  string    // comment text (excluding '\n' for //-style comments)
}

func (c *Comment) Pos() token.Pos { return c.Slash }
func (c *Comment) End() token.Pos { return token.Pos(int(c.Slash) + len(c.Text)) }

// A CommentGroup represents a sequence of comments
// with no other tokens and no empty lines between.
//
type CommentGroup struct {
	List []*Comment // len(List) > 0
}

func (g *CommentGroup) Pos() token.Pos { return g.List[0].Pos() }
func (g *CommentGroup) End() token.Pos { return g.List[len(g.List)-1].End() }

// ----------------------------------------------------------------------------
// Expressions and types

// A Field represents a Field declaration list in a struct type,
// a method list in an interface type, or a parameter/result declaration
// in a signature.
//
type Field struct {
	Doc     *CommentGroup // associated documentation; or nil
	Names   []*Ident      // field/method/parameter names; or nil if anonymous field
	Type    Expr          // field/method/parameter type
	Tag     *BasicLit     // field tag; or nil
	Comment *CommentGroup // line comments; or nil
}

func (f *Field) Pos() token.Pos {
	if len(f.Names) > 0 {
		return f.Names[0].Pos()
	}
	return f.Type.Pos()
}

func (f *Field) End() token.Pos {
	if f.Tag != nil {
		return f.Tag.End()
	}
	return f.Type.End()
}

// A FieldList represents a list of Fields, enclosed by parentheses or braces.
type FieldList struct {
	Opening token.Pos // position of opening parenthesis/brace, if any
	List    []*Field  // field list; or nil
	Closing token.Pos // position of closing parenthesis/brace, if any
}

func (f *FieldList) Pos() token.Pos {
	if f.Opening.IsValid() {
		return f.Opening
	}
	// the list should not be empty in this case;
	// be conservative and guard against bad ASTs
	if len(f.List) > 0 {
		return f.List[0].Pos()
	}
	return token.NoPos
}

func (f *FieldList) End() token.Pos {
	if f.Closing.IsValid() {
		return f.Closing + 1
	}
	// the list should not be empty in this case;
	// be conservative and guard against bad ASTs
	if n := len(f.List); n > 0 {
		return f.List[n-1].End()
	}
	return token.NoPos
}

// NumFields returns the number of (named and anonymous fields) in a FieldList.
func (f *FieldList) NumFields() int {
	n := 0
	if f != nil {
		for _, g := range f.List {
			m := len(g.Names)
			if m == 0 {
				m = 1 // anonymous field
			}
			n += m
		}
	}
	return n
}

// An expression is represented by a tree consisting of one
// or more of the following concrete expression nodes.
//
type (
	// A BadExpr node is a placeholder for expressions containing
	// syntax errors for which no correct expression nodes can be
	// created.
	//
	BadExpr struct {
		From, To token.Pos // position range of bad expression
	}

	// An Ident node represents an identifier.
	Ident struct {
		NamePos token.Pos   // identifier position
		Name    string      // identifier name
		Obj     *Object     // denoted object; or nil
		Kind    token.Token // CONST,VAR,TYPE,PACKAGE,FUNC,CLOSURE,TYPEVAR	
		Offset  int         // if a local var or parameter, the stack offset in the current frame
	}

	// egh 0 A Constant node represents a Constant identifier.
	Constant struct {
		NamePos token.Pos // identifier position
		Name    string    // identifier name
		Obj     *Object   // denoted object; or nil
	}

	// An Ellipsis node stands for the "..." type in a
	// parameter list or the "..." length in an array type.
	//
	Ellipsis struct {
		Ellipsis token.Pos // position of "..."
		Elt      Expr      // ellipsis element type (parameter lists only); or nil
	}

	// A BasicLit node represents a literal of basic type.
	BasicLit struct {
		ValuePos token.Pos   // literal position
		Kind     token.Token // token.INT, token.FLOAT, token.IMAG, token.CHAR, or token.STRING
		Value    string      // literal string; e.g. 42, 0x7f, 3.14, 1e-9, 2.4i, 'a', '\x7f', "foo" or `\m\n\o`
	}

	// A FuncLit node represents a function literal.
	FuncLit struct {
		Type *FuncType  // function type
		Body *BlockStmt // function body
	}

	// A CompositeLit node represents a composite literal.
	CompositeLit struct {
		Type   Expr      // literal type; or nil
		Lbrace token.Pos // position of "{"
		Elts   []Expr    // list of composite elements; or nil
		Rbrace token.Pos // position of "}"
	}

	// A ParenExpr node represents a parenthesized expression.
	ParenExpr struct {
		Lparen token.Pos // position of "("
		X      Expr      // parenthesized expression
		Rparen token.Pos // position of ")"
	}

	// A SelectorExpr node represents an expression followed by a selector.
	SelectorExpr struct {
		X   Expr   // expression
		Sel *Ident // field selector
	}

	// An IndexExpr node represents an expression followed by [ index ].
	//
	// This variant, [?index], for maps, returns whether the value is in the map
	// This variant, [index?], for maps, returns the value or a zero value, and whether the value is in the map
	//
	IndexExpr struct {
		X      Expr      // expression
		Lbrack token.Pos // position of "["
		Index  Expr      // index expression
		Rbrack token.Pos // position of "]"
		AssertExists bool  // the expression explicitly asserts that the value or a zero value is to be returned even if not found in map
		AskWhether bool // the expression is supposed to return just whether the key is in the map
	}

	// An SliceExpr node represents an expression followed by slice indices.
	SliceExpr struct {
		X      Expr      // expression
		Lbrack token.Pos // position of "["
		Low    Expr      // begin of slice range; or nil
		High   Expr      // end of slice range; or nil
		Rbrack token.Pos // position of "]"
	}

	// egh A TypeAssertion node represents an expression preceded by a
	// type assertion.
	//
	TypeAssertion struct {
		Type TypeSpec // asserted type		
		X    Expr     // expression
	}

	TypeAssertExpr struct {
		X    Expr // expression
		Type Expr // asserted type; nil means type switch X.(type)
	}

	// A CallExpr node represents an expression followed by an argument list.
	CallExpr struct {
		Fun      Expr      // function expression
		Lparen   token.Pos // position of "("
		Args     []Expr    // function arguments; or nil
		Ellipsis token.Pos // position of "...", if any
		Rparen   token.Pos // position of ")"
	}

	// EGH A MethodCall node represents an expression followed by an argument list, optionally including keyword-form args.
	// Keyword-form args, if present, must be found in the call expression after all required positional args and before variadic args if any.
	MethodCall struct {
		Fun      Expr      // function expression
		Args     []Expr    // function arguments; or nil
		Ellipsis token.Pos // position of "...", if any
		KeywordArgs map[string]Expr
		NumPositionalArgs int32 // -1 if not known. Will be known if there are keyword args
	}

	// EGH A ListConstruction node represents a list constructor invocation, which may be a list literal, a new empty list of a type, or
	// a list with a db sql query where clause specified as the source of list members.
	ListConstruction struct {
        Type *TypeSpec     // Includes the CollectionTypeSpec which must be a spec of a List.
		Elements  []Expr    // explicitly listed elements; or nil  
		Generator *RangeStatement // A for-range generator which will yield elements for the list, or nil      
		Query     Expr     // must be an expression evaluating to a String containing a SQL WHERE clause (without the "WHERE"), or nil
		                   // Note eventually it should be more like OQL where you can say e.g. engine.horsePower > 120 when fetching []Car
	}	
	
	// EGH A SetConstruction node represents a set constructor invocation, which may be a set literal, a new empty set of a type, or
	// a set with a db sql query where clause specified as the source of set members.
	SetConstruction struct {
        Type *TypeSpec     // Includes the CollectionTypeSpec which must be a spec of a Set.
		Elements  []Expr    // explicitly listed elements; or nil        
		Generator *RangeStatement // A for-range generator which will yield elements for the set, or nil		
		Query     Expr     // must be an expression evaluating to a String containing a SQL WHERE clause (without the "WHERE"), or nil
		                   // Note eventually it should be more like OQL where you can say e.g. engine.horsePower > 120 when fetching []Car
	}
	
	// EGH A MapConstruction node represents a Map constructor invocation, which may be a map literal, 
	// or a new empty map of some type to another.
	MapConstruction struct {
        Type *TypeSpec     // Includes the CollectionTypeSpec which must be a spec of a Map.
        ValType *TypeSpec     // Type of the values
        Keys []Expr         // explicitly listed keys; or nil
		Elements  []Expr    // explicitly listed elements; or nil  
		Generator *RangeStatement // A for-range generator which will yield key/map pairs for the map, or nil			      
	}		


	// A StarExpr node represents an expression of the form "*" Expression.
	// Semantically it could be a unary "*" expression, or a pointer type.
	//
	StarExpr struct {
		Star token.Pos // position of "*"
		X    Expr      // operand
	}

	// A UnaryExpr node represents a unary expression.
	// Unary "*" expressions are represented via StarExpr nodes.
	//
	UnaryExpr struct {
		OpPos token.Pos   // position of Op
		Op    token.Token // operator
		X     Expr        // operand
	}

	// A BinaryExpr node represents a binary expression.
	BinaryExpr struct {
		X     Expr        // left operand
		OpPos token.Pos   // position of Op
		Op    token.Token // operator
		Y     Expr        // right operand
	}

	// A KeyValueExpr node represents (key : value) pairs
	// in composite literals.
	//
	KeyValueExpr struct {
		Key   Expr
		Colon token.Pos // position of ":"
		Value Expr
	}
)

// The direction of a channel type is indicated by one
// of the following constants.
//
type ChanDir int

const (
	SEND ChanDir = 1 << iota
	RECV
)

// A type is represented by a tree consisting of one
// or more of the following type-specific expression
// nodes.
//
type (
	// An ArrayType node represents an array or slice type.
	ArrayType struct {
		Lbrack token.Pos // position of "["
		Len    Expr      // Ellipsis node for [...]T array types, nil for slice types
		Elt    Expr      // element type
	}

	// A StructType node represents a struct type.
	StructType struct {
		Struct     token.Pos  // position of "struct" keyword
		Fields     *FieldList // list of field declarations
		Incomplete bool       // true if (source) fields are missing in the Fields list
	}

	// Pointer types are represented via StarExpr nodes.

	// A FuncType node represents a function type.
	FuncType struct {
		Func      token.Pos        // position of subroutine name or "lambda" keyword
		AfterName token.Pos        // position after end of function name or "lambda"
		Params    []*InputArgDecl  // input parameter declarations. Can be empty list.
		Results   []*ReturnArgDecl // (outgoing) result declarations; Can be empty list.
	}

	// An InterfaceType node represents an interface type.
	InterfaceType struct {
		Interface  token.Pos  // position of "interface" keyword
		Methods    *FieldList // list of methods
		Incomplete bool       // true if (source) methods are missing in the Methods list
	}

	// A MapType node represents a map type.
	MapType struct {
		Map   token.Pos // position of "map" keyword
		Key   Expr
		Value Expr
	}

	// A ChanType node represents a channel type.
	ChanType struct {
		Begin token.Pos // position of "chan" keyword or "<-" (whichever comes first)
		Dir   ChanDir   // channel direction
		Value Expr      // value type
	}
)

// Pos and End implementations for expression/type nodes.
//
func (x *BadExpr) Pos() token.Pos  { return x.From }
func (x *Ident) Pos() token.Pos    { return x.NamePos }
func (x *Ellipsis) Pos() token.Pos { return x.Ellipsis }
func (x *BasicLit) Pos() token.Pos { return x.ValuePos }
func (x *FuncLit) Pos() token.Pos  { return x.Type.Pos() }
func (x *CompositeLit) Pos() token.Pos {
	if x.Type != nil {
		return x.Type.Pos()
	}
	return x.Lbrace
}
func (x *ParenExpr) Pos() token.Pos      { return x.Lparen }
func (x *SelectorExpr) Pos() token.Pos   { return x.X.Pos() }
func (x *IndexExpr) Pos() token.Pos      { return x.X.Pos() }
func (x *SliceExpr) Pos() token.Pos      { return x.X.Pos() }
func (x *TypeAssertExpr) Pos() token.Pos { return x.X.Pos() }
func (x *TypeAssertion) Pos() token.Pos  { return x.Type.Pos() }
func (x *MethodCall) Pos() token.Pos     { return x.Fun.Pos() }
func (x *ListConstruction) Pos() token.Pos     { return x.Type.Pos() }
func (x *SetConstruction) Pos() token.Pos     { return x.Type.Pos() }
func (x *MapConstruction) Pos() token.Pos     { return x.Type.Pos() }
func (x *CallExpr) Pos() token.Pos       { return x.Fun.Pos() }
func (x *StarExpr) Pos() token.Pos       { return x.Star }
func (x *UnaryExpr) Pos() token.Pos      { return x.OpPos }
func (x *BinaryExpr) Pos() token.Pos     { return x.X.Pos() }
func (x *KeyValueExpr) Pos() token.Pos   { return x.Key.Pos() }
func (x *ArrayType) Pos() token.Pos      { return x.Lbrack }
func (x *StructType) Pos() token.Pos     { return x.Struct }
func (x *FuncType) Pos() token.Pos       { return x.Func }
func (x *InterfaceType) Pos() token.Pos  { return x.Interface }
func (x *MapType) Pos() token.Pos        { return x.Map }
func (x *ChanType) Pos() token.Pos       { return x.Begin }

func (x *BadExpr) End() token.Pos { return x.To }
func (x *Ident) End() token.Pos   { return token.Pos(int(x.NamePos) + len(x.Name)) }
func (x *Ellipsis) End() token.Pos {
	if x.Elt != nil {
		return x.Elt.End()
	}
	return x.Ellipsis + 3 // len("...")
}
func (x *BasicLit) End() token.Pos     { return token.Pos(int(x.ValuePos) + len(x.Value)) }
func (x *FuncLit) End() token.Pos      { return x.Body.End() }
func (x *CompositeLit) End() token.Pos { return x.Rbrace + 1 }
func (x *ParenExpr) End() token.Pos    { return x.Rparen + 1 }
func (x *SelectorExpr) End() token.Pos { return x.Sel.End() }
func (x *IndexExpr) End() token.Pos    { return x.Rbrack + 1 }
func (x *SliceExpr) End() token.Pos    { return x.Rbrack + 1 }
func (x *TypeAssertExpr) End() token.Pos {
	if x.Type != nil {
		return x.Type.End()
	}
	return x.X.End()
}

func (x *TypeAssertion) End() token.Pos {
	return x.X.End() + 1
}

func (x *CallExpr) End() token.Pos { return x.Rparen + 1 }
func (x *MethodCall) End() token.Pos {
	if n := len(x.Args); n > 0 {
		return x.Args[n-1].End()
	}
	return x.Fun.End()
}
func (x *ListConstruction) End() token.Pos     { 
	if  x.Query != nil {
		return x.Query.End() 
	}
	return x.Type.End()
}

func (x *SetConstruction) End() token.Pos     { 
	if  x.Query != nil {
		return x.Query.End() 
	}
	return x.Type.End()
}

func (x *MapConstruction) End() token.Pos     { 
	return x.ValType.End()
}

func (x *StarExpr) End() token.Pos     { return x.X.End() }
func (x *UnaryExpr) End() token.Pos    { return x.X.End() }
func (x *BinaryExpr) End() token.Pos   { return x.Y.End() }
func (x *KeyValueExpr) End() token.Pos { return x.Value.End() }
func (x *ArrayType) End() token.Pos    { return x.Elt.End() }
func (x *StructType) End() token.Pos   { return x.Fields.End() }
func (x *FuncType) End() token.Pos {
	if x.Results != nil {
		return x.Results[len(x.Results)-1].End()
	}
	if x.Params != nil {
		return x.Params[len(x.Params)-1].End()
	}
	return x.AfterName
}
func (x *InterfaceType) End() token.Pos { return x.Methods.End() }
func (x *MapType) End() token.Pos       { return x.Value.End() }
func (x *ChanType) End() token.Pos      { return x.Value.End() }

// exprNode() ensures that only expression/type nodes can be
// assigned to an ExprNode.
//
func (x *BadExpr) exprNode()        {}
func (x *Ident) exprNode()          {}
func (x *Ellipsis) exprNode()       {}
func (x *BasicLit) exprNode()       {}
func (x *FuncLit) exprNode()        {}
func (x *CompositeLit) exprNode()   {}
func (x *ParenExpr) exprNode()      {}
func (x *SelectorExpr) exprNode()   {}
func (x *IndexExpr) exprNode()      {}
func (x *SliceExpr) exprNode()      {}
func (x *TypeAssertExpr) exprNode() {}
func (x *TypeAssertion) exprNode()  {}
func (x *CallExpr) exprNode()       {}
func (x *MethodCall) exprNode()     {}
func (x *ListConstruction) exprNode()     {}
func (x *SetConstruction) exprNode()     {}
func (x *MapConstruction) exprNode()     {}
func (x *StarExpr) exprNode()       {}
func (x *UnaryExpr) exprNode()      {}
func (x *BinaryExpr) exprNode()     {}
func (x *KeyValueExpr) exprNode()   {}

func (x *ArrayType) exprNode()     {}
func (x *StructType) exprNode()    {}
func (x *FuncType) exprNode()      {}
func (x *InterfaceType) exprNode() {}
func (x *MapType) exprNode()       {}
func (x *ChanType) exprNode()      {}

// ----------------------------------------------------------------------------
// Convenience functions for Idents

var noPos token.Pos

// NewIdent creates a new Ident without position.
// Useful for ASTs generated by code other than the Go parser.
//
func NewIdent(name string, kind token.Token) *Ident { return &Ident{noPos, name, nil, kind, -1} }

// IsExported returns whether name is an exported Go symbol
// (i.e., whether it begins with an uppercase letter).
//
func IsExported(name string) bool {
	ch, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(ch)
}

// IsExported returns whether id is an exported Go symbol
// (i.e., whether it begins with an uppercase letter).
//
func (id *Ident) IsExported() bool { return IsExported(id.Name) }

func (id *Ident) String() string {
	if id != nil {
		return id.Name
	}
	return "<nil>"
}

// ----------------------------------------------------------------------------
// Statements

// A statement is represented by a tree consisting of one
// or more of the following concrete statement nodes.
//
type (
	// A BadStmt node is a placeholder for statements containing
	// syntax errors for which no correct statement nodes can be
	// created.
	//
	BadStmt struct {
		From, To token.Pos // position range of bad statement
	}

	// A DeclStmt node represents a declaration in a statement list.
	DeclStmt struct {
		Decl Decl
	}

	// An EmptyStmt node represents an empty statement.
	// The "position" of the empty statement is the position
	// of the immediately preceding semicolon.
	//
	EmptyStmt struct {
		Semicolon token.Pos // position of preceding ";"
	}

	// A LabeledStmt node represents a labeled statement.
	LabeledStmt struct {
		Label *Ident
		Colon token.Pos // position of ":"
		Stmt  Stmt
	}

	// An ExprStmt node represents a (stand-alone) expression
	// in a statement list.
	//
	ExprStmt struct {
		X Expr // expression
	}

	// A SendStmt node represents a send statement.
	SendStmt struct {
		Chan  Expr
		Arrow token.Pos // position of "<-"
		Value Expr
	}

	// An IncDecStmt node represents an increment or decrement statement.
	IncDecStmt struct {
		X      Expr
		TokPos token.Pos   // position of Tok
		Tok    token.Token // INC or DEC
	}

	// EGH An AssignmentStatement node represents an assignment 
	//
	AssignmentStatement struct {
		Lhs    []Expr
		TokPos token.Pos   // position of Tok
		Tok    token.Token // assignment token, DEFINE
		Rhs    []Expr
	}

	// An AssignStmt node represents an assignment or
	// a short variable declaration.
	//
	AssignStmt struct {
		Lhs    []Expr
		TokPos token.Pos   // position of Tok
		Tok    token.Token // assignment token, DEFINE
		Rhs    []Expr
	}

	// EGH A GoStatement node represents a go statement.
	GoStatement struct {
		Go   token.Pos // position of "go" keyword
		Call *MethodCall
	}
	
	// A GoStmt node represents a go statement.
	GoStmt struct {
		Go   token.Pos // position of "go" keyword
		Call *CallExpr
	}	

	// EGH A DeferStatement node represents a defer statement.
	DeferStatement struct {
		Defer token.Pos // position of "defer" keyword
		Call  *MethodCall
	}
	
	// A DeferStmt node represents a defer statement.
	DeferStmt struct {
		Defer token.Pos // position of "defer" keyword
		Call  *CallExpr
	}
	
	

	// EGH A ReturnStatement node represents a return statement.
	ReturnStatement struct {
		Return  token.Pos // position of "=>" keyword
		Results []Expr    // result expressions; or nil
		IsYield bool      // If true, this is a generator "yield" statement 
	}

	// EGH A BreakStatement node represents a break statement.
	BreakStatement struct {
		Break  token.Pos // position of "break" keyword
	}

	// EGH A ContinueStatement node represents a continue statement.
	ContinueStatement struct {
		Continue  token.Pos // position of "continue" keyword
	}	


	// A ReturnStmt node represents a return statement.
	ReturnStmt struct {
		Return  token.Pos // position of "return" keyword
		Results []Expr    // result expressions; or nil
	}

	// A BranchStmt node represents a break, continue, goto,
	// or fallthrough statement.
	//
	BranchStmt struct {
		TokPos token.Pos   // position of Tok
		Tok    token.Token // keyword token (BREAK, CONTINUE, GOTO, FALLTHROUGH)
		Label  *Ident      // label name; or nil
	}

	// A BlockStmt node represents a braced statement list.
	BlockStmt struct {
		Lbrace token.Pos // position of "{"
		List   []Stmt
		Rbrace token.Pos // position of "}"
	}

	// EGH A BlockStatement node represents a statement list.
	BlockStatement struct {
		Start token.Pos // position of beginning of first statement
		List  []Stmt
	}

	// EGH An IfStatement node represents an if statement.
	IfStatement struct {
		If   token.Pos // position of "if" keyword
		Cond Expr      // condition
		Body *BlockStatement
		Else Stmt // else branch; or nil
	}

	// An IfStmt node represents an if statement.
	IfStmt struct {
		If   token.Pos // position of "if" keyword
		Init Stmt      // initialization statement; or nil
		Cond Expr      // condition
		Body *BlockStmt
		Else Stmt // else branch; or nil
	}

	// A CaseClause represents a case of an expression or type switch statement.
	CaseClause struct {
		Case  token.Pos // position of "case" or "default" keyword
		List  []Expr    // list of expressions or types; nil means default case
		Colon token.Pos // position of ":"
		Body  []Stmt    // statement list; or nil
	}

	// A SwitchStmt node represents an expression switch statement.
	SwitchStmt struct {
		Switch token.Pos  // position of "switch" keyword
		Init   Stmt       // initialization statement; or nil
		Tag    Expr       // tag expression; or nil
		Body   *BlockStmt // CaseClauses only
	}

	// An TypeSwitchStmt node represents a type switch statement.
	TypeSwitchStmt struct {
		Switch token.Pos  // position of "switch" keyword
		Init   Stmt       // initialization statement; or nil
		Assign Stmt       // x := y.(type) or y.(type)
		Body   *BlockStmt // CaseClauses only
	}

	// A CommClause node represents a case of a select statement.
	CommClause struct {
		Case  token.Pos // position of "case" or "default" keyword
		Comm  Stmt      // send or receive statement; nil means default case
		Colon token.Pos // position of ":"
		Body  []Stmt    // statement list; or nil
	}

	// An SelectStmt node represents a select statement.
	SelectStmt struct {
		Select token.Pos  // position of "select" keyword
		Body   *BlockStmt // CommClauses only
	}

	// EGH A WhileStatement represents a while statement.
	WhileStatement struct {
		While token.Pos // position of "while" keyword		
		Cond  Expr      // condition
		Body  *BlockStatement
		Else  Stmt // else branch; or nil		
	}

	// EGH A ForStatement represents a for statement.
	ForStatement struct {
		For  token.Pos // position of "for" keyword
		Init Stmt      // initialization statement; or nil
		Cond Expr      // condition; or nil
		Post []Stmt    // post iteration statement; or nil - number of Post statements must equal number of Init variables
		Body *BlockStatement
	}

	// EGH A RangeStatement represents a for statement which ranges over one or more collections.
	// Number n of value exprs and X exprs should match, or else there must be one X expr returning n values
	RangeStatement struct {
		For          token.Pos // position of "for" keyword
		KeyAndValues []Expr    // One or more of Key or the Values may be nil - minimum is one value or one key
		// This is not correct!!! we need to handle multiple values
		X    []Expr // value to range over NOT CORRECT!!! Need to handle multiple expressions.
		Body *BlockStatement
	}
	

	// A ForStmt represents a for statement.
	ForStmt struct {
		For  token.Pos // position of "for" keyword
		Init Stmt      // initialization statement; or nil
		Cond Expr      // condition; or nil
		Post Stmt      // post iteration statement; or nil
		Body *BlockStmt
	}

	// A RangeStmt represents a for statement with a range clause.
	RangeStmt struct {
		For        token.Pos   // position of "for" keyword
		Key, Value Expr        // Value may be nil
		TokPos     token.Pos   // position of Tok
		Tok        token.Token // ASSIGN, DEFINE
		X          Expr        // value to range over
		Body       *BlockStmt
	}
)

// Pos and End implementations for statement nodes.
//
func (s *BadStmt) Pos() token.Pos             { return s.From }
func (s *DeclStmt) Pos() token.Pos            { return s.Decl.Pos() }
func (s *EmptyStmt) Pos() token.Pos           { return s.Semicolon }
func (s *LabeledStmt) Pos() token.Pos         { return s.Label.Pos() }
func (s *ExprStmt) Pos() token.Pos            { return s.X.Pos() }
func (s *SendStmt) Pos() token.Pos            { return s.Chan.Pos() }
func (s *IncDecStmt) Pos() token.Pos          { return s.X.Pos() }
func (s *AssignStmt) Pos() token.Pos          { return s.Lhs[0].Pos() }
func (s *AssignmentStatement) Pos() token.Pos { return s.Lhs[0].Pos() }
func (s *GoStatement) Pos() token.Pos              { return s.Go }
func (s *DeferStatement) Pos() token.Pos           { return s.Defer }
func (s *GoStmt) Pos() token.Pos              { return s.Go }
func (s *DeferStmt) Pos() token.Pos           { return s.Defer }
func (s *ReturnStmt) Pos() token.Pos          { return s.Return }
func (s *ReturnStatement) Pos() token.Pos     { return s.Return }
func (s *BreakStatement) Pos() token.Pos      { return s.Break }
func (s *ContinueStatement) Pos() token.Pos   { return s.Continue }
func (s *BranchStmt) Pos() token.Pos          { return s.TokPos }
func (s *BlockStmt) Pos() token.Pos           { return s.Lbrace }
func (s *BlockStatement) Pos() token.Pos      { return s.Start }
func (s *IfStmt) Pos() token.Pos              { return s.If }
func (s *IfStatement) Pos() token.Pos         { return s.If }
func (s *CaseClause) Pos() token.Pos          { return s.Case }
func (s *SwitchStmt) Pos() token.Pos          { return s.Switch }
func (s *TypeSwitchStmt) Pos() token.Pos      { return s.Switch }
func (s *CommClause) Pos() token.Pos          { return s.Case }
func (s *SelectStmt) Pos() token.Pos          { return s.Select }
func (s *ForStmt) Pos() token.Pos             { return s.For }
func (s *ForStatement) Pos() token.Pos        { return s.For }
func (s *RangeStatement) Pos() token.Pos      { return s.For }
func (s *WhileStatement) Pos() token.Pos      { return s.While }
func (s *RangeStmt) Pos() token.Pos           { return s.For }

func (s *BadStmt) End() token.Pos  { return s.To }
func (s *DeclStmt) End() token.Pos { return s.Decl.End() }
func (s *EmptyStmt) End() token.Pos {
	return s.Semicolon + 1 /* len(";") */
}
func (s *LabeledStmt) End() token.Pos { return s.Stmt.End() }
func (s *ExprStmt) End() token.Pos    { return s.X.End() }
func (s *SendStmt) End() token.Pos    { return s.Value.End() }
func (s *IncDecStmt) End() token.Pos {
	return s.TokPos + 2 /* len("++") */
}
func (s *AssignStmt) End() token.Pos          { return s.Rhs[len(s.Rhs)-1].End() }
func (s *AssignmentStatement) End() token.Pos { return s.Rhs[len(s.Rhs)-1].End() }
func (s *GoStmt) End() token.Pos              { return s.Call.End() }
func (s *DeferStmt) End() token.Pos           { return s.Call.End() }
func (s *GoStatement) End() token.Pos              { return s.Call.End() }
func (s *DeferStatement) End() token.Pos           { return s.Call.End() }
func (s *ReturnStmt) End() token.Pos {
	if n := len(s.Results); n > 0 {
		return s.Results[n-1].End()
	}
	return s.Return + 6 // len("return")
}

func (s *ReturnStatement) End() token.Pos {
	if n := len(s.Results); n > 0 {
		return s.Results[n-1].End()
	}
	return s.Return + 2 // len("=>")
}

func (s *BreakStatement) End() token.Pos {
	return s.Break + 5 
}

func (s *ContinueStatement) End() token.Pos {
	return s.Continue + 8
}

func (s *BranchStmt) End() token.Pos {
	if s.Label != nil {
		return s.Label.End()
	}
	return token.Pos(int(s.TokPos) + len(s.Tok.String()))
}
func (s *BlockStmt) End() token.Pos { return s.Rbrace + 1 }

func (s *BlockStatement) End() token.Pos { return s.List[len(s.List)-1].End() }

func (s *IfStatement) End() token.Pos {
	if s.Else != nil {
		return s.Else.End()
	}
	return s.Body.End()
}

func (s *IfStmt) End() token.Pos {
	if s.Else != nil {
		return s.Else.End()
	}
	return s.Body.End()
}
func (s *CaseClause) End() token.Pos {
	if n := len(s.Body); n > 0 {
		return s.Body[n-1].End()
	}
	return s.Colon + 1
}
func (s *SwitchStmt) End() token.Pos     { return s.Body.End() }
func (s *TypeSwitchStmt) End() token.Pos { return s.Body.End() }
func (s *CommClause) End() token.Pos {
	if n := len(s.Body); n > 0 {
		return s.Body[n-1].End()
	}
	return s.Colon + 1
}
func (s *SelectStmt) End() token.Pos     { return s.Body.End() }
func (s *ForStmt) End() token.Pos        { return s.Body.End() }
func (s *WhileStatement) End() token.Pos { return s.Body.End() }
func (s *RangeStmt) End() token.Pos      { return s.Body.End() }
func (s *ForStatement) End() token.Pos   { return s.Body.End() }
func (s *RangeStatement) End() token.Pos { return s.Body.End() }

// stmtNode() ensures that only statement nodes can be
// assigned to a StmtNode.
//
func (s *BadStmt) stmtNode()             {}
func (s *DeclStmt) stmtNode()            {}
func (s *EmptyStmt) stmtNode()           {}
func (s *LabeledStmt) stmtNode()         {}
func (s *ExprStmt) stmtNode()            {}
func (s *SendStmt) stmtNode()            {}
func (s *IncDecStmt) stmtNode()          {}
func (s *AssignStmt) stmtNode()          {}
func (s *AssignmentStatement) stmtNode() {}
func (s *GoStmt) stmtNode()              {}
func (s *DeferStmt) stmtNode()           {}
func (s *GoStatement) stmtNode()         {}
func (s *DeferStatement) stmtNode()      {}
func (s *ReturnStmt) stmtNode()          {}
func (s *ReturnStatement) stmtNode()     {}
func (s *BreakStatement) stmtNode()      {}
func (s *ContinueStatement) stmtNode()   {}
func (s *BranchStmt) stmtNode()          {}
func (s *BlockStmt) stmtNode()           {}
func (s *IfStmt) stmtNode()              {}
func (s *BlockStatement) stmtNode()      {}
func (s *IfStatement) stmtNode()         {}
func (s *CaseClause) stmtNode()          {}
func (s *SwitchStmt) stmtNode()          {}
func (s *TypeSwitchStmt) stmtNode()      {}
func (s *CommClause) stmtNode()          {}
func (s *SelectStmt) stmtNode()          {}
func (s *ForStmt) stmtNode()             {}
func (s *ForStatement) stmtNode()        {}
func (s *RangeStatement) stmtNode()      {}
func (s *WhileStatement) stmtNode()      {}
func (x *MethodCall) stmtNode()          {}
func (s *RangeStmt) stmtNode()           {}

// ----------------------------------------------------------------------------
// Declarations

// A Spec node represents a single (non-parenthesized) import,
// constant, type, or variable declaration.
//
type (
	// The Spec type stands for any of *ImportSpec, *RelishImportSpec, *ValueSpec, and *TypeSpec.
	Spec interface {
		Node
		specNode()
	}

	// An ImportSpec node represents a single package import.
	RelishImportSpec struct {		
		Alias string   // Local name of the imported package - either last part of package path, or an alias
		OriginAndArtifactName string
		PackageName string
	}

	// An ImportSpec node represents a single package import.
	ImportSpec struct {
		Doc     *CommentGroup // associated documentation; or nil
		Name    *Ident        // local package name (including "."); or nil
		Path    *BasicLit     // import path
		Comment *CommentGroup // line comments; or nil
	}

	// A ValueSpec node represents a constant or variable declaration
	// (ConstSpec or VarSpec production).
	//
	ValueSpec struct {
		Doc     *CommentGroup // associated documentation; or nil
		Names   []*Ident      // value names (len(Names) > 0)
		Type    Expr          // value type; or nil
		Values  []Expr        // initial values; or nil
		Comment *CommentGroup // line comments; or nil
	}

	// A TypeSpec node represents a type declaration (TypeSpec production).
	// egh Can now refer to a type variable too.
	// e.g.
	// Int
	// Bag of Int
	// Bag of T
	// Bag of
	//    T <: Numeric
	// T1
	// T2 <: T
	// T3 <: Ordered
	// 
	TypeSpec struct {
		Doc            *CommentGroup       // associated documentation; or nil
		Name           *Ident              // type name (or type variable name)
		Type           Expr                // *Ident, *ParenExpr, *SelectorExpr, *StarExpr, or any of the *XxxTypes
		Comment        *CommentGroup       // line comments; or nil
		Params         []*TypeSpec         // Type parameters (egh)
		SuperTypes     []*TypeSpec         // Only valid if this is a type variable (egh)
		CollectionSpec *CollectionTypeSpec // nil or a collection specification
		NilAllowed     bool                // Whether a nil value is considered to satisfy the type constraint.
		NilElementsAllowed bool            // If a collection, if nil values are permitted. Replace by Params[i].NilAllowed
		typ interface{}  // The actual *RType - associated at generation time.
		
		// This is wrong !! The Method object needs to get a list of actual *RTypes of its return types
	}
	
	/*

	*/
	AritySpec struct {
		MinCard    int64
		MaxCard    int64     // -1 means N
		RangeStart token.Pos // start position in file of the first integer in the spec
		RangeEnd   token.Pos // just after the end of the last integer in the spec
	}
)

func (ts *TypeSpec) IsTypeVariable() bool {
	return ts.Name.Kind == token.TYPEVAR
}

// Pos and End implementations for spec nodes.
//
func (s *ImportSpec) Pos() token.Pos {
	if s.Name != nil {
		return s.Name.Pos()
	}
	return s.Path.Pos()
}
func (s *RelishImportSpec) Pos() token.Pos { return token.NoPos }

func (s *ValueSpec) Pos() token.Pos { return s.Names[0].Pos() }
func (s *TypeSpec) Pos() token.Pos  { return s.Name.Pos() }
func (s *AritySpec) Pos() token.Pos { return s.RangeStart }

func (s *ImportSpec) End() token.Pos { return s.Path.End() }
func (s *RelishImportSpec) End() token.Pos { return token.NoPos }
func (s *ValueSpec) End() token.Pos {
	if n := len(s.Values); n > 0 {
		return s.Values[n-1].End()
	}
	if s.Type != nil {
		return s.Type.End()
	}
	return s.Names[len(s.Names)-1].End()
}
func (s *TypeSpec) End() token.Pos { return s.Type.End() }

func (s *AritySpec) End() token.Pos { return s.RangeEnd }

// specNode() ensures that only spec nodes can be
// assigned to a Spec.
//
func (s *ImportSpec) specNode() {}
func (s *RelishImportSpec) specNode() {}
func (s *ValueSpec) specNode()  {}
func (s *TypeSpec) specNode()   {}
func (s *AritySpec) specNode()  {}

// A declaration is represented by one of the following declaration nodes.
//
type (
	// A BadDecl node is a placeholder for declarations containing
	// syntax errors for which no correct declaration nodes can be
	// created.
	//
	BadDecl struct {
		From, To token.Pos // position range of bad declaration
	}

	// A GenDecl node (generic declaration node) represents an import,
	// constant, type or variable declaration. A valid Lparen position
	// (Lparen.Line > 0) indicates a parenthesized declaration.
	//
	// Relationship between Tok value and Specs element type:
	//
	//	token.IMPORT  *ImportSpec
	//	token.CONST   *ValueSpec
	//	token.TYPE    *TypeSpec
	//	token.VAR     *ValueSpec
	//
	GenDecl struct {
		Doc    *CommentGroup // associated documentation; or nil
		TokPos token.Pos     // position of Tok
		Tok    token.Token   // IMPORT, CONST, TYPE, VAR
		Lparen token.Pos     // position of '(', if any
		Specs  []Spec
		Rparen token.Pos // position of ')', if any
	}

	// A FuncDecl node represents a function declaration.
	FuncDecl struct {
		Doc  *CommentGroup // associated documentation; or nil
		Recv *FieldList    // receiver (methods); or nil (functions)
		Name *Ident        // function/method name
		Type *FuncType     // position of Func keyword, parameters and results
		Body *BlockStmt    // function body; or nil (forward declaration)
	}
)

// Pos and End implementations for declaration nodes.
//
func (d *BadDecl) Pos() token.Pos  { return d.From }
func (d *GenDecl) Pos() token.Pos  { return d.TokPos }
func (d *FuncDecl) Pos() token.Pos { return d.Type.Pos() }

func (d *MethodDeclaration) Pos() token.Pos { return d.Name.Pos() }

func (d *ConstantDecl) Pos() token.Pos { return d.Name.Pos() }
func (d *InputArgDecl) Pos() token.Pos { return d.Name.Pos() }
func (d *ReturnArgDecl) Pos() token.Pos {
	if d.Name != nil {
		return d.Name.Pos()
	}
	return d.Type.Pos()
}

func (d *AttributeDecl) Pos() token.Pos { return d.Name.Pos() }

func (d *BadDecl) End() token.Pos { return d.To }
func (d *GenDecl) End() token.Pos {
	if d.Rparen.IsValid() {
		return d.Rparen + 1
	}
	return d.Specs[0].End()
}
func (d *FuncDecl) End() token.Pos {
	if d.Body != nil {
		return d.Body.End()
	}
	return d.Type.End()
}

func (d *MethodDeclaration) End() token.Pos {
	if d.Body != nil {
		return d.Body.End()
	}
	return d.Type.End()
}

func (d *ConstantDecl) End() token.Pos {
	return d.Value.End()
}

func (d *InputArgDecl) End() token.Pos {
	return d.Type.End()
}

func (d *ReturnArgDecl) End() token.Pos {
	return d.Type.End()
}

func (d *AttributeDecl) End() token.Pos {
	return d.Type.End()
}

// declNode() ensures that only declaration nodes can be
// assigned to a DeclNode.
//
func (d *BadDecl) declNode()  {}
func (d *GenDecl) declNode()  {}
func (d *FuncDecl) declNode() {}

func (d *RelationDecl) declNode()      {}
func (d *TypeDecl) declNode()          {}
func (d *MethodDeclaration) declNode() {}
func (d *ConstantDecl) declNode()      {}
func (d *InputArgDecl) declNode()      {}
func (d *ReturnArgDecl) declNode()     {}
func (d *AttributeDecl) declNode()     {}

// ----------------------------------------------------------------------------
// Files and packages

// A File node represents a Go source file.
//
// The Comments list contains all comments in the source file in order of
// appearance, including the comments that are pointed to from other nodes
// via Doc and Comment fields.
//
/*
type File struct {
	Doc        *CommentGroup   // associated documentation; or nil
	Package    token.Pos       // position of "package" keyword
	Name       *Ident          // package name
	Decls      []Decl          // top-level declarations; or nil
	Scope      *Scope          // package scope (this file only)
	Imports    []*ImportSpec   // imports in this file
	Unresolved []*Ident        // unresolved identifiers in this file
	Comments   []*CommentGroup // list of all comments in the source file
}
*/
/*
func (f *File) Pos() token.Pos { return f.Package }
func (f *File) End() token.Pos {
	if n := len(f.Decls); n > 0 {
		return f.Decls[n-1].End()
	}
	return f.Name.End()
}
*/

type File struct {
	Doc           *CommentGroup        // associated documentation; or nil (for relish should be a single comment)	
	Top           token.Pos            // position of first character of file
	Package       token.Pos            // position of "package" keyword	
	Name          *Ident               // package name
	Decls         []Decl               // top-level declarations; or nil (deprecated)	
	ConstantDecls []*ConstantDecl      // top-level declarations; or nil	
	TypeDecls     []*TypeDecl          // top-level declarations; or nil	
	RelationDecls []*RelationDecl      // top-level declarations; or nil
	MethodDecls   []*MethodDeclaration // top-level declarations; or nil
	Scope         *Scope               // package scope (this file only)
	Imports       []*ImportSpec        // imports in this file
	RelishImports []*RelishImportSpec        // imports in this file	
	Unresolved    []*Ident             // unresolved identifiers in this file
	Comments      []*CommentGroup      // list of all comments in the source file
	FileName      string
	FileSize      int
	FileLines     []int
}



/*
Converts a compact token.Pos representation of a position in a source code file to the expanded
token.Position form of the info. This can then be used in formatting error messages.
*/
func (f *File) Position(p token.Pos) token.Position {
   file := f.tokenFile()
   return file.Position(p)
}

/*
Stores the file name and source line position info into the ast.File in a form that can
be persisted with the ast.File when it is GOB-serialized into a .rlc file.
*/
func (f *File) StoreSourceFilePositionInfo(file *token.File) {
   f.FileName = file.Name()
   f.FileSize = file.Size()
   f.FileLines = file.Lines()
}

/*
  Reconstructs a token.File from the FileName, FileSize, FileLines values.
  This token.File can then be used to convert a token.Pos value of an ast Node into
  a token.Position value which can be used to give meaningful source file location info for runtime errors. 
*/
func (f *File) tokenFile() (file *token.File) {
   fset := token.NewFileSet()
   file = fset.AddFile(f.FileName, fset.Base(), f.FileSize)
   if ! file.SetLines(f.FileLines) {
   	   panic("Invalid source file line positions info for file " + f.FileName)
   }
   return
}


func (f *File) Pos() token.Pos { return f.Top }
func (f *File) End() token.Pos {
	var constantsEnd, typesEnd, relationsEnd, methodsEnd token.Pos
	if n := len(f.ConstantDecls); n > 0 {
		constantsEnd = f.ConstantDecls[n-1].End()
	}
	if n := len(f.TypeDecls); n > 0 {
		typesEnd = f.TypeDecls[n-1].End()
	}
	if n := len(f.RelationDecls); n > 0 {
		relationsEnd = f.RelationDecls[n-1].End()
	}
	if n := len(f.MethodDecls); n > 0 {
		methodsEnd = f.MethodDecls[n-1].End()
	}
	end := constantsEnd
	if typesEnd > end {
		end = typesEnd
	}
	if relationsEnd > end {
		end = relationsEnd
	}
	if methodsEnd > end {
		end = methodsEnd
	}
	if end > 0 {
		return end
	}
	return f.Name.End()
}

// A Package node represents a set of source files
// collectively building a Go package.
//
type Package struct {
	Name    string             // package name
	Scope   *Scope             // package scope across all files
	Imports map[string]*Object // map of package id -> package object
	Files   map[string]*File   // EGH relishd (Go) source files by filename
}

func (p *Package) Pos() token.Pos { return token.NoPos }
func (p *Package) End() token.Pos { return token.NoPos }
