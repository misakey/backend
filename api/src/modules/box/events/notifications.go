package events

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events/cache"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events/etype"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/notifications"
)

// increment count for all identities except the sender
// and send event to all members
func notify(ctx context.Context, e *Event, exec boil.ContextExecutor, redConn *redis.Client, identities entrypoints.IdentityIntraprocessInterface) error {
	// 1. retrieve member ids
	memberIDs, err := ListBoxMemberIDs(ctx, exec, redConn, e.BoxID)
	if err != nil {
		return merror.Transform(err).Describe("notifying member: listing members")
	}

	// 2. increase events count
	// delete the notification sender id from the list
	for i, id := range memberIDs {
		if id == e.SenderID {
			memberIDs = append(memberIDs[:i], memberIDs[i+1:]...)
			break
		}
	}

	// incr counts for a given box for all received identityIDs
	if err := IncrCounts(ctx, redConn, memberIDs, e.BoxID); err != nil {
		return merror.Transform(err).Describe("increasing events count")
	}

	// 3. send updates
	// build a set to ensure unicity
	uniqRecipientsIDs := make(map[string]bool)
	for _, memberID := range memberIDs {
		uniqRecipientsIDs[memberID] = true
	}

	// add sender_id and kicked_member_id
	uniqRecipientsIDs[e.SenderID] = true
	if e.Type == "member.kick" {
		var content MemberKickContent
		if err := e.JSONContent.Unmarshal(&content); err != nil {
			return err
		}
		uniqRecipientsIDs[content.KickedMemberID] = true
	}

	box, err := Compute(ctx, e.BoxID, exec, identities, e)
	if err != nil {
		return err
	}

	for memberID := range uniqRecipientsIDs {
		box.EventsCount = ComputeCount(ctx, redConn, memberID, e.BoxID)
		bu := notifications.Update{
			Type:   "box",
			Object: box,
		}
		notifications.SendBoxUpdate(ctx, redConn, memberID, &bu)
	}

	return nil

}

// send event to realtime channels
func publish(ctx context.Context, e *Event, exec boil.ContextExecutor, redConn *redis.Client, identities entrypoints.IdentityIntraprocessInterface) error {
	senderIdentities, err := MapSenderIdentities(ctx, []Event{*e}, identities)
	if err != nil {
		return merror.Transform(err).Describe("getting sender information")
	}

	view, err := FormatEvent(*e, senderIdentities)
	if err != nil {
		return merror.Transform(err).Describe("formatting event")
	}

	serializedEvent, err := view.ToJSON()
	if err != nil {
		return merror.Internal().Describe("encoding event to json")
	}
	redConn.Publish(e.BoxID+":events", serializedEvent)

	return nil
}

// send interrupt messages to close realtime channels
func interrupt(ctx context.Context, e *Event, exec boil.ContextExecutor, redConn *redis.Client, identities entrypoints.IdentityIntraprocessInterface) error {
	// on a leave event
	// close sockets for the leaving member
	if e.Type == etype.Memberleave {
		notifications.SendInterruption(ctx, redConn, e.SenderID, e.BoxID)
		// on a kicked event
		// close sockets for the kicked member
	} else if e.Type == etype.Memberkick {
		var c MemberKickContent
		if err := e.JSONContent.Unmarshal(&c); err != nil {
			return merror.Transform(err).Describe("marshalling lifecycle content")
		}
		notifications.SendInterruption(ctx, redConn, c.KickedMemberID, e.BoxID)
		// on a close event
		// close sockets for all members except the admin
	} else if e.Type == etype.Statelifecycle {
		var c StateLifecycleContent
		if err := e.JSONContent.Unmarshal(&c); err != nil {
			return merror.Transform(err).Describe("marshalling lifecycle content")
		}
		if c.State == "closed" {
			// get all members
			memberIDs, err := ListBoxMemberIDs(ctx, exec, redConn, e.BoxID)
			if err != nil {
				return merror.Transform(err).Describe("getting members list")
			}
			// get admin
			adminID, err := GetAdminID(ctx, exec, e.BoxID)
			if err != nil {
				return merror.Transform(err).Describe("getting admin id")
			}
			// send interruption if not creator
			for _, memberID := range memberIDs {
				if memberID != adminID {
					notifications.SendInterruption(ctx, redConn, memberID, e.BoxID)
				}
			}
		}
	}
	return nil
}

func invalidateCaches(ctx context.Context, e *Event, exec boil.ContextExecutor, redConn *redis.Client, identities entrypoints.IdentityIntraprocessInterface) error {
	_, err := redConn.Del(cache.GetBoxMembersKey(e.BoxID)).Result()
	if err != nil {
		logger.FromCtx(ctx).Warn().Msgf("could not invalidate cache %s:members", e.BoxID)
	}

	var senderID string
	if e.Type == etype.Memberkick {
		var content MemberKickContent
		if err := e.JSONContent.Unmarshal(&content); err != nil {
			logger.FromCtx(ctx).Warn().Msgf("could not unmarshall member.kick %s content", e.ID)
		}
		senderID = content.KickedMemberID
	} else {
		senderID = e.SenderID
	}
	_, err = redConn.Del(cache.GetSenderBoxesKey(senderID)).Result()
	if err != nil {
		logger.FromCtx(ctx).Warn().Msgf("could not invalidate cache %s:boxes", senderID)
	}

	return nil
}

type DeletedBox struct {
	BoxID    string `json:"id"`
	SenderID string `json:"sender_id"`
}

func SendDeleteBox(ctx context.Context, redConn *redis.Client, boxID, senderID string, memberIDs []string) {
	deletedBox := DeletedBox{
		BoxID:    boxID,
		SenderID: senderID,
	}
	bu := notifications.Update{
		Type:   "box.delete",
		Object: deletedBox,
	}
	for _, memberID := range memberIDs {
		notifications.SendBoxUpdate(ctx, redConn, memberID, &bu)
	}
}
