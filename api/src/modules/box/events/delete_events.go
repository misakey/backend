package events

import (
	"context"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

func DeleteAllForBox(
	ctx context.Context,
	exec boil.ContextExecutor,
	boxID string,
) error {
	mods := []qm.QueryMod{
		sqlboiler.EventWhere.BoxID.EQ(boxID),
	}
	rowsAff, err := sqlboiler.Events(mods...).DeleteAll(ctx, exec)
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return merror.NotFound().Detail("box_id", merror.DVNotFound)
	}
	return nil
}
