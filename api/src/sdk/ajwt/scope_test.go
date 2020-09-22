package ajwt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScope(t *testing.T) {
	t.Run("test Is(Not)User", func(t *testing.T) {
		// 1. the expected user
		claims := AccessClaims{
			Scope:   "openid user rol:1:2",
			Subject: "18b88d48-ad48-43b9-a323-eab1de68b280",
		}
		assert.Equal(t, false, claims.IsNotUser("18b88d48-ad48-43b9-a323-eab1de68b280"))
		assert.Equal(t, true, claims.IsUser("18b88d48-ad48-43b9-a323-eab1de68b280"))

		// 2. not the expected user
		claims = AccessClaims{
			Scope:   "openid user rol:1:2",
			Subject: "another-uuid",
		}
		assert.Equal(t, true, claims.IsNotUser("18b88d48-ad48-43b9-a323-eab1de68b280"))
		assert.Equal(t, false, claims.IsUser("18b88d48-ad48-43b9-a323-eab1de68b280"))

		// 3. not even a user
		claims = AccessClaims{
			Scope:   "openid service",
			Subject: "2e9394f2-fd9f-4a07-beb5-748c35062cad",
		}
		assert.Equal(t, true, claims.IsNotUser("18b88d48-ad48-43b9-a323-eab1de68b280"))
		assert.Equal(t, false, claims.IsUser("18b88d48-ad48-43b9-a323-eab1de68b280"))

		// 4. empty subject
		claims = AccessClaims{
			Scope:   "openid user rol:1:2",
			Subject: "",
		}
		assert.Equal(t, true, claims.IsNotUser("18b88d48-ad48-43b9-a323-eab1de68b280"))
		assert.Equal(t, false, claims.IsUser("18b88d48-ad48-43b9-a323-eab1de68b280"))

		// 5. empty input
		claims = AccessClaims{
			Scope:   "openid user rol:1:2",
			Subject: "2e9394f2-fd9f-4a07-beb5-748c35062cad",
		}
		assert.Equal(t, true, claims.IsNotUser(""))
		assert.Equal(t, false, claims.IsUser(""))
	})

	t.Run("test Is(Not)App", func(t *testing.T) {
		// 1. the expected application
		claims := AccessClaims{
			Scope:   "openid application",
			Subject: "00000000-1111-2222-3333-444444444444",
		}
		assert.Equal(t, false, claims.IsNotApp("00000000-1111-2222-3333-444444444444"))
		assert.Equal(t, true, claims.IsApp("00000000-1111-2222-3333-444444444444"))

		// 2. not the expected app
		claims = AccessClaims{
			Scope:   "openid application",
			Subject: "another_uuid",
		}
		assert.Equal(t, true, claims.IsNotApp("00000000-1111-2222-3333-444444444444"))
		assert.Equal(t, false, claims.IsApp("00000000-1111-2222-3333-444444444444"))

		// 3. not even an application
		claims = AccessClaims{
			Scope:   "openid user",
			Subject: "18b88d48-ad48-43b9-a323-eab1de68b280",
		}
		assert.Equal(t, true, claims.IsNotApp("00000000-1111-2222-3333-444444444444"))
		assert.Equal(t, false, claims.IsApp("00000000-1111-2222-3333-444444444444"))

		// 4. empty subject
		claims = AccessClaims{
			Scope:   "openid application",
			Subject: "",
		}
		assert.Equal(t, true, claims.IsNotApp("00000000-1111-2222-3333-444444444444"))
		assert.Equal(t, false, claims.IsApp("00000000-1111-2222-3333-444444444444"))

		// 5. empty input
		claims = AccessClaims{
			Scope:   "openid application",
			Subject: "00000000-1111-2222-3333-444444444444",
		}
		assert.Equal(t, true, claims.IsNotApp(""))
		assert.Equal(t, false, claims.IsApp(""))
	})

	t.Run("test Is(Not)Service", func(t *testing.T) {
		// 1. the expected service
		claims := AccessClaims{
			Scope:   "openid service",
			Subject: "2e9394f2-fd9f-4a07-beb5-748c35062cad",
		}
		assert.Equal(t, false, claims.IsNotService("2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, true, claims.IsService("2e9394f2-fd9f-4a07-beb5-748c35062cad"))

		// 2. not the expected service
		claims = AccessClaims{
			Scope:   "openid service",
			Subject: "databox-backend",
		}
		assert.Equal(t, true, claims.IsNotService("2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, false, claims.IsService("2e9394f2-fd9f-4a07-beb5-748c35062cad"))

		// 3. not even a service
		claims = AccessClaims{
			Scope:   "openid user",
			Subject: "2e9394f2-fd9f-4a07-beb5-748c35062cad",
		}
		assert.Equal(t, true, claims.IsNotService("2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, false, claims.IsService("2e9394f2-fd9f-4a07-beb5-748c35062cad"))

		// 4. empty subject
		claims = AccessClaims{
			Scope:   "openid service",
			Subject: "",
		}
		assert.Equal(t, true, claims.IsNotService("00000000-1111-2222-3333-444444444444"))
		assert.Equal(t, false, claims.IsService("00000000-1111-2222-3333-444444444444"))

		// 5. empty input
		claims = AccessClaims{
			Scope:   "openid service",
			Subject: "00000000-1111-2222-3333-444444444444",
		}
		assert.Equal(t, true, claims.IsNotService(""))
		assert.Equal(t, false, claims.IsService(""))
	})

	t.Run("test Is(Not)Misadmin", func(t *testing.T) {
		// 1. a valid misadmin
		claims := AccessClaims{
			Scope: "openid user misadmin",
		}
		assert.Equal(t, false, claims.IsNotMisadmin())
		assert.Equal(t, true, claims.IsMisadmin())

		// 2. an invalid misadmin - missing user scope
		claims = AccessClaims{
			Scope: "openid misadmin",
		}
		assert.Equal(t, true, claims.IsNotMisadmin())
		assert.Equal(t, false, claims.IsMisadmin())

		// 3. not even a user
		claims = AccessClaims{
			Scope:   "openid service",
			Subject: "2e9394f2-fd9f-4a07-beb5-748c35062cad",
		}
		assert.Equal(t, true, claims.IsNotMisadmin())
		assert.Equal(t, false, claims.IsMisadmin())

		// 3. a user without misadmin scope
		claims = AccessClaims{
			Scope:   "openid user charlatan",
			Subject: "2e9394f2-fd9f-4a07-beb5-748c35062cad",
		}
		assert.Equal(t, true, claims.IsNotMisadmin())
		assert.Equal(t, false, claims.IsMisadmin())
	})

	t.Run("test IsMisadminScope", func(t *testing.T) {
		assert.Equal(t, true, IsMisadminScope("misadmin"))
		assert.Equal(t, false, IsMisadminScope("rol.admin.2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, false, IsMisadminScope("rol.dpo.2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, false, IsMisadminScope("rol.pirate.2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, false, IsMisadminScope("openid"))
		assert.Equal(t, false, IsMisadminScope("user"))
		assert.Equal(t, false, IsMisadminScope("application"))
		assert.Equal(t, false, IsMisadminScope("service"))
		assert.Equal(t, false, IsMisadminScope("pur.minimum_required"))
		assert.Equal(t, false, IsMisadminScope("pur:minimum_required"))
	})

	t.Run("test IsAllowedScope", func(t *testing.T) {
		assert.Equal(t, true, IsAllowedScope("misadmin"))
		assert.Equal(t, true, IsAllowedScope("rol.admin.2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, true, IsAllowedScope("rol.dpo.2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, true, IsAllowedScope("openid"))
		assert.Equal(t, true, IsAllowedScope("user"))
		assert.Equal(t, true, IsAllowedScope("application"))
		assert.Equal(t, true, IsAllowedScope("service"))
		assert.Equal(t, true, IsAllowedScope("pur.minimum_required"))
		assert.Equal(t, true, IsAllowedScope("pur:minimum_required"))
		assert.Equal(t, false, IsAllowedScope("rol.pirate.2e9394f2-fd9f-4a07-beb5-748c35062cad"))
		assert.Equal(t, false, IsAllowedScope("bamba_triste"))
	})
}
