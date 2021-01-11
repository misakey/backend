package bubble

import (
	"testing"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/stretchr/testify/assert"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
)

type StructToValidate struct {
	FName    string   `json:"json_tag"`
	Password string   `json:"password_tag"`
	Enum     string   `json:"enum_tag"`
	UUIDs    []string `json:"u"`
	Email    string   `json:"email"`
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"already_snake", "already_snake"},
		{"A", "a"},
		{"AA", "aa"},
		{"AaAa", "aa_aa"},
		{"HTTPRequest", "http_request"},
		{"ThisTestHasBeenCopyPasted", "this_test_has_been_copy_pasted"},
		{"With23Ya", "with23_ya"},
		{"URL43324Hu", "url43324_hu"},
	}
	for _, test := range tests {
		have := NewOzzoNeedle().toSnakeCase(test.input)
		assert.Equalf(t, test.want, have, "wrong snake case")
	}
}

func TestOzzoExplode(t *testing.T) {
	// test with fully invalid structure
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
		v.Field(&data.Email, v.Required, is.EmailFormat),
	)

	err = NewOzzoNeedle().Explode(err)
	mErr := err.(merr.Error)
	assert.Equalf(t, merr.BadRequestCode, mErr.Co, "code test")
	assert.Equalf(t, merr.DVRequired, mErr.Details["json_tag"], "required test")
	assert.Equalf(t, merr.DVMalformed, mErr.Details["password_tag"], "length test")
	assert.Equalf(t, merr.DVMalformed, mErr.Details["enum_tag"], "enum test")
	assert.Equalf(t, merr.DVMalformed, mErr.Details["u"], "slice of uuid test")
	assert.Equalf(t, merr.DVMalformed, mErr.Details["email"], "email test")

	// test with a valid structure - mostly for the email format validation
	// the library plans to invalidate the uppercase characters in the is.Email
	// rule so we want to detect it while upgrading the library
	data = StructToValidate{
		FName:    "NS",
		Password: "1234567890",
		Enum:     "val1",
		UUIDs:    []string{"42d9f0e3-68bf-45e1-b77a-023087d69586"},
		Email:    "ThisEmail@hAs.uppErCase",
	}
	err = v.ValidateStruct(&data,
		v.Field(&data.FName, v.Required),
		v.Field(&data.Password, v.Length(10, 10)),
		v.Field(&data.Enum, v.In("val1", "val2")),
		v.Field(&data.UUIDs, v.Each(is.UUIDv4)),
		v.Field(&data.Email, v.Required, is.EmailFormat),
	)

	err = NewOzzoNeedle().Explode(err)
	assert.Nilf(t, err, "validating the clean structure")
}
