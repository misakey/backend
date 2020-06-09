package argon2

import (
	"encoding/json"

	"github.com/volatiletech/sqlboiler/types"
)

type pwdMetadata struct {
	HashedPassword
}

// ToMetadata password conversion from a RawJSON message
func ToMetadata(msg types.JSON) (ret pwdMetadata, err error) {
	msgJSON, err := msg.MarshalJSON()
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(msgJSON, &ret)
	return ret, err
}
