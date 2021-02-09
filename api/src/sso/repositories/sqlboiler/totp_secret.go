// Code generated by SQLBoiler 4.4.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
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
	"github.com/volatiletech/sqlboiler/v4/types"
	"github.com/volatiletech/strmangle"
)

// TotpSecret is an object representing the database table.
type TotpSecret struct {
	ID         int               `boil:"id" json:"id" toml:"id" yaml:"id"`
	IdentityID string            `boil:"identity_id" json:"identity_id" toml:"identity_id" yaml:"identity_id"`
	Secret     string            `boil:"secret" json:"secret" toml:"secret" yaml:"secret"`
	Backup     types.StringArray `boil:"backup" json:"backup,omitempty" toml:"backup" yaml:"backup,omitempty"`
	CreatedAt  time.Time         `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *totpSecretR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L totpSecretL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var TotpSecretColumns = struct {
	ID         string
	IdentityID string
	Secret     string
	Backup     string
	CreatedAt  string
}{
	ID:         "id",
	IdentityID: "identity_id",
	Secret:     "secret",
	Backup:     "backup",
	CreatedAt:  "created_at",
}

// Generated where

type whereHelpertypes_StringArray struct{ field string }

func (w whereHelpertypes_StringArray) EQ(x types.StringArray) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, false, x)
}
func (w whereHelpertypes_StringArray) NEQ(x types.StringArray) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, true, x)
}
func (w whereHelpertypes_StringArray) IsNull() qm.QueryMod { return qmhelper.WhereIsNull(w.field) }
func (w whereHelpertypes_StringArray) IsNotNull() qm.QueryMod {
	return qmhelper.WhereIsNotNull(w.field)
}
func (w whereHelpertypes_StringArray) LT(x types.StringArray) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpertypes_StringArray) LTE(x types.StringArray) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpertypes_StringArray) GT(x types.StringArray) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpertypes_StringArray) GTE(x types.StringArray) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

var TotpSecretWhere = struct {
	ID         whereHelperint
	IdentityID whereHelperstring
	Secret     whereHelperstring
	Backup     whereHelpertypes_StringArray
	CreatedAt  whereHelpertime_Time
}{
	ID:         whereHelperint{field: "\"totp_secret\".\"id\""},
	IdentityID: whereHelperstring{field: "\"totp_secret\".\"identity_id\""},
	Secret:     whereHelperstring{field: "\"totp_secret\".\"secret\""},
	Backup:     whereHelpertypes_StringArray{field: "\"totp_secret\".\"backup\""},
	CreatedAt:  whereHelpertime_Time{field: "\"totp_secret\".\"created_at\""},
}

// TotpSecretRels is where relationship names are stored.
var TotpSecretRels = struct {
	Identity string
}{
	Identity: "Identity",
}

// totpSecretR is where relationships are stored.
type totpSecretR struct {
	Identity *Identity `boil:"Identity" json:"Identity" toml:"Identity" yaml:"Identity"`
}

// NewStruct creates a new relationship struct
func (*totpSecretR) NewStruct() *totpSecretR {
	return &totpSecretR{}
}

// totpSecretL is where Load methods for each relationship are stored.
type totpSecretL struct{}

var (
	totpSecretAllColumns            = []string{"id", "identity_id", "secret", "backup", "created_at"}
	totpSecretColumnsWithoutDefault = []string{"identity_id", "secret", "backup", "created_at"}
	totpSecretColumnsWithDefault    = []string{"id"}
	totpSecretPrimaryKeyColumns     = []string{"id"}
)

type (
	// TotpSecretSlice is an alias for a slice of pointers to TotpSecret.
	// This should generally be used opposed to []TotpSecret.
	TotpSecretSlice []*TotpSecret

	totpSecretQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	totpSecretType                 = reflect.TypeOf(&TotpSecret{})
	totpSecretMapping              = queries.MakeStructMapping(totpSecretType)
	totpSecretPrimaryKeyMapping, _ = queries.BindMapping(totpSecretType, totpSecretMapping, totpSecretPrimaryKeyColumns)
	totpSecretInsertCacheMut       sync.RWMutex
	totpSecretInsertCache          = make(map[string]insertCache)
	totpSecretUpdateCacheMut       sync.RWMutex
	totpSecretUpdateCache          = make(map[string]updateCache)
	totpSecretUpsertCacheMut       sync.RWMutex
	totpSecretUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

// One returns a single totpSecret record from the query.
func (q totpSecretQuery) One(ctx context.Context, exec boil.ContextExecutor) (*TotpSecret, error) {
	o := &TotpSecret{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "sqlboiler: failed to execute a one query for totp_secret")
	}

	return o, nil
}

// All returns all TotpSecret records from the query.
func (q totpSecretQuery) All(ctx context.Context, exec boil.ContextExecutor) (TotpSecretSlice, error) {
	var o []*TotpSecret

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "sqlboiler: failed to assign all query results to TotpSecret slice")
	}

	return o, nil
}

// Count returns the count of all TotpSecret records in the query.
func (q totpSecretQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to count totp_secret rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q totpSecretQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "sqlboiler: failed to check if totp_secret exists")
	}

	return count > 0, nil
}

// Identity pointed to by the foreign key.
func (o *TotpSecret) Identity(mods ...qm.QueryMod) identityQuery {
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
func (totpSecretL) LoadIdentity(ctx context.Context, e boil.ContextExecutor, singular bool, maybeTotpSecret interface{}, mods queries.Applicator) error {
	var slice []*TotpSecret
	var object *TotpSecret

	if singular {
		object = maybeTotpSecret.(*TotpSecret)
	} else {
		slice = *maybeTotpSecret.(*[]*TotpSecret)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &totpSecretR{}
		}
		args = append(args, object.IdentityID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &totpSecretR{}
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
		foreign.R.TotpSecret = object
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.IdentityID == foreign.ID {
				local.R.Identity = foreign
				if foreign.R == nil {
					foreign.R = &identityR{}
				}
				foreign.R.TotpSecret = local
				break
			}
		}
	}

	return nil
}

// SetIdentity of the totpSecret to the related item.
// Sets o.R.Identity to related.
// Adds o to related.R.TotpSecret.
func (o *TotpSecret) SetIdentity(ctx context.Context, exec boil.ContextExecutor, insert bool, related *Identity) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"totp_secret\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"identity_id"}),
		strmangle.WhereClause("\"", "\"", 2, totpSecretPrimaryKeyColumns),
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
		o.R = &totpSecretR{
			Identity: related,
		}
	} else {
		o.R.Identity = related
	}

	if related.R == nil {
		related.R = &identityR{
			TotpSecret: o,
		}
	} else {
		related.R.TotpSecret = o
	}

	return nil
}

// TotpSecrets retrieves all the records using an executor.
func TotpSecrets(mods ...qm.QueryMod) totpSecretQuery {
	mods = append(mods, qm.From("\"totp_secret\""))
	return totpSecretQuery{NewQuery(mods...)}
}

// FindTotpSecret retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindTotpSecret(ctx context.Context, exec boil.ContextExecutor, iD int, selectCols ...string) (*TotpSecret, error) {
	totpSecretObj := &TotpSecret{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"totp_secret\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, totpSecretObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "sqlboiler: unable to select from totp_secret")
	}

	return totpSecretObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *TotpSecret) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("sqlboiler: no totp_secret provided for insertion")
	}

	var err error
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	nzDefaults := queries.NonZeroDefaultSet(totpSecretColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	totpSecretInsertCacheMut.RLock()
	cache, cached := totpSecretInsertCache[key]
	totpSecretInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			totpSecretAllColumns,
			totpSecretColumnsWithDefault,
			totpSecretColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(totpSecretType, totpSecretMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(totpSecretType, totpSecretMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"totp_secret\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"totp_secret\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "sqlboiler: unable to insert into totp_secret")
	}

	if !cached {
		totpSecretInsertCacheMut.Lock()
		totpSecretInsertCache[key] = cache
		totpSecretInsertCacheMut.Unlock()
	}

	return nil
}

// Update uses an executor to update the TotpSecret.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *TotpSecret) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	key := makeCacheKey(columns, nil)
	totpSecretUpdateCacheMut.RLock()
	cache, cached := totpSecretUpdateCache[key]
	totpSecretUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			totpSecretAllColumns,
			totpSecretPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("sqlboiler: unable to update totp_secret, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"totp_secret\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, totpSecretPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(totpSecretType, totpSecretMapping, append(wl, totpSecretPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "sqlboiler: unable to update totp_secret row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to get rows affected by update for totp_secret")
	}

	if !cached {
		totpSecretUpdateCacheMut.Lock()
		totpSecretUpdateCache[key] = cache
		totpSecretUpdateCacheMut.Unlock()
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values.
func (q totpSecretQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to update all for totp_secret")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to retrieve rows affected for totp_secret")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o TotpSecretSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), totpSecretPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"totp_secret\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, totpSecretPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to update all in totpSecret slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to retrieve rows affected all in update all totpSecret")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *TotpSecret) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("sqlboiler: no totp_secret provided for upsert")
	}
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	nzDefaults := queries.NonZeroDefaultSet(totpSecretColumnsWithDefault, o)

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

	totpSecretUpsertCacheMut.RLock()
	cache, cached := totpSecretUpsertCache[key]
	totpSecretUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			totpSecretAllColumns,
			totpSecretColumnsWithDefault,
			totpSecretColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			totpSecretAllColumns,
			totpSecretPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("sqlboiler: unable to upsert totp_secret, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(totpSecretPrimaryKeyColumns))
			copy(conflict, totpSecretPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"totp_secret\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(totpSecretType, totpSecretMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(totpSecretType, totpSecretMapping, ret)
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
		return errors.Wrap(err, "sqlboiler: unable to upsert totp_secret")
	}

	if !cached {
		totpSecretUpsertCacheMut.Lock()
		totpSecretUpsertCache[key] = cache
		totpSecretUpsertCacheMut.Unlock()
	}

	return nil
}

// Delete deletes a single TotpSecret record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *TotpSecret) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("sqlboiler: no TotpSecret provided for delete")
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), totpSecretPrimaryKeyMapping)
	sql := "DELETE FROM \"totp_secret\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to delete from totp_secret")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to get rows affected by delete for totp_secret")
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q totpSecretQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("sqlboiler: no totpSecretQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to delete all from totp_secret")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to get rows affected by deleteall for totp_secret")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o TotpSecretSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), totpSecretPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"totp_secret\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, totpSecretPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to delete all from totpSecret slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to get rows affected by deleteall for totp_secret")
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *TotpSecret) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindTotpSecret(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *TotpSecretSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := TotpSecretSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), totpSecretPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"totp_secret\".* FROM \"totp_secret\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, totpSecretPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "sqlboiler: unable to reload all in TotpSecretSlice")
	}

	*o = slice

	return nil
}

// TotpSecretExists checks if the TotpSecret row exists.
func TotpSecretExists(ctx context.Context, exec boil.ContextExecutor, iD int) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"totp_secret\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "sqlboiler: unable to check if totp_secret exists")
	}

	return exists, nil
}
