package events

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/volatiletech/sqlboiler/boil"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
)

func MustBoxExists(ctx context.Context, exec boil.ContextExecutor, boxID string) error {
	fmt.Println("le box id", boxID)
	_, err := sqlboiler.Events(
		sqlboiler.EventWhere.BoxID.EQ(boxID),
		sqlboiler.EventWhere.Type.EQ("create"),
	).One(ctx, exec)
	if err != nil && err == sql.ErrNoRows {
		return merror.NotFound().Detail("box_id", merror.DVNotFound)
	}
	return err
}
