package uuid

import (
	"github.com/google/uuid"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

func NewString() (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", merror.Transform(err).Describe("could not generate random UUIDv4")
	}

	return u.String(), nil
}
