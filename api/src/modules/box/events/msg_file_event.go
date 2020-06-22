package events

import (
	"context"
	"database/sql"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/volatiletech/sqlboiler/types"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type msgFileContent struct {
	Encrypted       string `json:"encrypted"`
	EncryptedFileID string `json:"encrypted_file_id"`
}

func (c *msgFileContent) Unmarshal(json types.JSON) error {
	return json.Unmarshal(c)
}

func (c msgFileContent) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.Encrypted, v.Required, is.Base64),
		v.Field(&c.EncryptedFileID, v.Required, is.UUIDv4),
	)
}

func NewMsgFile(
	ctx context.Context,
	boxID string, senderID string,
	encryptedContent string,
) (Event, string, error) {
	e := Event{}

	// generate a new uuid as a file ID
	fileID, err := uuid.NewString()
	if err != nil {
		return e, "", merror.Transform(err).Describe("file id")
	}

	// build the event content
	content := msgFileContent{
		Encrypted:       encryptedContent,
		EncryptedFileID: fileID,
	}

	e, err = NewWithAnyContent("msg.file", &content, boxID, senderID)
	if err != nil {
		return e, "", merror.Transform(err).Describe("new event")
	}
	return e, fileID, nil
}

func GetMsgFile(
	ctx context.Context,
	db *sql.DB,
	boxID, fileID string,
) (Event, error) {
	jsonQuery := `{"encrypted_file_id": "` + fileID + `"}`
	return FindByTypeContent(ctx, db, boxID, "msg.file", &jsonQuery)
}
