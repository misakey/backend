package bubble

import (
	"database/sql"

	"github.com/lib/pq"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
)

// PSQLNeedle ...
type PSQLNeedle struct {
}

// Explode ...
func (n PSQLNeedle) Explode(err error) error {
	// try to consider error cause as pq error to understand deeper the error
	pqErr, ok := merr.Cause(err).(*pq.Error)
	if !ok {
		// still handle NotFound in case of sql errors
		if err == sql.ErrNoRows {
			return merr.NotFound().Desc(err.Error())
		}
		return nil
	}

	switch pqErr.Code.Name() {
	case "foreign_key", "unique_violation", "foreign_key_violation":
		return merr.Conflict().Desc(err.Error())
	case "invalid_text_representation", "not_null_violation":
		return merr.BadRequest().Desc(err.Error())
	case "string_data_right_truncation":
		return merr.RequestEntityTooLarge().Desc(err.Error())
	case "query_canceled":
		return merr.ClientClosedRequest().Desc(err.Error())
	case "too_many_connections":
		return merr.ServiceUnavailable().Desc(err.Error())
	}
	return err
}
