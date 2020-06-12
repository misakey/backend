package events

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/volatiletech/sqlboiler/types"
)

type messageContent struct {
	Encrypted string `json:"encrypted"`
}

func (c *messageContent) Unmarshal(content types.JSON) error {
	return content.Unmarshal(c)
}

func (c messageContent) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.Encrypted, v.Required, is.Base64, v.Length(1, 1024)),
	)
}
