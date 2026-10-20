package identity

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

// signRS256 is what an issuer does; the tests need it, the gateway does not.
func signRS256(key *rsa.PrivateKey, signingInput string) ([]byte, error) {
	sum := sha256.Sum256([]byte(signingInput))
	return rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, sum[:])
}
