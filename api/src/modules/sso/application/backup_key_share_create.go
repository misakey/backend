package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type CreateBackupKeyShareCmd struct {
	AccountID      string `json:"account_id"`
	Share          string `json:"share"`
	OtherShareHash string `json:"other_share_hash"`
}

func (cmd CreateBackupKeyShareCmd) Validate() error {

	if err := v.ValidateStruct(&cmd,
		v.Field(&cmd.AccountID, v.Required, is.UUIDv4.Error("account_id must be an UUIDv4")),
		v.Field(&cmd.OtherShareHash, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
		v.Field(&cmd.Share, v.Required, is.Base64),
	); err != nil {
		return merror.Transform(err).Describe("validating create backup key share command")
	}
	return nil
}

func (sso SSOService) BackupKeyShareCreate(ctx context.Context, cmd CreateBackupKeyShareCmd) error {

	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return merror.Forbidden()
	}

	// retrieve the concerned identity
	identity, err := sso.identityService.Get(ctx, acc.Subject)
	if err != nil {
		return err
	}
	// the identity must have an account
	if identity.AccountID.IsZero() {
		return merror.Conflict().
			Describe("identity must have a linked account").
			Detail("account_id", merror.DVConflict)
	}
	// the account id must be the same than the identity linked account
	if identity.AccountID.String != cmd.AccountID {
		return merror.Forbidden().Describe("account_id does not match the querier account").Detail("account_id", merror.DVForbidden)
	}

	backupKeyShare := domain.BackupKeyShare{
		AccountID:      cmd.AccountID,
		Share:          cmd.Share,
		OtherShareHash: cmd.OtherShareHash,
	}

	return sso.backupKeyShareService.CreateBackupKeyShare(ctx, backupKeyShare)
}
