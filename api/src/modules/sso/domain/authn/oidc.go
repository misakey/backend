package authn

import (
	"strings"
)

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

func (amrs MethodRefs) String() string {
	tmp := make([]string, len(amrs))
	for i, v := range amrs {
		tmp[i] = string(v)
	}
	return strings.Join(tmp, " ")
}

// Official Authentication Context Class Reference (https://openid.net/specs/openid-connect-core-1_0.html#IDToken) enum
// used to store in the ID & Access Tokens the context class the authentication satisfied
// higher is the more secure
type ClassRef string
type ClassRefs []ClassRef

const (
	ACR0 ClassRef = "0" // long-lived browser cookie
	ACR1 ClassRef = "1" // mca
	ACR2 ClassRef = "2" // pwd
)

func (acr ClassRef) String() string {
	return string(acr)
}

// Multiple ACRValues capability is ignored so it always takes the first one
func (acrs ClassRefs) Get() ClassRef {
	if len(acrs) > 0 {
		return acrs[0]
	}
	return ACR1
}

// Multiple ACRValues capability is ignored
func (acrs *ClassRefs) Set(acr ClassRef) {
	if acrs == nil {
		*acrs = []ClassRef{acr}
	} else {
		(*acrs) = append(*acrs, acr)
	}
}

// RememberFor return an integer corresponding to seconds, according to the authentication context class
func (acr ClassRef) RememberFor() int {
	switch acr {
	case ACR1:
		return 3600 // 1h
	case ACR2:
		return 2592000 // 30d
	}
	return 0
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
