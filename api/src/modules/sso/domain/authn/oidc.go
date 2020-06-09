package authn

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
)

func (amr *MethodRefs) Add(method MethodRef) {
	*amr = append(*amr, method)
}

func (amr MethodRefs) Has(method MethodRef) bool {
	for _, registeredMethod := range amr {
		if method == registeredMethod {
			return true
		}
	}
	return false
}

func (amr MethodRefs) String() string {
	var tmp []string
	for _, v := range amr {
		tmp = append(tmp, string(v))
	}
	return strings.Join(tmp, " ")
}

// Official Authentication Context Class Reference (https://openid.net/specs/openid-connect-core-1_0.html#IDToken) enum
// used to store in the ID & Access Tokens the context class the authentication satisfied
// higher is the more secure
type ClassRef string

const (
	ACR0 ClassRef = "0" // long-lived browser cookie
	ACR1 ClassRef = "1" // mca
	ACR2 ClassRef = "2" // pwd
)

func (acr ClassRef) String() string {
	return string(acr)
}

// Context format used to forward information to Open ID server
type Context map[string]string

func NewContext() Context {
	return make(map[string]string)
}

func (ctx Context) SetAMR(amr MethodRefs) Context {
	ctx["amr"] = amr.String()
	return ctx
}

func (ctx Context) GetAMR() []string {
	return strings.Split(ctx["amr"], " ")
}
