package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/volatiletech/sqlboiler/boil"
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

func NewWithAnyContent(eType string, content anyContent, boxID string, senderID string) (Event, error) {
	contentBytes, err := json.Marshal(content)
	if err != nil {
		return Event{}, merror.Transform(err).Describe("marshalling anyContent into bytes")
	}
	jsonContent := types.JSON{}
	if err := jsonContent.UnmarshalJSON(contentBytes); err != nil {
		return Event{}, merror.Transform(err).Describe("unmarshalling content bytes into types.JSON")
	}

	return New(eType, jsonContent, boxID, senderID)
}

func ListByBoxID(ctx context.Context, exec boil.ContextExecutor, boxID string) ([]Event, error) {
	mods := []qm.QueryMod{
		sqlboiler.EventWhere.BoxID.EQ(boxID),
		qm.OrderBy(sqlboiler.EventColumns.CreatedAt + " DESC"),
	}

	dbEvents, err := sqlboiler.Events(mods...).All(ctx, exec)
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

func FindByTypeContent(ctx context.Context, exec boil.ContextExecutor, boxID, eType string, jsonQuery *string) (Event, error) {
	var e Event

	// build query
	mods := []qm.QueryMod{
		sqlboiler.EventWhere.BoxID.EQ(boxID),
		sqlboiler.EventWhere.Type.EQ(eType),
		qm.OrderBy(sqlboiler.EventColumns.CreatedAt + " DESC"),
	}
	// add content query if existing
	if jsonQuery != nil {
		mods = append(mods, qm.Where(`content::jsonb @> ?`, *jsonQuery))
	}

	dbEvent, err := sqlboiler.Events(mods...).One(ctx, exec)
	if err != nil {
		return e, merror.Transform(err).Describe("retrieving type/content db event")
	}
	return FromSqlBoiler(dbEvent), nil
}
