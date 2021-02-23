package datatag

import (
	"context"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories/sqlboiler"
)

func Get(ctx context.Context, exec boil.ContextExecutor, id string) (*sqlboiler.Datatag, error) {
	return sqlboiler.FindDatatag(ctx, exec, id)
}
