package events

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events/cache"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/notifications"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

// increment count for all identities except the sender
// and send event to all members
func notify(ctx context.Context, e *Event, exec boil.ContextExecutor, redConn *redis.Client, identities *IdentityMapper, _ files.FileStorageRepo, _ Metadata) error {
	// 1. retrieve member ids
	memberIDs, err := ListBoxMemberIDs(ctx, exec, redConn, e.BoxID)
	if err != nil {
		return merror.Transform(err).Describe("notifying member: listing members")
	}

	// delete the notification sender id from the list
	for i, id := range memberIDs {
		if id == e.SenderID {
			memberIDs = append(memberIDs[:i], memberIDs[i+1:]...)
			break
		}
	}

	// incr counts for a given box for all received identityIDs
	return IncrCounts(ctx, redConn, memberIDs, e.BoxID)

}

// send event to realtime channels
func publish(ctx context.Context, e *Event, exec boil.ContextExecutor, redConn *redis.Client, identities *IdentityMapper, _ files.FileStorageRepo, _ Metadata) error {
	memberIDs, err := ListBoxMemberIDs(ctx, exec, redConn, e.BoxID)
	if err != nil {
		return merror.Transform(err).Describe("notifying member: listing members")
	}

	// build a set to ensure unicity
	uniqRecipientsIDs := make(map[string]bool)
	for _, memberID := range memberIDs {
		uniqRecipientsIDs[memberID] = true
	}

	// add sender_id
	uniqRecipientsIDs[e.SenderID] = true

	// non-transparent mode for published events
	formattedEvent, err := e.Format(ctx, identities, false)
	if err != nil {
		return merror.Transform(err).Describe("formatting event")
	}

	for memberID := range uniqRecipientsIDs {
		bu := notifications.Update{
			Type:   "event.new",
			Object: formattedEvent,
		}
		notifications.SendUpdate(ctx, redConn, memberID, &bu)
	}

	return nil
}

func invalidateCaches(ctx context.Context, e *Event, exec boil.ContextExecutor, redConn *redis.Client, identities *IdentityMapper, _ files.FileStorageRepo, _ Metadata) error {
	_, err := redConn.Del(cache.GetBoxMembersKey(e.BoxID)).Result()
	if err != nil {
		logger.FromCtx(ctx).Warn().Msgf("could not invalidate cache %s:members", e.BoxID)
	}

	_, err = redConn.Del(cache.GetSenderBoxesKey(e.SenderID)).Result()
	if err != nil {
		logger.FromCtx(ctx).Warn().Msgf("could not invalidate cache %s:boxes", e.SenderID)
	}

	return nil
}

type DeletedBox struct {
	BoxID     string `json:"id"`
	SenderID  string `json:"sender_id"`
	PublicKey string `json:"public_key"`
}

func SendDeleteBox(ctx context.Context, redConn *redis.Client, boxID, senderID string, memberIDs []string, publicKey string) {
	deletedBox := DeletedBox{
		BoxID:     boxID,
		SenderID:  senderID,
		PublicKey: publicKey,
	}
	bu := notifications.Update{
		Type:   "box.delete",
		Object: deletedBox,
	}
	for _, memberID := range memberIDs {
		notifications.SendUpdate(ctx, redConn, memberID, &bu)
	}
}
