// Code generated by SQLBoiler 4.2.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package sqlboiler

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// IdentityProfileSharingConsent is an object representing the database table.
type IdentityProfileSharingConsent struct {
	ID              int       `boil:"id" json:"id" toml:"id" yaml:"id"`
	IdentityID      string    `boil:"identity_id" json:"identity_id" toml:"identity_id" yaml:"identity_id"`
	InformationType string    `boil:"information_type" json:"information_type" toml:"information_type" yaml:"information_type"`
	CreatedAt       time.Time `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	RevokedAt       null.Time `boil:"revoked_at" json:"revoked_at,omitempty" toml:"revoked_at" yaml:"revoked_at,omitempty"`

	R *identityProfileSharingConsentR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L identityProfileSharingConsentL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var IdentityProfileSharingConsentColumns = struct {
	ID              string
	IdentityID      string
	InformationType string
	CreatedAt       string
	RevokedAt       string
}{
	ID:              "id",
	IdentityID:      "identity_id",
	InformationType: "information_type",
	CreatedAt:       "created_at",
	RevokedAt:       "revoked_at",
}

// Generated where

var IdentityProfileSharingConsentWhere = struct {
	ID              whereHelperint
	IdentityID      whereHelperstring
	InformationType whereHelperstring
	CreatedAt       whereHelpertime_Time
	RevokedAt       whereHelpernull_Time
}{
	ID:              whereHelperint{field: "\"identity_profile_sharing_consent\".\"id\""},
	IdentityID:      whereHelperstring{field: "\"identity_profile_sharing_consent\".\"identity_id\""},
	InformationType: whereHelperstring{field: "\"identity_profile_sharing_consent\".\"information_type\""},
	CreatedAt:       whereHelpertime_Time{field: "\"identity_profile_sharing_consent\".\"created_at\""},
	RevokedAt:       whereHelpernull_Time{field: "\"identity_profile_sharing_consent\".\"revoked_at\""},
}

// IdentityProfileSharingConsentRels is where relationship names are stored.
var IdentityProfileSharingConsentRels = struct {
	Identity string
}{
	Identity: "Identity",
}

// identityProfileSharingConsentR is where relationships are stored.
type identityProfileSharingConsentR struct {
	Identity *Identity `boil:"Identity" json:"Identity" toml:"Identity" yaml:"Identity"`
}

// NewStruct creates a new relationship struct
func (*identityProfileSharingConsentR) NewStruct() *identityProfileSharingConsentR {
	return &identityProfileSharingConsentR{}
}

// identityProfileSharingConsentL is where Load methods for each relationship are stored.
type identityProfileSharingConsentL struct{}

var (
	identityProfileSharingConsentAllColumns            = []string{"id", "identity_id", "information_type", "created_at", "revoked_at"}
	identityProfileSharingConsentColumnsWithoutDefault = []string{"identity_id", "information_type", "revoked_at"}
	identityProfileSharingConsentColumnsWithDefault    = []string{"id", "created_at"}
	identityProfileSharingConsentPrimaryKeyColumns     = []string{"id"}
)

type (
	// IdentityProfileSharingConsentSlice is an alias for a slice of pointers to IdentityProfileSharingConsent.
	// This should generally be used opposed to []IdentityProfileSharingConsent.
	IdentityProfileSharingConsentSlice []*IdentityProfileSharingConsent

	identityProfileSharingConsentQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	identityProfileSharingConsentType                 = reflect.TypeOf(&IdentityProfileSharingConsent{})
	identityProfileSharingConsentMapping              = queries.MakeStructMapping(identityProfileSharingConsentType)
	identityProfileSharingConsentPrimaryKeyMapping, _ = queries.BindMapping(identityProfileSharingConsentType, identityProfileSharingConsentMapping, identityProfileSharingConsentPrimaryKeyColumns)
	identityProfileSharingConsentInsertCacheMut       sync.RWMutex
	identityProfileSharingConsentInsertCache          = make(map[string]insertCache)
	identityProfileSharingConsentUpdateCacheMut       sync.RWMutex
	identityProfileSharingConsentUpdateCache          = make(map[string]updateCache)
	identityProfileSharingConsentUpsertCacheMut       sync.RWMutex
	identityProfileSharingConsentUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

// One returns a single identityProfileSharingConsent record from the query.
func (q identityProfileSharingConsentQuery) One(ctx context.Context, exec boil.ContextExecutor) (*IdentityProfileSharingConsent, error) {
	o := &IdentityProfileSharingConsent{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "sqlboiler: failed to execute a one query for identity_profile_sharing_consent")
	}

	return o, nil
}

// All returns all IdentityProfileSharingConsent records from the query.
func (q identityProfileSharingConsentQuery) All(ctx context.Context, exec boil.ContextExecutor) (IdentityProfileSharingConsentSlice, error) {
	var o []*IdentityProfileSharingConsent

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "sqlboiler: failed to assign all query results to IdentityProfileSharingConsent slice")
	}

	return o, nil
}

// Count returns the count of all IdentityProfileSharingConsent records in the query.
func (q identityProfileSharingConsentQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to count identity_profile_sharing_consent rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q identityProfileSharingConsentQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "sqlboiler: failed to check if identity_profile_sharing_consent exists")
	}

	return count > 0, nil
}

// Identity pointed to by the foreign key.
func (o *IdentityProfileSharingConsent) Identity(mods ...qm.QueryMod) identityQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.IdentityID),
	}

	queryMods = append(queryMods, mods...)

	query := Identities(queryMods...)
	queries.SetFrom(query.Query, "\"identity\"")

	return query
}

// LoadIdentity allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (identityProfileSharingConsentL) LoadIdentity(ctx context.Context, e boil.ContextExecutor, singular bool, maybeIdentityProfileSharingConsent interface{}, mods queries.Applicator) error {
	var slice []*IdentityProfileSharingConsent
	var object *IdentityProfileSharingConsent

	if singular {
		object = maybeIdentityProfileSharingConsent.(*IdentityProfileSharingConsent)
	} else {
		slice = *maybeIdentityProfileSharingConsent.(*[]*IdentityProfileSharingConsent)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &identityProfileSharingConsentR{}
		}
		args = append(args, object.IdentityID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &identityProfileSharingConsentR{}
			}

			for _, a := range args {
				if a == obj.IdentityID {
					continue Outer
				}
			}

			args = append(args, obj.IdentityID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`identity`),
		qm.WhereIn(`identity.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Identity")
	}

	var resultSlice []*Identity
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Identity")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for identity")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for identity")
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.Identity = foreign
		if foreign.R == nil {
			foreign.R = &identityR{}
		}
		foreign.R.IdentityProfileSharingConsents = append(foreign.R.IdentityProfileSharingConsents, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.IdentityID == foreign.ID {
				local.R.Identity = foreign
				if foreign.R == nil {
					foreign.R = &identityR{}
				}
				foreign.R.IdentityProfileSharingConsents = append(foreign.R.IdentityProfileSharingConsents, local)
				break
			}
		}
	}

	return nil
}

