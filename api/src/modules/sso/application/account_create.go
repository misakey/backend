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
	identityID string

	Password   argon2.HashedPassword `json:"prehashed_password"`
	BackupData string                `json:"backup_data"`
}

func (cmd *CreateAccountCmd) SetIdentityID(id string) {
	cmd.identityID = id
}

func (cmd CreateAccountCmd) Validate() error {

	if err := v.ValidateStruct(&cmd,
		v.Field(&cmd.Password),
		v.Field(&cmd.identityID, v.Required, is.UUIDv4.Error("identity_id must be an UUIDv4")),
		v.Field(&cmd.BackupData, v.Required),
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
	BackupData    string `json:"backup_data"`
	BackupVersion int    `json:"backup_version"`
}

func (sso SSOService) CreateAccount(ctx context.Context, cmd CreateAccountCmd) (CreateAccountView, error) {
	view := CreateAccountView{}

	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return view, merror.Forbidden()
	}

	if acc.Subject != cmd.identityID {
		return view, merror.Forbidden()
	}

	// retrieve the concerned identity
	identity, err := sso.identityService.Get(ctx, cmd.identityID)
	if err != nil {
		return view, err
	}
	// the identity must be account free
	if identity.AccountID.String != "" {
		return view, merror.Conflict().
			Describe("identity already attached to an account").
			Detail("account_id", merror.DVConflict)
	}

	// prepare the account to be created
	account := domain.Account{
		BackupData:    cmd.BackupData,
		BackupVersion: 1,
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
	view.BackupVersion = account.BackupVersion // should be always 1
	view.PrehashedPassword.Params = cmd.Password.Params
	return view, nil
}
