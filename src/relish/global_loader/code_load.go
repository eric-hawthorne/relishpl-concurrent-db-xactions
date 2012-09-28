// Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// loads relish code packages into the runtime, from the local relish artifacts directory tree, 
// or failing that, from the Internet, assuming strong choice-free conventions
// about where a given version of a given artifact will reside or how it will be found, 
// both on the Internet as a whole, and on each server that hosts it.

package global_loader

import (
    "fmt"
//    "net/http"
	"io/ioutil"
//    "regexp"
    "io"
	"bytes"
    "strings"
    "strconv"
    "errors"
    "os"
//	. "relish/runtime/data"
    "relish/compiler/ast"
    "relish/compiler/parser"
    "relish/compiler/token"
	"relish/compiler/generator"
	"relish"
    "archive/zip"	
)

type Loader struct {
	RelishRuntimeLocation string
	LoadedArtifacts map [string]int  // map from originAndArtifactPath to loaded version number	
    LoadedArtifactKnownToBeLocal map[string]bool // was/is being loaded from local artfcts repo
    LoadedArtifactKnownToBeShared map[string]bool // was/is being loaded from shared artfcts repo 


	LoadedPackages map [string]int  // map from originAndArtifactPath/pkg/packagePath to loaded version number
	PackagesBeingLoaded map [string]bool  // map from originAndArtifactPath/pkg/packagePath to whether it is in the middle of loading	
	PackageLocalId map [string] string // map from originAndArtifactPath/pkg/packagePath to local id (short name) of package
	LocalIdPackage map [string] string // map from local id (short name) of package to package full path.
    SharedCodeOnly bool // if true, do not consider the local artifacts dir tree to find any code, only shared/relish/artifacts
    DatabaseName string // name, not full path, of SQLITE database file.
}

//
// Note: Need in each relish database a Packages table with fields id and name.
// Need to load this at program initialization time.
//

/*
Construct and return an artifact loader.
Params:
   relishRuntimeLocation = the directory which is the root of the relish runtime directories; i.e.
   the root of where relish code artifacts can be expected to reside on the computer. Path must end in relish or in relish/rt
   e.g. /opt/relish or /opt/relish/rt
*/
func NewLoader(relishRuntimeLocation string, sharedCodeOnly bool, databaseName string) *Loader {

	ldr := &Loader{relishRuntimeLocation,make(map[string]int),make(map[string]bool),make(map[string]bool),make(map[string]int),make(map[string]bool),make(map[string]string),make(map[string]string), sharedCodeOnly, databaseName}
	return ldr
}



func (ldr *Loader) LoadWebPackages (originAndArtifactPath string, version int, mustBeFromShared bool) (err error) {

    err = ldr.loadPackageTree(originAndArtifactPath, version, "web", mustBeFromShared)
    return
}

/*
Load the specified package and, recursively, all packages found in sub directories of the package src directory.
Currently used only to pre-load the web dialog handler packages.
*/
func (ldr *Loader) loadPackageTree (originAndArtifactPath string, version int, packagePath string, mustBeFromShared bool) (err error) {

    _, err = ldr.LoadPackage(originAndArtifactPath, version, packagePath, mustBeFromShared)
    if err != nil {
	    return
    }
    // Now load web subdir packages.

    if version == 0 {
	   version = ldr.LoadedArtifacts[originAndArtifactPath]
    }

    versionStr := fmt.Sprintf("/v%04d",version)
    artifactVersionDir := ldr.artifactDirPath(originAndArtifactPath) + versionStr

    packageSourcePath := artifactVersionDir + "/src/" + packagePath 

    // Read the filenames of sub directories in the /src/... package directory

    sourceDirFile, err := os.Open(packageSourcePath) 
    if err != nil {
      return
    }
    defer sourceDirFile.Close()

    filenames,err := sourceDirFile.Readdirnames(-1) 
    if err != nil {
	   return
    }

	for _,filename := range filenames {
	
	    if ! strings.Contains(filename,".") { // discard relish source files and other rubbish in the dir.
		
	       // TODO add in here the on-demand compilation as found in relish.go
	       // Doing it NOW!!
	       subDirFilePath := packageSourcePath + "/" + filename	

	       subDirFileInfo,statErr := os.Stat(subDirFilePath)	
	       if statErr != nil {
	           if ! os.IsNotExist(statErr) {
			   	  err = fmt.Errorf("Can't stat relish source directory or file '%s': %v\n", subDirFilePath, statErr)
			  	  return		       
	           }
	       } else if subDirFileInfo.IsDir() {
               packagePath += "/" + filename
                  err = ldr.loadPackageTree (originAndArtifactPath , version, packagePath, mustBeFromShared)	
		       if err != nil {
			      return
		       }	 
           }
       }
    }
    return
}

