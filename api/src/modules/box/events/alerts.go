package events

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events/cache"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/realtime"
)

//
// this file contains handlers about alerting end-user about some activities in boxes
// type of alerts:
// - boxCount: count the number of event that has occured in a box for a given user, displayed to the end-user in-app.
// - digestCount: also count the number of event that has occured in a box for a given user, displayed in digests send to the user out-of-the-app.
// - realtime: send to the active user app through websocket updates.

// for all identities except the event sender
// - increment box count
// - increment digest count
func countActivity(ctx context.Context, e *Event, exec boil.ContextExecutor, redConn *redis.Client, identities *IdentityMapper, _ files.FileStorageRepo, _ Metadata) error {
	// 1. retrieve member ids
	memberIDs, err := ListBoxMemberIDs(ctx, exec, redConn, e.BoxID)
	if err != nil {
		return merror.Transform(err).Describe("notifying member: listing members")
	}

	// get box settings for all users
	filters := BoxSettingFilters{
		BoxIDs: []string{e.BoxID},
	}
	boxSettings, err := ListBoxSettings(ctx, exec, filters)
	if err != nil {
		return err
	}
	isMuted := make(map[string]bool, len(boxSettings))
	for _, boxSetting := range boxSettings {
		isMuted[boxSetting.IdentityID] = boxSetting.Muted
	}

	// delete the notification sender id and the
	// senders who muted the box from the list
	filteredMemberIDs := memberIDs[:0]
	for _, id := range memberIDs {
		muted, ok := isMuted[id]
		if id != e.SenderID && !(ok && muted) {
			filteredMemberIDs = append(filteredMemberIDs, id)
		}
	}

	// incr digest count for a given box for all received identityIDs
	if err := IncrDigestCount(ctx, redConn, filteredMemberIDs, e.BoxID); err != nil {
		return err
	}

	// incr counts for a given box for all received identityIDs
	return IncrBoxCounts(ctx, redConn, filteredMemberIDs, e.BoxID)

}

// invalidates all redis caches for the boxID & event.senderID
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

// send Realtime Update to all members of the box about the given event
func sendRealtimeUpdate(ctx context.Context, e *Event, exec boil.ContextExecutor, redConn *redis.Client, identities *IdentityMapper, _ files.FileStorageRepo, _ Metadata) error {
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
		bu := realtime.Update{
			Type:   "event.new",
			Object: formattedEvent,
		}
		realtime.SendUpdate(ctx, redConn, memberID, &bu)
	}

	return nil
}
