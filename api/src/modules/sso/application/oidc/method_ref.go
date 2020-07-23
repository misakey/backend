package oidc

import "strings"

// Official Authentication Method Reference (https://tools.ietf.org/html/rfc8176) enum
// used to store in the ID Token authentication methods that have been used to
// authenticate the user
type MethodRef string
type MethodRefs []MethodRef

const (
	AMRBrowserCookie     MethodRef = "browser_cookie"
	AMREmailedCode       MethodRef = "emailed_code"
	AMRPrehashedPassword MethodRef = "prehashed_password"
	AMRAccountCreation   MethodRef = "account_creation"
	AMRResetPassword     MethodRef = "reset_password"
)

func (amrs *MethodRefs) Add(method MethodRef) {
	*amrs = append(*amrs, method)
}

func (amrs MethodRefs) Has(method MethodRef) bool {
	for _, amr := range amrs {
		if method == amr {
			return true
		}
	}
	return false
}

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

func (amrs MethodRefs) String() string {
	tmp := make([]string, len(amrs))
	for i, v := range amrs {
		tmp[i] = string(v)
	}
	return strings.Join(tmp, " ")
}
