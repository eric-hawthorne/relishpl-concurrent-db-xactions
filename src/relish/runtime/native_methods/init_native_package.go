// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU LESSER GPL v3 license, found in the LICENSE_LGPL3 file.

package native_methods

/*
   init_native_package.go 
*/

import (
	"relish/runtime/native_methods/standard_lib/files_methods"
	"relish/runtime/native_methods/standard_lib/http_methods"	
   "relish/runtime/native_methods/extensions/protocols/modbus_methods"
)

/*
IMPORTANT

If you add a package which has some Go native methods implemented in go packages under the
relish/runtime/native_methods/standard_lib directory or the  
relish/runtime/native_methods/extensions directory,
you must include a wrapper-creating function with the set of go methods, 
such as the files.InitFilesMethods method, and you must create an entry
in the map below recording the association of the full package path (the package identifier)
with the corresponding wrapper-creating function.

*/
var nativeMethodPackageMap = map [string] func() {
	"shared.relish.pl2012/relish_lib/pkg/files" : files_methods.InitFilesMethods,
	"shared.relish.pl2012/relish_lib/pkg/http_srv" : http_methods.InitHttpMethods,	
    "gait.bcit.ca2012/protocols/pkg/modbus" : modbus_methods.InitModbusMethods,   
}


/*
   Should be called only at the end of global_loader.LoadPackage(...)
   (So only called once per loaded package)

   Checks a map to see if the relish package has any native methods.
   If so, calls the function which initializes the package's native methods by
   creating a relish RMethod wrapper object for each Go method.
*/
func WrapNativeMethods(fullPackagePath string) {
   nativeMethodWrapper, packageHasNativeMethods := nativeMethodPackageMap[fullPackagePath]
   if packageHasNativeMethods {
   	  nativeMethodWrapper()
   }
}


