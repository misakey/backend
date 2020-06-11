package argon2

import (
	"crypto/subtle"
	"encoding/base64"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
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

// Matches checks whether the input password matches the current password hash
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

// DecodeParams attempts to decode a string from DB
// as a Argon2+HMAC hash, and if successful returns the argon2 parameters.
// This is used to send the parameters to the client during authentication.
func DecodeParams(encodedHash string) (params Params, err error) {
	params, _, _, err = decode(encodedHash)
	if err != nil {
		return params, err
	}
	return params, nil
}

// Validate the password format
func (p HashedPassword) Validate() error {
	if err := v.Validate(&p.HashBase64, v.Required); err != nil {
		return merror.Transform(err).Describe("validating prehashed password")
	}

	if err := v.ValidateStruct(&p.Params,
		v.Field(&p.Params.Memory, v.Required),
		v.Field(&p.Params.Iterations, v.Required),
		v.Field(&p.Params.Parallelism, v.Required),
		v.Field(&p.Params.SaltBase64, v.Required, is.Base64.Error("salt_base64 must be base64 encoded")),
	); err != nil {
		return merror.Transform(err).Describe("validating prehashed password params")
	}

	return nil
}
