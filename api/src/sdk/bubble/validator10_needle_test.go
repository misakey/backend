package bubble

import (
	"reflect"
	"strings"
	"testing"

	validator "github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

func TestValidator10Explode(t *testing.T) {
	// init validator
	val := validator.New()
	val.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	err := val.Struct(StructWithRequiredField{Password: "1", Enum: "val3", UUIDs: []string{"u3132"}})

	err = Validator10Needle{}.Explode(err)
	mErr := err.(merror.Error)
	assert.Equalf(t, merror.BadRequestCode, mErr.Co, "code test")
	assert.Equalf(t, "required", string(mErr.Details["json_tag"]), "required test")
	assert.Equalf(t, "invalid", string(mErr.Details["password_tag"]), "length test")
	assert.Equalf(t, "malformed", string(mErr.Details["enum_tag"]), "oneof test")
	assert.Equalf(t, "malformed", string(mErr.Details["u"]), "dive,uuid test")
}
