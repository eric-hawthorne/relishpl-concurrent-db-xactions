// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

/*
relish interpreter main program.

Note: This is not expected to work on Windows yet due to path separator differences.

Current usage: 

cd $RELISH_HOME/artifacts/some.codeorigin.com2007/some/artifact/path/v0001
relish path/to/package

or

cd $RELISH_HOME/artifacts/some.codeorigin.com2007/some/artifact/path/v0001/pkg/path/to/package
relish

or

relish some.codeorigin.com2007/some/artifact/path 2 path/to/package

or 

relish some.codeorigin.com2007/some/artifact/path path/to/package

which chooses the current version as specified in some/artifact/path/metadata.txt
- note that this will first check if the artifact has been installed locally, and will use the
  local copy of the artifact's metadata.txt to determine the current version. This could be
  out of date. TODO We need another command or command option to force an internet search for the 
  true authoritative current version of the artifact. 


Command line options:

-log 1|2    The logging level: 1 (some debugging info), 2 more. Very minimal logging of key runtime events if not supplied.

-web <port#>  The http listening port - if not supplied, does not listen for http requests



-db <dbname>  The database name. A SQLITE database file called <dbname>.db will be created/used in artifact data directory.
              Defaults to db1   i.e. db1.db	

-cpuprofile <filepath>.prof  Write cpu profile to file. Then use go tool pprof /opt/devel/relish/bin/relish somerun.prof 


-publish origin/artifact [version#]    Copies to shared/relish/artifacts directory tree   - served if sharing

-makecurrent origin/artifact version#

-share <port#> Serve source code (contents of the shared directory) on the specified port. Port should be 80
               or failing that 8421, or, if behind apache2 modproxy, any other port is fine but apache2 should
               present it as port 80 or port 8421. It is ok for the share port to be the same as the web port.

TODO 
-initproject origin/artifact 
-initwebproject origin/artifact

-refresh origin[/artifact] delete the local replica metadata.txt files for the specified origin, or 
 just the single metadata.txt file for the specified artifact, so that a newest metadata.txt file will
 be downloaded from from their originating (or other replica) server. Actually, just move the 
 metadata.txt file, and have it do a comparison of metadata dates to see which one to keep.

TODO Further future, speculative:
Note that maybe there should also be a special relish method which does this, universally, or by origin or originAndArtifact.
This method could be executed by a web request, if desired and if the web request handler was written, so as to
allow remote renewal of some or all of the code running in a relish installation.
We would have to have an unload-and-reload feature in the relish runtime to allow in-situ recompilation and linking
of stuff. This is complicated, as it would require stopping background "threads" and re-initializing things that
were initialized at the first startup and load in the runtime. (It's also a bit of a potential security nightmare.)

*/
package main

import (
        "fmt"
        "flag"
        "strings"
        "os"
//		"relish/compiler/token"
//		"relish/compiler/ast"	
//		"relish/compiler/parser"
		"relish/compiler/generator"
		"relish/runtime/native_methods/builtin"
		"relish/runtime/web"	  
		"relish/dbg"
		"relish/global_loader"
		"relish/global_publisher"		
		"util/crypto_util"
		"regexp"
//		"strconv"
		"runtime/pprof"
)

var reVersionAtEnd *regexp.Regexp = regexp.MustCompile("/v([0-9]+\\.[0-9]+\\.[0-9]+)$")
var reVersionedPackage *regexp.Regexp = regexp.MustCompile("/v([0-9]+\\.[0-9]+\\.[0-9]+)/pkg/")
var reVersion *regexp.Regexp = regexp.MustCompile("([0-9]+\\.[0-9]+\\.[0-9]+)")

