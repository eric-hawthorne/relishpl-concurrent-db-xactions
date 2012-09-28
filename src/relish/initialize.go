package relish

import (
	. "relish/runtime/data"
	. "relish/runtime/persist"	
)	
	
func DatabaseURI() string {
   return RT.DatabaseURI
}

func SetDatabaseURI(uri string) {
   if RT.DB() == nil {	
      RT.DatabaseURI = uri
   } else if uri != RT.DatabaseURI {
   	   panic("Cannot rename database after connection to database has been initialized.")
   }
}	

func EnsureDatabase() {	
   if RT.DB() == nil {
      db := NewDB(RT.DatabaseURI)
      RT.SetDB(db)
   }
}



