// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU LESSER GPL v3 license, found in the LICENSE_LGPL3 file.

package crypto_util

/*
   crypto_util.go - convenience methods for doing common cryptographic functions, including public key cryptography.
*/

import (
	"bytes"
	"crypto"
	"crypto/rand"
    "crypto/rsa"
    "crypto/sha256"
    "crypto/x509"
    "errors"
    "encoding/pem"
)

const RSAKeySize = 3072  // Default or suggested key length in bits.

/*
Generates an RSA key pair of the specified key length in bits. 
Uses Go's crypto/rand rand.Reader as the source of entropy.

If password is a non-empty string, encrypts the private key so that password is required
to decrypt and use it. If password == "", the private key returned is unencrypted.
*/
func GenerateKeyPair(keyLenBits int, password string) (privateKeyPEM string, publicKeyPEM string, err error) {
	
    priv, err := rsa.GenerateKey(rand.Reader, keyLenBits) 
    if err != nil {
	   return
	}
	err = priv.Validate()
	if err != nil {
	   return
	}	
	
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(priv)
	
	if len(password) > 0 {
		// TODO Need to encrypt the key here
	} 
	
    privateKeyPEM, err = EncodePrivateKeyPEM(privateKeyBytes) 	
    if err != nil {
       return
    }
	
	pub := &(priv.PublicKey)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(pub)
    if err != nil {
       return
	}
	
    publicKeyPEM, err = EncodePublicKeyPEM(publicKeyBytes) 	
    if err != nil {
       return
    }	

    return
}


// Helper functions

func EncodePrivateKeyPEM(binaryKey []byte) (pemBlock string, err error) {
	pemBlock, err = EncodePEM(binaryKey, "RSA PRIVATE KEY")
	return
}

func EncodePublicKeyPEM(binaryKey []byte) (pemBlock string, err error) {
	pemBlock, err = EncodePEM(binaryKey, "RSA PUBLIC KEY")
	return
}

func EncodeSignaturePEM(binarySignature []byte) (pemBlock string, err error) {
	pemBlock, err = EncodePEM(binarySignature, "SIGNATURE")
	return
}

func EncodePEM(binary []byte, blockType string) (pemBlock string, err error) {
    blk := new(pem.Block)
    blk.Type = blockType
    blk.Bytes = binary
    buf := new(bytes.Buffer)

    err = pem.Encode(buf, blk)
    if err != nil {
	   return
	}

    pemBlock = buf.String()
    return
}

func DecodePrivateKeyPEM(privateKeyPEM string) (priv *rsa.PrivateKey, err error) {
	blockBytes, blockType, err := DecodePEM(privateKeyPEM)
	if err != nil {
		return
	}
	if blockType != "RSA PRIVATE KEY" {
 		err = errors.New("DecodePrivateKeyPEM: Expecting RSA PRIVATE KEY, found " + blockType + ".")
        return	    
	}
 	priv, err = x509.ParsePKCS1PrivateKey(blockBytes)	
    return
}

func DecodePublicKeyPEM(publicKeyPEM string) (pub *rsa.PublicKey, err error) {
	blockBytes, blockType, err := DecodePEM(publicKeyPEM)
	if err != nil {
		return
	}
	if blockType != "RSA PUBLIC KEY" {
 		err = errors.New("DecodePublicKeyPEM: Expecting RSA PUBLIC KEY, found " + blockType + ".")
        return	    
	}

	var in interface{}
	in, err = x509.ParsePKIXPublicKey(blockBytes)
	if err != nil {
		return
	}
	pub = in.(*rsa.PublicKey)
    return	
}


func DecodeSignaturePEM(signaturePEM string) (signatureBytes []byte, err error) {
	signatureBytes, blockType, err := DecodePEM(signaturePEM)
	if err != nil {
		return
	}
	if blockType != "SIGNATURE" {
 		err = errors.New("DecodeSignaturePEM: Expecting SIGNATURE, found " + blockType + ".")
        return	    
	}
    return	
}


