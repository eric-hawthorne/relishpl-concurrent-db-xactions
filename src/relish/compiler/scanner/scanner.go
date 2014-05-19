// Substantial portions of the source code in this file 
// are Copyright 2009 The Go Authors. All rights reserved.
// Use of such source code is governed by a BSD-style
// license that can be found in the GO_LICENSE file.

// Modifications and additions which convert code to be part of a relish-language compiler 
// are Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of such source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// Package scanner implements a scanner for relish source text. Takes a []byte as
// source which can then be tokenized through repeated calls to the Scan
// function. Typical use:
//
//	var s Scanner
//	fset := token.NewFileSet()  // position information is relative to fset
//      file := fset.AddFile(filename, fset.Base(), len(src))  // register file
//	s.Init(file, src, nil /* no error handler */, 0)
//	for {
//		pos, tok, lit := s.Scan()
//		if tok == token.EOF {
//			break
//		}
//		// do something here with pos, tok, and lit
//	}
//
package scanner

import (
	"bytes"
	"fmt"
	"path/filepath"
	"relish/compiler/token"
	"strconv"
	"unicode"
	"unicode/utf8"
	. "relish/dbg"
)

/*
   saveable scanning state
*/
type ScanningState struct {
	Ch         rune    // current character
	Offset     int    // character offset
	RdOffset   int    // reading offset (position after current character)
	LineOffset int    // current line offset
	RuneColumn int    // position of the current unicode character on the current line (counting in runes).
	Str        string // most recently parsed identifier or numeric literal or literal string contents
}

const (
	INDENT  = 3
	MAX_COL = 120
)

// A Scanner holds the scanner's internal state while processing
// a given text.  It can be allocated as part of another data
// structure but must be initialized via Init before use.
//
type Scanner struct {
	// immutable state
	file *token.File  // source file handle
	dir  string       // directory portion of file.Name()
	src  []byte       // source
	err  ErrorHandler // error reporting; or nil
	mode uint         // scanning mode

	// scanning state
	ch         rune    // current character
	offset     int    // character offset
	rdOffset   int    // reading offset (position after current character)
	lineOffset int    // current line offset
	runeColumn int    // position of the current unicode character on the current line (counting in runes). 
	insertSemi bool   // insert a semicolon before next newline
	str        string // most recently parsed identifier or numeric literal or literal string contents
	prevState  ScanningState

	failStartOffset int // start of the string that parsing failed on
	failEndOffset   int // end of the string that parsing failed on

    // Following two items are used in generation of appropriate indentation-error compiler error messages.

	IndentWobble int // 0 if most recent text-starts-at-column test succeeded or is way off. 
	                 // 1 if indentation is +1 from there. -1 if indentation is -1 from there.

    IndentWobblePos token.Pos // If IndentWobble is +1 or -1, the file position where the non-space text starts

	// public state - ok to modify
	ErrorCount int // number of errors encountered
}

/*
The string that was being parsed when the scanner's Fail method was last called.
*/
func (S *Scanner) FailedOnString() string {

    if S.failStartOffset == S.failEndOffset {
	   return ""
	}
	
    r := rune(S.src[S.failStartOffset])

    if IsAsciiLetter(r) || IsAsciiDigit(r) {
    	srcLen := len(S.src)
    	for off := S.failStartOffset+1; off < srcLen; off++ {
           r = rune(S.src[off])  
           S.failEndOffset = off                  
           if ! (IsAsciiLetter(r) || IsAsciiDigit(r) || r == '_')  {	
           	   break
           }           
    	}
    }

	//DEBUG fmt.Printf("fail offsets: %v %v ch: '%s'\n",S.failStartOffset,S.failEndOffset,string(S.ch))	
	return string(S.src[S.failStartOffset:S.failEndOffset])
}




/*
The string created by slicing the src byte-slice from the startOffset to the endOffset.
*/
func (S *Scanner) Substring(startOffset, endOffset int) string {
	return string(S.src[startOffset:endOffset])
}

/*
   Returns the current scanning state (position etc) of the scanner in the file.
*/
func (S *Scanner) State() ScanningState {
	S.failStartOffset = 1
	S.failEndOffset = 1 // No FailedOnString yet.
	return ScanningState{Ch: S.ch, Offset: S.offset, RdOffset: S.rdOffset, LineOffset: S.lineOffset, RuneColumn: S.runeColumn, Str: S.str}
}

/*
   Restores the state of the scanner and returns false, indicating a failure to parse whatever was after the argument
   scanning state.
*/
func (S *Scanner) Fail(s ScanningState) bool {
	S.failStartOffset = S.offset
	S.failEndOffset = S.rdOffset
	S.ch = s.Ch
	S.offset = s.Offset
	S.rdOffset = s.RdOffset
	S.lineOffset = s.LineOffset
	S.runeColumn = s.RuneColumn
	S.str = s.Str
	S.prevState.Offset = -1 // Can't call Prev() after Restore() without an intervening Next()
	return false
}

/*
Sets the scanner state back by one position in the file so that Next() will return the same character at the same
position as it did on its last call.

Call after a Next() i.e. Only if YOU did the Next() and need to rescind it. Can't call after Prev() or Restore()
*/
func (S *Scanner) Prev() bool {
	if S.prevState.Offset == -1 {
		panic("Attempt to call Prev() other that after Next()")
	}
	S.failStartOffset = S.offset
	S.failEndOffset = S.rdOffset

	S.ch = S.prevState.Ch
	S.offset = S.prevState.Offset
	S.rdOffset = S.prevState.RdOffset
	S.lineOffset = S.prevState.LineOffset
	S.runeColumn = S.prevState.RuneColumn
	S.str = S.prevState.Str
	S.prevState.Offset = -1 // declare previous scanning state invalid. Can't call prev() at first or repeatedly.
	return false
}

