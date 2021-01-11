package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
)

// BackupKeyShareQuery ...
type BackupKeyShareQuery struct {
	otherShareHash string
}

// BindAndValidate ...
func (query *BackupKeyShareQuery) BindAndValidate(eCtx echo.Context) error {
	query.otherShareHash = eCtx.Param("other-share-hash")

	if err := v.ValidateStruct(query,
		v.Field(&query.otherShareHash, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
	); err != nil {
		return merr.From(err).Desc("validating create backup key share query")
	}
	return nil
}

// GetBackupKeyShare ...
func (sso *SSOService) GetBackupKeyShare(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*BackupKeyShareQuery)

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Forbidden()
	}

	backupKeyShare, err := sso.backupKeyShareService.GetBackupKeyShare(ctx, query.otherShareHash)
	if err != nil {
		return nil, err
	}

	// the request must bear authorization for an account
	if acc.AccountID.IsZero() {
		return nil, merr.Conflict().
			Desc("no account id in authorization").
			Add("account_id", merr.DVConflict)
	}
	// the account id must be the same than the identity linked account
	if acc.AccountID.String != backupKeyShare.AccountID {
		return nil, merr.NotFound()
	}

	return backupKeyShare, nil
}
