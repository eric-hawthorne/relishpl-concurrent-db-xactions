// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// For each relish code artifact in shared/relish/artifacts, creates a zip file of each version 
// of the artifact if such a zip file does not already exist.

package global_publisher

import (
    "fmt"
//    "net/http"
	"io/ioutil"
//    "regexp"
    "io"
	"bytes"
    "strings"
//    "strconv"
    "os"
//	. "relish/runtime/data"
//    "relish/compiler/ast"
//    "relish/compiler/parser"
//    "relish/compiler/token"
//	"relish/compiler/generator"
//	"relish"
    "archive/zip"	
//    . "relish/dbg"

    "crypto/sha256"
    "encoding/base64"  
    "util/crypto_util"  
    "errors"
)

/*
Remove the existing shared copy of the version of the artifact (in a failsafe manner).
Copy source code for the version of the package to the shared sourcecode tree.
Create a zip file of the version of the artifact. 
Wrap that with another zip file which will eventually contain a certificate 
of the origin's public key, and a signature of (the SHA256 hash of) the inner zip file contents.
Name the outer zip file by an unambiguous name corresponding to the version of the artifact from the origin.
Check to see if there is a metadata file in the shared directory for the artifact.
If no metadata file in shared, add it. Metadata file name should again be fully qualified unambiguous name,
to ensure findability by Google search. Or the content should contain a standard findable title like
"Relish Artifact Metadata".
Note that the metadata.txt file should eventually also include,
below its meaningful text content, a certificate of the origin's public key, and a signature of (the SHA256 hash of)
the meaningful content.
Should copy local artifact version directory tree that was published to create the next version to continue work on.

Note about signed code plans
============================
The plan is to have every copy of relish include the public key of shared.relish.pl.
This key can be directly compared manually at any time with the public key published at the site shared.relish.pl.

Each code-origin owner who wants to officially publish relish code should register their origin at shared.relish.pl
and should receive back both 
1) a certificate (signed by shared.relish.pl and verifiable with shared.relish.pl's public key)
   where such certificate attests to the association between another public key and the origin name.
2) The actual origin public key
3) A corresponding private key.

The origin owner should keep their private key secret but install it in a standard place
within their relish development environment so that their relish instance can use it to sign code.

The origin owner should also keep their certificate of their public key installed at a standard place
in their relish development environment, so they can include that cert in signed-code outer zip files.
The certified public key should actually be published at a standard place in the canonical server
for the origin, and also at shared.relish.pl. This is so that such a key can periodically be verified
as being a currently valid one.

Note: If the private key is stolen, someone else can produce code signed as coming from your origin,
so maybe the same password used to sign up to register your origin should serve as a decryption key
for a symmetrically encrypted version of the private key. So you would be prompted to enter the password
when publishing (and signing) some code.

If not using encrypted private key, then if it is discovered that the key has been stolen (e.g. some code
fraudulently claiming to be from the origin is discovered), then a solution would be to apply for a 
new public key for the origin, and re-publish all legitimate code signed by the new key.

Periodically, a imported-code-using instance of relish should re-verify (at the canonical server or shared.relish.pl)
that the public key that is signing each of their imported artifacts is still the valid public key for the origin.

Perhaps shared.relish.pl can contain a timestamp of when the most recent re-keying incident happened,
and relish instances can periodically check that to see if re-verification of origin public-keys, and
possible redownloading of re-signed code artifacts is necessary. 

*/
func PublishSourceCode(relishRoot string, originAndArtifact string, version string) (err error) {

   slashPos := strings.Index(originAndArtifact,"/")
   originId := originAndArtifact[:slashPos]

   // Obtain the private key for the origin that is publishing.

   originPrivateKey, err := crypto_util.GetPrivateKey("origin", originId) 
   if err != nil {
      return
   }

   // Obtain the public key certificate for the origin that is publishing.

   originPublicKeyCertificate, err := crypto_util.GetPublicKeyCert("origin", originId)  
   if err != nil {
      return
   }

   // Obtain the public key certificate of shared.relish.pl2012

   sharedRelishPublicKeyCertificate, err := crypto_util.GetPublicKeyCert("origin", "shared.relish.pl2012")  
   if err != nil {
      return
   }

   // Validate that it is signed properly, obtaining the shared.relish.pl2012 publicKeyPEM.

   sharedRelishPublicKey := crypto_util.VerifiedPublicKey("", sharedRelishPublicKeyCertificate, "origin", "shared.relish.pl2012") 

    if sharedRelishPublicKey == "" {
        err = errors.New("Invalid shared.relish.pl2012 public key certificate.")
        return
    }

   // Do a quick validation of publishing origin's public key cert

   originPublicKey := crypto_util.VerifiedPublicKey(sharedRelishPublicKey, originPublicKeyCertificate, "origin", originId) 

    if originPublicKey == "" {
        err = errors.New("Invalid " + originId + " public key certificate.")
        return
    }



    // prompt for the publishing origin's private key password.

    var buf *bufio.Reader = bufio.NewReader(os.Stdin)
    fmt.Print("Enter code-origin administration password:")
    input,err := buf.ReadString('\n')
    if err != nil {
       return
    }
   originPrivateKeyPassword := input[:len(input)-1]





   localArtifactPath := relishRoot + "/artifacts/" + originAndArtifact + "/"
   sharedArtifactPath := relishRoot + "/shared/relish/artifacts/" + originAndArtifact + "/"

   // Check if metadata.txt file exists in shared. If not, create it by copying local metadata.txt file

   sharedMetadataPath := sharedArtifactPath + "metadata.txt"
   _,err = os.Stat(sharedMetadataPath)    
   if err != nil {
        if os.IsNotExist(err) {

           err = os.MkdirAll(sharedArtifactPath,0777)       
           if err != nil {
              fmt.Printf("Error making shared artifact directory %s: %s\n", sharedArtifactPath,err)
              return 
           }  

           localMetadataPath := localArtifactPath + "metadata.txt"
           var content []byte
           content, err = ioutil.ReadFile(localMetadataPath)
           if err != nil {
             return
           }
           err = ioutil.WriteFile(sharedMetadataPath, content, 0666)
           if err != nil {
              return
           }              
        } else {    
           fmt.Printf("Can't stat  '%s' : %v\n", sharedArtifactPath + "metadata.txt", err)
           return              
        }
    } 

   // Note. This does not update the shared metadata.txt file from the local if the shared metadata.txt
   // file already existed.

   //

   versionPath := "v" + version

   sharedArtifactVersionPath := sharedArtifactPath + versionPath 

   _,err = os.Stat(sharedArtifactVersionPath)    
   foundVersionShared := false
   if err != nil {
        if ! os.IsNotExist(err) {
           fmt.Printf("Can't stat directory '%s' : %v\n", sharedArtifactVersionPath, err)
           return              
        }
    } else {
       foundVersionShared = true
    } 
    if foundVersionShared {
        fmt.Printf("%s already exists. Cannot republish the same version.\n",sharedArtifactVersionPath)
        return 
    }

    localSrcDirPath := localArtifactPath + versionPath + "/src"  
    srcDirPath := sharedArtifactVersionPath + "/src"

    // mkdir the version of the shared directory

    err = os.MkdirAll(sharedArtifactVersionPath,0777)

    if err != nil {
        fmt.Printf("Error making shared artifact version directory %s: %s\n", sharedArtifactVersionPath,err)
        return 
    }  

    // Copy source code directory tree to "shared/relish/artifacts" tree root.

    err = copySrcDirTree(localSrcDirPath, srcDirPath)
    if err != nil {
        fmt.Printf("Error copying local src dir to create %s: %s\n", srcDirPath,err)
        return 
    }   
    // TBD

    // Zip the source !


    srcZipFilePath := sharedArtifactVersionPath + "/src.zip"
    err = zipSrcDirTree(srcDirPath,srcZipFilePath)
    if err != nil {
        fmt.Printf("Error zipping %s: %s\n", srcDirPath,err)
        return 
    }
    

    // Now have to sign it and put into an outer zip file.

    err = signZippedSrc(srcZipFilePath, originPrivateKey, originPrivateKeyPassword, originPublicKeyCertificate, sharedArtifactPath,originAndArtifact,version)
    if err != nil {
        fmt.Printf("Error signing %s: %s\n", srcZipFilePath,err)
        return 
    } 

    err = os.Remove(srcZipFilePath)
    if err != nil {
       fmt.Printf("Error removing %s: %s\n", srcZipFilePath,err)
       return 
    }
    return
}


