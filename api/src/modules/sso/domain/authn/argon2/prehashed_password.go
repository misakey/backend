package argon2

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
)

// We consider received password following argon2 server relief principle
// see https://password-hashing.net/submissions/specs/Argon-v3.pdf for more information

// Params are the Argon2 parameters the client used to hash the password,
// which we must store so that the client can use the same parameters next time
type Params struct {
	Memory int `json:"memory"`
	// called "time" in argon2-browser JS library
	Iterations int `json:"iterations"`
	// should always be "1" in JS
	// but we still include the param for the sake of rigor
	Parallelism int    `json:"parallelism"`
	SaltBase64  string `json:"salt_base64"`
}

// Argon2Prehashed represents a password hashed with Argon2.
// This is the object the client sends instead of a password
// when we use "server relief".
//
// HashedPassword satisfies interface "Password"
type HashedPassword struct {
	// argon2 parameters the client used to hash the password
	Params     Params `json:"params"`
	HashBase64 string `json:"hash_base64"`
}

// Hash generates a new random salt
// and hashes the password with HMAC-SHA256 using this salt as the HMAC key.
// Note that here the input is a password that was *already hashed*
// by the client (using Argon2) but we still have to hash it again
// so that an attacker getting his hand on a snapshot of our database
// is still unable to log himself in
// (cannot compute the Argon2 hash from the Argon2+HMAC hash).
// Note that unlike Argon2, HMAC consumes very few resources.
func (p HashedPassword) Hash() (encodedHash string, err error) {
	clientHash, err := base64.StdEncoding.DecodeString(p.HashBase64)
	if err != nil {
		return "", err
	}

	serverSalt, err := generateRandomSalt()
	if err != nil {
		return "", err
	}

	finalHash := hash(clientHash, serverSalt)

	return encode(p.Params, serverSalt, finalHash), nil
}

// Matches checks whether the password matches a password hash from the database
func (p HashedPassword) Matches(encodedHash string) (bool, error) {
	_, serverSalt, expectedHash, err := decode(encodedHash)
	if err != nil {
		return false, err
	}

	clientHash, err := base64.StdEncoding.DecodeString(p.HashBase64)
	if err != nil {
		return false, err
	}

	computedHash := hash(clientHash, serverSalt)

	// note the constant time comparison to avoid timing attacks
	matches := subtle.ConstantTimeCompare(computedHash, expectedHash) == 1

	return matches, nil
}

// DecodeArgon2Params attempts to decode a string from DB
// as a Argon2+HMAC hash, and if successful returns the argon2 parameters.
// This is used to send the parameters to the client during authentication.
func DecodeArgon2Params(encodedHash string) (params *Params, err error) {
	params, _, _, err = decode(encodedHash)
	if err != nil {
		return params, err
	}
	return params, nil
}

const (
	algorithmID = "com.misakey.argon2-relief-v1"
)

var (
	ErrBadFormat = errors.New(fmt.Sprintf(`the stored hash is not in format "%s"`, algorithmID))
)
