// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
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
	"errors"
)

// Looking for "afunc" or " aFunc" or "aFuncName123" or " aFuncName123"
//var re *regexp.Regexp = regexp.MustCompile(`[a-z][A-Za-z0-9]*     ^\(   `)

// Before May 20 2014
//var re *regexp.Regexp = regexp.MustCompile(`([a-z][A-Za-z0-9]*)(?:$|[^(A-Za-z0-9])`)

var re *regexp.Regexp = regexp.MustCompile(`(?:^|[^.A-Za-z0-9])([a-z][A-Za-z0-9]*)(?:$|[^(.A-Za-z0-9])`)	
var re2 *regexp.Regexp = regexp.MustCompile(`([a-z][A-Za-z0-9]*)\.([a-z][A-Za-z0-9]*)(?:$|[^(A-Za-z0-9])`)

/*
queryArgs are values to be substituted by the SQL engine into ? parameters in the where clause.
There may be zero or more of these. The number must match the number of ?s.
*/
func (db *SqliteDB) FetchN(typ *RType, oqlSelectionCriteria string, queryArgs []RObject, coll RCollection, radius int, objs *[]RObject) (mayContainProxies bool, err error) {
	defer Un(Trace(PERSIST_TR, "FetchN", oqlSelectionCriteria, radius))
	
	var sqlQuery string
	var numPrimitiveAttrColumns int
	
    polymorphic := typ.HasSubtypes()
	
	idsOnly := (radius == 0) || polymorphic

	mayContainProxies = idsOnly
	
    sqlQuery, numPrimitiveAttrColumns, err = db.oqlWhereToSQLSelect(typ, oqlSelectionCriteria, coll, idsOnly) 	
    if err != nil {
       if strings.HasPrefix(err.Error(), "In asList") {
	      err = fmt.Errorf("%v\n  (call to asList with selection criteria:\n   \"%s\")",err, oqlSelectionCriteria)
       } else {
	      err = fmt.Errorf("Query syntax error:\n%v\n while translating selection criteria:\n\"%s\"",err, oqlSelectionCriteria)
       }
	   return
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
func (db *SqliteDB) oqlWhereToSQLSelect(objType *RType, oqlWhereCriteria string, coll RCollection, idsOnly bool) (sqlSelectQuery string, numPrimAttributeColumns int, err error) {

   // 1. Parse the OQL into an ast?

   // 2. Collect attribute references from the OQL: words not inside single-quotes.

   // 3. Search the supertypes list/lattice, to find the correct type for each attribute

   // 4. create aliases for each type table

   // 5. replace each attribute in the OQL with alias.attribute in sql query

   // 6. select from a join of the needed tables based on the attributes.	

   // Do we do this lazy or all in one?

    attributeNames := make(map[string]bool)
    otherAttributeNames := make(map[string][]string)  // map from join attr name to list of other attr names
    joinAttrs := make(map[string]*AttributeSpec)

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
	    "like": true,
	    "order": true,
	    "by": true,
    }

    s := removeLiteralStrings(oqlWhereCriteria)
 
    // Find b.c and d.e in expression like b.c < 2 and d.e = 'foo'
    //
	joinedAttrMatches := re2.FindAllStringSubmatch(s,-1)
	
    hasJoinedConditions := len(joinedAttrMatches) > 0


	joinAttrWords := []string{}
	joinAttrOtherAttrWords := []string{}	
	for _,match := range joinedAttrMatches {
		joinAttrWords = append(joinAttrWords, match[1])
		joinAttrOtherAttrWords = append(joinAttrOtherAttrWords, match[2])		
	}

    for i := range joinAttrWords {
        joinAttrWord := joinAttrWords[i]
        joinAttrOtherAttrWord := joinAttrOtherAttrWords[i]

        otherAttrs,othersFound := otherAttributeNames[joinAttrWord]
        if othersFound {
        	otherAttributeNames[joinAttrWord] = append(otherAttrs,joinAttrOtherAttrWord)
        } else {
        	otherAttributeNames[joinAttrWord] = []string{joinAttrOtherAttrWord}
        	attr, attrFound := objType.GetAttribute(joinAttrWord)
            if ! attrFound {
            	err = fmt.Errorf("Attribute '%s' not found in type %s or supertypes.", joinAttrWord, objType.Name)
                return    
            }    	
        	joinAttrs[joinAttrWord] = attr
        }
    }











    // Original code, before join-table handling, is below here.

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

    // tableNameAliases is a map from tableName to alias
    // aliasedAttrNames has each attribute b prefixed by its table's alias e.g. t1.b
    tableNameAliases, aliasedAttrNames, err := db.findAliases(objType, attributeNames, idsOnly, hasJoinedConditions)
    if err != nil {
       return 	
    }

    var literalMap map[string]string 
    oqlWhereCriteria, literalMap = substituteLiteralStrings(oqlWhereCriteria) 

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
		

// Question: What are the right t1, t2 ... table aliases to prepend if hasJoinedConditions ?

        var tableName string 
        var alias string

        if hasJoinedConditions {
        	tableName = db.TableNameIfy(objType.ShortName())
        	alias = tableNameAliases[tableName] + "."
        }

		for _, attr := range objType.Attributes {

			if attr.Part.Type.IsPrimitive && attr.Part.CollectionType == "" {
				if attr.Part.Type == TimeType {
				   sqlSelectQuery += sep + alias + attr.Part.Name + "," + attr.Part.Name + "_loc" 
				   numPrimAttributeColumns += 2							
				} else if attr.Part.Type == ComplexType || attr.Part.Type == Complex32Type {
				   sqlSelectQuery += sep + alias + attr.Part.Name	+ "_r," + attr.Part.Name + "_i" 	
				   numPrimAttributeColumns += 2
				} else {
				   sqlSelectQuery += sep + alias + attr.Part.Name
				   numPrimAttributeColumns ++
			    }
				sep = ","
			}
		}		
		
		for _, typ := range objType.Up {

	        if hasJoinedConditions {
	        	tableName = db.TableNameIfy(typ.ShortName())
	        	alias = tableNameAliases[tableName]
	        } else {
	        	alias = ""
	        }

			for _, attr := range typ.Attributes {
					if attr.Part.Type.IsPrimitive && attr.Part.CollectionType == "" {
					if attr.Part.Type == TimeType {
					   sqlSelectQuery += sep + alias + attr.Part.Name	+ "," + attr.Part.Name + "_loc" 
					   numPrimAttributeColumns += 2							
					} else if attr.Part.Type == ComplexType || attr.Part.Type == Complex32Type {
					   sqlSelectQuery += sep + alias + attr.Part.Name	+ "_r," + attr.Part.Name + "_i" 	
					   numPrimAttributeColumns += 2
					} else {
					   sqlSelectQuery += sep + alias + attr.Part.Name
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
	       if hasJoinedConditions {
		      sqlSelectQuery += " JOIN " + tableName + tableAlias + " ON ro.id =" + tableAlias + ".id"			       	   
	       } else {	
		      sqlSelectQuery += " JOIN " + tableName + tableAlias + " USING (id)"		
		   }
	    }
	    first = false
    }

    collectionMembershipWhereFilter := ""

    if coll != nil {
       var collectionId int64
       var collectionTableName string	
       if coll.Owner() == nil { // Independent persistent collection
       	  if ! coll.IsStoredLocally() {
       	  	 err = errors.New("In asList with OQL query, the collection must be persistent!")
       	  	 return
       	  }
          collectionId = coll.DBID()
          collectionTableName,_,_,_,_  = db.TypeDescriptor(coll)          
       } else {
       	  if ! coll.Owner().IsStoredLocally() {
       	  	 err = errors.New("In asList with OQL query, the object with multi-valued attribute must be persistent!")
       	  	 return
       	  }       	
          collectionId = coll.Owner().DBID()
          collectionTableName = db.TableNameIfy(coll.Attribute().ShortName())
       }

       sqlSelectQuery += " JOIN " + collectionTableName + " ctbl ON ro.id = ctbl.id1"		

       collectionMembershipWhereFilter = fmt.Sprintf("ctbl.id0 = %d AND ",collectionId)     
    }



/////////////// JOINED OTHER OBJECT ATTR TABLES

    // otherAttributeNames is a map from joinAttrName to list of other obj primitive attrNames
    // joinAttrs is a map form joinAttrName to actual attribute object (from which can get table name)

    // THE FOLLOWING MIGHT HAVE TO BE MOVED LATER!!

    otherAttrNamesToAliasedAttrNames :=  make(map[string]string)

    i := 0
    j := 0
    for joinAttrName, joinAttr := range joinAttrs {
        i++
    	otherAttrNames := otherAttributeNames[joinAttrName]

    	joinTable := db.TableNameIfy(joinAttr.ShortName())
    	joinTableAlias := fmt.Sprintf("jt%d",i)


        sqlSelectQuery += " JOIN " + joinTable + " " + joinTableAlias + " ON ro.id = " + joinTableAlias + ".id0" 

        joinedObjectType := joinAttr.Part.Type




        otherObjectTableAliases := make ( map[string]string )   // table name to alias

    	// find table name for each other attrName, given joinAttr.Part.Type

    	for _,otherAttrName := range otherAttrNames {
            otherAttr, attrFound := joinedObjectType.GetAttribute(otherAttrName)
            if ! attrFound {
            	err = fmt.Errorf("Attribute '%s' not found in attribute type %s or supertypes.", otherAttrName, joinedObjectType.Name)
                return 
            }
            otherObjectTypeWithAttr := otherAttr.WholeType
            otherTableName := db.TableNameIfy(otherObjectTypeWithAttr.ShortName())
            otherTableAlias, found := otherObjectTableAliases[otherTableName]
            if ! found {
            	j++
    	        otherTableAlias = fmt.Sprintf("ot%d",j) 
    	        otherObjectTableAliases[otherTableName] = otherTableAlias      

               sqlSelectQuery += " JOIN " + otherTableName + " " + otherTableAlias + " ON " + joinTableAlias + ".id1 = " + otherTableAlias + ".id" 

            }

            qualifiedOtherAttrName := joinAttrName + "." + otherAttrName
            otherAttrNamesToAliasedAttrNames[qualifiedOtherAttrName] = otherTableAlias + "." + otherAttrName
    	} 
    }




    // replace references to attributes (columns) with table-alias prefixed versions
 
    for qualifiedOtherAttrName, aliasedOtherAttrName := range otherAttrNamesToAliasedAttrNames {
	   oqlWhereCriteria = strings.Replace(oqlWhereCriteria, qualifiedOtherAttrName, aliasedOtherAttrName, -1)
    }

    oqlWhereCriteria = restoreLiteralStrings(oqlWhereCriteria, literalMap)



/////////////// END OF JOINED OTHER OBJECT ATTR TABLES








    sqlSelectQuery += " WHERE " + collectionMembershipWhereFilter + oqlWhereCriteria 

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

func substituteLiteralStrings(queryString string) (subbed string, literalMap map[string]string) {
	literalMap = make(map[string]string)
	strs := strings.Split(queryString, "'")
	for i, r := range strs {
		if i%2 == 0 {
			subbed += r
		} else {
			token := fmt.Sprintf("'%d'",i)

			subbed += token
			literalMap[token] = "'" + r + "'"
		}
	}
	return 
}

func restoreLiteralStrings(queryString string, literalMap map[string]string) string {
	for token,original := range literalMap {
       queryString = strings.Replace(queryString, token, original, 1)
	}
	return queryString
}

/*
Given a type, and a set of attribute names that are included in where clause criteria, return
a map from table names to table aliases, and a map from attribute names to table-alias-qualified attribute names
Returns an error if an attribute name is not found as a primitive-valued attribute in the type or its supertypes.

If idsOnly is true, returns only table names which hold one of the attributes, plus the tableName for the argument type.
If idsOnly is false, returns tables for the type and all supertypes.

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

func (db *SqliteDB) findAliases(typ *RType, attributeNames map[string]bool, idsOnly bool, hasJoinedCondition bool) (tableNamesToAliases map[string]string, attrNamesToAliasedAttrNames map[string]string, err error) {
   tableNamesToAliases = make(map[string]string)
   attrNamesToAliasedAttrNames =  make(map[string]string)
   aliasNum := 1
   var alias string
   typeTableName := db.TableNameIfy(typ.ShortName())
   var foundAttribute bool  // Whether type has at least one primitive attribute
   var matchedAttribute bool   // Whether an attribute from the type is mentioned in the where conditions of the query
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

   if hasJoinedCondition && alias == "" {
	  alias = fmt.Sprintf("t%v",aliasNum)
	  aliasNum ++   	   
   }
   tableNamesToAliases[typeTableName] = alias  // original type must always be included in the join. 	
   /*
   if (foundAttribute && ! idsOnly) || matchedAttribute {
      tableNamesToAliases[typeTableName] = alias	
   }
   */

   // TODO: Need to ensure that if idsOnly, the argument typ is included in tableNamesToAliases

   for _, superType := range typ.Up {	
	  foundAttribute = false   // Whether at least one attribute of this type has been found in the query expression
	  matchedAttribute = false // Whether an attribute from the type is mentioned in the where conditions of the query
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
	     if hasJoinedCondition && alias == "" {
		     alias = fmt.Sprintf("t%v",aliasNum)
		     aliasNum ++   	   
	     }
	     tableNamesToAliases[typeTableName] = alias	
	  }	      	
   }

   // Make sure each (non-join) attribute in select conditions is a primitive attribute of the object type
   // being selected or of one of the object type's supertypes.
   for attrName := range attributeNames {
	   if attrNamesToAliasedAttrNames[attrName] == "" {
           // attribute name from query was not matched by a primitive-valued attribute of the type or its supertypes.
	       err = fmt.Errorf("'%s' is not a primitive-valued attribute of type %v or its supertypes",attrName,typ)
	       return	   
	   }
   }
   return
}



