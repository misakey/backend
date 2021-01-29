package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events/etype"
	"gitlab.misakey.dev/misakey/backend/api/src/box/repositories/sqlboiler"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/slice"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"
)

// Event ...
type Event struct {
	ID          string
	CreatedAt   time.Time
	BoxID       string
	SenderID    string
	Type        string
	ReferrerID  null.String
	JSONContent types.JSON

	Content             interface{}
	MetadataForHandlers MetadataForUsedSpaceHandler
	ownerOrgID          null.String
}

// New ...
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
		return event, merr.From(err).Desc("binding content")
	}

	event.ID, err = uuid.NewString()
	if err != nil {
		return event, merr.From(err).Desc("generating event id")
	}
	return event, nil
}

func newWithAnyContent(eType string, content anyContent, boxID, senderID string, referrerID *string) (Event, error) {
	contentBytes, err := json.Marshal(content)
	if err != nil {
		return Event{}, merr.From(err).Desc("marshalling anyContent into bytes")
	}
	jsonContent := types.JSON{}
	if err := jsonContent.UnmarshalJSON(contentBytes); err != nil {
		return Event{}, merr.From(err).Desc("unmarshalling content bytes into types.JSON")
	}

	return New(eType, jsonContent, boxID, senderID, referrerID)
}

func (e *Event) persist(ctx context.Context, exec boil.ContextExecutor) error {
	// finally insert
	if err := e.ToSQLBoiler().Insert(ctx, exec, boil.Infer()); err != nil {
		return merr.From(err).Desc("inserting event in DB")
	}
	return nil
}

// GetLast ...
func GetLast(ctx context.Context, exec boil.ContextExecutor, boxID string) (Event, error) {
	return get(ctx, exec, eventFilters{
		boxID:  null.StringFrom(boxID),
		eTypes: etype.MembersCanSee,
	})
}

// ListForMembersByBoxID ...
func ListForMembersByBoxID(ctx context.Context, exec boil.ContextExecutor, boxID string, offset, limit *int) ([]Event, error) {
	return list(ctx, exec, eventFilters{
		boxID:  null.StringFrom(boxID),
		offset: offset,
		limit:  limit,
		eTypes: etype.MembersCanSee,
	})
}

// ListLastestForEachBoxID returns the latest events of each box id.
func ListLastestForEachBoxID(ctx context.Context, exec boil.ContextExecutor, boxIDs []string) ([]Event, error) {
	mods := []qm.QueryMod{
		qm.Select("DISTINCT ON (box_id) box_id, event.*"),
		sqlboiler.EventWhere.BoxID.IN(boxIDs),
		qm.OrderBy(sqlboiler.EventColumns.BoxID),
		qm.OrderBy(sqlboiler.EventColumns.CreatedAt + " DESC"),
	}

	// retrieve last events
	records, err := sqlboiler.Events(mods...).All(ctx, exec)
	if err != nil {
		return []Event{}, merr.From(err).Desc("retrieving last events")
	}

	// build the final last events list
	lastEvents := make([]Event, len(records))
	for idx, event := range records {
		lastEvents[idx] = FromSQLBoiler(event)
	}
	sort.Slice(lastEvents, func(i, j int) bool { return lastEvents[i].CreatedAt.Unix() > lastEvents[j].CreatedAt.Unix() })
	return lastEvents, nil
}

// ListFilesForMembersByBoxID ...
func ListFilesForMembersByBoxID(ctx context.Context, exec boil.ContextExecutor, boxID string, offset, limit *int) ([]Event, error) {
	return list(ctx, exec, eventFilters{
		boxID:      null.StringFrom(boxID),
		offset:     offset,
		limit:      limit,
		eTypes:     []string{etype.Msgfile},
		unreferred: true,
	})
}

// ListForBuild ...
func ListForBuild(ctx context.Context, exec boil.ContextExecutor, boxID string) ([]Event, error) {
	return list(ctx, exec, eventFilters{
		boxID:  null.StringFrom(boxID),
		eTypes: etype.RequireToBuild,
	})
}

// ListByTypeAndBoxIDAndSenderID ...
func ListByTypeAndBoxIDAndSenderID(ctx context.Context, exec boil.ContextExecutor, eventType, boxID, senderID string) ([]Event, error) {
	mods := []qm.QueryMod{
		sqlboiler.EventWhere.BoxID.EQ(boxID),
		sqlboiler.EventWhere.Type.EQ(eventType),
		sqlboiler.EventWhere.SenderID.EQ(senderID),
		qm.OrderBy(sqlboiler.EventColumns.CreatedAt + " DESC"),
	}

	dbEvents, err := sqlboiler.Events(mods...).All(ctx, exec)
	if err != nil {
		return nil, merr.From(err).Desc("retrieving db events")
	}

	events := make([]Event, len(dbEvents))
	for i, record := range dbEvents {
		events[i] = FromSQLBoiler(record)
	}

	if len(events) == 0 {
		return events, merr.NotFound().Add("id", merr.DVNotFound)
	}

	return events, nil
}

