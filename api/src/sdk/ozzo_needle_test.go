package sdk

import (
	"testing"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/stretchr/testify/assert"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type StructToValidate struct {
	FName    string   `json:"json_tag"`
	Password string   `json:"password_tag"`
	Enum     string   `json:"enum_tag"`
	UUIDs    []string `json:"u"`
	Email    string   `json:"email"`
}

func TestOzzoExplode(t *testing.T) {
	// init structure
	data := StructToValidate{
		Password: "1",
		Enum:     "val3",
		UUIDs:    []string{"u3132"},
		Email:    "invalidemail",
	}

	// init validator
	err := v.ValidateStruct(&data,
		v.Field(&data.FName, v.Required),
		v.Field(&data.Password, v.Length(10, 10)),
		v.Field(&data.Enum, v.In("val1", "val2")),
		v.Field(&data.UUIDs, v.Each(is.UUIDv4)),
		v.Field(&data.Email, v.Required, is.Email),
	)

	err = OzzoNeedle{}.Explode(err)
	mErr := err.(merror.Error)
	assert.Equalf(t, merror.BadRequestCode, mErr.Co, "code test")
	assert.Equalf(t, merror.DVRequired, string(mErr.Details["json_tag"]), "required test")
	assert.Equalf(t, merror.DVMalformed, string(mErr.Details["password_tag"]), "length test")
	assert.Equalf(t, merror.DVMalformed, string(mErr.Details["enum_tag"]), "enum test")
	assert.Equalf(t, merror.DVMalformed, string(mErr.Details["u"]), "slice of uuid test")
	assert.Equalf(t, merror.DVInvalid, string(mErr.Details["email"]), "email test")
}
