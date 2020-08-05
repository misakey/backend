package atomic

import (
	"context"
	"database/sql"

	"gitlab.misakey.dev/misakey/msk-sdk-go/logger"
)

func SQLRollback(ctx context.Context, tr *sql.Tx, err error) {
	if err == nil {
		return
	}
	if rErr := tr.Rollback(); rErr != nil {
		logger.FromCtx(ctx).Warn().Msgf("rolling back: %v", err)
	}
}
