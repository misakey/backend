package coupon

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type usedCouponRepo interface {
	Insert(context.Context, domain.UsedCoupon) error
}

type UsedCouponService struct {
	usedCoupons usedCouponRepo
}

func NewUsedCouponService(
	usedCouponRepo usedCouponRepo,
) UsedCouponService {
	return UsedCouponService{
		usedCoupons: usedCouponRepo,
	}
}

func (bkr UsedCouponService) CreateUsedCoupon(ctx context.Context, usedCoupon domain.UsedCoupon) error {
	return bkr.usedCoupons.Insert(ctx, usedCoupon)
}
