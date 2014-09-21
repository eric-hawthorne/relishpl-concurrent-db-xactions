// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU LESSER GPL v3 license, found in the LICENSE_LGPL3 file.

package crypto_util

/*
   crypto_util.go - convenience methods for doing common cryptographic functions, including public key cryptography.
*/

import (
	"bytes"
	"strings"
	"crypto"
	"crypto/rand"
    "crypto/rsa"
    "crypto/sha256"
    "crypto/x509"
    "errors"
    "encoding/pem"
    "os"
    "util/gos"
    "fmt"
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
	
	
    privateKeyPEM, err = EncodePrivateKeyPEM(privateKeyBytes, password) 	
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

func EncodePrivateKeyPEM(binaryKey []byte, password string) (pemBlock string, err error) {
	pemBlock, err = EncodePEM(binaryKey, "RSA PRIVATE KEY", password)
	return
}

func EncodePublicKeyPEM(binaryKey []byte) (pemBlock string, err error) {
	pemBlock, err = EncodePEM(binaryKey, "RSA PUBLIC KEY","")
	return
}

func EncodeSignaturePEM(binarySignature []byte) (pemBlock string, err error) {
	pemBlock, err = EncodePEM(binarySignature, "SIGNATURE", "")
	return
}

func EncodePEM(binary []byte, blockType string, password string) (pemBlock string, err error) {
	
	var blk *pem.Block
/* Awaiting Go 1.1 */
	if password != "" {
	   passwordBytes := ([]byte)(password)
	   blk, err = x509.EncryptPEMBlock(rand.Reader, blockType, binary, passwordBytes, x509.PEMCipherAES256) 
	   if err != nil {
		  return
	   }
    } else {
 /* */
       blk = new(pem.Block)
       blk.Type = blockType
       blk.Bytes = binary
/* Awaiting Go 1.1 */
    }
/* */

    buf := new(bytes.Buffer)

    err = pem.Encode(buf, blk)
    if err != nil {
	   return
	}

    pemBlock = buf.String()
    return
}

/*
If the password is not "", the private key is decrypted using the password.
*/
func DecodePrivateKeyPEM(privateKeyPEM string, privateKeyPassword string) (priv *rsa.PrivateKey, err error) {
	blockBytes, blockType, err := DecodePEM(privateKeyPEM, privateKeyPassword)
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
	blockBytes, blockType, err := DecodePEM(publicKeyPEM,"")
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
	signatureBytes, blockType, err := DecodePEM(signaturePEM,"")
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
func DecodePEM(pemBlock string, password string) (decoded []byte, blockType string, err error) {


 	var blk *pem.Block
    pb := ([]byte)(pemBlock)

 	blk, _ = pem.Decode(pb)
 	if blk == nil {
 		err = errors.New("DecodePEM: No PEM block found.")
        return
 	}

/* Awaiting Go 1.1 */
	if password != "" {
		passwordBytes := ([]byte)(password)
		decoded, err = x509.DecryptPEMBlock(blk, passwordBytes)
    } else {
/* */    	
        decoded = blk.Bytes
 /* Awaiting Go 1.1 */        
    }
/* */

    blockType = blk.Type
    return
}






/*
Returns a private key and a public key certificate. 
The public key certificate is a text string that includes:
1) 
entityType entityname certifies with this signature
that the public key for entityTypeName entityName is
----BEGIN RSA PUBLIC KEY----
----END RSA PUBLIC KEY----
2) a signature PEM for the statement above it
----BEGIN SIGNATURE----
----END RSIGNATURE----

If the certifying private key is the private key of the origin shared.relish.pl, then
the signature can be verified by using the public key of the origin shared.relish.pl, which
is included in each relish distribution (and should be published at shared.relish.pl).

As a special case, a self-signed cert can be created by supplying the empty string as the
certifyingPrivateKeyPEM. This will result in a public key certificate signed by the very private
key that corresponds to the certified public key.
*/
func GenerateCertifiedKeyPair(keyLenBits int, 
                              certifyingEntityType string,
                              certifyingEntityName string, 
                              certifyingPrivateKeyPEM string,
                              passwordForCertifyingPrivateKey string,	
	                          entityType string,
	                          entityNameAssociatedWithKeyPair string,
	                          passwordForPrivateKey string) (privateKeyPEM string, publicKeyCertificate string, err error) {


    if strings.ToLower(certifyingEntityName) == "shared.relish.pl2012" {	
       err = errors.New("No.")
       return
    }
    
    selfSigned := false
    
    if certifyingEntityName == "" {
    	certifyingEntityType = "origin"
    	certifyingEntityName = "shared.relish.pl2012"

     
    	certifyingPrivateKeyPEM, err = GetPrivateKey(certifyingEntityType, certifyingEntityName) 
 	    if err != nil {
	       if entityNameAssociatedWithKeyPair == "shared.relish.pl2012" {
		      selfSigned = true
		   } else {
		   	  // Only allowed to not have a pre-existing shaerd relish private key here 
		   	  // if creating shared relish key pair for first time.
              err = errors.New("No. No. " + err.Error()) 	
		      return
	       }
	    } else if entityType == "origin" && entityNameAssociatedWithKeyPair == "shared.relish.pl2012" {
	    	  // Trying to re-create shared relish origin key pair certified by another private key or 
	    	  // by an old shared relish private key.
              err = errors.New("No. No. No. ") 	
              return
	    } else {
	       passwordForCertifyingPrivateKey = GetDefaultToken() 
        }
    } else if certifyingEntityName == entityNameAssociatedWithKeyPair && certifyingEntityType == entityType && certifyingPrivateKeyPEM == "" {
    	selfSigned = true
    }

    privateKeyPEM, publicKeyPEM, err := GenerateKeyPair(keyLenBits, passwordForPrivateKey) 
	if err != nil {
		return
	}
	
	if selfSigned {
	   certifyingPrivateKeyPEM = privateKeyPEM
	   passwordForCertifyingPrivateKey = passwordForPrivateKey    		   
	}

    assertion := fmt.Sprintf("%s %s certifies with the signature above\nthat the public key for %s %s is\n%s",
    	                     certifyingEntityType,
                             certifyingEntityName,
                             entityType,
                             entityNameAssociatedWithKeyPair,
                             publicKeyPEM,
    	                    )
	signaturePEM, err := Sign(certifyingPrivateKeyPEM, passwordForCertifyingPrivateKey, strings.TrimSpace(assertion))
	if err != nil {
		return
	}	

	publicKeyCertificate = fmt.Sprintf("%s%s\n",
	                         signaturePEM,
    	                     assertion,
    	                    )

/* Debugging creation of self-signed certs
    if selfSigned {  // A little extra check here.
    	vpk := VerifiedPublicKey(publicKeyPEM, 
	                   publicKeyCertificate, 
	                   entityType, 
	                   entityNameAssociatedWithKeyPair) 
    	if vpk == "" {
              err = errors.New("Self-signed cert did not verify.")  
              return   		
    	} 
    }
*/

	return
}


	
	
func HashSha256(content string) []byte {
	hasher := sha256.New()
	hasher.Write([]byte(content))
	return hasher.Sum(nil)
}	

/*
Signs based on SHA256 hash of the content.
*/
func Sign(privateKeyPEM string, privateKeyPassword string, content string) (signaturePEM string, err error) {
	
	priv, err := DecodePrivateKeyPEM(privateKeyPEM, privateKeyPassword)
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
		fmt.Println("Could not decode publicKey PEM")		
		fmt.Println(err)
		return false
	}

	signatureBytes, err := DecodeSignaturePEM(signaturePEM)
	if err != nil {
		fmt.Println("Could not decode signature PEM")
		fmt.Println(err)		
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
If the certifierPublicKeyPEM is "", it means verify a self-signed public key certificate.
*/
func VerifiedPublicKey(certifierPublicKeyPEM string, 
	                   publicKeyCertificate string, 
	                   entityType string, 
	                   entityName string) (publicKeyPEM string) {
   signatureEndPos := strings.Index(publicKeyCertificate, "-----END SIGNATURE-----\n")
   if signatureEndPos == -1 {
	  return
   }
   assertionStartPos := signatureEndPos + 24
   assertion := strings.TrimSpace(publicKeyCertificate[assertionStartPos:])
   // assertion := publicKeyCertificate[assertionStartPos:]
   signaturePEM := publicKeyCertificate[:assertionStartPos]

   ownerStatement := "\nthat the public key for " + entityType + " " + entityName + " is\n-----BEGIN RSA PUBLIC KEY-----"
   ownerStatementPos := strings.Index(assertion, ownerStatement)
   if ownerStatementPos == -1 {
   	   return
   }

   pubKeyPos := ownerStatementPos + len(ownerStatement) - 30

   publicKeyPEM = assertion[pubKeyPos:]

   if certifierPublicKeyPEM == "" {
   	  certifierPublicKeyPEM = publicKeyPEM
   }
   if ! Verify(certifierPublicKeyPEM, signaturePEM, assertion) {
	   fmt.Println("++++++++++++++++")		
	   fmt.Println(publicKeyPEM)
	   fmt.Println("++++++++++++++++")		
	   fmt.Println(certifierPublicKeyPEM)	
	   fmt.Println("++++++++++++++++")	
	   fmt.Println(signaturePEM)
	   fmt.Println("++++++++++++++++")		
	   fmt.Println(assertion)		
	   fmt.Println("++++++++++++++++")		


    		fmt.Println("len(Certificate)=", len(publicKeyCertificate))
    		fmt.Println("len(trimmed assertion)=", len(assertion))

   	   publicKeyPEM = ""
   }

   return 
}



var defaultToken string

/*
 Sets a default token for OAUTH.
*/
func SetDefaultToken(token string) {
   if defaultToken == "" {
	   defaultToken = token
   }
} 


var relishRuntimeLocation string

func SetRelishRuntimeLocation(path string) {
	relishRuntimeLocation = path
}

func GetRelishRuntimeLocation() string {
   return relishRuntimeLocation
}

/*
*/
func GetDefaultToken() string {
   return defaultToken
}


/*
Get a private key in PEM format from the standard directory in the relish installation, 
using standard file naming convention. 
*/
func GetPrivateKey(entityType string, entityName string) (privateKeyPEM string, err error) {
	fileName := entityType + "__" + entityName + "__private_key.pem"
	path := relishRuntimeLocation + "/keys/private/" + fileName
	
	bts, err := gos.ReadFile(path) 
	if err != nil {
		return
	}
	privateKeyPEM = string(bts)
	return
}


/*
Get a public key certificate in PEM format from the standard directory in the relish installation, 
using standard file naming convention. 
*/
func GetPublicKeyCert(entityType string, entityName string) (publicKeyCertPEM string, err error) {
	fileName := entityType + "__" + entityName + "__public_key.pem"
	path := relishRuntimeLocation + "/keys/public/" + fileName
	
	bts, err := gos.ReadFile(path) 
	if err != nil {
		return
	}
	publicKeyCertPEM = string(bts)
	return
}

/*
Get a private key in PEM format from the standard directory in the relish installation, 
using standard file naming convention. 
*/
func GetTLSwebServerCertAndKeyFilePaths() (certPath string, keyPath string, err error) {
	certPath = relishRuntimeLocation + "/keys/public/tls_web_server_cert.pem"
	keyPath = relishRuntimeLocation + "/keys/private/tls_web_server_key.pem"

   _,err = gos.Stat(certPath)	
   if err != nil {
   	  return
   }

   _,err = gos.Stat(keyPath)	
   if err != nil {
   	  return
   }

	certPath = gos.ToOsSpecificPath(certPath)
	keyPath = gos.ToOsSpecificPath(keyPath)
	return
}



/*
Store a private key in PEM format into a file in the standard directory in the relish installation, 
using standard file naming convention. 
*/
func StorePrivateKey(entityType string, entityName string, privateKeyPEM string) (err error) {
	fileName := entityType + "__" + entityName + "__private_key.pem"
	path := relishRuntimeLocation + "/keys/private/" + fileName
	var perm os.FileMode = 0666
    err = gos.WriteFile(path, ([]byte)(privateKeyPEM), perm) 	
    return    
}


/*
Store a public key certificate in PEM format into a file in the standard directory in the relish installation, 
using standard file naming convention. 
*/
func StorePublicKeyCert(entityType string, entityName string, publicKeyCertPEM string) (err error) {
	fileName := entityType + "__" + entityName + "__public_key.pem"

	keyDirPath := relishRuntimeLocation + "/keys/public"
	path := keyDirPath + "/" + fileName

    var perm os.FileMode = 0777

    err = gos.MkdirAll(keyDirPath, perm)
    if err != nil {
       return
    }
    
	perm = 0666
    err = gos.WriteFile(path, ([]byte)(publicKeyCertPEM), perm) 	
    return
}