// Read the next Unicode char into S.ch.
// S.ch < 0 means end-of-file.
//
func (S *Scanner) Next() {
	S.prevState.Ch = S.ch
	S.prevState.Offset = S.offset
	S.prevState.RdOffset = S.rdOffset
	S.prevState.LineOffset = S.lineOffset
	S.prevState.RuneColumn = S.runeColumn
	S.prevState.Str = S.str

	if S.rdOffset < len(S.src) {
		S.offset = S.rdOffset
		if S.ch == '\n' {
			S.lineOffset = S.offset
			//S.file.AddLine(S.offset)     // What were we doing here? !!!!!!!!!  I now do this for all lines in Init()
			S.runeColumn = 0
		}
		r, w := rune(S.src[S.rdOffset]), 1
		switch {
		case r == 0:
			S.error(S.offset, "illegal character NUL")
		case r == '\t':
			S.error(S.offset, "Tab characters are not permitted in a Relish source code file.")
		case r >= 0x80:
			// not ASCII
			r, w = utf8.DecodeRune(S.src[S.rdOffset:])
			if r == utf8.RuneError && w == 1 {
				S.error(S.offset, "illegal UTF-8 encoding")
			}
		}
		S.rdOffset += w
		S.runeColumn++
		S.ch = r
		pr := r
		if pr == '\n' {
			pr = '@'
		}
		Log(PARSE_,"%s - %v - %v\n", string(pr), pr, S.runeColumn)
		if S.runeColumn > MAX_COL && S.ch != '\n' {
			S.error(S.offset, fmt.Sprintf("A Relish source code file line cannot exceed %v characters (Unicode codepoints) in width.", MAX_COL))
		}
	} else {
		S.offset = len(S.src)
		if S.ch == '\n' {
			S.lineOffset = S.offset
			//S.file.AddLine(S.offset)
		}
		S.ch = -1 // eof
	}
}

/*
   The position of the character just teed up by Next()
*/
func (S *Scanner) Pos() token.Pos {
	return S.file.Pos(S.offset)
}

func (S *Scanner) FailedPos() token.Pos {
	if S.failStartOffset < S.failEndOffset {
		return S.file.Pos(S.failStartOffset)
	}
	return S.Pos()
}

/*
   The current character (rune).
*/
func (S *Scanner) Ch() rune {
	return S.ch
}

/*
   If the file contents starting at the scanner position match the argument string s, 
   the scanner is positioned just after the matched string occurrence, and this function
   returns true.
   If the match is not found, the scanner position is returned to as before the call to
   this function, and this function returns false.
*/
func (S *Scanner) Match(s string) bool {
	st := S.State()
	for _, c := range s {
		if S.ch == c {
			S.Next()
		} else {
			return S.Fail(st)
		}
	}
	return true
}

/*
Return true and advance if the current and next scanned characters match the arguments, respectively. 
Do not advance, and return false, if not.
*/
func (S *Scanner) Match1(c1 rune) bool {
	if S.ch != c1 {
		return false
	}
	S.Next()
	return true
}

/*
Return true and advance if the current and next scanned characters match the arguments, respectively. 
Do not advance, and return false, if not.
*/
func (S *Scanner) Match2(c1, c2 rune) bool {
	if S.ch != c1 {
		return false
	}
	S.Next()
	if S.ch != c2 {
		return S.Prev()
	}
	S.Next()
	return true
}

/*
Return true and advance if there is a single space. Do not advance, and return false, if not.
*/
func (S *Scanner) Space() bool {
	if S.ch != ' ' {
		return false
	}
	S.Next()
	return true
}

/*
Parse two spaces, or fail back to the starting position.
*/
func (S *Scanner) DoubleSpace() bool {
	return S.Match2(' ', ' ')
}

/*
Parse three spaces, or fail back to the starting position.
*/
func (S *Scanner) TripleSpace() bool {
	return S.Match("   ")
}

/*
Parse the beginning token of a line-comment or rest-of-line comment.
*/
func (S *Scanner) lineCommentSymbol() bool {
	return S.Match2('/', '/')
}

/*
Returns true if it gobbles nothing but spaces til it hits EOL or EOF.
Leaves the scanner looking at the EOL or EOF.
*/
func (S *Scanner) BlankToEOL() bool {
	st := S.State()
	for ; S.ch == ' '; S.Next() {
	}
	if S.ch != '\n' && S.ch != -1 {
		return S.Fail(st)
	}
	return true
}

/*
Returns true if there is nothing but spaces or comments til end of file.
Do we want to insist on particular columns for the comments?
*/
func (S *Scanner) BlankOrCommentsToEOF() bool {
	st := S.State()

	if !S.BlankOrCommentToEOL() {
		return false
	}
	if S.ch == -1 {
		return true
	}
	for S.BlanksAndBelow(1, false) {
		if S.ch != -1 && !S.LineComment() {
			return S.Fail(st)
		}
	}
	for ; S.ch == ' ' || S.ch == '\n'; S.Next() {
	}
	if S.ch == -1 {
		return true
	}
	return S.Fail(st)
}

/*
Consume all characters until matching the string s starting at the specified column.
Returns whether the terminator pattern was found, and if so, the byte-offset in the src
at which the pattern was found.
If the end is not found, the scanner state is returned to as before the function was called.
*/
func (S *Scanner) ConsumeTilMatchAtColumn(s string, column int) (found bool, contentEndOffset int) {
	st := S.State()

	c, _ := utf8.DecodeRuneInString(s)

	for ; S.ch != -1; S.Next() {
		if S.runeColumn == column && S.ch == c {
			contentEndOffset = S.offset
			if S.Match(s) {
				found = true
				return
			}
		}
	}
	S.Fail(st)
	return
}

