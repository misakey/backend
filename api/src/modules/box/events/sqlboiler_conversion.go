package events

import (
	"bytes"

	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
)

func (e *Event) ToSQLBoiler() *sqlboiler.Event {
	result := sqlboiler.Event{
		BoxID: e.BoxID,

		ID:        e.ID,
		SenderID:  e.SenderID,
		CreatedAt: e.CreatedAt,

		Type:    e.Type,
		Content: null.JSONFrom(bytes.TrimSpace([]byte(e.JSONContent))),

		ReferrerID: e.ReferrerID,
	}
	return &result
}

func FromSQLBoiler(src *sqlboiler.Event) Event {
	dst := Event{
		ID:          src.ID,
		CreatedAt:   src.CreatedAt,
		SenderID:    src.SenderID,
		Type:        src.Type,
		JSONContent: src.Content.JSON,
		BoxID:       src.BoxID,
		ReferrerID:   src.ReferrerID,
	}
	return dst
}
