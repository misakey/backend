package oidc

// ClassRef Official Authentication Context Class Reference (https://openid.net/specs/openid-connect-core-1_0.html#IDToken) enum
// used to store in the ID & Access Tokens the context class the authentication satisfied
// higher is the more secure
type ClassRef string

// ClassRefs ...
type ClassRefs []ClassRef

const (
	// ACR0 ...
	ACR0 ClassRef = "0"
	// ACR1 ...
	ACR1 ClassRef = "1"
	// ACR2 ...
	ACR2 ClassRef = "2"
	// ACR3 ...
	ACR3 ClassRef = "3"
	// ACR4 ...
	ACR4 ClassRef = "4"
)

// String ...
func (acr ClassRef) String() string {
	return string(acr)
}

var acrToInt = map[ClassRef]int{
	ACR0: 0,
	ACR1: 1,
	ACR2: 2,
	ACR3: 3,
	ACR4: 4,
}

// LessThan ...
func (acr ClassRef) LessThan(minimum ClassRef) bool {
	return acrToInt[acr] < acrToInt[minimum]
}

// NewClassRefs ...
func NewClassRefs(acr ClassRef) ClassRefs {
	return []ClassRef{acr}
}

// Get the highest acr values
func (acrs ClassRefs) Get() ClassRef {
	max := ACR0
	for _, acr := range acrs {
		if max.LessThan(acr) {
			max = acr
		}
	}
	return max
}

// Set ...
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
	case ACR2, ACR3, ACR4:
		return 2592000 // 30d
	}
	return 0
}

var methodToACR = map[MethodRef]ClassRef{
	AMREmailedCode:       ACR1,
	AMRPrehashedPassword: ACR2,
	AMRTOTP:              ACR3,
	AMRWebauthn:          ACR4,
}

// GetMethodACR based on the static map above
func GetMethodACR(methodRefStr string) ClassRef {
	return methodToACR[MethodRef(methodRefStr)]
}

var acrToMethod = map[ClassRef]MethodRef{
	ACR1: AMREmailedCode,
	ACR2: AMRPrehashedPassword,
	ACR3: AMRTOTP,
	ACR4: AMRWebauthn,
}

// GetNextMethord returns next expected authn method
// and return nil if no next method is expected
func GetNextMethod(currentACR ClassRef, expectedACR ClassRef) *MethodRef {
	// if the current ACR equals or is higher than the expected,
	// there is no more authn method to perform
	if acrToInt[currentACR] >= acrToInt[expectedACR] {
		return nil
	}

	method, ok := acrToMethod[expectedACR]
	if ok {
		return &method
	}
	return nil
}