//p.ConsumeToEOL() &&

//p.EmptyOrBelowOrRightOf(2)

/*
Returns true if it gobbles nothing but spaces or a rest-of-line comment til it hits EOL or EOF.
Leaves the scanner looking at the EOL or EOF.
If anything else is found before EOL or EOF, resets to where it was called at and returns false.
*/
func (S *Scanner) BlankOrCommentToEOL() bool {
	return S.RestOfLineComment() || S.BlankToEOL()
}

func (S *Scanner) RestOfLineComment() bool {
	st := S.State()

	if !S.DoubleSpace() {
		return false
	}

	for S.Space() {
	}

	//Does something.
	if !S.LineComment() {
		return S.Fail(st)
	}

	return true
}

func (S *Scanner) BlankToEOLOrLineCommentAtColumn(col int) bool {
	st := S.State()
	
    for ; S.ch == ' '; S.Next() {   // gobble spaces til first non-space
    }
    if S.Col() == col && S.LineComment() {
	   return true
    }
	
	if S.ch != '\n' && S.ch != -1 {
		return S.Fail(st)
	}
	return true		
}

/*
Parse a '//' comment which must begin at the current position. 
Requires at least one blank space or before the beginning of the comment.
*/
func (S *Scanner) LineComment() bool {

	if !S.lineCommentSymbol() {
		return false
	}

	hasSpace := false

	for ; S.ch != '\n' && S.ch != -1; S.Next() {
		if S.ch == ' ' {
			hasSpace = true
		} else if (!hasSpace) && (isLetter(S.ch) || isDigit(S.ch)) {
			S.error(S.offset, "Must follow // by at least one space before comment words.")
		}
	}
	return true
}


/*
Parse a '//' comment which must begin at the current position. 
Requires at least one blank space or before the beginning of the comment.
*/
func (S *Scanner) LineComments() bool {

   col := S.Col()

   if ! S.LineComment() {
      return false	
   }
   st := S.State()

   for S.Below(col) {
	  if ! S.LineComment() {
	     S.Fail(st)	
	     break
	  }
      st = S.State()	
   }
   return true
}


func (S *Scanner) LineCommentOrBlankLine(col int) bool {

	st := S.State()

	if !S.BlankOrCommentToEOL() {
		return false
	}

	if S.ch == -1 { // EOF
		return S.Fail(st)
	}
	S.Next()
	if !S.BlankToEOLOrLineCommentAtColumn(col) {
		return S.Fail(st)
	}

	//DEBUG pos := S.file.Position(S.Pos())
	//DEBUG fmt.Printf("Blankline puts us at %5d:%3d:\n", pos.Line, pos.Column)	

	return true 	
}


/*
Returns true if it gobbles whitespace or line-comment up to EOL then a blank or empty line.
Leaves the scanner looking at the blank or empty line's trailing EOL or EOF.
*/
func (S *Scanner) BlankLine() bool {

	st := S.State()

	if !S.BlankOrCommentToEOL() {
		return false
	}

	if S.ch == -1 { // EOF
		return S.Fail(st)
	}
	S.Next()
	if !S.BlankToEOL() {
		return S.Fail(st)
	}

	//DEBUG pos := S.file.Position(S.Pos())
	//DEBUG fmt.Printf("Blankline puts us at %5d:%3d:\n", pos.Line, pos.Column)	

	return true
}

/*
Returns true if there is nothing but rest-of-line comment or blanks then an EOL
and then the first non-blank character occurs at rune-column col.
Does not allow blank lines.
Does not succeed if it is sitting at EOL or EOF after it has gobbled the right number of spaces.
The assertion is that there is some language token starting at the specified position.
*/
func (S *Scanner) Below(col int) bool {

	S.IndentWobble = 0

	if S.ch != ' ' && S.ch != '\n' && S.ch != -1 && S.Col() == col {
		return true // Make it idempotent
	}

	st := S.State()

	if !S.BlankOrCommentToEOL() {
		return false
	}

	if S.ch == -1 { // EOF
		return S.Fail(st)
	}

	S.Next() // Gobble EOL		

	for S.Space() {
	} // Gobble spaces

    scannerCol := S.Col()
	if scannerCol != col {
		if scannerCol == col + 1 {
           S.IndentWobble = -1			
           S.IndentWobblePos = S.Pos()
		} else if scannerCol == col - 1 {
           S.IndentWobble = -1
           S.IndentWobblePos = S.Pos()           
		}
		return S.Fail(st)
	}

	if S.ch == '\n' || S.ch == -1 { // EOL or EOF
		return S.Fail(st)
	}


	return true
}

/*
Returns true if there is nothing but rest-of-line comment or blanks then an EOL
and then the first non-blank character occurs at rune-column col.

Gobbles blank lines before the indented thing. Note: Should I have a version
of this that only allows a certain number of blank lines or fewer?
*/
func (S *Scanner) BlanksAndBelow(col int, lineCommentsAllowed bool) bool {

	S.IndentWobble = 0

	if S.ch != ' ' && S.ch != '\n' && S.ch != -1 && S.Col() == col {
		return true // Make it idempotent
	}

	st := S.State()

	if !S.BlankOrCommentToEOL() {
		return false
	}

	if S.ch == -1 { // EOF
		return S.Fail(st)
	}

    if lineCommentsAllowed {
  	    for S.LineCommentOrBlankLine(col) {
	    } // Gobble blank lines with line comments starting indented at the correct column allowed
    } else {
		for S.BlankLine() {
		} // Gobble blank lines
    }
	S.Next() // Gobble EOL	

	for S.Space() {
	} // Gobble spaces

    scannerCol := S.Col()
	if scannerCol != col {
		if scannerCol == col + 1 {
           S.IndentWobble = 1			
           S.IndentWobblePos = S.Pos()
		} else if scannerCol == col - 1 {
           S.IndentWobble = -1
           S.IndentWobblePos = S.Pos()           
		}
		return S.Fail(st)
	}

	return true
}




