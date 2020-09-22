package bubble

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	validator "gopkg.in/go-playground/validator.v9"
)

type StructWithRequiredField struct {
	FName    string   `json:"json_tag" validate:"required"`
	Password string   `json:"password_tag" validate:"omitempty,len=10"`
	Enum     string   `json:"enum_tag" validate:"oneof=val1 val2"`
	UUIDs    []string `json:"u" validate:"dive,uuid"`
}

func TestValidatorExplode(t *testing.T) {
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

	err = ValidatorNeedle{}.Explode(err)
	mErr := err.(merror.Error)
	assert.Equalf(t, merror.BadRequestCode, mErr.Co, "code test")
	assert.Equalf(t, "required", string(mErr.Details["json_tag"]), "required test")
	assert.Equalf(t, "invalid", string(mErr.Details["password_tag"]), "length test")
	assert.Equalf(t, "malformed", string(mErr.Details["enum_tag"]), "oneof test")
	assert.Equalf(t, "malformed", string(mErr.Details["u"]), "dive,uuid test")
}