// SetIdentity of the identityProfileSharingConsent to the related item.
// Sets o.R.Identity to related.
// Adds o to related.R.IdentityProfileSharingConsents.
func (o *IdentityProfileSharingConsent) SetIdentity(ctx context.Context, exec boil.ContextExecutor, insert bool, related *Identity) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"identity_profile_sharing_consent\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"identity_id"}),
		strmangle.WhereClause("\"", "\"", 2, identityProfileSharingConsentPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, updateQuery)
		fmt.Fprintln(writer, values)
	}
	if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.IdentityID = related.ID
	if o.R == nil {
		o.R = &identityProfileSharingConsentR{
			Identity: related,
		}
	} else {
		o.R.Identity = related
	}

	if related.R == nil {
		related.R = &identityR{
			IdentityProfileSharingConsents: IdentityProfileSharingConsentSlice{o},
		}
	} else {
		related.R.IdentityProfileSharingConsents = append(related.R.IdentityProfileSharingConsents, o)
	}

	return nil
}

// IdentityProfileSharingConsents retrieves all the records using an executor.
func IdentityProfileSharingConsents(mods ...qm.QueryMod) identityProfileSharingConsentQuery {
	mods = append(mods, qm.From("\"identity_profile_sharing_consent\""))
	return identityProfileSharingConsentQuery{NewQuery(mods...)}
}

// FindIdentityProfileSharingConsent retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindIdentityProfileSharingConsent(ctx context.Context, exec boil.ContextExecutor, iD int, selectCols ...string) (*IdentityProfileSharingConsent, error) {
	identityProfileSharingConsentObj := &IdentityProfileSharingConsent{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"identity_profile_sharing_consent\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, identityProfileSharingConsentObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "sqlboiler: unable to select from identity_profile_sharing_consent")
	}

	return identityProfileSharingConsentObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *IdentityProfileSharingConsent) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("sqlboiler: no identity_profile_sharing_consent provided for insertion")
	}

	var err error
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	nzDefaults := queries.NonZeroDefaultSet(identityProfileSharingConsentColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	identityProfileSharingConsentInsertCacheMut.RLock()
	cache, cached := identityProfileSharingConsentInsertCache[key]
	identityProfileSharingConsentInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			identityProfileSharingConsentAllColumns,
			identityProfileSharingConsentColumnsWithDefault,
			identityProfileSharingConsentColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(identityProfileSharingConsentType, identityProfileSharingConsentMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(identityProfileSharingConsentType, identityProfileSharingConsentMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"identity_profile_sharing_consent\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"identity_profile_sharing_consent\" %sDEFAULT VALUES%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			queryReturning = fmt.Sprintf(" RETURNING \"%s\"", strings.Join(returnColumns, "\",\""))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}

	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}

	if err != nil {
		return errors.Wrap(err, "sqlboiler: unable to insert into identity_profile_sharing_consent")
	}

	if !cached {
		identityProfileSharingConsentInsertCacheMut.Lock()
		identityProfileSharingConsentInsertCache[key] = cache
		identityProfileSharingConsentInsertCacheMut.Unlock()
	}

	return nil
}

