package datatag

import (
	"context"
	"database/sql"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories/sqlboiler"
)

type Filters struct {
	OrganizationID string
	IDs            []string
}

func Get(ctx context.Context, exec boil.ContextExecutor, id string) (*sqlboiler.Datatag, error) {
	datatag, err := sqlboiler.FindDatatag(ctx, exec, id)
	if err != nil && err == sql.ErrNoRows {
		return nil, merr.NotFound()
	}

	return datatag, nil
}

func List(ctx context.Context, exec boil.ContextExecutor, filters Filters) ([]*sqlboiler.Datatag, error) {
	mods := []qm.QueryMod{}

	if filters.OrganizationID != "" {
		mods = append(mods, sqlboiler.DatatagWhere.OrganizationID.EQ(filters.OrganizationID))
	}

	if len(filters.IDs) != 0 {
		mods = append(mods, sqlboiler.DatatagWhere.ID.IN(filters.IDs))
	}

	datatags, err := sqlboiler.Datatags(mods...).All(ctx, exec)
	if err != nil {
		return nil, err
	}
	if datatags == nil {
		return []*sqlboiler.Datatag{}, nil
	}

	return datatags, nil
}

func CheckExistencyAndOrg(ctx context.Context, exec boil.ContextExecutor, datatagID, orgID string) error {
	datatag, err := Get(ctx, exec, datatagID)
	if err != nil && merr.IsANotFound(err) {
		return merr.From(err).Desc("getting datatag").Add("datatag_id", merr.DVNotFound)
	}
	if err != nil {
		return merr.From(err).Desc("getting datatag")
	}
	if datatag.OrganizationID != orgID {
		return merr.Forbidden().Desc("datatag must belong to organization").Add("datatag_id", merr.DVForbidden)
	}

	return nil
}
