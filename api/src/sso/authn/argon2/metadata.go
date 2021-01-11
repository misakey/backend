package argon2

import (
	"encoding/json"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
)

// PwdMetadata ...
type PwdMetadata struct {
	HashedPassword
}

// ToMetadata password conversion from a RawJSON message
func ToMetadata(msg json.Marshaler) (ret PwdMetadata, err error) {
	msgJSON, err := msg.MarshalJSON()
	if err != nil {
		return ret, merr.From(err).Desc("password metadata")
	}
	err = json.Unmarshal(msgJSON, &ret)
	return ret, err
}
