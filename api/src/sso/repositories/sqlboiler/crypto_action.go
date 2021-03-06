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
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// CryptoAction is an object representing the database table.
type CryptoAction struct {
	ID                  string      `boil:"id" json:"id" toml:"id" yaml:"id"`
	AccountID           string      `boil:"account_id" json:"account_id" toml:"account_id" yaml:"account_id"`
	SenderIdentityID    null.String `boil:"sender_identity_id" json:"sender_identity_id,omitempty" toml:"sender_identity_id" yaml:"sender_identity_id,omitempty"`
	Type                string      `boil:"type" json:"type" toml:"type" yaml:"type"`
	BoxID               null.String `boil:"box_id" json:"box_id,omitempty" toml:"box_id" yaml:"box_id,omitempty"`
	EncryptionPublicKey string      `boil:"encryption_public_key" json:"encryption_public_key" toml:"encryption_public_key" yaml:"encryption_public_key"`
	Encrypted           string      `boil:"encrypted" json:"encrypted" toml:"encrypted" yaml:"encrypted"`
	CreatedAt           time.Time   `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *cryptoActionR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L cryptoActionL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var CryptoActionColumns = struct {
	ID                  string
	AccountID           string
	SenderIdentityID    string
	Type                string
	BoxID               string
	EncryptionPublicKey string
	Encrypted           string
	CreatedAt           string
}{
	ID:                  "id",
	AccountID:           "account_id",
	SenderIdentityID:    "sender_identity_id",
	Type:                "type",
	BoxID:               "box_id",
	EncryptionPublicKey: "encryption_public_key",
	Encrypted:           "encrypted",
	CreatedAt:           "created_at",
}

// Generated where

var CryptoActionWhere = struct {
	ID                  whereHelperstring
	AccountID           whereHelperstring
	SenderIdentityID    whereHelpernull_String
	Type                whereHelperstring
	BoxID               whereHelpernull_String
	EncryptionPublicKey whereHelperstring
	Encrypted           whereHelperstring
	CreatedAt           whereHelpertime_Time
}{
	ID:                  whereHelperstring{field: "\"crypto_action\".\"id\""},
	AccountID:           whereHelperstring{field: "\"crypto_action\".\"account_id\""},
	SenderIdentityID:    whereHelpernull_String{field: "\"crypto_action\".\"sender_identity_id\""},
	Type:                whereHelperstring{field: "\"crypto_action\".\"type\""},
	BoxID:               whereHelpernull_String{field: "\"crypto_action\".\"box_id\""},
	EncryptionPublicKey: whereHelperstring{field: "\"crypto_action\".\"encryption_public_key\""},
	Encrypted:           whereHelperstring{field: "\"crypto_action\".\"encrypted\""},
	CreatedAt:           whereHelpertime_Time{field: "\"crypto_action\".\"created_at\""},
}

// CryptoActionRels is where relationship names are stored.
var CryptoActionRels = struct {
	Account        string
	SenderIdentity string
}{
	Account:        "Account",
	SenderIdentity: "SenderIdentity",
}

// cryptoActionR is where relationships are stored.
type cryptoActionR struct {
	Account        *Account  `boil:"Account" json:"Account" toml:"Account" yaml:"Account"`
	SenderIdentity *Identity `boil:"SenderIdentity" json:"SenderIdentity" toml:"SenderIdentity" yaml:"SenderIdentity"`
}

// NewStruct creates a new relationship struct
func (*cryptoActionR) NewStruct() *cryptoActionR {
	return &cryptoActionR{}
}

// cryptoActionL is where Load methods for each relationship are stored.
type cryptoActionL struct{}

var (
	cryptoActionAllColumns            = []string{"id", "account_id", "sender_identity_id", "type", "box_id", "encryption_public_key", "encrypted", "created_at"}
	cryptoActionColumnsWithoutDefault = []string{"id", "account_id", "sender_identity_id", "type", "box_id", "encryption_public_key", "encrypted"}
	cryptoActionColumnsWithDefault    = []string{"created_at"}
	cryptoActionPrimaryKeyColumns     = []string{"id"}
)

type (
	// CryptoActionSlice is an alias for a slice of pointers to CryptoAction.
	// This should generally be used opposed to []CryptoAction.
	CryptoActionSlice []*CryptoAction

	cryptoActionQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	cryptoActionType                 = reflect.TypeOf(&CryptoAction{})
	cryptoActionMapping              = queries.MakeStructMapping(cryptoActionType)
	cryptoActionPrimaryKeyMapping, _ = queries.BindMapping(cryptoActionType, cryptoActionMapping, cryptoActionPrimaryKeyColumns)
	cryptoActionInsertCacheMut       sync.RWMutex
	cryptoActionInsertCache          = make(map[string]insertCache)
	cryptoActionUpdateCacheMut       sync.RWMutex
	cryptoActionUpdateCache          = make(map[string]updateCache)
	cryptoActionUpsertCacheMut       sync.RWMutex
	cryptoActionUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

// One returns a single cryptoAction record from the query.
func (q cryptoActionQuery) One(ctx context.Context, exec boil.ContextExecutor) (*CryptoAction, error) {
	o := &CryptoAction{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "sqlboiler: failed to execute a one query for crypto_action")
	}

	return o, nil
}

// All returns all CryptoAction records from the query.
func (q cryptoActionQuery) All(ctx context.Context, exec boil.ContextExecutor) (CryptoActionSlice, error) {
	var o []*CryptoAction

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "sqlboiler: failed to assign all query results to CryptoAction slice")
	}

	return o, nil
}

// Count returns the count of all CryptoAction records in the query.
func (q cryptoActionQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to count crypto_action rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q cryptoActionQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "sqlboiler: failed to check if crypto_action exists")
	}

	return count > 0, nil
}

// Account pointed to by the foreign key.
func (o *CryptoAction) Account(mods ...qm.QueryMod) accountQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.AccountID),
	}

	queryMods = append(queryMods, mods...)

	query := Accounts(queryMods...)
	queries.SetFrom(query.Query, "\"account\"")

	return query
}

// SenderIdentity pointed to by the foreign key.
func (o *CryptoAction) SenderIdentity(mods ...qm.QueryMod) identityQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.SenderIdentityID),
	}

	queryMods = append(queryMods, mods...)

	query := Identities(queryMods...)
	queries.SetFrom(query.Query, "\"identity\"")

	return query
}

// LoadAccount allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (cryptoActionL) LoadAccount(ctx context.Context, e boil.ContextExecutor, singular bool, maybeCryptoAction interface{}, mods queries.Applicator) error {
	var slice []*CryptoAction
	var object *CryptoAction

	if singular {
		object = maybeCryptoAction.(*CryptoAction)
	} else {
		slice = *maybeCryptoAction.(*[]*CryptoAction)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &cryptoActionR{}
		}
		args = append(args, object.AccountID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &cryptoActionR{}
			}

			for _, a := range args {
				if a == obj.AccountID {
					continue Outer
				}
			}

			args = append(args, obj.AccountID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`account`),
		qm.WhereIn(`account.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Account")
	}

	var resultSlice []*Account
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Account")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for account")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for account")
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.Account = foreign
		if foreign.R == nil {
			foreign.R = &accountR{}
		}
		foreign.R.CryptoActions = append(foreign.R.CryptoActions, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.AccountID == foreign.ID {
				local.R.Account = foreign
				if foreign.R == nil {
					foreign.R = &accountR{}
				}
				foreign.R.CryptoActions = append(foreign.R.CryptoActions, local)
				break
			}
		}
	}

	return nil
}

