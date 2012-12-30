// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// loads relish code artifacts from the Internet, assuming strong choice-free conventions
// about where a given version of a given artifact will reside or how it will be found, 
// both on the Internet as a whole, and on each server that hosts it.

package global_loader

import (
    "fmt"
    "net/http"
	"io/ioutil"
    "regexp"
//	"bytes"
//    "strings"
    "strconv"
    "errors"
//	. "relish/runtime/data"
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
http://register.relishing.org. This ensures origin uniqueness. Originators can also register each software artifact 
there if they wish. Doing so will ensure their software appears in a directory of relish code artifacts.
 
An artifact must be named according to the following conventions:
some_artifact_name or some/long_path/to/some_artifact_name

Each publicly released relish code artifact should be hosted at (by example):
http://mysuborg.example.org/relish/artifacts/some/long_path/to/some_artifact_name 
but in any case, must be hosted on some Google-searchable website at the path:
/relish/artifacts/some/long_path/to/some_artifact_name 
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
origin: some.origin.org2012
artifact: some/artifact/name
current version: 2
release date: 2012/03/27
standard tags: example tag, second example tag
more tags:

Start the 500 word or less summary on this line.
Continue it like this.
---------------------------------


v0002/built.txt  - metadata listing the version of each dependency which was used in the latest successful test execution of this
                   version of this artifact's test suite.

---------------------------------
some.otherorigin.com2011/some/artifact_name 26
yet.anotherorigin.com2012/artifact_name 7
---------------------------------


v0002/dependencies.txt  - metadata listing the versions of each dependency which this version of this artifact have been tested against successfully.
                          Note that this does not mean that all combinations of the listed versions of dependency A have been tested 
                          with all listed versions of dependency B. 

---------------------------------
some.otherorigin.com2011/some/artifact_name 2-14,17-26
yet.anotherorigin.com2012/artifact_name 4-7
---------------------------------

*/

// 
var re *regexp.Regexp = regexp.MustCompile(`current version: ([0-9]*)`)

/*
   Params:
   - hostUrl: http://code.example.com
   - originAndArtifactPath: some.origin.org2012/some/artifact/name
   - version: the version to fetch. If version is zero, fetches the current version.
*/
func fetchArtifactZipFile(hostUrl string, originAndArtifactPath string, version int) (fileContents []byte, err error) {
    var url string

    // TODO The metadata.txt file fetch should happen every time, even if version # is supplied.
    // and the metadata.txt file should replace the existing local one if the fetched one is newer. <- how determine if newer?

	if version == 0 {
	   	url = hostUrl + "/relish/artifacts/" + originAndArtifactPath + "/metadata.txt"
	    
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)	

	    match := re.FindSubmatchIndex(body)
        if match == nil {
	       return nil, errors.New("metadata.txt file must have a line like current version: 14")
        }
	    versionNumStart := match[2]
	    versionNumEnd := match[3]	
	    
	    s := string(body[versionNumStart:versionNumEnd])
	
        v64, err := strconv.ParseInt(s, 0, 32)  
	    if err != nil {
		   return nil, err
	    }
	
	    version = int(v64)	
	}
	
	// TODO %4d or whatever v0014
	url = fmt.Sprintf("%s/relish/artifacts/%s/v%s.zip",hostUrl,originAndArtifactPath,version)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)	
	fmt.Println(len(body))
	
	return body,nil
}

