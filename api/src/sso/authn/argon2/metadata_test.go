package argon2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/sqlboiler/v4/types"
)

func TestToMetadata(t *testing.T) {
	jsonStr := `
  {
    "hash_base_64": "Ym9uam91ciBmbG9yZW50IGNvbW1lbnQgdmFzLXR1IGVuIGNldHRlIGJlbGxlIGpvdXJuw6llID8h",
    "params": {
      "salt_base_64": "Yydlc3QgdmFjaGVtZW50IHNhbMOpZSBjb21tZSBwaHJhc2UgZW5jb2TDqWUgZW4gYmFzZSA2NA==",
      "memory": 1024,
      "iterations": 1,
      "parallelism": 1
    }
  }
  `
	msg := types.JSON{}
	err := msg.Scan(jsonStr)
	assert.Nil(t, err)

	ret, err := ToMetadata(msg)
	assert.Nil(t, err)
	assert.Equal(t, "Ym9uam91ciBmbG9yZW50IGNvbW1lbnQgdmFzLXR1IGVuIGNldHRlIGJlbGxlIGpvdXJuw6llID8h", ret.HashBase64)
	assert.Equal(t, "Yydlc3QgdmFjaGVtZW50IHNhbMOpZSBjb21tZSBwaHJhc2UgZW5jb2TDqWUgZW4gYmFzZSA2NA==", ret.Params.SaltBase64)
	assert.Equal(t, 1024, ret.Params.Memory)
	assert.Equal(t, 1, ret.Params.Iterations)
	assert.Equal(t, 1, ret.Params.Parallelism)
}
