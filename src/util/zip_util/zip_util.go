// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// convenience functions for working with zip files

package zip_util

import (

  "io/ioutil"
    "io"
  "bytes"
    "strings"
    "os"
    "archive/zip" 
    . "relish/dbg"
)


/*
Returns the extracted (decompressed) content of one file that is contained in the zip archive.
The inner file to extract is specified by its relative path name (path name in the zip file).

*/
func ExtractFileFromZipFileContents(zipFileContents []byte, innerFileName string) (innerFileContent []byte, err error) {
   wrapperByteSliceReader := bytes.NewReader(zipFileContents) 

   var rWrapper *zip.Reader
   rWrapper, err = zip.NewReader(wrapperByteSliceReader, int64(len(zipFileContents)))
   if err != nil {
      return
   }

   // Iterate through the files in the wrapper archive,
   // looking for the named file
   for _, fWrapper := range rWrapper.File {

      //fileInfo := fWrapper.FileHeader.FileInfo()  
      //if fileInfo.Name() == innerFileName {
      // replaced with following line to be Go1.2 compatible.
      if fWrapper.FileHeader.Name == innerFileName {


         var zippedFileReader io.ReadCloser
         zippedFileReader, err = fWrapper.Open()
         if err != nil {
           return
         }  
         innerFileContent, err = ioutil.ReadAll(zippedFileReader)
         if err != nil {
            return
         }  
         err = zippedFileReader.Close()
         if err != nil {
            return
         }  
         break
      }
   }
   return
}




/*
Extracts contents of a zip file into the specified directory, which must exist and be writeable.
The directory path must not end with a "/".
Always excludes __MACOSX folders from what it writes to the target directory tree.
*/
func ExtractZipFileContents(zipFileContents []byte, dirPath string) (err error) {

   var perm os.FileMode = 0777

   byteSliceReader := bytes.NewReader(zipFileContents) 

   var r *zip.Reader
   r, err = zip.NewReader(byteSliceReader, int64(len(zipFileContents)))
   if err != nil {
      return
   }


   // Iterate through the files in the archive,
   // copying their contents.
   for _, f := range r.File {

      fileInfo := f.FileHeader.FileInfo()  
      // if strings.Index(fileInfo.Name(),"__MACOSX") == 0 {
      // replaced with following line to be Go1.2 compatible.      
      if strings.Index(f.FileHeader.Name,"__MACOSX") == 0 {

         continue
      }
      if fileInfo.IsDir() {  
          Log(LOAD2_,"Directory %s:\n", f.Name) 
          err = os.MkdirAll(dirPath + "/" + f.Name,perm)
          if err != nil {
             return
          } 
      } else {
          Log(LOAD2_,"Copying file %s:\n", f.Name)
          var rc io.ReadCloser
          rc, err = f.Open()
          if err != nil {
             return
          }

          slashPos := strings.LastIndex(f.Name,"/")
          if slashPos != -1 {
             relativeDirPath := f.Name[:slashPos]
             err = os.MkdirAll(dirPath + "/" + relativeDirPath,perm)
             if err != nil {
                return
             } 
          }

          var outFile *os.File
          outFile, err = os.Create(dirPath + "/" + f.Name)  
          if err != nil {
            return
          }
      
          _, err = io.Copy(outFile, rc)
          if err != nil {
             return
          }
          rc.Close()
          outFile.Close()
          Logln(LOAD2_)
      }
   }
   return
}




