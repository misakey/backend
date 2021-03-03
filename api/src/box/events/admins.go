package events

import (
	"context"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events/etype"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
)

// MustBeAdmin ...
func MustBeAdmin(ctx context.Context, exec boil.ContextExecutor, boxID, senderID string) error {
	isCreator, err := isCreator(ctx, exec, boxID, senderID)
	if err != nil {
		return err
	}
	if !isCreator {
		return merr.Forbidden().Desc("not the creator")
	}
	return nil
}

// IsAdmin ...
func IsAdmin(ctx context.Context, exec boil.ContextExecutor, boxID, senderID string) (bool, error) {
	err := MustBeAdmin(ctx, exec, boxID, senderID)
	// return no error if it is a forbidden, just set admin boolean to false
	if merr.IsAForbidden(err) {
		return false, nil
	}
	// still return false if an error has occurred
	return (err == nil), err
}

// GetAdminID ...
func GetAdminID(ctx context.Context, exec boil.ContextExecutor, boxID string) (string, error) {
	createEvent, err := get(ctx, exec, eventFilters{
		eType: null.StringFrom(etype.Create),
		boxID: null.StringFrom(boxID),
	})
	if err != nil {
		return "", err
	}

	return createEvent.SenderID, nil
}
