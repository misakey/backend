package authentication

import (
	"crypto/rand"
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

// generateCodeMetadata cryptographically: a code with secure pseudorandom number generated.
func generateCodeMetadata() (ret types.JSON, err error) {
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

func toCodeMetadata(msg types.JSON) (ret codeMetadata, err error) {
	msgJSON, err := msg.MarshalJSON()
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(msgJSON, &ret)
	return ret, err
}
