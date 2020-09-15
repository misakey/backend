package events

import (
	"context"

	"github.com/volatiletech/sqlboiler/boil"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events/etype"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
)

func KickDeprecatedMembers(
	ctx context.Context,
	exec boil.ContextExecutor, identities entrypoints.IdentityIntraprocessInterface,
	boxID string, kickerID string,
) ([]Event, error) {
	var kicks []Event

	// 1. list action joins then iterate on it and verify acces is still granted
	activeJoins, err := listBoxActiveJoinEvents(ctx, exec, boxID)
	if err != nil {
		return kicks, err
	}

	// 2. check if we must kick active joins
	for _, joinEvent := range activeJoins {
		if err := MustHaveAccess(ctx, exec, identities, boxID, joinEvent.SenderID); err != nil {
			// if the member has no access anymore then kick them by creation a member.kick event
			if merror.HasCode(err, merror.ForbiddenCode) {
				kickEvent, err := New(etype.Memberkick, nil, boxID, kickerID, &joinEvent.ID)
				if err != nil {
					return kicks, merror.Transform(err).Describe("newing kick event")
				}
				if err := kickEvent.persist(ctx, exec); err != nil {
					return kicks, err
				}
				kicks = append(kicks, kickEvent)
				continue
			}

			return kicks, err
		}
	}
	return kicks, nil
}