// LoadSenderIdentity allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (cryptoActionL) LoadSenderIdentity(ctx context.Context, e boil.ContextExecutor, singular bool, maybeCryptoAction interface{}, mods queries.Applicator) error {
	var slice []*CryptoAction
	var object *CryptoAction

	if singular {
		object = maybeCryptoAction.(*CryptoAction)
	} else {
		slice = *maybeCryptoAction.(*[]*CryptoAction)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &cryptoActionR{}
		}
		if !queries.IsNil(object.SenderIdentityID) {
			args = append(args, object.SenderIdentityID)
		}

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &cryptoActionR{}
			}

			for _, a := range args {
				if queries.Equal(a, obj.SenderIdentityID) {
					continue Outer
				}
			}

			if !queries.IsNil(obj.SenderIdentityID) {
				args = append(args, obj.SenderIdentityID)
			}

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
		object.R.SenderIdentity = foreign
		if foreign.R == nil {
			foreign.R = &identityR{}
		}
		foreign.R.SenderIdentityCryptoActions = append(foreign.R.SenderIdentityCryptoActions, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if queries.Equal(local.SenderIdentityID, foreign.ID) {
				local.R.SenderIdentity = foreign
				if foreign.R == nil {
					foreign.R = &identityR{}
				}
				foreign.R.SenderIdentityCryptoActions = append(foreign.R.SenderIdentityCryptoActions, local)
				break
			}
		}
	}

	return nil
}

