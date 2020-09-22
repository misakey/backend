package bubble

import (
	"database/sql"

	"github.com/lib/pq"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

type PSQLNeedle struct {
}

func (n PSQLNeedle) Explode(err error) error {
	// try to consider error cause as pq error to understand deeper the error
	pqErr, ok := merror.Cause(err).(*pq.Error)
	if !ok {
		// still handle NotFound in case of sql errors
		if err == sql.ErrNoRows {
			return merror.NotFound().Describe(err.Error())
		}
		return nil
	}

	switch pqErr.Code.Name() {
	case "foreign_key", "unique_violation", "foreign_key_violation":
		return merror.Conflict().Describe(err.Error())
	case "invalid_text_representation", "not_null_violation":
		return merror.BadRequest().Describe(err.Error())
	case "string_data_right_truncation":
		return merror.RequestEntityTooLarge().Describe(err.Error())
	case "query_canceled":
		return merror.ClientClosedRequest().Describe(err.Error())
	case "too_many_connections":
		return merror.ServiceUnavailable().Describe(err.Error())
	}
	return err
}
