// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// This package serves as a dependency reduction bridge between other packages that need to
// use the methods in this package but don't want to depend on the packages that implement
// these functions.

// Database initialization functions.

package relish

import (
	. "relish/runtime/data"
	. "relish/runtime/persist"	
   "relish/rterr"
)	
	
func DatabaseURI() string {
   return RT.DatabaseURI
}

func SetDatabaseURI(uri string) {
   if RT.DB() == nil {	
      RT.DatabaseURI = uri
   } else if uri != RT.DatabaseURI {
   	   rterr.Stop("Cannot rename database after connection to database has been initialized.")
   }
}	

func EnsureDatabase() {	
   if RT.DB() == nil {
      db := NewDB(RT.DatabaseURI)
      RT.SetDB(db)
   }
}



