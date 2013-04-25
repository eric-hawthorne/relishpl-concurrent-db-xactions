// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU LESSER GPL v3 license, found in the LICENSE_LGPL3 file.

package files_methods

/*
   files.go - native methods for relish standard library 'files' package. 
*/

import (
	. "relish/runtime/data"
	"os"
	"fmt"
	"io"
	"bufio"
)


func InitFilesMethods() {

    // buf = Bytes 1000
    // n err = read file buf
	readMethod, err := RT.CreateMethod("relish.pl2012/relish_lib/pkg/files",nil,"read", []string{"file","buf"}, []string{"relish.pl2012/relish_lib/pkg/files/File","Bytes"}, []string{"Int","String"}, false, 0, false)
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
	readAllTextMethod, err := RT.CreateMethod("relish.pl2012/relish_lib/pkg/files",nil,"readAllText", []string{"file"}, []string{"relish.pl2012/relish_lib/pkg/files/File"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	readAllTextMethod.PrimitiveCode = readAllText

	readAllTextMethod2, err := RT.CreateMethod("relish.pl2012/relish_lib/pkg/files",nil,"readAllText", []string{"file","addMissingLinefeed"}, []string{"relish.pl2012/relish_lib/pkg/files/File","Bool"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	readAllTextMethod2.PrimitiveCode = readAllText
	
	// readAllBinary
	//    f File 
	// > 
	//    fileContent String err String
	//
	readAllBinaryMethod, err := RT.CreateMethod("relish.pl2012/relish_lib/pkg/files",nil,"readAllBinary", []string{"file"}, []string{"relish.pl2012/relish_lib/pkg/files/File"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	readAllBinaryMethod.PrimitiveCode = readAllBinary


	// n err = write file val
	writeMethod, err := RT.CreateMethod("relish.pl2012/relish_lib/pkg/files",nil,"write", []string{"file","val"}, []string{"relish.pl2012/relish_lib/pkg/files/File","Any"}, []string{"Int","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	writeMethod.PrimitiveCode = write	

    // err = close file
	closeMethod, err := RT.CreateMethod("relish.pl2012/relish_lib/pkg/files",nil,"close", []string{"file"}, []string{"relish.pl2012/relish_lib/pkg/files/File"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	closeMethod.PrimitiveCode = close


	// sync f File > err String
	// """
	//  Commits the current contents of the file to stable storage.
	//  Typically, this means flushing the file system's in-memory copy
	//  of recently written data to disk.
	// 
    //  err = sync file
    // """
	syncMethod, err := RT.CreateMethod("relish.pl2012/relish_lib/pkg/files",nil,"sync", []string{"file"}, []string{"relish.pl2012/relish_lib/pkg/files/File"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	syncMethod.PrimitiveCode = sync	

	//
	//
	// remove path String > err String
	// """
	//  Removes the named file or directory.
	//
	//  err = remove path
	// """
	//
	removeMethod, err := RT.CreateMethod("relish.pl2012/relish_lib/pkg/files",nil,"remove", []string{"path"}, []string{"String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	removeMethod.PrimitiveCode = remove		


	// removeAll path String > err String  
	// """
	//  Removes the named file or directory and all sub directories and contained files recursively.
	//
	//  err = removeAll path	
	// """
	//
	removeAllMethod, err := RT.CreateMethod("relish.pl2012/relish_lib/pkg/files",nil,"removeAll", []string{"path"}, []string{"String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	removeAllMethod.PrimitiveCode = removeAll		

	// rename oldName String newName String > err String
	// """
	//  Renames the file or directory.
	//
	//  err = rename oldName newName	
	// """
	//
	renameMethod, err := RT.CreateMethod("relish.pl2012/relish_lib/pkg/files",nil,"rename", []string{"oldName","newName"}, []string{"String","String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	renameMethod.PrimitiveCode = rename		





	// name size mode modTime isDir fileExists err =
	//    statPrimitive path
	//
	statPrimitiveMethod, err := RT.CreateMethod("relish.pl2012/relish_lib/pkg/files",nil,"statPrimitive", 
		                                        []string{"path"}, []string{"String"}, 
	                                            []string{"String","Int","Uint32","Time","Bool","Bool","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	statPrimitiveMethod.PrimitiveCode = statPrimitive			


	// name size mode modTime isDir fileExists err =
	//    lstatPrimitive path
	//
	lstatPrimitiveMethod, err := RT.CreateMethod("relish.pl2012/relish_lib/pkg/files",nil,"lstatPrimitive", 
		                                        []string{"path"}, []string{"String"}, 
	                                            []string{"String","Int","Uint32","Time","Bool","Bool","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	lstatPrimitiveMethod.PrimitiveCode = lstatPrimitive	











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





// write f File buf Bytes > n Int err String
// write f File val Any > n Int err String
// If not Bytes, converts to String
//
func write(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)
	var b []byte
	buf,isBytes := objects[1].(Bytes)
	if isBytes {
		b = ([]byte)(buf)
	} else {
		b = ([]byte)(string(objects[1].String()))
	}
	
	file := wrapper.GoObj.(*os.File)
	n, err := file.Write(b)
	errStr := ""
	if err != nil {
	   errStr = err.Error()
	}
	return []RObject{Int(n),String(errStr)}
}


// sync f File > err String
//
func sync(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)
	file := wrapper.GoObj.(*os.File)
	err := file.Sync()
	errStr := ""
	if err != nil {
	   errStr = err.Error()
	}
	return []RObject{String(errStr)}
}


// remove path String > err String
// """
//  Removes the named file or directory.
// """
//
func remove(th InterpreterThread, objects []RObject) []RObject {
	
    path := string(objects[0].(String))

	err := os.Remove(path)
	errStr := ""
	if err != nil {
	   errStr = err.Error()
	}
	return []RObject{String(errStr)}
}


// removeAll path String > err String  
// """
//  Removes the named file or directory and all sub directories and contained files recursively.
// """
//
func removeAll(th InterpreterThread, objects []RObject) []RObject {
	
    path := string(objects[0].(String))

	err := os.RemoveAll(path)
	errStr := ""
	if err != nil {
	   errStr = err.Error()
	}
	return []RObject{String(errStr)}
}


// rename oldName String newName String > err String
// """
//  Renames the file or directory.
//  Valid os filesystem pathnames (relative or absolute) to file/directory are expected as the names.
// """
func rename(th InterpreterThread, objects []RObject) []RObject {
	
    path1 := string(objects[0].(String))
    path2 := string(objects[1].(String))
	err := os.Rename(path1, path2)
	errStr := ""
	if err != nil {
	   errStr = err.Error()
	}
	return []RObject{String(errStr)}
}



// name size mode modTime isDir fileExists err =
//    statPrimitive path
//
func statPrimitive(th InterpreterThread, objects []RObject) []RObject {
	
    path := string(objects[0].(String))

	fi, err := os.Stat(path)
	errStr := ""
	var name String
	var size Int
	var mode Uint32
	var modTime RTime
	var isDir Bool
	var fileExists Bool
	if err == nil {
		fileExists = true
		name = String(fi.Name())
		size = Int(fi.Size())
		mode = Uint32(uint32(fi.Mode()))
        modTime = RTime(fi.ModTime()) 
		isDir = Bool(fi.IsDir())
	} else if os.IsNotExist(err) {
       fileExists = false
	} else { // Some other "can't stat" error - report it out
	   errStr = err.Error()
	}
	return []RObject{name,size,mode,modTime,isDir,fileExists,String(errStr)}
}      


// name size mode modTime isDir fileExists err =
//    lstatPrimitive path
//
func lstatPrimitive(th InterpreterThread, objects []RObject) []RObject {
	
    path := string(objects[0].(String))

	fi, err := os.Lstat(path)
	errStr := ""
	var name String
	var size Int
	var mode Uint32
	var modTime RTime
	var isDir Bool
	var fileExists Bool
	if err == nil {
		fileExists = true
		name = String(fi.Name())
		size = Int(fi.Size())
		mode = Uint32(uint32(fi.Mode()))
        modTime = RTime(fi.ModTime()) 
		isDir = Bool(fi.IsDir())
	} else if os.IsNotExist(err) {
       fileExists = false
	} else { // Some other "can't stat" error - report it out
	   errStr = err.Error()
	}
	return []RObject{name,size,mode,modTime,isDir,fileExists,String(errStr)}
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
	
	}
    var perm os.FileMode
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
		     fmt.Println(permUser, permGroup, permOther)
		     fmt.Println(p)
		     fmt.Println(perm)
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

