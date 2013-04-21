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

	readMethod, err := RT.CreateMethod("relish.pl2012/relish_lib/pkg/files",nil,"read", []string{"file","buf"}, []string{"relish.pl2012/relish_lib/pkg/files/File","Bytes"}, []string{"Int","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	readMethod.PrimitiveCode = read

	fileInitMethod1, err := RT.CreateMethod("relish.pl2012/relish_lib/pkg/files",nil,"relish.pl2012/relish_lib/pkg/files/initFile", []string{"file","filePath"}, []string{"relish.pl2012/relish_lib/pkg/files/File","String"},  []string{"relish.pl2012/relish_lib/pkg/files/File","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	fileInitMethod1.PrimitiveCode = initFile  
	
	fileInitMethod2, err := RT.CreateMethod("relish.pl2012/relish_lib/pkg/files",nil,"relish.pl2012/relish_lib/pkg/files/initFile", []string{"file","filePath","mode"}, []string{"relish.pl2012/relish_lib/pkg/files/File","String","String"},  []string{"relish.pl2012/relish_lib/pkg/files/File","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	fileInitMethod2.PrimitiveCode = initFile
	
	fileInitMethod3, err := RT.CreateMethod("relish.pl2012/relish_lib/pkg/files",nil,"relish.pl2012/relish_lib/pkg/files/initFile", []string{"file","filePath","mode","permission"}, []string{"relish.pl2012/relish_lib/pkg/files/File","String","String","String"},  []string{"relish.pl2012/relish_lib/pkg/files/File","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	fileInitMethod3.PrimitiveCode = initFile		
}




 
///////////////////////////////////////////////////////////////////////////////////////////
// I/O functions

// read r File buf Bytes > n Int err String
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
//        "rw_rw____"
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
	
	}
    var perm os.FileMode
    var permUser, permGroup, permOther byte
    if permStr != "0" {
	   n := len(permStr)
       if n == 4 || n == 3 {
	      if n == 4 {
	         permStr = permStr[1:]
	      }
          permUser = permStr[0] - '0'	     
          permGroup = permStr[1] - '0'
          permOther = permStr[2] - '0'
          if permUser < 0 || permUser > 7 ||  permGroup < 0 || permGroup > 7 ||  permOther < 0 || permOther > 7 {
	         errStr = `Allowed formats for permission are e.g. "666" or "rwxrw_r__"`
	      } else {
		     perm = os.FileMode((64 * permUser) + (8 * permGroup) + permOther)
		  }
	   } else if n == 9 {   // "rw_______" style
		  permStrUser := permStr[0:3]
		  switch permStrUser {
	      case "___":
		     permUser = 0
		  case "__x":
			 permUser = 1
		  case "_w_":
			 permUser = 2
	      case "_wx":
		     permUser = 3
		  case "r__":
			 permUser = 4
		  case "r_x":
			 permUser = 5
	      case "rw_":
		     permUser = 6
	      case "rwx":
		     permUser = 7	
		  default:	
		     errStr = `Allowed formats for permission are e.g. "666" or "rwxrw_r__"`					
		  }	
		  permStrGroup := permStr[3:6]
		  switch permStrGroup {
	      case "___":
		     permGroup = 0
		  case "__x":
			 permGroup = 1
		  case "_w_":
			 permGroup = 2
	      case "_wx":
		     permGroup = 3
		  case "r__":
			 permGroup = 4
		  case "r_x":
			 permGroup = 5
	      case "rw_":
		     permGroup = 6
	      case "rwx":
		     permGroup = 7	
		  default:	
		     errStr = `Allowed formats for permission are e.g. "666" or "rwxrw_r__"`					
		  }		
		  permStrOther := permStr[6:]
		  switch permStrOther {
	      case "___":
		     permOther = 0
		  case "__x":
			 permOther = 1
		  case "_w_":
			 permOther = 2
	      case "_wx":
		     permOther = 3
		  case "r__":
			 permOther = 4
		  case "r_x":
			 permOther = 5
	      case "rw_":
		     permOther = 6
	      case "rwx":
		     permOther = 7	
		  default:	
		     errStr = `Allowed formats for permission are e.g. "666" or "rwxrw_r__"`					
		  }
		  if errStr == "" {
		     perm = os.FileMode((64 * permUser) + (8 * permGroup) + permOther)			
		  }		
	   } else {
		  errStr = `Allowed formats for permission are e.g. "666" or "rwxrw_r__"`
	   }
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

