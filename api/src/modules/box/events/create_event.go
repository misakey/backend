package events

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/types"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
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
