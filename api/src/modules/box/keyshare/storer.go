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
	invitHash, share string,
) error {
	ks := KeyShare{invitHash, share}
	return ks.toSqlBoiler().Insert(ctx, exec, boil.Infer())
}

func Get(
	ctx context.Context, exec boil.ContextExecutor,
	invitHash string,
) (ret KeyShare, err error) {
	record, err := sqlboiler.FindKeyShare(ctx, exec, invitHash)
	if err == sql.ErrNoRows {
		return ret, merror.NotFound().Detail("invitation_hash", merror.DVNotFound)
	}
	if err != nil {
		return ret, err
	}
	return fromSqlBoiler(record), nil
}
