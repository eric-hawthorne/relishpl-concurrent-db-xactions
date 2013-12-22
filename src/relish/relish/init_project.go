// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

/*
Creates new relish project directory tree templates.
*/
package main

import (
        "fmt"
        "os"
        "strings"
        "time"
        "io/ioutil"
        "encoding/base64"
)


const APP_METADATA_FILE = `relish artifact metadata: %s
origin: %s
artifact: %s
current version: 0.1.0
release date: 2099/09/19
tags: application,ADD YOUR TAGS HERE COMMA SEPARATED

Please put a concise description of your software artifact (application or library)
here. A paragraph would do nicely. What is it for? In what context is it to be used?
`


const WEB_APP_METADATA_FILE = `relish artifact metadata: %s
origin: %s
artifact: %s
current version: 0.1.0
release date: 2099/09/19
tags: web application,ADD YOUR TAGS HERE COMMA SEPARATED

Please put a concise description of your web application software artifact 
here. A paragraph would do nicely. What is it for? In what context is it to be used?
`


/*
relish artifact metadata: 2013/09/25
origin: relish.pl2012
artifact: relish_website
current version: 0.1.0
release date: 2012/09/29
tags: web application,relish

The relish.pl website.
*/




const APP_MAIN_PROGRAM_FILE = `origin   %s
artifact %s
package  main

"""
 main.rel
 
 The main program (program entry point) for the %s
 software artifact.
"""


main
"""
 The main method.
 Initializes this and that and the other then prints a start confirmation message on the
 standard output stream. 
"""
   print "\nHello from the %s/%s program.\n"
`


const WEB_APP_MAIN_PROGRAM_FILE = `origin   %s
artifact %s
package  main

"""
 main.rel
 
 The main program (program entry point) for the %s
 web application.
"""


main
"""
 The main method.
 Initializes this and that and the other then prints a start confirmation message on the
 standard output stream. 
"""
   print "\n%s/%s web application started.\n"
`


const INDEX_HTML_FILE = `<html>
  <head>
  <title>Current Local Time</title>

  <link rel="stylesheet" type="text/css" id="stylesheet"
        href="/styles/default.css" />  
  
  </head>
  <body>
      <center> 
        <h1><img src="img/relish_logo_small.png"/> Current Local Time</h1>
        <br/>
        <br/>
        <p>
        It is always handy to know that
        the current time is <span class="time">{{.time}}</span>
        </p>
        <br/>
        <p>
        <a href="example.html">An example static web page</a>
        </p>
     </center>
  </body>
</html>
`

const DIALOG_FILE = `origin   %s
artifact %s
package  web

"""
 dialog.rel
 
 Web application dialog handling methods.
 Typically divided into:
 -action methods which accept and process html form input data or 
  AJAX requests, and
 -dynamic-page methods, which lookup persistent data from the database
 and insert the data into a dynamic web page template to create and serve a web page.
"""


import
   datetime


index > String Map
"""
 Handles a request for the root (i.e. empty path) url on the server and port.
"""
   t = now "Local"
   ts = format t "3:04PM"
   
   => "index.html"
      {
         "time" => ts
      }String > String


default > String String
"""
 Handles all url paths on this server and port which are not otherwise handled.
"""
   => "HTML"
      "<html><body><center><br/><br/><h3>Oops. Nothing here.</h3></center></body></html>"
`

const CSS_FILE = `
body
{
   font-family: sans-serif;
}

span.time
{
   font-size: 30px;
   font-weight: bold;
   color: green;
}
`


const STATIC_HTML_FILE = `<html>
  <head>
  <title>Example Static Html File</title>
  </head>
  <body>
     <h1>Example Static Html File</h1>
     <p>
     No dynamic template content here.
     </p>
     <p>
     Just a plain old html file. In relish, it goes in the web/static directory,
     or in a subdirectory of web/static. But the url to reach the html file that is
     called ~/relish/artifacts/you.com2014/v1.0.0/src/web/static/products/example1.html would be 
     http://you.com/products/example1.html
     </p>
  </body>
</html>
`


const RELISH_LOGO_PNG = `iVBORw0KGgoAAAANSUhEUgAAAAsAAAAQCAYAAADAvYV-AAAA_UlEQVR42mNgwAL8CutaN9zK_chAFGhg-L_-P8N_ohTCFG_4x4lF
Qw2DJsMNIASZNguhGLvpfyCCIg0yYIUwiF3xfyTFSJD9D7ozQCb9QZgg8F30P8Mihv-KHzXBtuidcP8E0rj5p-x_BuUmk38o7qtk
-I1iWD2yk6BW2j7zgihOZFgwbdrFlNnLN-TBnYhsWMCBrPfoEoTDGdkUEJzJ4Igir5Di9JypngViynk0k_9DwhyEs9pnHGeQr7cE
c3Z99IdIQIFJcxhcIVsND8KDIAZTLZoboSbnHykEy--9O_EUAy4PRe-r3gOSS7zrA5RjJJyoKlctso-_HglXCACGvcrTaHgvvAAA
AABJRU5ErkJggg==`


