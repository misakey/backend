package crypto

import (
	"context"
	"database/sql"

	"github.com/volatiletech/null/v8"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"
)

type IntraprocessHelper struct {
	sqlDB *sql.DB
}

func NewIntraprocessHelper(sqlDB *sql.DB) *IntraprocessHelper {
	return &IntraprocessHelper{sqlDB: sqlDB}
}

func (ih IntraprocessHelper) CreateCryptoActions(ctx context.Context, actions []Action) error {
	return CreateActions(ctx, ih.sqlDB, actions)
}

func (ih IntraprocessHelper) CreateInvitationActions(ctx context.Context, senderID, boxID, boxTitle, identifierValue string, actionsData null.JSON) error {
	tr, err := ih.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer atomic.SQLRollback(ctx, tr, err)
	err = CreateInvitationActions(ctx, tr, senderID, boxID, boxTitle, identifierValue, actionsData)
	if err != nil {
		return err
	}
	return tr.Commit()
}
