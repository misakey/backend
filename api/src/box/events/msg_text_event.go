package events

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type MsgTextContent struct {
	Encrypted    string    `json:"encrypted"`
	PublicKey    string    `json:"public_key"`
	LastEditedAt null.Time `json:"last_edited_at"`
}

func (c *MsgTextContent) Unmarshal(content types.JSON) error {
	return content.Unmarshal(c)
}

func (c MsgTextContent) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.Encrypted, v.Required, is.Base64),
		v.Field(&c.PublicKey, v.Required),
	)
}
