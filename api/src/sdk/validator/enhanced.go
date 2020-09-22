package validator

import (
	"reflect"
	"strings"

	validator "gopkg.in/go-playground/validator.v9"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/patch"
)

// Enhanced wraps the go playground validator with some added Misakey rules.
type Enhanced struct {
	*validator.Validate
}

// New creates a new validator.
func New() *Enhanced {
	baseValidator := validator.New()
	// get alternate names for StructFields (use json tag) - to build details error
	baseValidator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Enhanced{baseValidator}
}

// EnhancedValidate the interface using the go playground validator package
// We handle Patch Validation using patch package
func (v *Enhanced) EnhancedValidate(i interface{}) error {
	patchInfo, ok := i.(*patch.Info)
	if ok {
		return v.StructPartial(patchInfo.Model, patchInfo.Input...)
	}
	return v.Struct(i)
}
