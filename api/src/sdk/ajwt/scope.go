package ajwt

import (
	"strings"
)

// IsNotUser returns false if claims' scopes defines a user and its subject match given user ID
func (c *AccessClaims) IsNotUser(userID string) bool {
	return !c.IsUser(userID)
}

// IsUser returns true if claims' scopes defines a user and its subject match given user ID
func (c *AccessClaims) IsUser(userID string) bool {
	// check it is at least a user
	if c.IsNotAnyUser() {
		return false
	}
	// check the subject is not empty
	if c.Subject == "" {
		return false
	}
	// check the input is not empty
	if userID == "" {
		return false
	}
	// check if subject and input don't match
	if c.Subject != userID {
		return false
	}
	return true
}

// IsAnyUser returns false if claims' scopes defines a user
func (c *AccessClaims) IsNotAnyUser() bool {
	return !c.hasScope(userScope)
}

// IsAnyUser returns true if claims' scopes defines a user
func (c *AccessClaims) IsAnyUser() bool {
	return c.hasScope(userScope)
}

// IsNotApp returns false if claims' scopes defines an application and its subject match given app ID
func (c *AccessClaims) IsNotApp(appID string) bool {
	return !c.IsApp(appID)
}

// IsApp returns true if claims' scopes defines an application and its subject match given app ID
func (c *AccessClaims) IsApp(appID string) bool {
	if c.IsNotAnyApp() {
		return false
	}
	// check the subject is not empty
	if c.Subject == "" {
		return false
	}
	// check the input is not empty
	if appID == "" {
		return false
	}
	if c.Subject != appID {
		return false
	}
	return true
}

// IsNotAnyApp returns false if claims' scopes defines an application
func (c *AccessClaims) IsNotAnyApp() bool {
	return !c.hasScope(applicationScope)
}

// IsAnyApp returns false if claims' scopes defines an application
func (c *AccessClaims) IsAnyApp() bool {
	return c.hasScope(applicationScope)
}

// IsNotService returns false if claims' scopes defines a service and its subject match given service ID
func (c *AccessClaims) IsNotService(serviceID string) bool {
	return !c.IsService(serviceID)
}

// IsService returns true if claims' scopes defines a service and its subject match given service ID
func (c *AccessClaims) IsService(serviceID string) bool {
	if c.IsNotAnyService() {
		return false
	}
	// check the subject is not empty
	if c.Subject == "" {
		return false
	}
	// check the input is not empty
	if serviceID == "" {
		return false
	}
	if c.Subject != serviceID {
		return false
	}
	return true
}

// IsNotAnyService returns false if claims' scopes defines a service
func (c *AccessClaims) IsNotAnyService() bool {
	return !c.hasScope(serviceScope)
}

// IsAnyService returns true if claims' scopes defines a service
func (c *AccessClaims) IsAnyService() bool {
	return c.hasScope(serviceScope)
}

// IsNotMisadmin returns false if claims' scopes defines a user & a misadmin
func (c *AccessClaims) IsNotMisadmin() bool {
	return !c.IsMisadmin()
}

// IsMisadmin returns true if claims' scopes defines a user & a misadmin
func (c *AccessClaims) IsMisadmin() bool {
	if c.IsNotAnyUser() {
		return false
	}
	if !c.hasScope(misadminScope) {
		return false
	}
	return true
}

// IsMisadminScope return true if given scope string is considered as a misadmin scope.
func IsMisadminScope(candidate string) bool {
	return scope(candidate) == misadminScope
}

// IsAllowedScope return true if given candidate string is considered as an allowed scope
// A scope can be:
// - openid scope
// - caller type scope: user|service|application
// - misadmin scope
// - role scope
// - purpose scope
func IsAllowedScope(candidate string) bool {
	sco := scope(candidate)
	return sco == openIDScope ||
		sco == userScope ||
		sco == applicationScope ||
		sco == serviceScope ||
		sco == misadminScope ||
		IsPurposeScope(candidate) ||
		IsRoleScope(candidate)
}

// hasScope returns true if the candidate scope is contained inside the access claims scope list
// scope list is a space-separatated list string
func (c *AccessClaims) hasScope(candidate scope) bool {
	return c.hasStrScope(string(candidate))
}

func (c *AccessClaims) hasStrScope(candidate string) bool {
	return strings.Contains(c.Scope, candidate)
}
