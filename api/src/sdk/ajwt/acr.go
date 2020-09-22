package ajwt

import "gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

// Security Level - Authentication Context Class Reference (a.k.a. ACR values)
// Corresponding to a level of certainty, while the token has been generated,
// of the identity of the token requester
// The higher the security level is, stronger is the possibility the subject claims
// correspond in reality to the person who generated an access token
type ACRSecLvl string

func (s ACRSecLvl) String() string {
	return string(s)
}

var (
	noSecLevel ACRSecLvl = "0"
	ACRSecLvl1 ACRSecLvl = "1"
	ACRSecLvl2 ACRSecLvl = "2"
)

// ACRIsGTE compares AccessClaims ACR (Authentication Context Class - Sec Level) with the secLvl parameter
// It uses kind of >= operator principle and return an error if the current ACR is stricty inferior to asked one
func (c *AccessClaims) ACRIsGTE(secLvl ACRSecLvl) error {
	if c.ACR == noSecLevel {
		return merror.Forbidden().From(merror.OriACR).
			Describe("token acr is 0").
			Detail("acr", merror.DVForbidden).
			Detail("required_acr", secLvl.String())
	}

	// highest sec level is by definition always greater than other or equal
	if c.ACR == ACRSecLvl2 {
		return nil
	}

	// lowest sec level is by definition always smaller than other so always valid
	if secLvl == ACRSecLvl1 {
		return nil
	}

	// there is no other combinaison to check today
	return merror.Forbidden().From(merror.OriACR).
		Describe("token acr is too weak").
		Detail("acr", merror.DVForbidden).Detail("required_acr", secLvl.String())
}
