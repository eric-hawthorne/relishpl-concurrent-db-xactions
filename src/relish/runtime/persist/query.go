// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package implements data persistence in the relish language environment.

package persist

// This file contains methods that transform query syntax from an object & object-attribute query language to SQL.
// Currently is possibly specific to SQLITE 3 dbs.

import (
	"fmt"
	. "relish/dbg"
	. "relish/runtime/data"
	"strings"
	"regexp"
)

// Looking for "afunc" or " aFunc" or "aFuncName123" or " aFuncName123"
//var re *regexp.Regexp = regexp.MustCompile(`[a-z][A-Za-z0-9]*     ^\(   `)

var re *regexp.Regexp = regexp.MustCompile(`([a-z][A-Za-z0-9]*)(?:$|[^(A-Za-z0-9])`)


/*
queryArgs are values to be substituted by the SQL engine into ? parameters in the where clause.
There may be zero or more of these. The number must match the number of ?s.
*/
func (db *SqliteDB) FetchN(typ *RType, oqlSelectionCriteria string, queryArgs []RObject, radius int, objs *[]RObject) (mayContainProxies bool, err error) {
	defer Un(Trace(PERSIST_TR, "FetchN", oqlSelectionCriteria, radius))
	
	var sqlQuery string
	var numPrimitiveAttrColumns int
	
    polymorphic := typ.HasSubtypes()
	
	idsOnly := (radius == 0) || polymorphic

	mayContainProxies = idsOnly
	
    sqlQuery, numPrimitiveAttrColumns, err = db.oqlWhereToSQLSelect(typ, oqlSelectionCriteria, idsOnly) 	
    if err != nil {
	      err = fmt.Errorf("Query syntax error:\n%v\n while translating selection criteria:\n\"%s\"",err, oqlSelectionCriteria)
    }	
	
	checkCache := true
	errSuffix := ""
	err = db.fetchMultiple(sqlQuery, queryArgs, idsOnly, radius, numPrimitiveAttrColumns, errSuffix, checkCache, objs) 
	return
}




/*
Converts object and object-attribute query language expressions to SQL queries.

e.g. vehicles/Car, "speed > 60"   ==> "select id from [vehicles/Vehicle] where speed > 60"

If lazy is true, the select statement selects only ids.
If false, it selects everything from all tables: the type and all of its supertypes
*/
func (db *SqliteDB) oqlWhereToSQLSelect(objType *RType, oqlWhereCriteria string, idsOnly bool) (sqlSelectQuery string, numPrimAttributeColumns int, err error) {

   // 1. Parse the OQL into an ast?

   // 2. Collect attribute references from the OQL: words not inside single-quotes.

   // 3. Search the supertypes list/lattice, to find the correct type for each attribute

   // 4. create aliases for each type table

   // 5. replace each attribute in the OQL with alias.attribute in sql query

   // 6. select from a join of the needed tables based on the attributes.	

   // Do we do this lazy or all in one?

    attributeNames := make(map[string]bool)

    // These will have to be made illegal for attribute names in relish TODO !!!
    reservedWords := map[string]bool {
	    "and" : true,
	    "or" : true,
	    "not" : true,
	    "in" : true,
	    "is" : true,
	    "null": true,
	    "desc": true,
	    "asc": true,
	    "order": true,
	    "by": true,
    }

    s := removeLiteralStrings(oqlWhereCriteria)
 
    // words := re.FindAllString(s,-1)
	matches := re.FindAllStringSubmatch(s,-1)
	words := []string{}
	for _,match := range matches {
		words = append(words, match[1])
	}


    for _,word := range words {
  	    if ! reservedWords[word] {
	       attributeNames[word] = true
        }
    } 

    tableNameAliases, aliasedAttrNames, err := db.findAliases(objType, attributeNames, idsOnly)
    if err != nil {
       return 	
    }

    // replace references to attributes (columns) with table-alias prefixed versions

    for attrName, aliasedAttrName := range aliasedAttrNames {
	   oqlWhereCriteria = strings.Replace(oqlWhereCriteria, attrName, aliasedAttrName, -1)
    }

    var first bool
    if idsOnly {
       sqlSelectQuery = "SELECT ro.id FROM RObject ro"  // TODO This is going to be an ambiguous column name
       //   first = true
    } else {

		sqlSelectQuery = "SELECT ro.id,id2,flags,typeName"
		sep := ","		
		
		for _, attr := range objType.Attributes {
			if attr.Part.Type.IsPrimitive && attr.Part.CollectionType == "" {
				if attr.Part.Type == TimeType {
				   sqlSelectQuery += sep + attr.Part.Name	+ "," + attr.Part.Name + "_loc" 
				   numPrimAttributeColumns += 2							
				} else if attr.Part.Type == ComplexType || attr.Part.Type == Complex32Type {
				   sqlSelectQuery += sep + attr.Part.Name	+ "_r," + attr.Part.Name + "_i" 	
				   numPrimAttributeColumns += 2
				} else {
				   sqlSelectQuery += sep + attr.Part.Name
				   numPrimAttributeColumns ++
			    }
				sep = ","
			}
		}		
		
		for _, typ := range objType.Up {
			for _, attr := range typ.Attributes {
					if attr.Part.Type.IsPrimitive && attr.Part.CollectionType == "" {
					if attr.Part.Type == TimeType {
					   sqlSelectQuery += sep + attr.Part.Name	+ "," + attr.Part.Name + "_loc" 
					   numPrimAttributeColumns += 2							
					} else if attr.Part.Type == ComplexType || attr.Part.Type == Complex32Type {
					   sqlSelectQuery += sep + attr.Part.Name	+ "_r," + attr.Part.Name + "_i" 	
					   numPrimAttributeColumns += 2
					} else {
					   sqlSelectQuery += sep + attr.Part.Name
					   numPrimAttributeColumns ++
				    }
					sep = ","				
				}
			}
		}				
		sqlSelectQuery += " FROM RObject ro"
    }

    for tableName,alias := range tableNameAliases {
	    tableAlias := alias
	    if tableAlias != "" {
		   tableAlias = " " + tableAlias 
	    }
	    if first {
		   sqlSelectQuery += tableName + tableAlias
	    } else {
		   sqlSelectQuery += " JOIN " + tableName + tableAlias + " USING (id)"		
	    }
	    first = false
    }
    sqlSelectQuery += " WHERE " + oqlWhereCriteria

	return
}

