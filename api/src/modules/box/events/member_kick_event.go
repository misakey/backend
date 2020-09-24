package events

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events/etype"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
)

type MemberKickContent struct {
	// Stored but not in json
	KickedMemberID string `json:"kicked_member_id,omitempty"`
	// in json
	KickedMember *SenderView `json:"kicked_member,omitempty"`
}

func (c *MemberKickContent) Unmarshal(content types.JSON) error {
	return content.Unmarshal(c)
}

func (c MemberKickContent) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.KickedMemberID, v.Required, is.UUIDv4),
		v.Field(&c.KickedMember, v.Nil),
	)
}

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
				content := MemberKickContent{KickedMemberID: joinEvent.SenderID}
				kickEvent, err := newWithAnyContent(etype.Memberkick, &content, boxID, kickerID, &joinEvent.ID)
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
