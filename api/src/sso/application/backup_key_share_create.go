package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/crypto"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
)

// BackupKeyShareCreateCmd ...
type BackupKeyShareCreateCmd struct {
	AccountID      string `json:"account_id"`
	SaltBase64     string `json:"salt_base_64"`
	Share          string `json:"share"`
	OtherShareHash string `json:"other_share_hash"`
}

// BindAndValidate ...
func (cmd *BackupKeyShareCreateCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merr.BadRequest().Ori(merr.OriBody).Desc(err.Error())
	}

	if err := v.ValidateStruct(cmd,
		v.Field(&cmd.AccountID, v.Required, is.UUIDv4.Error("account_id must be an UUIDv4")),
		v.Field(&cmd.SaltBase64, is.Base64),
		v.Field(&cmd.OtherShareHash, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
		v.Field(&cmd.Share, v.Required, is.Base64),
	); err != nil {
		return merr.From(err).Desc("validating create backup key share command")
	}
	return nil
}

// CreateBackupKeyShare ...
func (sso *SSOService) CreateBackupKeyShare(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*BackupKeyShareCreateCmd)

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Forbidden()
	}

	// the request must bear authorization for an account
	identity, err := identity.Get(ctx, sso.ssoDB, acc.IdentityID)
	if err != nil {
		return nil, err
	}
	if identity.AccountID.IsZero() {
		return nil, merr.Conflict().Desc("no account id in authorization").Add("account_id", merr.DVConflict)
	}
	// the account id must be the same than the identity linked account
	if identity.AccountID.String != cmd.AccountID {
		return nil, merr.Forbidden().Desc("account_id does not match the querier account").Add("account_id", merr.DVForbidden)
	}

	backupKeyShare := crypto.BackupKeyShare{
		AccountID:      cmd.AccountID,
		SaltBase64:     cmd.SaltBase64,
		Share:          cmd.Share,
		OtherShareHash: cmd.OtherShareHash,
	}
	err = sso.backupKeyShareService.CreateBackupKeyShare(ctx, backupKeyShare)
	if err != nil {
		return nil, merr.From(err).Desc("creating")
	}
	return backupKeyShare, nil
}