/*
Given the full path name of a loaded package, return the filesystem path of the directory
which contains the source code files and subdir package directories.
*/
func (ldr *Loader) PackageSrcDirPath(fullPackagePath string) string {
	version := ldr.LoadedPackages[fullPackagePath]
	
	pkgPos := strings.Index(fullPackagePath,"/pkg/")
	originAndArtifactPath := fullPackagePath[:pkgPos]
	packagePath := fullPackagePath[pkgPos + 5:]
	
    versionStr := fmt.Sprintf("/v%04d",version)
    artifactVersionDir := ldr.artifactDirPath(originAndArtifactPath) + versionStr	
    packageSrcPath := artifactVersionDir + "/src/" + packagePath
	return packageSrcPath
}

/*
Given an originAndArtifactPath path segment, return the full filesystem path to the artifact directory
that the artifact was loaded from (either local private artifacts dir or shared artifacts dir).
Requires that at least one package from the artifact has been started to be loaded before this is called.
*/
func (ldr *Loader) artifactDirPath(originAndArtifactPath string) string {
    var artifactsRepoPathSegment string
    if ldr.LoadedArtifactKnownToBeLocal[originAndArtifactPath] { // was loaded from local artfcts repo
       artifactsRepoPathSegment = "/artifacts/"
    } else {
       artifactsRepoPathSegment = "/shared/relish/artifacts/"
    }
 

    return ldr.RelishRuntimeLocation + artifactsRepoPathSegment + originAndArtifactPath
}

/*
Return the database file path to use for this artifact in this run of relish interpreter.
If the artifact was loaded from local (private) artifact dir tree, 
a path like /opt/relish/rt/data/origin/artifact/<databaseName>.db will be used.
otherwise if the artifact was loaded from shared artifact dir tree,
a path like /opt/relish/rt/data_for_shared/origin/artifact/<databaseName>.db will be used.
*/
func (ldr *Loader) databaseDirPath(originAndArtifact string) string {
    var dataPathSegment string
    if ldr.LoadedArtifactKnownToBeLocal[originAndArtifact] { // was loaded from local artfcts repo
       dataPathSegment = "/data/"
    } else {
       dataPathSegment = "/data_for_shared/"
    }
    return ldr.RelishRuntimeLocation + dataPathSegment + originAndArtifact 
}




/*
Functions which load relish abstract syntax trees (intermediate code) into memory.

These functions rely on mandatory code locating conventions in relish.
*/

