package events

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events/etype"

	"github.com/volatiletech/null"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
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
	uniqueMemberIDs[createEvent.SenderID] = true

	// 2. compute people that has joined the box
	activeJoins, err := listBoxActiveJoinEvents(ctx, exec, boxID)
	if err != nil {
		return nil, err
	}
	for _, e := range activeJoins {
		uniqueMemberIDs[e.SenderID] = true
	}

	memberIDs := make([]string, len(uniqueMemberIDs))
	idx := 0
	for memberID := range uniqueMemberIDs {
		memberIDs[idx] = memberID
		idx++
	}
	return memberIDs, nil
}

func MustBeMember(
	ctx context.Context,
	exec boil.ContextExecutor,
	boxID, senderID string,
) error {
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

func isMember(ctx context.Context, exec boil.ContextExecutor, boxID, senderID string) (bool, error) {
	err := MustBeMember(ctx, exec, boxID, senderID)
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
