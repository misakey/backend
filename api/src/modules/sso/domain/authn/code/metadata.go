package code

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"

	"github.com/volatiletech/sqlboiler/types"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

var (
	codeSize = 6
	table    = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
)

type codeMetadata struct {
	Code string `json:"code"`
}

// GenerateAsRawJSON a code cryptographically: a code with secure pseudorandom number generated.
func GenerateAsRawJSON() (ret types.JSON, err error) {
	b := make([]byte, codeSize)
	n, err := io.ReadAtLeast(rand.Reader, b, codeSize)
	if err != nil {
		return ret, merror.Transform(err).Describe("generate code")
	}
	if n != codeSize {
		return ret, fmt.Errorf("generate code: read less than the wished size: %d vs %d", n, codeSize)
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	data, err := json.Marshal(codeMetadata{Code: string(b)})
	if err != nil {
		return ret, err
	}
	return ret, ret.UnmarshalJSON(data)
}

// ToMetadata code conversion from a RawJSON message
func ToMetadata(msg json.Marshaler) (ret codeMetadata, err error) {
	msgJSON, err := msg.MarshalJSON()
	if err != nil {
		return ret, merror.Transform(err).Describe("code metadata")
	}
	err = json.Unmarshal(msgJSON, &ret)
	return ret, err
}

// Matches checks whether an input code the current code matches
func (c codeMetadata) Matches(input codeMetadata) bool {
	return subtle.ConstantTimeCompare([]byte(input.Code), []byte(c.Code)) == 1
}
