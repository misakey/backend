package identity

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories/sqlboiler"
)

// Identifier ...
type Identifier struct {
	ID    string         `json:"id"`
	Value string         `json:"value"`
	Kind  IdentifierKind `json:"kind"`
}

// IdentifierKind ...
type IdentifierKind string

const (
	// EmailIdentifier ...
	EmailIdentifier IdentifierKind = "email"
)

// GetIdentifier ...
func GetIdentifier(ctx context.Context, exec boil.ContextExecutor, id string) (Identifier, error) {
	mods := []qm.QueryMod{
		sqlboiler.IdentifierWhere.ID.EQ(id),
	}

	record, err := sqlboiler.Identifiers(mods...).One(ctx, exec)
	if err == sql.ErrNoRows {
		return Identifier{}, merror.NotFound().Detail("id", merror.DVNotFound)
	}
	if err != nil {
		return Identifier{}, err
	}

	return Identifier{
		ID:    record.ID,
		Value: record.Value,
		Kind:  IdentifierKind(record.Kind),
	}, nil
}

// RequireIdentifier ...
func RequireIdentifier(ctx context.Context, exec boil.ContextExecutor, identifier *Identifier) error {
	mods := []qm.QueryMod{
		sqlboiler.IdentifierWhere.Kind.EQ(string(identifier.Kind)),
		sqlboiler.IdentifierWhere.Value.EQ(strings.ToLower(identifier.Value)),
	}

	existingRecord, err := sqlboiler.Identifiers(mods...).One(ctx, exec)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// on found, bind it and return
	if err == nil {
		identifier.ID = existingRecord.ID
		identifier.Value = existingRecord.Value
		identifier.Kind = IdentifierKind(existingRecord.Kind)
		return nil
	}

	// otherwise at this point we know we got a not found error so we create the identifier
	return createIdentifier(ctx, exec, identifier)
}

// GetIdentifierByKindValue ...
func GetIdentifierByKindValue(ctx context.Context, exec boil.ContextExecutor, identifier Identifier) (Identifier, error) {
	mods := []qm.QueryMod{
		sqlboiler.IdentifierWhere.Kind.EQ(string(identifier.Kind)),
		sqlboiler.IdentifierWhere.Value.EQ(strings.ToLower(identifier.Value)),
	}

	record, err := sqlboiler.Identifiers(mods...).One(ctx, exec)
	if err == sql.ErrNoRows {
		return Identifier{}, merror.NotFound().Detail("kind", merror.DVNotFound).Detail("value", merror.DVNotFound)
	}
	if err != nil {
		return Identifier{}, err
	}

	return Identifier{
		ID:    record.ID,
		Value: record.Value,
		Kind:  IdentifierKind(record.Kind),
	}, nil
}

func createIdentifier(ctx context.Context, exec boil.ContextExecutor, identifier *Identifier) error {
	// generate new UUID for new record
	id, err := uuid.NewRandom()
	if err != nil {
		return merror.Transform(err).Describe("could not generate uuid v4")
	}
	identifier.ID = id.String()

	// convert domain to sql model
	sqlIdentifier := sqlboiler.Identifier{
		ID:    identifier.ID,
		Value: strings.ToLower(identifier.Value),
		Kind:  string(identifier.Kind),
	}

	return sqlIdentifier.Insert(ctx, exec, boil.Infer())
}