func copySrcDirTree(fromSrcDirPath string, toSrcDirPath string) (err error) {
   
   var dir *os.File
   var filesInDir []os.FileInfo
   dir, err = os.Open(fromSrcDirPath)
   filesInDir, err = dir.Readdir(0)
   if err != nil {
     return
   }
   err = dir.Close()

   err = os.Mkdir(toSrcDirPath,0777)

   for _, fileInfo := range filesInDir {
        fromItemPath := fromSrcDirPath + "/" + fileInfo.Name()
        toItemPath := toSrcDirPath + "/" + fileInfo.Name()    
        if fileInfo.IsDir() {
           err = copySrcDirTree(fromItemPath, toItemPath)
           if err != nil {
              return
           }
        } else { // plain old file to be copied.
           if strings.HasSuffix(fileInfo.Name(), ".rel") {
              var content []byte
              content, err = ioutil.ReadFile(fromItemPath)
              if err != nil {
                 return
              }
              err = ioutil.WriteFile(toItemPath, content, 0666)
              if err != nil {
                 return
              }              
           }
        }
    }
    return
}














/*
  Given a zip file of the source code directory tree, 
  1. Computes the SHA256 hash of the source code zip file contents, then signs the hash using
  the private key of the origin.
  2. Adds 
     a. the certificate of the origin's public key (including that public key), and
     b. the signature of the source zip file (which can be verified with that public key)
     c. the source zip file
     to an outer (wrapper) zip file that it is creating.
  3. Writes the wrapper zip file as e.g. a.b.com2013--my_artifact_name--1.0.3.zip to the 
     shared artifact's root directory.

  NOTE: STEPS 1 and 2. a. b. are TBD !!!! Just re-zips the src.zip file presently.   
*/
func signZippedSrc(srcZipPath string, originPrivateKey string,originPrivateKeyPassword string, originPublicKeyCertificate string, sharedArtifactPath string, originAndArtifact string, version string) (err error) {
   originAndArtifactFilenamePart := strings.Replace(originAndArtifact, "/","--",-1)
   wrapperFilename := originAndArtifactFilenamePart + "---" + version + ".zip"
   wrapperFilePath := sharedArtifactPath + "/" + wrapperFilename
 
    var srcZipContents []byte
    srcZipContents, err = ioutil.ReadFile(srcZipPath) 
    if err != nil {
        return
    }

@@@@@@@@@@@@ Use crypto_util method here. Also, prepend the outer zip file name before signing!!!!!
    hasher := sha256.New()
    hasher.Write(srcZipContents)
    sha := hasher.Sum(nil)
    b64 := base64.URLEncoding.EncodeToString(sha) 

    signature := b64 // Temporary - not actually creating a signature here yet
    // signature = sign(sha, originPrivateKey, originPrivateKeyPassword)


    var buf *bytes.Buffer 
    buf, err = signZippedSrc1(srcZipPath, originPublicKeyCertificate, signature)


    var file *os.File
    file, err = os.Create(wrapperFilePath)
    if err != nil {
        return
    }

    _, err = buf.WriteTo(file)
    if err != nil {
        return
    }
    err = file.Close();   

    return
}

