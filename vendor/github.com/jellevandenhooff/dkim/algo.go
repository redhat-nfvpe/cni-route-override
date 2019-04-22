package dkim

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"errors"
	"hash"
)

// In violation of RFC 6373, we require a minimum RSA key complexity.
const minRSAKeyComplexity = 1024 // minimum length in bits

// An algo is a verification algorithm as defined in the RFC 6376.
type algo struct {
	hasher   func() hash.Hash
	checkSig func(pubkey []byte, data []byte, signature []byte) error
}

// checkRsa verifies that a signature over data was signed by pubkey using RSA
// and with the given hash function
func checkRsa(pubkey []byte, data []byte, signature []byte, hash crypto.Hash) error {
	key, err := x509.ParsePKIXPublicKey(pubkey)
	if err != nil {
		return err
	}
	rsaPub, ok := key.(*rsa.PublicKey)
	if !ok {
		return errors.New("not an RSA public key")
	}
	if rsaPub.N.BitLen() < minRSAKeyComplexity {
		return errors.New("RSA key too short")
	}
	return rsa.VerifyPKCS1v15(rsaPub, hash, data, signature)
}

// helper function to verify an RSA key with SHA1 as hash function
func checkRsaSha1(pubkey []byte, data []byte, signature []byte) error {
	return checkRsa(pubkey, data, signature, crypto.SHA1)
}

// helper function to verify an RSA key with SHA256 as hash function
func checkRsaSha256(pubkey []byte, data []byte, signature []byte) error {
	return checkRsa(pubkey, data, signature, crypto.SHA256)
}

var algos = map[string]*algo{
	"rsa-sha1":   {hasher: sha1.New, checkSig: checkRsaSha1},
	"rsa-sha256": {hasher: sha256.New, checkSig: checkRsaSha256},
}
