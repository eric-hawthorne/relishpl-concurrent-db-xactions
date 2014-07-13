// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// loads relish code artifacts from the Internet, assuming strong choice-free conventions
// about where a given version of a given artifact will reside or how it will be found, 
// both on the Internet as a whole, and on each server that hosts it.

// TODO The extensive comments below are out of date.

package global_loader

import (
    "fmt"
    "net/http"
    "regexp"
    "strings"
    "os"
    "util/gos"
    "util/net_util"
    "io/ioutil"
)

/*
Fetch code over the internet and install it locally.

Relish code is distributed in (versioned) units called artifacts.
An artifact will either be a code library, or an application. It is expected to be common for relish artifacts to depend on 
other relish artifacts, from other code originators, so there are strict code location and organization conventions 
to ensure that dependencies can be fetched automatically from the Internet during a software build.

A relish artifact has an originating organization (or organizational unit.) This is known as the origin.
The origin must be identified by a domain name such as example.org or mysuborg.example.org. 
Actually, so that origins are completely unique through time, the origin identifier is the domain name with 
a 4 digit year appended. The year should be the first calendar year that the domain is/was registered to the 
originating organization/organizational unit for the majority of the year. The year is NOT the year that the
software artifact was produced.

Relish code artifact originators can and should register their code origin domain at the website 
http://register.relish.pl . This ensures origin uniqueness. Originators can also register each software artifact 
there if they wish. Doing so will ensure their software appears in a directory of relish code artifacts.
 
An artifact must be named according to the following conventions:
some_artifact_name or some/long_path/to/some_artifact_name

Each publicly released relish code artifact should be hosted at (by example):
http://mysuborg.example.org/relish/artifacts/mysuborg.example.org2012/some/long_path/to/some_artifact_name 
but in any case, must be hosted on some Google-searchable website at the path:
/relish/artifacts/mysuborg.example.org2012/some/long_path/to/some_artifact_name 
in order to be found by the relish code fetching-and-building system.

For each artifact that is published, the minimum file set that must be present on the website
is:
relish/artifacts/some.origin.org2012/some/artifact/name/metadata.txt
relish/artifacts/some.origin.org2012/some/artifact/name/v0001.zip
relish/artifacts/some.origin.org2012/some/artifact/name/v0002.zip
...
etc. depending on number of available released versions.


A relish code projects file structure must be organized on the computer file system as follows:


--------------
relish/artifacts/some.origin.org2012/some/artifact/name/v0002/src/path/to/thepackage/some_file.rel
relish/artifacts/some.origin.org2012/some/artifact/name/v0002/src/path/to/thepackage/test_something.rel
relish/artifacts/some.origin.org2012/some/artifact/name/v0002/src/migrate/down_from.rel
relish/artifacts/some.origin.org2012/some/artifact/name/v0002/src/migrate/up_to.rel
relish/artifacts/some.origin.org2012/some/artifact/name/v0002/pkg/path/to/thepackage/some_file.rlc
relish/artifacts/some.origin.org2012/some/artifact/name/v0002/pkg/path/to/thepackage/test_something.rlc
relish/artifacts/some.origin.org2012/some/artifact/name/v0002/pkg/migrate/down_from.rlc
relish/artifacts/some.origin.org2012/some/artifact/name/v0002/pkg/migrate/up_to.rlc
relish/artifacts/some.origin.org2012/some/artifact/name/v0002/doc/readme.html
relish/artifacts/some.origin.org2012/some/artifact/name/v0002/dependencies.txt
relish/artifacts/some.origin.org2012/some/artifact/name/v0002.zip    // ready to be copied to public web directory to release it

relish/artifacts/some.origin.org2012/some/artifact/name/metadata.txt
relish/artifacts/some.origin.org2012/some/artifact/name/v0002/built.txt

relish/relish0.1


The relish build system will automatically look for relish code in the following directories,
and will stop looking after finding relish code in one of these places.
If relish is not there, the RELISH_HOME environment variable must be set to the active relish
code directory.

/usr/local/relish
/opt/relish
/opt/devel/relish
C:\relish
C:\devel\relish

Standard files conventions
==========================

metadata.txt - name, current version, classification, and brief description of the software artifact.
---------------------------------
relish artifact metadata: 2013/09/25
origin: some.origin.org2012
artifact: some/artifact/name
current version: 1.0.23
release date: 2012/03/27
tags: application,game

Start the 500 word or less summary on this line.
Continue it like this.
---------------------------------


v1.0.23/built.txt  - metadata listing the version of each dependency which was used in the latest successful test execution of this
                    version of this artifact's test suite.

---------------------------------
some.otherorigin.com2011/some/artifact_name 0.9.26
yet.anotherorigin.com2012/artifact_name 2.3.1
---------------------------------


v01.0.23/dependencies.txt  - metadata listing the versions of each dependency which this version of this artifact have been tested against successfully.
                          Note that this does not mean that all combinations of the listed versions of dependency A have been tested 
                          with all listed versions of dependency B. 

---------------------------------
some.otherorigin.com2011/some/artifact_name 2.3.1-3.0.14,3.0.17-3.0.26
yet.anotherorigin.com2012/artifact_name 4.2.0-7.0.0
---------------------------------

*/


