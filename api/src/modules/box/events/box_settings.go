package events

import (
	"context"
	"database/sql"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
)

type BoxSetting struct {
	IdentityID string `json:"identity_id"`
	BoxID      string `json:"box_id"`
	Muted      bool   `json:"muted"`
}

type BoxSettingFilters struct {
	BoxIDs     []string
	IdentityID string
}

func UpdateBoxSetting(ctx context.Context, exec boil.ContextExecutor, boxSetting BoxSetting) error {
	toUpsert := sqlboiler.BoxSetting{
		IdentityID: boxSetting.IdentityID,
		BoxID:      boxSetting.BoxID,
		Muted:      boxSetting.Muted,
	}
	return toUpsert.Upsert(ctx, exec, true, []string{sqlboiler.BoxSettingColumns.BoxID, sqlboiler.BoxSettingColumns.IdentityID}, boil.Infer(), boil.Infer())
}

func GetBoxSetting(ctx context.Context, exec boil.ContextExecutor, identityID, boxID string) (*BoxSetting, error) {
	mods := []qm.QueryMod{
		sqlboiler.BoxSettingWhere.BoxID.EQ(boxID),
		sqlboiler.BoxSettingWhere.IdentityID.EQ(identityID),
	}

	boxSetting, err := sqlboiler.BoxSettings(mods...).One(ctx, exec)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err != nil && err == sql.ErrNoRows {
		// we return a default value
		return GetDefaultBoxSetting(identityID, boxID), nil
	}

	return &BoxSetting{
		IdentityID: boxSetting.IdentityID,
		BoxID:      boxSetting.BoxID,
		Muted:      boxSetting.Muted,
	}, nil
}

func GetDefaultBoxSetting(identityID, boxID string) *BoxSetting {
	return &BoxSetting{
		IdentityID: identityID,
		BoxID:      boxID,
		Muted:      false,
	}
}

func ListBoxSettings(ctx context.Context, exec boil.ContextExecutor, filters BoxSettingFilters) ([]*BoxSetting, error) {
	mods := []qm.QueryMod{}

	if filters.IdentityID != "" {
		sqlboiler.BoxSettingWhere.IdentityID.EQ(filters.IdentityID)
	}

	if len(filters.BoxIDs) != 0 {
		sqlboiler.BoxSettingWhere.BoxID.IN(filters.BoxIDs)
	}

	dbBoxSettings, err := sqlboiler.BoxSettings(mods...).All(ctx, exec)
	if err != nil {
		return nil, err
	}

	boxSettings := make([]*BoxSetting, len(dbBoxSettings))
	for idx, boxSetting := range dbBoxSettings {
		boxSettings[idx] = &BoxSetting{
			IdentityID: boxSetting.IdentityID,
			BoxID:      boxSetting.BoxID,
			Muted:      boxSetting.Muted,
		}
	}

	return boxSettings, nil

}
