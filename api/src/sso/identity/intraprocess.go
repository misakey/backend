package identity

import (
	"context"
	"database/sql"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
)

// IntraprocessHelper offers a set of functions to interact with the identity package entities without having to pass
// any storage context executor - in general from external modules wishing to access some data but not owning
// the storage connection behind it.
// NOTE: no transaction logic can be then used.
type IntraprocessHelper struct {
	sqlDB   *sql.DB
	redConn *redis.Client
}

// NewIntraprocessHelper ...
func NewIntraprocessHelper(ssoDB *sql.DB, redConn *redis.Client) *IntraprocessHelper {
	return &IntraprocessHelper{sqlDB: ssoDB, redConn: redConn}
}

// Get ...
func (ih IntraprocessHelper) Get(ctx context.Context, identityID string) (Identity, error) {
	return Get(ctx, ih.sqlDB, identityID)
}

// List ...
func (ih IntraprocessHelper) List(ctx context.Context, filters Filters) ([]*Identity, error) {
	return List(ctx, ih.sqlDB, filters)
}

// NotificationBulkCreate ...
func (ih IntraprocessHelper) NotificationBulkCreate(ctx context.Context, identityIDs []string, nType string, details null.JSON) error {
	return NotificationBulkCreate(ctx, ih.sqlDB, ih.redConn, identityIDs, nType, details)
}
