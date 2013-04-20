// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU LESSER GPL v3 license, found in the LICENSE_LGPL3 file.

package files_methods

/*
   files.go - native methods for relish standard library 'files' package. 
*/

import (
	. "relish/runtime/data"
	"os"
)


func InitFilesMethods() {

	readMethod, err := RT.CreateMethod("relish.pl2012/relish_lib/pkg/files",nil,"read", []string{"file","buf"}, []string{"relish.pl2012/relish/package/files/File","Bytes"}, []string{"Int","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	readMethod.PrimitiveCode = read

	fileInitMethod, err := RT.CreateMethod("relish.pl2012/relish_lib/pkg/files",nil,"initFile", []string{"mode","filePath"}, []string{"String","String"},  []string{"relish.pl2012/relish/package/files/File","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	fileInitMethod.PrimitiveCode = initFile  
}




 
///////////////////////////////////////////////////////////////////////////////////////////
// I/O functions

func read(th InterpreterThread, objects []RObject) []RObject {
    // TODO
	return nil
}






///////////////////////////////////////////////////////////////////////////////////////////
// Type init functions


// file err = File mode filePath
func initFile(th InterpreterThread, objects []RObject) []RObject {
   
    fileWrapper := objects[0].(*GoWrapper)

    // mode := string(objects[1].(String))

    filePath := string(objects[2].(String))

    var errStr string    

    // Accept a read/write etc mode parameter
//    switch mode 
    file,err := os.Open(filePath) // For read access.

	if err != nil {
		errStr = err.Error()
	} else {
	   fileWrapper.GoObj = file
    }
	return []RObject{fileWrapper,String(errStr)}
}

