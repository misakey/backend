package crypto

import (
	"context"
	"database/sql"

	"github.com/go-redis/redis/v7"

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
	tr, err := ih.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer atomic.SQLRollback(ctx, tr, &err)
	err = CreateActions(ctx, ih.sqlDB, actions)
	if err != nil {
		return err
	}
	return tr.Commit()
}
