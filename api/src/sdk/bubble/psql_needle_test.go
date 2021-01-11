package bubble

import (
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
)

func TestACRIsGTE(t *testing.T) {
	tests := map[string]struct {
		inputErr    error
		expectedErr error
	}{
		"foreign_key pq error shall return a conflict merr": {
			inputErr:    &pq.Error{Code: "23503", Message: "dummy error"},
			expectedErr: merr.Conflict().Desc("pq: dummy error"),
		},
		"unique_violation pq error shall return a conflict merr": {
			inputErr:    &pq.Error{Code: "23505", Message: "dummy error"},
			expectedErr: merr.Conflict().Desc("pq: dummy error"),
		},
		"foreign_key_violation pq error shall return a conflict merr": {
			inputErr:    &pq.Error{Code: "23503", Message: "dummy error"},
			expectedErr: merr.Conflict().Desc("pq: dummy error"),
		},
		"invalid_text_representation pq error shall return a bad request merr": {
			inputErr:    &pq.Error{Code: "22P02", Message: "dummy error"},
			expectedErr: merr.BadRequest().Desc("pq: dummy error"),
		},
		"not_null_violation pq error shall return a bad request merr": {
			inputErr:    &pq.Error{Code: "23502", Message: "dummy error"},
			expectedErr: merr.BadRequest().Desc("pq: dummy error"),
		},
		"string_data_right_truncation pq error shall return a entity too large merr": {
			inputErr:    &pq.Error{Code: "01004", Message: "dummy error"},
			expectedErr: merr.RequestEntityTooLarge().Desc("pq: dummy error"),
		},
		"query_canceled pq error shall return a client closed request merr": {
			inputErr:    &pq.Error{Code: "57014", Message: "dummy error"},
			expectedErr: merr.ClientClosedRequest().Desc("pq: dummy error"),
		},
	}
	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			result := PSQLNeedle{}.Explode(test.inputErr)
			assert.Equal(t, test.expectedErr, result)
		})
	}
}
