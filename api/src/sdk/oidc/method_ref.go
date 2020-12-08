package oidc

import "strings"

// MethodRef ...
// Official Authentication Method Reference (https://tools.ietf.org/html/rfc8176) enum
// used to store in the ID Token authentication methods that have been used to
// authenticate the user
type MethodRef string

// MethodRefs ...
type MethodRefs []MethodRef

const (
	// AMRBrowserCookie ...
	AMRBrowserCookie MethodRef = "browser_cookie"
	// AMREmailedCode ...
	AMREmailedCode MethodRef = "emailed_code"
	// AMRPrehashedPassword ...
	AMRPrehashedPassword MethodRef = "prehashed_password"
	// AMRAccountCreation ...
	AMRAccountCreation MethodRef = "account_creation"
	// AMRResetPassword ...
	AMRResetPassword MethodRef = "reset_password"
)

// Add ...
func (amrs *MethodRefs) Add(method MethodRef) {
	*amrs = append(*amrs, method)
}

// Has ...
func (amrs MethodRefs) Has(method MethodRef) bool {
	for _, amr := range amrs {
		if method == amr {
			return true
		}
	}
	return false
}

// ToACR ...
func (amrs MethodRefs) ToACR() ClassRef {
	if amrs.Has(AMRPrehashedPassword) ||
		amrs.Has(AMRResetPassword) ||
		amrs.Has(AMRAccountCreation) {
		return ACR2
	}
	if amrs.Has(AMREmailedCode) {
		return ACR1
	}
	return ACR0
}

// String ...
func (amrs MethodRefs) String() string {
	tmp := make([]string, len(amrs))
	for i, v := range amrs {
		tmp[i] = string(v)
	}
	return strings.Join(tmp, " ")
}
