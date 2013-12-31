// Portions of the source code in this file 
// are Copyright 2009 The Go Authors. All rights reserved.
// Use of such source code is governed by a BSD-style
// license that can be found in the GO_LICENSE file.

// Modifications and additions which convert code to be part of a relish-language compiler test suite
// are Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of such source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

package generator

import (
	"relish/compiler/token"
	"relish/compiler/ast"	
	"relish/compiler/parser"
	"relish"
	"relish/runtime/builtin"	
	"relish/dbg"
	// "os"
	"testing"
)

var fset = token.NewFileSet()

var illegalInputs = []interface{}{
	nil,
	3.14,
	[]byte(nil),
	"foo!",
	`package p; func f() { if /* should have condition */ {} };`,
	`package p; func f() { if ; /* should have condition */ {} };`,
	`package p; func f() { if f(); /* should have condition */ {} };`,
	`package p; const c; /* should have constant value */`,
	`package p; func f() { if _ = range x; true {} };`,
	`package p; func f() { switch _ = range x; true {} };`,
	`package p; func f() { for _ = range x ; ; {} };`,
	`package p; func f() { for ; ; _ = range x {} };`,
	`package p; func f() { for ; _ = range x ; {} };`,
	`package p; var a = [1]int; /* illegal expression */`,
	`package p; var a = [...]int; /* illegal expression */`,
	`package p; var a = struct{} /* illegal expression */`,
	`package p; var a = func(); /* illegal expression */`,
	`package p; var a = interface{} /* illegal expression */`,
	`package p; var a = []int /* illegal expression */`,
	`package p; var a = map[int]int /* illegal expression */`,
	`package p; var a = chan int; /* illegal expression */`,
	`package p; var a = []int{[]int}; /* illegal expression */`,
	`package p; var a = ([]int); /* illegal expression */`,
	`package p; var a = a[[]int:[]int]; /* illegal expression */`,
	`package p; var a = <- chan int; /* illegal expression */`,
	`package p; func f() { select { case _ <- chan int: } };`,
}

/*
func TestParseIllegalInputs(t *testing.T) {
	for _, src := range illegalInputs {
		_, err := ParseFile(fset, "", src, 0)
		if err == nil {
			t.Errorf("ParseFile(%v) should have failed", src)
		}
	}
}

var validPrograms = []interface{}{
	"package p\n",
	`package p;`,
	`package p; import "fmt"; func f() { fmt.Println("Hello, World!") };`,
	`package p; func f() { if f(T{}) {} };`,
	`package p; func f() { _ = (<-chan int)(x) };`,
	`package p; func f() { _ = (<-chan <-chan int)(x) };`,
	`package p; func f(func() func() func());`,
	`package p; func f(...T);`,
	`package p; func f(float, ...int);`,
	`package p; func f(x int, a ...int) { f(0, a...); f(1, a...,) };`,
	`package p; type T []int; var a []bool; func f() { if a[T{42}[0]] {} };`,
	`package p; type T []int; func g(int) bool { return true }; func f() { if g(T{42}[0]) {} };`,
	`package p; type T []int; func f() { for _ = range []int{T{42}[0]} {} };`,
	`package p; var a = T{{1, 2}, {3, 4}}`,
	`package p; func f() { select { case <- c: case c <- d: case c <- <- d: case <-c <- d: } };`,
	`package p; func f() { select { case x := (<-c): } };`,
	`package p; func f() { if ; true {} };`,
	`package p; func f() { switch ; {} };`,
	`package p; func f() { for _ = range "foo" + "bar" {} };`,
}

func TestParseValidPrograms(t *testing.T) {
	for _, src := range validPrograms {
		_, err := ParseFile(fset, "", src, 0)
		if err != nil {
			t.Errorf("ParseFile(%q): %v", src, err)
		}
	}
}

var validFiles = []string{
	"parser.go",
	"parser_test.go",
}
*/

var validFiles = []string{
	"../../test_interpreter.rel",
	// "../../test_parser.rel",
}

func TestParse3(t *testing.T) {
	dbg.InitLogging(0)
	relish.InitRuntime("relish.db")
    builtin.InitBuiltinFunctions()	
	var g *Generator
	for _, filename := range validFiles {
		fileNode, err := parser.ParseFile(fset, filename, nil, parser.DeclarationErrors | parser.Trace)
		if err != nil {
			t.Errorf("ParseFile(%s): %v", filename, err)
		}
		ast.Print(fset,fileNode)
		
	    fileNameRoot := filename[:len(filename)-4]	
		
	    g = NewGenerator(fileNode,fileNameRoot)
        g.GenerateCode()	
	    //g.TestWalk()
	}
	
	g.Interp.RunMain()
}

/*
func nameFilter(filename string) bool {
	switch filename {
	case "parser.go":
	case "interface.go":
	case "parser_test.go":
	default:
		return false
	}
	return true
}

func dirFilter(f *os.FileInfo) bool { return nameFilter(f.Name) }

func TestParse4(t *testing.T) {
	path := "."
	pkgs, err := ParseDir(fset, path, dirFilter, 0)
	if err != nil {
		t.Fatalf("ParseDir(%s): %v", path, err)
	}
	if len(pkgs) != 1 {
		t.Errorf("incorrect number of packages: %d", len(pkgs))
	}
	pkg := pkgs["parser"]
	if pkg == nil {
		t.Errorf(`package "parser" not found`)
		return
	}
	for filename := range pkg.Files {
		if !nameFilter(filename) {
			t.Errorf("unexpected package file: %s", filename)
		}
	}
}
*/
