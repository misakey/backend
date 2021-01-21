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
	// AMRBrowserCookie is the use of browser cookie to store an auth session
	AMRBrowserCookie MethodRef = "browser_cookie"
	// AMREmailedCode is the entering of a code received by email
	AMREmailedCode MethodRef = "emailed_code"
	// AMRPrehashedPassword is the entering of a password
	AMRPrehashedPassword MethodRef = "prehashed_password"
	// AMRAccountCreation is the creation of an account
	AMRAccountCreation MethodRef = "account_creation"
	// AMRResetPassword is the use of reset password flow
	AMRResetPassword MethodRef = "reset_password"
	// AMRTOTP is the use of a totp
	AMRTOTP MethodRef = "totp"
	// AMRWebauthn is the use of webauthn protocol
	AMRWebauthn MethodRef = "webauthn"
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
// password + webauthn = acr 4
// password + totp: acr 3
// password, reset password, account creation: acr 2
// emailed code: acr 1
func (amrs MethodRefs) ToACR() ClassRef {
	// acr 4
	if amrs.Has(AMRWebauthn) && amrs.Has(AMRPrehashedPassword) {
		return ACR4
	}

	// acr 3
	if amrs.Has(AMRTOTP) && amrs.Has(AMRPrehashedPassword) {
		return ACR3
	}

	// acr 2
	if amrs.Has(AMRPrehashedPassword) ||
		amrs.Has(AMRResetPassword) ||
		amrs.Has(AMRAccountCreation) {
		return ACR2
	}

	// acr 1
	if amrs.Has(AMREmailedCode) {
		return ACR1
	}

	// acr 0
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
