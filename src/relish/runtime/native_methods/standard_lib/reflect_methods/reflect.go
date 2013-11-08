// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU LESSER GPL v3 license, found in the LICENSE_LGPL3 file.

package reflect_methods

/*
   reflect.go - native methods for reflection into language structure (metadata) of relish data.
   These methods are used by types defined in the relish standard library 'reflect' package. 
*/

import (
	. "relish/runtime/data"
    "fmt"
    "sort"
)

///////////
// Go Types

// None so far


/////////////////////////////////////
// relish method to go method binding

func InitReflectMethods() {

    // typeNames > [] String
    // """
    //  Should be alphabetical.
    //
    //  Should I have options like
    //   exclude builtin, exclude relish lib, exclude primitive, exclude collection types
    // """
    //
	typeNamesMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"typeNames", []string{}, []string{}, []string{"List_of_String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	typeNamesMethod.PrimitiveCode = typeNames


    // type name String > ?DataType
	typeMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"type", []string{"name"}, []string{"String"}, []string{"shared.relish.pl2012/relish_lib/pkg/reflect/DataType"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	typeMethod.PrimitiveCode = typ

    // attributeNames d dataType includeInherited Bool > [] String
	attributeNamesMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"attributeNames", []string{"d"}, []string{"shared.relish.pl2012/relish_lib/pkg/reflect/DataType"}, []string{"List_of_String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	attributeNamesMethod.PrimitiveCode = attributeNames 

    // attribute d DataType attributeName String > ?Attribute 
	attributeMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"attribute", 
		                                    []string{"d","attributeName"}, 
		                                    []string{"shared.relish.pl2012/relish_lib/pkg/reflect/DataType","String"}, 
		                                    []string{"shared.relish.pl2012/relish_lib/pkg/reflect/Attribute"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	attributeMethod.PrimitiveCode = attribute



	// typeOf a Any > DataType
	typeOfMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"typeOf", 
		                                    []string{"a"}, 
		                                    []string{"Any"}, 
		                                    []string{"shared.relish.pl2012/relish_lib/pkg/reflect/DataType"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	typeOfMethod.PrimitiveCode = typeOf

	// isa a Any d DataType > Bool
	isaMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"isa", 
		                                    []string{"a","d"}, 
		                                    []string{"Any","shared.relish.pl2012/relish_lib/pkg/reflect/DataType"}, 
		                                    []string{"Bool"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	isaMethod.PrimitiveCode = isa

   
	// supertypes d DataType > [] DataType
	// """
	//  Direct supertypes of d
	// """
	supertypesMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"supertypes", 
		                                    []string{"d","DataType"}, 
		                                    []string{"shared.relish.pl2012/relish_lib/pkg/reflect/DataType"}, 
		                                    []string{"List"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	supertypesMethod.PrimitiveCode = supertypes

	// subtypes d DataType > [] DataType
	// """
	//  Direct subtypes of d
	// """
	subtypesMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"subtypes", 
		                                    []string{"d","DataType"}, 
		                                    []string{"shared.relish.pl2012/relish_lib/pkg/reflect/DataType"}, 
		                                    []string{"List"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	subtypesMethod.PrimitiveCode = subtypes 


	// supertypeClosure d DataType > [] DataType
	// """
	//  All direct and indirect supertypes of d, not including d itself.
	// """
	supertypeClosureMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"supertypeClosure", 
		                                    []string{"d","DataType"}, 
		                                    []string{"shared.relish.pl2012/relish_lib/pkg/reflect/DataType"}, 
		                                    []string{"List"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	supertypeClosureMethod.PrimitiveCode = supertypeClosure

	// subtypeClosure d DataType > [] DataType
	// """
	//  All direct and indirect subtypes of d, not including d itself.
	// """
	subtypeClosureMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"subtypeClosure", 
		                                    []string{"d","DataType"}, 
		                                    []string{"shared.relish.pl2012/relish_lib/pkg/reflect/DataType"}, 
		                                    []string{"List"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	subtypeClosureMethod.PrimitiveCode = subtypeClosure 	


	// attrVal obj Any attr Attribute > val Any found Bool
	// """
	//  The value of the specified attribute of the object. May be a collection.
	//  If the object instance has no defined value for the specifeid attribute, found is false.
	// """
	attrValMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"attrVal", 
		                                    []string{"obj","attr"}, 
		                                    []string{"Any","shared.relish.pl2012/relish_lib/pkg/reflect/Attribute"}, 
		                                    []string{"Any","Bool"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	attrValMethod.PrimitiveCode = attrVal  	

}




 
///////////////////////////////////////////////////////////////////////////////////////////
// Reflection functions


    // typeNames > [] String
    // """
    //  Should be alphabetical.
    //
    //  Should I have options like
    //   exclude builtin, exclude relish lib, exclude primitive, exclude collection types
    // """
    //
func typeNames(th InterpreterThread, objects []RObject) []RObject {

    typeNameList, err := RT.Newrlist(StringType, 0, -1, nil, nil)
    if err != nil {
	   panic(err)
    }

    var typeNameSlice []string
    for typeName := range RT.Types {
    	typeNameSlice = append(typeNameSlice, typeName)
    } 
    sort.Strings(typeNameSlice)

    for _,typeName := range typeNameSlice {
    	typeNameList.AddSimple(String(typeName))
    }

	return []RObject{typeNameList}
}


    // type name String > ?DataType
//
func typ(th InterpreterThread, objects []RObject) []RObject {
	
	name := string(objects[0].(String))

	datatype,err := ensureDataType(name) 
	if err != nil {
		datatype = NIL
	}
	return []RObject{datatype}
}


    // attributeNames d dataType includeInherited Bool > [] String
//
func attributeNames(th InterpreterThread, objects []RObject) []RObject {
	
	datatype := objects[0].(*GoWrapper)
	t := datatype.GoObj.(*RType)
	includeInherited := bool(objects[1].(Bool))



    attrNameList, err := RT.Newrlist(StringType, 0, -1, nil, nil)
    if err != nil {
	   panic(err)
    }

    var attrNameSlice []string

	for attrName := range t.AttributesByName {
		attrNameSlice = append(attrNameSlice, attrName)
	}
	if includeInherited {
		for _, supertype := range t.Up {
	       for attrName := range supertype.AttributesByName {
		      attrNameSlice = append(attrNameSlice, attrName)
	       }
		}		
	}

    sort.Strings(attrNameSlice)

    for _,attrName := range attrNameSlice {
    	attrNameList.AddSimple(String(attrName))
    }

	return []RObject{attrNameList}
}


    // attribute d DataType attributeName String > Attribute
//
func attribute(th InterpreterThread, objects []RObject) []RObject {
	
	datatype := objects[0].(*GoWrapper)
	t := datatype.GoObj.(*RType)

	attributeName := string(objects[1].(String))

    attribute, err := ensureAttribute(t, attributeName)
	if err != nil {
		attribute = NIL
	}
	return []RObject{attribute}
}


	// typeOf a Any > DataType
//
func typeOf(th InterpreterThread, objects []RObject) []RObject {
	
    t := objects[0].Type()

	datatype,err := ensureDataType(t.Name) 
	if err != nil {
		panic(err)
	}

	return []RObject{datatype}
}


	// isa a Any d DataType > Bool
func isa(th InterpreterThread, objects []RObject) []RObject {
	
    t := objects[0].Type()
    datatype := objects[1].(*GoWrapper)	
	t2 := datatype.GoObj.(*RType)
	typeCompatible := t.LessEq(t2)
	
	return []RObject{Bool(typeCompatible)}
}



	// supertypes d DataType > [] DataType
	// """
	//  Direct supertypes of d
	// """
func supertypes(th InterpreterThread, objects []RObject) []RObject {
	
    datatype := objects[0].(*GoWrapper)	
	t := datatype.GoObj.(*RType)

    dataTypeType, typFound := RT.Types["shared.relish.pl2012/relish_lib/pkg/reflect/DataType"]
    if ! typFound {
    	panic("reflect.DataType is not defined.")
    }
    supertypeList, err := RT.Newrlist(dataTypeType, 0, -1, nil, nil)
    if err != nil {
	   panic(err)
    }

    for _,st := range t.Parents {

	   datatype,err := ensureDataType(st.Name) 
	   if err != nil {
		   panic(err)
	   }    	
       supertypeList.AddSimple(datatype)
    }

	return []RObject{supertypeList}
}






	// subtypes d DataType > [] DataType
	// """
	//  Direct subtypes of d
	// """
func subtypes(th InterpreterThread, objects []RObject) []RObject {
	
    datatype := objects[0].(*GoWrapper)	
	t := datatype.GoObj.(*RType)

    dataTypeType, typFound := RT.Types["shared.relish.pl2012/relish_lib/pkg/reflect/DataType"]
    if ! typFound {
    	panic("reflect.DataType is not defined.")
    }
    subtypeList, err := RT.Newrlist(dataTypeType, 0, -1, nil, nil)
    if err != nil {
	   panic(err)
    }

    for _,st := range t.Children {

	   datatype,err := ensureDataType(st.Name) 
	   if err != nil {
		   panic(err)
	   }    	
       subtypeList.AddSimple(datatype)
    }

	return []RObject{subtypeList}
}



	// supertypeClosure d DataType > [] DataType
	// """
	//  All direct and indirect supertypes of d, not including d itself.
	// """
func supertypeClosure(th InterpreterThread, objects []RObject) []RObject {
	
    datatype := objects[0].(*GoWrapper)	
	t := datatype.GoObj.(*RType)

    dataTypeType, typFound := RT.Types["shared.relish.pl2012/relish_lib/pkg/reflect/DataType"]
    if ! typFound {
    	panic("reflect.DataType is not defined.")
    }
    supertypeList, err := RT.Newrlist(dataTypeType, 0, -1, nil, nil)
    if err != nil {
	   panic(err)
    }

    for _,st := range t.Up {

	   datatype,err := ensureDataType(st.Name) 
	   if err != nil {
		   panic(err)
	   }    	
       supertypeList.AddSimple(datatype)
    }

	return []RObject{supertypeList}
}

	// subtypeClosure d DataType > [] DataType
	// """
	//  All direct and indirect subtypes of d, not including d itself.
	// """
func subtypeClosure(th InterpreterThread, objects []RObject) []RObject {
	
    datatype := objects[0].(*GoWrapper)	
	t := datatype.GoObj.(*RType)

    dataTypeType, typFound := RT.Types["shared.relish.pl2012/relish_lib/pkg/reflect/DataType"]
    if ! typFound {
    	panic("reflect.DataType is not defined.")
    }
    subtypeList, err := RT.Newrlist(dataTypeType, 0, -1, nil, nil)
    if err != nil {
	   panic(err)
    }

    for _,st := range t.SubtypeClosure() {

	   datatype,err := ensureDataType(st.Name) 
	   if err != nil {
		   panic(err)
	   }    	
       subtypeList.AddSimple(datatype)
    }

	return []RObject{subtypeList}
}


	// attrVal obj Any attr Attribute > val Any found Bool
	// """
	//  The value of the specified attribute of the object. May be a collection.
	// """
func attrVal(th InterpreterThread, objects []RObject) []RObject {
	
	obj := objects[0]
	attribute := objects[1].(*GoWrapper)	
	attr := attribute.GoObj.(*AttributeSpec)
		
    val, found := RT.AttrVal(obj, attr)
    if ! found {
    	val = NIL
    } 

	return []RObject{val, Bool(found)}
}




///////////////////////////////////////////////////////////////////////////////////////////
// Type init functions

/*
  Given the full name of a relish datatype, ensure that there is a reflect.DataType object
  for it and return that object. There is a singleton reflect.DataType for each RType whose
  reflected DataType has been ensured.
  A hashtable in the RunTime is first searched, then if the reflect.DataType is not found,
  a new one is created and put in the RunTime hashtable.
  The reflect.DataType is a *GoWrapper with its GoObj pointing to the underlying *RType.
*/
func ensureDataType(typeName string) (datatype RObject, err error) {

    rtype, found := RT.Types[typeName]
    if ! found {
    	err = fmt.Errorf("No data type has name '%s'", typeName)
    	return
    }

    datatype,found = RT.ReflectedDataTypes[typeName]
    if ! found {
       datatype, err = RT.NewObject("shared.relish.pl2012/relish_lib/pkg/reflect/DataType")
       if err != nil {
    	  return
       }
       dt :=  datatype.(*GoWrapper)
       dt.GoObj = rtype
       RT.ReflectedDataTypes[typeName] = datatype    

       // Now set the attributes of the reflect.DataType object

       nameAttr, found := datatype.Type().GetAttribute("name")
       if ! found {
    	  err = fmt.Errorf("Hmm. Why does reflect.DataType not have a name attribute?")
    	  return
       }
       RT.RestoreAttr(datatype,  nameAttr, String(typeName)) 
    }
	return 
}



/*
  Given a reflect.DataType and the name of a direct or inherited attribute, ensure that there is a reflect.Attribute object
  for the attribute and return that relish object. There is a singleton reflect.Attribute for each RType whose
  reflected DataType has been ensured.
  A hashtable in the RunTime is first searched, then if the reflect.Attribute is not found,
  a new one is created and put in the RunTime hashtable.
  The reflect.Attribute is a *GoWrapper with its GoObj pointing to the underlying *AttributeSpec.
*/
func ensureAttribute(t *RType, attrName string) (attribute RObject, err error) {


    attr, found := t.GetAttribute(attrName)
    if ! found {
    	err = fmt.Errorf("DataType '%s' has no direct or inherited attribute '%s'.", t.Name, attrName)
    	return
    }

    attrKey := attr.WholeType.Name + "_|_" + attr.Part.Name

    attribute,found = RT.ReflectedAttributes[attrKey]
    if ! found {
       attribute, err = RT.NewObject("shared.relish.pl2012/relish_lib/pkg/reflect/Attribute")
       if err != nil {
    	  return
       }
       att :=  attribute.(*GoWrapper)
       att.GoObj = attr
       RT.ReflectedAttributes[attrKey] = attribute

       // Now set the attributes of the reflect.Attribute object from the underlying AttributeSpec and RelEnd.
       // NOTE: Or should we leave these as native accessor methods? 

       nameAttr, found := attribute.Type().GetAttribute("name")
       if ! found {
    	  err = fmt.Errorf("Hmm. Why does reflect.Attribute not have a name attribute?")
    	  return
       }
       RT.RestoreAttr(attribute,  nameAttr, String(attrName)) 


       partType := attr.Part.Type
       var partDataType RObject
       partDataType, err = ensureDataType(partType.Name) 
       if err != nil {
       	  return
       }
       typeAttr, found := attribute.Type().GetAttribute("type")
       if ! found {
    	  err = fmt.Errorf("Hmm. Why does reflect.Attribute not have a type attribute?")
    	  return
       }
       RT.RestoreAttr(attribute,  typeAttr, partDataType) 

       minArityAttr, found := attribute.Type().GetAttribute("minArity")
       if ! found {
    	  err = fmt.Errorf("Hmm. Why does reflect.Attribute not have a minArity attribute?")
    	  return
       }
       RT.RestoreAttr(attribute,  minArityAttr, Int(int64(attr.Part.ArityLow))) 

       maxArityAttr, found := attribute.Type().GetAttribute("maxArity")
       if ! found {
    	  err = fmt.Errorf("Hmm. Why does reflect.Attribute not have a maxArity attribute?")
    	  return
       }
       RT.RestoreAttr(attribute,  maxArityAttr, Int(int64(attr.Part.ArityHigh)))        

       if attr.IsRelation() {
          inverseAttr, found := attribute.Type().GetAttribute("inverse")
          if ! found {
    	     err = fmt.Errorf("Hmm. Why does reflect.Attribute not have an inverse attribute?")
    	     return
          }
 
          var inverseAttribute RObject

          inverseAttrKey := inverseAttr.WholeType.Name + "_|_" + inverseAttr.Part.Name

          inverseAttribute,found = RT.ReflectedAttributes[inverseAttrKey]
          if ! found {
             inverseAttribute, err = ensureAttribute(inverseAttr.WholeType, inverseAttr.Part.Name)
             if err != nil {
       	        return
             }       	

             // Now set the inverse attributes for both Attribute objects.

             RT.RestoreAttr(attribute,  inverseAttr, inverseAttribute)             

             RT.RestoreAttr(inverseAttribute,  inverseAttr, attribute) 
          }  // If inverseAttribute already exists, it will be the one setting up the inverses on both sides.
       }
    }
	return 
}





    // attribute d DataType attributeName String > Attribute