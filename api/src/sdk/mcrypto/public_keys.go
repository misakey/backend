package mcrypto

import (
	"errors"
	"strings"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/volatiletech/null/v8"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
)

const (
	// NACL_ENC_ALGO ("com.misakey.nacl-enc") is the default encryption algorithm
	NACL_ENC_ALGO = "com.misakey.nacl-enc"
	// AES_RSA_ENC_ALGO ("com.misakey.aes-rsa-enc") is an alternate encryption algorithm
	// designed for the needs of EPI Use Labs
	AES_RSA_ENC_ALGO = "com.misakey.aes-rsa-enc"
)

/*
 * XXX Maybe we should consider creating a dedicated type:
 *
 *    type PublicKey string
 *
 * but this requires going through the whole codebase
 * and replacing the type of every public key.
 * It also means we will need casting to and from null.String
 * every time we interact with the SQLBoiler layer.
 */

func ValidateNaClPublicKey(pk interface{}) error {
	s, ok := pk.(string)
	if !ok {
		n, ok := pk.(null.String)
		if !ok {
			return errors.New("public key is not a string")
		}
		s = n.String
	}
	err := v.Validate(s, v.Match(format.UnpaddedURLSafeBase64))
	if err != nil {
		// setting the code ourselves because our "ozzo_needle" does not recognize the error
		// as a validation one (and creates a "HTTP 500 Internal Server Error")
		return merr.From(err).Code(merr.BadRequestCode)
	}

	return nil
}

// Validate implements interface v.Rule
// (so that it can be used as "v.By(ValidatePublicKey)")
func ValidatePublicKey(pk interface{}) error {
	s, ok := pk.(string)
	if !ok {
		n, ok := pk.(null.String)
		if !ok {
			return errors.New("public key is not a string")
		}
		s = n.String
	}

	if strings.HasPrefix(s, AES_RSA_ENC_ALGO) {
		start := len(AES_RSA_ENC_ALGO) + 1
		err := v.Validate(s[start:], v.Match(format.UnpaddedURLSafeBase64))
		if err != nil {
			// setting the code ourselves because our "ozzo_needle" does not recognize the error
			// as a validation one (and creates a "HTTP 500 Internal Server Error")
			return merr.From(err).Code(merr.BadRequestCode)
		}
		return nil
	}

	err := ValidateNaClPublicKey(pk)
	if err != nil {
		// setting the code ourselves because our "ozzo_needle" does not recognize the error
		// as a validation one (and creates a "HTTP 500 Internal Server Error")
		return merr.From(err).Code(merr.BadRequestCode)
	}
	return nil
}

func GetPublicKeyEncryptionAlgorithm(pk string) string {
	if strings.HasPrefix(pk, AES_RSA_ENC_ALGO) {
		return AES_RSA_ENC_ALGO
	}

	return NACL_ENC_ALGO
}
