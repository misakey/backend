package events

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/box/external"
	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
)

func doMessage(ctx context.Context, e *Event, _ null.JSON, exec boil.ContextExecutor, redConn *redis.Client, _ *IdentityMapper, _ external.CryptoRepo, _ files.FileStorageRepo) (Metadata, error) {
	// check that the current sender is a member of the box
	if err := MustBeMember(ctx, exec, redConn, e.BoxID, e.SenderID); err != nil {
		return nil, err
	}

	if e.ReferrerID.Valid {
		return nil, merror.BadRequest().Describe("referrer id cannot be set").Detail("referrer_id", merror.DVForbidden)
	}

	if err := e.persist(ctx, exec); err != nil {
		return nil, err
	}

	msg, err := buildMessage(ctx, exec, e.ID)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
