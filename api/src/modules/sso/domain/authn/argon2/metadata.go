package argon2

import (
	"encoding/json"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

type pwdMetadata struct {
	HashedPassword
}

// ToMetadata password conversion from a RawJSON message
func ToMetadata(msg json.Marshaler) (ret pwdMetadata, err error) {
	msgJSON, err := msg.MarshalJSON()
	if err != nil {
		return ret, merror.Transform(err).Describe("password metadata")
	}
	err = json.Unmarshal(msgJSON, &ret)
	return ret, err
}
