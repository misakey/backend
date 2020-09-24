package repositories

import (
	"context"
	"database/sql"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/repositories/sqlboiler"
)

type UsedCouponSQLBoiler struct {
	db *sql.DB
}

func NewUsedCouponSQLBoiler(db *sql.DB) *UsedCouponSQLBoiler {
	return &UsedCouponSQLBoiler{
		db: db,
	}
}

func (repo UsedCouponSQLBoiler) toSQLBoiler(domModel *domain.UsedCoupon) *sqlboiler.UsedCoupon {
	return &sqlboiler.UsedCoupon{
		ID:         domModel.ID,
		IdentityID: domModel.IdentityID,
		Value:      domModel.Value,
	}
}

func (repo *UsedCouponSQLBoiler) Insert(ctx context.Context, UsedCoupon domain.UsedCoupon) error {
	// convert domain to sql model
	sqlUsedCoupon := repo.toSQLBoiler(&UsedCoupon)
	return sqlUsedCoupon.Insert(ctx, repo.db, boil.Infer())
}
