package events

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
)

func doDeleteMsg(ctx context.Context, e *Event, exec boil.ContextExecutor, redConn *redis.Client, identities *IdentityMapper, filesRepo files.FileStorageRepo) (Metadata, error) {
	// Authorization-related checks should come as soon as possible
	// so we put them first.
	if err := MustMemberHaveAccess(ctx, exec, redConn, identities, e.BoxID, e.SenderID); err != nil {
		return nil, err
	}

	// check that the event contains a referrer_id
	if err := checkReferrer(*e); err != nil {
		return nil, err
	}

	msg, err := buildMessage(ctx, exec, e.ReferrerID.String)
	if err != nil {
		return nil, merror.Transform(err).Describe("building message")
	}

	if e.SenderID != msg.InitialSenderID {
		// box admins can delete messages even if they are not the author
		if err := MustBeAdmin(ctx, exec, msg.BoxID, e.SenderID); err != nil {
			return nil, merror.Transform(err).Describe("checking admins")
		}
	}
	// if the message is already deleted, do not go further
	if !msg.DeletedAt.IsZero() {
		return nil, merror.Gone().Describe("cannot delete an already deleted message")
	}

	if err := e.persist(ctx, exec); err != nil {
		return nil, err
	}
	// add the recently created event to the built message
	if err := msg.addEvent(*e); err != nil {
		return nil, err
	}

	// NOTE: side effect, could be in after handler
	// (potential) removal of the actual encrypted file (on S3)
	// is done at the very end because the operation cannot be rolled back
	if !msg.FileID.IsZero() {
		isOrphan, err := IsFileOrphan(ctx, exec, msg.FileID.String)
		if err != nil {
			return nil, merror.Transform(err).Describe("checking if file is orphan")
		}
		if isOrphan {
			if err := files.Delete(ctx, exec, filesRepo, msg.FileID.String); err != nil {
				return nil, merror.Transform(err).Describe("deleting stored file")
			}
		}
	}
	return &msg, nil
}
