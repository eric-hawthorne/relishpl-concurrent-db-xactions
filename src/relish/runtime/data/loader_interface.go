// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.

package data

/*
   loader_interface.go - Abstraction of code-package loading service.
*/


type PackageLoader interface {
	/*
	Loads the code package into the runtime. May fetch it from the Internet first.
	Returns nil if it succeeds.
	Version may be "" in which case the current version of the package, or the same version
	as other packages that have been loaded from the artifact, is used. TODO IS THAT LATTER BIT TRUE????
	*/
	LoadRelishCodePackage (originAndArtifactPath string, version string, packagePath string, mustBeFromShared bool) (err error) 
}