//TODO (perf): struct of size 240 bytes could be of size 224 bytes (maligned)
type eventFilters struct {
	// focus on one column filter
	idOnly    bool
	boxIDOnly bool

	// classic filters
	id          null.String
	boxID       null.String
	boxIDs      []string
	eType       null.String
	eTypes      []string
	senderID    null.String
	referrerID  null.String
	referrerIDs []string

	// filters triggering in jsonb research
	content          *string
	fileID           null.String
	restrictionType  null.String
	restrictionTypes []string
	accessValue      null.String

	// ensure the event is not referred by another one
	unreferred bool

	// pagination
	offset *int
	limit  *int
}

func get(ctx context.Context, exec boil.ContextExecutor, filters eventFilters) (Event, error) {
	var e Event

	// NOTE: buildMods should always sort event with the most recent one on top of it
	mods, err := buildMods(ctx, exec, filters)
	if err != nil {
		return e, merr.From(err).Desc("building mods for event get")
	}

	dbEvent, err := sqlboiler.Events(mods...).One(ctx, exec)
	if err == sql.ErrNoRows {
		return e, merr.NotFound().Add("id", merr.DVNotFound)
	}
	if err != nil {
		return e, merr.From(err).Desc("getting event")
	}
	return FromSQLBoiler(dbEvent), nil
}

func listEventAndReferrers(ctx context.Context, exec boil.ContextExecutor, id string) ([]Event, error) {
	mods := []qm.QueryMod{
		sqlboiler.EventWhere.ID.EQ(id),
		qm.Or2(sqlboiler.EventWhere.ReferrerID.EQ(null.StringFrom(id))),
		qm.OrderBy(sqlboiler.EventColumns.CreatedAt + " ASC"),
	}

	dbEvents, err := sqlboiler.Events(mods...).All(ctx, exec)
	if err != nil {
		return nil, err
	}
	if len(dbEvents) == 0 {
		return nil, merr.NotFound().Desc("listing events")
	}

	events := make([]Event, len(dbEvents))
	for i, record := range dbEvents {
		events[i] = FromSQLBoiler(record)
	}

	return events, nil
}

func list(ctx context.Context, exec boil.ContextExecutor, filters eventFilters) ([]Event, error) {
	mods, err := buildMods(ctx, exec, filters)
	if err != nil {
		return nil, merr.From(err).Desc("building mods for events list")
	}

	dbEvents, err := sqlboiler.Events(mods...).All(ctx, exec)
	if err != nil {
		return nil, err
	}

	events := make([]Event, len(dbEvents))
	for i, record := range dbEvents {
		events[i] = FromSQLBoiler(record)
	}
	return events, nil
}

func buildMods(ctx context.Context, exec boil.ContextExecutor, filters eventFilters) ([]qm.QueryMod, error) {
	mods := []qm.QueryMod{
		// NOTE: get() function count on this to always get the latest event
		qm.OrderBy(sqlboiler.EventColumns.CreatedAt + " DESC"),
	}
	// select only the id if asked
	if filters.idOnly {
		mods = append(mods, qm.Select(sqlboiler.EventColumns.ID))
	}
	// select only the id if asked
	if filters.boxIDOnly {
		mods = append(mods, qm.Select(sqlboiler.EventColumns.BoxID))
	}
	// add id
	if filters.id.Valid {
		mods = append(mods, sqlboiler.EventWhere.ID.EQ(filters.id.String))
	}
	// add sender id
	if filters.senderID.Valid {
		mods = append(mods, sqlboiler.EventWhere.SenderID.EQ(filters.senderID.String))
	}
	// add type
	// NOTE: must be handled before filters.eTypes
	if filters.eType.Valid {
		filters.eTypes = append(filters.eTypes, filters.eType.String)
	}
	// add types
	// NOTE: must be handled after filters.eType
	if len(filters.eTypes) > 0 {
		mods = append(mods, sqlboiler.EventWhere.Type.IN(filters.eTypes))
	}
	// add referrer id
	if filters.referrerID.Valid {
		filters.referrerIDs = append(filters.referrerIDs, filters.referrerID.String)
	}
	// add referrers ids
	if len(filters.referrerIDs) > 0 {
		// Note that there is no risk of SQL injection
		// even if we build a query string ourselves
		// because there is no way the attacker can control `sqlboiler.EventColumns.ReferrerID`
		mods = append(mods, qm.AndIn(sqlboiler.EventColumns.ReferrerID+" IN ?", slice.StringSliceToInterfaceSlice(filters.referrerIDs)...))
	}
	// add box id queries
	if filters.boxID.Valid {
		filters.boxIDs = append(filters.boxIDs, filters.boxID.String)
	}
	if len(filters.boxIDs) > 0 {
		mods = append(mods, sqlboiler.EventWhere.BoxID.IN(filters.boxIDs))
	}

	// add JSONB matching
	if filters.content != nil {
		mods = append(mods, qm.Where(`content::jsonb @> ?`, *filters.content))
	}
	// add encrypted file id JSONB matching
	if filters.fileID.Valid {
		mods = append(mods, qm.Where(`content->>'encrypted_file_id' = ?`, filters.fileID.String))
	}

	// add restriction type in restrictionTypes slice
	if filters.restrictionType.Valid {
		filters.restrictionTypes = append(filters.restrictionTypes, filters.restrictionType.String)
	}

	// add restriction types slices in JSONB matching
	if len(filters.restrictionTypes) > 0 {
		mods = append(mods, qm.WhereIn(`content->>'restriction_type' IN ?`, slice.StringSliceToInterfaceSlice(filters.restrictionTypes)...))
	}

	// add access value type in JSONB matching
	if filters.accessValue.Valid {
		mods = append(mods, qm.Where(`content->>'value' = ?`, filters.accessValue.String))
	}

	// add offset for pagination
	if filters.offset != nil {
		mods = append(mods, qm.Offset(*filters.offset))
	}
	// add limit for pagination
	if filters.limit != nil {
		mods = append(mods, qm.Limit(*filters.limit))
	}

	// add unreferred query
	// TODO: merge this query into the main one as a sub query
	if filters.unreferred {
		notInIDs, err := referentIDs(ctx, exec, filters)
		if err != nil {
			return mods, merr.From(err).Desc("sub selecting referents")
		}
		if len(notInIDs) > 0 {
			mods = append(mods, qm.WhereIn(sqlboiler.EventColumns.ID+" NOT IN ?", slice.StringSliceToInterfaceSlice(notInIDs)...))
		}
	}
	return mods, nil
}

