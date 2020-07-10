package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

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

	// retrieve the concerned identity
	identity, err := sso.identityService.Get(ctx, acc.Subject)
	if err != nil {
		return nil, err
	}
	// the identity must have an account
	if identity.AccountID.IsZero() {
		return nil, merror.Conflict().
			Describe("identity must have a linked account").
			Detail("account_id", merror.DVConflict)
	}
	// the account id must be the same than the identity linked account
	if identity.AccountID.String != backupKeyShare.AccountID {
		return nil, merror.NotFound()
	}

	return backupKeyShare, nil
}
