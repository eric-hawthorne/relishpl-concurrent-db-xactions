// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
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
    "strings"
    "bufio"
//    "strconv"
    "os"
//	. "relish/runtime/data"
    "relish/compiler/ast"
    "relish/compiler/parser"
    "relish/compiler/token"
	"relish/compiler/generator"
    "relish/runtime/native_methods"
	"relish"	
    "util/zip_util"
    "util/crypto_util"    
    . "relish/dbg"
    "errors"
)


const STANDARD_SOURCE_CODE_SHARING_PORT = "8421"  // relish source code may be shared on port 80 or 8421


type Loader struct {
	RelishRuntimeLocation string
	LoadedArtifacts map [string]string  // map from originAndArtifactPath to loaded version number
    LoadedArtifactKnownToBeLocal map[string]bool // was/is being loaded from local artfcts repo
    LoadedArtifactKnownToBePublished map[string]bool // was/is being loaded from shared artfcts repo 
    LoadedArtifactKnownToBeReplica map[string]bool // was/is imported and was/is being loaded from shared replicas repo 


	LoadedPackages map [string]string  // map from originAndArtifactPath/pkg/packagePath to loaded version number
	PackagesBeingLoaded map [string]bool  // map from originAndArtifactPath/pkg/packagePath to whether it is in the middle of loading	
	PackageLocalId map [string] string // map from originAndArtifactPath/pkg/packagePath to local id (short name) of package
	LocalIdPackage map [string] string // map from local id (short name) of package to package full path.
    SharedCodeOnly bool // if true, do not consider the local artifacts dir tree to find any code, only shared/relish/artifacts
    DatabaseName string // name, not full path, of SQLITE database file.

    nonSearchedOriginHostUrls map[string][]string  // cached list of urls of hosts where code from an origin may be found
                                                   // this list comes from configuration files in the rt/xtras directory
                                                   // and from the standard host-for-origin naming convention.
    
    stagingServerUrls map[string][]string  
    originUrls map[string][]string  
    replicaUrls map[string][]string  
    repositoryUrls []string  
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

	ldr := &Loader{relishRuntimeLocation,make(map[string]string),make(map[string]bool),
                   make(map[string]bool),make(map[string]bool),make(map[string]string),
                   make(map[string]bool),make(map[string]string),make(map[string]string), 
                   sharedCodeOnly, databaseName, make(map[string][]string),
                   make(map[string][]string),make(map[string][]string),make(map[string][]string),nil}

    ldr.initCodeLocations()
	return ldr
}




/*
If serving a webapp, the relish packages in the web package directory tree of the running artifact must all be loaded
into the runtime. 
*/
func (ldr *Loader) LoadWebPackages (originAndArtifactPath string, version string, mustBeFromShared bool) (err error) {

    err = ldr.loadPackageTree(originAndArtifactPath, version, "web", mustBeFromShared)
    return
}

