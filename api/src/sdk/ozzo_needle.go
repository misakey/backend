package sdk

import (
	"fmt"
	"reflect"

	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	v "github.com/go-ozzo/ozzo-validation/v4"
)

type OzzoNeedle struct {
}

func (n OzzoNeedle) Explode(err error) error {
	// try to consider error cause as validator error to understand deeper the error
	valErrs, ok := merror.Cause(err).(v.Errors)
	if !ok {
		fmt.Println(reflect.TypeOf(merror.Cause(err)))
		return nil
	}

	mErr := merror.Transform(err).Code(merror.BadRequestCode)
	return recHandleErrors(mErr, valErrs, nil)
}

func recHandleErrors(mErr merror.Error, valErrs v.Errors, replaceFieldTag *string) merror.Error {
	// v.Errors is basically a mao["structure_tag"]error, we parse it
	for fieldTag, valErr := range valErrs {
		if replaceFieldTag != nil {
			fieldTag = *replaceFieldTag
		}

		// v.Errors can be nested - for slice validation as an example since
		// errors can be different between index 0 and 1
		// it is case we use recursive to handle it and override the fieldTag
		// future fieldTags for slice are index, we don't want to map detail on indexes
		if reValErrs, ok := valErr.(v.Errors); ok {
			mErr = recHandleErrors(mErr, reValErrs, &fieldTag)
			continue
		}

		// v.ErrorObject is the final object we want to examinate to set details
		errObj, ok := valErr.(v.ErrorObject)
		if !ok {
			continue
		}

		switch errObj.Code() {
		case
			"validation_in_invalid",
			"validation_length_invalid",
			"validation_length_too_long",
			"validation_length_out_of_range",
			"validation_length_empty_required",
			"validation_is_email",
			"validation_is_uuid_v4":
			mErr.Detail(fieldTag, merror.DVMalformed)
		case
			"validation_required",
			"validation_nil_or_not_empty_required":
			mErr.Detail(fieldTag, merror.DVRequired)
		default:
			mErr.Detail(fieldTag, merror.DVInvalid).Detail("please_inform_about_this_unknown_code", errObj.Code())
		}
	}
	return mErr
}
