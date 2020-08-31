package events

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/volatiletech/sqlboiler/types"
)

// MsgDeleteContent is exported
// because application layer need to access the "EventID" field
type MsgDeleteContent struct {
	EventID string `json:"event_id"`
}

func (c *MsgDeleteContent) Unmarshal(content types.JSON) error {
	return content.Unmarshal(c)
}

func (c MsgDeleteContent) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.EventID, v.Required, is.UUIDv4),
	)
}
