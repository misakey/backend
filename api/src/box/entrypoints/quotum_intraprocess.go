package entrypoints

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/box/application"
)

// QuotaIntraprocessInterface ...
type QuotaIntraprocessInterface interface {
	CreateBase(ctx context.Context, identityID string) (interface{}, error)
}

// QuotaIntraprocess ...
type QuotaIntraprocess struct {
	service application.BoxApplication
}

// NewQuotumIntraprocess ...
func NewQuotumIntraprocess(boxService application.BoxApplication) QuotaIntraprocess {
	return QuotaIntraprocess{
		service: boxService,
	}
}

// CreateBase ...
func (qii *QuotaIntraprocess) CreateBase(ctx context.Context, identityID string) (interface{}, error) {
	req := application.CreateQuotumRequest{
		IdentityID: identityID,
		Value:      104857600,
		Origin:     "base",
	}
	return qii.service.CreateQuotum(ctx, &req)
}
