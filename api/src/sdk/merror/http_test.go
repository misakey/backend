package merror

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleErr(t *testing.T) {
	tests := map[string]struct {
		inputErr     error
		expectedCode int
		expectedErr  error
	}{
		"raw error raises internal code error": {
			inputErr:     errors.New("raw error"),
			expectedCode: http.StatusInternalServerError,
			expectedErr:  Error{InternalError, InternalCode, OriNotDefined, "raw error", make(map[string]string), false},
		},
		"merror raises a badrequest code error": {
			inputErr:     BadRequest().Describe("uuid is required").From(OriBody),
			expectedCode: http.StatusBadRequest,
			expectedErr:  Error{BadRequestError, BadRequestCode, OriBody, "uuid is required", make(map[string]string), false},
		},
		"merror raises a client closed request code error": {
			inputErr:     ClientClosedRequest().Describe("canceled context").From(OriBody),
			expectedCode: StatusClientClosedRequest,
			expectedErr:  Error{ClientClosedRequestError, ClientClosedRequestCode, OriBody, "canceled context", make(map[string]string), false},
		},
	}
	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			resultCode, resultErr := HandleErr(test.inputErr)
			assert.Equal(t, test.expectedCode, resultCode)
			assert.Equal(t, test.expectedErr, resultErr)
		})
	}
}

func TestTransformHTTPCode(t *testing.T) {
	tests := map[string]struct {
		code        int
		expectedErr error
	}{
		"NotFound": {
			code:        404,
			expectedErr: NotFound(),
		},
		"Unknown": {
			code:        455,
			expectedErr: Internal(),
		},
		"Internal": {
			code:        500,
			expectedErr: Internal(),
		},
		"Conflict": {
			code:        409,
			expectedErr: Conflict(),
		},
		"ClientClosedRequest": {
			code:        499,
			expectedErr: ClientClosedRequest(),
		},
	}
	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			resultErr := TransformHTTPCode(test.code)
			assert.Equal(t, test.expectedErr, resultErr)
		})
	}
}