/*
   Helper. 
*/
func signZippedSrc1(srcZipPath string, originPublicKeyCertificate string, signature string) (buf *bytes.Buffer, err error) {

   buf = new(bytes.Buffer)

   // Create a new zip archive.
   w := zip.NewWriter(buf)

   err = signZippedSrc2(w, srcZipPath, originPublicKeyCertificate, signature)
   if err != nil {
      return
   }    

   err = w.Close()

   return
}   

/*
   Helper. Write the wrapper zip file using the zip.Writer.
   
Are the cert and the signature actually []byte arguments????
*/
func signZippedSrc2(w *zip.Writer, srcZipPath string, originPublicKeyCertificate string, signature string) (err error) {

   var zw io.Writer
   zw, err = w.Create("src.zip")
   if err != nil {
      return
   }   

   var f *os.File
   f,err = os.Open(srcZipPath)
   if err != nil {
      return
   }            
   _, err = io.Copy(zw, f)
   err = f.Close()   
   if err != nil {
      return
   }      

   zw, err = w.Create("certifiedOriginPublicKey.txt")
   if err != nil {
      return
   }
   _, err = zw.Write([]byte(originPublicKeyCertificate))
   if err != nil {
      return
   }

   zw, err = w.Create("signatureOfSrcZip.txt")
   if err != nil {
      return
   }
   _, err = zw.Write([]byte(signature))
   if err != nil {
      return
   }   

   return
}   