// Update uses an executor to update the IdentityProfileSharingConsent.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *IdentityProfileSharingConsent) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	key := makeCacheKey(columns, nil)
	identityProfileSharingConsentUpdateCacheMut.RLock()
	cache, cached := identityProfileSharingConsentUpdateCache[key]
	identityProfileSharingConsentUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			identityProfileSharingConsentAllColumns,
			identityProfileSharingConsentPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("sqlboiler: unable to update identity_profile_sharing_consent, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"identity_profile_sharing_consent\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, identityProfileSharingConsentPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(identityProfileSharingConsentType, identityProfileSharingConsentMapping, append(wl, identityProfileSharingConsentPrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, values)
	}
	var result sql.Result
	result, err = exec.ExecContext(ctx, cache.query, values...)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to update identity_profile_sharing_consent row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to get rows affected by update for identity_profile_sharing_consent")
	}

	if !cached {
		identityProfileSharingConsentUpdateCacheMut.Lock()
		identityProfileSharingConsentUpdateCache[key] = cache
		identityProfileSharingConsentUpdateCacheMut.Unlock()
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values.
func (q identityProfileSharingConsentQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to update all for identity_profile_sharing_consent")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to retrieve rows affected for identity_profile_sharing_consent")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o IdentityProfileSharingConsentSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("sqlboiler: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), identityProfileSharingConsentPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"identity_profile_sharing_consent\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, identityProfileSharingConsentPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to update all in identityProfileSharingConsent slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to retrieve rows affected all in update all identityProfileSharingConsent")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *IdentityProfileSharingConsent) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("sqlboiler: no identity_profile_sharing_consent provided for upsert")
	}
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	nzDefaults := queries.NonZeroDefaultSet(identityProfileSharingConsentColumnsWithDefault, o)

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	if updateOnConflict {
		buf.WriteByte('t')
	} else {
		buf.WriteByte('f')
	}
	buf.WriteByte('.')
	for _, c := range conflictColumns {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	identityProfileSharingConsentUpsertCacheMut.RLock()
	cache, cached := identityProfileSharingConsentUpsertCache[key]
	identityProfileSharingConsentUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			identityProfileSharingConsentAllColumns,
			identityProfileSharingConsentColumnsWithDefault,
			identityProfileSharingConsentColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			identityProfileSharingConsentAllColumns,
			identityProfileSharingConsentPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("sqlboiler: unable to upsert identity_profile_sharing_consent, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(identityProfileSharingConsentPrimaryKeyColumns))
			copy(conflict, identityProfileSharingConsentPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"identity_profile_sharing_consent\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(identityProfileSharingConsentType, identityProfileSharingConsentMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(identityProfileSharingConsentType, identityProfileSharingConsentMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}
	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(returns...)
		if err == sql.ErrNoRows {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}
	if err != nil {
		return errors.Wrap(err, "sqlboiler: unable to upsert identity_profile_sharing_consent")
	}

	if !cached {
		identityProfileSharingConsentUpsertCacheMut.Lock()
		identityProfileSharingConsentUpsertCache[key] = cache
		identityProfileSharingConsentUpsertCacheMut.Unlock()
	}

	return nil
}

// Delete deletes a single IdentityProfileSharingConsent record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *IdentityProfileSharingConsent) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("sqlboiler: no IdentityProfileSharingConsent provided for delete")
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), identityProfileSharingConsentPrimaryKeyMapping)
	sql := "DELETE FROM \"identity_profile_sharing_consent\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to delete from identity_profile_sharing_consent")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to get rows affected by delete for identity_profile_sharing_consent")
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q identityProfileSharingConsentQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("sqlboiler: no identityProfileSharingConsentQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to delete all from identity_profile_sharing_consent")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to get rows affected by deleteall for identity_profile_sharing_consent")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o IdentityProfileSharingConsentSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), identityProfileSharingConsentPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"identity_profile_sharing_consent\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, identityProfileSharingConsentPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to delete all from identityProfileSharingConsent slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to get rows affected by deleteall for identity_profile_sharing_consent")
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *IdentityProfileSharingConsent) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindIdentityProfileSharingConsent(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *IdentityProfileSharingConsentSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := IdentityProfileSharingConsentSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), identityProfileSharingConsentPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"identity_profile_sharing_consent\".* FROM \"identity_profile_sharing_consent\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, identityProfileSharingConsentPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "sqlboiler: unable to reload all in IdentityProfileSharingConsentSlice")
	}

	*o = slice

	return nil
}

// IdentityProfileSharingConsentExists checks if the IdentityProfileSharingConsent row exists.
func IdentityProfileSharingConsentExists(ctx context.Context, exec boil.ContextExecutor, iD int) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"identity_profile_sharing_consent\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "sqlboiler: unable to check if identity_profile_sharing_consent exists")
	}

	return exists, nil
}