// Code generated by SQLBoiler 3.7.1 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
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
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"github.com/volatiletech/sqlboiler/queries/qmhelper"
	"github.com/volatiletech/sqlboiler/strmangle"
)

// SavedFile is an object representing the database table.
type SavedFile struct {
	ID                string    `boil:"id" json:"id" toml:"id" yaml:"id"`
	IdentityID        string    `boil:"identity_id" json:"identity_id" toml:"identity_id" yaml:"identity_id"`
	EncryptedFileID   string    `boil:"encrypted_file_id" json:"encrypted_file_id" toml:"encrypted_file_id" yaml:"encrypted_file_id"`
	EncryptedMetadata string    `boil:"encrypted_metadata" json:"encrypted_metadata" toml:"encrypted_metadata" yaml:"encrypted_metadata"`
	KeyFingerprint    string    `boil:"key_fingerprint" json:"key_fingerprint" toml:"key_fingerprint" yaml:"key_fingerprint"`
	CreatedAt         time.Time `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *savedFileR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L savedFileL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var SavedFileColumns = struct {
	ID                string
	IdentityID        string
	EncryptedFileID   string
	EncryptedMetadata string
	KeyFingerprint    string
	CreatedAt         string
}{
	ID:                "id",
	IdentityID:        "identity_id",
	EncryptedFileID:   "encrypted_file_id",
	EncryptedMetadata: "encrypted_metadata",
	KeyFingerprint:    "key_fingerprint",
	CreatedAt:         "created_at",
}

// Generated where

var SavedFileWhere = struct {
	ID                whereHelperstring
	IdentityID        whereHelperstring
	EncryptedFileID   whereHelperstring
	EncryptedMetadata whereHelperstring
	KeyFingerprint    whereHelperstring
	CreatedAt         whereHelpertime_Time
}{
	ID:                whereHelperstring{field: "\"saved_file\".\"id\""},
	IdentityID:        whereHelperstring{field: "\"saved_file\".\"identity_id\""},
	EncryptedFileID:   whereHelperstring{field: "\"saved_file\".\"encrypted_file_id\""},
	EncryptedMetadata: whereHelperstring{field: "\"saved_file\".\"encrypted_metadata\""},
	KeyFingerprint:    whereHelperstring{field: "\"saved_file\".\"key_fingerprint\""},
	CreatedAt:         whereHelpertime_Time{field: "\"saved_file\".\"created_at\""},
}

// SavedFileRels is where relationship names are stored.
var SavedFileRels = struct {
	EncryptedFile string
}{
	EncryptedFile: "EncryptedFile",
}

// savedFileR is where relationships are stored.
type savedFileR struct {
	EncryptedFile *EncryptedFile
}

// NewStruct creates a new relationship struct
func (*savedFileR) NewStruct() *savedFileR {
	return &savedFileR{}
}

// savedFileL is where Load methods for each relationship are stored.
type savedFileL struct{}

var (
	savedFileAllColumns            = []string{"id", "identity_id", "encrypted_file_id", "encrypted_metadata", "key_fingerprint", "created_at"}
	savedFileColumnsWithoutDefault = []string{"id", "identity_id", "encrypted_file_id", "encrypted_metadata", "key_fingerprint", "created_at"}
	savedFileColumnsWithDefault    = []string{}
	savedFilePrimaryKeyColumns     = []string{"id"}
)

type (
	// SavedFileSlice is an alias for a slice of pointers to SavedFile.
	// This should generally be used opposed to []SavedFile.
	SavedFileSlice []*SavedFile

	savedFileQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	savedFileType                 = reflect.TypeOf(&SavedFile{})
	savedFileMapping              = queries.MakeStructMapping(savedFileType)
	savedFilePrimaryKeyMapping, _ = queries.BindMapping(savedFileType, savedFileMapping, savedFilePrimaryKeyColumns)
	savedFileInsertCacheMut       sync.RWMutex
	savedFileInsertCache          = make(map[string]insertCache)
	savedFileUpdateCacheMut       sync.RWMutex
	savedFileUpdateCache          = make(map[string]updateCache)
	savedFileUpsertCacheMut       sync.RWMutex
	savedFileUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

// One returns a single savedFile record from the query.
func (q savedFileQuery) One(ctx context.Context, exec boil.ContextExecutor) (*SavedFile, error) {
	o := &SavedFile{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "sqlboiler: failed to execute a one query for saved_file")
	}

	return o, nil
}

// All returns all SavedFile records from the query.
func (q savedFileQuery) All(ctx context.Context, exec boil.ContextExecutor) (SavedFileSlice, error) {
	var o []*SavedFile

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "sqlboiler: failed to assign all query results to SavedFile slice")
	}

	return o, nil
}

// Count returns the count of all SavedFile records in the query.
func (q savedFileQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to count saved_file rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q savedFileQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "sqlboiler: failed to check if saved_file exists")
	}

	return count > 0, nil
}

// EncryptedFile pointed to by the foreign key.
func (o *SavedFile) EncryptedFile(mods ...qm.QueryMod) encryptedFileQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.EncryptedFileID),
	}

	queryMods = append(queryMods, mods...)

	query := EncryptedFiles(queryMods...)
	queries.SetFrom(query.Query, "\"encrypted_file\"")

	return query
}

// LoadEncryptedFile allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (savedFileL) LoadEncryptedFile(ctx context.Context, e boil.ContextExecutor, singular bool, maybeSavedFile interface{}, mods queries.Applicator) error {
	var slice []*SavedFile
	var object *SavedFile

	if singular {
		object = maybeSavedFile.(*SavedFile)
	} else {
		slice = *maybeSavedFile.(*[]*SavedFile)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &savedFileR{}
		}
		args = append(args, object.EncryptedFileID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &savedFileR{}
			}

			for _, a := range args {
				if a == obj.EncryptedFileID {
					continue Outer
				}
			}

			args = append(args, obj.EncryptedFileID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(qm.From(`encrypted_file`), qm.WhereIn(`encrypted_file.id in ?`, args...))
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load EncryptedFile")
	}

	var resultSlice []*EncryptedFile
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice EncryptedFile")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for encrypted_file")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for encrypted_file")
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.EncryptedFile = foreign
		if foreign.R == nil {
			foreign.R = &encryptedFileR{}
		}
		foreign.R.SavedFiles = append(foreign.R.SavedFiles, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.EncryptedFileID == foreign.ID {
				local.R.EncryptedFile = foreign
				if foreign.R == nil {
					foreign.R = &encryptedFileR{}
				}
				foreign.R.SavedFiles = append(foreign.R.SavedFiles, local)
				break
			}
		}
	}

	return nil
}

// SetEncryptedFile of the savedFile to the related item.
// Sets o.R.EncryptedFile to related.
// Adds o to related.R.SavedFiles.
func (o *SavedFile) SetEncryptedFile(ctx context.Context, exec boil.ContextExecutor, insert bool, related *EncryptedFile) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"saved_file\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"encrypted_file_id"}),
		strmangle.WhereClause("\"", "\"", 2, savedFilePrimaryKeyColumns),
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

	o.EncryptedFileID = related.ID
	if o.R == nil {
		o.R = &savedFileR{
			EncryptedFile: related,
		}
	} else {
		o.R.EncryptedFile = related
	}

	if related.R == nil {
		related.R = &encryptedFileR{
			SavedFiles: SavedFileSlice{o},
		}
	} else {
		related.R.SavedFiles = append(related.R.SavedFiles, o)
	}

	return nil
}

// SavedFiles retrieves all the records using an executor.
func SavedFiles(mods ...qm.QueryMod) savedFileQuery {
	mods = append(mods, qm.From("\"saved_file\""))
	return savedFileQuery{NewQuery(mods...)}
}

// FindSavedFile retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindSavedFile(ctx context.Context, exec boil.ContextExecutor, iD string, selectCols ...string) (*SavedFile, error) {
	savedFileObj := &SavedFile{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"saved_file\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, savedFileObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "sqlboiler: unable to select from saved_file")
	}

	return savedFileObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *SavedFile) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("sqlboiler: no saved_file provided for insertion")
	}

	var err error
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	nzDefaults := queries.NonZeroDefaultSet(savedFileColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	savedFileInsertCacheMut.RLock()
	cache, cached := savedFileInsertCache[key]
	savedFileInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			savedFileAllColumns,
			savedFileColumnsWithDefault,
			savedFileColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(savedFileType, savedFileMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(savedFileType, savedFileMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"saved_file\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"saved_file\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "sqlboiler: unable to insert into saved_file")
	}

	if !cached {
		savedFileInsertCacheMut.Lock()
		savedFileInsertCache[key] = cache
		savedFileInsertCacheMut.Unlock()
	}

	return nil
}

// Update uses an executor to update the SavedFile.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *SavedFile) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	key := makeCacheKey(columns, nil)
	savedFileUpdateCacheMut.RLock()
	cache, cached := savedFileUpdateCache[key]
	savedFileUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			savedFileAllColumns,
			savedFilePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("sqlboiler: unable to update saved_file, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"saved_file\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, savedFilePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(savedFileType, savedFileMapping, append(wl, savedFilePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "sqlboiler: unable to update saved_file row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to get rows affected by update for saved_file")
	}

	if !cached {
		savedFileUpdateCacheMut.Lock()
		savedFileUpdateCache[key] = cache
		savedFileUpdateCacheMut.Unlock()
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values.
func (q savedFileQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to update all for saved_file")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to retrieve rows affected for saved_file")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o SavedFileSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), savedFilePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"saved_file\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, savedFilePrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to update all in savedFile slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to retrieve rows affected all in update all savedFile")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *SavedFile) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("sqlboiler: no saved_file provided for upsert")
	}
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	nzDefaults := queries.NonZeroDefaultSet(savedFileColumnsWithDefault, o)

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

	savedFileUpsertCacheMut.RLock()
	cache, cached := savedFileUpsertCache[key]
	savedFileUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			savedFileAllColumns,
			savedFileColumnsWithDefault,
			savedFileColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			savedFileAllColumns,
			savedFilePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("sqlboiler: unable to upsert saved_file, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(savedFilePrimaryKeyColumns))
			copy(conflict, savedFilePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"saved_file\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(savedFileType, savedFileMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(savedFileType, savedFileMapping, ret)
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

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, vals)
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
		return errors.Wrap(err, "sqlboiler: unable to upsert saved_file")
	}

	if !cached {
		savedFileUpsertCacheMut.Lock()
		savedFileUpsertCache[key] = cache
		savedFileUpsertCacheMut.Unlock()
	}

	return nil
}

// Delete deletes a single SavedFile record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *SavedFile) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("sqlboiler: no SavedFile provided for delete")
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), savedFilePrimaryKeyMapping)
	sql := "DELETE FROM \"saved_file\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to delete from saved_file")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to get rows affected by delete for saved_file")
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q savedFileQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("sqlboiler: no savedFileQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to delete all from saved_file")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to get rows affected by deleteall for saved_file")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o SavedFileSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), savedFilePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"saved_file\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, savedFilePrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to delete all from savedFile slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to get rows affected by deleteall for saved_file")
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *SavedFile) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindSavedFile(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *SavedFileSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := SavedFileSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), savedFilePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"saved_file\".* FROM \"saved_file\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, savedFilePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "sqlboiler: unable to reload all in SavedFileSlice")
	}

	*o = slice

	return nil
}

// SavedFileExists checks if the SavedFile row exists.
func SavedFileExists(ctx context.Context, exec boil.ContextExecutor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"saved_file\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "sqlboiler: unable to check if saved_file exists")
	}

	return exists, nil
}
