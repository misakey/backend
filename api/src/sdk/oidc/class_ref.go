package oidc

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

func (acr ClassRef) LessThan(minimum ClassRef) bool {
	// empty acr is always lesser
	if acr == "" {
		return true
	}

	// ACR0 has only one equivalent: 0
	if acr == ACR0 && minimum != ACR0 {
		return true
	}

	// ACR1 is beaten if minimum is ACR2
	if acr == ACR1 && minimum == ACR2 {
		return true
	}

	// ACR2 can't be beaten
	return false
}

func NewClassRefs(acr ClassRef) ClassRefs {
	return []ClassRef{acr}
}

// Multiple ACRValues capability is ignored so it always takes the first one
func (acrs ClassRefs) Get() ClassRef {
	acr := ACR0
	if len(acrs) > 0 {
		switch acrs[0] {
		case ACR0, ACR1, ACR2:
			acr = acrs[0]
		}
	}
	return acr
}

// Multiple ACRValues capability is ignored
// the full slice is replaced
func (acrs *ClassRefs) Set(acr ClassRef) {
	*acrs = ClassRefs([]ClassRef{acr})
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
