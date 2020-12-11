package crypto

import (
	"context"
	"database/sql"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"
)

// IntraprocessHelper ...
type IntraprocessHelper struct {
	sqlDB   *sql.DB
	redConn *redis.Client
}

// NewIntraprocessHelper ...
func NewIntraprocessHelper(sqlDB *sql.DB, redConn *redis.Client) *IntraprocessHelper {
	return &IntraprocessHelper{sqlDB: sqlDB, redConn: redConn}
}

// CreateActions ...
func (ih IntraprocessHelper) CreateActions(ctx context.Context, actions []Action) error {
	return CreateActions(ctx, ih.sqlDB, actions)
}

// CreateInvitationActionsForIdentity ...
func (ih IntraprocessHelper) CreateInvitationActionsForIdentity(ctx context.Context, senderID, boxID, boxTitle, identityValue string, actionsData null.JSON) error {
	tr, err := ih.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer atomic.SQLRollback(ctx, tr, &err)
	err = CreateInvitationActionsForIdentity(ctx, tr, ih.redConn, senderID, boxID, boxTitle, identityValue, actionsData)
	if err != nil {
		return err
	}
	return tr.Commit()
}

// CreateInvitationActionsForIdentifier ...
func (ih IntraprocessHelper) CreateInvitationActionsForIdentifier(ctx context.Context, senderID, boxID, boxTitle, identifierValue string, actionsData null.JSON) error {
	tr, err := ih.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer atomic.SQLRollback(ctx, tr, &err)
	err = CreateInvitationActionsForIdentifier(ctx, tr, ih.redConn, senderID, boxID, boxTitle, identifierValue, actionsData)
	if err != nil {
		return err
	}
	return tr.Commit()
}
