package keyshares

import (
	"context"
	"database/sql"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"gitlab.misakey.dev/misakey/backend/api/src/box/repositories/sqlboiler"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"github.com/volatiletech/sqlboiler/v4/boil"
)

// Create a new keyshare
func Create(
	ctx context.Context, exec boil.ContextExecutor,
	invitHash, share, encShare, boxID, creatorID string,
) error {
	ks := BoxKeyShare{invitHash, share, boxID, null.StringFrom(encShare), creatorID}
	return ks.toSQLBoiler().Insert(ctx, exec, boil.Infer())
}

// Get a keyshare from its invitHash
func Get(
	ctx context.Context, exec boil.ContextExecutor,
	invitHash string,
) (ret BoxKeyShare, err error) {
	record, err := sqlboiler.FindBoxKeyShare(ctx, exec, invitHash)
	if err == sql.ErrNoRows {
		return ret, merror.NotFound().Detail("other_share_hash", merror.DVNotFound)
	}
	if err != nil {
		return ret, err
	}
	return fromSQLBoiler(record), nil
}

// GetLastForBoxID returns the last keyShare of a given box
func GetLastForBoxID(
	ctx context.Context, exec boil.ContextExecutor,
	boxID string,
) (*BoxKeyShare, error) {
	mods := []qm.QueryMod{
		sqlboiler.BoxKeyShareWhere.BoxID.EQ(boxID),
		qm.OrderBy(sqlboiler.BoxKeyShareColumns.CreatedAt + " DESC"),
	}
	record, err := sqlboiler.BoxKeyShares(mods...).One(ctx, exec)
	if err != nil {
		return nil, err
	}

	keyShare := fromSQLBoiler(record)

	return &keyShare, nil
}

// EmptyAll keyshares for a given box
func EmptyAll(
	ctx context.Context, exec boil.ContextExecutor,
	boxID string,
) error {
	mods := []qm.QueryMod{
		sqlboiler.BoxKeyShareWhere.BoxID.EQ(boxID),
	}
	// ignore the zero affected rows case since the deletion is a success
	// if no record exist (it empties all)
	_, err := sqlboiler.BoxKeyShares(mods...).DeleteAll(ctx, exec)
	return err
}
