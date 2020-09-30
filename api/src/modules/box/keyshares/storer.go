package keyshares

import (
	"context"
	"database/sql"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"github.com/volatiletech/sqlboiler/v4/boil"
)

func Create(
	ctx context.Context, exec boil.ContextExecutor,
	invitHash, share, boxID, creatorID string,
) error {
	ks := BoxKeyShare{invitHash, share, boxID, creatorID}
	return ks.toSQLBoiler().Insert(ctx, exec, boil.Infer())
}

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

func GetLastForBoxID(
	ctx context.Context, exec boil.ContextExecutor,
	boxID string,
) (*BoxKeyShare, error) {
	mods := []qm.QueryMod{
		sqlboiler.BoxKeyShareWhere.BoxID.EQ(boxID),
		qm.OrderBy("created_at DESC"),
	}
	record, err := sqlboiler.BoxKeyShares(mods...).One(ctx, exec)
	if err != nil {
		return nil, err
	}

	keyShare := fromSQLBoiler(record)

	return &keyShare, nil
}

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
