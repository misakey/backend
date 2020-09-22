package ajwt

import (
	"strings"
)

// IsRoleScope return true if given candidate string is considered as a role scope.
// Forbidden usage of * in role scopes.
func IsRoleScope(candidate string) bool {
	return (strings.HasPrefix(candidate, string(adminRolePrefix)) ||
		strings.HasPrefix(candidate, string(dpoRolePrefix))) && !strings.Contains(candidate, "*")
}

// GetRole takes a role scope and return the appID and the roleID contained inside
// Warning: GetRole is not meant to receive anything else than a role scope
// there is no error handling about getting an invalid role scope
func GetRole(scope string) AppRole {
	elems := strings.Split(scope, ".")
	if len(elems) != 3 {
		return AppRole{}
	}

	return AppRole{
		RoleLabel:     elems[1],
		ApplicationID: elems[2],
	}
}

// hasRole returns true if the app & role couple is found inside scopes
// it also checks claims have user scopes since only a user can own a role
func (c *AccessClaims) hasRole(appID string, roleLabel string) bool {
	var role strings.Builder
	_, _ = role.WriteString("rol.")
	_, _ = role.WriteString(roleLabel)
	_, _ = role.WriteString(".")
	_, _ = role.WriteString(appID)
	return !c.IsNotAnyUser() && c.hasStrScope(role.String())
}

func (c *AccessClaims) IsNotAdminOn(appID string) bool {
	return !c.hasRole(appID, adminRoleLabel)
}

func (c *AccessClaims) IsAdminOn(appID string) bool {
	return c.hasRole(appID, adminRoleLabel)
}

func (c *AccessClaims) IsAnyAdmin() bool {
	return !c.IsNotAdmin()
}

func (c *AccessClaims) IsNotAdmin() bool {
	if c.IsNotAnyUser() {
		return true
	}
	scopes := strings.Split(c.Scope, " ")
	for _, scope := range scopes {
		role := GetRole(scope)
		if role.RoleLabel == adminRoleLabel {
			return false
		}
	}
	return true
}

func (c *AccessClaims) IsNotDPOOn(appID string) bool {
	return !c.hasRole(appID, dpoRoleLabel)
}

func (c *AccessClaims) IsDPOOn(appID string) bool {
	return c.hasRole(appID, dpoRoleLabel)
}

func (c *AccessClaims) IsAnyDPO() bool {
	return !c.IsNotDPO()
}

func (c *AccessClaims) IsNotDPO() bool {
	if c.IsNotAnyUser() {
		return true
	}
	scopes := strings.Split(c.Scope, " ")
	for _, scope := range scopes {
		role := GetRole(scope)
		if role.RoleLabel == dpoRoleLabel {
			return false
		}
	}
	return true
}

// GetDPOAppID return application on which the user is DPO on
// if not found, nil is returned
func (c *AccessClaims) GetDPOAppID() *string {
	scopes := strings.Split(c.Scope, " ")
	for _, scope := range scopes {
		role := GetRole(scope)
		if role.RoleLabel == dpoRoleLabel {
			return &role.ApplicationID
		}
	}
	return nil
}

// GetAdminAppID return application on which the user is Admin on
// if not found, nil is returned
func (c *AccessClaims) GetAdminAppID() *string {
	scopes := strings.Split(c.Scope, " ")
	for _, scope := range scopes {
		role := GetRole(scope)
		if role.RoleLabel == adminRoleLabel {
			return &role.ApplicationID
		}
	}
	return nil
}
