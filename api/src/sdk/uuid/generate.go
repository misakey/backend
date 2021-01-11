package uuid

import (
	"github.com/google/uuid"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
)

// NewString ...
func NewString() (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", merr.From(err).Desc("could not generate random UUIDv4")
	}

	return u.String(), nil
}
