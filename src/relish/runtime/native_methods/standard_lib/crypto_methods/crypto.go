// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU LESSER GPL v3 license, found in the LICENSE_LGPL3 file.

package crypto_methods

/*
   crypto.go - native methods for a few stereotypical crypto operations.
*/

import (
	. "relish/runtime/data"
//	"os"
//	"io"
//	"bufio"
	"util/crypto_util"
)

///////////
// Go Types

// None so far needed here.

/////////////////////////////////////
// relish method to go method binding

func InitCryptoMethods() {

    // generateKeyPair keyLenBits Int passwordForPrivateKey String > privateKeyPEM String publicKeyPEM String err String
    // 
	generateKeyPairMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/crypto",nil,"generateKeyPair", []string{"keyLenBits","passwordForPrivateKey"}, []string{"Int","String"}, []string{"String","String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	generateKeyPairMethod.PrimitiveCode = generateKeyPair





}




 
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Stereotypical Crypto functions including public-key crypto key-pair generation, 
// digital-object signing, and signature verification.

// generateKeyPair keyLenBits Int passwordForPrivateKey String > privateKeyPem String publicKeyPem String err String
// """
//  Generates an RSA private key and public key as ascii-armoured (base-64) PEM strings.
//  Minimum keyLenBits recommended is 2048, 3072 is better.
//  If passwordForPrivateKey is not "", then the private key returned is encrypted and, before its use, 
//  it must be decrypted using the password.
// """
//
//
func generateKeyPair (th InterpreterThread, objects []RObject) []RObject {
	keyLenBits := int(int64(objects[0].(Int)))
    password := string(objects[1].(String))
    privateKeyPEM, publicKeyPEM, err := crypto_util.GenerateKeyPair(keyLenBits, password) 	
    var errStr string
    if err != nil {
       errStr = err.Error()
    }
	return []RObject{String(privateKeyPEM), String(publicKeyPEM), String(errStr)}    
}

/*
sign privateKeyPem String privateKeyPassword String content String > signaturePem String err String

Signs based on SHA256 hash of the content.
*/
func Sign (th InterpreterThread, objects []RObject) []RObject {
    privateKeyPEM := string(objects[0].(String))
    password := string(objects[1].(String))
    content := string(objects[2].(String))
    signaturePEM, err := crypto_util.Sign(privateKeyPEM, password, content) 
	
    var errStr string
    if err != nil {
       errStr = err.Error()
    }
	return []RObject{String(signaturePEM), String(errStr)}    
}


/*
Verifies the signature as being the signature for the content.
Returns true if the signature is the signature of the content, as signed by the private key
that corresponds to the argument public key.
Assumes that SHA256 was used as the hash function to hash the content for signing.
*/
func Verify(publicKeyPEM string, signaturePEM string, content string) bool {


///////////////////////////////////////////////////////////////////////////////////////////
// Type init functions


// None yet -  if any relish types for crypto, remember to add the artifact to defs.go