/*
Load the specified package and, recursively, all packages found in sub directories of the package src directory.
Currently used only to pre-load the web dialog handler packages.
version may be "" which means use the current version of the artifact.
*/
func (ldr *Loader) loadPackageTree (originAndArtifactPath string, version string, packagePath string, mustBeFromShared bool) (err error) {

    _, err = ldr.LoadPackage(originAndArtifactPath, version, packagePath, mustBeFromShared)
    if err != nil {
	    return
    }
    // Now load web subdir packages.

    if version == "" {
	   version = ldr.LoadedArtifacts[originAndArtifactPath]
    }

    artifactVersionDir := ldr.artifactDirPath(originAndArtifactPath) + "/v" + version

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

    parentPackagePath := packagePath
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
	
               packagePath = parentPackagePath + "/" + filename
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
	
    versionStr := "/v" + version
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
    } else if ldr.LoadedArtifactKnownToBeReplica[originAndArtifactPath] {
       artifactsRepoPathSegment = "/shared/relish/replicas/"	
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


var parserDebugMode uint = parser.DeclarationErrors






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

If not found locally, tries to load from the Internet (at several standard locations).

TODO signed-code integrity checks
*/
func (ldr *Loader) LoadPackage (originAndArtifactPath string, version string, packagePath string, mustBeFromShared bool) (gen *generator.Generator, err error) {

	if Logging(PARSE_) {
	   parserDebugMode |= parser.Trace
	}

    // First, see if the package is already loaded. If so, return 

    packageIdentifier := originAndArtifactPath + "/pkg/" + packagePath   

    beingLoaded := ldr.PackagesBeingLoaded[packageIdentifier]
    if beingLoaded {
	   err = fmt.Errorf("Package dependency loop. Package '%s' is also present in the tree of packages it imports.",packageIdentifier)	
	   return	
    }

    loadedVersion,found := ldr.LoadedPackages[packageIdentifier]
    if found {
	   if loadedVersion != version && version != "" {
	       err = fmt.Errorf("Can't load version %s of '%s' since version %s is already loaded into runtime.",version,packageIdentifier,loadedVersion)	
	   }
	   return
    }


    localArtifactMetadataFilePath := ldr.RelishRuntimeLocation + "/artifacts/" + originAndArtifactPath + "/metadata.txt"	
	
    sharedArtifactMetadataFilePath := ldr.RelishRuntimeLocation + "/shared/relish/artifacts/" + originAndArtifactPath + "/metadata.txt"	
    
    sharedReplicaMetadataFilePath := ldr.RelishRuntimeLocation + "/shared/relish/replicas/" + originAndArtifactPath + "/metadata.txt"	

    // Current version of artifact according to shared artifact metadata found in this relish directory tree.
    sharedCurrentVersion := ""

    // Date of the artifact metadata that is being relied on for the package load.
    metadataDate := ""

    ldr.PackagesBeingLoaded[packageIdentifier] = true

    Log(ALWAYS_,"Loading package %s\n",packageIdentifier)

    mustBeFromShared = mustBeFromShared || ldr.SharedCodeOnly  // Set whether will consider local code for this package.
    var mustBeFromLocal bool                 // We may end up constrained to load from local artifact.


    // Package is not loaded. But see if any other packages from the same artifact are loaded.
    // If so, make sure they don't have an incompatible version.

    var artifactAlreadyLoaded bool  // if true, at least one package from the currently-being-loaded artifact has already been loade.
                                    // This means the needed version of the artifact and the artifacts it depends on have already
                                    // been loaded from built.txt into LoadedArtifacts map.

    var artifactKnownToBeLocal bool  // if artifact is loaded or being loaded, is it loaded from local 
    var artifactKnownToBePublished bool // if artifact is loaded or being loaded, is it loaded from shared
    var artifactKnownToBeReplica bool // if artifact is loaded or being loaded, is it downloaded and loaded from shared/replicas

    loadedVersion,artifactAlreadyLoaded = ldr.LoadedArtifacts[originAndArtifactPath]
    if artifactAlreadyLoaded {
	   if loadedVersion != version && version != "" {
	       err = fmt.Errorf("Can't load package '%s' from version %s of '%s'. Another package from version %s of the artifact is already loaded into runtime.",packagePath,version,originAndArtifactPath,loadedVersion)	
	       return
	   }

       artifactKnownToBeLocal = ldr.LoadedArtifactKnownToBeLocal[originAndArtifactPath] 
       artifactKnownToBePublished = ldr.LoadedArtifactKnownToBePublished[originAndArtifactPath]    
       artifactKnownToBeReplica = ldr.LoadedArtifactKnownToBeReplica[originAndArtifactPath]  
       mustBeFromShared = mustBeFromShared || artifactKnownToBePublished || artifactKnownToBeReplica // Set whether will consider local code for this package.         

       Log(LOAD2_,"%s %s mustBeFromShared=%v\n",originAndArtifactPath,packagePath,mustBeFromShared)
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


    if version == "" {

        if ! mustBeFromShared {
		
            version, metadataDate, err = ldr.readMetadataFile(localArtifactMetadataFilePath) 	
		    if err != nil {
			   return
		    }
		}

        if version == "" {

        	if mustBeFromLocal {
        	   // We already loaded a package from the local artifact, but somehow the local artifact is no longer there on filesystem.	
	       	   // This should never happen if everything is being loaded at once at beginning of run. Check anyway.
		       err = fmt.Errorf("Can't load package '%s' from local artifact '%s'. Local artifact not found.",packagePath,originAndArtifactPath)	
		       return               		
        	}

            version, metadataDate, err = ldr.readMetadataFile(sharedArtifactMetadataFilePath) 	
		    if err != nil {
			   return
		    }		
		    if version == "" {
	            version, metadataDate, err = ldr.readMetadataFile(sharedReplicaMetadataFilePath) 	
			    if err != nil {
				   return
			    }	
			    if version != "" {
				   artifactKnownToBeReplica = true
				}		
			} else {
				artifactKnownToBePublished = true
			}
			sharedCurrentVersion = version				
		}
	}




    // stat the artifact version dir to see if the version of the artifact exists in the filesystem.
    //
    // try local then shared artifacts and replicas dir trees as allowed by constraints so far

    artifactVersionDirFound := false
    var artifactVersionDir string


    if version != "" {
	    versionStr := "/v" + version

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
		        artifactKnownToBeLocal = true
		    }
	    }

	    if ! artifactVersionDirFound {

           // this version not found in local artifacts dir tree
	       // try published shared artifacts dir tree 
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
	            artifactKnownToBePublished = true
	        }
	    }
	
	    if ! artifactVersionDirFound {

           // this version not found in local artifacts dir tree or published shared artifacts dir tree
	       // try downloaded shared replicas dir tree 
	       artifactVersionDir = ldr.RelishRuntimeLocation + "/shared/relish/replicas/" + originAndArtifactPath + versionStr

	       _,statErr := os.Stat(artifactVersionDir)
	       if statErr != nil {
	           if ! os.IsNotExist(statErr) {
			   	   err = fmt.Errorf("Can't stat relish artifact version directory '%s': %v\n", artifactVersionDir, statErr)
				   return		       
	           }
	        } else {
	            artifactVersionDirFound = true
	            mustBeFromShared = true  // locked to shared artifact now
	            artifactKnownToBeReplica = true
	        }
	    }	
	
    } 
    

    if ! artifactVersionDirFound {

	    // Have not found the artifact version locally. Fetch it from the Internet.
	
	
	
        // TODO Need this path in order to install or update the artifact metadata file from remote, if there is none locally
        // or if the remote one is more recent.
        //
	    // artifactMetadataFilePath := ldr.RelishRuntimeLocation + "/shared/relish/artifacts/" + originAndArtifactPath + "/metadata.txt"	



	    // Note: We will always be fetching into the shared artifacts directory tree.
	    // If programmer wants to copy an artifact version into the local artifacts directory tree to develop/modify it,
	    // they must currently do that copy separately manually.

	    var zipFileContents []byte

        if sharedCurrentVersion == "" {
           sharedCurrentVersion,metadataDate, err = ldr.readMetadataFile(sharedReplicaMetadataFilePath)
        }

        // replaced by stuff below
        // defaultHostURL := ldr.DefaultCodeHost(originAndArtifactPath)
        // hostURL := defaultHostURL        
        //
        // REVIEW THIS COMMENT 
        // Note: TODO The correct order to do things is to load the metadata.txt file from the default host
        // (if possible) then to search for secondary hosts to get the version zip file, selecting
        // one AT RANDOM, 
        // then if all of those (some number of) mirrors fail, get it from the default host
        // Also, use port 80 then 8421. 

        // TODO: RE: SERVER SEARCH ORDER 
        // We really ought to consider using a different order of servers tried 
        // for fetching artifact source-code zip files than the order we use for trying to find
        // the smaller artifact metadata.txt files. Specifically, it is better to find
        // metadata.txt files at servers owned or controlled by the origin, because the metadata.txt
        // file will be up to date. shared.relish.pl is next best for that consideration.
        // However, from a performance (load sharing when scaled) perspective, it is better to 
        // download the actual source code zip files from randomly found replica servers or
        // secondary general repositories.

        hostURLs := ldr.NonSearchedCodeHostURLs(originAndArtifactPath,"")   // NEW

        // If we did not have the metadata file on filesystem before, or if remote metadata file is newer,
        // we should download and cache the metadata.txt file from the remote repository.
        // 
        // Then, if we do not have a specified version yet, we should set version # from that,
    
        var currentVersion string

        var usingCentralRepo bool = false  // whether getting code from http://shared.relish.pl

        var hostURL string

        for _,hostURL = range hostURLs {

	       // Read remote metadata file. Store it locally if its date is >= the local shared artifact metadata date.
	       currentVersion, err = fetchArtifactMetadata(hostURL, originAndArtifactPath, metadataDate, sharedCurrentVersion, sharedReplicaMetadataFilePath) 	
	       if err == nil {
               break
           } else if currentVersion != "" {
               return  // Inability to create or write local metadata file. Bad error.
           }
        }
        if currentVersion == "" {
	
            // TODO Now try google search 
    
            // Should really do a Google search for metadata found anywhere (except shared.relish.pl) now,
            // so as to limit load and single point of failure on shared.relish.pl.
/*
		    hostURLs,err = ldr.FindSecondaryCodeHosts(originAndArtifactPath, hostURLs)
		    if err != nil {
			   return 
		    }
	        for _,hostURL = range hostURLs {

		       // Read remote metadata file. Store it locally if its date is >= the local shared artifact metadata date.
		       currentVersion, err = fetchArtifactMetadata(hostURL, originAndArtifactPath, metadataDate, sharedCurrentVersion, sharedReplicaMetadataFilePath) 	
		       if err == nil {
	               break
	           } else if currentVersion != "" {
	               return  // Inability to create or write local metadata file. Bad error.
	           }
	        }	
*/		
            //
            // Only if trying other replica sites fails should shared.relish.pl be tried.
            //
            // However, for now, we're going straight to trying shared.relish.pl, because Google searching and
            // signed-metadata verification and signed-code verification aren't implemented yet.
        }

        if currentVersion == "" {

            // Now try shared.relish.pl 

            usingCentralRepo = true
            hostURL = "http://shared.relish.pl"        

            // Read remote metadata file. Store it locally if its date is >= the local shared artifact metadata date.
            currentVersion, err = fetchArtifactMetadata(hostURL, originAndArtifactPath, metadataDate, sharedCurrentVersion, sharedReplicaMetadataFilePath)  
            if err != nil {
               return  // We really couldn't find and download metadata for this artifact anywhere we looked. Too bad.
            }    
        }

        // metadataHostURL := hostURL  // If we need to keep track of where we got the metadata from.

        if usingCentralRepo {
            hostURLs = []string{hostURL}
        } else {
            hostURLs = ldr.NonSearchedCodeHostURLs (originAndArtifactPath, hostURL)  // NEW	    
	    }

		if version == "" {
			version = currentVersion
		}
		
        // Version must now be a proper version number string, not ""
		
		var zipFileName string

        for _,hostURL = range hostURLs {
	        zipFileContents, zipFileName, err = fetchArtifactZipFile(hostURL, originAndArtifactPath, version) 
	        if err == nil {
		       break
		    }     
		    // TODO consider logging the missed fetch and or developing a bad reputation for the host.
	    }

	    if zipFileContents == nil {
		   err = fmt.Errorf("Search of Internet did not find relish software artifact '%s'",originAndArtifactPath)
		   return
	    }
	
	    artifactKnownToBeReplica = true

        // Unzip the artifact into the proper local directory tree

	    // TODO TODO Really don't know the artifact version here in some case, (in case there was nothing
	    // not even a metadata.txt file locally, and no version was specified on command line) so
	    // we don't have the correct path for artifactVersionDir known yet in that case !!!
	    // WE DO KNOW IT HAS TO BE A SHARED REPLICASE DIR PATH however.

	   versionStr := "/v" +version
	   artifactVersionDir = ldr.RelishRuntimeLocation + "/shared/relish/replicas/" + originAndArtifactPath + versionStr

	   //os.MkdirAll(name string, perm FileMode) error
	   var perm os.FileMode = 0777
	   err = os.MkdirAll(artifactVersionDir, perm)
	   if err != nil {
	      return
	   }

       zipFilePath := ldr.RelishRuntimeLocation + "/shared/relish/replicas/" + originAndArtifactPath + "/" + zipFileName
       err = ioutil.WriteFile(zipFilePath, zipFileContents, perm)
       if err != nil {
          return
       }

	   // Open an artifact version zip archive for reading.
	
       var srcZipFileContents []byte
       srcZipFileContents, err = zip_util.ExtractFileFromZipFileContents(zipFileContents, "artifactVersionContents.zip") 
	   if err != nil {
	      return
	   }

       /////////////////////////////////////////////////////////////////////////////////////
       // Verify the signed contents using digital signature verification.

       var sharedRelishPublicKeyCertificateBytes []byte
       var sharedRelishPublicKeyCertificate string
       var installationSharedRelishPublicKeyCert string       
       var originPublicKeyCertificateBytes []byte
       var originPublicKeyCertificate string
       var signatureBytes []byte
       var signature string

       sharedRelishPublicKeyCertificateBytes, err = zip_util.ExtractFileFromZipFileContents(zipFileContents, "sharedRelishPublicKeyCertificate.pem") 
       if err != nil {
          return
       }

       sharedRelishPublicKeyCertificate = strings.TrimSpace(string(sharedRelishPublicKeyCertificateBytes))

       originPublicKeyCertificateBytes, err = zip_util.ExtractFileFromZipFileContents(zipFileContents, "originPublicKeyCertificate.pem") 
       if err != nil {
          return
       }

       originPublicKeyCertificate = strings.TrimSpace(string(originPublicKeyCertificateBytes))

       signatureBytes, err = zip_util.ExtractFileFromZipFileContents(zipFileContents, "signatureOfArtifactVersionContents.pem") 
       if err != nil {
          return
       }

       signature = strings.TrimSpace(string(signatureBytes))

       installationSharedRelishPublicKeyCert, err = crypto_util.GetPublicKeyCert("origin", "shared.relish.pl2012")
       installationSharedRelishPublicKeyCert = strings.TrimSpace(installationSharedRelishPublicKeyCert)
       if err != nil {
          return
       }

       //   Is the shared relish public key cert in the artifact zip file identical to the cert that came with my
       //   relish distribution? If not, panic.

       if sharedRelishPublicKeyCertificate != installationSharedRelishPublicKeyCert {
          err = fmt.Errorf("Did not install downloaded artifact because shared.relish.pl2012 public key certificate\n" + 
                           "in artifact %s (v%s) downloaded from %s\n" + 
                           "is different than shared.relish.pl2012 public key certificate in this relish installation.\n", 
                           originAndArtifactPath, version, hostURL)    
          return      
       }

       // Validate that shared.relish.pl2012 public key is signed properly, obtaining the shared.relish.pl2012 publicKeyPEM.

       sharedRelishPublicKey := crypto_util.VerifiedPublicKey("", sharedRelishPublicKeyCertificate, "origin", "shared.relish.pl2012") 

       if sharedRelishPublicKey == "" {
          err = errors.New("Invalid shared.relish.pl2012 public key certificate.")
          return
        }

        // Validate the artifact-publishing origin's public key cert

        slashPos := strings.Index(originAndArtifactPath,"/")
        originId := originAndArtifactPath[:slashPos]
        
        originPublicKey := crypto_util.VerifiedPublicKey(sharedRelishPublicKey, originPublicKeyCertificate, "origin", originId) 

        if originPublicKey == "" {
             err = fmt.Errorf("Did not install downloaded artifact because %s public key certificate\n" + 
                              "in artifact %s (v%s) downloaded from %s\n" + 
                              "is invalid.\n", 
                              originId, originAndArtifactPath, version, hostURL)             
            return
        }

        signedContent := zipFileName + "_|_" + string(srcZipFileContents)
        if ! crypto_util.Verify(originPublicKey, signature, signedContent) {
             err = fmt.Errorf("Did not install downloaded artifact because artifact version content\n" + 
                              "in artifact %s (v%s) downloaded from %s\n" + 
                              "does not match (was not verified by) its digital signature.\n", 
                              originAndArtifactPath, version, hostURL)             
            return
        }

       // Woohoo! Contents are verified. 
       //
       /////////////////////////////////////////////////////////////////////////////////////
       //
       // Write them to relish installation shared code directory tree.

       // Note: Assuming the artifactVersionContents.zip file starts with src/ pkg/ doc/ etc not with v0002/       

       err = zip_util.ExtractZipFileContents(srcZipFileContents, artifactVersionDir) 
	   if err != nil {
	      return
	   }
	
       Log(ALWAYS_,"Downloaded %s (v%s) from %s\n", originAndArtifactPath, version, hostURL)		
	
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
		       artifactVersion := artifactsVersionsStrs[i+1] 

			   alreadyDesiredVersion,versionFound := ldr.LoadedArtifacts[artifactPath]	
			   if versionFound {
				   if artifactVersion != alreadyDesiredVersion {
				      Log(ALWAYS_,"Using v%s of %s. %s (v%s) may prefer v%s of %s.\n",alreadyDesiredVersion, artifactPath, originAndArtifactPath, version, artifactVersion, artifactPath)
			       }
			   } else {
                  // Tell the loader to prefer the version of the other artifact 
                  // that the being-loaded artifact built.txt file specifies. 
			      ldr.LoadedArtifacts[artifactPath] = artifactVersion	
			   }
	        }		

        } else if ! os.IsNotExist(statErr) {
			err = fmt.Errorf("Can't stat '%s': %v\n", builtFilePath, statErr)
			return		       
	    }     
        // else there is no built.txt file. Accept that for now
        // and subsequently load the current versions of artifacts where no other version is preferred.
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

    var sourceDirFile *os.File
    sourceDirFile, err = os.Open(packageSourcePath) 
    if err != nil {
      return
    }
    defer sourceDirFile.Close()

    var filenames []string
    filenames,err = sourceDirFile.Readdirnames(-1) 
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

	ldr.LoadedArtifactKnownToBePublished[originAndArtifactPath] = strings.Contains(artifactVersionDir,"/shared/relish/artifacts/") 
	ldr.LoadedArtifactKnownToBeReplica[originAndArtifactPath] = artifactKnownToBeReplica 	
	  
    Log(LOAD2_,"ldr.LoadedArtifactKnownToBePublished[%s]=%v\n",originAndArtifactPath,ldr.LoadedArtifactKnownToBePublished[originAndArtifactPath])
    Log(LOAD2_,"ldr.LoadedArtifactKnownToBeReplica[%s]=%v\n",originAndArtifactPath,ldr.LoadedArtifactKnownToBeReplica[originAndArtifactPath])
    Logln(LOAD_,"artifactVersionDir="+artifactVersionDir)


    ldr.LoadedArtifactKnownToBeLocal[originAndArtifactPath] = ! (ldr.LoadedArtifactKnownToBePublished[originAndArtifactPath] || artifactKnownToBeReplica)	

    
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

    // Collect a map of file nodes to the root of the filename. We will be passing this to the generator to generate runtime
    // code for the whole package all at once.
    //
    astFileNodes := make(map[*ast.File]string)    

    for _,filename := range filenames {
		var sourceFound bool
		var pickledFound bool	
	    if strings.HasSuffix(filename,".rel") { // consider only the relish source files in the dir.
		// This is actually quite controversial, since it means that source code MUST be present
		// or we won't bother looking for the compiled file to load.
		// This is a somewhat political opinionated decision. Will have to be seriously mulled if not pondered.
			
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
		  	  fileNode, err = parser.ParseFile(fset, sourceFilePath, nil, parserDebugMode)
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
		
		   astFileNodes[fileNode] = fileNameRoot
//         gen = generator.NewGenerator(fileNode, fileNameRoot) // TODO NOW add a isLocal =ldr.LoadedArtifactKnownToBeLocal[originAndArtifactPath]
                                                                // argument so that we can flag the RPackage object as local or shared.
//         gen.GenerateCode()	


           if packageIdentifier != fileNode.Name.Name {
              err = fmt.Errorf("\nThe origin, artifact, or package metadata at top of source code file\n'%s'\ndoes not match the package directory path where the file resides.\n",sourceFilePath)  
              return   
           }


	       if parseNeeded {
		      Log(ALWAYS_,"Compiled %s\n", sourceFilePath)		
		   } 
		
	    } // end of if it is a code file.
    } // end of loop over each file in the package.

    if len(astFileNodes) > 0 {
       gen = generator.NewGenerator(astFileNodes)
       gen.GenerateCode()
    }

    native_methods.WrapNativeMethods(packageIdentifier)  // Check if package has native methods, if so, make RMethod wrappers.
    
    ldr.LoadedPackages[packageIdentifier] = version

    delete(ldr.PackagesBeingLoaded,packageIdentifier)

    Log(ALWAYS_,"Loaded %s\n", packageCompiledPath)	
    
   return
}




  	
/*
If a file at the specified path exists, reads the current version info, and the metadata date, from it.
If file does not exist, returned currentVersion is "" as is metadataDate, with no error.
If file exists but not a properly formatted metadata file, returns an error.
*/
func (ldr *Loader) readMetadataFile(path string) (currentVersion string, metadataDate string, err error) {

	_,statErr := os.Stat(path)
	if statErr != nil {
	    if ! os.IsNotExist(statErr) {
			err = fmt.Errorf("Can't stat '%s': %v\n", path, statErr)
			return		       
	    }

	    // did not find the metadata.txt file in the local (private) artifact dir tree
	    //
	    // so there is no local (private) artifact

	} else { // found the metadata.txt file in the local (private) artifact dir tree
 
	    var body []byte
		body, err = ioutil.ReadFile(path)	
		if err != nil {
			return
		}
        currentVersion, metadataDate, err = readCurrentVersion(body, path) 			
	}
	return
}

