// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

/*
Creates new relish project directory tree templates.
*/
package main

import (
        "fmt"
        "os"
)


const RELISH_ICON_PNG = `iVBORw0KGgoAAAANSUhEUgAAAAsAAAAQCAYAAADAvYV-AAAA_UlEQVR42mNgwAL8CutaN9zK_chAFGhg-L_-P8N_ohTCFG_4x4lF
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
   err = os.MkdirAll(artifactDir,0777)       
   if err != nil {
      return 
   } 
   mainPackageDir := artifactDir + "/v0.1.0/src/main"
   err = os.MkdirAll(mainPackageDir,0777)       
   if err != nil {
      return 
   } 
   
   // TODO create main.rel file  
   
     
   if projectType == "webapp" {
      // TODO Create web package directory with template index.html and dialog.rel and static/ with styles and an image and a static html file.
      webPackageDir := artifactDir + "/v0.1.0/src/web"
      err = os.MkdirAll(webPackageDir,0777)       
      if err != nil {
         return 
      } 
      stylesDir := webPackageDir + "/static/styles"
      err = os.MkdirAll(stylesDir,0777)       
      if err != nil {
         return 
      }          
      
      // TODO create index.html, dialog.rel,default.css,an image??,a static html file.
   } else if projectType == "" {
      // TODO create example main.rel file with print hello world in it.
   } else {
      err = fmt.Errorf("Unrecognized relish project type '%s'.",projectType)
   }
   return
}