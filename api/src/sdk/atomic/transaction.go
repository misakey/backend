package atomic

import (
	"context"
	"database/sql"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
)

func SQLRollback(ctx context.Context, tr *sql.Tx, ptrErr *error) {
	if ptrErr == nil {
		return
	}
	if *ptrErr == nil {
		return
	}
	if rErr := tr.Rollback(); rErr != nil {
		logger.FromCtx(ctx).Warn().Msgf("rolling back: %v", rErr)
	}
}
