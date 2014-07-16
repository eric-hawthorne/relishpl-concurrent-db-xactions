// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU LESSER GPL v3 license, found in the LICENSE_LGPL3 file.

package http_methods

/*
   request.go - native methods for http request objects and their attributes, such as uploaded files and cookies.
   These methods are used by types defined in the relish standard library 'http_srv' package. 
*/

import (
	. "relish/runtime/data"
	"io"
	"bufio"
	"net/http"
	"mime/multipart"
	"fmt"
)

///////////
// Go Types

/*
 An instance of this type is the wrapped native object referred to by a relish http_srv.UploadedFile instance.
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
	nameMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"name", []string{"file"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/UploadedFile"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	nameMethod.PrimitiveCode = name


    // err = open file 
	openMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"open", []string{"file"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/UploadedFile"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	openMethod.PrimitiveCode = open


    // buf = Bytes 1000
    // n err = read file buf
	readMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"read", []string{"file","buf"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/UploadedFile","Bytes"}, []string{"Int","String"}, false, 0, false)
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
	readAllTextMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"readAllText", []string{"file"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/UploadedFile"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	readAllTextMethod.PrimitiveCode = readAllText

	readAllTextMethod2, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"readAllText", []string{"file","addMissingLinefeed"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/UploadedFile","Bool"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	readAllTextMethod2.PrimitiveCode = readAllText
	
	// readAllBinary
	//    f File 
	// > 
	//    fileContent String err String
	//
	readAllBinaryMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"readAllBinary", []string{"file"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/UploadedFile"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	readAllBinaryMethod.PrimitiveCode = readAllBinary


	

    // err = close file
	closeMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"close", []string{"file"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/UploadedFile"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	closeMethod.PrimitiveCode = close

















    // http_srv.Request methods

    // First, make sure we have the appropriate list type in existence.
    uploadedFileType := RT.Types["shared.relish.pl2012/relish_lib/pkg/http_srv/UploadedFile"]
    uploadedFileListType, err := RT.GetListType(uploadedFileType) 
	if err != nil {
		panic(err)
	}    

    // First, make sure we have the appropriate list type in existence.
    cookieType := RT.Types["shared.relish.pl2012/relish_lib/pkg/http_srv/Cookie"]
    cookieListType, err := RT.GetListType(cookieType) 
	if err != nil {
		panic(err)
	}  	

    // files r Request key String > fs [] UploadedFile
	uploadedFilesMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"uploadedFiles", []string{"request","key"},  []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/Request","String"}, []string{uploadedFileListType.Name}, false, 0, false)
	if err != nil {
		panic(err)
	}
	uploadedFilesMethod.PrimitiveCode = uploadedFiles

    // file r Request key String > f UploadedFile err String
	uploadedFileMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"uploadedFile", []string{"request","key"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/Request","String"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/UploadedFile","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	uploadedFileMethod.PrimitiveCode = uploadedFile

    //   TODO
    //
    // cookies r Request > c [] Cookie
	cookiesMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"cookies", []string{"request"},  []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/Request"}, []string{cookieListType.Name}, false, 0, false)
	if err != nil {
		panic(err)
	}
	cookiesMethod.PrimitiveCode = cookies

    // cookie r Request name String > c Cookie err String
	cookieMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"cookie", []string{"request","name"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/Request","String"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/Cookie","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	cookieMethod.PrimitiveCode = cookie


	// contentLength r Request > Int 
	//
	contentLengthMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"contentLength", []string{"request"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/Request"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	contentLengthMethod.PrimitiveCode = contentLength	
	

	// requestUri r Request > String 
	//
	requestUriMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"requestUri", []string{"request"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/Request"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	requestUriMethod.PrimitiveCode = requestUri	


	// referer r Request > String 
	//
	// Note: Misspelling of "referrer" inherited from http_srv standard.
	//
	refererMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"referer", []string{"request"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/Request"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	refererMethod.PrimitiveCode = referer	
	

	// method r Request > String 
	//
	// GET POST PUT
	//
	methodMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"method", []string{"request"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/Request"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	methodMethod.PrimitiveCode = method	
	
	
	// host r Request > String 
	//
	// host or host:port
	//
	hostMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"host", []string{"request"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/Request"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	hostMethod.PrimitiveCode = host
	
		

	// remoteAddr r Request > String 
	//
	// The client address
	//
	// IP:port
	//
	remoteAddrMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"remoteAddr", []string{"request"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/Request"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	remoteAddrMethod.PrimitiveCode = remoteAddr




    // http_srv.Cookie methods

    // See golang.org/pkg/net/http/#Cookie for attribute meanings

    // name c Cookie > String
    //
	cookieNameMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"name", []string{"c"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/Cookie"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	cookieNameMethod.PrimitiveCode = cookieName

    // value c Cookie > String
	//
	cookieValueMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"value", []string{"c"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/Cookie"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	cookieValueMethod.PrimitiveCode = cookieValue

    // path c Cookie > String
    //
	cookiePathMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"path", []string{"c"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/Cookie"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	cookiePathMethod.PrimitiveCode = cookiePath

    // domain c Cookie > String
    //
	cookieDomainMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"domain", []string{"c"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/Cookie"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	cookieDomainMethod.PrimitiveCode = cookieDomain 

    // expires c Cookie > Time 
    //
	cookieExpiresMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"expires", []string{"c"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/Cookie"}, []string{"Time"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	cookieExpiresMethod.PrimitiveCode = cookieExpires

    // rawExpires c Cookie > String
    //
	cookieRawExpiresMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"rawExpires", []string{"c"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/Cookie"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	cookieRawExpiresMethod.PrimitiveCode = cookieRawExpires	

	// // maxAge=0 means no 'Max-Age' attribute specified.
	// // maxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
	// // maxAge>0 means Max-Age attribute present and given in seconds
	// maxAge c Cookie > Int

	cookieMaxAgeMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"maxAge", []string{"c"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/Cookie"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	cookieMaxAgeMethod.PrimitiveCode = cookieMaxAge

    // secure c Cookie > Bool
	//
	cookieSecureMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"secure", []string{"c"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/Cookie"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	cookieSecureMethod.PrimitiveCode = cookieSecure

    // httpOnly c Cookie > Bool
    //
	cookieHttpOnlyMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"httpOnly", []string{"c"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/Cookie"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	cookieHttpOnlyMethod.PrimitiveCode = cookieHttpOnly


    // raw c Cookie > String      
    //
	cookieRawMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"raw", []string{"c"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/Cookie"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	cookieRawMethod.PrimitiveCode = cookieRaw

    // initString s String c Cookie > String  
    //
	initStringFromCookieMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"initString", []string{"s","c"}, []string{"String","shared.relish.pl2012/relish_lib/pkg/http_srv/Cookie"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	initStringFromCookieMethod.PrimitiveCode = initStringFromCookie

    // setCookie c Cookie > String    
	// """
	//  Return a proper header line for setting the cookie in an http response
	// """
    //
	cookieSetMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"setCookie", []string{"c"}, []string{"shared.relish.pl2012/relish_lib/pkg/http_srv/Cookie"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	cookieSetMethod.PrimitiveCode = cookieSetHeader

    // setCookie name String value String maxAge Int path String domain String > String
    //
	cookieSet2Method, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"setCookie", []string{"name","value","maxAge","path","domain"}, []string{"String","String","Int","String","String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	cookieSet2Method.PrimitiveCode = cookieSetHeader2


    // header h1 String > String
    //
	headerMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"header", []string{"h1"}, []string{"String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	headerMethod.PrimitiveCode = header

    // headers h1 String > String
    //
	headers1Method, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"headers", []string{"h1"}, []string{"String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	headers1Method.PrimitiveCode = headers

    // headers h1 String h2 String > String
    //
	headers2Method, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"headers", []string{"h1","h2"}, []string{"String","String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	headers2Method.PrimitiveCode = headers

    // headers h1 String h2 String h3 String > String
    //
	headers3Method, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"headers", []string{"h1","h2","h3"}, []string{"String","String","String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	headers3Method.PrimitiveCode = headers

    // headers h1 String h2 String h3 String h4 String > String
    //
	headers4Method, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"headers", []string{"h1","h2","h3","h4"}, []string{"String","String","String","String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	headers4Method.PrimitiveCode = headers

    // headers h1 String h2 String h3 String h4 String h5 String > String
    //
	headers5Method, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"headers", []string{"h1","h2","h3","h4","h5"}, []string{"String","String","String","String","String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	headers5Method.PrimitiveCode = headers

    // headers h1 String h2 String h3 String h4 String h5 String h6 String > String
    //
	headers6Method, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"headers", []string{"h1","h2","h3","h4","h5","h6"}, []string{"String","String","String","String","String","String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	headers6Method.PrimitiveCode = headers

    // headers h1 String h2 String h3 String h4 String h5 String h6 String h7 String > String
    //
	headers7Method, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"headers", []string{"h1","h2","h3","h4","h5","h6","h7"}, []string{"String","String","String","String","String","String","String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	headers7Method.PrimitiveCode = headers

    // headers h1 String h2 String h3 String h4 String h5 String h6 String h7 String h8 String > String
    //
	headers8Method, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/http_srv",nil,"headers", []string{"h1","h2","h3","h4","h5","h6","h7","h8"}, []string{"String","String","String","String","String","String","String","String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	headers8Method.PrimitiveCode = headers						
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
// uploadedFiles r Request key String > fs [] UploadedFile
//
func uploadedFiles(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)
	request := wrapper.GoObj.(*http.Request)	

    key := string(objects[1].(String))


    uploadedFileType := RT.Types["shared.relish.pl2012/relish_lib/pkg/http_srv/UploadedFile"]

    fileList,err := RT.Newrlist(uploadedFileType,0,-1,nil,nil,nil)
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
// uploadedFile r Request key String > f UploadedFile err String
//
func uploadedFile(th InterpreterThread, objects []RObject) []RObject {
	
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
        errStr = "http_srv: no such file"   	
    } 

	return []RObject{uploadedFile, String(errStr)}
}



//   TODO
//
// cookies r Request > c [] Cookie

func cookies(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)	
	request := wrapper.GoObj.(*http.Request)
		

    cookieType := RT.Types["shared.relish.pl2012/relish_lib/pkg/http_srv/Cookie"]

    cookieList,err := RT.Newrlist(cookieType,0,-1,nil,nil,nil)
    if err != nil {
    	panic(err)
    }

    cookies := request.Cookies()

    var cookieObj RObject

	for _,cookie := range cookies {
       cookieObj, err = createCookie(cookie)
       if err != nil {
	       panic(err)
       }    		
	   cookieList.AddSimple(cookieObj)
    }
	return []RObject{cookieList}
}

// cookie r Request name String > c Cookie err String

func cookie(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)	
	request := wrapper.GoObj.(*http.Request)

    name := string(objects[1].(String))		

    var cookieObj RObject = NIL

    cookie, err := request.Cookie(name) 
    if err == nil {
        cookieObj, err = createCookie(cookie)
    }

    var errStr string
    if err != nil {
    	errStr = err.Error()
    } else if cookieObj == NIL {
        errStr = "http_srv: no such cookie"   	
    } 

	return []RObject{cookieObj, String(errStr)}    
}



// contentLength r Request > Int 
//
func contentLength(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)	
	request := wrapper.GoObj.(*http.Request)
		
	return []RObject{Int(request.ContentLength)}
}

// requestUri r Request > String 
//
func requestUri(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)	
	request := wrapper.GoObj.(*http.Request)
		
	return []RObject{String(request.RequestURI)}
}

// referer r Request > String 
//
// Note: Misspelling of "referrer" inherited from http standard.
//
func referer(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)	
	request := wrapper.GoObj.(*http.Request)
		
	return []RObject{String(request.Referer())}
}




// method r Request > String 
//
// GET POST PUT
//
func method(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)	
	request := wrapper.GoObj.(*http.Request)
		
	return []RObject{String(request.Method)}
}


// host r Request > String 
//
// host or host:port
//
func host(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)	
	request := wrapper.GoObj.(*http.Request)
		
	return []RObject{String(request.Host)}
}


// remoteAddr r Request > String 
//
// The client address
//
// IP:port
//
func remoteAddr(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)	
	request := wrapper.GoObj.(*http.Request)
		
	return []RObject{String(request.RemoteAddr)}
}





///////////////////////////////////////////////////////////////////////////////////////////
// Cookie functions

// See golang.org/pkg/net/http/#Cookie for attribute meanings

// name c Cookie > String
//
func cookieName(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)	
	cookie := wrapper.GoObj.(*http.Cookie)
		
	return []RObject{String(cookie.Name)}
}


// value c Cookie > String
//
func cookieValue(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)	
	cookie := wrapper.GoObj.(*http.Cookie)
		
	return []RObject{String(cookie.Value)}
}


// path c Cookie > String
//
func cookiePath(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)	
	cookie := wrapper.GoObj.(*http.Cookie)
		
	return []RObject{String(cookie.Path)}
}


// domain c Cookie > String
//
func cookieDomain(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)	
	cookie := wrapper.GoObj.(*http.Cookie)
		
	return []RObject{String(cookie.Domain)}
}


// expires c Cookie > Time 
//
func cookieExpires(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)	
	cookie := wrapper.GoObj.(*http.Cookie)
		
	return []RObject{RTime(cookie.Expires)}
}


// rawExpires c Cookie > String
//
func cookieRawExpires(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)	
	cookie := wrapper.GoObj.(*http.Cookie)
		
	return []RObject{String(cookie.RawExpires)}
}


// // maxAge=0 means no 'Max-Age' attribute specified.
// // maxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
// // maxAge>0 means Max-Age attribute present and given in seconds
// maxAge c Cookie > Int
//
func cookieMaxAge(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)	
	cookie := wrapper.GoObj.(*http.Cookie)
		
	return []RObject{Int(cookie.MaxAge)}
}


// secure c Cookie > Bool
//
func cookieSecure(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)	
	cookie := wrapper.GoObj.(*http.Cookie)
		
	return []RObject{Bool(cookie.Secure)}
}


// httpOnly c Cookie > Bool
//
func cookieHttpOnly(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)	
	cookie := wrapper.GoObj.(*http.Cookie)
		
	return []RObject{Bool(cookie.HttpOnly)}
}


// raw c Cookie > String      
//
func cookieRaw(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)	
	cookie := wrapper.GoObj.(*http.Cookie)
		
	return []RObject{String(cookie.Raw)}
}


// initString s String c Cookie > String  
//
func initStringFromCookie(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[1].(*GoWrapper)	
	cookie := wrapper.GoObj.(*http.Cookie)
		
	return []RObject{String(cookie.String())}
}

// setCookie c Cookie > String
// """
//  Return a proper header line for setting the cookie in an http response
// """
func cookieSetHeader(th InterpreterThread, objects []RObject) []RObject {
	
	wrapper := objects[0].(*GoWrapper)	
	cookie := wrapper.GoObj.(*http.Cookie)
		
	return []RObject{String("Set-Cookie: " + cookie.String())}
}

// setCookie name String value String maxAge Int path String domain String > String
// """
//  Return a proper header line for a cookie in an http response.
//  The maxAge is in seconds.
//
//  the returned string can be used in the HEADERS directive returned from a relish web dialog method. 
//  This is most easily facilitated by using the header or headers methods of the http_srv package.
//  Example:
//  
//     => headers 
//           setCookie "SESS" "A1YJ243E2G6FSL8" 3600 "/" "example.com"
//           NO_CACHE
//        "foo.html"
//        fooDialogArgs
//
//  If maxAge < 0, no max-age attribute is included in the cookie.
//  If path is "", it is not included in the cookie, and the cookie works only for the exact path
//  of the http request that the http response is for.
//  If domain is "", it is not included in the cookie, and whatever domain is responding to the 
//  request becomes the effective cookie domain on the client.
//  
//
// """
func cookieSetHeader2(th InterpreterThread, objects []RObject) []RObject {
	name := string(objects[0].(String))
	value := string(objects[1].(String))
	maxAge := int(int64(objects[2].(Int)))
	domain := string(objects[3].(String))
	path := string(objects[4].(String))				
    
	cookie := &http.Cookie{ Name: name,
		                    Value: value,
	                      }
    if path != "" {
    	cookie.Path = path
    }
    if domain != "" {
       cookie.Domain = domain
    }
    if maxAge >= 0 {
       cookie.MaxAge = maxAge
	}
	return []RObject{String("Set-Cookie: " + cookie.String())}
}

// header h1 String > String
// """
//  Return a HEADERS argument with one header.
//  The result of this call can be returned as the first return-value of a relish web dialog method
//  to set the header in the http response.
// """
func header(th InterpreterThread, objects []RObject) []RObject {
	
	h1 := string(objects[0].(String))	
    h := `HEADERS
%s
`
    h = fmt.Sprintf(h,h1)

	return []RObject{String(h)}
}


// headers 
//    h1 String 
//    h2 String
//    h3 String
//    hN String
// > 
//    String
// """
//  Return a HEADERS argument with the specified header lines
//  The result of this call can be returned as the first return-value of a relish web dialog method
//  to set the headers in the http response.
// """
func headers(th InterpreterThread, objects []RObject) []RObject {
	
	h := "HEADERS\n"
	for _,arg := range objects {
	   h += "%s\n"
	   hi := string(arg.(String))	
       h = fmt.Sprintf(h,hi)
    }

	return []RObject{String(h)}
}




///////////////////////////////////////////////////////////////////////////////////////////
// Type init functions


/*
Construct and initialize an http_srv.UploadedFile object.
*/
func createUploadedFile(fh *multipart.FileHeader) (uf RObject, err error) {
   
    uf,err = RT.NewObject("shared.relish.pl2012/relish_lib/pkg/http_srv/UploadedFile")
    if err != nil {
    	return
    }
    ufWrapper := uf.(*GoWrapper)

    uploadedFile := &UploadedFile{fh,nil}

    ufWrapper.GoObj = uploadedFile

	return 
}


/*
Construct and initialize an http_srv.Cookie object.
*/
func createCookie(cookie *http.Cookie) (cookieObj RObject, err error) {
   
    cookieObj,err = RT.NewObject("shared.relish.pl2012/relish_lib/pkg/http_srv/Cookie")
    if err != nil {
    	return
    }
    cookieWrapper := cookieObj.(*GoWrapper)
    cookieWrapper.GoObj = cookie

	return 
}


