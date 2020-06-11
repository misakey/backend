package events

import (
	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/boxes/repositories/sqlboiler"
)

func (e *Event) ToSqlBoiler() *sqlboiler.Event {
	result := sqlboiler.Event{
		ID:        e.ID,
		BoxID:     e.BoxID,
		CreatedAt: e.CreatedAt,
		Type:      e.Type,
		Content:   null.JSONFrom(e.Content),
		SenderID:  e.SenderID,
	}

	return &result
}

func FromSqlBoiler(src *sqlboiler.Event) *Event {
	dst := Event{}
	dst.BoxID = src.BoxID
	dst.ID = src.ID
	dst.CreatedAt = src.CreatedAt
	dst.Type = src.Type
	dst.Content = src.Content.JSON

	return &dst
}
