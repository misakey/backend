package ajwt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPurposes(t *testing.T) {
	t.Run("test IsPurposeScope", func(t *testing.T) {
		assert.Equal(t, true, IsPurposeScope("pur.minimum_required"))
		assert.Equal(t, true, IsPurposeScope("pur:minimum_required"))
		assert.Equal(t, false, IsPurposeScope("rol.admin.2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, false, IsPurposeScope("rol.dpo.2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, false, IsPurposeScope("rol.pirate.2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, false, IsPurposeScope("openid"))
		assert.Equal(t, false, IsPurposeScope("user"))
		assert.Equal(t, false, IsPurposeScope("application"))
		assert.Equal(t, false, IsPurposeScope("service"))
		assert.Equal(t, false, IsPurposeScope("misadmin"))
	})

	t.Run("test TrimPurposes", func(t *testing.T) {
		// with all kind of scopes
		scopes := []string{"user", "openid", "pur:1", "pur:2", "rol.admin.3"}
		trimmedScopes := TrimPurposes(scopes)
		assert.Equal(t, trimmedScopes, []string{"user", "openid", "rol.admin.3"})
		assert.Equal(t, []string{"user", "openid", "pur:1", "pur:2", "rol.admin.3"}, scopes)

		// only with purpose scopes
		scopes = []string{"pur:1", "pur:2"}
		trimmedScopes = TrimPurposes(scopes)
		assert.Equal(t, trimmedScopes, []string{})
		assert.Equal(t, []string{"pur:1", "pur:2"}, scopes)

		// without any purpose scopes
		scopes = []string{"service", "openid"}
		trimmedScopes = TrimPurposes(scopes)
		assert.Equal(t, trimmedScopes, []string{"service", "openid"})
		assert.Equal(t, []string{"service", "openid"}, scopes)
	})
}