// SetAccount of the cryptoAction to the related item.
// Sets o.R.Account to related.
// Adds o to related.R.CryptoActions.
func (o *CryptoAction) SetAccount(ctx context.Context, exec boil.ContextExecutor, insert bool, related *Account) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"crypto_action\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"account_id"}),
		strmangle.WhereClause("\"", "\"", 2, cryptoActionPrimaryKeyColumns),
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

	o.AccountID = related.ID
	if o.R == nil {
		o.R = &cryptoActionR{
			Account: related,
		}
	} else {
		o.R.Account = related
	}

	if related.R == nil {
		related.R = &accountR{
			CryptoActions: CryptoActionSlice{o},
		}
	} else {
		related.R.CryptoActions = append(related.R.CryptoActions, o)
	}

	return nil
}

// SetSenderIdentity of the cryptoAction to the related item.
// Sets o.R.SenderIdentity to related.
// Adds o to related.R.SenderIdentityCryptoActions.
func (o *CryptoAction) SetSenderIdentity(ctx context.Context, exec boil.ContextExecutor, insert bool, related *Identity) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"crypto_action\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"sender_identity_id"}),
		strmangle.WhereClause("\"", "\"", 2, cryptoActionPrimaryKeyColumns),
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

	queries.Assign(&o.SenderIdentityID, related.ID)
	if o.R == nil {
		o.R = &cryptoActionR{
			SenderIdentity: related,
		}
	} else {
		o.R.SenderIdentity = related
	}

	if related.R == nil {
		related.R = &identityR{
			SenderIdentityCryptoActions: CryptoActionSlice{o},
		}
	} else {
		related.R.SenderIdentityCryptoActions = append(related.R.SenderIdentityCryptoActions, o)
	}

	return nil
}

// RemoveSenderIdentity relationship.
// Sets o.R.SenderIdentity to nil.
// Removes o from all passed in related items' relationships struct (Optional).
func (o *CryptoAction) RemoveSenderIdentity(ctx context.Context, exec boil.ContextExecutor, related *Identity) error {
	var err error

	queries.SetScanner(&o.SenderIdentityID, nil)
	if _, err = o.Update(ctx, exec, boil.Whitelist("sender_identity_id")); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	if o.R != nil {
		o.R.SenderIdentity = nil
	}
	if related == nil || related.R == nil {
		return nil
	}

	for i, ri := range related.R.SenderIdentityCryptoActions {
		if queries.Equal(o.SenderIdentityID, ri.SenderIdentityID) {
			continue
		}

		ln := len(related.R.SenderIdentityCryptoActions)
		if ln > 1 && i < ln-1 {
			related.R.SenderIdentityCryptoActions[i] = related.R.SenderIdentityCryptoActions[ln-1]
		}
		related.R.SenderIdentityCryptoActions = related.R.SenderIdentityCryptoActions[:ln-1]
		break
	}
	return nil
}

// CryptoActions retrieves all the records using an executor.
func CryptoActions(mods ...qm.QueryMod) cryptoActionQuery {
	mods = append(mods, qm.From("\"crypto_action\""))
	return cryptoActionQuery{NewQuery(mods...)}
}

