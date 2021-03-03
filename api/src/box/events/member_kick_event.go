package events

import (
	"context"
	"encoding/json"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events/etype"
	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
)

// MemberKickContent ...
type MemberKickContent struct {
	// Stored but not return in json view
	KickerID string `json:"kicker_id,omitempty"`
	// in json
	Kicker *SenderView `json:"kicker,omitempty"`
}

// Unmarshal ...
func (c *MemberKickContent) Unmarshal(content types.JSON) error {
	return content.Unmarshal(c)
}

// Validate ...
func (c MemberKickContent) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.KickerID, v.Required, is.UUIDv4),
		v.Field(&c.Kicker, v.Nil),
	)
}

// KickDeprecatedMembers by checking if the member is still in the identifier accesses list.
func KickDeprecatedMembers(
	ctx context.Context,
	exec boil.ContextExecutor, identities *IdentityMapper,
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
		if err := MustBeLegitimate(ctx, exec, identities, boxID, joinEvent.SenderID); err != nil {
			// if the member has no access anymore then kick them by creation a member.kick event
			if merr.IsAForbidden(err) {
				content := MemberKickContent{KickerID: kickerID}
				kickEvent, err := newWithAnyContent(etype.Memberkick, &content, boxID, joinEvent.SenderID, &joinEvent.ID)
				if err != nil {
					return kicks, merr.From(err).Desc("newing kick event")
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

// notifyKick by creating identity notifications for kicked members
func notifyKick(ctx context.Context, e *Event, exec boil.ContextExecutor, _ *redis.Client, identities *IdentityMapper, _ files.FileStorageRepo, _ Metadata) error {
	// TODO (perf): use metadahandlers to bear already retrieved data across handlers ?
	box, err := Compute(ctx, e.BoxID, exec, identities, e)
	if err != nil {
		return merr.From(err).Desc("computing box")
	}

	// notify the kicked identity they have been kicked.
	kickDetails := struct {
		BoxID      string `json:"id"`
		BoxTitle   string `json:"title"`
		OwnerOrgID string `json:"owner_org_id"`
	}{
		BoxID:      box.ID,
		BoxTitle:   box.Title,
		OwnerOrgID: box.OwnerOrgID,
	}
	bytes, err := json.Marshal(kickDetails)
	if err != nil {
		return merr.From(err).Desc("marshalling kick details")
	}
	identities.CreateNotifs(ctx, []string{e.SenderID}, etype.Memberkick, null.JSONFrom(bytes))
	return nil
}