/*
Given a PEM block in a string e.g.
----- BEGIN RSA PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQClbkoOcBAXWJpRh9x+qEHRVvLs
DjatUqRN/rHmH3rZkdjFEFb/7bFitMDyg6EqiKOU3/Umq3KRy7MHzqv84LHf1c2V
CAltWyuLbfXWce9jd8CSHLI8Jwpw4lmOb/idGfEFrMLT8Ms18pKA4Thrb2TE7yLh
4fINDOjP+yJJvZohNwIDAQAB
----- END RSA PUBLIC KEY-----
Returns the decoding of the base-64 into a byte slice.
Also returns the block type, in this case "RSA PUBLIC KEY"
*/
func DecodePEM(pemBlock string) (decoded []byte, blockType string, err error) {

 	var blk *pem.Block
    pb := ([]byte)(pemBlock)
 	blk, _ = pem.Decode(pb)
 	if blk == nil {
 		err = errors.New("DecodePEM: No PEM block found.")
        return
 	}
    decoded = blk.Bytes
    blockType = blk.Type
    return
}






/*
Returns a private key and a public key certificate. 
The public key certificate is a text string that includes
0) the entity type and entity name that certifies
1) a signature PEM for the statement below it
EMPTY LINE
2) The statement:
The public key for
entityTypeName entityName
is
---- BEGIN RSA PUBLIC KEY ----
---- END RSA PUBLIC KEY ----

If the certifying private key is the private key of the origin shared.relish.ple, then
the signature can be verified by using the public key of the origin shared.relish.pl, which
is included in each relish distribution (and should be published at shared.relish.pl).

As a special case, a self-signed cert can be created by supplying the empty string as the
certifyingPrivateKeyPEM. This will result in a public key certificate signed by the very private
key that corresponds to the certified public key.
*/
func GenerateCertifiedRsaKeyPair(keyLen int, 
	                             entityType string,
	                             entityNameAssociatedWitKeyPair string,
	                             certifyingEntityType string,
	                             certifyingEntityName string, 
	                             certifyingPrivateKeyPEM string) (privateKeyPEM string, publicKeyCertificate string) {
	return // TODO Implement
} 
	
	
func HashSha256(content string) []byte {
	hasher := sha256.New()
	hasher.Write([]byte(content))
	return hasher.Sum(nil)
}	

/*
Signs based on SHA256 hash of the content.
*/
func Sign(privateKeyPEM string, content string) (signaturePEM string, err error) {
	
	priv, err := DecodePrivateKeyPEM(privateKeyPEM)
	if err != nil {
		return
	}
		
	hashed := HashSha256(content)

    signatureBytes, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hashed) 
    if err != nil {
	   return
    }

    signaturePEM, err = EncodeSignaturePEM(signatureBytes)
    return
}

/*
Verifies the signature as being the signature for the content.
Returns true if the signature is the signature of the content, as signed by the private key
that corresponds to the argument public key.
Assumes that SHA256 was used as the hash function to hash the content for signing.
*/
func Verify(publicKeyPEM string, signaturePEM string, content string) bool {
	pubKey, err := DecodePublicKeyPEM(publicKeyPEM)
	if err != nil {
		return false
	}

	signatureBytes, err := DecodeSignaturePEM(signaturePEM)
	if err != nil {
		return false
	}
	
	hashed := HashSha256(content)
	
    err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed, signatureBytes)	
    if err != nil {
	   return false
    }
    return true	
}


/*
If the public key certificate was signed by the certifier private key corresponding to certifierPublicKeyPEM, and
indeed certifies that entityType entityName has the public key in the certificate, then returns the
public key PEM for for the entity. Otherwise returns an empty string.
*/
func VerifiedPublicKey(certifierPublicKeyPEM string, 
	                   publicKeyCertificate string, 
	                   entityType string, 
	                   entityName string) (publicKeyPEM string) {
	return // TODO Implement
}