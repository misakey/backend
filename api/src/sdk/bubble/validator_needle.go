package bubble

import (
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	validator "gopkg.in/go-playground/validator.v9"
)

type ValidatorNeedle struct {
}

func (n ValidatorNeedle) Explode(err error) error {
	// try to consider error cause as validator error to understand deeper the error
	valErrs, ok := merror.Cause(err).(validator.ValidationErrors)
	if !ok {
		return nil
	}

	mErr := merror.Transform(err).Code(merror.BadRequestCode)
	for _, valErr := range valErrs {
		// get, format key
		key := valErr.Field()
		// remove slice position if slice detected
		// an array with a dive rule will have `field_name[error_pos]` as Field() value
		// we only want field_name to return as it is the real key - pos info is ignored today
		if len(key) > 3 &&
			key[len(key)-3] == '[' &&
			key[len(key)-1] == ']' {
			key = key[:len(key)-3]
		}

		// add details corresponding to tags
		switch valErr.ActualTag() {
		case "required":
			mErr = mErr.Detail(key, merror.DVRequired)
		case "oneof", "uuid", "email":
			mErr = mErr.Detail(key, merror.DVMalformed)
		default:
			mErr = mErr.Detail(key, merror.DVInvalid)
		}
	}
	return mErr
}
