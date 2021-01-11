package merr

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
			expectedErr:  Error{ErrInternal, InternalCode, OriNotDefined, "raw error", make(map[string]string)},
		},
		"merr raises a badrequest code error": {
			inputErr:     BadRequest().Desc("uuid is required").Ori(OriBody),
			expectedCode: http.StatusBadRequest,
			expectedErr:  Error{ErrBadRequest, BadRequestCode, OriBody, "uuid is required", make(map[string]string)},
		},
		"merr raises a client closed request code error": {
			inputErr:     ClientClosedRequest().Desc("canceled context").Ori(OriBody),
			expectedCode: StatusClientClosedRequest,
			expectedErr:  Error{ErrClientClosedRequest, ClientClosedRequestCode, OriBody, "canceled context", make(map[string]string)},
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