/*
Loads the code from the specified package into the runtime.

Recurses to load packages which the argument package depends on.

TODO artifact version preferences are NOT being handled correctly yet.
What should happen is that all artifacts whose versions "ARE" recursively preferred should have their version preferences
consulted and respected first (closer to top of load-tree has preference priority) and then and only then should
a single no-version-preferred artifact be loaded by asking what its latest version is, then re-visit if that constrains other
formerly no-version-preferred ones, then load newly preferred versions, then load the next no-version-preferred artifact.
Currently, an artifact version may be loaded in a no-version-preferred way, because the current load-tree descent path does
not prefer a version, but it could be that a subsequently loaded artifact somewhere else in the load-tree COULD express
a preference for a different version of the artifact, but its preference is consulted too late, after the artifact is already loaded.  

Handles searching first in a local (private) artifacts repository (directory tree) then a shared artifacts repository,
but only if a directive is not in effect to load from shared only, and only if another package from 
the same artifact has not already been loaded, because in that case, the local (private) or shared decision has already
been made and must apply to subsequent packages loaded from the same artifact.

TODO MD5 sum integrity checks
*/
func (ldr *Loader) LoadPackage (originAndArtifactPath string, version int, packagePath string, mustBeFromShared bool) (gen *generator.Generator, err error) {

    // First, see if the package is already loaded. If so, return 

    packageIdentifier := originAndArtifactPath + "/pkg/" + packagePath   

    beingLoaded := ldr.PackagesBeingLoaded[packageIdentifier]
    if beingLoaded {
	   err = fmt.Errorf("Package dependency loop. Package '%s' is also present in the tree of packages it imports.",packageIdentifier)	
	   return	
    }

    loadedVersion,found := ldr.LoadedPackages[packageIdentifier]
    if found {
	   if loadedVersion != version && version != 0 {
	       err = fmt.Errorf("Can't load version %d of '%s' since version %d is already loaded into runtime.",version,packageIdentifier,loadedVersion)	
	   }
	   return
    }

    ldr.PackagesBeingLoaded[packageIdentifier] = true

    fmt.Printf("Loading package %s\n",packageIdentifier)

    mustBeFromShared = mustBeFromShared || ldr.SharedCodeOnly  // Set whether will consider local code for this package.
    var mustBeFromLocal bool                 // We may end up constrained to load from local artifact.


    // Package is not loaded. But see if any other packages from the same artifact are loaded.
    // If so, make sure they don't have an incompatible version.

    var artifactAlreadyLoaded bool  // if true, at least one package from the currently-being-loaded artifact has already been loade.
                                    // This means the needed version of the artifact and the artifacts it depends on have already
                                    // been loaded from built.txt into LoadedArtifacts map.

    var artifactKnownToBeLocal bool  // if artifact is loaded or being loaded, is it loaded from local 
    var artifactKnownToBeShared bool // if artifact is loaded or being loaded, is it loaded from shared

    loadedVersion,artifactAlreadyLoaded = ldr.LoadedArtifacts[originAndArtifactPath]
    if artifactAlreadyLoaded {
	   if loadedVersion != version && version != 0 {
	       err = fmt.Errorf("Can't load package '%s' from version %d of '%s'. Another package from version %d of the artifact is already loaded into runtime.",packagePath,version,originAndArtifactPath,loadedVersion)	
	       return
	   }

       artifactKnownToBeLocal = ldr.LoadedArtifactKnownToBeLocal[originAndArtifactPath] 
       artifactKnownToBeShared = ldr.LoadedArtifactKnownToBeShared[originAndArtifactPath]    
       mustBeFromShared = mustBeFromShared || artifactKnownToBeShared  // Set whether will consider local code for this package.         

       fmt.Printf("%s %s mustBeFromShared=%v",originAndArtifactPath,packagePath,mustBeFromShared)
       if artifactKnownToBeLocal {
	       if mustBeFromShared {
	       	   // This should never happen I think. Check anyway.
		       err = fmt.Errorf("Can't load package '%s' from shared artifact '%s'. Another package from the local copy of the artifact is already loaded into runtime.",packagePath,originAndArtifactPath)	
		       return       	
	       } else {
	       	   mustBeFromLocal = true
	       } 
	   }
    }

    // Now try to load the package from local file system. 

    // If allowed, need to try twice, trying to read from
    // local artifacts dir tree, then if not found from shared artifacts dir tree.

    //// If no version has been specified, but some version of the artifact exists in local disk,
    //// the first thing to do is to read the metadata.txt of the local artifact, and set the version number desired
    //// to the current version as specified by the local artifact copy. Note this could be out of date, but we need a
    //// different command to go check if there is a later version of the artifact out there and download it. 

    //// If we can find a metadata.txt file for the artifact locally, set the version with it.


    if version == 0 {

        if ! mustBeFromShared {

		    localArtifactMetadataFilePath := ldr.RelishRuntimeLocation + "/artifacts/" + originAndArtifactPath + "/metadata.txt"	
		
		    _,statErr := os.Stat(localArtifactMetadataFilePath)
		    if statErr != nil {
		        if ! os.IsNotExist(statErr) {
					err = fmt.Errorf("Can't stat '%s': %v\n", localArtifactMetadataFilePath, statErr)
					return		       
		        }

		        // did not find the metadata.txt file in the local (private) artifact dir tree
		        //
		        // so there is no local (private) artifact

		    } else { // found the metadata.txt file in the local (private) artifact dir tree
		     
		        var body []byte
				body, err = ioutil.ReadFile(localArtifactMetadataFilePath)	
				if err != nil {
					return
				}

			    match := re.FindSubmatchIndex(body)
		        if match == nil {
			       err = errors.New("metadata.txt file must have a line like current version: 14")
			       return
		        }
			    versionNumStart := match[2]
			    versionNumEnd := match[3]	
		    
			    s := string(body[versionNumStart:versionNumEnd])
		 
		        var v64 int64
		        v64, err = strconv.ParseInt(s, 0, 32)  
			    if err != nil {
				   return
			    }
		
			    version = int(v64)	
		
		    }	
		}

        if version == 0 {

        	if mustBeFromLocal {
        	   // We already loaded a package from the local artifact, but somehow the local artifact is no longer there on filesystem.	
	       	   // This should never happen if everything is being loaded at once at beginning of run. Check anyway.
		       err = fmt.Errorf("Can't load package '%s' from local artifact '%s'. Local artifact not found.",packagePath,originAndArtifactPath)	
		       return               		
        	}

		    sharedArtifactMetadataFilePath := ldr.RelishRuntimeLocation + "/shared/relish/artifacts/" + originAndArtifactPath + "/metadata.txt"	
		
		    _,statErr := os.Stat(sharedArtifactMetadataFilePath)
		    if statErr != nil {
		        if ! os.IsNotExist(statErr) {
					err = fmt.Errorf("Can't stat '%s': %v\n", sharedArtifactMetadataFilePath, statErr)
					return		       
		        }

		        // did not find the metadata.txt file in the shared artifact dir tree
		        //
		        // so there is no lshared artifact in the filesystem.

		    } else { // found the metadata.txt file in the shared artifact dir tree
		     
		        var body []byte
				body, err = ioutil.ReadFile(sharedArtifactMetadataFilePath)	
				if err != nil {
					return
				}

			    match := re.FindSubmatchIndex(body)
		        if match == nil {
			       err = errors.New("metadata.txt file must have a line like current version: 14")
			       return
		        }
			    versionNumStart := match[2]
			    versionNumEnd := match[3]	
		    
			    s := string(body[versionNumStart:versionNumEnd])
		 
		        var v64 int64
		        v64, err = strconv.ParseInt(s, 0, 32)  
			    if err != nil {
				   return
			    }
		
			    version = int(v64)	
		
		    }	
		}
	}




    // stat the artifact version dir to see if the version of the artifact exists in the filesystem.
    //
    // try local then shared artifacts dir trees as allowed by constraints so far

    artifactVersionDirFound := false
    var artifactVersionDir string


    if version > 0 {
	    versionStr := fmt.Sprintf("/v%04d",version)

        if ! mustBeFromShared {

            // try local artifacts dir tree
		    artifactVersionDir = ldr.RelishRuntimeLocation + "/artifacts/" + originAndArtifactPath + versionStr

		    _,statErr := os.Stat(artifactVersionDir)
		    if statErr != nil {
		        if ! os.IsNotExist(statErr) {
					err = fmt.Errorf("Can't stat relish artifact version directory '%s': %v\n", artifactVersionDir, statErr)
					return		       
		        }


		    } else {
		        artifactVersionDirFound = true
		        mustBeFromLocal = true  // locked to local (private) artifact now
		    }
	    }

	    if ! artifactVersionDirFound {

           // this version not found in local artifacts dir tree
	       // try shared artifacts dir tree 
	       artifactVersionDir = ldr.RelishRuntimeLocation + "/shared/relish/artifacts/" + originAndArtifactPath + versionStr

	       _,statErr := os.Stat(artifactVersionDir)
	       if statErr != nil {
	           if ! os.IsNotExist(statErr) {
			   	   err = fmt.Errorf("Can't stat relish artifact version directory '%s': %v\n", artifactVersionDir, statErr)
				   return		       
	           }
	        } else {
	            artifactVersionDirFound = true
	            mustBeFromShared = true  // locked to shared artifact now
	        }
	    }
    } 
    

    if ! artifactVersionDirFound {

        // TODO Need this path in order to install or update the artifact metadata file from remote, if there is none locally
        // or if the remote one is more recent.
        //
	    // artifactMetadataFilePath := ldr.RelishRuntimeLocation + "/shared/relish/artifacts/" + originAndArtifactPath + "/metadata.txt"	

	    // Have not found the artifact version locally. Fetch it from the Internet.

	    // Note: We will always be fetching into the shared artifacts directory tree.
	    // If programmer wants to copy an artifact version into the local artifacts directory tree to develop/modify it,
	    // they must currently do that copy separately manually.

	    var zipFileContents []byte

	    hostURL := ldr.DefaultCodeHost(originAndArtifactPath)

	    zipFileContents, err = fetchArtifactZipFile(hostURL, originAndArtifactPath, version) 
	    if err != nil {
		    var hostURLs []string
		    hostURLs,err = ldr.FindSecondaryCodeHosts(originAndArtifactPath, hostURL)
		    if err != nil {
			   return 
		    }
		    for _,hostURL = range hostURLs {
		        zipFileContents, err = fetchArtifactZipFile(hostURL, originAndArtifactPath, version) 
		        if err == nil {
			       break
			    }     
			    // consider logging the missed fetch and or developing a bad reputation for the host.
		    }
	    }

        // TODO TODO TODO MAYBE UP HIGHER A BIT 
        // if we did not have the metadata file on filesystem before, or if remote metadata file is newer,
        // we should download and cache the metadata.txt file from the remote repository.
        // 
        // Then, if we do not have a specified version yet, we should set version # from that,







	    if zipFileContents == nil {
		   err = fmt.Errorf("Search of Internet did not find relish software artifact '%s'",originAndArtifactPath)
		   return
	    }

       // TODO Unzip the artifact into the proper local directory tree


   
////////////////////////////////////////////////////////////////////////

	   // TODO Unzip the artifact into the proper local directory tree

	    // TODO TODO Really don't know the artifact version here in some case, (in case there was nothing
	    // not even a metadata.txt file locally, and no version was specified on command line) so
	    // we don't have the correct path for artifactVersionDir known yet in that case !!!
	    // WE DO KNOW IT HAS TO BE A SHARED ARTIFACTS DIR PATH however.




	   //os.MkdirAll(name string, perm FileMode) error
	   var perm os.FileMode = 0777
	   err = os.MkdirAll(artifactVersionDir, perm)
	   if err != nil {
	      return
	   }

	   // Open a zip archive for reading.
	
	   // Note: Assuming the zip file starts with src/ pkg/ doc/ etc not with v0002/

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
	      if strings.Index(fileInfo.Name(),"__MACOSX") == 0 {
		     continue
	      }
	      if fileInfo.IsDir() {  
		      fmt.Printf("Directory %s:\n", f.Name)	
		      err = os.MkdirAll(artifactVersionDir + "/" + f.Name,perm)
	          if err != nil {
	             return
	          }	
		  } else {
		      fmt.Printf("Copying file %s:\n", f.Name)
		      var rc io.ReadCloser
		      rc, err = f.Open()
		      if err != nil {
		         return
		      }

              var outFile *os.File
	          outFile, err = os.Create(artifactVersionDir + "/" + f.Name)  
	          if err != nil {
	            return
	          }
          
		      _, err = io.Copy(outFile, rc)
		      if err != nil {
		         return
		      }
		      rc.Close()
		      outFile.Close()
		      fmt.Println()
	      }
	   }
	}
		
    if ! artifactAlreadyLoaded { // Read built.txt from the artifact version directory
        builtFilePath := artifactVersionDir + "/built.txt"
	    var builtFileContents []byte

		_,statErr := os.Stat(builtFilePath)

		if statErr == nil {

			builtFileContents, err = ioutil.ReadFile(builtFilePath)	
			if err != nil {
				return
			}
			
	        artifactsVersionsStrs := strings.Fields(string(builtFileContents))
	        n := len(artifactsVersionsStrs)
	        for i := 0; i < n; i += 2 {
		       artifactPath := artifactsVersionsStrs[i] 
		       artifactVersionStr := artifactsVersionsStrs[i+1] 
		       var v64 int64
		       v64, err = strconv.ParseInt(artifactVersionStr, 0, 32)  
			   if err != nil {
				   return
			   }
			   artifactVersion := int(v64)	
			   alreadyDesiredVersion,versionFound := ldr.LoadedArtifacts[artifactPath]	
			   if versionFound {
				   if artifactVersion != alreadyDesiredVersion {
				      // Should be logging this, not printing it to stdout
				      fmt.Printf("Using v%d of %s. %s (v%d) may prefer v%d of %s.\n",alreadyDesiredVersion, artifactPath, originAndArtifactPath, version, artifactVersion, artifactPath)
			       }
			   } else {
			      ldr.LoadedArtifacts[artifactPath] = artifactVersion	
			   }
	        }		

        } else if ! os.IsNotExist(statErr) {
			err = fmt.Errorf("Can't stat '%s': %v\n", builtFilePath, statErr)
			return		       
	    }        
	}	

