// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU LESSER GPL v3 license, found in the LICENSE_LGPL3 file.

package http_methods

/*
   request.go - native methods for http request objects and their attributes, such as uploaded files and cookies.
   These methods are used by types defined in the relish standard library 'http' package. 
*/

import (
	. "relish/runtime/data"
//	"os"
	"io"
	"bufio"
	"net/http"
	"mime/multipart"
)

///////////
// Go Types

/*
 An instance of this type is the wrapped native object referred to by a relish http.UploadedFile instance.
*/
type UploadedFile struct {
   header *multipart.FileHeader
   file multipart.File
}

/*
Ensures that uf.File refers to an open multipart.File which can be read and closed.
Harmless to call this if uf is already open. 
*/
func (uf *UploadedFile) Open() (err error) {
   if uf.file == nil {
   	  uf.file, err = uf.header.Open()
   }
   return
}

func (uf *UploadedFile) Name() string {
	return uf.header.Filename
}

func (uf *UploadedFile) File() multipart.File {
	return uf.file
}


/////////////////////////////////////
// relish method to go method binding

func InitHttpMethods() {

    // print name file 
	nameMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http",nil,"name", []string{"file"}, []string{"shared.relish.pl2012/relish_lib/pkg/http/UploadedFile"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	nameMethod.PrimitiveCode = name


    // err = open file 
	openMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http",nil,"open", []string{"file"}, []string{"shared.relish.pl2012/relish_lib/pkg/http/UploadedFile"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	openMethod.PrimitiveCode = open


    // buf = Bytes 1000
    // n err = read file buf
	readMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http",nil,"read", []string{"file","buf"}, []string{"shared.relish.pl2012/relish_lib/pkg/http/UploadedFile","Bytes"}, []string{"Int","String"}, false, 0, false)
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
	readAllTextMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http",nil,"readAllText", []string{"file"}, []string{"shared.relish.pl2012/relish_lib/pkg/http/UploadedFile"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	readAllTextMethod.PrimitiveCode = readAllText

	readAllTextMethod2, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http",nil,"readAllText", []string{"file","addMissingLinefeed"}, []string{"shared.relish.pl2012/relish_lib/pkg/http/UploadedFile","Bool"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	readAllTextMethod2.PrimitiveCode = readAllText
	
	// readAllBinary
	//    f File 
	// > 
	//    fileContent String err String
	//
	readAllBinaryMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http",nil,"readAllBinary", []string{"file"}, []string{"shared.relish.pl2012/relish_lib/pkg/http/UploadedFile"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	readAllBinaryMethod.PrimitiveCode = readAllBinary


	

    // err = close file
	closeMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http",nil,"close", []string{"file"}, []string{"shared.relish.pl2012/relish_lib/pkg/http/UploadedFile"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	closeMethod.PrimitiveCode = close

















    // http.Request methods

    // First, make sure we have the appropriate list type in existence.
    uploadedFileType := RT.Types["shared.relish.pl2012/relish_lib/pkg/http/UploadedFile"]
    uploadedFileListType, err := RT.GetListType(uploadedFileType) 
	if err != nil {
		panic(err)
	}    

    // files r Request key String > fs [] UploadedFile
	filesMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http",nil,"files", []string{"request","key"},  []string{"shared.relish.pl2012/relish_lib/pkg/http/Request","String"}, []string{uploadedFileListType.Name}, false, 0, false)
	if err != nil {
		panic(err)
	}
	filesMethod.PrimitiveCode = files

    // file r Request key String > f UploadedFile err String
	fileMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http",nil,"file", []string{"request","key"}, []string{"shared.relish.pl2012/relish_lib/pkg/http/Request","String"}, []string{"shared.relish.pl2012/relish_lib/pkg/http/UploadedFile","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	fileMethod.PrimitiveCode = file

    //   TODO
    //
    // cookies r Request > c [] Cookie
    //
    // cookie r Request key String > c Cookie err String








}




 
///////////////////////////////////////////////////////////////////////////////////////////
// I/O functions


// name f UploadedFile  > String
//
func name(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)
	file := wrapper.GoObj.(*UploadedFile)
	return []RObject{String(file.Name())}
}


// open f UploadedFile  > err String
//
func open(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)
	file := wrapper.GoObj.(*UploadedFile)
	err := file.Open()
	errStr := ""
	if err != nil {
	   errStr = err.Error()
	}
	return []RObject{String(errStr)}
}


// read f UploadedFile buf Bytes > n Int err String
//
func read(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)
	buf := objects[1].(Bytes)
	b := ([]byte)(buf)
	uf := wrapper.GoObj.(*UploadedFile)	
	file := uf.File()
	n, err := file.Read(b)
	errStr := ""
	if err != nil {
	   errStr = err.Error()
	}
	return []RObject{Int(n),String(errStr)}
}



// readAllText 
//    f UploadedFile 
//    addMissingLinefeed Bool = false
// > 
//    fileContent String err String
//
func readAllText(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)
	uf := wrapper.GoObj.(*UploadedFile)	
	file := uf.File()	
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
//    f UploadedFile 
// > 
//    fileContent String err String
//
func readAllBinary(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)
	uf := wrapper.GoObj.(*UploadedFile)	
	file := uf.File()	
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
	uf := wrapper.GoObj.(*UploadedFile)	
	file := uf.File()	
	err := file.Close()
	errStr := ""
	if err != nil {
	   errStr = err.Error()
	}
	return []RObject{String(errStr)}
}



///////////////////////////////////////////////////////////////////////////////////////////
// Request Processing functions


//
// files r Request key String > fs [] UploadedFile
//
func files(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)
	request := wrapper.GoObj.(*http.Request)	

    key := string(objects[1].(String))


    uploadedFileType := RT.Types["shared.relish.pl2012/relish_lib/pkg/http/UploadedFile"]

    fileList,err := RT.Newrlist(uploadedFileType,0,-1,nil,nil)
    if err != nil {
    	panic(err)
    }

    if request.MultipartForm != nil && request.MultipartForm.File != nil {
    	fhs := request.MultipartForm.File[key]
    	for _,fh := range fhs {
    		uploadedFile,err := createUploadedFile(fh)
            if err != nil {
    	       panic(err)
            }    		
    		fileList.AddSimple(uploadedFile)
    	}
    }
	return []RObject{fileList}
}






//
// file r Request key String > f UploadedFile err String
//
func file(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)
	request := wrapper.GoObj.(*http.Request)	

    key := string(objects[1].(String))

    var err error

    var uploadedFile RObject = NIL

    if request.MultipartForm != nil && request.MultipartForm.File != nil {
    	fhs := request.MultipartForm.File[key]
        if len(fhs) > 0 {
     		uploadedFile,err = createUploadedFile(fhs[0])       	
        }
    }

    var errStr string
    if err != nil {
    	errStr = err.Error()
    } else if uploadedFile == NIL {
        errStr = "http: no such file"   	
    } 

	return []RObject{uploadedFile, String(errStr)}
}



//   TODO
//
// cookies r Request > c [] Cookie
//
// cookie r Request key String > c Cookie err String





///////////////////////////////////////////////////////////////////////////////////////////
// Type init functions


/*
Construct and initialize an http.UploadedFile object.
*/
func createUploadedFile(fh *multipart.FileHeader) (uf RObject, err error) {
   
    uf,err = RT.NewObject("shared.relish.pl2012/relish_lib/pkg/http/UploadedFile")

    ufWrapper := uf.(*GoWrapper)

    uploadedFile := &UploadedFile{fh,nil}

    ufWrapper.GoObj = uploadedFile

	return 
}




