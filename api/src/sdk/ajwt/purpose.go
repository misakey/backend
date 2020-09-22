package ajwt

import "strings"

// TrimPurposes removes application purpose scopes from given list of scopes and return it
// It does not modify scope directly but return scopes after the trim as a slice of strings
func TrimPurposes(scopes []string) []string {
	ret := []string{}
	for _, sco := range scopes {
		// add scope if not an application purpose scope
		if !IsPurposeScope(sco) {
			ret = append(ret, sco)
		}
	}
	return ret
}

// IsPurposeScope return true if given candidate string is considered as a purpose scope.
func IsPurposeScope(candidate string) bool {
	return strings.HasPrefix(candidate, string(purposePrefix)) ||
		// to deprecate
		strings.HasPrefix(candidate, string(oldPurposePrefix))
}

// GetPurpose takes a purpose scope and return the application purpose IDs contained inside
// Warning: GetPurpose is not meant to receive anything else than a purpose scope
// there is no error handling about getting an invalid purpose scope
func GetPurpose(scope string) string {
	return strings.TrimPrefix(scope, "pur.")
}

// BuildPurposeScope takes a purpose label and return a purpose scope
func BuildPurposeScope(label string) string {
	return "pur." + label
}