////////////////////////////////////////////////////////////////////////

	// TODO
	// Load all the code files in the package of the artifact. 
	// compile files if necessary?
	//
	// record the version of the artifact and package in the loader's registry of loaded packages and artifacts.

    packageSourcePath := artifactVersionDir + "/src/" + packagePath 
    packageCompiledPath := artifactVersionDir + "/pkg/" + packagePath 

    // Read the filenames of source files etc. in the /src/... package directory


    sourceDirFile, err := os.Open(packageSourcePath) 
    if err != nil {
      return
    }
    defer sourceDirFile.Close()

    filenames,err := sourceDirFile.Readdirnames(-1) 
    if err != nil {
	   return
    }



    // Create /pkg/ dir tree if does not exist already
    _,statErr := os.Stat(packageCompiledPath)
     if statErr != nil {
         if ! os.IsNotExist(statErr) {
			 err = fmt.Errorf("Can't stat relish intermediate-code directory '%s': %v\n", packageCompiledPath, statErr)
			 return		       
         } 
	     var perm os.FileMode = 0777
	     err = os.MkdirAll(packageCompiledPath, perm)
	     if err != nil {
	        return
         }
     }



    ldr.LoadedArtifacts[originAndArtifactPath] = version

	ldr.LoadedArtifactKnownToBeShared[originAndArtifactPath] = strings.Contains(artifactVersionDir,"/shared/relish/artifacts/")   
    fmt.Printf("ldr.LoadedArtifactKnownToBeShared[%s]=%v\n",originAndArtifactPath,ldr.LoadedArtifactKnownToBeShared[originAndArtifactPath])
    fmt.Println("artifactVersionDir="+artifactVersionDir)


    ldr.LoadedArtifactKnownToBeLocal[originAndArtifactPath] = ! ldr.LoadedArtifactKnownToBeShared[originAndArtifactPath]	

    
    if relish.DatabaseURI() == "" {
       dbDirPath := ldr.databaseDirPath(originAndArtifactPath) 

	   var perm os.FileMode = 0777
	   err = os.MkdirAll(dbDirPath, perm)
	   if err != nil {
	      return
	   }

       dbFilePath := dbDirPath + "/" + ldr.DatabaseName
       relish.SetDatabaseURI(dbFilePath)   // TODO NOT TRUE AT ALL YET 
                                           // This can be overridden with a statement in the program, as long as a persistence op has not been used first.
    }

    for _,filename := range filenames {
		var sourceFound bool
		var pickledFound bool	
	    if strings.HasSuffix(filename,".rel") { // consider only the relish source files in the dir.
			
           // TODO add in here the on-demand compilation as found in relish.go
           // Doing it NOW!!
	       sourceFilePath := packageSourcePath + "/" + filename	
	
	       fileNameRoot := filename[:len(filename)-4]
	
           pickleFilePath := packageCompiledPath + "/" + fileNameRoot + ".rlc"



           sourceFileInfo,statErr := os.Stat(sourceFilePath)	
           if statErr != nil {
	           if ! os.IsNotExist(statErr) {
			   	  err = fmt.Errorf("Can't stat relish source file '%s': %v\n", sourceFilePath, statErr)
			  	  return		       
	           }
           } else {
              sourceFound = true
           }

           pickleFileInfo,statErr := os.Stat(pickleFilePath)
           if statErr != nil {
	           if ! os.IsNotExist(statErr) {
				 err = fmt.Errorf("Can't stat relish intermediate-code file '%s': %v\n", pickleFilePath, statErr)
				 return		       
	           } 
           } else {
              pickledFound = true
           }

           var parseNeeded bool
           if sourceFound {
	          if pickledFound {
		         if sourceFileInfo.ModTime().After(pickleFileInfo.ModTime()) {
			        parseNeeded = true
		         }
	          } else {
		         parseNeeded = true
	          }
           } else if ! pickledFound {
			  err = fmt.Errorf("Error: Found neither relish source file '%s' nor intermediate-code file '%s'.\n", sourceFilePath, pickleFilePath)
			  return	
           }

           var fileNode *ast.File
           if parseNeeded {	
              var fset = token.NewFileSet()	
		  	  fileNode, err = parser.ParseFile(fset, sourceFilePath, nil, parser.DeclarationErrors | parser.Trace)
			  if err != nil {
				 err = fmt.Errorf("Error parsing file '%s': %v\n", sourceFilePath, err)
				 return
			  }

	          err = ast.Pickle(fileNode, pickleFilePath) 	
			  if err != nil {
				 err = fmt.Errorf("Error pickling file '%s': %v\n", sourceFilePath, err)
				 return
			  }
		   } else { // read the pickled (intermediate-code) file

	          fileNode, err = ast.Unpickle(pickleFilePath) 	
			  if err != nil {
				 err = fmt.Errorf("Error unpickling file '%s': %v\n", pickleFilePath, err)
				 return
			  }					
		   }


 
		
		   err = ldr.ensureImportsAreLoaded(fileNode)
	       if err != nil {
	          return
	       }		
		
           gen = generator.NewGenerator(fileNode, fileNameRoot) // TODO NOW add a isLocal =ldr.LoadedArtifactKnownToBeLocal[originAndArtifactPath]
                                                                // argument so that we can flag the RPackage object as local or shared.
           gen.GenerateCode()	

	       if parseNeeded {
		      fmt.Printf("Compiled %s\n", sourceFilePath)		
		   } else {
		      fmt.Printf("Loaded %s\n", pickleFilePath)	
	       }
	    }
    }

    ldr.LoadedPackages[packageIdentifier] = version



    delete(ldr.PackagesBeingLoaded,packageIdentifier)

    return
}
     
	