// FindCryptoAction retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindCryptoAction(ctx context.Context, exec boil.ContextExecutor, iD string, selectCols ...string) (*CryptoAction, error) {
	cryptoActionObj := &CryptoAction{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"crypto_action\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, cryptoActionObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "sqlboiler: unable to select from crypto_action")
	}

	return cryptoActionObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *CryptoAction) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("sqlboiler: no crypto_action provided for insertion")
	}

	var err error
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	nzDefaults := queries.NonZeroDefaultSet(cryptoActionColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	cryptoActionInsertCacheMut.RLock()
	cache, cached := cryptoActionInsertCache[key]
	cryptoActionInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			cryptoActionAllColumns,
			cryptoActionColumnsWithDefault,
			cryptoActionColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(cryptoActionType, cryptoActionMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(cryptoActionType, cryptoActionMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"crypto_action\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"crypto_action\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "sqlboiler: unable to insert into crypto_action")
	}

	if !cached {
		cryptoActionInsertCacheMut.Lock()
		cryptoActionInsertCache[key] = cache
		cryptoActionInsertCacheMut.Unlock()
	}

	return nil
}

// Update uses an executor to update the CryptoAction.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *CryptoAction) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	key := makeCacheKey(columns, nil)
	cryptoActionUpdateCacheMut.RLock()
	cache, cached := cryptoActionUpdateCache[key]
	cryptoActionUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			cryptoActionAllColumns,
			cryptoActionPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("sqlboiler: unable to update crypto_action, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"crypto_action\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, cryptoActionPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(cryptoActionType, cryptoActionMapping, append(wl, cryptoActionPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "sqlboiler: unable to update crypto_action row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to get rows affected by update for crypto_action")
	}

	if !cached {
		cryptoActionUpdateCacheMut.Lock()
		cryptoActionUpdateCache[key] = cache
		cryptoActionUpdateCacheMut.Unlock()
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values.
func (q cryptoActionQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to update all for crypto_action")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to retrieve rows affected for crypto_action")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o CryptoActionSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), cryptoActionPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"crypto_action\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, cryptoActionPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to update all in cryptoAction slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to retrieve rows affected all in update all cryptoAction")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *CryptoAction) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("sqlboiler: no crypto_action provided for upsert")
	}
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	nzDefaults := queries.NonZeroDefaultSet(cryptoActionColumnsWithDefault, o)

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

	cryptoActionUpsertCacheMut.RLock()
	cache, cached := cryptoActionUpsertCache[key]
	cryptoActionUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			cryptoActionAllColumns,
			cryptoActionColumnsWithDefault,
			cryptoActionColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			cryptoActionAllColumns,
			cryptoActionPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("sqlboiler: unable to upsert crypto_action, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(cryptoActionPrimaryKeyColumns))
			copy(conflict, cryptoActionPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"crypto_action\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(cryptoActionType, cryptoActionMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(cryptoActionType, cryptoActionMapping, ret)
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
		return errors.Wrap(err, "sqlboiler: unable to upsert crypto_action")
	}

	if !cached {
		cryptoActionUpsertCacheMut.Lock()
		cryptoActionUpsertCache[key] = cache
		cryptoActionUpsertCacheMut.Unlock()
	}

	return nil
}

// Delete deletes a single CryptoAction record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *CryptoAction) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("sqlboiler: no CryptoAction provided for delete")
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cryptoActionPrimaryKeyMapping)
	sql := "DELETE FROM \"crypto_action\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to delete from crypto_action")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to get rows affected by delete for crypto_action")
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q cryptoActionQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("sqlboiler: no cryptoActionQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to delete all from crypto_action")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to get rows affected by deleteall for crypto_action")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o CryptoActionSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), cryptoActionPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"crypto_action\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, cryptoActionPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: unable to delete all from cryptoAction slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboiler: failed to get rows affected by deleteall for crypto_action")
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *CryptoAction) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindCryptoAction(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *CryptoActionSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := CryptoActionSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), cryptoActionPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"crypto_action\".* FROM \"crypto_action\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, cryptoActionPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "sqlboiler: unable to reload all in CryptoActionSlice")
	}

	*o = slice

	return nil
}

// CryptoActionExists checks if the CryptoAction row exists.
func CryptoActionExists(ctx context.Context, exec boil.ContextExecutor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"crypto_action\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "sqlboiler: unable to check if crypto_action exists")
	}

	return exists, nil
}