/*
Gobbles blank lines until the first line comment indented one level from col. The line comment is also gobbled.
*/
func (S *Scanner) BlanksThenIndentedLineComment(col int) bool {
	st := S.State()
    if ! S.BlanksAndIndent(col, false) {
    	return false
    }
	if ! S.LineComment() {
		return S.Fail(st)
	} 
	return true
}

/*
Returns true if there is nothing but rest-of-line comment or blanks then an EOL
and then the first non-blank character occurs at a rune-column which is a single indent-level 
in from the fromCol.

should this gobble blank lines or with extra arg a max # of them? 
*/
func (S *Scanner) Indent(fromCol int) bool {
	return S.Below(fromCol + INDENT)
}

/*
Returns true if there is nothing but rest-of-line comment or blanks then an EOL
then zero or more blank lines
and then the first non-blank character occurs at a rune-column which is a single indent-level 
in from the fromCol.

should this gobble blank lines or with extra arg a max # of them? 
*/
func (S *Scanner) BlanksAndIndent(fromCol int, lineCommentsAllowed bool) bool {
	return S.BlanksAndBelow(fromCol + INDENT, lineCommentsAllowed)
}

/*
Returns true if there is nothing but rest-of-line comment or blanks then an EOL
then zero or more blank lines
and then the first non-blank character occurs at a rune-column which is two chars less than a single indent-level 
in from the fromCol. Used for < or > prefixes

should this gobble blank lines or with extra arg a max # of them? 
*/
func (S *Scanner) BlanksAndMiniIndent(fromCol int, lineCommentsAllowed bool) bool {
	return S.BlanksAndBelow(fromCol + INDENT - 2, lineCommentsAllowed)
}

// -----------------------------------------------------
// Scanning identifiers

// off-topic: A real person uri
// eric.george.hawthorne.1963.irvine.scotland.uk.people.relishing.org 2011
//
// Can only ever have one or the other as your public relish identity - cannot switch
// eric.george.hawthorne.1963.now.burnaby.bc.ca.people.relishing.org 2011

/*
   Parses a DNS-valid (ASCII) domain name with a few more restrictions such as
   any letters in it must be all lowercase. Also, disallows digits and hyphens
   in the top-level-domain part of the name.
*/
func (S *Scanner) ScanDomainName() bool {
	st := S.State()
	if !S.NonTopLevelDomainNamePart() {
		return false
	}
	for S.NonTopLevelDomainNamePart() {
	}

	if !S.TopLevelDomainLabel() {
		return S.Fail(st)
	}
	return true
}

func (S *Scanner) TopLevelDomainLabel() bool {
	st := S.State()
	if !IsAsciiLowercaseLetter(S.ch) {
		return false
	}
	S.Next()
	if !IsAsciiLowercaseLetter(S.ch) {
		return S.Fail(st)
	}
	S.Next()
	for ; IsAsciiLowercaseLetter(S.ch); S.Next() {
	}
	return true
}

/*
   Parses the non-top-level part of a valid dns domain name.
   Includes the trailing dot.
   Accepts lowercase names only.

   Good: abc.  abc1. a-1. xn-afb-23c-d3.
*/
func (S *Scanner) NonTopLevelDomainNamePart() bool {
	st := S.State()
	if !IsAsciiLowercaseLetter(S.ch) {
		return false
	}
	lastCharWasOkForEnd := true
	for {
		S.Next()
		if IsAsciiLowercaseLetterOrAsciiDigit(S.ch) {
			lastCharWasOkForEnd = true
		} else if S.ch == '-' { // only allowed in the middle
			lastCharWasOkForEnd = false
		} else if S.ch == '.' {
			if lastCharWasOkForEnd {
				S.Next()
				return true
			}
			return S.Fail(st)
		} else { // We came to another character in the wrong place.
			return S.Fail(st)
		}
	}
	return false // should never get here. Go is weird.
}

func (S *Scanner) ScanArtifactName() bool {
	return S.ScanPackageName()
}

func (S *Scanner) ScanPackageName() bool {
	st := S.State()
	if found, _ := S.PackageNamePart(); !found {
		return false
	}
	
	for S.ch == '/' {
	    st2 := S.State()		
		S.Next()
		if found, _ := S.PackageNamePart(); !found {
			S.Fail(st2)
			return true
		}
	}
	if S.ch == '.' {  // cannot end a full package name with a . because it looks like a domain name part.
		return S.Fail(st)
	}
	return true 
}

func (S *Scanner) ScanPackageAlias() (bool, string) {
	return S.PackageNamePart()
}

func (S *Scanner) PackageNamePart() (bool, string) {
	st := S.State()
	if !IsAsciiLowercaseLetter(S.ch) {
		return false, ""
	}
	lastCharWasOkForEnd := true
	for {
		S.Next()
		if IsAsciiLowercaseLetterOrAsciiDigit(S.ch) {
			lastCharWasOkForEnd = true
		} else if S.ch == '_' { // only allowed in the middle
			lastCharWasOkForEnd = false
	    } else if lastCharWasOkForEnd {
            if IsAsciiCapitalLetter(S.ch) {  // this is just the first part of a varname/methodname
               break
            }	    	
		    namePart := string(S.src[st.Offset:S.offset])
		    if namePart == "pkg" {   // not allowed as a package name part
			   break
		    }
		    if namePart == "artifacts" {   // not allowed as a package name part
			   break
		    }

			return true, namePart
		} else {
			break
		}
	}
	return S.Fail(st), ""
}

