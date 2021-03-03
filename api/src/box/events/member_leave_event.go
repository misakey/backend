package events

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events/etype"
	"gitlab.misakey.dev/misakey/backend/api/src/box/external"
	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
)

func doLeave(ctx context.Context, e *Event, _ null.JSON, exec boil.ContextExecutor, redConn *redis.Client, _ *IdentityMapper, _ external.CryptoRepo, _ files.FileStorageRepo) (Metadata, error) {
	// check that the current sender has access to the box
	if err := MustBeMember(ctx, exec, redConn, e.BoxID, e.SenderID); err != nil {
		// user is a not a box member
		// so we just return
		return nil, err
	}

	// check that the current sender is not the admin
	// admin can’t leave their own box
	if err := MustBeAdmin(ctx, exec, e.BoxID, e.SenderID); err == nil {
		return nil, merr.Forbidden().Desc("admin can’t leave their own box")
	}

	// get the last join event to set the referrer id
	joinEvent, err := get(ctx, exec, eventFilters{
		eType:    null.StringFrom(etype.Memberjoin),
		senderID: null.StringFrom(e.SenderID),
		boxID:    null.StringFrom(e.BoxID),
		// exclude referred joins
		excludeOnRef: &referentsFilters{
			eTypes:   []string{etype.Memberleave, etype.Memberkick},
			senderID: null.StringFrom(e.SenderID),
			boxID:    null.StringFrom(e.BoxID),
		},
	})
	if err != nil {
		return nil, merr.From(err).Desc("getting last join event")
	}
	e.ReferrerID = null.StringFrom(joinEvent.ID)

	return nil, e.persist(ctx, exec)
}
