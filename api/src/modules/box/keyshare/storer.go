package keyshare

import (
	"context"
	"database/sql"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"github.com/volatiletech/sqlboiler/boil"
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