// -----------------------------------------------------
// Scanning other basic things

func (S *Scanner) ScanYear() bool {
	st := S.State()
	if S.ch != '1' && S.ch != '2' { // valid up to the year 2999
		return false
	}
	S.Next()
	if !IsAsciiDigit(S.ch) {
		return S.Fail(st)
	}
	S.Next()
	if !IsAsciiDigit(S.ch) {
		return S.Fail(st)
	}
	S.Next()
	if !IsAsciiDigit(S.ch) {
		return S.Fail(st)
	}
	S.Next()	
	return true
}

// -----------------------------------------------------
//

func (S *Scanner) ScanNumber() (bool, token.Token, string) {

	// current token start

	offs := S.offset
	tok := token.ILLEGAL
	found := false

	// determine token value
	switch ch := S.ch; {
	case digitVal(ch) < 10:
		tok = S.scanNumber(false)
		found = true
	default:
		S.Next() // always make progress
		switch ch {
		case '.':
			if digitVal(S.ch) < 10 {
				tok = S.scanNumber(true)
				found = true
			}
		default:
			S.Prev()
		}
	}

	// TODO(gri): The scanner API should change such that the literal string
	//            is only valid if an actual literal was scanned. This will
	//            permit a more efficient implementation.
	return found, tok, string(S.src[offs:S.offset])
}

// -----------------------------------------------------
// 

/*
   The rune column at which the current character was found.
*/
func (S *Scanner) Col() int {
	return S.runeColumn
}

// The mode parameter to the Init function is a set of flags (or 0).
// They control scanner behavior.
//
const (
	ScanComments      = 1 << iota // return comments as COMMENT tokens
	AllowIllegalChars             // do not report an error for illegal chars
	InsertSemis                   // automatically insert semicolons
)

// Init prepares the scanner S to tokenize the text src by setting the
// scanner at the beginning of src. The scanner uses the file set file
// for position information and it adds line information for each line.
// It is ok to re-use the same file when re-scanning the same file as
// line information which is already present is ignored. Init causes a
// panic if the file size does not match the src size.
//
// Calls to Scan will use the error handler err if they encounter a
// syntax error and err is not nil. Also, for each error encountered,
// the Scanner field ErrorCount is incremented by one. The mode parameter
// determines how comments, illegal characters, and semicolons are handled.
//
// Note that Init may call err if there is an error in the first character
// of the file.
//
func (S *Scanner) Init(file *token.File, src []byte, err ErrorHandler, mode uint) {
	// Explicitly initialize all fields since a scanner may be reused.
	if file.Size() != len(src) {
		panic("file size does not match src len")
	}
	S.file = file
	S.dir, _ = filepath.Split(file.Name())
	S.src = src
	S.err = err
	S.mode = mode

	S.ch = ' '
	S.offset = 0
	S.rdOffset = 0
	S.lineOffset = 0
	S.runeColumn = 0
	S.insertSemi = false
	S.ErrorCount = 0

	S.prevState.Offset = -1 // declare previous scanning state invalid. Can't call prev() at first or repeatedly.

	file.SetLinesForContent(src) // calculate line-end positions.
	S.Next()
}

func (S *Scanner) error(offs int, msg string) {
	if S.err != nil {
		S.err.Error(S.file.Position(S.file.Pos(offs)), msg)
	}
	S.ErrorCount++
}

var prefix = []byte("//line ")

// EGH May be obsolete
//
func (S *Scanner) interpretLineComment(text []byte) {
	if bytes.HasPrefix(text, prefix) {
		// get filename and line number, if any
		if i := bytes.LastIndex(text, []byte{':'}); i > 0 {
			if line, err := strconv.Atoi(string(text[i+1:])); err == nil && line > 0 {
				// valid //line filename:line comment;
				filename := filepath.Clean(string(text[len(prefix):i]))
				if !filepath.IsAbs(filename) {
					// make filename relative to current directory
					filename = filepath.Join(S.dir, filename)
				}
				// update scanner position
				S.file.AddLineInfo(S.lineOffset, filename, line-1) // -1 since comment applies to next line
			}
		}
	}
}

// EGH Obsolete
//
func (S *Scanner) scanComment() {
	// initial '/' already consumed; S.ch == '/' || S.ch == '*'
	offs := S.offset - 1 // position of initial '/'

	if S.ch == '/' {
		//-style comment
		S.Next()
		for S.ch != '\n' && S.ch >= 0 {
			S.Next()
		}
		if offs == S.lineOffset {
			// comment starts at the beginning of the current line
			S.interpretLineComment(S.src[offs:S.offset])
		}
		return
	}

	/*-style comment */
	S.Next()
	for S.ch >= 0 {
		ch := S.ch
		S.Next()
		if ch == '*' && S.ch == '/' {
			S.Next()
			return
		}
	}

	S.error(offs, "comment not terminated")
}

// EGH Something to do with comments
//
func (S *Scanner) findLineEnd() bool {
	// initial '/' already consumed

	defer func(offs int) {
		// reset scanner state to where it was upon calling findLineEnd
		S.ch = '/'
		S.offset = offs
		S.rdOffset = offs + 1
		S.Next() // consume initial '/' again
	}(S.offset - 1)

	// read ahead until a newline, EOF, or non-comment token is found
	for S.ch == '/' || S.ch == '*' {
		if S.ch == '/' {
			//-style comment always contains a newline
			return true
		}
		/*-style comment: look for newline */
		S.Next()
		for S.ch >= 0 {
			ch := S.ch
			if ch == '\n' {
				return true
			}
			S.Next()
			if ch == '*' && S.ch == '/' {
				S.Next()
				break
			}
		}
		S.skipWhitespace() // S.insertSemi is set
		if S.ch < 0 || S.ch == '\n' {
			return true
		}
		if S.ch != '/' {
			// non-comment token
			return false
		}
		S.Next() // consume '/'
	}

	return false
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch >= 0x80 && unicode.IsLetter(ch)
}

func IsAsciiLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || ch >= 0x80 && unicode.IsDigit(ch)
}

func IsAsciiDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func IsAsciiCapitalLetter(ch rune) bool {
	return 'A' <= ch && ch <= 'Z'
}

func IsAsciiLowercaseLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z'
}

func IsAsciiLowercaseLetterOrAsciiDigit(ch rune) bool {
	return ('a' <= ch && ch <= 'z') || ('0' <= ch && ch <= '9')
}

func IsAsciiCapitalLetterOrAsciiDigit(ch rune) bool {
	return ('A' <= ch && ch <= 'Z') || ('0' <= ch && ch <= '9')
}

func (S *Scanner) ScanTypeName() (bool, string) {
	st := S.State()

	// Capital Letter
	if !IsAsciiCapitalLetter(S.ch) {
		return false, ""
	}
	S.Next()
	if !IsAsciiLowercaseLetter(S.ch) {
		return S.Fail(st), ""
	}

	S.Next()
	for IsAsciiLowercaseLetterOrAsciiDigit(S.ch) {
		S.Next()
	}

    capsCount := 0
	// Capital Letter
	for IsAsciiCapitalLetter(S.ch) {
		capsCount++
		S.Next()
		if IsAsciiLowercaseLetter(S.ch) {
			if capsCount > 1 {
			   return S.Fail(st), ""	// Allow multiple caps together only at end.						
			}
			capsCount = 0
			S.Next()
			for IsAsciiLowercaseLetterOrAsciiDigit(S.ch) {
				S.Next()
			}
	    } else {
	    	for IsAsciiDigit(S.ch) {
		    	S.Next()	
	            if IsAsciiLowercaseLetter(S.ch) {
				   if capsCount > 1 {
				      return S.Fail(st), ""	// Allow multiple caps together only at end.						
				   }
				   capsCount = 0
				   S.Next()
				   for IsAsciiLowercaseLetterOrAsciiDigit(S.ch) {
					  S.Next()
				   }
				}	    	    	
		    }    
	    }
	}
	return true, string(S.src[st.Offset:S.offset])
}


func (S *Scanner) ScanMethodName() (bool, string) {
	if S.Match2('<','-') {
		return true, "<-"
	}
	return S.ScanVarName()
}

func (S *Scanner) ScanRelEndName() (bool, string) {
	return S.ScanVarName()
}

func (S *Scanner) ScanVarName() (bool, string) {
	st := S.State()

	if !IsAsciiLowercaseLetter(S.ch) {
		return false, ""
	}

	S.Next()
	for IsAsciiLowercaseLetterOrAsciiDigit(S.ch) {
		S.Next()
	}

	// Capital Letter
	for IsAsciiCapitalLetter(S.ch) {
		S.Next()
		if !IsAsciiLowercaseLetter(S.ch) {
			return S.Fail(st), ""
		}
		S.Next()
		for IsAsciiLowercaseLetterOrAsciiDigit(S.ch) {
			S.Next()
		}
	}

	return true, string(S.src[st.Offset:S.offset])
}

/*
True if the identifier in the source matches the word. 
Note, if the identifier is longer than the word, does not match the word.
*/
func (S *Scanner) MatchWord(word string) bool {
	st := S.State()
	found,ident := S.ScanVarName()
	if ! found {
		return false
	}
    if ident != word {
	   return S.Fail(st)
    }
    return true
}

func (S *Scanner) ScanConstName() (bool, string) {
	st := S.State()
	couldBeTypeVar := false
	length := 0

	if !IsAsciiCapitalLetter(S.ch) {
		return false, ""
	}
	length++
	if S.ch == 'T' {
		couldBeTypeVar = true
	}

	S.Next()

	if !IsAsciiDigit(S.ch) {
		couldBeTypeVar = false
	}

	for IsAsciiCapitalLetterOrAsciiDigit(S.ch) {
		length++
		S.Next()
	}

	for S.ch == '_' {
		S.Next()
		if !IsAsciiCapitalLetterOrAsciiDigit(S.ch) {
			return S.Fail(st), ""
		}
		length++
		S.Next()
		for IsAsciiCapitalLetterOrAsciiDigit(S.ch) {
			length++
			S.Next()
		}
	}

	if couldBeTypeVar && (length == 2) {
		return S.Fail(st), ""
	}

	if IsAsciiLowercaseLetter(S.ch) {
		return S.Fail(st), ""
	}

	return true, string(S.src[st.Offset:S.offset])
}

func (S *Scanner) ScanTypeVarName() (bool, string) {
	st := S.State()

	if !(S.ch == 'T') {
		return false, ""
	}

	S.Next()
	if !IsAsciiDigit(S.ch) {
		return S.Fail(st), ""
	}
	S.Next()
	if IsAsciiLowercaseLetterOrAsciiDigit(S.ch) {
		return S.Fail(st), ""
	}
	return true, string(S.src[st.Offset:S.offset])
}

func (S *Scanner) scanIdentifier() token.Token {
	offs := S.offset
	for isLetter(S.ch) || isDigit(S.ch) {
		S.Next()
	}
	return token.Lookup(S.src[offs:S.offset])
}

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= ch && ch <= 'f':
		return int(ch - 'a' + 10)
	case 'A' <= ch && ch <= 'F':
		return int(ch - 'A' + 10)
	}
	return 16 // larger than any legal digit val
}

