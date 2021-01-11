package events

import (
	"encoding/json"

	"gitlab.misakey.dev/misakey/backend/api/src/box/repositories/sqlboiler"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
)

// IsDeleted ...
func IsDeleted(event *sqlboiler.Event) (bool, error) {
	deletedContent := DeletedContent{}
	err := json.Unmarshal(event.Content.JSON, &deletedContent)
	if err != nil {
		return false, merr.Internal().Desc("unmarshaling content json")
	}

	if deletedContent.Deleted.AtTime.IsZero() {
		return false, nil
	}

	return true, nil
}
