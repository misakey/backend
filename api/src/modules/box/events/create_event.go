package events

import (
	"context"
	"encoding/json"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/volatiletech/sqlboiler/types"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/utils"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type creationContent struct {
	stateLifecycleContent
	PublicKey string `json:"public_key"`
	Title     string `json:"title"`
}

func NewCreationJSON(publicKey string, title string) (types.JSON, error) {
	c := creationContent{
		PublicKey:             publicKey,
		Title:                 title,
		stateLifecycleContent: stateLifecycleContent{State: "open"},
	}
	ret, err := json.Marshal(c)
	if err != nil {
		return ret, merror.Transform(err).Describe("marshalling creation content")
	}
	return ret, nil
}

func (c *creationContent) Unmarshal(content types.JSON) error {
	return content.Unmarshal(c)
}

func (c creationContent) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.State, v.Required, v.In("open")),
		v.Field(&c.PublicKey, v.Required, v.Match(utils.RxUnpaddedURLsafeBase64)),
		v.Field(&c.Title, v.Required, v.Length(5, 50)),
	)
}

func (c *boxComputer) playCreate(ctx context.Context, e Event) error {
	c.box.CreatedAt = e.CreatedAt
	creationContent := creationContent{}
	if err := creationContent.Unmarshal(e.Content); err != nil {
		return err
	}
	c.box.Lifecycle = creationContent.stateLifecycleContent.State
	c.box.PublicKey = creationContent.PublicKey
	c.box.Title = creationContent.Title

	// set the creator information
	identity, err := c.identityRepo.GetIdentity(ctx, e.SenderID)
	if err != nil {
		return merror.Transform(err).Describe("retrieving creator")
	}
	c.box.Creator = NewSenderView(identity)
	return nil
}
