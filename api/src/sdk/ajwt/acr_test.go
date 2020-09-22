package ajwt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

func TestACRIsGTE(t *testing.T) {
	tests := map[string]struct {
		claims AccessClaims
		secLvl ACRSecLvl
		err    error
	}{
		"wished SecLevel 1 vs claim ACR 0 should return an error": {
			claims: AccessClaims{
				ACR: noSecLevel,
			},
			secLvl: ACRSecLvl1,
			err:    merror.Forbidden().From(merror.OriACR).Describe("token acr is 0").Detail("acr", merror.DVForbidden).Detail("required_acr", "1"),
		},
		"wished SecLevel 2 vs claim ACR 0 should return an error": {
			claims: AccessClaims{
				ACR: noSecLevel,
			},
			secLvl: ACRSecLvl2,
			err:    merror.Forbidden().From(merror.OriACR).Describe("token acr is 0").Detail("acr", merror.DVForbidden).Detail("required_acr", "2"),
		},
		"wished SecLevel 1 vs claim ACR 2 should return no error": {
			claims: AccessClaims{
				ACR: ACRSecLvl2,
			},
			secLvl: ACRSecLvl1,
			err:    nil,
		},
		"wished SecLevel 2 vs claim ACR 2 should return no error": {
			claims: AccessClaims{
				ACR: ACRSecLvl2,
			},
			secLvl: ACRSecLvl2,
			err:    nil,
		},
		"wished SecLevel 1 vs claim ACR 1 should return no error": {
			claims: AccessClaims{
				ACR: ACRSecLvl1,
			},
			secLvl: ACRSecLvl1,
			err:    nil,
		},
		"wished SecLevel 2 vs claim ACR 1 should return an error": {
			claims: AccessClaims{
				ACR: ACRSecLvl1,
			},
			secLvl: ACRSecLvl2,
			err:    merror.Forbidden().From(merror.OriACR).Describe("token acr is too weak").Detail("acr", merror.DVForbidden).Detail("required_acr", "2"),
		},
	}
	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			result := test.claims.ACRIsGTE(test.secLvl)
			assert.Equal(t, test.err, result)
		})
	}
}
