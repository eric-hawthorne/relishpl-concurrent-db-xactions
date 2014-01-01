// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.

package data

/*
   transcoder_json.go -  encoding and decoding of objects/values to JSON Strings. 
*/

import (
	    "encoding/json"
        "bytes"
)

func JsonMarshal(obj RObject, includePrivate bool) (encoded string, err error) {
   visited := make(map[RObject]bool)		
   tree, err := obj.ToMapListTree(includePrivate, visited)
   if err != nil {
      return		
   } 
   b, err := json.Marshal(tree)
   if err != nil {
      return	
   } 
   var buf bytes.Buffer
   json.HTMLEscape(&buf, b)
   encoded = buf.String()
  
   return
}
