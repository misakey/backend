package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/ajwt"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type BackupKeyShareQuery struct {
	OtherShareHash string
}

func (cmd BackupKeyShareQuery) Validate() error {

	if err := v.ValidateStruct(&cmd,
		v.Field(&cmd.OtherShareHash, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
	); err != nil {
		return merror.Transform(err).Describe("validating create backup key share command")
	}
	return nil
}

func (sso SSOService) BackupKeyShareGet(ctx context.Context, query BackupKeyShareQuery) (*domain.BackupKeyShare, error) {

	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Forbidden()
	}

	backupKeyShare, err := sso.backupKeyShareService.GetBackupKeyShare(ctx, query.OtherShareHash)
	if err != nil {
		return nil, err
	}

	// the request must bear authorization for an account
	if acc.AccountID.IsZero() {
		return nil, merror.Conflict().
			Describe("no account id in authorization").
			Detail("account_id", merror.DVConflict)
	}
	// the account id must be the same than the identity linked account
	if acc.AccountID.String != backupKeyShare.AccountID {
		return nil, merror.NotFound()
	}

	return backupKeyShare, nil
}
