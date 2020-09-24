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
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// BoxKeyShare is an object representing the database table.
type BoxKeyShare struct {
	OtherShareHash string    `boil:"other_share_hash" json:"other_share_hash" toml:"other_share_hash" yaml:"other_share_hash"`
	Share          string    `boil:"share" json:"share" toml:"share" yaml:"share"`
	CreatedAt      time.Time `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	BoxID          string    `boil:"box_id" json:"box_id" toml:"box_id" yaml:"box_id"`
	CreatorID      string    `boil:"creator_id" json:"creator_id" toml:"creator_id" yaml:"creator_id"`

	R *boxKeyShareR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L boxKeyShareL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var BoxKeyShareColumns = struct {
	OtherShareHash string
	Share          string
	CreatedAt      string
	BoxID          string
	CreatorID      string
}{
	OtherShareHash: "other_share_hash",
	Share:          "share",
	CreatedAt:      "created_at",
	BoxID:          "box_id",
	CreatorID:      "creator_id",
}

// Generated where

type whereHelperstring struct{ field string }

func (w whereHelperstring) EQ(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperstring) NEQ(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperstring) LT(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperstring) LTE(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperstring) GT(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperstring) GTE(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }
func (w whereHelperstring) IN(slice []string) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelperstring) NIN(slice []string) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

type whereHelpertime_Time struct{ field string }

func (w whereHelpertime_Time) EQ(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.EQ, x)
}
func (w whereHelpertime_Time) NEQ(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.NEQ, x)
}
func (w whereHelpertime_Time) LT(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpertime_Time) LTE(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpertime_Time) GT(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpertime_Time) GTE(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

var BoxKeyShareWhere = struct {
	OtherShareHash whereHelperstring
	Share          whereHelperstring
	CreatedAt      whereHelpertime_Time
	BoxID          whereHelperstring
	CreatorID      whereHelperstring
}{
	OtherShareHash: whereHelperstring{field: "\"box_key_share\".\"other_share_hash\""},
	Share:          whereHelperstring{field: "\"box_key_share\".\"share\""},
	CreatedAt:      whereHelpertime_Time{field: "\"box_key_share\".\"created_at\""},
	BoxID:          whereHelperstring{field: "\"box_key_share\".\"box_id\""},
	CreatorID:      whereHelperstring{field: "\"box_key_share\".\"creator_id\""},
}

// BoxKeyShareRels is where relationship names are stored.
var BoxKeyShareRels = struct {
}{}

// boxKeyShareR is where relationships are stored.
type boxKeyShareR struct {
}

// NewStruct creates a new relationship struct
func (*boxKeyShareR) NewStruct() *boxKeyShareR {
	return &boxKeyShareR{}
}

// boxKeyShareL is where Load methods for each relationship are stored.
type boxKeyShareL struct{}

var (
	boxKeyShareAllColumns            = []string{"other_share_hash", "share", "created_at", "box_id", "creator_id"}
	boxKeyShareColumnsWithoutDefault = []string{"other_share_hash", "share", "created_at", "box_id", "creator_id"}
	boxKeyShareColumnsWithDefault    = []string{}
	boxKeySharePrimaryKeyColumns     = []string{"other_share_hash"}
)

type (
	// BoxKeyShareSlice is an alias for a slice of pointers to BoxKeyShare.
	// This should generally be used opposed to []BoxKeyShare.
	BoxKeyShareSlice []*BoxKeyShare

	boxKeyShareQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	boxKeyShareType                 = reflect.TypeOf(&BoxKeyShare{})
	boxKeyShareMapping              = queries.MakeStructMapping(boxKeyShareType)
	boxKeySharePrimaryKeyMapping, _ = queries.BindMapping(boxKeyShareType, boxKeyShareMapping, boxKeySharePrimaryKeyColumns)
	boxKeyShareInsertCacheMut       sync.RWMutex
	boxKeyShareInsertCache          = make(map[string]insertCache)
	boxKeyShareUpdateCacheMut       sync.RWMutex
	boxKeyShareUpdateCache          = make(map[string]updateCache)
	boxKeyShareUpsertCacheMut       sync.RWMutex
	boxKeyShareUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

// One returns a single boxKeyShare record from the query.
func (q boxKeyShareQuery) One(ctx context.Context, exec boil.ContextExecutor) (*BoxKeyShare, error) {
	o := &BoxKeyShare{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "sqlboiler: failed to execute a one query for box_key_share")
	}

	return o, nil
}

// All returns all BoxKeyShare records from the query.
func (q boxKeyShareQuery) All(ctx context.Context, exec boil.ContextExecutor) (BoxKeyShareSlice, error) {
	var o []*BoxKeyShare

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "sqlboiler: failed to assign all query results to BoxKeyShare slice")
	}

	return o, nil
}

// Count returns the count of all BoxKeyShare records in the query.
func (q boxKeyShareQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to count box_key_share rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q boxKeyShareQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "sqlboiler: failed to check if box_key_share exists")
	}

	return count > 0, nil
}

// BoxKeyShares retrieves all the records using an executor.
func BoxKeyShares(mods ...qm.QueryMod) boxKeyShareQuery {
	mods = append(mods, qm.From("\"box_key_share\""))
	return boxKeyShareQuery{NewQuery(mods...)}
}

// FindBoxKeyShare retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindBoxKeyShare(ctx context.Context, exec boil.ContextExecutor, otherShareHash string, selectCols ...string) (*BoxKeyShare, error) {
	boxKeyShareObj := &BoxKeyShare{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"box_key_share\" where \"other_share_hash\"=$1", sel,
	)

	q := queries.Raw(query, otherShareHash)

	err := q.Bind(ctx, exec, boxKeyShareObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "sqlboiler: unable to select from box_key_share")
	}

	return boxKeyShareObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *BoxKeyShare) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("sqlboiler: no box_key_share provided for insertion")
	}

	var err error
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	nzDefaults := queries.NonZeroDefaultSet(boxKeyShareColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	boxKeyShareInsertCacheMut.RLock()
	cache, cached := boxKeyShareInsertCache[key]
	boxKeyShareInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			boxKeyShareAllColumns,
			boxKeyShareColumnsWithDefault,
			boxKeyShareColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(boxKeyShareType, boxKeyShareMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(boxKeyShareType, boxKeyShareMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"box_key_share\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"box_key_share\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "sqlboiler: unable to insert into box_key_share")
	}

	if !cached {
		boxKeyShareInsertCacheMut.Lock()
		boxKeyShareInsertCache[key] = cache
		boxKeyShareInsertCacheMut.Unlock()
	}

	return nil
}

// Update uses an executor to update the BoxKeyShare.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *BoxKeyShare) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	key := makeCacheKey(columns, nil)
	boxKeyShareUpdateCacheMut.RLock()
	cache, cached := boxKeyShareUpdateCache[key]
	boxKeyShareUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			boxKeyShareAllColumns,
			boxKeySharePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("sqlboiler: unable to update box_key_share, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"box_key_share\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, boxKeySharePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(boxKeyShareType, boxKeyShareMapping, append(wl, boxKeySharePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "sqlboiler: unable to update box_key_share row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to get rows affected by update for box_key_share")
	}

	if !cached {
		boxKeyShareUpdateCacheMut.Lock()
		boxKeyShareUpdateCache[key] = cache
		boxKeyShareUpdateCacheMut.Unlock()
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values.
func (q boxKeyShareQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to update all for box_key_share")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to retrieve rows affected for box_key_share")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o BoxKeyShareSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), boxKeySharePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"box_key_share\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, boxKeySharePrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to update all in boxKeyShare slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to retrieve rows affected all in update all boxKeyShare")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *BoxKeyShare) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("sqlboiler: no box_key_share provided for upsert")
	}
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	nzDefaults := queries.NonZeroDefaultSet(boxKeyShareColumnsWithDefault, o)

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

	boxKeyShareUpsertCacheMut.RLock()
	cache, cached := boxKeyShareUpsertCache[key]
	boxKeyShareUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			boxKeyShareAllColumns,
			boxKeyShareColumnsWithDefault,
			boxKeyShareColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			boxKeyShareAllColumns,
			boxKeySharePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("sqlboiler: unable to upsert box_key_share, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(boxKeySharePrimaryKeyColumns))
			copy(conflict, boxKeySharePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"box_key_share\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(boxKeyShareType, boxKeyShareMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(boxKeyShareType, boxKeyShareMapping, ret)
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
		return errors.Wrap(err, "sqlboiler: unable to upsert box_key_share")
	}

	if !cached {
		boxKeyShareUpsertCacheMut.Lock()
		boxKeyShareUpsertCache[key] = cache
		boxKeyShareUpsertCacheMut.Unlock()
	}

	return nil
}

// Delete deletes a single BoxKeyShare record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *BoxKeyShare) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("sqlboiler: no BoxKeyShare provided for delete")
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), boxKeySharePrimaryKeyMapping)
	sql := "DELETE FROM \"box_key_share\" WHERE \"other_share_hash\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to delete from box_key_share")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to get rows affected by delete for box_key_share")
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q boxKeyShareQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("sqlboiler: no boxKeyShareQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to delete all from box_key_share")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to get rows affected by deleteall for box_key_share")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o BoxKeyShareSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), boxKeySharePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"box_key_share\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, boxKeySharePrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to delete all from boxKeyShare slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to get rows affected by deleteall for box_key_share")
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *BoxKeyShare) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindBoxKeyShare(ctx, exec, o.OtherShareHash)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *BoxKeyShareSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := BoxKeyShareSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), boxKeySharePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"box_key_share\".* FROM \"box_key_share\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, boxKeySharePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "sqlboiler: unable to reload all in BoxKeyShareSlice")
	}

	*o = slice

	return nil
}

// BoxKeyShareExists checks if the BoxKeyShare row exists.
func BoxKeyShareExists(ctx context.Context, exec boil.ContextExecutor, otherShareHash string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"box_key_share\" where \"other_share_hash\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, otherShareHash)
	}
	row := exec.QueryRowContext(ctx, sql, otherShareHash)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "sqlboiler: unable to check if box_key_share exists")
	}

	return exists, nil
}
