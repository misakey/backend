package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn/argon2"
)

type CreateAccountCmd struct {
	IdentityID string

	Password   argon2.HashedPassword `json:"prehashed_password"`
	BackupData string                `json:"backup_data"`
}

func (cmd CreateAccountCmd) Validate() error {
	// validate nested structure separately
	err := v.ValidateStruct(&cmd.Password, v.Field(&cmd.Password.HashBase64, v.Required))
	if err != nil {
		return merror.Transform(err).Describe("validating prehashed password")
	}

	if err := v.ValidateStruct(&cmd.Password.Params,
		v.Field(&cmd.Password.Params.Memory, v.Required),
		v.Field(&cmd.Password.Params.Iterations, v.Required),
		v.Field(&cmd.Password.Params.Parallelism, v.Required),
		v.Field(&cmd.Password.Params.SaltBase64, v.Required, is.Base64.Error("salt_base64 must be base64 encoded")),
	); err != nil {
		return merror.Transform(err).Describe("validating prehashed password params")
	}

	if err := v.ValidateStruct(&cmd,
		v.Field(&cmd.IdentityID, v.Required, is.UUIDv4.Error("identity_id must be an UUIDv4")),
		v.Field(&cmd.BackupData, v.Required, is.Base64.Error("backup_data must be base64 encoded")),
	); err != nil {
		return merror.Transform(err).Describe("validating create account command")
	}
	return nil
}

type CreateAccountView struct {
	ID                string `json:"id"`
	PrehashedPassword struct {
		Params argon2.Params `json:"params"`
	} `json:"prehashed_password"`
	BackupData string `json:"backup_data"`
}

func (sso SSOService) CreateAccount(ctx context.Context, cmd CreateAccountCmd) (CreateAccountView, error) {
	view := CreateAccountView{}
	// TODO: check the identity is inside the token
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return view, merror.Forbidden()
	}

	if acc.Subject != cmd.IdentityID {
		return view, merror.Forbidden()
	}

	// retrieve the concerned identity
	identity, err := sso.identityService.Get(ctx, cmd.IdentityID)
	if err != nil {
		return view, err
	}
	// the identity must be account free
	if identity.AccountID.Valid {
		return view, merror.Conflict().Describe("identity already attached to an account").
			Detail("account_id", merror.DVConflict)
	}

	// prepare the account to be created
	account := domain.Account{
		BackupData: cmd.BackupData,
	}
	account.Password, err = cmd.Password.Hash()
	if err != nil {
		return view, merror.Transform(err).Describe("could not hash the password")
	}

	// hash the password before storing it
	if err := sso.accountService.Create(ctx, &account); err != nil {
		return view, err
	}

	// update the identity's account id column
	identity.AccountID = null.StringFrom(account.ID)
	if err := sso.identityService.Update(ctx, &identity); err != nil {
		return view, err
	}

	// fill view model with domain model
	view.ID = account.ID
	view.BackupData = account.BackupData
	view.PrehashedPassword.Params = cmd.Password.Params
	return view, nil
}
