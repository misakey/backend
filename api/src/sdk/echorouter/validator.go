package echorouter

import "gitlab.misakey.dev/misakey/backend/api/src/sdk/validator"

// WrappedValidator wraps our enhanced Misakey validator around echo framework.
type WrappedValidator struct {
	validator *validator.Enhanced
}

// NewWrappedValidator's constructor.
func NewWrappedValidator(validator *validator.Enhanced) *WrappedValidator {
	return &WrappedValidator{validator: validator}
}

// Validate implements the echo framework validator interface.
func (wrapper *WrappedValidator) Validate(i interface{}) error {
	return wrapper.validator.EnhancedValidate(i)
}
