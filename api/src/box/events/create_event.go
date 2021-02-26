package events

import (
	"context"

	"github.com/volatiletech/null/v8"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
	"gitlab.misakey.dev/misakey/backend/api/src/box/events/cache"
	"gitlab.misakey.dev/misakey/backend/api/src/box/events/etype"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"
)

// CreationContent ...
type CreationContent struct {
	OwnerOrgID        string  `json:"owner_org_id"`
	DatatagID         *string `json:"datatag_id,omitempty"`
	SubjectIdentityID *string `json:"subject_identity_id,omitempty"`
	PublicKey         string  `json:"public_key"`
	Title             string  `json:"title"`
}

// Unmarshal ...
func (c *CreationContent) Unmarshal(json types.JSON) error {
	return json.Unmarshal(c)
}

// Validate ...
func (c CreationContent) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.OwnerOrgID, v.Required),
		v.Field(&c.DatatagID, is.UUIDv4),
		v.Field(&c.SubjectIdentityID, is.UUIDv4),
		v.Field(&c.PublicKey, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
		v.Field(&c.Title, v.Required, v.Length(1, 50)),
	)
}

// CreateCreateEvent ...
func CreateCreateEvent(
	ctx context.Context,
	exec boil.ContextExecutor, redConn *redis.Client, identities *IdentityMapper,
	title, publicKey, ownerOrgID string,
	datatagID, subjectIdentityID *string,
	senderID string,
) (Event, error) {
	// generate an id for the created box
	boxID, err := uuid.NewString()
	if err != nil {
		return Event{}, merr.From(err).Desc("generating box ID")
	}

	c := CreationContent{
		OwnerOrgID:        ownerOrgID,
		PublicKey:         publicKey,
		Title:             title,
		DatatagID:         datatagID,
		SubjectIdentityID: subjectIdentityID,
	}
	event, err := newWithAnyContent(etype.Create, &c, boxID, senderID, nil)
	if err != nil {
		return Event{}, merr.From(err).Desc("creating create event")
	}

	// persist the event in storage
	if err = event.persist(ctx, exec); err != nil {
		return Event{}, merr.From(err).Desc("inserting event")
	}

	// clean box cache for the creator
	err = cache.CleanUserBoxByUserOrg(ctx, redConn, senderID, ownerOrgID)
	if err != nil {
		logger.FromCtx(ctx).Warn().Msgf("clean user box cache %s: %v", senderID, err)
	}

	// send notification to creator
	if err := sendRealtimeUpdate(ctx, &event, exec, redConn, identities, nil, nil); err != nil {
		logger.FromCtx(ctx).Warn().Err(err).Msgf("could not send create notification for box %s", event.BoxID)
	}

	return event, nil
}

type createInfo struct {
	OwnerOrgID string
	Pubkey     string
	Title      string
	CreatorID  string
}

// GetCreateInfo retrieves the create event of the box and marshal its content into a CreationContent structure aside its creator id
func GetCreateInfo(
	ctx context.Context,
	exec boil.ContextExecutor,
	boxID string,
) (info createInfo, err error) {
	content := CreationContent{}
	e, err := get(ctx, exec, eventFilters{
		boxID: null.StringFrom(boxID),
		eType: null.StringFrom(etype.Create),
	})
	if err != nil {
		return info, merr.From(err).Desc("getting create event")
	}
	if err := e.JSONContent.Unmarshal(&content); err != nil {
		return info, merr.From(err).Descf("unmarshaling creation event content")
	}
	// fill info
	info.OwnerOrgID = content.OwnerOrgID
	info.Pubkey = content.PublicKey
	info.Title = content.Title
	info.CreatorID = e.SenderID
	return info, nil
}

// ListCreateEventsByCreatorID ...
func ListCreateByCreatorID(ctx context.Context, exec boil.ContextExecutor, creatorID string) ([]Event, error) {
	createEvents, err := list(ctx, exec, eventFilters{
		eType:    null.StringFrom(etype.Create),
		senderID: null.StringFrom(creatorID),
	})
	return createEvents, err
}

// MapCreationContentByBoxID retrieves the create event of the boxes and marshal its content
// it returns a map[boxID]CreationContent
func MapCreationContentByBoxID(
	ctx context.Context,
	exec boil.ContextExecutor,
	boxIDs []string,
) (map[string]CreationContent, error) {
	events, err := list(ctx, exec, eventFilters{
		boxIDs: boxIDs,
		eType:  null.StringFrom(etype.Create),
	})
	if err != nil {
		return nil, merr.From(err).Desc("listing create event")
	}
	contentByBoxID := make(map[string]CreationContent, len(events))
	for _, e := range events {
		c := CreationContent{}
		if err := e.JSONContent.Unmarshal(&c); err != nil {
			return nil, merr.From(err).Descf("unmarshaling creation event content")
		}
		contentByBoxID[e.BoxID] = c
	}
	return contentByBoxID, nil
}

func isCreator(ctx context.Context, exec boil.ContextExecutor, boxID, senderID string) (bool, error) {
	_, err := get(ctx, exec, eventFilters{
		eType:    null.StringFrom(etype.Create),
		senderID: null.StringFrom(senderID),
		boxID:    null.StringFrom(boxID),
	})
	// if no error, the user is the creator
	if err == nil {
		return true, nil
	}

	// not found error means the user is not the creator
	if merr.IsANotFound(err) {
		return false, nil
	}
	// return the error if unexpected
	return false, err
}

func isDataSubject(ctx context.Context, exec boil.ContextExecutor, boxID, senderID string) (bool, error) {
	event, err := get(ctx, exec, eventFilters{
		eType: null.StringFrom(etype.Create),
		boxID: null.StringFrom(boxID),
	})
	if err != nil {
		return false, err
	}

	var content CreationContent
	if err := event.JSONContent.Unmarshal(&content); err != nil {
		return false, err
	}

	if content.SubjectIdentityID != nil && *content.SubjectIdentityID == senderID {
		return true, nil
	}

	return false, nil
}
