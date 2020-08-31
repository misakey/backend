package events

import (
	"context"
	"database/sql"
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
		events[i] = FromSQLBoiler(record)
	}

	if len(events) == 0 {
		return events, merror.NotFound().Detail("id", merror.DVNotFound)
	}

	return events, nil
}

func ListByBoxIDAndType(ctx context.Context, exec boil.ContextExecutor, boxID, eventType string) ([]Event, error) {
	mods := []qm.QueryMod{
		sqlboiler.EventWhere.BoxID.EQ(boxID),
		sqlboiler.EventWhere.Type.EQ(eventType),
		qm.OrderBy(sqlboiler.EventColumns.CreatedAt + " DESC"),
	}

	dbEvents, err := sqlboiler.Events(mods...).All(ctx, exec)
	if err != nil {
		return nil, merror.Transform(err).Describe("retrieving db events")
	}

	events := make([]Event, len(dbEvents))
	for i, record := range dbEvents {
		events[i] = FromSQLBoiler(record)
	}

	return events, nil
}

func ListByTypeAndBoxIDAndSenderID(ctx context.Context, exec boil.ContextExecutor, eventType, boxID, senderID string) ([]Event, error) {
	mods := []qm.QueryMod{
		sqlboiler.EventWhere.BoxID.EQ(boxID),
		sqlboiler.EventWhere.Type.EQ(eventType),
		sqlboiler.EventWhere.SenderID.EQ(senderID),
		qm.OrderBy(sqlboiler.EventColumns.CreatedAt + " DESC"),
	}

	dbEvents, err := sqlboiler.Events(mods...).All(ctx, exec)
	if err != nil {
		return nil, merror.Transform(err).Describe("retrieving db events")
	}

	events := make([]Event, len(dbEvents))
	for i, record := range dbEvents {
		events[i] = FromSQLBoiler(record)
	}

	if len(events) == 0 {
		return events, merror.NotFound().Detail("id", merror.DVNotFound)
	}

	return events, nil
}

func newWithAnyContent(eType string, content anyContent, boxID string, senderID string) (Event, error) {
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

func findByTypeContent(ctx context.Context, exec boil.ContextExecutor, boxID, eType string, jsonQuery *string) (Event, error) {
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
	if err == sql.ErrNoRows {
		return e, merror.NotFound().
			Detail("box_id", merror.DVNotFound).
			Detail("type", merror.DVNotFound).
			Describef("finding %s by type %s content", boxID, eType)
	}
	if err != nil {
		return e, merror.Transform(err).Describe("retrieving type/content db event")
	}
	return FromSQLBoiler(dbEvent), nil
}

func FindByEncryptedFileID(ctx context.Context, exec boil.ContextExecutor, encryptedFileID string) ([]Event, error) {
	// build query
	jsonQuery := `{"encrypted_file_id": "` + encryptedFileID + `"}`
	mods := []qm.QueryMod{
		sqlboiler.EventWhere.Type.EQ("msg.file"),
		qm.Where(`content::jsonb @> ?`, jsonQuery),
	}

	dbEvents, err := sqlboiler.Events(mods...).All(ctx, exec)
	if err != nil {
		return nil, err
	}

	events := make([]Event, len(dbEvents))
	for i, record := range dbEvents {
		events[i] = FromSQLBoiler(record)
	}

	if len(events) == 0 {
		return events, merror.NotFound().Detail("id", merror.DVNotFound)
	}

	return events, nil
}

func ListFilesID(ctx context.Context, exec boil.ContextExecutor, boxID string) ([]string, error) {
	events, err := ListByBoxIDAndType(ctx, exec, boxID, "msg.file")
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(events))
	var content MsgFileContent
	for idx, event := range events {
		err = json.Unmarshal(event.Content, &content)
		if err != nil {
			return nil, merror.Internal().Describe("unmarshaling content json")
		}
		ids[idx] = content.EncryptedFileID
	}

	return ids, nil
}