/*
Read the current version and metadata date information from the contents of an artifact's metadata.txt file
*/
func readCurrentVersion(metadata []byte, metadataFilePath string) (currentVersion string, metadataDate string, err error) {
    match := reMetadataDate.FindSubmatchIndex(metadata)
    if match == nil {
       err = fmt.Errorf("%s file first line must be formatted like this example: relish artifact metadata: 2013/02/19", metadataFilePath)
       return
    } 

    dateStart := match[2]
    dateEnd := match[3]   

    metadataDate = string(metadata[dateStart:dateEnd])    

    match = reCurrentVersion.FindSubmatchIndex(metadata)
    if match == nil {
       err = fmt.Errorf("%s file must have a line like: current version: 1.0.23", metadataFilePath)
       return
    }
    versionNumStart := match[2]
    versionNumEnd := match[3]	

    currentVersion = string(metadata[versionNumStart:versionNumEnd])
    return
}


/*
Returns the URL of the default host that should host the artifact.
*/
func (ldr *Loader) DefaultCodeHost (originAndArtifactPath string) string {
   return "http://" + originAndArtifactPath[:strings.Index(originAndArtifactPath,"/")-4]
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
Returns a list of urls of http servers, in order, to check for the presence of the artifact.
Starts with any staging servers for the code origin (from xtras/relish_code_staging_servers.txt)
Then comes the default standard host name for the code origin.
Then comes alternate hosts for the origin (from xtras/relish_code_origins.txt)
Then comes known replicating hosts of the origin (from xtras/relish_code_replicas.txt)
Then comes other general relish code repository hosts (from xtras/relish_code_repositories.txt)
Does not include Google-searched servers.
Does not include the shared.relish.pl server

If startingHost is present, the returned list of hosts will begin at that host instead of at the beginning
of the list that would otherwise be returned. This allows hosts that have been unsuccessfully checked for
metadata.txt files to not be checked for artifact zip files either.
*/
func (ldr *Loader) NonSearchedCodeHostURLs (originAndArtifactPath string, startingHost string) (hostURLs []string) {

    origin := ldr.Origin(originAndArtifactPath)

    hostURLs, found := ldr.nonSearchedOriginHostUrls[origin]
    if ! found {
        var hosts []string
        hosts = ldr.nonSearchedCodeHostURLs(origin)
        if startingHost != "" {
            for i,host := range hosts {
                if host == startingHost {
                    hostURLs = hosts[i:]
                    ldr.nonSearchedOriginHostUrls[origin] = hostURLs
                    return
                }
            }
        }
        hostURLs = hosts
    }
    return
}

/*
Helper. Constructs a list in a particular order of host urls to search for relish source code from a particular code origin.
*/
func (ldr *Loader) nonSearchedCodeHostURLs (origin string) (hostURLs []string) {

    defaultOriginHostURL := "http://" + origin[:len(origin)-4]

    hostURLs = ldr.stagingCodeHostURLs(origin)
    hostURLs = append(hostURLs, defaultOriginHostURL)
    hostURLs = append(hostURLs, defaultOriginHostURL + ":" + STANDARD_SOURCE_CODE_SHARING_PORT)
    hostURLs = append(hostURLs, ldr.originCodeHostURLs(origin)...)
    hostURLs = append(hostURLs, ldr.replicaCodeHostURLs(origin)...)
    hostURLs = append(hostURLs, ldr.repositoryHostURLs()...)     
    return
}

// TODO These should be maps and lists in the loader, initialized at startup from the rt/xtras config files.

func (ldr *Loader) stagingCodeHostURLs (origin string) []string {
    return ldr.stagingServerUrls[origin]
}

func (ldr *Loader) originCodeHostURLs (origin string) []string {
    return ldr.originUrls[origin]
}

func (ldr *Loader) replicaCodeHostURLs (origin string) []string {
    return ldr.replicaUrls[origin]
}

func (ldr *Loader) repositoryHostURLs () []string {
    return ldr.repositoryUrls
}


func (ldr *Loader) initCodeLocations() {
	xtrasDirPath := ldr.RelishRuntimeLocation + "/xtras/"
	stagingServersFilePath := xtrasDirPath + "relish_code_staging_servers.txt"
	originsFilePath := xtrasDirPath + "relish_code_origins.txt"
	replicasFilePath := xtrasDirPath + "relish_code_replicas.txt"
	repositoriesFilePath := xtrasDirPath + "relish_code_repositories.txt"			

	f,err := os.Open(stagingServersFilePath)
	if err == nil {
	   r := bufio.NewReader(f)
	   for {
		  line, err := r.ReadString('\n')
		  if err != nil {
			break
		  }
		  words := strings.Fields(line)
		  if len(words) < 2 {
			 fmt.Println("Error: xtras/relish_code_staging_servers.txt has a line with too few entries on it.")
		  } else {
			 origin := words[0]
			 servers := words[1:]
			 if ! strings.HasPrefix(origin, "#") {
				var serverURLs []string
				for _,server := range servers {
					serverURLs = append(serverURLs, "http://" + server)
					if strings.Index(server,":") == -1 {
					   serverURLs = append(serverURLs, "http://" + server + ":" + STANDARD_SOURCE_CODE_SHARING_PORT)						
					}
				} 
				ldr.stagingServerUrls[origin] = serverURLs
			 }
		  }
	   } 	
	   f.Close()	 
    }	

	f,err = os.Open(originsFilePath)
	if err == nil {
	   r := bufio.NewReader(f)
	   for {
		  line, err := r.ReadString('\n')
		  if err != nil {
			break
		  }
		  words := strings.Fields(line)
		  if len(words) < 2 {
			 fmt.Println("Error: xtras/relish_code_staging_servers.txt has a line with too few entries on it.")
		  } else {
			 origin := words[0]
			 servers := words[1:]
			 if ! strings.HasPrefix(origin, "#") {
				var serverURLs []string
				for _,server := range servers {
					serverURLs = append(serverURLs, "http://" + server)
					if strings.Index(server,":") == -1 {
					   serverURLs = append(serverURLs, "http://" + server + ":" + STANDARD_SOURCE_CODE_SHARING_PORT)						
					}
				} 				
				ldr.originUrls[origin] = serverURLs
			 }
		  }
	   } 	
	   f.Close()	 
    }

	f,err = os.Open(replicasFilePath)
	if err == nil {
	   r := bufio.NewReader(f)
	   for {
		  line, err := r.ReadString('\n')
		  if err != nil {
			break
		  }
		  words := strings.Fields(line)
		  if len(words) < 2 {
			 fmt.Println("Error: xtras/relish_code_staging_servers.txt has a line with too few entries on it.")
		  } else {
			 origin := words[0]
			 servers := words[1:]
			 if ! strings.HasPrefix(origin, "#") {
				var serverURLs []string
				for _,server := range servers {
					serverURLs = append(serverURLs, "http://" + server)
					if strings.Index(server,":") == -1 {
					   serverURLs = append(serverURLs, "http://" + server + ":" + STANDARD_SOURCE_CODE_SHARING_PORT)						
					}
				} 					
				ldr.replicaUrls[origin] = serverURLs
			 }
		  }
	   } 	
	   f.Close()	 
    }

    var b []byte
	b,err = ioutil.ReadFile(repositoriesFilePath)
	if err == nil {
		servers := strings.Fields(string(b))
		var serverURLs []string
		for _,server := range servers {
			serverURLs = append(serverURLs, "http://" + server)
			if strings.Index(server,":") == -1 {
			   serverURLs = append(serverURLs, "http://" + server + ":" + STANDARD_SOURCE_CODE_SHARING_PORT)						
			}
		}
		ldr.repositoryUrls = serverURLs				 
    }
}






/*
Return the origin part of the originAndArtifact path.
*/  
func (ldr *Loader) Origin(originAndArtifactName string) string {
    return originAndArtifactName[:strings.Index(originAndArtifactName,"/")]
}   




/*
Return the version of the artifact that is to be loaded. Can return "" (no preference)
*/	
func (ldr *Loader) ArtifactVersion(originAndArtifactName string) string {
	return ldr.LoadedArtifacts[originAndArtifactName]
}	

	
/*
Check the imports list of the relish intermediate-code file and load the packages if not already loaded.
Requires consultation of the current package's artifact's built.txt information on which version of each other artifact must be loaded.
*/	
func (ldr *Loader) ensureImportsAreLoaded(fileNode *ast.File) (err error) {
	imports := fileNode.RelishImports  // package specifications
	for _,importedPackageSpec := range imports {
		if importedPackageSpec.OriginAndArtifactName == "relish" {
			continue  // special case for the subset of standard lib packages which only contain Go native methods and
			          // have no relish package to load.
		}
		importedArtifactVersion := ldr.ArtifactVersion(importedPackageSpec.OriginAndArtifactName)	
		
		_,err = ldr.LoadPackage(importedPackageSpec.OriginAndArtifactName,
			                  importedArtifactVersion,
		                      importedPackageSpec.PackageName, 
		                      false)		
        if err != nil {
	       if importedArtifactVersion == "" {
		      Log(ALWAYS_,"Error loading package %s from current version of %s:  %v\n", importedPackageSpec.PackageName,importedPackageSpec.OriginAndArtifactName, err)		
		   } else {
		      Log(ALWAYS_,"Error loading package %s from version %s of %s:  %v\n", importedPackageSpec.PackageName,importedArtifactVersion,importedPackageSpec.OriginAndArtifactName, err)		
	       }
		   break	
       }			
	} 
	return
}	