func (S *Scanner) scanMantissa(base int) {
	for digitVal(S.ch) < base {
		S.Next()
	}
}

func (S *Scanner) scanNumber(seenDecimalPoint bool) token.Token {
	// digitVal(S.ch) < 10
	tok := token.INT

	if seenDecimalPoint {
		tok = token.FLOAT
		S.scanMantissa(10)
		goto exponent
	}

	if S.ch == '0' {
		// int or float
		offs := S.offset
		S.Next()
		if S.ch == 'x' || S.ch == 'X' {
			// hexadecimal int
			S.Next()
			S.scanMantissa(16)
			if S.offset-offs <= 2 {
				// only scanned "0x" or "0X"
				S.error(offs, "illegal hexadecimal number")
			}
		} else {
			// octal int or float
			seenDecimalDigit := false
			S.scanMantissa(8)
			if S.ch == '8' || S.ch == '9' {
				// illegal octal int or float
				seenDecimalDigit = true
				S.scanMantissa(10)
			}
			if S.ch == '.' || S.ch == 'e' || S.ch == 'E' || S.ch == 'i' {
				goto fraction
			}
			// octal int
			if seenDecimalDigit {
				S.error(offs, "illegal octal number")
			}
		}
		goto exit
	}

	// decimal int or float
	S.scanMantissa(10)

fraction:
	if S.ch == '.' {
		tok = token.FLOAT
		S.Next()
		S.scanMantissa(10)
	}

exponent:
	if S.ch == 'e' || S.ch == 'E' {
		tok = token.FLOAT
		S.Next()
		if S.ch == '-' || S.ch == '+' {
			S.Next()
		}
		S.scanMantissa(10)
	}

	if S.ch == 'i' {
		tok = token.IMAG
		S.Next()
	}

exit:
	return tok
}

func (S *Scanner) scanEscape(quote rune) {
	offs := S.offset

	var i, base, max uint32
	switch S.ch {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', quote:
		S.Next()
		return
	case '0', '1', '2', '3', '4', '5', '6', '7':
		i, base, max = 3, 8, 255
	case 'x':
		S.Next()
		i, base, max = 2, 16, 255
	case 'u':
		S.Next()
		i, base, max = 4, 16, unicode.MaxRune
	case 'U':
		S.Next()
		i, base, max = 8, 16, unicode.MaxRune
	default:
		S.Next() // always make progress
		S.error(offs, "unknown escape sequence")
		return
	}

	var x uint32
	for ; i > 0 && S.ch != quote && S.ch >= 0; i-- {
		d := uint32(digitVal(S.ch))
		if d >= base {
			S.error(S.offset, "illegal character in escape sequence")
			break
		}
		x = x*base + d
		S.Next()
	}
	// in case of an error, consume remaining chars
	for ; i > 0 && S.ch != quote && S.ch >= 0; i-- {
		S.Next()
	}
	if x > max || 0xd800 <= x && x < 0xe000 {
		S.error(offs, "escape sequence is invalid Unicode code point")
	}
}

func (S *Scanner) scanChar() {
	// '\'' opening already consumed
	offs := S.offset - 1

	n := 0
	for S.ch != '\'' {
		ch := S.ch
		n++
		S.Next()
		if ch == '\n' || ch < 0 {
			S.error(offs, "character literal not terminated")
			n = 1
			break
		}
		if ch == '\\' {
			S.scanEscape('\'')
		}
	}

	S.Next()

	if n != 1 {
		S.error(offs, "illegal character literal")
	}
}

// egh 
func (S *Scanner) ScanString() (bool, string) {
	if S.ch != '"' {
		return false, ""
	}
	S.Next()
	st := S.State()
	S.scanString()
	s := string(S.src[st.Offset : S.offset-1])
	//S.Next()
	return true, s
}

func (S *Scanner) scanString() {
	S.scanString1()

	S.Next()
}

func (S *Scanner) scanString1() {
	// '"' opening already consumed
	offs := S.offset - 1

	for S.ch != '"' {
		ch := S.ch
		S.Next()
		if ch == '\n' || ch < 0 {
			S.error(offs, "string not terminated")
			break
		}
		if ch == '\\' {
			S.scanEscape('"')
		}
	}
}

func (S *Scanner) scanRawString() {
	// '`' opening already consumed
	offs := S.offset - 1

	for S.ch != '`' {
		ch := S.ch
		S.Next()
		if ch < 0 {
			S.error(offs, "string not terminated")
			break
		}
	}

	S.Next()
}

func (S *Scanner) skipWhitespace() {
	for S.ch == ' ' || S.ch == '\t' || S.ch == '\n' && !S.insertSemi || S.ch == '\r' {
		S.Next()
	}
}

// Helper functions for scanning multi-byte tokens such as >> += >>= .
// Different routines recognize different length tok_i based on matches
// of ch_i. If a token ends in '=', the result is tok1 or tok3
// respectively. Otherwise, the result is tok0 if there was no other
// matching character, or tok2 if the matching character was ch2.

func (S *Scanner) switch2(tok0, tok1 token.Token) token.Token {
	if S.ch == '=' {
		S.Next()
		return tok1
	}
	return tok0
}

func (S *Scanner) switch3(tok0, tok1 token.Token, ch2 rune, tok2 token.Token) token.Token {
	if S.ch == '=' {
		S.Next()
		return tok1
	}
	if S.ch == ch2 {
		S.Next()
		return tok2
	}
	return tok0
}

func (S *Scanner) switch4(tok0, tok1 token.Token, ch2 rune, tok2, tok3 token.Token) token.Token {
	if S.ch == '=' {
		S.Next()
		return tok1
	}
	if S.ch == ch2 {
		S.Next()
		if S.ch == '=' {
			S.Next()
			return tok3
		}
		return tok2
	}
	return tok0
}