// 
var reMetadataDate *regexp.Regexp = regexp.MustCompile(`relish artifact metadata: (2[0-2][0-9][0-9]/[0-9][0-9]/[0-9][0-9])`)
var reCurrentVersion *regexp.Regexp = regexp.MustCompile(`current version: ([0-9]+\.[0-9]+\.[0-9]+)`)


var artifactLoaderClient *http.Client = net_util.HttpTimeoutClient(10)

/*
   Params:
   - hostUrl: http://code.example.com
   - originAndArtifactPath: some.origin.org2012/some/artifact/name
   - version: the version to fetch. If version is "", fetches the current version.
*/
func fetchArtifactZipFile(hostUrl string, originAndArtifactPath string, version string) (fileContents []byte, zipFileName string, err error) {
    var url string

    // TODO The metadata.txt file fetch should happen every time, even if version # is supplied,
    // and the metadata.txt file should replace the existing local one if the fetched one is newer. <- how determine if newer?


    slashPos := strings.Index(originAndArtifactPath, "/")
    originDomain := originAndArtifactPath[0:slashPos-4]
    hostDomain := hostUrl[7:]
    var dir string
    if originDomain == hostDomain || strings.HasPrefix(hostDomain,"localhost") {
	   dir = "artifacts"
	} else {
	   dir = "replicas"
	}
   	
  zipFileName = strings.Replace(originAndArtifactPath,"/","--",-1) + "---" + version + ".zip"

	url = fmt.Sprintf("%s/relish/%s/%s/%s",hostUrl,dir,originAndArtifactPath,zipFileName)
	
  //resp, err := http.Get(url)  
	resp, err := artifactLoaderClient.Get(url)
	if err != nil {
		return nil, zipFileName, err
	}
	defer resp.Body.Close()

  if resp.StatusCode == 404 {  // HTTP Page/Resource Not Found 
     err = fmt.Errorf("%s (v%s) not found at servers tried.", originAndArtifactPath, version)
     return nil, zipFileName, err
  } else if resp.StatusCode >= 400 {  // Other HTTP error
     err = fmt.Errorf("%s (v%s) not found or HTTP errors at servers tried.",originAndArtifactPath, version)
     return nil, zipFileName, err
  }

	body, err := ioutil.ReadAll(resp.Body)	
	// fmt.Println(len(body))
	
	return body,zipFileName,nil
}


/*
Fetches and, if newer, installs metadata.txt file in shared/relish/replicas directory.
*/
func fetchArtifactMetadata(hostUrl string, originAndArtifactPath string, localMetadataDate string, existingSharedCurrentVersion string, sharedArtifactMetadataFilePath string)  (currentVersion string, err error) {

    slashPos := strings.Index(originAndArtifactPath, "/")
    originDomain := originAndArtifactPath[0:slashPos-4]
    hostDomain := hostUrl[7:]
    var dir string
    if originDomain == hostDomain || strings.HasPrefix(hostDomain,"localhost") {
  	   dir = "artifacts"
  	} else {
  	   dir = "replicas"
  	}
     	url := hostUrl + "/relish/" + dir + "/" + originAndArtifactPath + "/metadata.txt"
      
      var resp *http.Response
      var body []byte

  	//resp, err = http.Get(url)
    resp, err = artifactLoaderClient.Get(url)    
  	if err != nil {
  		return 
  	}
  	defer resp.Body.Close()

    if resp.StatusCode == 404 {  // HTTP Page/Resource Not Found 
       err = fmt.Errorf("%s not found at servers tried.",originAndArtifactPath)
       return
    } else if resp.StatusCode >= 400 {  // Other HTTP error
       err = fmt.Errorf("%s not found or HTTP errors at servers tried.",originAndArtifactPath)
       return
    }

  	body, err = ioutil.ReadAll(resp.Body)
  	if err != nil {
  		return
  	}	

    var remoteCurrentVersion string
    var remoteMetadataDate string
    remoteCurrentVersion, remoteMetadataDate, err = readCurrentVersion(body, url)
  	if err != nil {
  		return
  	}

    if remoteMetadataDate > localMetadataDate {
        currentVersion = remoteCurrentVersion

        // Create local directory for replica artifact, if necessary
 
        var perm os.FileMode = 0777

        artifactReplicaDir := sharedArtifactMetadataFilePath[:len(sharedArtifactMetadataFilePath)-len("/metadata.txt")]

        err = gos.MkdirAll(artifactReplicaDir, perm)
        if err != nil {
           return
        }

        // Write body to local shared metadata.txt file

        err = gos.WriteFile(sharedArtifactMetadataFilePath, body, perm)
	    if err != nil {
		   return 
	    }	
     } else {
        currentVersion = existingSharedCurrentVersion	
     }	
     return
}	