package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/crypto"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
)

func (sso *SSOService) GetSecretStorage(ctx context.Context, gen request.Request) (interface{}, error) {
	acc := oidc.GetAccesses(ctx)
	// querier must have an account
	if acc == nil || acc.AccountID.IsZero() {
		return nil, merr.Forbidden()
	}

	secrets, err := crypto.GetAccountSecrets(ctx, sso.sqlDB, acc.AccountID.String)
	if err != nil {
		if err == crypto.ErrNoRootKey {
			return nil, merr.Conflict().Desc("Account has no root key; it requires migration.")
		}

		return nil, merr.From(err)
	}

	return secrets, nil
}

type MigrateToSecretStorageQuery = crypto.SecretStorageSetupData

func (sso *SSOService) MigrateToSecretStorage(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*crypto.SecretStorageSetupData)

	acc := oidc.GetAccesses(ctx)
	// querier must have an account
	if acc == nil || acc.AccountID.IsZero() {
		return nil, merr.Forbidden()
	}

	tr, err := sso.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, &err)

	err = crypto.ResetAccountSecretStorage(ctx, tr, acc.AccountID.String, query)
	if err != nil {
		return nil, merr.From(err).Desc("migrating account")
	}

	curIdentity, err := identity.Get(ctx, tr, acc.IdentityID)
	if err != nil {
		return nil, merr.From(err).Desc("retrieving identity")
	}

	if curIdentity.Pubkey.String == "" || curIdentity.NonIdentifiedPubkey.String == "" {
		err = curIdentity.SetIdentityKeys(
			query.IdentityPublicKey,
			query.IdentityNonIdentifiedPublicKey,
		)
		if err != nil {
			return nil, merr.BadRequest().Ori(merr.OriBody).Desc(err.Error())
		}

		err = identity.Update(ctx, tr, &curIdentity)
		if err != nil {
			return nil, merr.From(err).Desc("updating identity")
		}
	} else {
		if query.IdentityPublicKey != "" || query.IdentityNonIdentifiedPublicKey != "" {
			return nil, merr.BadRequest().Ori(merr.OriBody).
				Desc("unexpected identity keys: identity already has identity keys")
		}
	}

	if cErr := tr.Commit(); cErr != nil {
		return nil, merr.From(cErr).Desc("committing transaction")
	}

	return nil, nil
}

func (sso *SSOService) CreateSecretStorageAsymKey(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*crypto.SecretStorageAsymKey)

	acc := oidc.GetAccesses(ctx)
	// querier must have an account
	if acc == nil || acc.AccountID.IsZero() {
		return nil, merr.Forbidden()
	}

	response, err := crypto.CreateSecretStorageAsymKey(ctx, sso.sqlDB, acc.AccountID.String, *cmd)

	return response, err
}

func (sso *SSOService) CreateSecretStorageBoxKeyShare(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*crypto.SecretStorageBoxKeyShare)

	acc := oidc.GetAccesses(ctx)
	// querier must have an account
	if acc == nil || acc.AccountID.IsZero() {
		return nil, merr.Forbidden()
	}

	tr, err := sso.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, &err)

	response, err := crypto.CreateSecretStorageBoxKeyShare(ctx, tr, acc.AccountID.String, *cmd)
	if err != nil {
		return nil, merr.From(err).Desc("creating box key share")
	}

	if cErr := tr.Commit(); cErr != nil {
		return nil, merr.From(cErr).Desc("committing transaction")
	}

	return response, err
}

type DeleteAsymKeysCmd struct {
	Pubkeys []string `json:"public_keys"`
}

// BindAndValidate implements request.Request.BindAndValidate
func (cmd *DeleteAsymKeysCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merr.From(err).Ori(merr.OriBody)
	}

	if err := v.ValidateStruct(cmd,
		v.Field(&cmd.Pubkeys, v.Required, v.Each(v.Match(format.UnpaddedURLSafeBase64))),
	); err != nil {
		return err
	}

	return nil
}

func (sso *SSOService) DeleteAsymKeys(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*DeleteAsymKeysCmd)

	acc := oidc.GetAccesses(ctx)
	// querier must have an account
	if acc == nil || acc.AccountID.IsZero() {
		return nil, merr.Forbidden()
	}

	tr, err := sso.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, &err)

	err = crypto.DeleteAsymKeys(ctx, tr, acc.AccountID.String, cmd.Pubkeys)
	if err != nil {
		return nil, merr.From(err).Desc("deleting asym keys")
	}

	if cErr := tr.Commit(); cErr != nil {
		return nil, merr.From(cErr).Desc("committing transaction")
	}

	return nil, err
}

type DeleteBoxKeySharesCmd struct {
	BoxIDs []string `json:"box_ids"`
}

// BindAndValidate implements request.Request.BindAndValidate
func (cmd *DeleteBoxKeySharesCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merr.From(err).Ori(merr.OriBody)
	}

	if err := v.ValidateStruct(cmd,
		v.Field(&cmd.BoxIDs, v.Required, v.Each(is.UUIDv4)),
	); err != nil {
		return err
	}

	return nil
}

func (sso *SSOService) DeleteBoxKeyShares(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*DeleteBoxKeySharesCmd)

	acc := oidc.GetAccesses(ctx)
	// querier must have an account
	if acc == nil || acc.AccountID.IsZero() {
		return nil, merr.Forbidden()
	}

	tr, err := sso.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, &err)

	err = crypto.DeleteBoxKeyShares(ctx, tr, acc.AccountID.String, cmd.BoxIDs)
	if err != nil {
		return nil, merr.From(err).Desc("deleting box key shares")
	}

	if cErr := tr.Commit(); cErr != nil {
		return nil, merr.From(cErr).Desc("committing transaction")
	}

	return nil, err
}