// Scan scans the next token and returns the token position,
// the token, and the literal string corresponding to the
// token. The source end is indicated by token.EOF.
//
// If the returned token is token.SEMICOLON, the corresponding
// literal string is ";" if the semicolon was present in the source,
// and "\n" if the semicolon was inserted because of a newline or
// at EOF.
//
// For more tolerant parsing, Scan will return a valid token if
// possible even if a syntax error was encountered. Thus, even
// if the resulting token sequence contains no illegal tokens,
// a client may not assume that no error occurred. Instead it
// must check the scanner's ErrorCount or the number of calls
// of the error handler, if there was one installed.
//
// Scan adds line information to the file added to the file
// set with Init. Token positions are relative to that file
// and thus relative to the file set.
//
func (S *Scanner) Scan() (token.Pos, token.Token, string) {
scanAgain:
	S.skipWhitespace()

	// current token start
	insertSemi := false
	offs := S.offset
	tok := token.ILLEGAL

	// determine token value
	switch ch := S.ch; {
	case isLetter(ch):
		tok = S.scanIdentifier()
		switch tok {
		case token.IDENT, token.BREAK, token.CONTINUE, token.FALLTHROUGH, token.RETURN:
			insertSemi = true
		}
	case digitVal(ch) < 10:
		insertSemi = true
		tok = S.scanNumber(false)
	default:
		S.Next() // always make progress
		switch ch {
		case -1:
			if S.insertSemi {
				S.insertSemi = false // EOF consumed
				return S.file.Pos(offs), token.SEMICOLON, "\n"
			}
			tok = token.EOF
		case '\n':
			// we only reach here if S.insertSemi was
			// set in the first place and exited early
			// from S.skipWhitespace()
			S.insertSemi = false // newline consumed
			return S.file.Pos(offs), token.SEMICOLON, "\n"
		case '"':
			insertSemi = true
			tok = token.STRING
			S.scanString()
		case '\'':
			insertSemi = true
			tok = token.CHAR
			S.scanChar()
		case '`':
			insertSemi = true
			tok = token.STRING
			S.scanRawString()
		case ':':
			tok = S.switch2(token.COLON, token.DEFINE)
		case '.':
			if digitVal(S.ch) < 10 {
				insertSemi = true
				tok = S.scanNumber(true)
			} else if S.ch == '.' {
				S.Next()
				if S.ch == '.' {
					S.Next()
					tok = token.ELLIPSIS
				}
			} else {
				tok = token.PERIOD
			}
		case ',':
			tok = token.COMMA
		case ';':
			tok = token.SEMICOLON
		case '(':
			tok = token.LPAREN
		case ')':
			insertSemi = true
			tok = token.RPAREN
		case '[':
			tok = token.LBRACK
		case ']':
			insertSemi = true
			tok = token.RBRACK
		case '{':
			tok = token.LBRACE
		case '}':
			insertSemi = true
			tok = token.RBRACE
		case '+':
			tok = S.switch3(token.ADD, token.ADD_ASSIGN, '+', token.INC)
			if tok == token.INC {
				insertSemi = true
			}
		case '-':
			tok = S.switch3(token.SUB, token.SUB_ASSIGN, '-', token.DEC)
			if tok == token.DEC {
				insertSemi = true
			}
		case '*':
			tok = S.switch2(token.MUL, token.MUL_ASSIGN)
		case '/':
			if S.ch == '/' || S.ch == '*' {
				// comment
				if S.insertSemi && S.findLineEnd() {
					// reset position to the beginning of the comment
					S.ch = '/'
					S.offset = offs
					S.rdOffset = offs + 1
					S.insertSemi = false // newline consumed
					return S.file.Pos(offs), token.SEMICOLON, "\n"
				}
				S.scanComment()
				if S.mode&ScanComments == 0 {
					// skip comment
					S.insertSemi = false // newline consumed
					goto scanAgain
				}
				tok = token.COMMENT
			} else {
				tok = S.switch2(token.QUO, token.QUO_ASSIGN)
			}
		case '%':
			tok = S.switch2(token.REM, token.REM_ASSIGN)
		case '^':
			tok = S.switch2(token.XOR, token.XOR_ASSIGN)
		case '<':
			if S.ch == '-' {
				S.Next()
				tok = token.ARROW
			} else {
				tok = S.switch4(token.LSS, token.LEQ, '<', token.SHL, token.SHL_ASSIGN)
			}
		case '>':
			tok = S.switch4(token.GTR, token.GEQ, '>', token.SHR, token.SHR_ASSIGN)
		case '=':
			tok = S.switch2(token.ASSIGN, token.EQL)
		case '!':
			tok = S.switch2(token.NOT, token.NEQ)
		case '&':
			if S.ch == '^' {
				S.Next()
				tok = S.switch2(token.AND_NOT, token.AND_NOT_ASSIGN)
			} else {
				tok = S.switch3(token.AND, token.AND_ASSIGN, '&', token.LAND)
			}
		case '|':
			tok = S.switch3(token.OR, token.OR_ASSIGN, '|', token.LOR)
		default:
			if S.mode&AllowIllegalChars == 0 {
				S.error(offs, fmt.Sprintf("illegal character %#U", ch))
			}
			insertSemi = S.insertSemi // preserve insertSemi info
		}
	}

	if S.mode&InsertSemis != 0 {
		S.insertSemi = insertSemi
	}

	// TODO(gri): The scanner API should change such that the literal string
	//            is only valid if an actual literal was scanned. This will
	//            permit a more efficient implementation.
	return S.file.Pos(offs), tok, string(S.src[offs:S.offset])
}
