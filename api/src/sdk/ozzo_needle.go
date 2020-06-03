package sdk

import (
	"regexp"
	"strings"

	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	v "github.com/go-ozzo/ozzo-validation/v4"
)

type ozzoNeedle struct {
	matchFirstCap *regexp.Regexp
	matchAllCap   *regexp.Regexp
}

func NewOzzoNeedle() ozzoNeedle {
	return ozzoNeedle{
		matchFirstCap: regexp.MustCompile("(.)([A-Z][a-z]+)"),
		matchAllCap:   regexp.MustCompile("([a-z0-9])([A-Z])"),
	}
}

func (n ozzoNeedle) toSnakeCase(str string) string {
	if n.matchFirstCap == nil {
		return "ozzo_needle_wrongly_allocated!"
	}
	snake := n.matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = n.matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func (n ozzoNeedle) Explode(err error) error {
	// try to consider error cause as validator error to understand deeper the error
	valErrs, ok := merror.Cause(err).(v.Errors)
	if !ok {
		return nil
	}

	mErr := merror.Transform(err).Code(merror.BadRequestCode)
	return n.recHandleErrors(mErr, valErrs, nil)
}

func (n ozzoNeedle) recHandleErrors(mErr merror.Error, valErrs v.Errors, replaceFieldTag *string) merror.Error {
	// v.Errors is basically a mao["structure_tag"]error, we parse it
	for fieldTag, valErr := range valErrs {
		if replaceFieldTag != nil {
			fieldTag = *replaceFieldTag
		} else {
			// we ensure the field is snake case
			fieldTag = n.toSnakeCase(fieldTag)
		}
		// v.Errors can be nested - for slice validation as an example since
		// errors can be different between index 0 and 1
		// it is case we use recursive to handle it and override the fieldTag
		// future fieldTags for slice are index, we don't want to map detail on indexes
		if reValErrs, ok := valErr.(v.Errors); ok {
			mErr = n.recHandleErrors(mErr, reValErrs, &fieldTag)
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
