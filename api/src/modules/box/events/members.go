package events

import (
	"context"

	"github.com/volatiletech/null"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/boil"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

// List identities ID that are members of the given box
func ListBoxMemberIDs(ctx context.Context, exec boil.ContextExecutor, boxID string) ([]string, error) {
	// 1. get the creator id which is a member
	createEvent, err := get(ctx, exec, eventFilters{
		eType: null.StringFrom("create"),
		boxID: null.StringFrom(boxID),
	})
	if err != nil {
		return nil, merror.Transform(err).Describe("getting create event")
	}
	// start the member IDs list with it
	uniqueMemberIDs := make(map[string]bool)
	uniqueMemberIDs[createEvent.SenderID] = true

	// 2. compute people having access to the box
	// get all the identity that has joined the box and did not leave it
	joinEvents, err := list(ctx, exec, eventFilters{
		eType:     null.StringFrom("member.join"),
		unrefered: true,
		boxID:     null.StringFrom(boxID),
	})
	if err != nil {
		return nil, merror.Transform(err).Describe("listing join events")
	}
	for _, e := range joinEvents {
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
		eType:     null.StringFrom("member.join"),
		unrefered: true,
		senderID:  null.StringFrom(senderID),
		boxID:     null.StringFrom(boxID),
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

func NotifyMembers(ctx context.Context, exec boil.ContextExecutor, redConn *redis.Client, senderID, boxID string) error {
	// retrieve all member ids excepted the sender id
	// we build a set to find all uniq actors
	uniqMembers := make(map[string]bool)
	events, err := list(ctx, exec, eventFilters{
		boxID:       null.StringFrom(boxID),
		notSenderID: null.StringFrom(senderID),
	})
	if err != nil {
		return err
	}
	for _, event := range events {
		uniqMembers[event.SenderID] = true
	}
	memberIDs := make([]string, len(uniqMembers))
	idx := 0
	for member := range uniqMembers {
		memberIDs[idx] = member
		idx += 1
	}

	// incr counts for a given box for all received identityIDs
	return incrCounts(ctx, redConn, memberIDs, boxID)
}
