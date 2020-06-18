package events

import (
	"context"
	"database/sql"
	"time"

	"github.com/volatiletech/sqlboiler/queries/qm"
	"github.com/volatiletech/sqlboiler/types"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"
)

type Event struct {
	ID        string
	CreatedAt time.Time
	SenderID  string
	Type      string
	Content   types.JSON
	BoxID     string
}

func New(eType string, jsonContent types.JSON, boxID string, senderID string) (Event, error) {
	event := Event{
		CreatedAt: time.Now(),
		SenderID:  senderID,
		Type:      eType,
		BoxID:     boxID,
		Content:   jsonContent,
	}
	// validate the shape of the event content
	err := validateContent(event)
	if err != nil {
		return event, merror.Transform(err).Describe("validating content")
	}

	event.ID, err = uuid.NewString()
	if err != nil {
		return event, merror.Transform(err).Describe("generating event id")
	}
	return event, nil
}

func ListByBoxID(ctx context.Context, db *sql.DB, boxID string) ([]Event, error) {
	mods := []qm.QueryMod{
		sqlboiler.EventWhere.BoxID.EQ(boxID),
		qm.OrderBy(sqlboiler.EventColumns.CreatedAt + " DESC"),
	}

	dbEvents, err := sqlboiler.Events(mods...).All(ctx, db)
	if err != nil {
		return nil, merror.Transform(err).Describe("retrieving db events")
	}

	events := make([]Event, len(dbEvents))
	for i, record := range dbEvents {
		events[i] = FromSqlBoiler(record)
	}

	if len(events) == 0 {
		return events, merror.NotFound().Detail("id", merror.DVNotFound)
	}

	return events, nil
}
