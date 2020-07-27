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
	SaltBase64     string `json:"salt_base64"`
	Share          string `json:"share"`
	OtherShareHash string `json:"other_share_hash"`
}

func (cmd CreateBackupKeyShareCmd) Validate() error {

	if err := v.ValidateStruct(&cmd,
		v.Field(&cmd.AccountID, v.Required, is.UUIDv4.Error("account_id must be an UUIDv4")),
		v.Field(&cmd.SaltBase64, is.Base64),
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

	// the request must bear authorization for an account
	identity, err := sso.identityService.Get(ctx, acc.IdentityID)
	if err != nil {
		return err
	}
	if identity.AccountID.IsZero() {
		return merror.Conflict().
			Describe("no account id in authorization").
			Detail("account_id", merror.DVConflict)
	}
	// the account id must be the same than the identity linked account
	if identity.AccountID.String != cmd.AccountID {
		return merror.Forbidden().Describe("account_id does not match the querier account").Detail("account_id", merror.DVForbidden)
	}

	backupKeyShare := domain.BackupKeyShare{
		AccountID:      cmd.AccountID,
		SaltBase64:     cmd.SaltBase64,
		Share:          cmd.Share,
		OtherShareHash: cmd.OtherShareHash,
	}

	return sso.backupKeyShareService.CreateBackupKeyShare(ctx, backupKeyShare)
}
