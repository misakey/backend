package ajwt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasRole(t *testing.T) {
	t.Run("test HasRole with exisiting role scopes", func(t *testing.T) {
		// init claims with existing scopes
		claim := AccessClaims{
			Scope: "user openid rol.admin.2e9394f2-fd9f-4a07-beb5-748c35062cad",
		}
		assert.Equal(t, true, claim.hasRole("2e9394f2-fd9f-4a07-beb5-748c35062cad", adminRoleLabel))
		assert.Equal(t, false, claim.IsNotAdminOn("2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, true, claim.IsAdminOn("2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, true, claim.IsNotDPOOn("2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, false, claim.IsDPOOn("2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, true, claim.IsNotDPO())
		assert.Equal(t, false, claim.IsNotAdmin())
		assert.Equal(t, false, claim.IsAnyDPO())
		assert.Equal(t, true, claim.IsAnyAdmin())
		assert.Equal(t, false, claim.IsAnyService())
	})

	t.Run("test HasRole with blank role scopes", func(t *testing.T) {
		// init claims with existing scopes
		claim := AccessClaims{
			Scope: "user rol.dpo.2e9394f2-fd9f-4a07-beb5-748c35062cad",
		}
		assert.Equal(t, true, claim.hasRole("2e9394f2-fd9f-4a07-beb5-748c35062cad", dpoRoleLabel))
		assert.Equal(t, true, claim.IsNotAdminOn("2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, false, claim.IsAdminOn("2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, false, claim.IsNotDPOOn("2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, true, claim.IsDPOOn("2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, false, claim.IsNotDPO())
		assert.Equal(t, true, claim.IsNotAdmin())
		assert.Equal(t, true, claim.IsAnyDPO())
		assert.Equal(t, false, claim.IsAnyAdmin())
		assert.Equal(t, false, claim.IsAnyService())
	})

	t.Run("test HasRole with simple user", func(t *testing.T) {
		// init claims with existing scopes
		claim := AccessClaims{
			Scope: "user openid",
		}
		assert.Equal(t, false, claim.hasRole("2e9394f2-fd9f-4a07-beb5-748c35062cad", dpoRoleLabel))
		assert.Equal(t, true, claim.IsNotAdminOn("2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, false, claim.IsAdminOn("2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, true, claim.IsNotDPOOn("2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, false, claim.IsDPOOn("2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, true, claim.IsNotDPO())
		assert.Equal(t, true, claim.IsNotAdmin())
		assert.Equal(t, false, claim.IsAnyDPO())
		assert.Equal(t, false, claim.IsAnyAdmin())
		assert.Equal(t, true, claim.IsAnyUser())
		assert.Equal(t, false, claim.IsNotAnyUser())
		assert.Equal(t, false, claim.IsAnyService())
	})

	t.Run("test HasRole with application", func(t *testing.T) {
		// init claims with existing scopes
		claim := AccessClaims{
			Scope: "application",
		}
		assert.Equal(t, false, claim.hasRole("2e9394f2-fd9f-4a07-beb5-748c35062cad", dpoRoleLabel))
		assert.Equal(t, true, claim.IsNotAdminOn("2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, false, claim.IsAdminOn("2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, true, claim.IsNotDPOOn("2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, false, claim.IsDPOOn("2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, true, claim.IsNotDPO())
		assert.Equal(t, true, claim.IsNotAdmin())
		assert.Equal(t, false, claim.IsAnyDPO())
		assert.Equal(t, false, claim.IsAnyAdmin())
		assert.Equal(t, false, claim.IsAnyUser())
		assert.Equal(t, true, claim.IsNotAnyUser())
		assert.Equal(t, true, claim.IsAnyApp())
		assert.Equal(t, false, claim.IsAnyService())
	})

	t.Run("test HasRole on a non-user", func(t *testing.T) {
		// init claims with existing scopes
		claim := AccessClaims{
			Scope: "openid application rol.admin.2e9394f2-fd9f-4a07-beb5-748c35062cad",
		}
		assert.Equal(t, false, claim.hasRole("2e9394f2-fd9f-4a07-beb5-748c35062cad", dpoRoleLabel))
		assert.Equal(t, true, claim.IsNotAdminOn("2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, true, claim.IsNotDPOOn("2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, false, claim.IsDPOOn("2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, true, claim.IsNotAdmin())
		assert.Equal(t, true, claim.IsNotDPO())
	})

	t.Run("test IsRoleScope", func(t *testing.T) {
		assert.Equal(t, true, IsRoleScope("rol.admin.2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, true, IsRoleScope("rol.dpo.2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, false, IsRoleScope("rol.pirate.2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, false, IsRoleScope("rol.dpo.*"))
		assert.Equal(t, false, IsRoleScope("openid"))
		assert.Equal(t, false, IsRoleScope("user"))
		assert.Equal(t, false, IsRoleScope("application"))
		assert.Equal(t, false, IsRoleScope("service"))
		assert.Equal(t, false, IsRoleScope("misadmin"))
		assert.Equal(t, false, IsRoleScope("pur.minimum_required"))
		assert.Equal(t, false, IsRoleScope("pur:minimum_required"))
	})

	t.Run("test GetRole", func(t *testing.T) {
		// GetRole on DPO
		appRole := GetRole("rol.dpo.2e9394f2-fd9f-4a07-beb5-748c35062cad")
		assert.Equal(t, dpoRoleLabel, appRole.RoleLabel)
		assert.Equal(t, "2e9394f2-fd9f-4a07-beb5-748c35062cad", appRole.ApplicationID)

		// GetRole on Admin
		appRole = GetRole("rol.admin.2e9394f2-fd9f-4a07-beb5-748c35062cad")
		assert.Equal(t, adminRoleLabel, appRole.RoleLabel)
		assert.Equal(t, "2e9394f2-fd9f-4a07-beb5-748c35062cad", appRole.ApplicationID)

		// GetRole on non-existing role
		appRole = GetRole("application")
		assert.Equal(t, "", appRole.RoleLabel)
		assert.Equal(t, "", appRole.ApplicationID)
	})

	t.Run("test GetDPOAppID", func(t *testing.T) {
		desc := "with dpo scope should return app ID"
		claim := AccessClaims{
			Scope: "openid application rol.dpo.2e9394f2-fd9f-4a07-beb5-748c35062cad",
		}
		appID := claim.GetDPOAppID()
		assert.Equalf(t, "2e9394f2-fd9f-4a07-beb5-748c35062cad", *appID, desc)

		desc = "without dpo scope should return nil"
		claim = AccessClaims{
			Scope: "openid application rol.admin.2e9394f2-fd9f-4a07-beb5-748c35062cad",
		}
		appID = claim.GetDPOAppID()
		assert.Nilf(t, appID, desc)

		desc = "without any scope should return nil"
		claim = AccessClaims{
			Scope: "openid application",
		}
		appID = claim.GetAdminAppID()
		assert.Nilf(t, appID, desc)
	})

	t.Run("test GetAdminAppID", func(t *testing.T) {
		// init claims with admin scope should return app ID
		claim := AccessClaims{
			Scope: "openid application rol.admin.2e9394f2-fd9f-4a07-beb5-748c35062cad",
		}
		appID := claim.GetAdminAppID()
		assert.Equal(t, "2e9394f2-fd9f-4a07-beb5-748c35062cad", *appID)

		// init claims without admin scope should return nil
		claim = AccessClaims{
			Scope: "openid application rol.dpo.2e9394f2-fd9f-4a07-beb5-748c35062cad",
		}
		appID = claim.GetAdminAppID()
		assert.Nil(t, appID)

	})
}