/*
Removes single-quoted literal strings from sql query text and returns all of the query text
except for the quotes and quoted literals. Handles double '' escapes in literals.
*/
func removeLiteralStrings(queryString string) string {
	s := ""
	strs := strings.Split(queryString, "'")
	for i, r := range strs {
		if i%2 == 0 {
			s += r
		}
	}
	return s
}

/*
Given a type, and a set of attribute names that are included in where clause criteria, return
a map from table names to table aliases, and a map from attribute names to table-alias-qualified attribute names
Returns an error if an attribute name is not found as a primitive-valued attribute in the type or its supertypes.

If lazy is true, returns only table names which hold one of the attributes.
If lay id false, returns tables for the type and all supertypes.

func (db *SqliteDB) findAliases(typ *RType, attributeNames map[string]bool, lazy bool) (tableNamesToAliases map[string]string, attrNamesToAliasedAttrNames map[string]string, err error) {
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
		          	 tableNamesToAliases[db.TableNameIfy(typ.ShortName())] = alias
	                 attrNamesToAliasedAttrNames[attrName] = alias + "." + attrName
		          	 continue QueryAttributeLoop
		          } else if ! la
			  }
		   }
	       for _, superType := range typ.Up {	
			  for _, attr := range superType.Attributes {
				  if attr.Part.Type.IsPrimitive {
			          if attr.Part.Name == attrName {
			          	 alias = fmt.Sprintf("t%v",aliasNum)
			          	 aliasNum ++
			          	 tableNamesToAliases[db.TableNameIfy(superType.ShortName())] = alias
		                 attrNamesToAliasedAttrNames[attrName] = alias + "." + attrName
			          	 continue QueryAttributeLoop
			          }
			      }
			  }	       	
	       }

           // attribute name from query was not matched by a primitive-valued attribute of the type or its supertypes.
	       err = fmt.Errorf("'%s' is not a primitive-valued attribute of type %v or its supertypes",attrName,typ)
	       return	   
	   }
	return
}
*/

func (db *SqliteDB) findAliases(typ *RType, attributeNames map[string]bool, idsOnly bool) (tableNamesToAliases map[string]string, attrNamesToAliasedAttrNames map[string]string, err error) {
   tableNamesToAliases = make(map[string]string)
   attrNamesToAliasedAttrNames =  make(map[string]string)
   aliasNum := 1
   var alias string
   typeTableName := db.TableNameIfy(typ.ShortName())
   var foundAttribute bool
   var matchedAttribute bool
   for _, attr := range typ.Attributes {
	  if attr.Part.Type.IsPrimitive && attr.Part.CollectionType == "" {
		 foundAttribute = true // type has at least one primitive attribute
		 alias = tableNamesToAliases[typeTableName]

         for attrName := range attributeNames {
	        if attr.Part.Name == attrName {	
		       matchedAttribute = true
		       if alias == "" {
			      alias = fmt.Sprintf("t%v",aliasNum)
		          aliasNum ++
		          tableNamesToAliases[typeTableName] = alias
		       }		
	           attrNamesToAliasedAttrNames[attrName] = alias + "." + attrName	
	        }
	     }		
	  }
   }

   if (foundAttribute && ! idsOnly) || matchedAttribute {
      tableNamesToAliases[typeTableName] = alias	
   }

   for _, superType := range typ.Up {	
	  foundAttribute = false
	  matchedAttribute = false
	  alias = ""
      typeTableName = db.TableNameIfy(superType.ShortName())	
	  for _, attr := range superType.Attributes {
		  if attr.Part.Type.IsPrimitive && attr.Part.CollectionType == "" {
		     foundAttribute = true // type has at least one primitive attribute			
			 alias = tableNamesToAliases[typeTableName]
	         for attrName := range attributeNames {
		        if attr.Part.Name == attrName {		
			       matchedAttribute = true
			       if alias == "" {
				      alias = fmt.Sprintf("t%v",aliasNum)
			          aliasNum ++
			          tableNamesToAliases[typeTableName] = alias
			       }			
		           attrNamesToAliasedAttrNames[attrName] = alias + "." + attrName	
		        }
		     }
	      }
	  }	
	  if (foundAttribute && ! idsOnly) || matchedAttribute {
	     tableNamesToAliases[typeTableName] = alias	
	  }	      	
   }

   for attrName := range attributeNames {
	   if attrNamesToAliasedAttrNames[attrName] == "" {
           // attribute name from query was not matched by a primitive-valued attribute of the type or its supertypes.
	       err = fmt.Errorf("'%s' is not a primitive-valued attribute of type %v or its supertypes",attrName,typ)
	       return	   
	   }
   }
   return
}



