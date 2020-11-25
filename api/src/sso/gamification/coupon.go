package gamification

import (
	"context"
	"time"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories/sqlboiler"
)

//
// models
//
type UsedCoupon struct {
	ID         int       `json:"id"`
	IdentityID string    `json:"identity_id"`
	Value      string    `json:"value"`
	CreatedAt  time.Time `json:"created_ad"`
}

func (uc UsedCoupon) toSQLBoiler() *sqlboiler.UsedCoupon {
	return &sqlboiler.UsedCoupon{
		ID:         uc.ID,
		IdentityID: uc.IdentityID,
		Value:      uc.Value,
	}
}

//
// functions
//

func UseCoupon(ctx context.Context, exec boil.ContextExecutor, usedCoupon UsedCoupon) error {
	return usedCoupon.toSQLBoiler().Insert(ctx, exec, boil.Infer())
}
