package events

import (
	"bytes"

	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
)

func (e *Event) ToSqlBoiler() *sqlboiler.Event {
	result := sqlboiler.Event{
		BoxID: e.BoxID,

		ID:        e.ID,
		SenderID:  e.SenderID,
		CreatedAt: e.CreatedAt,

		Type:    e.Type,
		Content: null.JSONFrom(bytes.TrimSpace([]byte(e.Content))),
	}
	return &result
}

func FromSqlBoiler(src *sqlboiler.Event) Event {
	dst := Event{
		ID:        src.ID,
		CreatedAt: src.CreatedAt,
		SenderID:  src.SenderID,
		Type:      src.Type,
		Content:   src.Content.JSON,
		BoxID:     src.BoxID,
	}
	return dst
}
