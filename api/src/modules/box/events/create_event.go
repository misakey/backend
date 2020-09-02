package events

import (
	"context"
	"database/sql"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/types"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"
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
	return newWithAnyContent("create", &c, boxID, senderID)
}

func GetCreateEvent(
	ctx context.Context,
	exec boil.ContextExecutor,
	boxID string,
) (Event, error) {
	return findByTypeContent(ctx, exec, boxID, "create", nil)
}

func CreateCreateEvent(ctx context.Context, title, publicKey, senderID string, exec boil.ContextExecutor) (Event, error) {
	event, err := NewCreate(title, publicKey, senderID)
	if err != nil {
		return Event{}, merror.Transform(err).Describe("creating create event")
	}

	// persist the event in storage
	if err = event.ToSQLBoiler().Insert(ctx, exec, boil.Infer()); err != nil {
		return Event{}, merror.Transform(err).Describe("inserting event")
	}

	return event, nil

}

func MustBoxExists(ctx context.Context, exec boil.ContextExecutor, boxID string) error {
	_, err := sqlboiler.Events(
		sqlboiler.EventWhere.BoxID.EQ(boxID),
		sqlboiler.EventWhere.Type.EQ("create"),
	).One(ctx, exec)
	if err != nil && err == sql.ErrNoRows {
		return merror.NotFound().Detail("box_id", merror.DVNotFound)
	}
	return err
}
