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



    // generateCertifiedKeyPair 
    //    keyLenBits Int 
    //    certifyingEntityType String 
    //    certifyingEntityName String 
    //    passwordForCertifyingPrivateKey String
    //    entityType String 
    //    entityName String 
    //    passwordForPrivateKey String 
    // > privateKeyPem String publicKeyPem String err String
    // 
	generateCertifiedKeyPairMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/crypto",
	                                                       nil,
	                                                       "generateCertifiedKeyPair", 
	                                                       []string{"keyLenBits","certifyingEntityType","certifyingEntityName","certifyingPrivateKeyPem","passwordForCertifyingPrivateKey","entityType","entityName","passwordForPrivateKey"}, 
	                                                       []string{"Int","String","String","String","String","String","String","String"}, 
	                                                       []string{"String","String","String"}, 
	                                                       false, 
	                                                       0, 
	                                                       false)
	if err != nil {
		panic(err)
	}
	generateCertifiedKeyPairMethod.PrimitiveCode = generateCertifiedKeyPair


    // sign privateKeyPem String privateKeyPassword String content String > signaturePem String err String
    // 
	signMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/crypto",nil,"sign", []string{"privateKeyPem","passwordForPrivateKey"}, []string{"String","String"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	signMethod.PrimitiveCode = sign




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
Verifies the signature as being the signature for the content.
Returns true if the signature is the signature of the content, as signed by the private key
that corresponds to the argument public key.
Assumes that SHA256 was used as the hash function to hash the content for signing.

func Verify(publicKeyPEM string, signaturePEM string, content string) bool {
}
*/

/*

getPrivateKey entityType entityName > privateKeyPem
"""
 Get from file.
"""

getPublicKeyCert entityType entityName > publicKeyPem
"""
 Get from file.
"""



func VerifiedPublicKey(certifierPublicKeyPEM string, 
	                   publicKeyCertificate string, 
	                   entityType string, 
	                   entityName string) (publicKeyPEM string)


*/












// generateCerifiedKeyPair 
//    keyLenBits Int 
//    certifyingEntityType String 
//    certifyingEntityName String String
//    certifyingPrivateKeyPem 
//    passwordForCertifyingPrivateKey String
//    entityType String 
//    entityName String 
//    passwordForPrivateKey String 
// > 
//    privateKeyPem String publicKeyCertPem String err String
// """
//  Generates an RSA private key and public key as ascii-armoured (base-64) PEM strings.
//  Minimum keyLenBits recommended is 2048, 3072 is better.
//  If passwordForPrivateKey is not "", then the private key returned is encrypted and, before its use, 
//  it must be decrypted using the password.
// """
//
//
func generateCertifiedKeyPair (th InterpreterThread, objects []RObject) []RObject {
	keyLenBits := int(int64(objects[0].(Int)))
    certifyingEntityType := string(objects[1].(String))	
    certifyingEntityName := string(objects[2].(String))	
    certifyingPrivateKeyPEM := string(objects[3].(String))
    certifyingPrivateKeyPassword := string(objects[4].(String))
    entityType := string(objects[5].(String))
    entityName := string(objects[6].(String))	
    password := string(objects[7].(String))
    privateKeyPEM, publicKeyCertPEM, err := crypto_util.GenerateCertifiedKeyPair(keyLenBits, certifyingEntityType, certifyingEntityName, certifyingPrivateKeyPEM, certifyingPrivateKeyPassword, entityType, entityName, password) 	
    var errStr string
    if err != nil {
       errStr = err.Error()
    }
	return []RObject{String(privateKeyPEM), String(publicKeyCertPEM), String(errStr)}    
}


/*
sign privateKeyPem String privateKeyPassword String content String > signaturePem String err String

Signs based on SHA256 hash of the content.
*/
func sign (th InterpreterThread, objects []RObject) []RObject {
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

///////////////////////////////////////////////////////////////////////////////////////////
// Type init functions


// None yet -  if any relish types for crypto, remember to add the artifact to defs.go




