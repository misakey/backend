package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"github.com/volatiletech/sqlboiler/types"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/slice"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"
)

type Event struct {
	ID          string
	CreatedAt   time.Time
	BoxID       string
	SenderID    string
	Type        string
	ReferrerID  null.String
	JSONContent types.JSON

	Content interface{}
}

func New(eType string, jsonContent types.JSON, boxID, senderID string, referrerID *string) (Event, error) {
	event := Event{
		CreatedAt:   time.Now(),
		SenderID:    senderID,
		Type:        eType,
		BoxID:       boxID,
		JSONContent: jsonContent,
		ReferrerID:  null.StringFromPtr(referrerID),
	}

	// bind/validate the shape of the event content
	err := bindAndValidateContent(&event)
	if err != nil {
		return event, merror.Transform(err).Describe("binding content")
	}

	event.ID, err = uuid.NewString()
	if err != nil {
		return event, merror.Transform(err).Describe("generating event id")
	}
	return event, nil
}

func ListForMembersByBoxID(ctx context.Context, exec boil.ContextExecutor, boxID string, offset, limit *int) ([]Event, error) {
	return list(ctx, exec, eventFilters{
		boxID:  null.StringFrom(boxID),
		eTypes: []string{Ecreate, Estatelifecycle, Emsgtext, Emsgfile, Emsgedit, Emsgdelete, Ememberjoin, Ememberleave},
		offset: offset,
		limit:  limit,
	})
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

func newWithAnyContent(eType string, content anyContent, boxID, senderID string) (Event, error) {
	contentBytes, err := json.Marshal(content)
	if err != nil {
		return Event{}, merror.Transform(err).Describe("marshalling anyContent into bytes")
	}
	jsonContent := types.JSON{}
	if err := jsonContent.UnmarshalJSON(contentBytes); err != nil {
		return Event{}, merror.Transform(err).Describe("unmarshalling content bytes into types.JSON")
	}

	return New(eType, jsonContent, boxID, senderID, nil)
}

type eventFilters struct {
	id          null.String
	boxID       null.String
	eType       null.String
	eTypes      []string
	senderID    null.String
	notSenderID null.String
	referrerID  null.String
	content     *string
	unrefered   bool

	offset *int
	limit  *int
}

func get(ctx context.Context, exec boil.ContextExecutor, filters eventFilters) (Event, error) {
	var e Event

	mods, err := buildMods(ctx, exec, filters)
	if err != nil {
		return e, merror.Transform(err).Describe("building mods for event get")
	}

	dbEvent, err := sqlboiler.Events(mods...).One(ctx, exec)
	if err == sql.ErrNoRows {
		return e, merror.NotFound().Detail("id", merror.DVNotFound)
	}
	if err != nil {
		return e, merror.Transform(err).Describe("getting event")
	}
	return FromSQLBoiler(dbEvent), nil
}

func list(ctx context.Context, exec boil.ContextExecutor, filters eventFilters) ([]Event, error) {
	mods, err := buildMods(ctx, exec, filters)
	if err != nil {
		return nil, merror.Transform(err).Describe("building mods for events list")
	}

	dbEvents, err := sqlboiler.Events(mods...).All(ctx, exec)
	if err != nil {
		return nil, merror.Transform(err).Describe("listing events")
	}

	events := make([]Event, len(dbEvents))
	for i, record := range dbEvents {
		events[i] = FromSQLBoiler(record)
	}
	return events, nil
}

func buildMods(ctx context.Context, exec boil.ContextExecutor, filters eventFilters) ([]qm.QueryMod, error) {
	mods := []qm.QueryMod{
		qm.OrderBy(sqlboiler.EventColumns.CreatedAt + " DESC"),
	}
	// add id
	if filters.id.Valid {
		mods = append(mods, sqlboiler.EventWhere.ID.EQ(filters.id.String))
	}
	// add sender id
	if filters.senderID.Valid {
		mods = append(mods, sqlboiler.EventWhere.SenderID.EQ(filters.senderID.String))
	}
	// remove not sender id
	if filters.notSenderID.Valid {
		mods = append(mods, sqlboiler.EventWhere.SenderID.NEQ(filters.notSenderID.String))
	}
	// add type
	if filters.eType.Valid {
		filters.eTypes = append(filters.eTypes, filters.eType.String)
	}
	// add types
	if len(filters.eTypes) > 0 {
		mods = append(mods, sqlboiler.EventWhere.Type.IN(filters.eTypes))
	}
	// add referrer
	if filters.referrerID.Valid {
		mods = append(mods, sqlboiler.EventWhere.ReferrerID.EQ(filters.referrerID))
	}
	// add box query
	if filters.boxID.Valid {
		mods = append(mods, sqlboiler.EventWhere.BoxID.EQ(filters.boxID.String))
	}
	// add JSONB matching
	if filters.content != nil {
		mods = append(mods, qm.Where(`content::jsonb @> ?`, *filters.content))
	}
	// add offset for pagination
	if filters.offset != nil {
		mods = append(mods, qm.Offset(*filters.offset))
	}
	// add limit for pagination
	if filters.limit != nil {
		mods = append(mods, qm.Limit(*filters.limit))
	}

	// add unrefered query
	// TODO: merge this query into the main one as a sub query
	if filters.unrefered {
		notInIDs, err := referentIDs(ctx, exec, filters)
		if err != nil {
			return mods, merror.Transform(err).Describe("sub selecting referents")
		}
		if len(notInIDs) > 0 {
			mods = append(mods, qm.WhereIn(sqlboiler.EventColumns.ID+" NOT IN ?", slice.StringSliceToInterfaceSlice(notInIDs)...))
		}
	}
	return mods, nil
}

func referentIDs(ctx context.Context, exec boil.ContextExecutor, filters eventFilters) ([]string, error) {
	// first we get events that refers other event: the referents
	subMods := []qm.QueryMod{}
	// either it selects event refering another specific event
	if filters.id.Valid {
		subMods = append(subMods, sqlboiler.EventWhere.ReferrerID.EQ(filters.id))
	} else { // or it selects event refering any event for a given box or sender
		if filters.boxID.Valid {
			subMods = append(subMods, sqlboiler.EventWhere.BoxID.EQ(filters.boxID.String))
		}
		if filters.senderID.Valid {
			subMods = append(subMods, sqlboiler.EventWhere.SenderID.EQ(filters.senderID.String))
		}
		// box id or sender id should have been set or the query will be to large
		if len(subMods) == 0 {
			return nil, merror.Internal().Describe("wrong unrefered use")
		}
		subMods = append(subMods, sqlboiler.EventWhere.ReferrerID.IsNotNull())
	}
	referents, err := sqlboiler.Events(subMods...).All(ctx, exec)
	if err != nil {
		return nil, merror.Transform(err).Describe("listing referents")
	}
	// compute the list of event that are refered according to retrieved referents
	notInIDs := make([]string, len(referents))
	for i, referent := range referents {
		notInIDs[i] = referent.ReferrerID.String
	}
	return notInIDs, nil
}

func FindByEncryptedFileID(ctx context.Context, exec boil.ContextExecutor, encryptedFileID string) ([]Event, error) {
	// build expected content
	content := `{"encrypted_file_id": "` + encryptedFileID + `"}`
	events, err := list(ctx, exec, eventFilters{
		eType:   null.StringFrom("msg.file"),
		content: &content,
	})
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return events, merror.NotFound().Detail("id", merror.DVNotFound)
	}
	return events, nil
}

func ListFilesID(ctx context.Context, exec boil.ContextExecutor, boxID string) ([]string, error) {
	events, err := list(ctx, exec, eventFilters{
		boxID: null.StringFrom(boxID),
		eType: null.StringFrom("msg.file"),
	})
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(events))
	var content MsgFileContent
	for idx, event := range events {
		err = json.Unmarshal(event.JSONContent, &content)
		if err != nil {
			return nil, merror.Internal().Describe("unmarshaling content json")
		}
		ids[idx] = content.EncryptedFileID
	}

	return ids, nil
}

func CountByBoxID(ctx context.Context, exec boil.ContextExecutor, boxID string) (int, error) {
	count, err := sqlboiler.Events(sqlboiler.EventWhere.BoxID.EQ(boxID)).Count(ctx, exec)
	if err != nil {
		return 0, merror.Transform(err).Describe("retrieving db events")
	}

	return int(count), nil
}
