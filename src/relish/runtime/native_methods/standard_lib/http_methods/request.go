// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU LESSER GPL v3 license, found in the LICENSE_LGPL3 file.

package http_methods

/*
   request.go - native methods for http request objects and their attributes, such as uploaded files and cookies.
   These methods are used by types defined in the relish standard library 'http' package. 
*/

import (
	. "relish/runtime/data"
	"os"
	"io"
	"bufio"
	"http"
	"mime/multipart"
)


func InitHttpMethods() {

    // buf = Bytes 1000
    // n err = read file buf
	readMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/files",nil,"read", []string{"file","buf"}, []string{"shared.relish.pl2012/relish_lib/pkg/files/File","Bytes"}, []string{"Int","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	readMethod.PrimitiveCode = read

	// readAllText 
	//    f File 
	//    addMissingLinefeed Bool = false
	// > 
	//    fileContent String err String
	//
	readAllTextMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/files",nil,"readAllText", []string{"file"}, []string{"shared.relish.pl2012/relish_lib/pkg/files/File"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	readAllTextMethod.PrimitiveCode = readAllText

	readAllTextMethod2, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/files",nil,"readAllText", []string{"file","addMissingLinefeed"}, []string{"shared.relish.pl2012/relish_lib/pkg/files/File","Bool"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	readAllTextMethod2.PrimitiveCode = readAllText
	
	// readAllBinary
	//    f File 
	// > 
	//    fileContent String err String
	//
	readAllBinaryMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/files",nil,"readAllBinary", []string{"file"}, []string{"shared.relish.pl2012/relish_lib/pkg/files/File"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	readAllBinaryMethod.PrimitiveCode = readAllBinary


	

    // err = close file
	closeMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/files",nil,"close", []string{"file"}, []string{"shared.relish.pl2012/relish_lib/pkg/files/File"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	closeMethod.PrimitiveCode = close







	fileInitMethod1, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/files",nil,"shared.relish.pl2012/relish_lib/pkg/files/initFile", []string{"file","filePath"}, []string{"shared.relish.pl2012/relish_lib/pkg/files/File","String"},  []string{"shared.relish.pl2012/relish_lib/pkg/files/File","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	fileInitMethod1.PrimitiveCode = initFile  
	
	fileInitMethod2, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/files",nil,"shared.relish.pl2012/relish_lib/pkg/files/initFile", []string{"file","filePath","mode"}, []string{"shared.relish.pl2012/relish_lib/pkg/files/File","String","String"},  []string{"shared.relish.pl2012/relish_lib/pkg/files/File","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	fileInitMethod2.PrimitiveCode = initFile
	
	fileInitMethod3, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/files",nil,"shared.relish.pl2012/relish_lib/pkg/files/initFile", []string{"file","filePath","mode","permission"}, []string{"shared.relish.pl2012/relish_lib/pkg/files/File","String","String","String"},  []string{"shared.relish.pl2012/relish_lib/pkg/files/File","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	fileInitMethod3.PrimitiveCode = initFile		
}




 
///////////////////////////////////////////////////////////////////////////////////////////
// I/O functions

// read f File buf Bytes > n Int err String
//
func read(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)
	buf := objects[1].(Bytes)
	b := ([]byte)(buf)
	file := wrapper.GoObj.(*os.File)
	n, err := file.Read(b)
	errStr := ""
	if err != nil {
	   errStr = err.Error()
	}
	return []RObject{Int(n),String(errStr)}
}



// readAllText 
//    f File 
//    addMissingLinefeed Bool = false
// > 
//    fileContent String err String
//
func readAllText(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)
	file := wrapper.GoObj.(*os.File)
    addMissingLinefeed := false
    if len(objects) == 2 {
    	addMissingLinefeed = bool(objects[1].(Bool))
    }


	br := bufio.NewReader(file)
	
	var err error 
	var content, line []byte  
	for {
	   line, err = br.ReadBytes('\n')
	   n := len(line)
	   if n > 1 && line[n-2] == '\r' {
		  line[n-2] = '\n'
		  line = line[:n-1]
	   }
	   content = append(content, line...)
	   if err != nil {
		  if err == io.EOF {
			 err = nil
		  }
		  break
	   }
    }
	errStr := ""
    if err == nil {
       if len(content) > 0 && addMissingLinefeed && content[len(content)-1] != '\n' {
	      content = append(content,'\n')
	   }	
	} else {
	   errStr = err.Error()
	}
	return []RObject{String(string(content)),String(errStr)}
}


// readAllBinary
//    f File 
// > 
//    fileContent String err String
//
func readAllBinary(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)
	file := wrapper.GoObj.(*os.File)	
	var buf []byte = make([]byte,8192)
	var content []byte
	var err error 	
	for {
	   n, err := file.Read(buf)
	   b := buf[:n]
	   content = append(content, b...)
	   if err != nil {
		  if err == io.EOF {
			 err = nil
		  }
		  break
	   }
    }
	errStr := ""
    if err != nil {
	   errStr = err.Error()
	}
	return []RObject{String(string(content)),String(errStr)}
}





// Close closes the File, rendering it unusable for I/O. It returns an error, if any.
func close(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)	
	file := wrapper.GoObj.(*os.File)
	err := file.Close()
	errStr := ""
	if err != nil {
	   errStr = err.Error()
	}
	return []RObject{String(errStr)}
}


///////////
// Go Types



///////////////////////////////////////////////////////////////////////////////////////////
// Type init functions


