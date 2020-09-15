package events

import (
	"context"

	"github.com/volatiletech/sqlboiler/boil"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

func MustBeAdmin(ctx context.Context, exec boil.ContextExecutor, boxID, senderID string) error {
	isCreator, err := isCreator(ctx, exec, boxID, senderID)
	if err != nil {
		return err
	}
	if !isCreator {
		return merror.Forbidden().Describe("not the creator")
	}
	return nil
}

func isAdmin(ctx context.Context, exec boil.ContextExecutor, boxID, senderID string) (bool, error) {
	err := MustBeAdmin(ctx, exec, boxID, senderID)
	if err != nil && merror.HasCode(err, merror.ForbiddenCode) {
		return false, nil
	}
	// return false admin if an error has occured
	return (err == nil), err
}
