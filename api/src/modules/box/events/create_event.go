package events

import (
	"context"

	"github.com/volatiletech/null/v8"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
)

type CreationContent struct {
	PublicKey string `json:"public_key"`
	Title     string `json:"title"`
}

func (c *CreationContent) Unmarshal(json types.JSON) error {
	return json.Unmarshal(c)
}

func (c CreationContent) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.PublicKey, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
		v.Field(&c.Title, v.Required, v.Length(1, 50)),
	)
}

func NewCreate(title, publicKey, senderID string) (e Event, err error) {
	// generate an id for the created box
	boxID, err := uuid.NewString()
	if err != nil {
		err = merror.Transform(err).Describe("generating box ID")
		return
	}

	c := CreationContent{
		PublicKey: publicKey,
		Title:     title,
	}
	return newWithAnyContent("create", &c, boxID, senderID, nil)
}

func GetCreateEvent(
	ctx context.Context,
	exec boil.ContextExecutor,
	boxID string,
) (Event, error) {
	return get(ctx, exec, eventFilters{
		boxID: null.StringFrom(boxID),
		eType: null.StringFrom("create"),
	})
}

func CreateCreateEvent(
	ctx context.Context,
	title, publicKey, senderID string,
	exec boil.ContextExecutor, redConn *redis.Client, identities *IdentityMapper, filesRepo files.FileStorageRepo,
) (Event, error) {
	event, err := NewCreate(title, publicKey, senderID)
	if err != nil {
		return Event{}, merror.Transform(err).Describe("creating create event")
	}

	// persist the event in storage
	if err = event.persist(ctx, exec); err != nil {
		return Event{}, merror.Transform(err).Describe("inserting event")
	}

	// invalidates cache for creator boxes list
	if err := invalidateCaches(ctx, &event, exec, redConn, identities, filesRepo, nil); err != nil {
		logger.FromCtx(ctx).Warn().Err(err).Msgf("invalidating the cache")
	}

	// send notification to creator
	if err := sendRealtimeUpdate(ctx, &event, exec, redConn, identities, filesRepo, nil); err != nil {
		logger.FromCtx(ctx).Warn().Err(err).Msgf("could not send create notification for box %s", event.BoxID)
	}

	return event, nil
}

func ListCreatorIDEvents(ctx context.Context, exec boil.ContextExecutor, creatorID string) ([]Event, error) {
	createEvents, err := list(ctx, exec, eventFilters{
		eType:    null.StringFrom("create"),
		senderID: null.StringFrom(creatorID),
	})
	return createEvents, err
}

func isCreator(ctx context.Context, exec boil.ContextExecutor, boxID, senderID string) (bool, error) {
	_, err := get(ctx, exec, eventFilters{
		eType:    null.StringFrom("create"),
		senderID: null.StringFrom(senderID),
		boxID:    null.StringFrom(boxID),
	})
	// if no error, the user is the creator
	if err == nil {
		return true, nil
	}

	// not found error means the user is not the creator
	if merror.HasCode(err, merror.NotFoundCode) {
		return false, nil
	}
	// return the error if unexpected
	return false, err
}

func getBoxCreatorID(ctx context.Context, exec boil.ContextExecutor, boxID string) (string, error) {
	e, err := get(ctx, exec, eventFilters{
		eType: null.StringFrom("create"),
		boxID: null.StringFrom(boxID),
	})

	if err != nil {
		return "", err
	}

	return e.SenderID, nil
}
