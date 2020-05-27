package events

import (
	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/backend/api/src/sqlboiler"
)

func (e *Event) ToSqlBoiler() *sqlboiler.Event {
	result := sqlboiler.Event{
		ID:        e.ID,
		BoxID:     e.BoxID,
		CreatedAt: e.CreatedAt,
		Type:      e.Type,
		Content:   null.JSONFrom(e.Content),
		// TODO (for now we put a constant because senders are ont implemented)
		SenderID: "c80b6bf4-d021-42d5-bd06-2769fa7a81b5",
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
