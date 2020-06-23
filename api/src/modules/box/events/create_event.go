package events

import (
	"context"
	"regexp"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/volatiletech/sqlboiler/types"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

var rxUnpaddedURLsafeBase64 = regexp.MustCompile("^[a-zA-Z0-9_-]+$")

type creationContent struct {
	stateLifecycleContent
	PublicKey string `json:"public_key"`
	Title     string `json:"title"`
}

func (c *creationContent) Unmarshal(json types.JSON) error {
	return json.Unmarshal(c)
}

func (c creationContent) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.State, v.Required, v.In("open")),
		v.Field(&c.PublicKey, v.Required, v.Match(rxUnpaddedURLsafeBase64)),
		v.Field(&c.Title, v.Required, v.Length(5, 50)),
	)
}

func NewCreate(title, publicKey, senderID string) (e Event, err error) {
	// generate an id for the created box
	boxID, err := uuid.NewString()
	if err != nil {
		err = merror.Transform(err).Describe("generating box ID")
		return
	}

	c := creationContent{
		PublicKey:             publicKey,
		Title:                 title,
		stateLifecycleContent: stateLifecycleContent{State: "open"},
	}
	return NewWithAnyContent("create", &c, boxID, senderID)
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
	identity, err := c.repo.Identities().Get(ctx, e.SenderID)
	if err != nil {
		return merror.Transform(err).Describe("retrieving creator")
	}
	c.box.Creator = NewSenderView(identity)
	return nil
}
