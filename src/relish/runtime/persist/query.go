// Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package implements data persistence in the relish language environment.

package persist

// This file contains methods that transform query syntax from an object & object-attribute query language to SQL.
// Currently is possibly specific to SQLITE 3 dbs.

import (
	"fmt"
	. "relish/dbg"
	. "relish/runtime/data"
)

	/*
	Converts object and object-attribute query language expressions to SQL queries.

	e.g. vehicles/Car, "speed > 60"   ==> "select id from [vehicles/Vehicle] where speed > 60"
	*/
func (db *SqliteDB) OQLWhereToSQLSelect(typ *RType, oqlWhereCriteria string) (sqlSelectQuery String, err Error) {

   // 1. Parse the OQL into an ast?

   // 2. Collect attribute references from the OQL: words not inside single-quotes.

   // 3. Search the supertypes list/lattice, to find the correct type for each attribute

   // 4. create aliases for each type table

   // 5. replace each attribute in the OQL with alias.attribute in sql query

   // 6. select from a join of the needed tables based on the attributes.	

   // Do we do this lazy or all in one?
}

/*
Given a type, and a set of attribute names that are included in where clause criteria, return
a map from table names to table aliases, and a map from attribute names to table-alias-qualified attribute names
Returns an error if an attribute name is not found as a primitive-valued attribute in the type or its supertypes.
*/
func (db *SqliteDB) findAliases(typ *RType, attributeNames map[string]bool) (tableNamesToAliases map[string]string, attrNamesToAliasedAttrNames map[string]string, err Error) {
   tableNamesToAliases = make(map[string]string)
   attrNamesToAliasedAttrNames =  make(map[string]string)
   aliasNum := 1
   var alias string
   QueryAttributeLoop:
	   for attrName := range attributeNames {
		   for _, attr := range typ.Attributes {
			  if attr.Part.Type.IsPrimitive {
		          if attr.Part.Name == attrName {
		          	 alias = fmt.Sprintf("t%v",aliasNum)
		          	 aliasNum ++
		          	 tableNamesToAliases[tableNameIfy(typ.ShortName())] = alias
	                 attrNamesToAliasedAttrNames[attrName] = alias + "." + attrName
		          	 continue QueryAttributeLoop
		          }
			  }
		   }
	       for _, superType := range typ.Up {	
			  for _, attr := range superType.Attributes {
				  if attr.Part.Type.IsPrimitive {
			          if attr.Part.Name == attrName {
			          	 alias = fmt.Sprintf("t%v",aliasNum)
			          	 aliasNum ++
			          	 tableNamesToAliases[tableNameIfy(superType.ShortName())] = alias
		                 attrNamesToAliasedAttrNames[attrName] = alias + "." + attrName
			          	 continue QueryAttributeLoop
			          }
			      }
			  }	       	
	       }	   
	   }
	return
}




