package argon2

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestEncodingDecoding(t *testing.T) {
	t.Run("encoding is consistent with decoding", func(t *testing.T) {
		params := Params{
			Memory:      1024,
			Iterations:  1,
			Parallelism: 1,
			SaltBase64:  "ZmFrZVJhbmRvbVNhbHQ=",
		}

		serverSalt := []byte{'w', 'h', 'a', 't', 'e', 'v', 'e', 'r'}
		finalHash := []byte{'a', 'l', 'l', 'f', 'a', 'k', 'e'}

		encoded := encode(params, serverSalt, finalHash)

		decodedParams, decodedServerSalt, decodedFinalHash, err := decode(encoded)
		if err != nil {
			t.Error(err.Error())
		}

		assert.Equal(t, decodedParams, params)
		assert.Equal(t, decodedServerSalt, serverSalt)
		assert.Equal(t, decodedFinalHash, finalHash)
	})
}
