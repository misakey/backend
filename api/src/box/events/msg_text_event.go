package events

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
)

// MsgTextContent ...
type MsgTextContent struct {
	Encrypted    string    `json:"encrypted"`
	PublicKey    string    `json:"public_key"`
	LastEditedAt null.Time `json:"last_edited_at"`
}

// Unmarshal ...
func (c *MsgTextContent) Unmarshal(content types.JSON) error {
	return content.Unmarshal(c)
}

// Validate ...
func (c MsgTextContent) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.Encrypted, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
		v.Field(&c.PublicKey, v.Required),
	)
}
