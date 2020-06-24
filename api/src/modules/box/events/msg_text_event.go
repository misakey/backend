package events

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/volatiletech/sqlboiler/types"
)

type msgTextContent struct {
	Encrypted string `json:"encrypted"`
	PublicKey string `json:"public_key"`
}

func (c *msgTextContent) Unmarshal(content types.JSON) error {
	return content.Unmarshal(c)
}

func (c msgTextContent) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.Encrypted, v.Required, is.Base64),
		v.Field(&c.PublicKey, v.Required),
	)
}