func main() {
    var loggingLevel int
    var webListeningPort int
    var shareListeningPort int  // port on which source code will be shared by http    
    var sharedCodeOnly bool  // do not use local artifacts - only those in shared directory.
    var explorerListeningPort int // port on which data explorer_api web service will be served.   
    var runningArtifactMustBeFromShared bool
    var dbName string 
    var cpuprofile string
    var publish bool
    var quiet bool
    var projectPath string

    //var fset = token.NewFileSet()
	flag.IntVar(&loggingLevel, "log", 0, "The logging level: 0 is least verbose, 2 most")	
	flag.IntVar(&webListeningPort, "web", 0, "The http listening port - if not supplied, does not listen for http requests")
	flag.IntVar(&explorerListeningPort, "explore", 0, "The explorer_api web service listening port - if supplied, the data explorer tool can connect to this program on this port")		
	flag.StringVar(&dbName, "db", "db1", "The database name. A SQLITE database file called <name>.db will be created/used in artifact data directory")			

	flag.BoolVar(&sharedCodeOnly, "shared", false, "Use shared version of all artifacts - ignore local/dev copy of artifacts")		
 
    flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")

	flag.IntVar(&shareListeningPort, "share", 0, "The code sharing http listening port - if not supplied, does not listen for source code sharing http requests")	    

    flag.BoolVar(&publish, "publish", false, "artifactpath version - copy specified version of artifact to shared/relish/artifacts")

    flag.BoolVar(&quiet, "quiet", false, "do not show package loading info or interpreter version in program output")

    flag.StringVar(&projectPath, "init", "", "<artifactpath> [webapp] - create directory tree and template files for a relish software project")    
    
    flag.Parse()


    pathParts := flag.Args() // full path to package, or originAndArtifact and path to package 
                             // (or originAndArtifact and version number if -publish)


    if cpuprofile != "" {
        f, err := os.Create(cpuprofile)
        if err != nil {
		     fmt.Println(err)
		     return 
		  }
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
    }

   	dbg.InitLogging(int32(loggingLevel))
	//relish.InitRuntime("relish.db")
    
//  if ! publish {
//    	builtin.InitBuiltinFunctions()	
//	}

	  var g *generator.Generator
	
	  var relishRoot string  // This actually has to be the root of the runtime environment
	                       // i.e. /opt/relish if this is a binary distribution,
	                       // or /opt/relish/rt if this is a source distribution
	
    workingDirectory, err := os.Getwd()
    if err != nil {
	   fmt.Printf("Cannot determine working directory: %s\n",err)	
    }


    var originAndArtifact string
    var version string
    var packagePath string
    isSubDir := false
    isSourceDist := false

    relishIndexInWd := strings.Index(workingDirectory,"/relish/")
    if relishIndexInWd > -1 {
       isSubDir = true
    } else {
       relishIndexInWd = strings.Index(workingDirectory,"/relish")
       if relishIndexInWd == -1 || relishIndexInWd != len(workingDirectory) - 7 {
         if projectPath != "" {  // Must be creating a relish dir under a new project dir
            relishRoot = workingDirectory + "/relish"
         
            err = os.MkdirAll(relishRoot,0777)       
            if err != nil {
              fmt.Printf("Error making relish project directory %s: %s\n", relishRoot,err)
              return 
            }                  
         } else { 
    		   fmt.Printf("relish command must be run from within a relish directory tree.\n")
    		   return  
    		}        
       }
    }
    if relishRoot == "" {
       relishRoot = workingDirectory[:relishIndexInWd + 7]
    }
    
    _,err = os.Stat(relishRoot + "/rt")	
    if err == nil {
       relishRoot += "/rt"
       isSourceDist = true
    } else if ! os.IsNotExist(err) {
		   fmt.Printf("Can't stat '%s' : %v\n", relishRoot + "/rt", err)
 	     return		       	
    }    
    
    // relishRoot is now established
    
    if projectPath != "" {
       if len(pathParts) < 1 {
          err = initProject(relishRoot, projectPath, "")
          if err == nil {
            fmt.Printf("Created relish project template %s/artifacts/%s\n", relishRoot,projectPath)           
            fmt.Printf("To run your project template's dummy main program, relish %s\n", projectPath)                 
          } else {          
            fmt.Printf("Error initializing project %s: %s\n", projectPath, err)
          }                 
       } else { 
          projectType := pathParts[0]
          err = initProject(relishRoot, projectPath, projectType)
          if err == nil {
             fmt.Printf("Created relish web-app project template %s/artifacts/%s.\n", relishRoot,projectPath)          
             fmt.Printf("To run the web-app, relish -web 8080 %s\n", projectPath) 
             fmt.Printf("Then enter localhost:8080 into your browser's address bar to view the web app.\n", projectPath)              
          } else {          
            fmt.Printf("Error initializing project %s: %s\n", projectPath, err)
          }                 
       }       
       return    
    }
    
    
    
    if isSubDir { // See where we are more specifically
       if isSourceDist {
          idx := strings.Index(workingDirectory,"/rt/shared") 
          if idx > -1 {
             runningArtifactMustBeFromShared = true
          } 
       } else {
          idx := strings.Index(workingDirectory,"/relish/shared") 
          if idx > -1 {
             runningArtifactMustBeFromShared = true
          }   
       }
     
   	   if ! publish {   

        // See if the current directory is a particular artifact version directory,
        // or even is a particular package directory.

   	     originPos := strings.Index(workingDirectory,"/artifacts/") + 11
   	     if originPos == 10 {
   	        originPos = strings.Index(workingDirectory,"/replicas/") + 10	       
	       }
	       if originPos >= 10 {
             match := reVersionedPackage.FindStringSubmatchIndex(workingDirectory)
             if match != nil {
      	        version = workingDirectory[match[2]:match[3]]
             
      	        originAndArtifact = workingDirectory[originPos:match[0]]
      	        packagePath = workingDirectory[match[3]+1:]	
             } else {
      	        match := reVersionAtEnd.FindStringSubmatch(workingDirectory)
      	        if match != nil {
      		         version = match[1]	
      	           originAndArtifact = workingDirectory[originPos:len(workingDirectory)-len(version)-2]		
      		      }
             }
          }
       }
    }

    if ! publish {
	    builtin.InitBuiltinFunctions(relishRoot)	
    }

    crypto_util.SetRelishRuntimeLocation(relishRoot)  // So that keys can be fetched.   

    if publish {
      if len(pathParts) < 2 {
          fmt.Println("Usage (example): relish -publish someorigin.com2013/artifact_name 1.0.23")
          return
      }
	    originAndArtifact = pathParts[0]
	    version = pathParts[1]

      err = global_publisher.PublishSourceCode(relishRoot, originAndArtifact, version)
	    if err != nil {
		    fmt.Println(err)
      }
      return
    }

 
    sourceCodeShareDir := ""
    if shareListeningPort != 0 {
      // sourceCodeShareDir hould be the "relish/shared" 
      // or "relish/rt/shared" of "relish/4production/shared" or "relish/rt/4production/shared" directory.		
      sourceCodeShareDir = relishRoot + "/shared"
    }
    onlyCodeSharing := (shareListeningPort != 0 && webListeningPort == 0)

    if onlyCodeSharing {

      if shareListeningPort < 1024 && shareListeningPort != 80 && shareListeningPort != 443 {
         fmt.Println("Error: The source-code sharing port must be 80, 443, or > 1023 (8421 is the standard if using a high port)")
         return		
      }  		

      web.ListenAndServeSourceCode(shareListeningPort, sourceCodeShareDir)	
    }


    var loader = global_loader.NewLoader(relishRoot, sharedCodeOnly, dbName + ".db", quiet)
  	
  	
    if originAndArtifact == "" {
  		
         if len(pathParts) == 3 {  // originAndArtifact version packagePath
  	
  		    originAndArtifact = pathParts[0]
  		    version = pathParts[1]
  		    packagePath = pathParts[2]		
  	
         } else if len(pathParts) == 2 {  // originAndArtifact packagePath	
  		
  		    originAndArtifact = pathParts[0]
  		    packagePathOrVersion := pathParts[1]
  		    
          // Determine if is a version by regexp or a package path has been supplied
         
          version = reVersion.FindString(packagePathOrVersion)
          if version == "" {
             packagePath = packagePathOrVersion

          } 
  	   } else if shareListeningPort == 0 || webListeningPort != 0 {
         if len(pathParts) != 1 {
  	       fmt.Println("Usage: relish [-web 80] originAndArtifact [version] [path/to/package]\n# package path defaults to main")            
  	       return
         }
         originAndArtifact = pathParts[0]       
      }
    } else if packagePath == "" {
       if len(pathParts) == 1 {
          packagePath = pathParts[0]        
       } else if len(pathParts) > 1 {
  	      fmt.Println("Usage (when in an artifact version directory): relish [-web 80] [path/to/package]\n# package path defaults to main")
          return
       }	
    } else {  // both originAndArtifact and packagePath are defined (non "")
       if len(pathParts) != 0 {
           fmt.Println("Usage (when in a package directory): relish [-web 80]")
           return
       }		
    }


  	if strings.HasSuffix(originAndArtifact,"/") {  // Strip trailing / if present
  	   originAndArtifact = originAndArtifact[:len(originAndArtifact)-1]	
  	}
    if strings.HasSuffix(packagePath,"/") {  // Strip trailing / if present
       packagePath = packagePath[:len(packagePath)-1]	
    }

    if packagePath == "" {
       packagePath = "main"  // substitute a default.
    }
    
    fullPackagePath := fmt.Sprintf("%s/v%s/pkg/%s",originAndArtifact,version, packagePath)
    fullUnversionedPackagePath := fmt.Sprintf("%s/pkg/%s",originAndArtifact, packagePath)
    
    g, err = loader.LoadPackage(originAndArtifact, version, packagePath, runningArtifactMustBeFromShared)

    if err != nil {
	    if version == "" {
		   fmt.Printf("Error loading package %s from current version of %s:  %v\n", packagePath, originAndArtifact, err)		
		} else {
		   fmt.Printf("Error loading package %s:  %v\n",fullPackagePath, err)
	    }
		return	
    }


    g.Interp.SetRunningArtifact(originAndArtifact) 

    g.Interp.SetPackageLoader(loader)


    // TODO the following rather twisty logic  (from here to end of main method) could be straightened out.
    // One of its purposes is to ensure that the last http listener is run in this goroutine rather
    // than in a background one. And there can be different numbers of listeners...

    // Count how many separate listeners there will be.

    numListeners := 0
    numListening := 0  // how many of those are already listening?

    if webListeningPort != 0 {
    	numListeners += 1
    	if shareListeningPort != 0 && shareListeningPort != webListeningPort {
    		numListeners += 1
    	}
    }
    if explorerListeningPort != 0 {
    	numListeners += 1
    }

    // end counting listeners


    // check for disallowed port numbers, and if not, load the packages needed for web app serving

	if webListeningPort != 0 {
	   if webListeningPort < 1024 && webListeningPort != 80 && webListeningPort != 443 {
			fmt.Println("Error: The web listening port must be 80, 443, or > 1023")
			return		
	   }
	
       if shareListeningPort != webListeningPort && shareListeningPort != 0 && shareListeningPort < 1024 && shareListeningPort != 80 && shareListeningPort != 443 {
	  	  fmt.Println("Error: The source-code sharing port must be 80, 443, or > 1023 (8421 is the standard if using a high port)")
		  return		
       }		
	
       err = loader.LoadWebPackages(originAndArtifact, version, runningArtifactMustBeFromShared)	
	   if err != nil {
		    if version == "" {
			   fmt.Printf("Error loading web packages from current version of %s:  %v\n", originAndArtifact, err)		
			} else {
			   fmt.Printf("Error loading web packages from version %s of %s:  %v\n", version, originAndArtifact, err)
		    }
			return	
	   }
	}

    // check for disallowed port numbers, and if not, load the package needed for explorer_api web service serving

	if explorerListeningPort != 0 {
	   if explorerListeningPort < 1024 && explorerListeningPort != 80 && explorerListeningPort != 443 {
			fmt.Println("Error: The explorer listening port must be 80, 443, or > 1023")
			return		
	   }


	   explorerApiOriginAndArtifact := "shared.relish.pl2012/explorer_api"
	   explorerApiPackagePath := "web"
       _, err = loader.LoadPackage(explorerApiOriginAndArtifact, "", explorerApiPackagePath, false)

       if err != nil {
		   fmt.Printf("Error loading package %s from current version of %s:  %v\n", 
		              explorerApiPackagePath, 
		              explorerApiOriginAndArtifact, 
		              err)		
   		   return	
       }	     
    }

	   
	if numListeners > 0 {  // If we'll be listening for http requests, run main in a background goroutine.
		go g.Interp.RunMain(fullUnversionedPackagePath, quiet)
	}


	if webListeningPort != 0 {

	   web.SetWebPackageSrcDirPath(loader.PackageSrcDirPath(originAndArtifact + "/pkg/web"))
	  
	   if shareListeningPort == webListeningPort {
          numListening += 1
          if numListening == numListeners {
	         web.ListenAndServe(webListeningPort, sourceCodeShareDir)
	      } else {
	         go web.ListenAndServe(webListeningPort, sourceCodeShareDir)	      	
	      }
	   } else {
          if shareListeningPort != 0 {
          	 numListening += 1
	         go web.ListenAndServeSourceCode(shareListeningPort, sourceCodeShareDir) 
	      }		

	      numListening += 1
	      if numListening == numListeners {	
	         web.ListenAndServe(webListeningPort, "")	
	      } else {
	         go web.ListenAndServe(webListeningPort, "")	      	
	      }
	   }
	   	
	} 
	if explorerListeningPort != 0 {         
      web.ListenAndServeExplorerApi(explorerListeningPort)	          
   }
   
   // This will only be reached if numListeners == 0 or there is an error starting listeners.
   // but don't want to re-run main if there was an error starting listeners.
   if numListeners == 0 {
      g.Interp.RunMain(fullUnversionedPackagePath,quiet)
   }
}