/*
Zips the specified directory tree of relish source code files into the specified zip file.
*/
func zipSrcDirTree(directoryPath string, zipFilePath string) (err error) {

    var buf *bytes.Buffer 
    buf, err = zipSrcDirTree1(directoryPath)

    var file *os.File
	file, err = os.Create(zipFilePath)
    if err != nil {
        return
    }

    _, err = buf.WriteTo(file)
    if err != nil {
        return
    }
    err = file.Close();

    return
}

/*
   Helper. Zip the contents of a directory tree into the byte buffer, which is returned.
   
   Filters so it only includes .rel files

   Note: this will not work if there are symbolic links in the src directory tree.
   (because Readdir does not follow links.)
*/
func zipSrcDirTree1(directoryPath string) (buf *bytes.Buffer, err error) {

   var rootDirFileInfo os.FileInfo
   rootDirFileInfo, err = os.Stat(directoryPath)
   if err != nil {
       return
   }
   if ! rootDirFileInfo.IsDir() {
      err = fmt.Errorf("%s is not a directory.", directoryPath)
      return
   }

   buf = new(bytes.Buffer)

   // Create a new zip archive.
   w := zip.NewWriter(buf)

   err = zipSrcDirTree2(w, directoryPath, rootDirFileInfo.Name())  // "/opt/relish/rt/artifacts/a.com2013/art1/v0001/src"  "src"
   if err != nil {
      return
   }    

   err = w.Close()

   return
}   

/*
   Helper. Recursively zip the contents of a directory tree using the zip.Writer.
   
   Filters so it only includes .rel files

   Note: this will not work if there are symbolic links in the src directory tree.
   (because Readdir does not follow links.)
*/
func zipSrcDirTree2(w *zip.Writer, directoryPath string, relativeDirName string) (err error) {

   var dir *os.File
   var filesInDir []os.FileInfo
   dir, err = os.Open(directoryPath)
   filesInDir, err = dir.Readdir(0)
   if err != nil {
     return
   }
   err = dir.Close()

   for _, fileInfo := range filesInDir {
        if fileInfo.IsDir() {
           subItemPath := directoryPath + "/" + fileInfo.Name()    
           subItemRelativePath := relativeDirName + "/" + fileInfo.Name()               
           err = zipSrcDirTree2(w, subItemPath, subItemRelativePath)
           if err != nil {
              return
           }
        } else { // plain old file to be added.
           if strings.HasSuffix(fileInfo.Name(), ".rel") {
              subItemPath := directoryPath + "/" + fileInfo.Name()    
              subItemRelativePath := relativeDirName + "/" + fileInfo.Name()   

              var fh *zip.FileHeader
              fh, err = zip.FileInfoHeader(fileInfo)
              if err != nil {
                 return
              }                
              fh.Name = subItemRelativePath

              var zw io.Writer
              zw, err = w.CreateHeader( fh )
              if err != nil {
                 return
              }    
              var f *os.File
              f,err = os.Open(subItemPath)
              if err != nil {
                 return
              }
            
              _, err = io.Copy(zw, f)
              err = f.Close()   
              if err != nil {
                 return
              }      
           }
        }
    }

    return
}   