/*
Creates a directory tree for a new relish artifact.
Creates template versions of some of the files you will need.
projectType can currently be "webapp" or ""
*/
func initProject(relishRoot, projectPath, projectType string) (err error) {
   artifactDir := relishRoot + "/artifacts/" + projectPath
   slashPos := strings.Index(projectPath,"/")
   origin := projectPath[0:slashPos]
   artifact := projectPath[slashPos+1:]
   metadataFilePath := artifactDir + "/metadata.txt"
   date := time.Now().Format("2006/01/02")  
   
   err = os.MkdirAll(artifactDir,0777)       
   if err != nil {
      return 
   } 
   mainPackageDir := artifactDir + "/v0.1.0/src/main"
   err = os.MkdirAll(mainPackageDir,0777)       
   if err != nil {
      return 
   } 
   mainFilePath := mainPackageDir + "/main.rel"    
     
   if projectType == "webapp" {
      
      metadata := fmt.Sprintf(WEB_APP_METADATA_FILE,date,origin,artifact)     
      
      err = ioutil.WriteFile(metadataFilePath, ([]byte)(metadata), 0777)  
      if err != nil {
         return 
      }      
      
      // TODO Create web package directory with template index.html and dialog.rel and static/ with styles and an image and a static html file.
      webPackageDir := artifactDir + "/v0.1.0/src/web"
      err = os.MkdirAll(webPackageDir,0777)       
      if err != nil {
         return 
      } 
      
      indexPath := webPackageDir + "/index.html"
      
      err = ioutil.WriteFile(indexPath, ([]byte)(INDEX_HTML_FILE), 0777)  
      if err != nil {
         return 
      }      
      
      dialogPath := webPackageDir + "/dialog.rel"
      
      dialogContent := fmt.Sprintf(DIALOG_FILE,origin,artifact) 
            
      err = ioutil.WriteFile(dialogPath, ([]byte)(dialogContent), 0777)  
      if err != nil {
         return 
      }      
      
      stylesDir := webPackageDir + "/static/styles"
      err = os.MkdirAll(stylesDir,0777)       
      if err != nil {
         return 
      }    
      
      cssPath := stylesDir + "/default.css"
      
      err = ioutil.WriteFile(cssPath, ([]byte)(CSS_FILE), 0777)  
      if err != nil {
         return 
      }      
      
      staticPath := webPackageDir + "/static/example.html"
      
      err = ioutil.WriteFile(staticPath, ([]byte)(STATIC_HTML_FILE), 0777)  
      if err != nil {
         return 
      }      
      
      
      imgDir := webPackageDir + "/static/img"
      err = os.MkdirAll(imgDir,0777)       
      if err != nil {
         return 
      }            
      
      imagePath := imgDir + "/relish_logo_small.png"    
   
      var imageBytes []byte
	   imageBytes,err = base64.URLEncoding.DecodeString(RELISH_LOGO_PNG) 
      if err != nil {
         return 
      }	       
      err = ioutil.WriteFile(imagePath, imageBytes, 0777)  
      if err != nil {
         return 
      }
      
      mainContent := fmt.Sprintf(WEB_APP_MAIN_PROGRAM_FILE,origin,artifact,artifact,origin,artifact) 
 
      err = ioutil.WriteFile(mainFilePath, ([]byte)(mainContent), 0777)  
      if err != nil {
         return 
      }      
      
      // TODO create index.html, dialog.rel,default.css,an image??,a static html file.
   } else if projectType == "" {
      
      metadata := fmt.Sprintf(APP_METADATA_FILE,date,origin,artifact)        
      
      err = ioutil.WriteFile(metadataFilePath, ([]byte)(metadata), 0777)  
      if err != nil {
         return 
      }            
      
      mainContent := fmt.Sprintf(APP_MAIN_PROGRAM_FILE,origin,artifact,artifact,origin,artifact)
      
      err = ioutil.WriteFile(mainFilePath, ([]byte)(mainContent), 0777)  
      if err != nil {
         return 
      }      
   } else {
      err = fmt.Errorf("Unrecognized relish project type '%s'.",projectType)
   }
   return
}