func referentIDs(ctx context.Context, exec boil.ContextExecutor, filters eventFilters) ([]string, error) {
	// first we get events that refers other event: the referents
	subMods := []qm.QueryMod{
		qm.Select(sqlboiler.EventColumns.ReferrerID),
	}
	// either it selects event referring another specific event
	if filters.id.Valid {
		subMods = append(subMods, sqlboiler.EventWhere.ReferrerID.EQ(filters.id))
	} else {
		subMods = append(subMods, sqlboiler.EventWhere.ReferrerID.IsNotNull())

		// check we don't face the cases we should never use
		if filters.boxID.IsZero() && filters.senderID.IsZero() {
			return nil, merr.Internal().Desc("wrong unreferred use")
		}

		// NOTE: boxID must be checked before senderID - both cannot be used at the same time
		// TODO (usage): need to improve this query to be more natural to build
		if filters.boxID.Valid {
			subMods = append(subMods, sqlboiler.EventWhere.BoxID.EQ(filters.boxID.String))
		} else if filters.senderID.Valid {
			subMods = append(subMods, sqlboiler.EventWhere.SenderID.EQ(filters.senderID.String))
		}
	}
	referents, err := sqlboiler.Events(subMods...).All(ctx, exec)
	if err != nil {
		return nil, merr.From(err).Desc("listing referents")
	}
	// compute the list of event that are referred according to retrieved referents
	notInIDs := make([]string, len(referents))
	for i, referent := range referents {
		notInIDs[i] = referent.ReferrerID.String
	}
	return notInIDs, nil
}

// ListFilesID ...
func ListFilesID(ctx context.Context, exec boil.ContextExecutor, boxID string) ([]string, error) {
	events, err := list(ctx, exec, eventFilters{
		boxID: null.StringFrom(boxID),
		eType: null.StringFrom(etype.Msgfile),
	})
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(events))
	var content MsgFileContent
	for idx, event := range events {
		err = json.Unmarshal(event.JSONContent, &content)
		if err != nil {
			return nil, merr.Internal().Desc("unmarshaling content json")
		}
		ids[idx] = content.EncryptedFileID
	}

	return ids, nil
}

// CountByBoxID ...
func CountByBoxID(ctx context.Context, exec boil.ContextExecutor, boxID string) (int, error) {
	mods := []qm.QueryMod{
		sqlboiler.EventWhere.BoxID.EQ(boxID),
		sqlboiler.EventWhere.Type.IN(etype.MembersCanSee),
	}
	count, err := sqlboiler.Events(mods...).Count(ctx, exec)
	if err != nil {
		return 0, merr.From(err).Desc("retrieving db events")
	}

	return int(count), nil
}

// CountFilesByBoxID ...
func CountFilesByBoxID(ctx context.Context, exec boil.ContextExecutor, boxID string) (int, error) {
	// by default, count only the events a member can see
	mods := []qm.QueryMod{
		sqlboiler.EventWhere.BoxID.EQ(boxID),
		sqlboiler.EventWhere.Type.IN([]string{etype.Msgfile}),
	}
	filters := eventFilters{
		boxID: null.StringFrom(boxID),
		eType: null.StringFrom(etype.Msgfile),
	}
	notInIDs, err := referentIDs(ctx, exec, filters)
	if err != nil {
		return 0, merr.From(err).Desc("sub selecting referents")
	}
	if len(notInIDs) > 0 {
		mods = append(mods, qm.WhereIn(sqlboiler.EventColumns.ID+" NOT IN ?", slice.StringSliceToInterfaceSlice(notInIDs)...))
	}
	count, err := sqlboiler.Events(mods...).Count(ctx, exec)
	if err != nil {
		return 0, merr.From(err).Desc("retrieving db events")
	}

	return int(count), nil
}
