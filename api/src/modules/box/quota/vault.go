package quota

import (
	"context"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
)

type VaultUsedSpace struct {
	Value      int64 `json:"value" boil:"total"`
}

func GetVault(ctx context.Context, exec boil.ContextExecutor, id string) (*VaultUsedSpace, error) {
	mods := []qm.QueryMod{
		qm.Select("COALESCE(SUM(encrypted_file.size), 0) as total"),
		qm.InnerJoin("encrypted_file ON encrypted_file.id = saved_file.encrypted_file_id "),
		sqlboiler.SavedFileWhere.IdentityID.EQ(id),
	}
	vaultUserSpace := VaultUsedSpace{}
	err := sqlboiler.SavedFiles(mods...).Bind(ctx, exec, &vaultUserSpace)
	if err != nil {
		return nil, err
	}

	return &vaultUserSpace, nil
}