/*
Finds hosts that host the artifact.
Returns an empty list if there are no such hosts.
Returns an error if the search service cannot be reached or does not return a valid result page.
*/
func (ldr *Loader) DefaultCodeHost (originAndArtifactPath string) (hostURL string) {
	hostURL = "http://" + originAndArtifactPath[:strings.Index(originAndArtifactPath,"/")-4]
	return hostURL
}
	
	
/*
Finds hosts that host the artifact.
Returns an empty list if there are no such hosts.
Returns an error if the search service cannot be reached or does not return a valid result page.
TODO
*/
func (ldr *Loader) FindSecondaryCodeHosts (originAndArtifactPath string, primaryHostURL string) (hostURLs []string, err error) {
	return hostURLs, nil
}


/*
Return the version of the artifact that is to be loaded. Can return 0 (no preference)
*/	
func (ldr *Loader) ArtifactVersion(originAndArtifactName string) int {
	return ldr.LoadedArtifacts[originAndArtifactName]
}	

	
/*
Check the imports list of the relish intermediate-code file and load the packages if not already loaded.
Requires consultation of the current package's artifact's built.txt information on which version of each other artifact must be loaded.
*/	
func (ldr *Loader) ensureImportsAreLoaded(fileNode *ast.File) (err error) {
	imports := fileNode.RelishImports  // package specifications
	for _,importedPackageSpec := range imports {
		
		importedArtifactVersion := ldr.ArtifactVersion(importedPackageSpec.OriginAndArtifactName)	
		
		_,err = ldr.LoadPackage(importedPackageSpec.OriginAndArtifactName,
			                  importedArtifactVersion,
		                      importedPackageSpec.PackageName, 
		                      false)		
        if err != nil {
	       if importedArtifactVersion == 0 {
		      fmt.Printf("Error loading package %s from current version of %s:  %v\n", importedPackageSpec.PackageName,importedPackageSpec.OriginAndArtifactName, err)		
		   } else {
		      fmt.Printf("Error loading package %s from version %d of %s:  %v\n", importedPackageSpec.PackageName,importedArtifactVersion,importedPackageSpec.OriginAndArtifactName, err)		
	       }
		   break	
       }			
	} 
	return
}	