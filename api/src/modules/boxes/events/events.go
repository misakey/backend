package events

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/volatiletech/sqlboiler/types"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/boxes/utils"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type UserSetFields struct {
	Type    string     `json:"type"`
	Content types.JSON `json:"content"`
}

type readOnlyFields struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"server_event_created_at"`
}

type Event struct {
	UserSetFields
	readOnlyFields
	BoxID string `json:"-"`
}

type messageContent struct {
	Encrypted string `json:"encrypted"`
}

type lifecycleContent struct {
	State string `json:"state"`
}

func New(fields UserSetFields, boxID string) (*Event, error) {
	err := validation.ValidateStruct(&fields,
		validation.Field(&fields.Type, validation.Required, validation.In("msg.text", "msg.file", "state.lifecycle")),
	)
	if err != nil {
		return nil, err
	}

	event := Event{
		UserSetFields: fields,
		BoxID:         boxID,
	}

	// Validating the shape of the event
	// (no business logic validation here)
	if event.Type == "msg.text" || event.Type == "msg.file" {
		content := messageContent{}
		err = event.Content.Unmarshal(&content)
		if err != nil {
			return nil, merror.Transform(err).Code(merror.BadRequestCode).Detail("content", "invalid")
		}
		err = validation.ValidateStruct(&content,
			validation.Field(&content.Encrypted, validation.Required, is.Base64, validation.Length(1, 1024)),
		)
		if err != nil {
			return nil, merror.Transform(err).Code(merror.BadRequestCode).Detail("encrypted", "invalid")
		}
	} else {
		switch event.Type {
		case "state.lifecycle":
			content := lifecycleContent{}
			err := event.Content.Unmarshal(&content)
			if err != nil {
				return nil, merror.Transform(err).Code(merror.BadRequestCode).Detail("content", "invalid")
			}
			err = validation.ValidateStruct(&content,
				validation.Field(&content.State, validation.Required, validation.In("closed")),
			)
			if err != nil {
				return nil, merror.Transform(err).Code(merror.BadRequestCode).Detail("state", "invalid")
			}
		default:
			// This is an internal error and not a bad request error
			// because "type" field should have already been checked
			return nil, merror.Internal().Describef(`no validation function for event type "%s"`, event.Type)
		}
	}

	event.ID, err = utils.RandomUUIDString()
	if err != nil {
		return nil, merror.Transform(err).Describe("could not generate id for event")
	}

	event.CreatedAt = time.Now()

	return &event, nil
}
