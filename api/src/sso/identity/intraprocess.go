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
	identity, err := Get(ctx, ih.sqlDB, identityID)
	sanitize(&identity)
	return identity, err

}

// Get ...
func (ih IntraprocessHelper) GetByIdentifierValue(ctx context.Context, identifierValue string) (Identity, error) {
	identity, err := GetByIdentifier(ctx, ih.sqlDB, identifierValue, AnyIdentifierKind)
	sanitize(&identity)
	return identity, err
}

// List ...
func (ih IntraprocessHelper) List(ctx context.Context, filters Filters) ([]*Identity, error) {
	identities, err := List(ctx, ih.sqlDB, filters)
	for _, identity := range identities {
		sanitize(identity)
	}
	return identities, err
}

// NotificationBulkCreate ...
func (ih IntraprocessHelper) NotificationBulkCreate(ctx context.Context, identityIDs []string, nType string, details null.JSON) error {
	return NotificationBulkCreate(ctx, ih.sqlDB, ih.redConn, identityIDs, nType, details)
}

// remove identity information that are never used by external modules
func sanitize(identity *Identity) {
	// Organizations have no identifier to display to other modules
	if identity.IdentifierKind == IdentifierKindOrgID {
		identity.IdentifierValue = ""
		identity.IdentifierKind = ""
	}
}
