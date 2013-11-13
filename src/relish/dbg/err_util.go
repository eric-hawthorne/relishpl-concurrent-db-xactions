package dbg

/*
Utilities for reporting errors.
*/

import (
        "relish/compiler/token"
        "strings"
        )

/*
Formats an error message nicely, giving the relish source file location of the error if possible.
*/
func FmtErr(pos token.Position, message string) string {
	if pos.Filename != "" || pos.IsValid() {
		// don't print "<unknown position>"
		// TODO(gri) reconsider the semantics of Position.IsValid
		
        fileNamePosString := pos.String()
        srcPos := strings.LastIndex(fileNamePosString, "/src/")
        artifactDir := fileNamePosString[:srcPos+5]
        packageFilePos := fileNamePosString[srcPos+5:] 
		
		return "\n" + packageFilePos + ":\n\n" + message + "\n\nError in software artifact\n" + artifactDir + "\n"
	}
	return message
}