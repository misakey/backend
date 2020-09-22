package events

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events/etype"

	"github.com/volatiletech/null"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
	"gitlab.misakey.dev/misakey/msk-sdk-go/logger"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

// List identities ID that are members of the given box
func ListBoxMemberIDs(ctx context.Context, exec boil.ContextExecutor, boxID string) ([]string, error) {
	// 1. get the creator id which is a member
	createEvent, err := get(ctx, exec, eventFilters{
		eType: null.StringFrom(etype.Create),
		boxID: null.StringFrom(boxID),
	})
	if err != nil {
		return nil, merror.Transform(err).Describe("getting create event")
	}
	// start the member IDs list with it
	uniqueMemberIDs := make(map[string]bool)
	memberIDs := []string{createEvent.SenderID}
	uniqueMemberIDs[createEvent.SenderID] = true

	// 2. compute people that has joined the box
	activeJoins, err := listBoxActiveJoinEvents(ctx, exec, boxID)
	if err != nil {
		return nil, err
	}

	// build the list and ensure unicity with a map
	for _, e := range activeJoins {
		_, ok := uniqueMemberIDs[e.SenderID]
		if !ok {
			uniqueMemberIDs[e.SenderID] = true
			memberIDs = append(memberIDs, e.SenderID)
		}
	}

	return memberIDs, nil
}

func MustBeMember(
	ctx context.Context,
	exec boil.ContextExecutor,
	redConn *redis.Client,
	boxID, senderID string,
) error {

	// check the membership in the cache if it exists
	exists, err := redConn.Exists(boxID + ":members").Result()
	if err == nil && exists == 1 {
		// if cache is valid
		senderIsMember, err := redConn.SIsMember(boxID+":members", senderID).Result()
		if err != nil {
			if senderIsMember {
				return nil
			}
			return merror.Forbidden().Describe("restricted to member").Detail("sender_id", merror.DVForbidden)
		}
	}

	// if the creator, returns immediatly
	isCreator, err := isCreator(ctx, exec, boxID, senderID)
	if err != nil {
		return err
	}
	if isCreator {
		return nil
	}

	_, err = get(ctx, exec, eventFilters{
		eType:      null.StringFrom("member.join"),
		unreferred: true,
		boxID:      null.StringFrom(boxID),
		// NOTE: today senderID is not used to build unreferred filter since boxID is considered before
		// this is necessary since the sender of member.kick is not the sender of the member.join event.
		senderID: null.StringFrom(senderID),
	})
	// if found, the sender is a member of the box
	if err == nil {
		return nil
	}
	return merror.Forbidden().Describe("restricted to member").Detail("sender_id", merror.DVForbidden)
}

func isMember(
	ctx context.Context,
	exec boil.ContextExecutor,
	redConn *redis.Client,
	boxID,
	senderID string,
) (bool, error) {
	err := MustBeMember(ctx, exec, redConn, boxID, senderID)
	if err != nil && merror.HasCode(err, merror.ForbiddenCode) {
		return false, nil
	}
	// return false admin if an error has occured
	return (err == nil), err
}

// increment count for all identities except the sender
func notify(ctx context.Context, e *Event, exec boil.ContextExecutor, redConn *redis.Client, identities entrypoints.IdentityIntraprocessInterface) error {
	// retrieve all member ids excepted the sender id
	// we build a set to find all uniq actors
	memberIDs, err := ListBoxMemberIDs(ctx, exec, e.BoxID)
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
	return incrCounts(ctx, redConn, memberIDs, e.BoxID)
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
		sendInterruption(ctx, redConn, e.SenderID, e.BoxID)
		// on a kicked event
		// close sockets for the kicked member
	} else if e.Type == etype.Memberkick {
		var c MemberKickContent
		if err := e.JSONContent.Unmarshal(&c); err != nil {
			return merror.Transform(err).Describe("marshalling lifecycle content")
		}
		sendInterruption(ctx, redConn, c.KickedMemberID, e.BoxID)
		// on a close event
		// close sockets for all members except the admin
	} else if e.Type == etype.Statelifecycle {
		var c StateLifecycleContent
		if err := e.JSONContent.Unmarshal(&c); err != nil {
			return merror.Transform(err).Describe("marshalling lifecycle content")
		}
		if c.State == "closed" {
			// get all members
			memberIDs, err := ListBoxMemberIDs(ctx, exec, e.BoxID)
			if err != nil {
				return merror.Transform(err).Describe("getting members list")
			}
			// get admin
			adminID, err := getAdminID(ctx, exec, e.BoxID)
			if err != nil {
				return merror.Transform(err).Describe("getting admin id")
			}
			// send interruption if not creator
			for _, memberID := range memberIDs {
				if memberID != adminID {
					sendInterruption(ctx, redConn, memberID, e.BoxID)
				}
			}
		}
	}
	return nil
}

func sendInterruption(ctx context.Context, redConn *redis.Client, senderID, boxID string) {
	logger.
		FromCtx(ctx).
		Debug().
		Msgf("sending interruption message to %s:%s", boxID, senderID)
	if _, err := redConn.Publish("interrupt:"+boxID+":"+senderID, []byte("stop")).Result(); err != nil {
		logger.FromCtx(ctx).Error().Err(err).Msgf("interrupting channel interrupt:%s:%s", boxID, senderID)
	}
}