// file err = File filePath [mode [perm]]
//
// mode = "r" - read only (the default)
//        "w" - write only, creates if not exist, or truncates if does
//        "a" - appends, creates if not exist
//        "r+" - read and write, creates if not exist, does not truncate
//        "w+" - read and write, creates if not exist, truncates if does
//        "a+" - read and append, creates if not exist, read pointer at end
//        "nw" - write only, creates. File must not exist
//        "nw+" - read write, creates, File must not exist
//
// perm = "rwxrwxrwx"
//        "rw-rw----"
//        "777"
//        "660"
//
func initFile(th InterpreterThread, objects []RObject) []RObject {
   
    fileWrapper := objects[0].(*GoWrapper)

     filePath := string(objects[1].(String))   
 
    modeStr := "r"
    if len(objects) > 2 {
       modeStr = string(objects[2].(String))    
    }
    permStr := "666"
    if len(objects) > 3 {
       permStr = string(objects[3].(String))    
    }

    var errStr string    

// Accept a read/write etc mode parameter
//    switch mode 

    var flag int 
    switch modeStr {
	   case "r":
		  flag = os.O_RDONLY
		  permStr = "0"
	   case "w":
		  flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	   case "a":
		  flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
       case "r+":
		  flag = os.O_RDWR | os.O_CREATE	
	   case "w+":
		  flag = os.O_RDWR | os.O_CREATE | os.O_TRUNC		
	   case "a+":
		  flag = os.O_RDWR | os.O_CREATE | os.O_APPEND			
	   case "nw":
		  flag = os.O_WRONLY | os.O_CREATE | os.O_EXCL			
	   case "nw+":
		  flag = os.O_RDWR | os.O_CREATE | os.O_EXCL			
	   default:
	   	  errStr = `File mode if supplied must be one of "r","w","a","r+","w+","a+","nw","nw+".`
	}

    var perm os.FileMode

    if errStr == "" {
       perm, errStr = getFilePermissions(permStr)
    }

    if errStr == "" {
	    file,err := os.OpenFile(filePath, flag, perm) 

		if err != nil {
			errStr = err.Error()
		} else {
		   fileWrapper.GoObj = file
	    }
    }
	return []RObject{fileWrapper,String(errStr)}
}



/*
Helper function.
Given the argument permissions string, which can be in a format like "666" "0666" or "rwxrw_r__",
Return the corresponding FileMode value with the appropriate permission bits set.
See Go os.FileMode specification.
Returns a non-empty errStr error message if the input is invalidly formatted.
Note. the input permStr value "0" is allowed and results in a perm with no permission bits set.
*/
func getFilePermissions(permStr string) (perm os.FileMode, errStr string) {
    var permUser, permGroup, permOther uint32
    if permStr != "0" {
	   n := len(permStr)
       if n == 4 || n == 3 {
	      if n == 4 {
	         permStr = permStr[1:]
	      }
          permUser = uint32(permStr[0] - '0')	     
          permGroup = uint32(permStr[1] - '0')
          permOther = uint32(permStr[2] - '0')
          if permUser < 0 || permUser > 7 ||  permGroup < 0 || permGroup > 7 ||  permOther < 0 || permOther > 7 {
	         errStr = `Allowed formats for permission are e.g. "666" or "rwxrw_r__"`
	      } else {
             var p uint32 = (64 * permUser) + (8 * permGroup) + permOther		
		     perm = os.FileMode(p)
		     //fmt.Println(permUser, permGroup, permOther)
		     //fmt.Println(p)
		     //fmt.Println(perm)
		  }
	   } else if n == 9 {   // "rw-------" style
		  permStrUser := permStr[0:3]
		  switch permStrUser {
	      case "---":
		     permUser = 0
		  case "--x":
			 permUser = 1
		  case "-w-":
			 permUser = 2
	      case "-wx":
		     permUser = 3
		  case "r--":
			 permUser = 4
		  case "r-x":
			 permUser = 5
	      case "rw-":
		     permUser = 6
	      case "rwx":
		     permUser = 7	
		  default:	
		     errStr = `Allowed formats for permission are e.g. "666" or "rwxrw-r--"`					
		  }	
		  permStrGroup := permStr[3:6]
		  switch permStrGroup {
	      case "---":
		     permGroup = 0
		  case "--x":
			 permGroup = 1
		  case "-w-":
			 permGroup = 2
	      case "-wx":
		     permGroup = 3
		  case "r--":
			 permGroup = 4
		  case "r-x":
			 permGroup = 5
	      case "rw-":
		     permGroup = 6
	      case "rwx":
		     permGroup = 7	
		  default:	
		     errStr = `Allowed formats for permission are e.g. "666" or "rwxrw-r--"`					
		  }		
		  permStrOther := permStr[6:]
		  switch permStrOther {
	      case "---":
		     permOther = 0
		  case "--x":
			 permOther = 1
		  case "-w-":
			 permOther = 2
	      case "-wx":
		     permOther = 3
		  case "r--":
			 permOther = 4
		  case "r-x":
			 permOther = 5
	      case "rw-":
		     permOther = 6
	      case "rwx":
		     permOther = 7	
		  default:	
		     errStr = `Allowed formats for permission are e.g. "666" or "rwxrw-r--"`					
		  }
		  if errStr == "" {
             var p uint32 = (64 * permUser) + (8 * permGroup) + permOther	
		     perm = os.FileMode(p)						
		  }		
	   } else {
		  errStr = `Allowed formats for permission are e.g. "666" or "rwxrw-r--"`
	   }
    }
    return
}