package domain

import "time"

type UsedCoupon struct {
	ID            int    `json:"id"`
	IdentityID    string    `json:"identity_id"`
	Value         string    `json:"value"`
	CreatedAt     time.Time `json:"created_ad"`
}
