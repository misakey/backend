package events

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/volatiletech/sqlboiler/types"
)

// MsgEditContent is exported
// because application layer need to access the "EventID" field
type MsgEditContent struct {
	NewEncrypted string `json:"new_encrypted"`
	NewPublicKey string `json:"new_public_key"`
}

func (c *MsgEditContent) Unmarshal(content types.JSON) error {
	return content.Unmarshal(c)
}

func (c MsgEditContent) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.NewEncrypted, v.Required, is.Base64),
		v.Field(&c.NewEncrypted, v.Required), // URL-safe base64
	)
}
