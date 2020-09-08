package events

import (
	"context"
	"github.com/volatiletech/null"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/types"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type MsgFileContent struct {
	Encrypted       string `json:"encrypted"`
	PublicKey       string `json:"public_key"`
	EncryptedFileID string `json:"encrypted_file_id"`
}

func (c *MsgFileContent) Unmarshal(json types.JSON) error {
	return json.Unmarshal(c)
}

func (c MsgFileContent) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.Encrypted, v.Required, is.Base64),
		v.Field(&c.PublicKey, v.Required),
		v.Field(&c.EncryptedFileID, v.Required, is.UUIDv4),
	)
}

func NewMsgFile(
	ctx context.Context,
	boxID string, senderID string,
	encContent string, pubKey string,
) (Event, string, error) {
	e := Event{}

	// generate a new uuid as a file ID
	fileID, err := uuid.NewString()
	if err != nil {
		return e, "", merror.Transform(err).Describe("file id")
	}

	// build the event content
	content := MsgFileContent{
		Encrypted:       encContent,
		PublicKey:       pubKey,
		EncryptedFileID: fileID,
	}

	e, err = newWithAnyContent("msg.file", &content, boxID, senderID)
	if err != nil {
		return e, "", merror.Transform(err).Describe("new event")
	}
	return e, fileID, nil
}

func GetMsgFile(
	ctx context.Context,
	exec boil.ContextExecutor,
	boxID, fileID string,
) (Event, error) {
	jsonQuery := `{"encrypted_file_id": "` + fileID + `"}`
	return get(ctx, exec, eventFilters{
		boxID:   null.StringFrom(boxID),
		eType:   null.StringFrom("msg.file"),
		content: &jsonQuery,
	})
}
