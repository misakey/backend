// TODO Move most of these utilities
// to make them usable everywhere else in the code
package utils

import (
	"regexp"

	"github.com/google/uuid"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

var RxUnpaddedURLsafeBase64 = regexp.MustCompile("^[a-zA-Z0-9_-]+$")

func RandomUUIDString() (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", merror.Transform(err).Describe("could not generate random UUIDv4")
	}

	return u.String(), nil
}
