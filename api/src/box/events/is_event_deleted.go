package events

import (
	"encoding/json"

	"gitlab.misakey.dev/misakey/backend/api/src/box/repositories/sqlboiler"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

func IsDeleted(event *sqlboiler.Event) (bool, error) {
	deletedContent := DeletedContent{}
	err := json.Unmarshal(event.Content.JSON, &deletedContent)
	if err != nil {
		return false, merror.Internal().Describe("unmarshaling content json")
	}

	if deletedContent.Deleted.AtTime.IsZero() {
		return false, nil
	}

	return true, nil
}
